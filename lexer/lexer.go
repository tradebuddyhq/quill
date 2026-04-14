package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

var keywords = map[string]TokenType{
	"is":        TOKEN_IS,
	"are":       TOKEN_ARE,
	"say":       TOKEN_SAY,
	"if":        TOKEN_IF,
	"otherwise": TOKEN_OTHERWISE,
	"for":       TOKEN_FOR,
	"each":      TOKEN_EACH,
	"in":        TOKEN_IN,
	"to":        TOKEN_TO,
	"give":      TOKEN_GIVE,
	"back":      TOKEN_BACK,
	"and":       TOKEN_AND,
	"or":        TOKEN_OR,
	"not":       TOKEN_NOT,
	"greater":   TOKEN_GREATER,
	"less":      TOKEN_LESS,
	"than":      TOKEN_THAN,
	"equal":     TOKEN_EQUAL,
	"contains":  TOKEN_CONTAINS,
	"while":     TOKEN_WHILE,
	"use":       TOKEN_USE,
	"test":      TOKEN_TEST,
	"expect":    TOKEN_EXPECT,
	"yes":       TOKEN_YES,
	"no":        TOKEN_NO,
	"true":      TOKEN_YES,
	"false":     TOKEN_NO,
	"describe":  TOKEN_DESCRIBE,
	"new":       TOKEN_NEW,
	"my":        TOKEN_MY,
	"await":     TOKEN_AWAIT,
	"as":        TOKEN_AS,
	"try":       TOKEN_TRY,
	"fails":     TOKEN_FAILS,
	"extends":   TOKEN_EXTENDS,
	"from":      TOKEN_FROM,
	"with":      TOKEN_WITH,
	"nothing":   TOKEN_NOTHING,
	"break":     TOKEN_BREAK,
	"continue":  TOKEN_CONTINUE,
	"match":     TOKEN_MATCH,
	"when":      TOKEN_WHEN,
	"define":    TOKEN_DEFINE,
	"of":        TOKEN_OF,
	"spawn":     TOKEN_SPAWN,
	"task":      TOKEN_TASK,
	"parallel":  TOKEN_PARALLEL,
	"race":      TOKEN_RACE,
	"channel":   TOKEN_CHANNEL,
	"send":      TOKEN_SEND,
	"receive":   TOKEN_RECEIVE,
	"select":    TOKEN_SELECT,
	"after":     TOKEN_AFTER,
	"buffer":    TOKEN_BUFFER,
	"component": TOKEN_COMPONENT,
	"state":     TOKEN_STATE,
	"mount":     TOKEN_MOUNT,
	"link":      TOKEN_LINK,
	"head":      TOKEN_HEAD,
	"style":     TOKEN_STYLE,
	"form":      TOKEN_FORM,
	"redirect":  TOKEN_REDIRECT,
	"load":      TOKEN_LOAD,
	"trait":     TOKEN_TRAIT,
	"where":     TOKEN_WHERE,
	"using":     TOKEN_USING,
	"self":      TOKEN_SELF,
	"yield":     TOKEN_YIELD,
	"loop":      TOKEN_LOOP,
	"mock":      TOKEN_MOCK,
	// "server" is no longer a global keyword — it is handled contextually
	"route":     TOKEN_ROUTE,
	"database":  TOKEN_DATABASE,
	"respond":   TOKEN_RESPOND,
	"status":    TOKEN_STATUS,
	"model":     TOKEN_MODEL,
	"connect":   TOKEN_CONNECT,
	"port":      TOKEN_PORT,
	"cancel":    TOKEN_CANCEL,
	"settled":   TOKEN_SETTLED,
	"private":   TOKEN_PRIVATE,
	"public":    TOKEN_PUBLIC,
	"type":      TOKEN_TYPE,
	"Partial":   TOKEN_PARTIAL,
	"Omit":      TOKEN_OMIT,
	"Pick":      TOKEN_PICK,
	"Record":    TOKEN_RECORD,
	"Readonly":  TOKEN_READONLY,
	"Required":  TOKEN_REQUIRED,
	"websocket": TOKEN_WEBSOCKET,
	"on":          TOKEN_ON,
	"broadcast":   TOKEN_BROADCAST,
	"command":     TOKEN_COMMAND,
	"described":   TOKEN_DESCRIBED,
	"reply":       TOKEN_REPLY,
	"embed":       TOKEN_EMBED,
	"worker":      TOKEN_WORKER,
	"screen":      TOKEN_SCREEN,
	"navigate":    TOKEN_NAVIGATE,
	"effect":      TOKEN_EFFECT,
	"every":       TOKEN_EVERY,
	"delete":      TOKEN_DELETE,
}

type Lexer struct {
	source      string
	pos         int
	line        int
	col         int
	tokens      []Token
	indentStack []int
	atLineStart bool
}

func New(source string) *Lexer {
	// Normalize line endings and ensure trailing newline
	source = strings.ReplaceAll(source, "\r\n", "\n")
	source = strings.ReplaceAll(source, "\r", "\n")
	if len(source) > 0 && source[len(source)-1] != '\n' {
		source += "\n"
	}

	return &Lexer{
		source:      source,
		pos:         0,
		line:        1,
		col:         1,
		indentStack: []int{0},
		atLineStart: true,
	}
}

func (l *Lexer) Tokenize() ([]Token, error) {
	for l.pos < len(l.source) {
		if l.atLineStart {
			l.handleIndentation()
			l.atLineStart = false
			if l.pos >= len(l.source) {
				break
			}
			// Skip blank lines
			if l.source[l.pos] == '\n' {
				l.pos++
				l.line++
				l.col = 1
				l.atLineStart = true
				continue
			}
		}

		ch := l.source[l.pos]

		switch {
		case ch == '\n':
			l.addToken(TOKEN_NEWLINE, "\\n")
			l.pos++
			l.line++
			l.col = 1
			l.atLineStart = true

		case ch == ' ' || ch == '\t':
			// Skip whitespace (mid-line)
			l.pos++
			l.col++

		case ch == '-' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '-':
			l.skipComment()

		case ch == '"':
			if err := l.readString(); err != nil {
				return nil, err
			}

		case isDigit(ch):
			l.readNumber()

		case isAlpha(ch) || ch == '_':
			l.readIdentOrKeyword()

		case ch == '+':
			l.addToken(TOKEN_PLUS, "+")
			l.advance()
		case ch == '*':
			l.addToken(TOKEN_STAR, "*")
			l.advance()
		case ch == '/':
			l.addToken(TOKEN_SLASH, "/")
			l.advance()
		case ch == '%':
			l.addToken(TOKEN_MODULO, "%")
			l.advance()
		case ch == '.':
			if l.pos+2 < len(l.source) && l.source[l.pos+1] == '.' && l.source[l.pos+2] == '.' {
				l.addToken(TOKEN_SPREAD, "...")
				l.advance()
				l.advance()
				l.advance()
			} else {
				l.addToken(TOKEN_DOT, ".")
				l.advance()
			}
		case ch == ':':
			l.addToken(TOKEN_COLON, ":")
			l.advance()
		case ch == ',':
			l.addToken(TOKEN_COMMA, ",")
			l.advance()
		case ch == '(':
			l.addToken(TOKEN_LPAREN, "(")
			l.advance()
		case ch == ')':
			l.addToken(TOKEN_RPAREN, ")")
			l.advance()
		case ch == '[':
			l.addToken(TOKEN_LBRACKET, "[")
			l.advance()
		case ch == ']':
			l.addToken(TOKEN_RBRACKET, "]")
			l.advance()
		case ch == '-':
			if l.pos+1 < len(l.source) && l.source[l.pos+1] == '>' {
				l.addToken(TOKEN_ARROW, "->")
				l.advance()
				l.advance()
			} else {
				l.addToken(TOKEN_MINUS, "-")
				l.advance()
			}

		case ch == '^':
			l.addToken(TOKEN_CARET, "^")
			l.advance()

		case ch == '|':
			l.addToken(TOKEN_PIPE, "|")
			l.advance()

		case ch == '?':
			l.addToken(TOKEN_QUESTION, "?")
			l.advance()

		case ch == '`':
			if err := l.readBacktickString(); err != nil {
				return nil, err
			}

		case ch == '@':
			l.addToken(TOKEN_AT, "@")
			l.advance()

		case ch == '{':
			l.addToken(TOKEN_LBRACE, "{")
			l.advance()
		case ch == '}':
			l.addToken(TOKEN_RBRACE, "}")
			l.advance()
		case ch == '=':
			l.addToken(TOKEN_ASSIGN, "=")
			l.advance()

		default:
			return nil, fmt.Errorf("line %d, column %d: unexpected character %q", l.line, l.col, string(ch))
		}
	}

	// Emit remaining dedents
	for len(l.indentStack) > 1 {
		l.addToken(TOKEN_DEDENT, "")
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
	}

	l.addToken(TOKEN_EOF, "")
	return l.tokens, nil
}

func (l *Lexer) handleIndentation() {
	indent := 0
	for l.pos < len(l.source) && (l.source[l.pos] == ' ' || l.source[l.pos] == '\t') {
		if l.source[l.pos] == '\t' {
			indent += 2 // 1 tab = 2 spaces
		} else {
			indent++
		}
		l.pos++
		l.col++
	}

	// Skip blank lines and comment-only lines
	if l.pos >= len(l.source) || l.source[l.pos] == '\n' {
		return
	}
	if l.source[l.pos] == '-' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '-' {
		return
	}

	currentIndent := l.indentStack[len(l.indentStack)-1]

	if indent > currentIndent {
		l.indentStack = append(l.indentStack, indent)
		l.addToken(TOKEN_INDENT, "")
	} else if indent < currentIndent {
		for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > indent {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			l.addToken(TOKEN_DEDENT, "")
		}
		if l.indentStack[len(l.indentStack)-1] != indent {
			// We'll handle this as a best-effort — don't crash
			l.addToken(TOKEN_DEDENT, "")
		}
	}
}

func (l *Lexer) addToken(tokenType TokenType, value string) {
	l.tokens = append(l.tokens, Token{
		Type:   tokenType,
		Value:  value,
		Line:   l.line,
		Column: l.col,
	})
}

func (l *Lexer) advance() {
	l.pos++
	l.col++
}

func (l *Lexer) skipComment() {
	// Skip from -- to end of line
	for l.pos < len(l.source) && l.source[l.pos] != '\n' {
		l.pos++
		l.col++
	}
}

func (l *Lexer) readString() error {
	// Check for triple-quote multiline string """..."""
	if l.pos+2 < len(l.source) && l.source[l.pos+1] == '"' && l.source[l.pos+2] == '"' {
		return l.readMultilineString()
	}

	l.pos++ // skip opening quote
	l.col++
	start := l.pos
	startLine := l.line

	for l.pos < len(l.source) && l.source[l.pos] != '"' {
		if l.source[l.pos] == '\\' && l.pos+1 < len(l.source) {
			l.pos += 2
			l.col += 2
			continue
		}
		if l.source[l.pos] == '\n' {
			return fmt.Errorf("line %d: unterminated string (started on line %d)", l.line, startLine)
		}
		l.pos++
		l.col++
	}

	if l.pos >= len(l.source) {
		return fmt.Errorf("line %d: unterminated string", startLine)
	}

	value := l.source[start:l.pos]
	l.addToken(TOKEN_STRING, value)
	l.pos++ // skip closing quote
	l.col++
	return nil
}

func (l *Lexer) readMultilineString() error {
	l.pos += 3 // skip opening """
	l.col += 3
	start := l.pos
	startLine := l.line

	// Skip leading newline after opening """
	if l.pos < len(l.source) && l.source[l.pos] == '\n' {
		l.pos++
		l.line++
		l.col = 0
		start = l.pos
	}

	for l.pos < len(l.source) {
		if l.source[l.pos] == '"' && l.pos+2 < len(l.source) && l.source[l.pos+1] == '"' && l.source[l.pos+2] == '"' {
			value := l.source[start:l.pos]
			// Trim trailing newline before closing """
			if len(value) > 0 && value[len(value)-1] == '\n' {
				value = value[:len(value)-1]
			}
			l.addToken(TOKEN_STRING, value)
			l.pos += 3 // skip closing """
			l.col += 3
			return nil
		}
		if l.source[l.pos] == '\\' && l.pos+1 < len(l.source) {
			l.pos += 2
			l.col += 2
			continue
		}
		if l.source[l.pos] == '\n' {
			l.line++
			l.col = 0
		} else {
			l.col++
		}
		l.pos++
	}

	return fmt.Errorf("line %d: unterminated multiline string (started on line %d)", l.line, startLine)
}

func (l *Lexer) readNumber() {
	start := l.pos
	for l.pos < len(l.source) && isDigit(l.source[l.pos]) {
		l.pos++
		l.col++
	}
	// Check for decimal point
	if l.pos < len(l.source) && l.source[l.pos] == '.' && l.pos+1 < len(l.source) && isDigit(l.source[l.pos+1]) {
		l.pos++
		l.col++
		for l.pos < len(l.source) && isDigit(l.source[l.pos]) {
			l.pos++
			l.col++
		}
	}
	l.addToken(TOKEN_NUMBER, l.source[start:l.pos])
}

func (l *Lexer) readIdentOrKeyword() {
	start := l.pos
	for l.pos < len(l.source) && (isAlpha(l.source[l.pos]) || isDigit(l.source[l.pos]) || l.source[l.pos] == '_') {
		l.pos++
		l.col++
	}
	word := l.source[start:l.pos]

	if tokenType, ok := keywords[word]; ok {
		l.addToken(tokenType, word)
	} else {
		l.addToken(TOKEN_IDENT, word)
	}
}

func (l *Lexer) readBacktickString() error {
	l.addToken(TOKEN_BACKTICK, "`")
	l.pos++ // skip opening backtick
	l.col++
	start := l.pos
	startLine := l.line

	for l.pos < len(l.source) && l.source[l.pos] != '`' {
		if l.source[l.pos] == '\n' {
			l.line++
			l.col = 0
		}
		l.pos++
		l.col++
	}

	if l.pos >= len(l.source) {
		return fmt.Errorf("line %d: unterminated backtick string (started on line %d)", l.line, startLine)
	}

	value := l.source[start:l.pos]
	l.addToken(TOKEN_STRING, value)
	l.pos++ // skip closing backtick
	l.col++
	l.addToken(TOKEN_BACKTICK, "`")
	return nil
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isAlpha(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}
