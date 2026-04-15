package parser

import (
	"quill/ast"
	"quill/lexer"
	"strconv"
)

func (p *Parser) isBracketAssignment() bool {
	if !p.check(lexer.TOKEN_IDENT) && !p.check(lexer.TOKEN_MY) {
		return false
	}
	// Scan ahead to find LBRACKET, then matching RBRACKET, then IS/ARE
	i := p.pos + 1
	// Skip dot accesses: ident.field.field
	for i+1 < len(p.tokens) && p.tokens[i].Type == lexer.TOKEN_DOT {
		i += 2 // skip dot and field
	}
	if i >= len(p.tokens) || p.tokens[i].Type != lexer.TOKEN_LBRACKET {
		return false
	}
	// Find matching RBRACKET
	depth := 1
	i++
	for i < len(p.tokens) && depth > 0 {
		if p.tokens[i].Type == lexer.TOKEN_LBRACKET {
			depth++
		} else if p.tokens[i].Type == lexer.TOKEN_RBRACKET {
			depth--
		}
		i++
	}
	if depth != 0 || i >= len(p.tokens) {
		return false
	}
	return p.tokens[i].Type == lexer.TOKEN_IS || p.tokens[i].Type == lexer.TOKEN_ARE
}

func (p *Parser) parseBracketAssignment() ast.Statement {
	line := p.current().Line
	// Parse the object expression (identifier, possibly with dot accesses) up to the bracket
	expr := p.parsePrimary()
	// Handle dot accesses before the bracket
	for p.check(lexer.TOKEN_DOT) {
		p.advance() // consume "."
		field := p.expectIdentOrKeyword()
		expr = &ast.DotExpr{Object: expr, Field: field.Value}
	}
	// Parse the bracket access
	p.expect(lexer.TOKEN_LBRACKET)
	index := p.parseExpression()
	p.expect(lexer.TOKEN_RBRACKET)
	// Consume "is" or "are"
	p.advance()
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.IndexAssignStatement{Object: expr, Index: index, Value: value, Line: line}
}

func (p *Parser) parseSay() *ast.SayStatement {
	line := p.current().Line
	p.advance() // consume "say"
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.SayStatement{Value: value, Line: line}
}

func (p *Parser) parseAssignment() *ast.AssignStatement {
	line := p.current().Line
	name := p.advance().Value // consume identifier
	p.advance()               // consume "is" or "are"
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.AssignStatement{Name: name, Value: value, Line: line}
}

// parseKeywordAssignment parses assignment where the variable name is a keyword token
// (e.g., style is "dark", screen is "home")
func (p *Parser) parseKeywordAssignment() *ast.AssignStatement {
	line := p.current().Line
	name := p.advance().Value // consume keyword token as variable name
	p.advance()               // consume "is" or "are"
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.AssignStatement{Name: name, Value: value, Line: line}
}

func (p *Parser) parseIf() *ast.IfStatement {
	line := p.current().Line
	p.advance() // consume "if"
	condition := p.parseExpression()
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()

	var elseIfs []ast.ElseIfClause
	var elseBody []ast.Statement

	for p.check(lexer.TOKEN_OTHERWISE) {
		p.advance() // consume "otherwise"
		if p.check(lexer.TOKEN_IF) {
			p.advance() // consume "if"
			cond := p.parseExpression()
			p.expect(lexer.TOKEN_COLON)
			p.expect(lexer.TOKEN_NEWLINE)
			b := p.parseBlock()
			elseIfs = append(elseIfs, ast.ElseIfClause{Condition: cond, Body: b})
		} else {
			p.expect(lexer.TOKEN_COLON)
			p.expect(lexer.TOKEN_NEWLINE)
			elseBody = p.parseBlock()
			break
		}
	}

	return &ast.IfStatement{
		Condition: condition,
		Body:      body,
		ElseIfs:   elseIfs,
		Else:      elseBody,
		Line:      line,
	}
}

func (p *Parser) parseFor() *ast.ForEachStatement {
	line := p.current().Line
	p.advance() // consume "for"

	// Check for "for await each"
	isAsync := false
	if p.check(lexer.TOKEN_AWAIT) {
		isAsync = true
		p.advance() // consume "await"
	}

	p.expect(lexer.TOKEN_EACH)

	// Check for destructuring pattern: {name, age} or [a, b]
	var destructPattern ast.DestructurePattern
	varName := ""

	if p.check(lexer.TOKEN_LBRACE) {
		destructPattern = p.parseObjectPattern()
	} else if p.check(lexer.TOKEN_LBRACKET) {
		destructPattern = p.parseArrayPattern()
	} else {
		varTok := p.expect(lexer.TOKEN_IDENT)
		varName = varTok.Value
	}

	p.expect(lexer.TOKEN_IN)
	iterable := p.parseExpression()
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()

	return &ast.ForEachStatement{
		Variable:           varName,
		Iterable:           iterable,
		Body:               body,
		IsAsync:            isAsync,
		DestructurePattern: destructPattern,
		Line:               line,
	}
}

func (p *Parser) parseWhile() *ast.WhileStatement {
	line := p.current().Line
	p.advance() // consume "while"
	condition := p.parseExpression()
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()

	return &ast.WhileStatement{
		Condition: condition,
		Body:      body,
		Line:      line,
	}
}

func (p *Parser) parseFuncDef() *ast.FuncDefinition {
	line := p.current().Line
	p.advance() // consume "to"
	nameTok := p.expectIdentOrKeyword() // accept keyword tokens as function names (e.g., "load")

	params := []string{}
	paramTypes := []string{}
	for p.check(lexer.TOKEN_IDENT) {
		params = append(params, p.advance().Value)
		// Capture optional type annotation: "as type" or "as list of type"
		typeHint := ""
		if p.check(lexer.TOKEN_AS) {
			p.advance() // consume "as"
			typeHint = p.advance().Value // consume the type name
			// Handle "list of X" or "map of X"
			if p.check(lexer.TOKEN_OF) {
				p.advance() // consume "of"
				typeHint = typeHint + " of " + p.advance().Value
			}
		}
		paramTypes = append(paramTypes, typeHint)
		// Skip comma between params
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
	}

	// Parse optional where clauses for generic constraints: "where T is Comparable"
	// These come before the return type arrow
	var typeParams []ast.TypeParam
	for p.check(lexer.TOKEN_WHERE) {
		p.advance() // consume "where"
		typeParamName := p.expect(lexer.TOKEN_IDENT).Value
		p.expect(lexer.TOKEN_IS)
		constraintName := p.expect(lexer.TOKEN_IDENT).Value
		typeParams = append(typeParams, ast.TypeParam{Name: typeParamName, Constraint: constraintName})
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
	}

	// Capture optional return type: "-> type"
	returnType := ""
	if p.check(lexer.TOKEN_ARROW) {
		p.advance() // consume "->"
		returnType = p.advance().Value
		if p.check(lexer.TOKEN_OF) {
			p.advance()
			returnType = returnType + " of " + p.advance().Value
		}
	}

	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()

	return &ast.FuncDefinition{
		Name:       nameTok.Value,
		Params:     params,
		ParamTypes: paramTypes,
		ReturnType: returnType,
		TypeParams: typeParams,
		Body:       body,
		Line:       line,
	}
}

func (p *Parser) parseReturn() *ast.ReturnStatement {
	line := p.current().Line
	p.advance() // consume "give"
	p.expect(lexer.TOKEN_BACK)
	// Allow bare "give back" without a value
	if p.check(lexer.TOKEN_NEWLINE) || p.check(lexer.TOKEN_EOF) || p.check(lexer.TOKEN_DEDENT) {
		p.consumeNewline()
		return &ast.ReturnStatement{Value: nil, Line: line}
	}
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.ReturnStatement{Value: value, Line: line}
}

func (p *Parser) parseRaise() *ast.RaiseStatement {
	line := p.current().Line
	p.advance() // consume "raise"
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.RaiseStatement{Value: value, Line: line}
}

func (p *Parser) parseUse() *ast.UseStatement {
	line := p.current().Line
	p.advance() // consume "use"
	path := p.expect(lexer.TOKEN_STRING)
	alias := ""
	if p.check(lexer.TOKEN_AS) {
		p.advance()
		aliasTok := p.expect(lexer.TOKEN_IDENT)
		alias = aliasTok.Value
	}
	p.consumeNewline()
	return &ast.UseStatement{Path: path.Value, Alias: alias, Line: line}
}

func (p *Parser) parseTest() *ast.TestBlock {
	line := p.current().Line
	p.advance() // consume "test"
	name := p.expect(lexer.TOKEN_STRING)
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()
	return &ast.TestBlock{Name: name.Value, Body: body, Line: line}
}

func (p *Parser) parseMock() *ast.MockStatement {
	line := p.current().Line
	p.advance() // consume "mock"
	funcName := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_WITH)
	var params []string
	for p.check(lexer.TOKEN_IDENT) {
		params = append(params, p.current().Value)
		p.advance()
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		} else {
			break
		}
	}
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()
	return &ast.MockStatement{FuncName: funcName.Value, Params: params, Body: body, Line: line}
}

func (p *Parser) parseExpect() *ast.ExpectStatement {
	line := p.current().Line
	p.advance() // consume "expect"

	// Check for mock assertion: expect <func> was called <n> time(s)
	if p.check(lexer.TOKEN_IDENT) && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_IDENT && p.tokens[p.pos+1].Value == "was" {
		funcName := p.current().Value
		p.advance() // consume func name
		p.advance() // consume "was"
		if p.check(lexer.TOKEN_IDENT) && p.current().Value == "called" {
			p.advance() // consume "called"
			count := 1
			if p.check(lexer.TOKEN_NUMBER) {
				n, _ := strconv.Atoi(p.current().Value)
				count = n
				p.advance() // consume number
			}
			// consume optional "time" or "times"
			if p.check(lexer.TOKEN_IDENT) && (p.current().Value == "time" || p.current().Value == "times") {
				p.advance()
			}
			p.consumeNewline()
			return &ast.ExpectStatement{
				Expr: &ast.MockAssertionExpr{FuncName: funcName, AssertType: "called", Count: count},
				Line: line,
			}
		}
	}

	expr := p.parseExpression()
	p.consumeNewline()
	return &ast.ExpectStatement{Expr: expr, Line: line}
}

func (p *Parser) parseDescribe() ast.Statement {
	line := p.current().Line
	p.advance() // consume "describe"

	// Check if this is "describe trait"
	if p.check(lexer.TOKEN_TRAIT) {
		return p.parseTrait(line)
	}

	name := p.expect(lexer.TOKEN_IDENT)

	// Check for "extends ParentClass"
	extends := ""
	if p.check(lexer.TOKEN_EXTENDS) {
		p.advance() // consume "extends"
		parentTok := p.expect(lexer.TOKEN_IDENT)
		extends = parentTok.Value
	}

	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	var props []ast.AssignStatement
	var methods []ast.FuncDefinition
	var propVisibilities []string
	var methodVisibilities []string

	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}

		// Check for visibility modifier
		visibility := ""
		if p.check(lexer.TOKEN_PRIVATE) || p.check(lexer.TOKEN_PUBLIC) {
			visibility = p.advance().Value
		}

		if p.check(lexer.TOKEN_TO) {
			method := p.parseFuncDef()
			methods = append(methods, *method)
			methodVisibilities = append(methodVisibilities, visibility)
		} else if p.check(lexer.TOKEN_IDENT) && p.checkNext(lexer.TOKEN_IS, lexer.TOKEN_ARE) {
			assign := p.parseAssignment()
			props = append(props, *assign)
			propVisibilities = append(propVisibilities, visibility)
		} else {
			p.error("expected a property or method inside describe block")
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.DescribeStatement{
		Name:                 name.Value,
		Extends:              extends,
		Properties:           props,
		Methods:              methods,
		PropertyVisibilities: propVisibilities,
		MethodVisibilities:   methodVisibilities,
		Line:                 line,
	}
}

func (p *Parser) parseTryCatch() *ast.TryCatchStatement {
	line := p.current().Line
	p.advance() // consume "try"
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	tryBody := p.parseBlock()

	// Expect "if it fails" or "if it fails error:"
	var errorVar string
	var catchBody []ast.Statement

	if p.check(lexer.TOKEN_IF) {
		p.advance() // consume "if"
		// expect "it"
		p.expect(lexer.TOKEN_IDENT) // "it"
		p.expect(lexer.TOKEN_FAILS) // "fails"

		// Optional error variable name
		if p.check(lexer.TOKEN_IDENT) {
			errorVar = p.advance().Value
		} else {
			errorVar = "error"
		}

		p.expect(lexer.TOKEN_COLON)
		p.expect(lexer.TOKEN_NEWLINE)
		catchBody = p.parseBlock()
	} else if p.check(lexer.TOKEN_FAILS) {
		// Alternative syntax: "fails with <var>:" or just "fails:"
		p.advance() // consume "fails"
		if p.check(lexer.TOKEN_WITH) {
			p.advance() // consume "with"
			if p.check(lexer.TOKEN_IDENT) {
				errorVar = p.advance().Value
			} else {
				errorVar = "error"
			}
		} else {
			errorVar = "error"
		}
		p.expect(lexer.TOKEN_COLON)
		p.expect(lexer.TOKEN_NEWLINE)
		catchBody = p.parseBlock()
	} else if p.check(lexer.TOKEN_IDENT) && p.current().Value == "catch" {
		// Alternative syntax: catch err:
		p.advance() // consume "catch"
		if p.check(lexer.TOKEN_IDENT) {
			errorVar = p.advance().Value
		} else {
			errorVar = "error"
		}
		p.expect(lexer.TOKEN_COLON)
		p.expect(lexer.TOKEN_NEWLINE)
		catchBody = p.parseBlock()
	}

	return &ast.TryCatchStatement{
		TryBody:   tryBody,
		ErrorVar:  errorVar,
		CatchBody: catchBody,
		Line:      line,
	}
}

func (p *Parser) parseDelete() *ast.DeleteStatement {
	line := p.current().Line
	p.advance() // consume "delete"
	target := p.parseExpression()
	p.consumeNewline()
	return &ast.DeleteStatement{Target: target, Line: line}
}

func (p *Parser) parseBreak() *ast.BreakStatement {
	line := p.current().Line
	p.advance() // consume "break"
	p.consumeNewline()
	return &ast.BreakStatement{Line: line}
}

func (p *Parser) parseContinue() *ast.ContinueStatement {
	line := p.current().Line
	p.advance() // consume "continue"
	p.consumeNewline()
	return &ast.ContinueStatement{Line: line}
}

func (p *Parser) parseYield() *ast.YieldStatement {
	line := p.current().Line
	p.advance() // consume "yield"
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.YieldStatement{Value: value, Line: line}
}

func (p *Parser) parseLoop() *ast.LoopStatement {
	line := p.current().Line
	p.advance() // consume "loop"
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()
	return &ast.LoopStatement{Body: body, Line: line}
}

func (p *Parser) parseMatch() *ast.MatchStatement {
	line := p.current().Line
	p.advance() // consume "match"
	value := p.parseExpression()
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	cases := []ast.MatchCase{}
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}

		if p.check(lexer.TOKEN_WHEN) {
			p.advance() // consume "when"

			// Check for type-based pattern: when text t, when number n, when list l, when nothing
			typePattern := ""
			binding := ""
			if p.check(lexer.TOKEN_NOTHING) {
				typePattern = "nothing"
				p.advance() // consume "nothing"
				// nothing has no binding
				// Optional guard
				var guard ast.Expression
				if p.check(lexer.TOKEN_IF) {
					p.advance()
					guard = p.parseExpression()
				}
				p.expect(lexer.TOKEN_COLON)
				p.expect(lexer.TOKEN_NEWLINE)
				body := p.parseBlock()
				cases = append(cases, ast.MatchCase{TypePattern: typePattern, Guard: guard, Body: body})
			} else if p.check(lexer.TOKEN_IDENT) && p.isTypeKeyword(p.current().Value) {
				typePattern = p.advance().Value // consume the type name
				// Optional binding variable
				if p.check(lexer.TOKEN_IDENT) {
					binding = p.advance().Value
				}
				// Optional guard
				var guard ast.Expression
				if p.check(lexer.TOKEN_IF) {
					p.advance()
					guard = p.parseExpression()
				}
				p.expect(lexer.TOKEN_COLON)
				p.expect(lexer.TOKEN_NEWLINE)
				body := p.parseBlock()
				cases = append(cases, ast.MatchCase{TypePattern: typePattern, Binding: binding, Guard: guard, Body: body})
			} else if p.check(lexer.TOKEN_LBRACE) {
				// Object destructuring pattern: when {status: 200, body}:
				objPattern := p.parseMatchObjectPattern()

				// Optional guard
				var guard ast.Expression
				if p.check(lexer.TOKEN_IF) {
					p.advance()
					guard = p.parseExpression()
				}

				p.expect(lexer.TOKEN_COLON)
				p.expect(lexer.TOKEN_NEWLINE)
				body := p.parseBlock()
				cases = append(cases, ast.MatchCase{Pattern: objPattern, Guard: guard, Body: body})
			} else {
				// Value-based pattern (existing behavior)
				pattern := p.parseExpression()

				// Check if this is a bare identifier binding with guard (when n if ...)
				var guard ast.Expression
				if p.check(lexer.TOKEN_IF) {
					p.advance()
					guard = p.parseExpression()
				}

				p.expect(lexer.TOKEN_COLON)
				p.expect(lexer.TOKEN_NEWLINE)
				body := p.parseBlock()
				cases = append(cases, ast.MatchCase{Pattern: pattern, Guard: guard, Body: body})
			}
		} else if p.check(lexer.TOKEN_OTHERWISE) {
			p.advance() // consume "otherwise"
			p.expect(lexer.TOKEN_COLON)
			p.expect(lexer.TOKEN_NEWLINE)
			body := p.parseBlock()
			cases = append(cases, ast.MatchCase{Pattern: nil, Body: body})
		} else {
			p.error("expected 'when' or 'otherwise' in match block")
		}
	}

	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.MatchStatement{Value: value, Cases: cases, Line: line}
}

func (p *Parser) parseDefine() *ast.DefineStatement {
	line := p.current().Line
	p.advance() // consume "define"
	name := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	variants := []ast.EnumVariant{}
	var enumMethods []ast.FuncDefinition
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		// Check for method definition inside enum
		if p.check(lexer.TOKEN_TO) {
			method := p.parseFuncDef()
			enumMethods = append(enumMethods, *method)
		} else {
			variantName := p.expect(lexer.TOKEN_IDENT)
			var variantValue ast.Expression
			fields := []string{}
			// Check for "is <value>" (enum with associated value)
			if p.check(lexer.TOKEN_IS) {
				p.advance() // consume "is"
				variantValue = p.parseExpression()
			} else {
				// Parse optional fields for algebraic data types
				for p.check(lexer.TOKEN_IDENT) {
					fields = append(fields, p.advance().Value)
				}
			}
			variants = append(variants, ast.EnumVariant{Name: variantName.Value, Fields: fields, Value: variantValue})
			p.consumeNewline()
		}
	}

	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.DefineStatement{Name: name.Value, Variants: variants, Methods: enumMethods, Line: line}
}

func (p *Parser) parseFromUse() *ast.FromUseStatement {
	line := p.current().Line
	p.advance() // consume "from"
	path := p.expect(lexer.TOKEN_STRING)
	p.expect(lexer.TOKEN_USE) // "use"

	names := []string{}
	aliases := []string{}

	name := p.expectIdentOrKeyword().Value
	names = append(names, name)
	alias := ""
	if p.check(lexer.TOKEN_AS) {
		p.advance() // consume "as"
		alias = p.expectIdentOrKeyword().Value
	}
	aliases = append(aliases, alias)

	for p.check(lexer.TOKEN_COMMA) {
		p.advance()
		name = p.expectIdentOrKeyword().Value
		names = append(names, name)
		alias = ""
		if p.check(lexer.TOKEN_AS) {
			p.advance() // consume "as"
			alias = p.expectIdentOrKeyword().Value
		}
		aliases = append(aliases, alias)
	}

	p.consumeNewline()
	return &ast.FromUseStatement{Names: names, Aliases: aliases, Path: path.Value, Line: line}
}

func (p *Parser) parseDotAssignment() ast.Statement {
	line := p.current().Line
	tok := p.advance() // consume identifier or "my"
	obj := tok.Value
	if tok.Type == lexer.TOKEN_MY {
		obj = "this"
	}
	p.advance()                // consume "."
	field := p.advance().Value // consume field name
	p.advance()                // consume "is"/"are"
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.DotAssignStatement{Object: obj, Field: field, Value: value, Line: line}
}

// isOnStatement checks if the current position is an "on" event handler pattern:
// ident on "event" with ... :  OR  ident.field on "event" with ... :
