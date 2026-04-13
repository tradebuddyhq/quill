package parser

import (
	"quill/ast"
	"quill/lexer"
	"strconv"
)

func (p *Parser) isOnStatement() bool {
	if !p.check(lexer.TOKEN_IDENT) {
		return false
	}
	// Look ahead past possible dot chains to find "on"
	offset := 1
	for p.pos+offset+1 < len(p.tokens) &&
		p.tokens[p.pos+offset].Type == lexer.TOKEN_DOT &&
		(p.tokens[p.pos+offset+1].Type == lexer.TOKEN_IDENT || isKeywordToken(p.tokens[p.pos+offset+1].Type)) {
		offset += 2
	}
	return p.pos+offset < len(p.tokens) && p.tokens[p.pos+offset].Type == lexer.TOKEN_ON
}

// parseOnStatement parses:
//   object on "event" with [params]: <block>            (event handler)
//   object on get|post|put|delete|patch|use "/" with [params]: <block>  (route handler)
func (p *Parser) parseOnStatement() *ast.OnStatement {
	line := p.current().Line
	// Parse the object expression (ident or ident.field.field...)
	objExpr := ast.Expression(&ast.Identifier{Name: p.advance().Value})
	for p.check(lexer.TOKEN_DOT) {
		p.advance() // consume "."
		field := p.expectIdentOrKeyword()
		objExpr = &ast.DotExpr{Object: objExpr, Field: field.Value}
	}
	p.expect(lexer.TOKEN_ON) // consume "on"

	// Check if this is a route handler: on get|post|put|delete|patch "path"
	httpMethods := map[string]bool{"get": true, "post": true, "put": true, "delete": true, "patch": true}
	if (p.check(lexer.TOKEN_IDENT) || p.check(lexer.TOKEN_DELETE)) && httpMethods[p.current().Value] {
		method := p.advance().Value
		path := p.expect(lexer.TOKEN_STRING).Value
		params := []string{}
		if p.check(lexer.TOKEN_WITH) {
			p.advance() // consume "with"
			for p.check(lexer.TOKEN_IDENT) {
				params = append(params, p.advance().Value)
				if p.check(lexer.TOKEN_COMMA) {
					p.advance()
				}
			}
		}
		p.expect(lexer.TOKEN_COLON)
		p.expect(lexer.TOKEN_NEWLINE)
		body := p.parseBlock()
		return &ast.OnStatement{Object: objExpr, Method: method, Path: path, Params: params, Body: body, Line: line}
	}

	event := p.expect(lexer.TOKEN_STRING).Value
	p.expect(lexer.TOKEN_WITH) // consume "with"
	params := []string{}
	for p.check(lexer.TOKEN_IDENT) {
		params = append(params, p.advance().Value)
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
	}
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()
	return &ast.OnStatement{Object: objExpr, Event: event, Params: params, Body: body, Line: line}
}

func (p *Parser) parseExprStatement() *ast.ExprStatement {
	line := p.current().Line
	expr := p.parseExpression()
	p.consumeNewline()
	return &ast.ExprStatement{Expr: expr, Line: line}
}


func (p *Parser) peek() lexer.Token {
	return p.current()
}

func (p *Parser) checkNextValue(types ...lexer.TokenType) bool {
	if p.pos+1 >= len(p.tokens) {
		return false
	}
	next := p.tokens[p.pos+1].Type
	for _, t := range types {
		if next == t {
			return true
		}
	}
	return false
}

func (p *Parser) parseServerBlock() ast.Statement {
	line := p.current().Line
	p.advance() // consume "server"
	p.expect(lexer.TOKEN_COLON)
	p.consumeNewline()

	server := &ast.ServerBlockStatement{
		Port: 3000, // default
		Line: line,
	}

	p.expect(lexer.TOKEN_INDENT)
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}

		if p.check(lexer.TOKEN_PORT) || (p.check(lexer.TOKEN_IDENT) && p.current().Value == "port") {
			p.advance() // consume "port"
			p.expect(lexer.TOKEN_IS)
			tok := p.expect(lexer.TOKEN_NUMBER)
			port, _ := strconv.Atoi(tok.Value)
			server.Port = port
			p.consumeNewline()
		} else if p.check(lexer.TOKEN_ROUTE) {
			route := p.parseRouteDefinition()
			server.Routes = append(server.Routes, route)
		} else if p.check(lexer.TOKEN_WEBSOCKET) {
			ws := p.parseWebSocketBlock()
			server.WebSockets = append(server.WebSockets, ws)
		} else {
			p.advance() // skip unknown
			p.consumeNewline()
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return server
}

func (p *Parser) parseRouteDefinition() ast.RouteDefinition {
	line := p.current().Line
	p.advance() // consume "route"

	// method: get, post, put, delete
	method := p.current().Value
	p.advance()

	// path string
	path := p.expect(lexer.TOKEN_STRING).Value

	p.expect(lexer.TOKEN_COLON)
	p.consumeNewline()

	body := p.parseBlock()

	return ast.RouteDefinition{
		Method: method,
		Path:   path,
		Body:   body,
		Line:   line,
	}
}

func (p *Parser) parseDatabaseBlock() ast.Statement {
	line := p.current().Line
	p.advance() // consume "database"
	p.expect(lexer.TOKEN_COLON)
	p.consumeNewline()

	db := &ast.DatabaseBlockStatement{
		Line: line,
	}

	p.expect(lexer.TOKEN_INDENT)
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}

		if p.check(lexer.TOKEN_CONNECT) || (p.check(lexer.TOKEN_IDENT) && p.current().Value == "connect") {
			p.advance() // consume "connect"
			db.ConnectString = p.expect(lexer.TOKEN_STRING).Value
			p.consumeNewline()
		} else if p.check(lexer.TOKEN_MODEL) || (p.check(lexer.TOKEN_IDENT) && p.current().Value == "model") {
			model := p.parseModelDefinition()
			db.Models = append(db.Models, model)
		} else {
			p.advance()
			p.consumeNewline()
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return db
}

func (p *Parser) parseModelDefinition() ast.ModelDef {
	p.advance() // consume "model"

	name := p.expect(lexer.TOKEN_IDENT).Value
	p.expect(lexer.TOKEN_COLON)
	p.consumeNewline()

	model := ast.ModelDef{Name: name}

	p.expect(lexer.TOKEN_INDENT)
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}

		fieldName := p.current().Value
		p.advance() // field name
		p.expect(lexer.TOKEN_AS)
		fieldType := p.current().Value
		p.advance() // type name
		p.consumeNewline()

		model.Fields = append(model.Fields, ast.ModelFieldDef{
			Name: fieldName,
			Type: fieldType,
		})
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return model
}

func (p *Parser) parseAuthBlock() ast.Statement {
	line := p.current().Line
	p.advance() // consume "auth"
	p.expect(lexer.TOKEN_COLON)
	p.consumeNewline()

	auth := &ast.AuthBlockStatement{
		Line: line,
	}

	p.expect(lexer.TOKEN_INDENT)
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}

		if p.check(lexer.TOKEN_IDENT) && p.current().Value == "secret" {
			p.advance() // consume "secret"
			p.expect(lexer.TOKEN_IS)
			auth.Secret = p.expect(lexer.TOKEN_STRING).Value
			p.consumeNewline()
		} else if p.check(lexer.TOKEN_ROUTE) {
			route := p.parseRouteDefinition()
			auth.Routes = append(auth.Routes, route)
		} else {
			p.advance()
			p.consumeNewline()
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return auth
}

func (p *Parser) parseRespond() ast.Statement {
	line := p.current().Line
	p.advance() // consume "respond"

	// Check for optional content type modifier: json, html
	contentType := ""
	if p.check(lexer.TOKEN_IDENT) && (p.current().Value == "json" || p.current().Value == "html") {
		contentType = p.advance().Value
	}

	// Support legacy "respond with <expr>" syntax
	if p.check(lexer.TOKEN_WITH) {
		p.advance() // consume "with"
	}

	value := p.parseExpression()

	statusCode := 0
	if p.check(lexer.TOKEN_STATUS) || (p.check(lexer.TOKEN_IDENT) && p.current().Value == "status") {
		p.advance() // consume "status"
		tok := p.expect(lexer.TOKEN_NUMBER)
		code, _ := strconv.Atoi(tok.Value)
		statusCode = code
	}

	p.consumeNewline()

	return &ast.RespondStatement{
		Value:       value,
		ContentType: contentType,
		StatusCode:  statusCode,
		Line:        line,
	}
}

