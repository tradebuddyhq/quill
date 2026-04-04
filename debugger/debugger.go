package debugger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"quill/codegen"
	"quill/lexer"
	"quill/parser"
	"regexp"
	"strings"
	"time"
)

// StackFrame represents a single frame in the Quill call stack.
type StackFrame struct {
	FunctionName string
	QuillLine    int
	JSLine       int
}

// Debugger is the main debug engine that bridges Quill source to Node.js inspector.
type Debugger struct {
	sourceFile   string
	compiledFile string
	mapFile      string
	sourceMap    *SourceMapData
	breakpoints  map[int]string // Quill line -> CDP breakpoint ID
	cdp          *CDPClient
	nodeCmd      *exec.Cmd
	scriptID     string
	quillLines   []string // original source lines (1-indexed via quillLines[i-1])
	currentLine  int      // current Quill line (1-based)
	paused       bool
	callFrames   []CallFrame
	done         chan struct{}
}

// New creates a new Debugger for the given Quill source file.
func New(quillFile string) *Debugger {
	return &Debugger{
		sourceFile:  quillFile,
		breakpoints: make(map[int]string),
		done:        make(chan struct{}),
	}
}

// Start compiles the Quill file, launches Node with --inspect-brk, and connects.
func (d *Debugger) Start() error {
	// Read source lines
	source, err := os.ReadFile(d.sourceFile)
	if err != nil {
		return fmt.Errorf("could not read %q: %w", d.sourceFile, err)
	}
	d.quillLines = strings.Split(string(source), "\n")

	// Compile to JS with source map
	l := lexer.New(string(source))
	tokens, err := l.Tokenize()
	if err != nil {
		return fmt.Errorf("lexer error: %w", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parser error: %w", err)
	}

	ext := filepath.Ext(d.sourceFile)
	base := d.sourceFile[:len(d.sourceFile)-len(ext)]
	d.compiledFile = base + ".debug.js"
	d.mapFile = d.compiledFile + ".map"

	g := codegen.New()
	js, sourceMapJSON := g.GenerateWithSourceMap(program, d.sourceFile, d.compiledFile)

	// Append source mapping URL
	js += "\n//# sourceMappingURL=" + filepath.Base(d.mapFile) + "\n"

	if err := os.WriteFile(d.compiledFile, []byte(js), 0644); err != nil {
		return fmt.Errorf("could not write compiled JS: %w", err)
	}
	if err := os.WriteFile(d.mapFile, []byte(sourceMapJSON), 0644); err != nil {
		return fmt.Errorf("could not write source map: %w", err)
	}

	// Load the source map for line translation
	sm, err := LoadSourceMapFromJSON(sourceMapJSON)
	if err != nil {
		return fmt.Errorf("could not parse source map: %w", err)
	}
	d.sourceMap = sm

	// Launch Node.js with --inspect-brk
	absPath, _ := filepath.Abs(d.compiledFile)
	d.nodeCmd = exec.Command("node", "--inspect-brk=0", absPath)
	d.nodeCmd.Stdin = nil
	d.nodeCmd.Stdout = os.Stdout

	// Capture stderr to find the WebSocket URL
	stderrPipe, err := d.nodeCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("could not create stderr pipe: %w", err)
	}

	if err := d.nodeCmd.Start(); err != nil {
		return fmt.Errorf("could not start Node.js: %w", err)
	}

	// Parse the inspector URL from Node's stderr
	wsURL, err := parseInspectorURL(stderrPipe)
	if err != nil {
		d.nodeCmd.Process.Kill()
		return fmt.Errorf("could not get inspector URL: %w", err)
	}

	// Give Node a moment to be ready
	time.Sleep(100 * time.Millisecond)

	// Connect CDP
	cdp, err := NewCDPClient(wsURL)
	if err != nil {
		d.nodeCmd.Process.Kill()
		return fmt.Errorf("could not connect to inspector: %w", err)
	}
	d.cdp = cdp

	// Enable the debugger domain
	if _, err := d.cdp.Send("Debugger.enable", nil); err != nil {
		d.Stop()
		return fmt.Errorf("Debugger.enable failed: %w", err)
	}

	// Enable runtime domain for evaluation
	if _, err := d.cdp.Send("Runtime.enable", nil); err != nil {
		d.Stop()
		return fmt.Errorf("Runtime.enable failed: %w", err)
	}

	// Listen for events in background
	go d.eventLoop()

	// Wait briefly for the initial pause (--inspect-brk pauses on first line)
	time.Sleep(200 * time.Millisecond)

	// Process any queued paused events
	d.drainEvents()

	return nil
}

// eventLoop processes CDP events.
func (d *Debugger) eventLoop() {
	for {
		select {
		case <-d.done:
			return
		case evt, ok := <-d.cdp.Events():
			if !ok {
				return
			}
			d.handleEvent(evt)
		}
	}
}

// drainEvents processes pending events without blocking.
func (d *Debugger) drainEvents() {
	for {
		select {
		case evt := <-d.cdp.Events():
			d.handleEvent(evt)
		default:
			return
		}
	}
}

// handleEvent processes a single CDP event.
func (d *Debugger) handleEvent(evt CDPMessage) {
	switch evt.Method {
	case "Debugger.paused":
		if evt.Params != nil {
			var pe PausedEvent
			json.Unmarshal(*evt.Params, &pe)
			d.paused = true
			d.callFrames = pe.CallFrames
			if len(pe.CallFrames) > 0 {
				jsLine := pe.CallFrames[0].Location.LineNumber
				if d.sourceMap != nil {
					ql := d.sourceMap.JSLineToQuillLine(jsLine)
					if ql > 0 {
						d.currentLine = ql
					}
				}
				// Capture script ID from first frame
				if d.scriptID == "" {
					d.scriptID = pe.CallFrames[0].Location.ScriptID
				}
			}
		}
	case "Debugger.resumed":
		d.paused = false
	case "Debugger.scriptParsed":
		if evt.Params != nil {
			var sp ScriptParsedEvent
			json.Unmarshal(*evt.Params, &sp)
			// Track our compiled script
			if strings.Contains(sp.URL, filepath.Base(d.compiledFile)) {
				d.scriptID = sp.ScriptID
			}
		}
	}
}

// waitForPause waits until the debugger pauses or a timeout expires.
func (d *Debugger) waitForPause() error {
	deadline := time.After(30 * time.Second)
	for {
		d.drainEvents()
		if d.paused {
			return nil
		}
		select {
		case <-deadline:
			return fmt.Errorf("timeout waiting for debugger to pause")
		case <-time.After(50 * time.Millisecond):
			// poll again
		}
	}
}

// SetBreakpoint sets a breakpoint on the given Quill line (1-based).
func (d *Debugger) SetBreakpoint(line int) error {
	if line < 1 || line > len(d.quillLines) {
		return fmt.Errorf("line %d is out of range (1-%d)", line, len(d.quillLines))
	}

	if _, exists := d.breakpoints[line]; exists {
		return fmt.Errorf("breakpoint already set on line %d", line)
	}

	jsLine := d.sourceMap.QuillLineToJSLine(line)
	if jsLine < 0 {
		return fmt.Errorf("line %d does not map to any compiled JavaScript line", line)
	}

	params := map[string]interface{}{
		"lineNumber": jsLine,
		"url":        filepath.Base(d.compiledFile),
	}

	result, err := d.cdp.Send("Debugger.setBreakpointByUrl", params)
	if err != nil {
		return fmt.Errorf("failed to set breakpoint: %w", err)
	}

	// Extract breakpoint ID
	var resp struct {
		BreakpointID string `json:"breakpointId"`
	}
	if result != nil {
		json.Unmarshal(*result, &resp)
	}

	d.breakpoints[line] = resp.BreakpointID
	return nil
}

// RemoveBreakpoint removes the breakpoint on the given Quill line (1-based).
func (d *Debugger) RemoveBreakpoint(line int) error {
	bpID, exists := d.breakpoints[line]
	if !exists {
		return fmt.Errorf("no breakpoint on line %d", line)
	}

	params := map[string]interface{}{
		"breakpointId": bpID,
	}

	if _, err := d.cdp.Send("Debugger.removeBreakpoint", params); err != nil {
		return fmt.Errorf("failed to remove breakpoint: %w", err)
	}

	delete(d.breakpoints, line)
	return nil
}

// Continue resumes execution.
func (d *Debugger) Continue() error {
	d.paused = false
	_, err := d.cdp.Send("Debugger.resume", nil)
	if err != nil {
		return err
	}
	return d.waitForPause()
}

// StepOver steps to the next line.
func (d *Debugger) StepOver() error {
	d.paused = false
	_, err := d.cdp.Send("Debugger.stepOver", nil)
	if err != nil {
		return err
	}
	return d.waitForPause()
}

// StepInto steps into a function call.
func (d *Debugger) StepInto() error {
	d.paused = false
	_, err := d.cdp.Send("Debugger.stepInto", nil)
	if err != nil {
		return err
	}
	return d.waitForPause()
}

// StepOut steps out of the current function.
func (d *Debugger) StepOut() error {
	d.paused = false
	_, err := d.cdp.Send("Debugger.stepOut", nil)
	if err != nil {
		return err
	}
	return d.waitForPause()
}

// Evaluate evaluates an expression in the current call frame.
func (d *Debugger) Evaluate(expr string) (string, error) {
	if !d.paused || len(d.callFrames) == 0 {
		return "", fmt.Errorf("not paused")
	}

	params := map[string]interface{}{
		"callFrameId": d.callFrames[0].CallFrameID,
		"expression":  expr,
	}

	result, err := d.cdp.Send("Debugger.evaluateOnCallFrame", params)
	if err != nil {
		return "", err
	}

	var resp struct {
		Result RemoteObject `json:"result"`
		ExceptionDetails *struct {
			Text string `json:"text"`
		} `json:"exceptionDetails"`
	}
	if result != nil {
		json.Unmarshal(*result, &resp)
	}

	if resp.ExceptionDetails != nil {
		return "", fmt.Errorf("evaluation error: %s", resp.ExceptionDetails.Text)
	}

	return formatRemoteObject(resp.Result), nil
}

// GetLocals returns local variable names and values from the current scope.
func (d *Debugger) GetLocals() (map[string]string, error) {
	if !d.paused || len(d.callFrames) == 0 {
		return nil, fmt.Errorf("not paused")
	}

	locals := make(map[string]string)

	// Find the local scope in the current frame
	for _, scope := range d.callFrames[0].ScopeChain {
		if scope.Type != "local" {
			continue
		}

		if scope.Object.ObjectID == "" {
			continue
		}

		params := map[string]interface{}{
			"objectId":               scope.Object.ObjectID,
			"ownProperties":          true,
			"generatePreview":        true,
		}

		result, err := d.cdp.Send("Runtime.getProperties", params)
		if err != nil {
			return nil, err
		}

		var resp struct {
			Result []PropertyDescriptor `json:"result"`
		}
		if result != nil {
			json.Unmarshal(*result, &resp)
		}

		for _, prop := range resp.Result {
			// Skip internal properties
			if strings.HasPrefix(prop.Name, "__") {
				continue
			}
			locals[prop.Name] = formatRemoteObject(prop.Value)
		}
		break
	}

	return locals, nil
}

// GetCallStack returns the call stack with Quill source line numbers.
func (d *Debugger) GetCallStack() ([]StackFrame, error) {
	if !d.paused {
		return nil, fmt.Errorf("not paused")
	}

	var frames []StackFrame
	for _, cf := range d.callFrames {
		jsLine := cf.Location.LineNumber
		quillLine := 0
		if d.sourceMap != nil {
			quillLine = d.sourceMap.JSLineToQuillLine(jsLine)
		}
		frames = append(frames, StackFrame{
			FunctionName: cf.FunctionName,
			QuillLine:    quillLine,
			JSLine:       jsLine,
		})
	}
	return frames, nil
}

// IsPaused returns whether the debugger is currently paused.
func (d *Debugger) IsPaused() bool {
	return d.paused
}

// CurrentLine returns the current Quill source line (1-based).
func (d *Debugger) CurrentLine() int {
	return d.currentLine
}

// QuillLines returns the source lines of the Quill file.
func (d *Debugger) QuillLines() []string {
	return d.quillLines
}

// Breakpoints returns the set of active breakpoint lines.
func (d *Debugger) Breakpoints() map[int]string {
	return d.breakpoints
}

// SourceFile returns the path of the source file being debugged.
func (d *Debugger) SourceFile() string {
	return d.sourceFile
}

// Stop kills the Node process and cleans up temporary files.
func (d *Debugger) Stop() error {
	select {
	case <-d.done:
		// Already closed
	default:
		close(d.done)
	}

	if d.cdp != nil {
		d.cdp.Close()
	}

	if d.nodeCmd != nil && d.nodeCmd.Process != nil {
		d.nodeCmd.Process.Kill()
		d.nodeCmd.Wait()
	}

	// Clean up debug files
	os.Remove(d.compiledFile)
	os.Remove(d.mapFile)

	return nil
}

// formatRemoteObject converts a RemoteObject to a display string.
func formatRemoteObject(obj RemoteObject) string {
	switch obj.Type {
	case "string":
		if s, ok := obj.Value.(string); ok {
			return fmt.Sprintf("%q", s)
		}
		return fmt.Sprintf("%q", obj.Description)
	case "number", "boolean":
		if obj.Description != "" {
			return obj.Description
		}
		return fmt.Sprintf("%v", obj.Value)
	case "undefined":
		return "undefined"
	case "object":
		if obj.Subtype == "null" {
			return "nothing"
		}
		if obj.Description != "" {
			return obj.Description
		}
		return "[object]"
	case "function":
		if obj.Description != "" {
			return obj.Description
		}
		return "[function]"
	default:
		if obj.Description != "" {
			return obj.Description
		}
		return fmt.Sprintf("%v", obj.Value)
	}
}

// parseInspectorURL reads Node's stderr to extract the WebSocket inspector URL.
func parseInspectorURL(stderr io.ReadCloser) (string, error) {
	scanner := bufio.NewScanner(stderr)

	wsRe := regexp.MustCompile(`ws://[^\s]+`)
	deadline := time.After(10 * time.Second)

	ch := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if matches := wsRe.FindString(line); matches != "" {
				ch <- matches
				return
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- err
		} else {
			errCh <- fmt.Errorf("Node.js stderr closed without providing WebSocket URL")
		}
	}()

	select {
	case url := <-ch:
		return url, nil
	case err := <-errCh:
		return "", err
	case <-deadline:
		return "", fmt.Errorf("timeout waiting for Node.js inspector URL")
	}
}
