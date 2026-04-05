package codegen

import (
	"quill/ast"
	"strings"
	"testing"
)

func TestWebSocketBlockGeneratesWSServer(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.WebSocketBlock{
				Path:       "/chat",
				ConnectVar: "client",
				MessageVar: "client",
				DataVar:    "data",
				CloseVar:   "client",
				OnConnect: []ast.Statement{
					&ast.SayStatement{Value: &ast.StringLiteral{Value: "Client joined"}, Line: 3},
				},
				OnMessage: []ast.Statement{
					&ast.SayStatement{Value: &ast.StringLiteral{Value: "got message"}, Line: 5},
				},
				OnClose: []ast.Statement{
					&ast.SayStatement{Value: &ast.StringLiteral{Value: "Client left"}, Line: 7},
				},
				Line: 1,
			},
		},
	}

	gen := New()
	output := gen.Generate(program)

	if !strings.Contains(output, "WebSocket.Server") {
		t.Errorf("expected WebSocket.Server creation, got:\n%s", output)
	}
	if !strings.Contains(output, "require('ws')") {
		t.Errorf("expected ws require, got:\n%s", output)
	}
}

func TestWebSocketConnectHandler(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.WebSocketBlock{
				Path:       "/chat",
				ConnectVar: "client",
				MessageVar: "client",
				DataVar:    "data",
				CloseVar:   "client",
				OnConnect: []ast.Statement{
					&ast.SayStatement{Value: &ast.StringLiteral{Value: "Client joined"}, Line: 3},
				},
				OnMessage:  []ast.Statement{},
				OnClose:    []ast.Statement{},
				Line:       1,
			},
		},
	}

	gen := New()
	output := gen.Generate(program)

	if !strings.Contains(output, "__wss.on('connection'") {
		t.Errorf("expected connection handler, got:\n%s", output)
	}
	if !strings.Contains(output, "__ws_clients.add(client)") {
		t.Errorf("expected client tracking, got:\n%s", output)
	}
	if !strings.Contains(output, `"Client joined"`) {
		t.Errorf("expected connect body, got:\n%s", output)
	}
}

func TestWebSocketMessageHandler(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.WebSocketBlock{
				Path:       "/chat",
				ConnectVar: "client",
				MessageVar: "client",
				DataVar:    "data",
				CloseVar:   "client",
				OnConnect:  []ast.Statement{},
				OnMessage: []ast.Statement{
					&ast.SayStatement{Value: &ast.Identifier{Name: "data"}, Line: 5},
				},
				OnClose: []ast.Statement{},
				Line:    1,
			},
		},
	}

	gen := New()
	output := gen.Generate(program)

	if !strings.Contains(output, "client.on('message'") {
		t.Errorf("expected message handler, got:\n%s", output)
	}
}

func TestWebSocketDisconnectHandler(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.WebSocketBlock{
				Path:       "/chat",
				ConnectVar: "client",
				MessageVar: "client",
				DataVar:    "data",
				CloseVar:   "client",
				OnConnect:  []ast.Statement{},
				OnMessage:  []ast.Statement{},
				OnClose: []ast.Statement{
					&ast.SayStatement{Value: &ast.StringLiteral{Value: "Client left"}, Line: 7},
				},
				Line: 1,
			},
		},
	}

	gen := New()
	output := gen.Generate(program)

	if !strings.Contains(output, "client.on('close'") {
		t.Errorf("expected close handler, got:\n%s", output)
	}
	if !strings.Contains(output, "__ws_clients.delete(client)") {
		t.Errorf("expected client removal on close, got:\n%s", output)
	}
}

func TestBroadcastGeneratesClientIteration(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.BroadcastStatement{
				Value: &ast.Identifier{Name: "data"},
				Line:  1,
			},
		},
	}

	gen := New()
	output := gen.Generate(program)

	if !strings.Contains(output, "__ws_clients") {
		t.Errorf("expected __ws_clients iteration, got:\n%s", output)
	}
	if !strings.Contains(output, "__c.send(data)") {
		t.Errorf("expected send to each client, got:\n%s", output)
	}
}

func TestWebSocketPathMatching(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.WebSocketBlock{
				Path:       "/notifications",
				ConnectVar: "ws",
				MessageVar: "ws",
				DataVar:    "msg",
				CloseVar:   "ws",
				OnConnect: []ast.Statement{
					&ast.SayStatement{Value: &ast.StringLiteral{Value: "connected"}, Line: 2},
				},
				OnMessage:  []ast.Statement{},
				OnClose:    []ast.Statement{},
				Line:       1,
			},
		},
	}

	gen := New()
	output := gen.Generate(program)

	if !strings.Contains(output, "request.url === '/notifications'") {
		t.Errorf("expected path matching for /notifications, got:\n%s", output)
	}
}
