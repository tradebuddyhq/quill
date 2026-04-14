//go:build js && wasm

package main

import (
	"quill/codegen"
	"quill/lexer"
	"quill/parser"
	"syscall/js"
)

// compile takes Quill source code and returns compiled JavaScript.
func compile(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{"error": "no source provided"}
	}

	source := args[0].String()

	// Lex
	l := lexer.New(source)
	tokens, err := l.Tokenize()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	// Parse
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	// Generate JavaScript
	g := codegen.NewBrowser()
	js_code := g.Generate(program)

	return map[string]interface{}{"js": js_code}
}

func main() {
	// Expose compile function to JavaScript
	js.Global().Get("window").Set("__quill_compile", js.FuncOf(compile))

	// Signal that WASM is ready
	js.Global().Get("window").Set("__quill_ready", true)

	// Keep the Go program running
	select {}
}
