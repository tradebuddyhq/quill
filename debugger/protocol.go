package debugger

import (
	"bufio"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// CDPMessage represents a Chrome DevTools Protocol message (request, response, or event).
type CDPMessage struct {
	ID     int              `json:"id,omitempty"`
	Method string           `json:"method,omitempty"`
	Params *json.RawMessage `json:"params,omitempty"`
	Result *json.RawMessage `json:"result,omitempty"`
	Error  *CDPError        `json:"error,omitempty"`
}

// CDPError represents an error in a CDP response.
type CDPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// PausedEvent contains data from Debugger.paused events.
type PausedEvent struct {
	CallFrames []CallFrame `json:"callFrames"`
	Reason     string      `json:"reason"`
	HitBreakpoints []string `json:"hitBreakpoints"`
}

// CallFrame represents a frame in the call stack.
type CallFrame struct {
	CallFrameID  string   `json:"callFrameId"`
	FunctionName string   `json:"functionName"`
	Location     Location `json:"location"`
	ScopeChain   []Scope  `json:"scopeChain"`
}

// Location represents a script location.
type Location struct {
	ScriptID     string `json:"scriptId"`
	LineNumber   int    `json:"lineNumber"`
	ColumnNumber int    `json:"columnNumber"`
}

// Scope represents a variable scope.
type Scope struct {
	Type   string          `json:"type"`
	Object RemoteObjectRef `json:"object"`
}

// RemoteObjectRef is a reference to a remote object in the debuggee.
type RemoteObjectRef struct {
	Type        string `json:"type"`
	ClassName   string `json:"className,omitempty"`
	Description string `json:"description,omitempty"`
	ObjectID    string `json:"objectId,omitempty"`
}

// RemoteObject represents a value returned from the debuggee.
type RemoteObject struct {
	Type        string `json:"type"`
	Subtype     string `json:"subtype,omitempty"`
	Value       interface{} `json:"value,omitempty"`
	Description string `json:"description,omitempty"`
	ClassName   string `json:"className,omitempty"`
	ObjectID    string `json:"objectId,omitempty"`
}

// PropertyDescriptor describes a property on a remote object.
type PropertyDescriptor struct {
	Name  string       `json:"name"`
	Value RemoteObject `json:"value"`
}

// ScriptParsedEvent is emitted when Node parses a script.
type ScriptParsedEvent struct {
	ScriptID string `json:"scriptId"`
	URL      string `json:"url"`
}

// CDPClient is a minimal Chrome DevTools Protocol client using raw WebSocket.
type CDPClient struct {
	conn      net.Conn
	reader    *bufio.Reader
	writeMu   sync.Mutex
	nextID    atomic.Int64
	pending   map[int]chan CDPMessage
	pendingMu sync.Mutex
	events    chan CDPMessage
	done      chan struct{}
	closed    atomic.Bool
}

// NewCDPClient connects to a CDP WebSocket endpoint.
func NewCDPClient(wsURL string) (*CDPClient, error) {
	// Parse ws://host:port/path
	wsURL = strings.TrimPrefix(wsURL, "ws://")
	slashIdx := strings.Index(wsURL, "/")
	host := wsURL
	path := "/"
	if slashIdx >= 0 {
		host = wsURL[:slashIdx]
		path = wsURL[slashIdx:]
	}

	conn, err := net.DialTimeout("tcp", host, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", host, err)
	}

	// Perform WebSocket handshake
	key := generateWSKey()
	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n", path, host, key)
	if _, err := conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshake write failed: %w", err)
	}

	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshake response failed: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 101 {
		conn.Close()
		return nil, fmt.Errorf("WebSocket handshake failed: status %d", resp.StatusCode)
	}

	client := &CDPClient{
		conn:    conn,
		reader:  reader,
		pending: make(map[int]chan CDPMessage),
		events:  make(chan CDPMessage, 64),
		done:    make(chan struct{}),
	}

	go client.readLoop()
	return client, nil
}

// Send sends a CDP command and waits for its response.
func (c *CDPClient) Send(method string, params interface{}) (*json.RawMessage, error) {
	id := int(c.nextID.Add(1))
	ch := make(chan CDPMessage, 1)

	c.pendingMu.Lock()
	c.pending[id] = ch
	c.pendingMu.Unlock()

	msg := struct {
		ID     int         `json:"id"`
		Method string      `json:"method"`
		Params interface{} `json:"params,omitempty"`
	}{
		ID:     id,
		Method: method,
		Params: params,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	if err := c.writeFrame(data); err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("CDP error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	case <-time.After(30 * time.Second):
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return nil, fmt.Errorf("timeout waiting for response to %s (id=%d)", method, id)
	case <-c.done:
		return nil, fmt.Errorf("connection closed")
	}
}

// Events returns the channel of CDP events (Debugger.paused, etc.).
func (c *CDPClient) Events() <-chan CDPMessage {
	return c.events
}

// Close shuts down the CDP connection.
func (c *CDPClient) Close() error {
	if c.closed.CompareAndSwap(false, true) {
		close(c.done)
		return c.conn.Close()
	}
	return nil
}

// readLoop continuously reads WebSocket frames and dispatches them.
func (c *CDPClient) readLoop() {
	defer func() {
		if c.closed.CompareAndSwap(false, true) {
			close(c.done)
		}
	}()

	for {
		payload, err := c.readFrame()
		if err != nil {
			return
		}

		var msg CDPMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			continue
		}

		if msg.ID > 0 {
			// Response to a request
			c.pendingMu.Lock()
			ch, ok := c.pending[msg.ID]
			if ok {
				delete(c.pending, msg.ID)
			}
			c.pendingMu.Unlock()
			if ok {
				ch <- msg
			}
		} else if msg.Method != "" {
			// Event
			select {
			case c.events <- msg:
			default:
				// Drop if full
			}
		}
	}
}

// readFrame reads a single WebSocket frame (handles fragmentation and continuation).
func (c *CDPClient) readFrame() ([]byte, error) {
	var fullPayload []byte
	for {
		// Read first 2 bytes
		header := make([]byte, 2)
		if _, err := io.ReadFull(c.reader, header); err != nil {
			return nil, err
		}

		fin := header[0]&0x80 != 0
		opcode := header[0] & 0x0f
		masked := header[1]&0x80 != 0
		payloadLen := uint64(header[1] & 0x7f)

		if opcode == 0x08 { // close frame
			return nil, fmt.Errorf("WebSocket closed")
		}

		if payloadLen == 126 {
			ext := make([]byte, 2)
			if _, err := io.ReadFull(c.reader, ext); err != nil {
				return nil, err
			}
			payloadLen = uint64(binary.BigEndian.Uint16(ext))
		} else if payloadLen == 127 {
			ext := make([]byte, 8)
			if _, err := io.ReadFull(c.reader, ext); err != nil {
				return nil, err
			}
			payloadLen = binary.BigEndian.Uint64(ext)
		}

		var maskKey []byte
		if masked {
			maskKey = make([]byte, 4)
			if _, err := io.ReadFull(c.reader, maskKey); err != nil {
				return nil, err
			}
		}

		payload := make([]byte, payloadLen)
		if payloadLen > 0 {
			if _, err := io.ReadFull(c.reader, payload); err != nil {
				return nil, err
			}
		}

		if masked {
			for i := range payload {
				payload[i] ^= maskKey[i%4]
			}
		}

		fullPayload = append(fullPayload, payload...)

		if fin {
			_ = opcode // text or binary, we treat both as text
			return fullPayload, nil
		}
	}
}

// writeFrame writes a masked WebSocket text frame.
func (c *CDPClient) writeFrame(payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// Build frame header
	var frame []byte

	// FIN + text opcode
	frame = append(frame, 0x81)

	// Payload length + mask bit
	maskBit := byte(0x80) // client must mask
	plen := len(payload)
	if plen < 126 {
		frame = append(frame, maskBit|byte(plen))
	} else if plen < 65536 {
		frame = append(frame, maskBit|126)
		frame = append(frame, byte(plen>>8), byte(plen))
	} else {
		frame = append(frame, maskBit|127)
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(plen))
		frame = append(frame, b...)
	}

	// Masking key
	maskKey := make([]byte, 4)
	rand.Read(maskKey)
	frame = append(frame, maskKey...)

	// Masked payload
	masked := make([]byte, plen)
	for i, b := range payload {
		masked[i] = b ^ maskKey[i%4]
	}
	frame = append(frame, masked...)

	_, err := c.conn.Write(frame)
	return err
}

// generateWSKey generates a random WebSocket key for the handshake.
func generateWSKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// computeWSAccept computes the expected Sec-WebSocket-Accept value (unused but kept for reference).
func computeWSAccept(key string) string {
	const magic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(key + magic))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
