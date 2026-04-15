package parser

import (
	"fmt"
	"quill/ast"
	"quill/lexer"
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

// isAssignment returns true if the current token can be used as a variable name
// and is followed by TOKEN_IS or TOKEN_ARE. This allows keywords like "command",
// "embed", "topic", etc. to be used as variable names in assignment context.
func (p *Parser) isAssignment() bool {
	if p.isAtEnd() || p.pos+1 >= len(p.tokens) {
		return false
	}
	next := p.tokens[p.pos+1].Type
	if next != lexer.TOKEN_IS && next != lexer.TOKEN_ARE {
		return false
	}
	cur := p.current().Type
	if cur == lexer.TOKEN_IDENT {
		return true
	}
	// Allow keyword tokens as variable names, except control flow keywords
	// that have their own statement syntax
	switch cur {
	case lexer.TOKEN_IF, lexer.TOKEN_FOR, lexer.TOKEN_WHILE, lexer.TOKEN_TO,
		lexer.TOKEN_GIVE, lexer.TOKEN_SAY, lexer.TOKEN_FROM,
		lexer.TOKEN_TRY, lexer.TOKEN_RAISE,
		lexer.TOKEN_BREAK, lexer.TOKEN_CONTINUE,
		lexer.TOKEN_AT, lexer.TOKEN_OTHERWISE,
		lexer.TOKEN_NEWLINE, lexer.TOKEN_INDENT, lexer.TOKEN_DEDENT, lexer.TOKEN_EOF:
		return false
	}
	// Any other keyword token followed by is/are is an assignment
	return isKeywordToken(cur)
}

// isDotAssignment returns true if the current token starts a dot assignment
// like "obj.field is value", allowing keyword tokens on the left side.
func (p *Parser) isDotAssignment() bool {
	if p.pos+3 >= len(p.tokens) {
		return false
	}
	cur := p.current().Type
	if cur != lexer.TOKEN_IDENT && !isKeywordToken(cur) {
		return false
	}
	if p.tokens[p.pos+1].Type != lexer.TOKEN_DOT {
		return false
	}
	t2 := p.tokens[p.pos+2].Type
	if t2 != lexer.TOKEN_IDENT && !isKeywordToken(t2) {
		return false
	}
	t3 := p.tokens[p.pos+3].Type
	return t3 == lexer.TOKEN_IS || t3 == lexer.TOKEN_ARE
}

func (p *Parser) parseStatement() ast.Statement {
	switch {
	// Assignment must be checked BEFORE keyword-specific handlers so that
	// "command is x" parses as an assignment, not as the "command" keyword stmt.
	case p.isDotAssignment():
		return p.parseDotAssignment()
	case p.isAssignment():
		return p.parseAssignment()
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
	case p.check(lexer.TOKEN_RAISE):
		return p.parseRaise()
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
	case p.check(lexer.TOKEN_NAVIGATE):
		return p.parseNavigateStmt()
	case p.check(lexer.TOKEN_EVERY):
		return p.parseEveryStatement()
	case p.check(lexer.TOKEN_DELETE):
		return p.parseDelete()
	case p.check(lexer.TOKEN_IDENT) && p.current().Value == "app" && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_IDENT && p.tokens[p.pos+1].Value == "navigation":
		return p.parseNavigationBlock()
	case p.check(lexer.TOKEN_IDENT) && p.current().Value == "context" && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_IDENT:
		return p.parseContextDecl()
	case p.check(lexer.TOKEN_IDENT) && p.current().Value == "server" && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_COLON:
		return p.parseServerBlock()
	case p.check(lexer.TOKEN_IDENT) && p.current().Value == "auth" && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_COLON:
		return p.parseAuthBlock()
	case p.check(lexer.TOKEN_DATABASE):
		return p.parseDatabaseBlock()
	case p.isStreamStatement():
		return p.parseStreamStatement()
	case p.isAgentStatement():
		return p.parseAgentStatement()
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
	case p.isBracketAssignment():
		return p.parseBracketAssignment()
	case p.isOnStatement():
		return p.parseOnStatement()
	default:
		return p.parseExprStatement()
	}
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

// isKeywordToken returns true if the token type is a keyword that can also
// be used as an identifier in certain contexts (after a dot, as object key).
func isKeywordToken(t lexer.TokenType) bool {
	switch t {
	case lexer.TOKEN_ON, lexer.TOKEN_USE, lexer.TOKEN_SEND, lexer.TOKEN_STATUS,
		lexer.TOKEN_TYPE, lexer.TOKEN_FROM, lexer.TOKEN_SELECT, lexer.TOKEN_SERVER,
		lexer.TOKEN_LOAD, lexer.TOKEN_AFTER, lexer.TOKEN_STATE,
		lexer.TOKEN_ROUTE, lexer.TOKEN_PORT, lexer.TOKEN_MODEL, lexer.TOKEN_CONNECT,
		lexer.TOKEN_HEAD, lexer.TOKEN_STYLE, lexer.TOKEN_FORM, lexer.TOKEN_LINK,
		lexer.TOKEN_BROADCAST, lexer.TOKEN_CHANNEL, lexer.TOKEN_RECEIVE,
		lexer.TOKEN_BUFFER, lexer.TOKEN_MATCH, lexer.TOKEN_RESPOND,
		lexer.TOKEN_DATABASE, lexer.TOKEN_MOUNT, lexer.TOKEN_COMPONENT,
		lexer.TOKEN_COMMAND, lexer.TOKEN_REPLY, lexer.TOKEN_EMBED,
		lexer.TOKEN_WORKER, lexer.TOKEN_CANCEL, lexer.TOKEN_SETTLED,
		lexer.TOKEN_SCREEN, lexer.TOKEN_NAVIGATE, lexer.TOKEN_EFFECT,
		lexer.TOKEN_EVERY,
		lexer.TOKEN_SPAWN, lexer.TOKEN_TASK, lexer.TOKEN_PARALLEL, lexer.TOKEN_RACE,
		lexer.TOKEN_DESCRIBE, lexer.TOKEN_NEW, lexer.TOKEN_TEST, lexer.TOKEN_EXPECT,
		lexer.TOKEN_MOCK, lexer.TOKEN_TRAIT, lexer.TOKEN_WHERE, lexer.TOKEN_USING,
		lexer.TOKEN_SELF, lexer.TOKEN_PRIVATE, lexer.TOKEN_PUBLIC,
		lexer.TOKEN_YIELD, lexer.TOKEN_LOOP, lexer.TOKEN_DEFINE,
		lexer.TOKEN_REDIRECT, lexer.TOKEN_WEBSOCKET, lexer.TOKEN_DESCRIBED,
		lexer.TOKEN_IS, lexer.TOKEN_ARE, lexer.TOKEN_SAY, lexer.TOKEN_IF,
		lexer.TOKEN_OTHERWISE, lexer.TOKEN_FOR, lexer.TOKEN_EACH, lexer.TOKEN_IN,
		lexer.TOKEN_TO, lexer.TOKEN_GIVE, lexer.TOKEN_BACK, lexer.TOKEN_AND,
		lexer.TOKEN_OR, lexer.TOKEN_NOT, lexer.TOKEN_GREATER, lexer.TOKEN_LESS,
		lexer.TOKEN_THAN, lexer.TOKEN_EQUAL, lexer.TOKEN_CONTAINS,
		lexer.TOKEN_WHILE, lexer.TOKEN_MY, lexer.TOKEN_AWAIT, lexer.TOKEN_AS,
		lexer.TOKEN_TRY, lexer.TOKEN_FAILS, lexer.TOKEN_EXTENDS, lexer.TOKEN_WITH,
		lexer.TOKEN_NOTHING, lexer.TOKEN_BREAK, lexer.TOKEN_CONTINUE,
		lexer.TOKEN_WHEN, lexer.TOKEN_OF, lexer.TOKEN_YES, lexer.TOKEN_NO:
		return true
	}
	return false
}

// expectIdentOrKeyword consumes the current token if it's an identifier or a
// keyword that can be used as an identifier in this context, and returns it.
func (p *Parser) expectIdentOrKeyword() lexer.Token {
	if p.current().Type == lexer.TOKEN_IDENT || isKeywordToken(p.current().Type) {
		return p.advance()
	}
	p.error(fmt.Sprintf("expected name but found %q", p.current().Value))
	return p.advance() // unreachable
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

// skipNewlinesAndIndent skips newlines along with indent/dedent tokens,
// used inside object/array literals to support multiline syntax.
func (p *Parser) skipNewlinesAndIndent() {
	for p.check(lexer.TOKEN_NEWLINE) || p.check(lexer.TOKEN_INDENT) || p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
}

// countIndentsBetween counts the net indent depth (indents minus dedents)
// in the token range [from, to).
func (p *Parser) countIndentsBetween(from, to int) int {
	depth := 0
	for i := from; i < to && i < len(p.tokens); i++ {
		if p.tokens[i].Type == lexer.TOKEN_INDENT {
			depth++
		} else if p.tokens[i].Type == lexer.TOKEN_DEDENT {
			depth--
		}
	}
	return depth
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