package parser

import (
	"quill/ast"
	"quill/lexer"
	"fmt"
	"strconv"
)

func (p *Parser) parseWorkerHandler() *ast.WorkerHandler {
	line := p.current().Line
	p.advance() // consume "worker"
	p.expect(lexer.TOKEN_ON) // consume "on"

	// Expect "fetch" as identifier
	fetchTok := p.expect(lexer.TOKEN_IDENT)
	if fetchTok.Value != "fetch" {
		panic(&ParseError{Line: line, Message: fmt.Sprintf("expected 'fetch' after 'worker on', got '%s'", fetchTok.Value)})
	}

	var params []string
	if p.check(lexer.TOKEN_WITH) {
		p.advance() // consume "with"
		for p.check(lexer.TOKEN_IDENT) {
			params = append(params, p.advance().Value)
		}
	}

	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()

	return &ast.WorkerHandler{
		Params: params,
		Body:   body,
		Line:   line,
	}
}

// --- Utility functions ---


func (p *Parser) parseMatchObjectPattern() *ast.ObjectMatchPattern {
	p.advance() // consume "{"
	var fields []ast.ObjectMatchField

	for !p.check(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		key := p.expect(lexer.TOKEN_IDENT).Value
		if p.check(lexer.TOKEN_COLON) {
			p.advance() // consume ":"
			value := p.parseExpression()
			fields = append(fields, ast.ObjectMatchField{Key: key, Value: value})
		} else {
			// Just a binding: when {body}: means bind __match.body to body
			fields = append(fields, ast.ObjectMatchField{Key: key, Value: nil})
		}
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
	}
	p.expect(lexer.TOKEN_RBRACE)
	return &ast.ObjectMatchPattern{Fields: fields}
}

// isTypeKeyword returns true if the identifier is a type name for pattern matching.
func (p *Parser) isTypeKeyword(name string) bool {
	switch name {
	case "text", "number", "list", "boolean":
		return true
	}
	return false
}

// (duplicate stubs removed - defined earlier in file)

func (p *Parser) parseTypeAlias() *ast.TypeAliasStatement {
	line := p.current().Line
	p.advance()
	name := p.expect(lexer.TOKEN_IDENT).Value
	p.expect(lexer.TOKEN_IS)
	var utility string
	switch {
	case p.check(lexer.TOKEN_PARTIAL):
		utility = "Partial"
		p.advance()
	case p.check(lexer.TOKEN_OMIT):
		utility = "Omit"
		p.advance()
	case p.check(lexer.TOKEN_PICK):
		utility = "Pick"
		p.advance()
	case p.check(lexer.TOKEN_RECORD):
		utility = "Record"
		p.advance()
	case p.check(lexer.TOKEN_READONLY):
		utility = "Readonly"
		p.advance()
	case p.check(lexer.TOKEN_REQUIRED):
		utility = "Required"
		p.advance()
	default:
		p.error("expected type utility (Partial, Omit, Pick, Record, Readonly, Required)")
	}
	p.expect(lexer.TOKEN_OF)
	baseType := p.advance().Value
	var args []string
	if p.check(lexer.TOKEN_COMMA) {
		p.advance()
		if p.check(lexer.TOKEN_LBRACKET) {
			p.advance()
			for !p.check(lexer.TOKEN_RBRACKET) && !p.isAtEnd() {
				if p.check(lexer.TOKEN_STRING) {
					args = append(args, p.advance().Value)
				} else if p.check(lexer.TOKEN_IDENT) {
					args = append(args, p.advance().Value)
				} else {
					p.advance()
				}
				if p.check(lexer.TOKEN_COMMA) {
					p.advance()
				}
			}
			if p.check(lexer.TOKEN_RBRACKET) {
				p.advance()
			}
		} else {
			args = append(args, p.advance().Value)
		}
	}
	p.consumeNewline()
	return &ast.TypeAliasStatement{Name: name, BaseType: baseType, Utility: utility, Args: args, Line: line}
}

func (p *Parser) parseDecorators() []ast.Decorator {
	var decorators []ast.Decorator
	for p.check(lexer.TOKEN_AT) {
		line := p.current().Line
		p.advance()
		name := p.expect(lexer.TOKEN_IDENT).Value
		var args []ast.Expression
		if p.check(lexer.TOKEN_LPAREN) {
			p.advance()
			for !p.check(lexer.TOKEN_RPAREN) && !p.isAtEnd() {
				args = append(args, p.parseExpression())
				if p.check(lexer.TOKEN_COMMA) {
					p.advance()
				}
			}
			p.expect(lexer.TOKEN_RPAREN)
		}
		decorators = append(decorators, ast.Decorator{Name: name, Args: args, Line: line})
		p.consumeNewline()
		p.skipNewlines()
	}
	return decorators
}

func (p *Parser) parseDecorated() ast.Statement {
	decorators := p.parseDecorators()
	if p.check(lexer.TOKEN_TO) {
		funcDef := p.parseFuncDef()
		return &ast.DecoratedFuncDefinition{Decorators: decorators, Func: funcDef, Line: funcDef.Line}
	} else if p.check(lexer.TOKEN_ROUTE) {
		route := p.parseRouteDefinition()
		return &ast.DecoratedRouteDefinition{Decorators: decorators, Route: route, Line: route.Line}
	}
	p.error("expected function definition or route after decorator")
	return nil
}

func (p *Parser) parseBroadcast() *ast.BroadcastStatement {
	line := p.current().Line
	p.advance()
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.BroadcastStatement{Value: value, Line: line}
}

func (p *Parser) parseCommand() *ast.CommandStatement {
	line := p.current().Line
	p.advance() // consume "command"
	name := p.expect(lexer.TOKEN_STRING).Value

	var params []string
	if p.check(lexer.TOKEN_WITH) {
		p.advance() // consume "with"
		for p.check(lexer.TOKEN_IDENT) {
			params = append(params, p.advance().Value)
		}
	}

	description := ""
	if p.check(lexer.TOKEN_DESCRIBED) {
		p.advance() // consume "described"
		description = p.expect(lexer.TOKEN_STRING).Value
	}

	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()

	return &ast.CommandStatement{
		Name:        name,
		Description: description,
		Params:      params,
		Body:        body,
		Line:        line,
	}
}

func (p *Parser) parseReply() *ast.ReplyStatement {
	line := p.current().Line
	p.advance() // consume "reply"
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.ReplyStatement{Value: value, Line: line}
}

func (p *Parser) parseEmbedLiteral() *ast.EmbedLiteral {
	p.advance() // consume "embed"
	title := p.expect(lexer.TOKEN_STRING).Value
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)

	var fields []ast.EmbedField
	p.expect(lexer.TOKEN_INDENT)
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		// Embed directives use identifiers as contextual keywords
		if p.check(lexer.TOKEN_IDENT) {
			directive := p.current().Value
			switch directive {
			case "color":
				p.advance()
				colorName := p.advance().Value
				fields = append(fields, ast.EmbedField{Type: "color", Value: colorName})
				p.consumeNewline()
			case "field":
				p.advance()
				fieldName := p.expect(lexer.TOKEN_STRING).Value
				fieldValue := p.expect(lexer.TOKEN_STRING).Value
				fields = append(fields, ast.EmbedField{Type: "field", Name: fieldName, Value: fieldValue})
				p.consumeNewline()
			case "footer":
				p.advance()
				footerText := p.expect(lexer.TOKEN_STRING).Value
				fields = append(fields, ast.EmbedField{Type: "footer", Value: footerText})
				p.consumeNewline()
			case "description":
				p.advance()
				descText := p.expect(lexer.TOKEN_STRING).Value
				fields = append(fields, ast.EmbedField{Type: "description", Value: descText})
				p.consumeNewline()
			case "thumbnail":
				p.advance()
				url := p.expect(lexer.TOKEN_STRING).Value
				fields = append(fields, ast.EmbedField{Type: "thumbnail", Value: url})
				p.consumeNewline()
			case "image":
				p.advance()
				url := p.expect(lexer.TOKEN_STRING).Value
				fields = append(fields, ast.EmbedField{Type: "image", Value: url})
				p.consumeNewline()
			default:
				p.advance() // skip unknown
				p.consumeNewline()
			}
		} else {
			p.advance() // skip unknown token
			p.consumeNewline()
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
	return &ast.EmbedLiteral{Title: title, Fields: fields}
}

func (p *Parser) parseWebSocketBlock() ast.WebSocketBlock {
	line := p.current().Line
	p.advance()
	path := p.expect(lexer.TOKEN_STRING).Value
	p.expect(lexer.TOKEN_COLON)
	p.consumeNewline()
	ws := ast.WebSocketBlock{Path: path, Line: line}
	p.expect(lexer.TOKEN_INDENT)
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		if p.check(lexer.TOKEN_ON) {
			p.advance()
			eventType := p.expect(lexer.TOKEN_IDENT).Value
			switch eventType {
			case "connect":
				ws.ConnectVar = p.expect(lexer.TOKEN_IDENT).Value
				p.expect(lexer.TOKEN_COLON)
				p.consumeNewline()
				ws.OnConnect = p.parseBlock()
			case "message":
				ws.MessageVar = p.expect(lexer.TOKEN_IDENT).Value
				ws.DataVar = p.expect(lexer.TOKEN_IDENT).Value
				p.expect(lexer.TOKEN_COLON)
				p.consumeNewline()
				ws.OnMessage = p.parseBlock()
			case "disconnect":
				ws.CloseVar = p.expect(lexer.TOKEN_IDENT).Value
				p.expect(lexer.TOKEN_COLON)
				p.consumeNewline()
				ws.OnClose = p.parseBlock()
			default:
				p.error(fmt.Sprintf("unexpected websocket event %q, expected connect/message/disconnect", eventType))
			}
		} else {
			p.advance()
			p.consumeNewline()
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
	return ws
}

// isStreamStatement checks if the current position is a "stream <provider>" statement.
func (p *Parser) isStreamStatement() bool {
	if p.pos+1 >= len(p.tokens) {
		return false
	}
	cur := p.tokens[p.pos]
	next := p.tokens[p.pos+1]
	return cur.Type == lexer.TOKEN_IDENT && cur.Value == "stream" &&
		next.Type == lexer.TOKEN_IDENT &&
		(next.Value == "claude" || next.Value == "openai" || next.Value == "gemini" || next.Value == "ollama")
}

// isAgentStatement checks if the current position is an "agent" statement.
func (p *Parser) isAgentStatement() bool {
	if p.pos+1 >= len(p.tokens) {
		return false
	}
	cur := p.tokens[p.pos]
	next := p.tokens[p.pos+1]
	return cur.Type == lexer.TOKEN_IDENT && cur.Value == "agent" &&
		next.Type == lexer.TOKEN_STRING
}

// parseAskExpression parses: ask claude "prompt" [with model "x" max_tokens N system "y" temperature N]
// or: ask claude <variable> (messages mode)
func (p *Parser) parseAskExpression() *ast.AskExpression {
	p.advance() // consume "ask"

	// Expect provider name (currently only "claude")
	provider := p.expect(lexer.TOKEN_IDENT).Value

	// Parse the prompt or messages variable
	var prompt ast.Expression
	isMessages := false

	if p.check(lexer.TOKEN_STRING) {
		prompt = &ast.StringLiteral{Value: p.advance().Value}
	} else if p.check(lexer.TOKEN_IDENT) {
		// Variable reference — treat as messages array
		prompt = &ast.Identifier{Name: p.advance().Value}
		isMessages = true
	} else {
		p.error("expected a string prompt or variable after 'ask claude'")
	}

	// Parse optional "with" options
	options := map[string]ast.Expression{}
	if p.check(lexer.TOKEN_WITH) {
		p.advance() // consume "with"
		for {
			optName := ""
			if p.check(lexer.TOKEN_IDENT) {
				optName = p.advance().Value
			} else if p.check(lexer.TOKEN_MODEL) {
				optName = p.advance().Value
			} else {
				break
			}
			switch optName {
			case "model":
				options["model"] = &ast.StringLiteral{Value: p.expect(lexer.TOKEN_STRING).Value}
			case "max_tokens":
				val, err := strconv.ParseFloat(p.expect(lexer.TOKEN_NUMBER).Value, 64)
				if err != nil {
					p.error("expected a number for max_tokens")
				}
				options["max_tokens"] = &ast.NumberLiteral{Value: val}
			case "system":
				options["system"] = &ast.StringLiteral{Value: p.expect(lexer.TOKEN_STRING).Value}
			case "temperature":
				val, err := strconv.ParseFloat(p.expect(lexer.TOKEN_NUMBER).Value, 64)
				if err != nil {
					p.error("expected a number for temperature")
				}
				options["temperature"] = &ast.NumberLiteral{Value: val}
			default:
				// Unknown option — stop parsing options
				p.pos-- // back up
				goto doneOptions
			}
		}
	}
doneOptions:

	// Parse optional "as {schema}" for structured output
	var structuredOutput map[string]string
	if p.check(lexer.TOKEN_AS) || (p.check(lexer.TOKEN_IDENT) && p.current().Value == "as") {
		p.advance() // consume "as"
		p.expect(lexer.TOKEN_LBRACE)
		structuredOutput = map[string]string{}
		for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
			fieldName := p.expect(lexer.TOKEN_IDENT).Value
			p.expect(lexer.TOKEN_COLON)
			fieldType := p.expect(lexer.TOKEN_IDENT).Value
			switch fieldType {
			case "text", "number", "bool", "list":
				// valid
			default:
				p.error("unknown structured output type '" + fieldType + "'; expected text, number, bool, or list")
			}
			structuredOutput[fieldName] = fieldType
			// optional comma separator
			if p.check(lexer.TOKEN_COMMA) {
				p.advance()
			}
		}
		p.expect(lexer.TOKEN_RBRACE)
	}

	return &ast.AskExpression{
		Provider:         provider,
		Prompt:           prompt,
		Options:          options,
		IsMessages:       isMessages,
		StructuredOutput: structuredOutput,
	}
}

// parseStreamStatement parses: stream claude "prompt": body
func (p *Parser) parseStreamStatement() *ast.StreamStatement {
	line := p.current().Line
	p.advance() // consume "stream" (an IDENT)

	provider := p.expect(lexer.TOKEN_IDENT).Value // consume "claude"

	var prompt ast.Expression
	if p.check(lexer.TOKEN_STRING) {
		prompt = &ast.StringLiteral{Value: p.advance().Value}
	} else if p.check(lexer.TOKEN_IDENT) {
		prompt = &ast.Identifier{Name: p.advance().Value}
	} else {
		p.error("expected a string prompt or variable after 'stream claude'")
	}

	// Parse optional "with" options
	options := map[string]ast.Expression{}
	if p.check(lexer.TOKEN_WITH) {
		p.advance() // consume "with"
		for {
			optName := ""
			if p.check(lexer.TOKEN_IDENT) {
				optName = p.advance().Value
			} else if p.check(lexer.TOKEN_MODEL) {
				optName = p.advance().Value
			} else {
				break
			}
			switch optName {
			case "model":
				options["model"] = &ast.StringLiteral{Value: p.expect(lexer.TOKEN_STRING).Value}
			case "max_tokens":
				val, err := strconv.ParseFloat(p.expect(lexer.TOKEN_NUMBER).Value, 64)
				if err != nil {
					p.error("expected a number for max_tokens")
				}
				options["max_tokens"] = &ast.NumberLiteral{Value: val}
			case "system":
				options["system"] = &ast.StringLiteral{Value: p.expect(lexer.TOKEN_STRING).Value}
			case "temperature":
				val, err := strconv.ParseFloat(p.expect(lexer.TOKEN_NUMBER).Value, 64)
				if err != nil {
					p.error("expected a number for temperature")
				}
				options["temperature"] = &ast.NumberLiteral{Value: val}
			default:
				p.pos-- // back up
				goto doneStreamOptions
			}
		}
	}
doneStreamOptions:

	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()

	return &ast.StreamStatement{
		Provider: provider,
		Prompt:   prompt,
		ChunkVar: "chunk",
		Body:     body,
		Options:  options,
		Line:     line,
	}
}

// parseAgentStatement parses: agent "name" with tools [tool1, tool2, tool3]:
func (p *Parser) parseAgentStatement() *ast.AgentStatement {
	line := p.current().Line
	p.advance() // consume "agent"

	name := p.expect(lexer.TOKEN_STRING).Value

	// Parse "with tools [...]"
	var tools []string
	if p.check(lexer.TOKEN_WITH) || (p.check(lexer.TOKEN_IDENT) && p.current().Value == "with") {
		p.advance() // consume "with"
		if !p.check(lexer.TOKEN_IDENT) || p.current().Value != "tools" {
			p.error("expected 'tools' after 'with' in agent statement")
		}
		p.advance() // consume "tools"
		p.expect(lexer.TOKEN_LBRACKET)
		for !p.check(lexer.TOKEN_RBRACKET) && !p.check(lexer.TOKEN_EOF) {
			tools = append(tools, p.expect(lexer.TOKEN_IDENT).Value)
			if p.check(lexer.TOKEN_COMMA) {
				p.advance()
			}
		}
		p.expect(lexer.TOKEN_RBRACKET)
	}

	// Parse optional prompt string
	var prompt ast.Expression
	if p.check(lexer.TOKEN_STRING) {
		prompt = &ast.StringLiteral{Value: p.advance().Value}
	} else if p.check(lexer.TOKEN_IDENT) {
		prompt = &ast.Identifier{Name: p.advance().Value}
	}

	// Parse optional options (model, system, max_tokens)
	options := map[string]ast.Expression{}
	for p.check(lexer.TOKEN_IDENT) || p.check(lexer.TOKEN_MODEL) {
		optName := p.advance().Value
		switch optName {
		case "model":
			options["model"] = &ast.StringLiteral{Value: p.expect(lexer.TOKEN_STRING).Value}
		case "max_tokens":
			val, err := strconv.ParseFloat(p.expect(lexer.TOKEN_NUMBER).Value, 64)
			if err != nil {
				p.error("expected a number for max_tokens")
			}
			options["max_tokens"] = &ast.NumberLiteral{Value: val}
		case "system":
			options["system"] = &ast.StringLiteral{Value: p.expect(lexer.TOKEN_STRING).Value}
		default:
			p.pos-- // back up
			goto doneAgentOptions
		}
	}
doneAgentOptions:

	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()

	return &ast.AgentStatement{
		Name:    name,
		Tools:   tools,
		Prompt:  prompt,
		Options: options,
		Body:    body,
		Line:    line,
	}
}

// parseEmbedExpression parses: embed(text), embed(text, "openai"), embed(text, "openai", "text-embedding-3-small")
func (p *Parser) parseEmbedExpression() *ast.EmbedExpression {
	line := p.current().Line
	p.advance() // consume "embed"
	p.expect(lexer.TOKEN_LPAREN)

	text := p.parseExpression()

	provider := ""
	model := ""

	if p.check(lexer.TOKEN_COMMA) {
		p.advance()
		provider = p.expect(lexer.TOKEN_STRING).Value
	}

	if p.check(lexer.TOKEN_COMMA) {
		p.advance()
		model = p.expect(lexer.TOKEN_STRING).Value
	}

	p.expect(lexer.TOKEN_RPAREN)

	return &ast.EmbedExpression{
		Text:     text,
		Provider: provider,
		Model:    model,
		Line:     line,
	}
}
