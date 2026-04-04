package lexer

import "fmt"

type TokenType int

const (
	TOKEN_EOF TokenType = iota
	TOKEN_NEWLINE
	TOKEN_INDENT
	TOKEN_DEDENT

	// Literals
	TOKEN_STRING
	TOKEN_NUMBER
	TOKEN_YES
	TOKEN_NO

	// Identifier
	TOKEN_IDENT

	// Keywords
	TOKEN_IS
	TOKEN_ARE
	TOKEN_SAY
	TOKEN_IF
	TOKEN_OTHERWISE
	TOKEN_FOR
	TOKEN_EACH
	TOKEN_IN
	TOKEN_TO
	TOKEN_GIVE
	TOKEN_BACK
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT
	TOKEN_GREATER
	TOKEN_LESS
	TOKEN_THAN
	TOKEN_EQUAL
	TOKEN_CONTAINS
	TOKEN_WHILE
	TOKEN_USE
	TOKEN_TEST
	TOKEN_EXPECT

	// Operators
	TOKEN_PLUS
	TOKEN_MINUS
	TOKEN_STAR
	TOKEN_SLASH
	TOKEN_MODULO
	TOKEN_DOT

	// Delimiters
	TOKEN_COLON
	TOKEN_COMMA
	TOKEN_LPAREN
	TOKEN_RPAREN
	TOKEN_LBRACKET
	TOKEN_RBRACKET
)

var tokenNames = map[TokenType]string{
	TOKEN_EOF:       "end of file",
	TOKEN_NEWLINE:   "newline",
	TOKEN_INDENT:    "indent",
	TOKEN_DEDENT:    "dedent",
	TOKEN_STRING:    "text",
	TOKEN_NUMBER:    "number",
	TOKEN_YES:       "yes",
	TOKEN_NO:        "no",
	TOKEN_IDENT:     "name",
	TOKEN_IS:        "is",
	TOKEN_ARE:       "are",
	TOKEN_SAY:       "say",
	TOKEN_IF:        "if",
	TOKEN_OTHERWISE: "otherwise",
	TOKEN_FOR:       "for",
	TOKEN_EACH:      "each",
	TOKEN_IN:        "in",
	TOKEN_TO:        "to",
	TOKEN_GIVE:      "give",
	TOKEN_BACK:      "back",
	TOKEN_AND:       "and",
	TOKEN_OR:        "or",
	TOKEN_NOT:       "not",
	TOKEN_GREATER:   "greater",
	TOKEN_LESS:      "less",
	TOKEN_THAN:      "than",
	TOKEN_EQUAL:     "equal",
	TOKEN_CONTAINS:  "contains",
	TOKEN_WHILE:     "while",
	TOKEN_USE:       "use",
	TOKEN_TEST:      "test",
	TOKEN_EXPECT:    "expect",
	TOKEN_PLUS:      "+",
	TOKEN_MINUS:     "-",
	TOKEN_STAR:      "*",
	TOKEN_SLASH:     "/",
	TOKEN_MODULO:    "%",
	TOKEN_DOT:       ".",
	TOKEN_COLON:     ":",
	TOKEN_COMMA:     ",",
	TOKEN_LPAREN:    "(",
	TOKEN_RPAREN:    ")",
	TOKEN_LBRACKET:  "[",
	TOKEN_RBRACKET:  "]",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", int(t))
}

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%s, %q, line %d)", t.Type, t.Value, t.Line)
}
