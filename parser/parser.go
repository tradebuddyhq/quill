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
	tokens []lexer.Token
	pos    int
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
	case p.check(lexer.TOKEN_IDENT) && p.pos+3 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_DOT && p.tokens[p.pos+2].Type == lexer.TOKEN_IDENT && (p.tokens[p.pos+3].Type == lexer.TOKEN_IS || p.tokens[p.pos+3].Type == lexer.TOKEN_ARE):
		return p.parseDotAssignment()
	case p.check(lexer.TOKEN_IDENT) && p.checkNext(lexer.TOKEN_IS, lexer.TOKEN_ARE):
		return p.parseAssignment()
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
	p.expect(lexer.TOKEN_EACH)
	varTok := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_IN)
	iterable := p.parseExpression()
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()

	return &ast.ForEachStatement{
		Variable: varTok.Value,
		Iterable: iterable,
		Body:     body,
		Line:     line,
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

func (p *Parser) parseExpect() *ast.ExpectStatement {
	line := p.current().Line
	p.advance() // consume "expect"
	expr := p.parseExpression()
	p.consumeNewline()
	return &ast.ExpectStatement{Expr: expr, Line: line}
}

func (p *Parser) parseDescribe() *ast.DescribeStatement {
	line := p.current().Line
	p.advance() // consume "describe"
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

	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		if p.check(lexer.TOKEN_TO) {
			method := p.parseFuncDef()
			methods = append(methods, *method)
		} else if p.check(lexer.TOKEN_IDENT) && p.checkNext(lexer.TOKEN_IS, lexer.TOKEN_ARE) {
			assign := p.parseAssignment()
			props = append(props, *assign)
		} else {
			p.error("expected a property or method inside describe block")
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.DescribeStatement{Name: name.Value, Extends: extends, Properties: props, Methods: methods, Line: line}
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
			pattern := p.parseExpression()

			// Optional guard: "if condition"
			var guard ast.Expression
			if p.check(lexer.TOKEN_IF) {
				p.advance()
				guard = p.parseExpression()
			}

			p.expect(lexer.TOKEN_COLON)
			p.expect(lexer.TOKEN_NEWLINE)
			body := p.parseBlock()
			cases = append(cases, ast.MatchCase{Pattern: pattern, Guard: guard, Body: body})
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
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		variantName := p.expect(lexer.TOKEN_IDENT)
		fields := []string{}
		// Parse optional fields for algebraic data types
		for p.check(lexer.TOKEN_IDENT) {
			fields = append(fields, p.advance().Value)
		}
		variants = append(variants, ast.EnumVariant{Name: variantName.Value, Fields: fields})
		p.consumeNewline()
	}

	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.DefineStatement{Name: name.Value, Variants: variants, Line: line}
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

func (p *Parser) parseExprStatement() *ast.ExprStatement {
	line := p.current().Line
	expr := p.parseExpression()
	p.consumeNewline()
	return &ast.ExprStatement{Expr: expr, Line: line}
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
		expr := p.parseUnary()
		return &ast.AwaitExpr{Expr: expr}
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
		p.advance()
		return &ast.Identifier{Name: tok.Value}

	case lexer.TOKEN_NOTHING:
		p.advance()
		return &ast.NothingLiteral{}

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
		className := p.expect(lexer.TOKEN_IDENT)
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
		return &ast.NewExpr{ClassName: className.Value, Args: args}

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

	if !p.check(lexer.TOKEN_RBRACE) {
		// First key: value pair
		key := p.expect(lexer.TOKEN_IDENT)
		keys = append(keys, key.Value)
		p.expect(lexer.TOKEN_COLON)
		values = append(values, p.parseExpression())

		for p.check(lexer.TOKEN_COMMA) {
			p.advance()
			if p.check(lexer.TOKEN_RBRACE) {
				break // trailing comma
			}
			key = p.expect(lexer.TOKEN_IDENT)
			keys = append(keys, key.Value)
			p.expect(lexer.TOKEN_COLON)
			values = append(values, p.parseExpression())
		}
	}

	p.expect(lexer.TOKEN_RBRACE)
	return &ast.ObjectLiteral{Keys: keys, Values: values}
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
