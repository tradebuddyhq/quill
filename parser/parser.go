package parser

import (
	"fmt"
	"quill/ast"
	"quill/lexer"
	"strconv"
)

type ParseError struct {
	Line    int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Message)
}

type Parser struct {
	tokens   []lexer.Token
	pos      int
	recovery *ErrorRecovery // nil means no recovery (panic on first error)
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) Parse() (program *ast.Program, err error) {
	defer func() {
		if r := recover(); r != nil {
			if pe, ok := r.(*ParseError); ok {
				err = pe
			} else {
				panic(r)
			}
		}
	}()

	stmts := []ast.Statement{}
	for !p.isAtEnd() {
		p.skipNewlines()
		if p.isAtEnd() {
			break
		}
		stmt := p.parseStatement()
		stmts = append(stmts, stmt)
	}

	return &ast.Program{Statements: stmts}, nil
}

// --- Statement parsing ---

func (p *Parser) parseStatement() ast.Statement {
	switch {
	case p.check(lexer.TOKEN_SAY):
		return p.parseSay()
	case p.check(lexer.TOKEN_IF):
		return p.parseIf()
	case p.check(lexer.TOKEN_FOR):
		return p.parseFor()
	case p.check(lexer.TOKEN_WHILE):
		return p.parseWhile()
	case p.check(lexer.TOKEN_TO):
		return p.parseFuncDef()
	case p.check(lexer.TOKEN_GIVE):
		return p.parseReturn()
	case p.check(lexer.TOKEN_USE):
		return p.parseUse()
	case p.check(lexer.TOKEN_TEST):
		return p.parseTest()
	case p.check(lexer.TOKEN_MOCK):
		return p.parseMock()
	case p.check(lexer.TOKEN_EXPECT):
		return p.parseExpect()
	case p.check(lexer.TOKEN_DESCRIBE):
		return p.parseDescribe()
	case p.check(lexer.TOKEN_TRY):
		return p.parseTryCatch()
	case p.check(lexer.TOKEN_BREAK):
		return p.parseBreak()
	case p.check(lexer.TOKEN_CONTINUE):
		return p.parseContinue()
	case p.check(lexer.TOKEN_FROM):
		return p.parseFromUse()
	case p.check(lexer.TOKEN_MATCH):
		return p.parseMatch()
	case p.check(lexer.TOKEN_DEFINE):
		return p.parseDefine()
	case p.check(lexer.TOKEN_CANCEL):
		return p.parseCancel()
	case p.check(lexer.TOKEN_SPAWN):
		return p.parseSpawn()
	case p.check(lexer.TOKEN_PARALLEL):
		return p.parseParallel()
	case p.check(lexer.TOKEN_RACE):
		return p.parseRace()
	case p.check(lexer.TOKEN_CHANNEL):
		return p.parseChannel()
	case p.check(lexer.TOKEN_SEND):
		return p.parseSend()
	case p.check(lexer.TOKEN_SELECT):
		return p.parseSelect()
	case p.check(lexer.TOKEN_YIELD):
		return p.parseYield()
	case p.check(lexer.TOKEN_LOOP):
		return p.parseLoop()
	case p.check(lexer.TOKEN_COMPONENT):
		return p.parseComponent()
	case p.check(lexer.TOKEN_MOUNT):
		return p.parseMount()
	case p.check(lexer.TOKEN_TYPE):
		return p.parseTypeAlias()
	case p.check(lexer.TOKEN_AT):
		return p.parseDecorated()
	case p.check(lexer.TOKEN_BROADCAST):
		return p.parseBroadcast()
	case p.check(lexer.TOKEN_COMMAND):
		return p.parseCommand()
	case p.check(lexer.TOKEN_REPLY):
		return p.parseReply()
	case p.check(lexer.TOKEN_WORKER):
		return p.parseWorkerHandler()
	case p.check(lexer.TOKEN_RESPOND):
		return p.parseRespond()
	case p.isStreamStatement():
		return p.parseStreamStatement()
	case p.check(lexer.TOKEN_LBRACE):
		// Check if this is a destructuring: {name, age} is expr
		if p.isObjectDestructure() {
			return p.parseObjectDestructure()
		}
		return p.parseExprStatement()
	case p.check(lexer.TOKEN_LBRACKET):
		// Check if this is a destructuring: [first, second] are expr
		if p.isArrayDestructure() {
			return p.parseArrayDestructure()
		}
		return p.parseExprStatement()
	case p.check(lexer.TOKEN_IDENT) && p.pos+3 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_DOT && p.tokens[p.pos+2].Type == lexer.TOKEN_IDENT && (p.tokens[p.pos+3].Type == lexer.TOKEN_IS || p.tokens[p.pos+3].Type == lexer.TOKEN_ARE):
		return p.parseDotAssignment()
	case p.check(lexer.TOKEN_IDENT) && p.checkNext(lexer.TOKEN_IS, lexer.TOKEN_ARE):
		return p.parseAssignment()
	case p.isOnStatement():
		return p.parseOnStatement()
	default:
		return p.parseExprStatement()
	}
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
	nameTok := p.expect(lexer.TOKEN_IDENT)

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
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.ReturnStatement{Value: value, Line: line}
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
	}

	return &ast.TryCatchStatement{
		TryBody:   tryBody,
		ErrorVar:  errorVar,
		CatchBody: catchBody,
		Line:      line,
	}
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
	names = append(names, p.expect(lexer.TOKEN_IDENT).Value)
	for p.check(lexer.TOKEN_COMMA) {
		p.advance()
		names = append(names, p.expect(lexer.TOKEN_IDENT).Value)
	}

	p.consumeNewline()
	return &ast.FromUseStatement{Names: names, Path: path.Value, Line: line}
}

func (p *Parser) parseDotAssignment() ast.Statement {
	line := p.current().Line
	obj := p.advance().Value  // consume identifier
	p.advance()                // consume "."
	field := p.advance().Value // consume field name
	p.advance()                // consume "is"/"are"
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.DotAssignStatement{Object: obj, Field: field, Value: value, Line: line}
}

// isOnStatement checks if the current position is an "on" event handler pattern:
// ident on "event" with ... :  OR  ident.field on "event" with ... :
func (p *Parser) isOnStatement() bool {
	if !p.check(lexer.TOKEN_IDENT) {
		return false
	}
	// Look ahead past possible dot chains to find "on"
	offset := 1
	for p.pos+offset+1 < len(p.tokens) &&
		p.tokens[p.pos+offset].Type == lexer.TOKEN_DOT &&
		p.tokens[p.pos+offset+1].Type == lexer.TOKEN_IDENT {
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
		field := p.expect(lexer.TOKEN_IDENT)
		objExpr = &ast.DotExpr{Object: objExpr, Field: field.Value}
	}
	p.expect(lexer.TOKEN_ON) // consume "on"

	// Check if this is a route handler: on get|post|put|delete|patch "path"
	httpMethods := map[string]bool{"get": true, "post": true, "put": true, "delete": true, "patch": true}
	if p.check(lexer.TOKEN_IDENT) && httpMethods[p.current().Value] {
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

func (p *Parser) parseComponent() *ast.ComponentStatement {
	line := p.current().Line
	p.advance() // consume "component"
	name := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	var states []ast.StateDeclaration
	var methods []ast.FuncDefinition
	var renderBody []ast.RenderElement
	var styles *ast.StyleBlock
	var loader *ast.LoadFunction
	var actions []ast.FormAction
	var head *ast.HeadBlock

	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		if p.check(lexer.TOKEN_STATE) {
			states = append(states, *p.parseStateDecl())
		} else if p.check(lexer.TOKEN_STYLE) {
			styles = p.parseStyleBlock()
		} else if p.check(lexer.TOKEN_HEAD) {
			head = p.parseHeadBlock()
		} else if p.check(lexer.TOKEN_TO) {
			if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Value == "render" {
				renderBody = p.parseRenderMethod()
			} else if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_LOAD {
				loader = p.parseLoadFunction()
			} else {
				method := p.parseFuncDef()
				methods = append(methods, *method)
			}
		} else if p.check(lexer.TOKEN_FORM) {
			action := p.parseFormAction()
			actions = append(actions, action)
		} else {
			p.error("expected state, style, head, method (to), form, or render inside component block")
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.ComponentStatement{
		Name:       name.Value,
		States:     states,
		Methods:    methods,
		RenderBody: renderBody,
		Styles:     styles,
		Loader:     loader,
		Actions:    actions,
		Head:       head,
		Line:       line,
	}
}

func (p *Parser) parseStyleBlock() *ast.StyleBlock {
	p.advance() // consume "style"
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)
	var rules []ast.CSSRule
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		rule := p.parseCSSRule()
		rules = append(rules, rule)
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
	return &ast.StyleBlock{Rules: rules}
}

func (p *Parser) parseCSSRule() ast.CSSRule {
	var selectorParts []string
	for !p.check(lexer.TOKEN_COLON) && !p.check(lexer.TOKEN_NEWLINE) && !p.isAtEnd() {
		tok := p.advance()
		selectorParts = append(selectorParts, tok.Value)
	}
	selector := ""
	for i, part := range selectorParts {
		if i > 0 && part != ":" && selectorParts[i-1] != "." {
			selector += " "
		}
		selector += part
	}
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)
	properties := make(map[string]string)
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		var propParts []string
		for !p.check(lexer.TOKEN_IS) && !p.check(lexer.TOKEN_NEWLINE) && !p.isAtEnd() {
			tok := p.advance()
			propParts = append(propParts, tok.Value)
		}
		propName := ""
		for i, part := range propParts {
			if i > 0 {
				propName += "-"
			}
			propName += part
		}
		if p.check(lexer.TOKEN_IS) {
			p.advance()
			val := p.expect(lexer.TOKEN_STRING)
			properties[propName] = val.Value
		}
		p.consumeNewline()
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
	return ast.CSSRule{Selector: selector, Properties: properties}
}

func (p *Parser) parseLoadFunction() *ast.LoadFunction {
	p.advance() // consume "to"
	p.advance() // consume "load"
	param := "request"
	if p.check(lexer.TOKEN_IDENT) {
		param = p.advance().Value
	}
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()
	return &ast.LoadFunction{Param: param, Body: body}
}

func (p *Parser) parseHeadBlock() *ast.HeadBlock {
	line := p.current().Line
	p.advance() // consume "head"
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)
	var entries []ast.HeadEntry
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		tagTok := p.expect(lexer.TOKEN_IDENT)
		tag := tagTok.Value
		entry := ast.HeadEntry{Tag: tag, Attrs: make(map[string]string)}
		if tag == "title" {
			text := p.expect(lexer.TOKEN_STRING)
			entry.Text = text.Value
		} else {
			for p.check(lexer.TOKEN_IDENT) {
				attrName := p.advance().Value
				if p.check(lexer.TOKEN_STRING) {
					attrVal := p.advance().Value
					entry.Attrs[attrName] = attrVal
				}
			}
		}
		p.consumeNewline()
		entries = append(entries, entry)
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
	return &ast.HeadBlock{Entries: entries, Line: line}
}

func (p *Parser) parseFormAction() ast.FormAction {
	p.advance()                                      // consume "form"
	p.expect(lexer.TOKEN_IDENT)                      // "action"
	handlerName := p.expect(lexer.TOKEN_IDENT).Value // handler name
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)
	var body []ast.Statement
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		body = append(body, p.parseStatement())
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
	return ast.FormAction{Name: handlerName, Body: body}
}

func (p *Parser) parseStateDecl() *ast.StateDeclaration {
	line := p.current().Line
	p.advance() // consume "state"
	name := p.expect(lexer.TOKEN_IDENT)
	p.advance() // consume "is" or "are"
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.StateDeclaration{Name: name.Value, Value: value, Line: line}
}

func (p *Parser) parseRenderMethod() []ast.RenderElement {
	p.advance() // consume "to"
	p.advance() // consume "render"
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	var elements []ast.RenderElement
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		elements = append(elements, *p.parseRenderElement())
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
	return elements
}

func (p *Parser) parseRenderElement() *ast.RenderElement {
	line := p.current().Line

	// Handle link element: link to="/path" "text"
	if p.check(lexer.TOKEN_LINK) {
		p.advance() // consume "link"
		props := make(map[string]ast.Expression)
		// Parse to="path"
		if p.check(lexer.TOKEN_TO) {
			p.advance() // consume "to"
			if p.check(lexer.TOKEN_STRING) {
				toVal := p.advance()
				props["to"] = &ast.StringLiteral{Value: toVal.Value}
			}
		}
		var children []ast.RenderNode
		if p.check(lexer.TOKEN_STRING) {
			text := p.advance()
			children = append(children, ast.RenderNode{Text: &ast.StringLiteral{Value: text.Value}})
		}
		p.consumeNewline()
		return &ast.RenderElement{
			Tag:      "__link",
			Props:    props,
			Children: children,
			Line:     line,
		}
	}

	// Handle conditional: if condition:
	if p.check(lexer.TOKEN_IF) {
		p.advance() // consume "if"
		condition := p.parseExpression()
		p.expect(lexer.TOKEN_COLON)
		p.expect(lexer.TOKEN_NEWLINE)
		p.expect(lexer.TOKEN_INDENT)
		var children []ast.RenderNode
		for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
			p.skipNewlines()
			if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
				break
			}
			el := p.parseRenderElement()
			children = append(children, ast.RenderNode{Element: el})
		}
		if p.check(lexer.TOKEN_DEDENT) {
			p.advance()
		}
		return &ast.RenderElement{
			Tag:       "__fragment",
			Condition: condition,
			Children:  children,
			Props:     map[string]ast.Expression{},
			Line:      line,
		}
	}

	// Handle for each: for each item in list:
	if p.check(lexer.TOKEN_FOR) {
		p.advance() // consume "for"
		p.expect(lexer.TOKEN_EACH)
		varTok := p.expect(lexer.TOKEN_IDENT)
		p.expect(lexer.TOKEN_IN)
		iterable := p.parseExpression()
		p.expect(lexer.TOKEN_COLON)
		p.expect(lexer.TOKEN_NEWLINE)
		p.expect(lexer.TOKEN_INDENT)
		var children []ast.RenderNode
		for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
			p.skipNewlines()
			if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
				break
			}
			el := p.parseRenderElement()
			children = append(children, ast.RenderNode{Element: el})
		}
		if p.check(lexer.TOKEN_DEDENT) {
			p.advance()
		}
		return &ast.RenderElement{
			Tag:      "__fragment",
			Iterator: &ast.RenderIterator{Variable: varTok.Value, Iterable: iterable},
			Children: children,
			Props:    map[string]ast.Expression{},
			Line:     line,
		}
	}

	// Regular element: tag [props...] [: "text" | NEWLINE INDENT children DEDENT]
	tag := p.expect(lexer.TOKEN_IDENT)
	props := make(map[string]ast.Expression)

	// Parse props: onClick handler, bind:value ident, key value, etc.
	for p.check(lexer.TOKEN_IDENT) && !p.isAtEnd() {
		propName := p.current().Value
		// Check for bind:value pattern
		if propName == "bind" && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_COLON {
			p.advance() // consume "bind"
			p.advance() // consume ":"
			bindTarget := p.expect(lexer.TOKEN_IDENT)
			props["bind:"+bindTarget.Value] = &ast.Identifier{Name: bindTarget.Value}
			continue
		}
		// Check if next token is an identifier (event handler or attr with value)
		if p.checkNext(lexer.TOKEN_IDENT) {
			p.advance() // consume prop name
			valTok := p.advance() // consume prop value (identifier)
			props[propName] = &ast.Identifier{Name: valTok.Value}
		} else if p.checkNext(lexer.TOKEN_STRING) {
			p.advance() // consume prop name
			valTok := p.advance() // consume string value
			props[propName] = &ast.StringLiteral{Value: valTok.Value}
		} else {
			break
		}
	}

	var children []ast.RenderNode

	if p.check(lexer.TOKEN_COLON) {
		p.advance() // consume ":"
		if p.check(lexer.TOKEN_STRING) {
			// Inline text: tag: "text"
			text := p.advance()
			children = append(children, ast.RenderNode{Text: &ast.StringLiteral{Value: text.Value}})
			p.consumeNewline()
		} else if p.check(lexer.TOKEN_IDENT) {
			// Inline expression: tag: expr
			expr := p.parseExpression()
			children = append(children, ast.RenderNode{Text: expr})
			p.consumeNewline()
		} else if p.check(lexer.TOKEN_NEWLINE) {
			// Block children
			p.advance() // consume newline
			p.expect(lexer.TOKEN_INDENT)
			for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
				p.skipNewlines()
				if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
					break
				}
				el := p.parseRenderElement()
				children = append(children, ast.RenderNode{Element: el})
			}
			if p.check(lexer.TOKEN_DEDENT) {
				p.advance()
			}
		}
	} else {
		p.consumeNewline()
	}

	return &ast.RenderElement{
		Tag:      tag.Value,
		Props:    props,
		Children: children,
		Line:     line,
	}
}

func (p *Parser) parseMount() *ast.MountStatement {
	line := p.current().Line
	p.advance() // consume "mount"
	comp := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_TO)
	selector := p.parseExpression()
	p.consumeNewline()
	return &ast.MountStatement{Component: comp.Value, Selector: selector, Line: line}
}

// --- Concurrency parsing ---

func (p *Parser) parseCancel() ast.Statement {
	line := p.current().Line
	p.advance() // consume "cancel"
	target := p.expect(lexer.TOKEN_IDENT)
	p.consumeNewline()
	return &ast.CancelStatement{Target: target.Value, Line: line}
}

func (p *Parser) parseSpawn() ast.Statement {
	line := p.current().Line
	p.advance() // consume "spawn"
	p.expect(lexer.TOKEN_TASK)
	nameTok := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()
	return &ast.SpawnStatement{Name: nameTok.Value, Body: body, Line: line}
}

func (p *Parser) parseParallel() ast.Statement {
	line := p.current().Line
	p.advance() // consume "parallel"

	// Check for "parallel settled:"
	isSettled := false
	if p.check(lexer.TOKEN_SETTLED) {
		isSettled = true
		p.advance() // consume "settled"
	}

	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()
	return &ast.ParallelBlock{Tasks: body, IsSettled: isSettled, Line: line}
}

func (p *Parser) parseRace() ast.Statement {
	line := p.current().Line
	p.advance() // consume "race"
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()
	return &ast.RaceBlock{Tasks: body, Line: line}
}

func (p *Parser) parseChannel() ast.Statement {
	line := p.current().Line
	p.advance() // consume "channel"
	nameTok := p.expect(lexer.TOKEN_IDENT)
	var bufferSize ast.Expression
	if p.check(lexer.TOKEN_WITH) {
		p.advance() // consume "with"
		p.expect(lexer.TOKEN_BUFFER)
		bufferSize = p.parseExpression()
	}
	p.consumeNewline()
	return &ast.ChannelStatement{Name: nameTok.Value, BufferSize: bufferSize, Line: line}
}

func (p *Parser) parseSend() ast.Statement {
	line := p.current().Line
	p.advance() // consume "send"
	value := p.parseExpression()
	p.expect(lexer.TOKEN_TO)
	channelTok := p.expect(lexer.TOKEN_IDENT)
	p.consumeNewline()
	return &ast.SendStatement{Value: value, Channel: channelTok.Value, Line: line}
}

func (p *Parser) parseSelect() ast.Statement {
	line := p.current().Line
	p.advance() // consume "select"
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	var cases []ast.SelectCase
	var afterMs ast.Expression
	var afterBody []ast.Statement

	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		if p.check(lexer.TOKEN_WHEN) {
			p.advance() // consume "when"
			p.expect(lexer.TOKEN_RECEIVE)
			p.expect(lexer.TOKEN_FROM)
			channelTok := p.expect(lexer.TOKEN_IDENT)
			p.expect(lexer.TOKEN_COLON)
			p.expect(lexer.TOKEN_NEWLINE)
			body := p.parseBlock()
			cases = append(cases, ast.SelectCase{Channel: channelTok.Value, Body: body})
		} else if p.check(lexer.TOKEN_AFTER) {
			p.advance() // consume "after"
			afterMs = p.parseExpression()
			p.expect(lexer.TOKEN_COLON)
			p.expect(lexer.TOKEN_NEWLINE)
			afterBody = p.parseBlock()
		} else {
			p.error("expected 'when' or 'after' in select block")
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.SelectStatement{Cases: cases, AfterMs: afterMs, AfterBody: afterBody, Line: line}
}

// --- Trait parsing ---

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

func (p *Parser) parseBlock() []ast.Statement {
	p.expect(lexer.TOKEN_INDENT)
	stmts := []ast.Statement{}
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		stmt := p.parseStatement()
		stmts = append(stmts, stmt)
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
	return stmts
}

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
	left := p.parseUnary()
	for p.check(lexer.TOKEN_STAR) || p.check(lexer.TOKEN_SLASH) || p.check(lexer.TOKEN_MODULO) {
		op := p.advance().Value
		right := p.parseUnary()
		left = &ast.BinaryExpr{Left: left, Operator: op, Right: right}
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
			field := p.expect(lexer.TOKEN_IDENT)
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
		// Check for "ask claude" expression
		if tok.Value == "ask" && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_IDENT && p.tokens[p.pos+1].Value == "claude" {
			return p.parseAskExpression()
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
			next := p.expect(lexer.TOKEN_IDENT)
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
		field := p.expect(lexer.TOKEN_IDENT)
		return &ast.DotExpr{Object: &ast.Identifier{Name: "this"}, Field: field.Value}

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
	body := p.parseExpression()
	return &ast.LambdaExpr{Params: params, Body: body}
}

func (p *Parser) parseObjectLiteral() ast.Expression {
	p.advance() // consume "{"
	keys := []string{}
	values := []ast.Expression{}
	var computed []ast.ComputedProperty

	if !p.check(lexer.TOKEN_RBRACE) {
		// Parse first property (regular or computed)
		if p.check(lexer.TOKEN_LBRACKET) {
			// Computed property: {[expr]: value}
			p.advance() // consume "["
			keyExpr := p.parseExpression()
			p.expect(lexer.TOKEN_RBRACKET)
			p.expect(lexer.TOKEN_COLON)
			val := p.parseExpression()
			computed = append(computed, ast.ComputedProperty{KeyExpr: keyExpr, Value: val})
		} else {
			key := p.expect(lexer.TOKEN_IDENT)
			keys = append(keys, key.Value)
			p.expect(lexer.TOKEN_COLON)
			values = append(values, p.parseExpression())
		}

		for p.check(lexer.TOKEN_COMMA) {
			p.advance()
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
			} else {
				key := p.expect(lexer.TOKEN_IDENT)
				keys = append(keys, key.Value)
				p.expect(lexer.TOKEN_COLON)
				values = append(values, p.parseExpression())
			}
		}
	}

	p.expect(lexer.TOKEN_RBRACE)
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
	elements := []ast.Expression{}

	if !p.check(lexer.TOKEN_RBRACKET) {
		elements = append(elements, p.parseExpression())
		for p.check(lexer.TOKEN_COMMA) {
			p.advance()
			if p.check(lexer.TOKEN_RBRACKET) {
				break // trailing comma
			}
			elements = append(elements, p.parseExpression())
		}
	}

	p.expect(lexer.TOKEN_RBRACKET)
	return &ast.ListLiteral{Elements: elements}
}

// --- Helper methods ---

// --- Full-stack block parsing ---

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

func (p *Parser) current() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF, Value: "", Line: -1}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() lexer.Token {
	tok := p.current()
	p.pos++
	return tok
}

func (p *Parser) check(tokenType lexer.TokenType) bool {
	return p.current().Type == tokenType
}

func (p *Parser) checkNext(types ...lexer.TokenType) bool {
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

func (p *Parser) expect(tokenType lexer.TokenType) lexer.Token {
	if p.current().Type != tokenType {
		p.error(fmt.Sprintf("expected %s but found %q", tokenType, p.current().Value))
	}
	return p.advance()
}

func (p *Parser) consumeNewline() {
	if p.check(lexer.TOKEN_NEWLINE) {
		p.advance()
	}
	// Also OK if we hit EOF or DEDENT — end of statement
}

func (p *Parser) skipNewlines() {
	for p.check(lexer.TOKEN_NEWLINE) {
		p.advance()
	}
}

func (p *Parser) isAtEnd() bool {
	return p.current().Type == lexer.TOKEN_EOF
}

func (p *Parser) error(msg string) {
	panic(&ParseError{
		Line:    p.current().Line,
		Message: msg,
	})
}

// parseMatchObjectPattern parses {key: value, key2, ...} patterns in match/when.
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

// isStreamStatement checks if the current position is a "stream claude" statement.
func (p *Parser) isStreamStatement() bool {
	if p.pos+1 >= len(p.tokens) {
		return false
	}
	cur := p.tokens[p.pos]
	next := p.tokens[p.pos+1]
	return cur.Type == lexer.TOKEN_IDENT && cur.Value == "stream" &&
		next.Type == lexer.TOKEN_IDENT && next.Value == "claude"
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

	return &ast.AskExpression{
		Provider:   provider,
		Prompt:     prompt,
		Options:    options,
		IsMessages: isMessages,
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
