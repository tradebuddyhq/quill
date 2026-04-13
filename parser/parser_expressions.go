package parser

import (
	"fmt"
	"quill/ast"
	"quill/lexer"
	"strconv"
)

// --- Expression parsing (precedence climbing) ---

func (p *Parser) parseExpression() ast.Expression {
	return p.parsePipe()
}

func (p *Parser) parsePipe() ast.Expression {
	left := p.parseOr()
	for p.check(lexer.TOKEN_PIPE) {
		p.advance() // consume "|"
		right := p.parseOr()
		left = &ast.PipeExpr{Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseOr() ast.Expression {
	left := p.parseAnd()
	for p.check(lexer.TOKEN_OR) {
		p.advance()
		right := p.parseAnd()
		left = &ast.LogicalExpr{Left: left, Operator: "or", Right: right}
	}
	return left
}

func (p *Parser) parseAnd() ast.Expression {
	left := p.parseNot()
	for p.check(lexer.TOKEN_AND) {
		p.advance()
		right := p.parseNot()
		left = &ast.LogicalExpr{Left: left, Operator: "and", Right: right}
	}
	return left
}

func (p *Parser) parseNot() ast.Expression {
	if p.check(lexer.TOKEN_NOT) {
		p.advance()
		operand := p.parseNot()
		return &ast.NotExpr{Operand: operand}
	}
	return p.parseComparison()
}

func (p *Parser) parseComparison() ast.Expression {
	left := p.parseAddition()

	if p.check(lexer.TOKEN_IS) {
		p.advance() // consume "is"

		if p.check(lexer.TOKEN_GREATER) {
			p.advance() // consume "greater"
			if p.check(lexer.TOKEN_THAN) {
				p.advance() // consume "than"
			}
			if p.check(lexer.TOKEN_OR) {
				p.advance() // "or"
				p.expect(lexer.TOKEN_EQUAL)
				p.expect(lexer.TOKEN_TO)
				right := p.parseAddition()
				return &ast.ComparisonExpr{Left: left, Operator: ">=", Right: right}
			}
			right := p.parseAddition()
			return &ast.ComparisonExpr{Left: left, Operator: ">", Right: right}
		}

		if p.check(lexer.TOKEN_LESS) {
			p.advance() // consume "less"
			if p.check(lexer.TOKEN_THAN) {
				p.advance() // consume "than"
			}
			if p.check(lexer.TOKEN_OR) {
				p.advance() // "or"
				p.expect(lexer.TOKEN_EQUAL)
				p.expect(lexer.TOKEN_TO)
				right := p.parseAddition()
				return &ast.ComparisonExpr{Left: left, Operator: "<=", Right: right}
			}
			right := p.parseAddition()
			return &ast.ComparisonExpr{Left: left, Operator: "<", Right: right}
		}

		if p.check(lexer.TOKEN_EQUAL) {
			p.advance() // consume "equal"
			if p.check(lexer.TOKEN_TO) {
				p.advance() // consume "to"
			}
			right := p.parseAddition()
			return &ast.ComparisonExpr{Left: left, Operator: "==", Right: right}
		}

		if p.check(lexer.TOKEN_NOT) {
			p.advance() // consume "not"
			right := p.parseAddition()
			return &ast.ComparisonExpr{Left: left, Operator: "!=", Right: right}
		}

		// Check for type check: "is text", "is number", "is nothing", "is boolean"
		if p.check(lexer.TOKEN_IDENT) {
			typeName := p.current().Value
			if typeName == "text" || typeName == "number" || typeName == "boolean" || typeName == "nothing" {
				p.advance()
				return &ast.TypeCheckExpr{Expr: left, TypeName: typeName}
			}
		}

		// "is <expr>" means equality
		right := p.parseAddition()
		return &ast.ComparisonExpr{Left: left, Operator: "==", Right: right}
	}

	if p.check(lexer.TOKEN_CONTAINS) {
		p.advance()
		right := p.parseAddition()
		return &ast.ComparisonExpr{Left: left, Operator: "contains", Right: right}
	}

	return left
}

func (p *Parser) parseAddition() ast.Expression {
	left := p.parseMultiplication()
	for p.check(lexer.TOKEN_PLUS) || p.check(lexer.TOKEN_MINUS) {
		op := p.advance().Value
		right := p.parseMultiplication()
		left = &ast.BinaryExpr{Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parseMultiplication() ast.Expression {
	left := p.parseExponentiation()
	for p.check(lexer.TOKEN_STAR) || p.check(lexer.TOKEN_SLASH) || p.check(lexer.TOKEN_MODULO) {
		op := p.advance().Value
		right := p.parseExponentiation()
		left = &ast.BinaryExpr{Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parseExponentiation() ast.Expression {
	left := p.parseUnary()
	for p.check(lexer.TOKEN_CARET) {
		p.advance() // consume "^"
		right := p.parseUnary()
		left = &ast.BinaryExpr{Left: left, Operator: "^", Right: right}
	}
	return left
}

func (p *Parser) parseUnary() ast.Expression {
	if p.check(lexer.TOKEN_AWAIT) {
		p.advance()
		// Check for "await all" or "await first"
		if p.check(lexer.TOKEN_IDENT) && (p.current().Value == "all" || p.current().Value == "first") {
			keyword := p.advance().Value
			return &ast.AwaitExpression{Target: &ast.Identifier{Name: keyword}}
		}
		expr := p.parseUnary()
		return &ast.AwaitExpr{Expr: expr}
	}
	// "try <expr>" as an expression (error wrapping, not try/catch block)
	if p.check(lexer.TOKEN_TRY) && !p.checkNext(lexer.TOKEN_COLON) {
		p.advance() // consume "try"
		expr := p.parseUnary()
		return &ast.TryExpression{Expr: expr}
	}
	if p.check(lexer.TOKEN_MINUS) {
		p.advance()
		operand := p.parseUnary()
		return &ast.UnaryMinusExpr{Operand: operand}
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() ast.Expression {
	expr := p.parsePrimary()

	for {
		if p.check(lexer.TOKEN_DOT) {
			p.advance() // consume "."
			field := p.expectIdentOrKeyword()
			expr = &ast.DotExpr{Object: expr, Field: field.Value}
		} else if p.check(lexer.TOKEN_LBRACKET) {
			p.advance() // consume "["
			index := p.parseExpression()
			p.expect(lexer.TOKEN_RBRACKET)
			expr = &ast.IndexExpr{Object: expr, Index: index}
		} else if p.check(lexer.TOKEN_LPAREN) {
			p.advance() // consume "("
			args := []ast.Expression{}
			if !p.check(lexer.TOKEN_RPAREN) {
				args = append(args, p.parseExpression())
				for p.check(lexer.TOKEN_COMMA) {
					p.advance()
					args = append(args, p.parseExpression())
				}
			}
			p.expect(lexer.TOKEN_RPAREN)
			expr = &ast.CallExpr{Function: expr, Args: args}
		} else if p.check(lexer.TOKEN_QUESTION) {
			p.advance() // consume "?"
			expr = &ast.PropagateExpr{Expr: expr}
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) parsePrimary() ast.Expression {
	tok := p.current()

	switch tok.Type {
	case lexer.TOKEN_STRING:
		p.advance()
		return &ast.StringLiteral{Value: tok.Value}

	case lexer.TOKEN_NUMBER:
		p.advance()
		val, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			p.error(fmt.Sprintf("invalid number: %s", tok.Value))
		}
		return &ast.NumberLiteral{Value: val}

	case lexer.TOKEN_YES:
		p.advance()
		return &ast.BoolLiteral{Value: true}

	case lexer.TOKEN_NO:
		p.advance()
		return &ast.BoolLiteral{Value: false}

	case lexer.TOKEN_IDENT:
		// Check for "ask <provider>" expression
		if tok.Value == "ask" && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_IDENT {
			nextVal := p.tokens[p.pos+1].Value
			if nextVal == "claude" || nextVal == "openai" || nextVal == "gemini" || nextVal == "ollama" {
				return p.parseAskExpression()
			}
		}
		// Check for "agent" statement used as expression (shouldn't happen, but handle gracefully)
		if tok.Value == "agent" && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_STRING {
			p.error("agent blocks are statements, not expressions")
		}
		p.advance()
		// Check for tagged template: identifier followed by backtick
		if p.check(lexer.TOKEN_BACKTICK) {
			return p.parseTaggedTemplate(tok.Value, tok.Line)
		}
		return &ast.Identifier{Name: tok.Value}

	case lexer.TOKEN_NOTHING:
		p.advance()
		return &ast.NothingLiteral{}

	case lexer.TOKEN_EMBED:
		// Check if this is embed(...) function call for embeddings
		if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_LPAREN {
			return p.parseEmbedExpression()
		}
		// Check if this is an embed literal: embed "title":
		if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_STRING {
			return p.parseEmbedLiteral()
		}
		// Otherwise treat as identifier
		p.advance()
		return &ast.Identifier{Name: "embed"}

	case lexer.TOKEN_WITH:
		return p.parseLambda()

	case lexer.TOKEN_LBRACE:
		return p.parseObjectLiteral()

	case lexer.TOKEN_SPREAD:
		p.advance() // consume "..."
		expr := p.parseUnary()
		return &ast.SpreadExpr{Expr: expr}

	case lexer.TOKEN_NEW:
		p.advance() // consume "new"
		className := p.expect(lexer.TOKEN_IDENT).Value
		// Support dotted class names like Discord.Client
		for p.check(lexer.TOKEN_DOT) {
			p.advance() // consume "."
			next := p.expectIdentOrKeyword()
			className = className + "." + next.Value
		}
		args := []ast.Expression{}
		if p.check(lexer.TOKEN_LPAREN) {
			p.advance()
			if !p.check(lexer.TOKEN_RPAREN) {
				args = append(args, p.parseExpression())
				for p.check(lexer.TOKEN_COMMA) {
					p.advance()
					args = append(args, p.parseExpression())
				}
			}
			p.expect(lexer.TOKEN_RPAREN)
		}
		return &ast.NewExpr{ClassName: className, Args: args}

	case lexer.TOKEN_RECEIVE:
		p.advance() // consume "receive"
		p.expect(lexer.TOKEN_FROM)
		channelTok := p.expect(lexer.TOKEN_IDENT)
		return &ast.ReceiveExpression{Channel: channelTok.Value, Line: tok.Line}

	// ask is handled via IDENT check below

	case lexer.TOKEN_MY:
		p.advance()
		p.expect(lexer.TOKEN_DOT)
		field := p.expectIdentOrKeyword()
		return &ast.DotExpr{Object: &ast.Identifier{Name: "this"}, Field: field.Value}

	case lexer.TOKEN_IF:
		return p.parseTernaryExpression()

	case lexer.TOKEN_LBRACKET:
		return p.parseListLiteral()

	case lexer.TOKEN_LPAREN:
		p.advance()
		expr := p.parseExpression()
		p.expect(lexer.TOKEN_RPAREN)
		return expr

	default:
		p.error(fmt.Sprintf("I didn't expect %q here", tok.Value))
		return nil
	}
}

// parseTernaryExpression parses: if <cond>: <then> otherwise: <else> as an expression.
func (p *Parser) parseTernaryExpression() ast.Expression {
	p.advance() // consume "if"
	condition := p.parseExpression()
	p.expect(lexer.TOKEN_COLON)
	thenExpr := p.parseExpression()
	p.expect(lexer.TOKEN_OTHERWISE)
	p.expect(lexer.TOKEN_COLON)
	elseExpr := p.parseExpression()
	return &ast.TernaryExpression{Condition: condition, Then: thenExpr, Else: elseExpr}
}

func (p *Parser) parseLambda() ast.Expression {
	p.advance() // consume "with"
	params := []string{}

	// Parse params until we hit ":"
	for p.check(lexer.TOKEN_IDENT) {
		params = append(params, p.advance().Value)
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
	}

	p.expect(lexer.TOKEN_COLON)

	// Check if the body is a statement (e.g., say "hello") rather than an expression
	if p.check(lexer.TOKEN_SAY) || p.check(lexer.TOKEN_GIVE) {
		stmt := p.parseStatement()
		return &ast.LambdaExpr{Params: params, BodyStatements: []ast.Statement{stmt}}
	}

	// Check for block body (indented)
	if p.check(lexer.TOKEN_NEWLINE) && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_INDENT {
		p.advance() // consume newline
		body := p.parseBlock()
		return &ast.LambdaExpr{Params: params, BodyStatements: body}
	}

	body := p.parseExpression()
	return &ast.LambdaExpr{Params: params, Body: body}
}

func (p *Parser) parseObjectLiteral() ast.Expression {
	p.advance() // consume "{"
	savedPos := p.pos
	p.skipNewlinesAndIndent()
	// Count how many net indents we consumed from the opening brace
	netIndents := p.countIndentsBetween(savedPos, p.pos)
	keys := []string{}
	values := []ast.Expression{}
	var computed []ast.ComputedProperty

	if !p.check(lexer.TOKEN_RBRACE) {
		// Parse first property (regular, computed, or string-keyed)
		if p.check(lexer.TOKEN_LBRACKET) {
			// Computed property: {[expr]: value}
			p.advance() // consume "["
			keyExpr := p.parseExpression()
			p.expect(lexer.TOKEN_RBRACKET)
			p.expect(lexer.TOKEN_COLON)
			val := p.parseExpression()
			computed = append(computed, ast.ComputedProperty{KeyExpr: keyExpr, Value: val})
		} else if p.check(lexer.TOKEN_STRING) {
			// String key: {"User-Agent": value}
			keyStr := p.advance().Value
			p.expect(lexer.TOKEN_COLON)
			val := p.parseExpression()
			computed = append(computed, ast.ComputedProperty{
				KeyExpr: &ast.StringLiteral{Value: keyStr},
				Value:   val,
			})
		} else {
			key := p.expectIdentOrKeyword()
			keys = append(keys, key.Value)
			p.expect(lexer.TOKEN_COLON)
			values = append(values, p.parseExpression())
		}

		for p.check(lexer.TOKEN_COMMA) {
			p.advance()
			beforeSkip := p.pos
			p.skipNewlinesAndIndent()
			netIndents += p.countIndentsBetween(beforeSkip, p.pos)
			if p.check(lexer.TOKEN_RBRACE) {
				break // trailing comma
			}
			if p.check(lexer.TOKEN_LBRACKET) {
				// Computed property
				p.advance() // consume "["
				keyExpr := p.parseExpression()
				p.expect(lexer.TOKEN_RBRACKET)
				p.expect(lexer.TOKEN_COLON)
				val := p.parseExpression()
				computed = append(computed, ast.ComputedProperty{KeyExpr: keyExpr, Value: val})
			} else if p.check(lexer.TOKEN_STRING) {
				// String key
				keyStr := p.advance().Value
				p.expect(lexer.TOKEN_COLON)
				val := p.parseExpression()
				computed = append(computed, ast.ComputedProperty{
					KeyExpr: &ast.StringLiteral{Value: keyStr},
					Value:   val,
				})
			} else {
				key := p.expectIdentOrKeyword()
				keys = append(keys, key.Value)
				p.expect(lexer.TOKEN_COLON)
				values = append(values, p.parseExpression())
			}
		}
	}

	beforeSkip := p.pos
	p.skipNewlinesAndIndent()
	netIndents += p.countIndentsBetween(beforeSkip, p.pos)
	p.expect(lexer.TOKEN_RBRACE)
	// After a multiline object literal, consume matching dedent tokens
	// for any indentation that was entered inside the braces
	for i := 0; i < netIndents; i++ {
		if p.check(lexer.TOKEN_NEWLINE) {
			p.advance()
		}
		if p.check(lexer.TOKEN_DEDENT) {
			p.advance()
		}
	}
	return &ast.ObjectLiteral{Keys: keys, Values: values, ComputedProperties: computed}
}

func (p *Parser) parseTaggedTemplate(tag string, line int) ast.Expression {
	p.advance() // consume opening TOKEN_BACKTICK
	// Next token should be the string content between backticks
	templateTok := p.expect(lexer.TOKEN_STRING)
	p.expect(lexer.TOKEN_BACKTICK) // consume closing TOKEN_BACKTICK

	template := templateTok.Value
	// Parse {expr} interpolations from the template string
	var expressions []ast.Expression
	i := 0
	for i < len(template) {
		if template[i] == '{' {
			j := i + 1
			depth := 1
			for j < len(template) && depth > 0 {
				if template[j] == '{' {
					depth++
				} else if template[j] == '}' {
					depth--
				}
				if depth > 0 {
					j++
				}
			}
			if depth == 0 {
				exprStr := template[i+1 : j]
				// Parse the expression string
				exprLexer := lexer.New(exprStr)
				exprTokens, err := exprLexer.Tokenize()
				if err == nil && len(exprTokens) > 1 {
					exprParser := New(exprTokens)
					expr := exprParser.parseExpression()
					expressions = append(expressions, expr)
				}
				i = j + 1
			} else {
				i++
			}
		} else {
			i++
		}
	}

	return &ast.TaggedTemplateExpr{
		Tag:         tag,
		Template:    template,
		Expressions: expressions,
		Line:        line,
	}
}

func (p *Parser) parseListLiteral() ast.Expression {
	p.advance() // consume "["
	savedPos := p.pos
	p.skipNewlinesAndIndent()
	netIndents := p.countIndentsBetween(savedPos, p.pos)
	elements := []ast.Expression{}

	if !p.check(lexer.TOKEN_RBRACKET) {
		elements = append(elements, p.parseExpression())
		for p.check(lexer.TOKEN_COMMA) {
			p.advance()
			beforeSkip := p.pos
			p.skipNewlinesAndIndent()
			netIndents += p.countIndentsBetween(beforeSkip, p.pos)
			if p.check(lexer.TOKEN_RBRACKET) {
				break // trailing comma
			}
			elements = append(elements, p.parseExpression())
		}
	}

	beforeSkip := p.pos
	p.skipNewlinesAndIndent()
	netIndents += p.countIndentsBetween(beforeSkip, p.pos)
	p.expect(lexer.TOKEN_RBRACKET)
	// After a multiline array literal, consume matching dedent tokens
	for i := 0; i < netIndents; i++ {
		if p.check(lexer.TOKEN_NEWLINE) {
			p.advance()
		}
		if p.check(lexer.TOKEN_DEDENT) {
			p.advance()
		}
	}
	return &ast.ListLiteral{Elements: elements}
}

// --- Helper methods ---

// --- Full-stack block parsing ---

