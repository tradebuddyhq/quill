package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"quill/analyzer"
	"quill/ast"
	"quill/lexer"
	"quill/parser"
	"quill/typechecker"
)

// Server is the Quill LSP server.
type Server struct {
	reader     *bufio.Reader
	writer     io.Writer
	docs       *DocumentStore
	hover      *HoverProvider
	completion *CompletionProvider
	shutdown   bool

	// Cache parsed programs per URI
	programs map[string]*ast.Program
}

// Start launches the LSP server on stdio.
func Start() {
	s := &Server{
		reader:     bufio.NewReader(os.Stdin),
		writer:     os.Stdout,
		docs:       NewDocumentStore(),
		hover:      NewHoverProvider(),
		completion: NewCompletionProvider(),
		programs:   make(map[string]*ast.Program),
	}
	s.run()
}

func (s *Server) run() {
	for {
		msg, err := ReadMessage(s.reader)
		if err != nil {
			// EOF or broken pipe — exit cleanly
			return
		}
		s.handleMessage(msg)
	}
}

func (s *Server) handleMessage(msg *JSONRPCMessage) {
	switch msg.Method {
	case "initialize":
		s.handleInitialize(msg)
	case "initialized":
		// No action needed
	case "shutdown":
		s.handleShutdown(msg)
	case "exit":
		s.handleExit()
	case "textDocument/didOpen":
		s.handleDidOpen(msg)
	case "textDocument/didChange":
		s.handleDidChange(msg)
	case "textDocument/didSave":
		s.handleDidSave(msg)
	case "textDocument/didClose":
		s.handleDidClose(msg)
	case "textDocument/hover":
		s.handleHover(msg)
	case "textDocument/completion":
		s.handleCompletion(msg)
	default:
		// Unknown method — if it has an ID, respond with method not found
		if msg.ID != nil {
			resp := MakeErrorResponse(msg.ID, -32601, "method not found: "+msg.Method)
			WriteMessage(s.writer, resp)
		}
	}
}

func (s *Server) handleInitialize(msg *JSONRPCMessage) {
	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: 1, // Full sync
			HoverProvider:    true,
			CompletionProvider: &CompletionOpt{
				TriggerCharacters: []string{".", " "},
				ResolveProvider:   false,
			},
		},
	}
	resp := MakeResponse(msg.ID, result)
	WriteMessage(s.writer, resp)
}

func (s *Server) handleShutdown(msg *JSONRPCMessage) {
	s.shutdown = true
	resp := MakeResponse(msg.ID, nil)
	WriteMessage(s.writer, resp)
}

func (s *Server) handleExit() {
	if s.shutdown {
		os.Exit(0)
	}
	os.Exit(1)
}

func (s *Server) handleDidOpen(msg *JSONRPCMessage) {
	var params DidOpenTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return
	}
	s.docs.Open(params.TextDocument.URI, params.TextDocument.Version, params.TextDocument.Text)
	s.validateAndPublish(params.TextDocument.URI)
}

func (s *Server) handleDidChange(msg *JSONRPCMessage) {
	var params DidChangeTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return
	}
	if len(params.ContentChanges) > 0 {
		// Full sync: take the last content change
		content := params.ContentChanges[len(params.ContentChanges)-1].Text
		s.docs.Update(params.TextDocument.URI, params.TextDocument.Version, content)
		s.validateAndPublish(params.TextDocument.URI)
	}
}

func (s *Server) handleDidSave(msg *JSONRPCMessage) {
	var params DidSaveTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return
	}
	s.validateAndPublish(params.TextDocument.URI)
}

func (s *Server) handleDidClose(msg *JSONRPCMessage) {
	var params struct {
		TextDocument TextDocumentIdentifier `json:"textDocument"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return
	}
	s.docs.Close(params.TextDocument.URI)
	delete(s.programs, params.TextDocument.URI)

	// Clear diagnostics
	s.publishDiagnostics(params.TextDocument.URI, []Diagnostic{})
}

func (s *Server) handleHover(msg *JSONRPCMessage) {
	var params TextDocumentPositionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		resp := MakeResponse(msg.ID, nil)
		WriteMessage(s.writer, resp)
		return
	}

	doc := s.docs.Get(params.TextDocument.URI)
	if doc == nil {
		resp := MakeResponse(msg.ID, nil)
		WriteMessage(s.writer, resp)
		return
	}

	program := s.programs[params.TextDocument.URI]
	hover := s.hover.GetHover(doc, params.Position, program)
	if hover == nil {
		resp := MakeResponse(msg.ID, nil)
		WriteMessage(s.writer, resp)
		return
	}

	resp := MakeResponse(msg.ID, hover)
	WriteMessage(s.writer, resp)
}

func (s *Server) handleCompletion(msg *JSONRPCMessage) {
	var params CompletionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		resp := MakeResponse(msg.ID, CompletionList{})
		WriteMessage(s.writer, resp)
		return
	}

	doc := s.docs.Get(params.TextDocument.URI)
	if doc == nil {
		resp := MakeResponse(msg.ID, CompletionList{})
		WriteMessage(s.writer, resp)
		return
	}

	program := s.programs[params.TextDocument.URI]
	completions := s.completion.GetCompletions(doc, params.Position, program)

	resp := MakeResponse(msg.ID, completions)
	WriteMessage(s.writer, resp)
}

// validateAndPublish runs the full compiler pipeline and publishes diagnostics.
func (s *Server) validateAndPublish(uri string) {
	doc := s.docs.Get(uri)
	if doc == nil {
		return
	}

	var diagnostics []Diagnostic

	// Lex
	l := lexer.New(doc.Content)
	tokens, err := l.Tokenize()
	if err != nil {
		diag := Diagnostic{
			Range:    doc.LineToRange(1),
			Severity: SeverityError,
			Source:   "quill-lexer",
			Message:  err.Error(),
		}
		diagnostics = append(diagnostics, diag)
		s.publishDiagnostics(uri, diagnostics)
		return
	}

	// Parse
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		diag := Diagnostic{
			Range:    doc.LineToRange(1),
			Severity: SeverityError,
			Source:   "quill-parser",
			Message:  err.Error(),
		}
		diagnostics = append(diagnostics, diag)
		s.publishDiagnostics(uri, diagnostics)
		return
	}

	// Cache the program for hover/completion
	s.programs[uri] = program

	// Analyze
	a := analyzer.New()
	analyzerDiags := a.Analyze(program)
	for _, d := range analyzerDiags {
		severity := SeverityWarning
		if d.Severity == analyzer.Error {
			severity = SeverityError
		} else if d.Severity == analyzer.Info {
			severity = SeverityInformation
		}
		msg := d.Message
		if d.Hint != "" {
			msg = fmt.Sprintf("%s (hint: %s)", d.Message, d.Hint)
		}
		diagnostics = append(diagnostics, Diagnostic{
			Range:    doc.LineToRange(d.Line),
			Severity: severity,
			Source:   "quill-analyzer",
			Message:  msg,
		})
	}

	// Type check
	tc := typechecker.New()
	typeDiags := tc.Check(program)
	for _, d := range typeDiags {
		severity := SeverityWarning
		if d.Severity == "error" {
			severity = SeverityError
		}
		diagnostics = append(diagnostics, Diagnostic{
			Range:    doc.LineToRange(d.Line),
			Severity: severity,
			Source:   "quill-typechecker",
			Message:  d.Message,
		})
	}

	s.publishDiagnostics(uri, diagnostics)
}

func (s *Server) publishDiagnostics(uri string, diagnostics []Diagnostic) {
	if diagnostics == nil {
		diagnostics = []Diagnostic{}
	}
	params := PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	}
	notification := MakeNotification("textDocument/publishDiagnostics", params)
	WriteMessage(s.writer, notification)
}
