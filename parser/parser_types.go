package parser

import (
	"quill/ast"
	"quill/lexer"
)

func (p *Parser) parseTrait(line int) *ast.TraitDeclaration {
	p.advance() // consume "trait"
	name := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	var methods []ast.TraitMethod
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		if p.check(lexer.TOKEN_TO) {
			p.advance() // consume "to"
			methodName := p.expect(lexer.TOKEN_IDENT)
			var params []ast.TypedParam
			// Parse parameters until -> or : or newline
			for p.check(lexer.TOKEN_IDENT) {
				paramName := p.advance().Value
				typeHint := ""
				if p.check(lexer.TOKEN_AS) {
					p.advance()
					typeHint = p.advance().Value
					if p.check(lexer.TOKEN_OF) {
						p.advance()
						typeHint = typeHint + " of " + p.advance().Value
					}
				}
				params = append(params, ast.TypedParam{Name: paramName, TypeHint: typeHint})
				if p.check(lexer.TOKEN_COMMA) {
					p.advance()
				}
			}
			returnType := ""
			if p.check(lexer.TOKEN_ARROW) {
				p.advance()
				retTok := p.advance()
				returnType = retTok.Value
				if p.check(lexer.TOKEN_OF) {
					p.advance()
					returnType = returnType + " of " + p.advance().Value
				}
			}
			methods = append(methods, ast.TraitMethod{
				Name:       methodName.Value,
				Params:     params,
				ReturnType: returnType,
			})
			p.consumeNewline()
		} else {
			p.error("expected 'to' method signature inside trait block")
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.TraitDeclaration{Name: name.Value, Methods: methods, Line: line}
}

// --- Destructuring parsing ---

func (p *Parser) isObjectDestructure() bool {
	// Scan ahead from { ... } to see if it ends with "is" or "are"
	depth := 0
	saved := p.pos
	for i := p.pos; i < len(p.tokens); i++ {
		tok := p.tokens[i]
		if tok.Type == lexer.TOKEN_LBRACE {
			depth++
		} else if tok.Type == lexer.TOKEN_RBRACE {
			depth--
			if depth == 0 {
				// Check if next token is "is" or "are"
				if i+1 < len(p.tokens) && (p.tokens[i+1].Type == lexer.TOKEN_IS || p.tokens[i+1].Type == lexer.TOKEN_ARE) {
					return true
				}
				return false
			}
		} else if tok.Type == lexer.TOKEN_NEWLINE || tok.Type == lexer.TOKEN_EOF {
			break
		}
	}
	_ = saved
	return false
}

func (p *Parser) isArrayDestructure() bool {
	depth := 0
	for i := p.pos; i < len(p.tokens); i++ {
		tok := p.tokens[i]
		if tok.Type == lexer.TOKEN_LBRACKET {
			depth++
		} else if tok.Type == lexer.TOKEN_RBRACKET {
			depth--
			if depth == 0 {
				if i+1 < len(p.tokens) && (p.tokens[i+1].Type == lexer.TOKEN_IS || p.tokens[i+1].Type == lexer.TOKEN_ARE) {
					return true
				}
				return false
			}
		} else if tok.Type == lexer.TOKEN_NEWLINE || tok.Type == lexer.TOKEN_EOF {
			break
		}
	}
	return false
}

func (p *Parser) parseObjectDestructure() *ast.DestructureStatement {
	line := p.current().Line
	pattern := p.parseObjectPattern()
	p.advance() // consume "is" or "are"
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.DestructureStatement{Pattern: pattern, Value: value, Line: line}
}

func (p *Parser) parseObjectPattern() *ast.ObjectPattern {
	p.advance() // consume "{"
	var fields []ast.ObjectPatternField
	rest := ""

	for !p.check(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		// Check for ...rest
		if p.check(lexer.TOKEN_SPREAD) {
			p.advance() // consume "..."
			rest = p.expect(lexer.TOKEN_IDENT).Value
			if p.check(lexer.TOKEN_COMMA) {
				p.advance()
			}
			continue
		}

		key := p.expect(lexer.TOKEN_IDENT).Value

		// Check for nested pattern: key: {nested} or key: [nested]
		if p.check(lexer.TOKEN_COLON) {
			p.advance() // consume ":"
			if p.check(lexer.TOKEN_LBRACE) {
				nested := p.parseObjectPattern()
				fields = append(fields, ast.ObjectPatternField{Key: key, Nested: nested})
			} else if p.check(lexer.TOKEN_LBRACKET) {
				nested := p.parseArrayPattern()
				fields = append(fields, ast.ObjectPatternField{Key: key, Nested: nested})
			} else {
				p.error("expected nested destructuring pattern after ':'")
			}
		} else {
			fields = append(fields, ast.ObjectPatternField{Key: key})
		}

		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
	}
	p.expect(lexer.TOKEN_RBRACE)
	return &ast.ObjectPattern{Fields: fields, Rest: rest}
}

func (p *Parser) parseArrayDestructure() *ast.DestructureStatement {
	line := p.current().Line
	pattern := p.parseArrayPattern()
	p.advance() // consume "is" or "are"
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.DestructureStatement{Pattern: pattern, Value: value, Line: line}
}

func (p *Parser) parseArrayPattern() *ast.ArrayPattern {
	p.advance() // consume "["
	var elements []ast.ArrayPatternElement
	rest := ""

	for !p.check(lexer.TOKEN_RBRACKET) && !p.isAtEnd() {
		if p.check(lexer.TOKEN_SPREAD) {
			p.advance() // consume "..."
			rest = p.expect(lexer.TOKEN_IDENT).Value
			if p.check(lexer.TOKEN_COMMA) {
				p.advance()
			}
			continue
		}

		if p.check(lexer.TOKEN_LBRACE) {
			nested := p.parseObjectPattern()
			elements = append(elements, ast.ArrayPatternElement{Nested: nested})
		} else if p.check(lexer.TOKEN_LBRACKET) {
			nested := p.parseArrayPattern()
			elements = append(elements, ast.ArrayPatternElement{Nested: nested})
		} else {
			name := p.expect(lexer.TOKEN_IDENT).Value
			elements = append(elements, ast.ArrayPatternElement{Name: name})
		}

		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
	}
	p.expect(lexer.TOKEN_RBRACKET)
	return &ast.ArrayPattern{Elements: elements, Rest: rest}
}

