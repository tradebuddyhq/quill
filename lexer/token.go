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
	TOKEN_DESCRIBE
	TOKEN_NEW
	TOKEN_MY
	TOKEN_AWAIT
	TOKEN_AS
	TOKEN_TRY
	TOKEN_FAILS
	TOKEN_RAISE
	TOKEN_EXTENDS
	TOKEN_FROM
	TOKEN_WITH
	TOKEN_NOTHING
	TOKEN_ARROW   // ->
	TOKEN_LBRACE  // {
	TOKEN_RBRACE  // }
	TOKEN_SPREAD  // ...
	TOKEN_BREAK
	TOKEN_CONTINUE
	TOKEN_MATCH
	TOKEN_WHEN
	TOKEN_DEFINE
	TOKEN_OF
	TOKEN_PIPE // |

	// Concurrency keywords
	TOKEN_SPAWN
	TOKEN_TASK
	TOKEN_PARALLEL
	TOKEN_RACE
	TOKEN_CHANNEL
	TOKEN_SEND
	TOKEN_RECEIVE
	TOKEN_SELECT
	TOKEN_AFTER
	TOKEN_BUFFER

	// Framework keywords
	TOKEN_COMPONENT
	TOKEN_STATE
	TOKEN_MOUNT
	TOKEN_LINK
	TOKEN_HEAD
	TOKEN_STYLE
	TOKEN_FORM
	TOKEN_REDIRECT
	TOKEN_LOAD

	// Type system keywords
	TOKEN_TRAIT
	TOKEN_WHERE
	TOKEN_USING
	TOKEN_SELF

	// Cancel keyword
	TOKEN_CANCEL
	// Question mark operator
	TOKEN_QUESTION
	// Settled keyword
	TOKEN_SETTLED

	// Visibility keywords
	TOKEN_PRIVATE
	TOKEN_PUBLIC

	// Iterator/generator keywords
	TOKEN_YIELD
	TOKEN_LOOP

	// Mock keyword
	TOKEN_MOCK

	// Template literal
	TOKEN_BACKTICK

	// Full-stack keywords
	TOKEN_SERVER
	TOKEN_ROUTE
	TOKEN_DATABASE
	TOKEN_RESPOND
	TOKEN_STATUS
	TOKEN_MODEL
	TOKEN_CONNECT
	TOKEN_PORT

	// Operators
	TOKEN_PLUS
	TOKEN_MINUS
	TOKEN_STAR
	TOKEN_SLASH
	TOKEN_MODULO
	TOKEN_CARET
	TOKEN_DOT

	// Type utility keywords
	TOKEN_TYPE
	TOKEN_PARTIAL
	TOKEN_OMIT
	TOKEN_PICK
	TOKEN_RECORD
	TOKEN_READONLY
	TOKEN_REQUIRED

	// Decorator
	TOKEN_AT

	// WebSocket keywords
	TOKEN_WEBSOCKET
	TOKEN_ON
	TOKEN_BROADCAST

	// Discord keywords
	TOKEN_COMMAND
	TOKEN_DESCRIBED
	TOKEN_REPLY
	TOKEN_EMBED

	// Worker keywords
	TOKEN_WORKER

	// AI keywords
	TOKEN_ASK
	TOKEN_STREAM

	// Expo / React Native keywords
	TOKEN_SCREEN
	TOKEN_NAVIGATE
	TOKEN_EFFECT

	// Cron keywords
	TOKEN_EVERY

	// Delete keyword
	TOKEN_DELETE

	// Assignment
	TOKEN_ASSIGN // =

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
	TOKEN_DESCRIBE:  "describe",
	TOKEN_NEW:       "new",
	TOKEN_MY:        "my",
	TOKEN_AWAIT:     "await",
	TOKEN_AS:        "as",
	TOKEN_TRY:       "try",
	TOKEN_FAILS:     "fails",
	TOKEN_RAISE:     "raise",
	TOKEN_EXTENDS:   "extends",
	TOKEN_FROM:      "from",
	TOKEN_WITH:      "with",
	TOKEN_NOTHING:   "nothing",
	TOKEN_ARROW:     "->",
	TOKEN_LBRACE:    "{",
	TOKEN_RBRACE:    "}",
	TOKEN_SPREAD:    "...",
	TOKEN_BREAK:     "break",
	TOKEN_CONTINUE:  "continue",
	TOKEN_MATCH:     "match",
	TOKEN_WHEN:      "when",
	TOKEN_DEFINE:    "define",
	TOKEN_OF:        "of",
	TOKEN_PIPE:      "|",
	TOKEN_SPAWN:     "spawn",
	TOKEN_TASK:      "task",
	TOKEN_PARALLEL:  "parallel",
	TOKEN_RACE:      "race",
	TOKEN_CHANNEL:   "channel",
	TOKEN_SEND:      "send",
	TOKEN_RECEIVE:   "receive",
	TOKEN_SELECT:    "select",
	TOKEN_AFTER:     "after",
	TOKEN_BUFFER:    "buffer",
	TOKEN_COMPONENT: "component",
	TOKEN_STATE:     "state",
	TOKEN_MOUNT:     "mount",
	TOKEN_LINK:      "link",
	TOKEN_HEAD:      "head",
	TOKEN_STYLE:     "style",
	TOKEN_FORM:      "form",
	TOKEN_REDIRECT:  "redirect",
	TOKEN_LOAD:      "load",
	TOKEN_TRAIT:     "trait",
	TOKEN_WHERE:     "where",
	TOKEN_USING:     "using",
	TOKEN_SELF:      "self",
	TOKEN_CANCEL:    "cancel",
	TOKEN_QUESTION:  "?",
	TOKEN_SETTLED:   "settled",
	TOKEN_PRIVATE:   "private",
	TOKEN_PUBLIC:    "public",
	TOKEN_YIELD:     "yield",
	TOKEN_LOOP:      "loop",
	TOKEN_MOCK:      "mock",
	TOKEN_BACKTICK:  "`",
	TOKEN_SERVER:    "server",
	TOKEN_ROUTE:     "route",
	TOKEN_DATABASE:  "database",
	TOKEN_RESPOND:   "respond",
	TOKEN_STATUS:    "status",
	TOKEN_MODEL:     "model",
	TOKEN_CONNECT:   "connect",
	TOKEN_PORT:      "port",
	TOKEN_PLUS:      "+",
	TOKEN_MINUS:     "-",
	TOKEN_STAR:      "*",
	TOKEN_SLASH:     "/",
	TOKEN_MODULO:    "%",
	TOKEN_CARET:     "^",
	TOKEN_DOT:       ".",
	TOKEN_COLON:     ":",
	TOKEN_COMMA:     ",",
	TOKEN_LPAREN:    "(",
	TOKEN_RPAREN:    ")",
	TOKEN_LBRACKET:  "[",
	TOKEN_RBRACKET:  "]",
	TOKEN_TYPE:      "type",
	TOKEN_PARTIAL:   "Partial",
	TOKEN_OMIT:      "Omit",
	TOKEN_PICK:      "Pick",
	TOKEN_RECORD:    "Record",
	TOKEN_READONLY:  "Readonly",
	TOKEN_REQUIRED:  "Required",
	TOKEN_AT:        "@",
	TOKEN_WEBSOCKET: "websocket",
	TOKEN_ON:          "on",
	TOKEN_BROADCAST:   "broadcast",
	TOKEN_COMMAND:     "command",
	TOKEN_DESCRIBED:   "described",
	TOKEN_REPLY:       "reply",
	TOKEN_EMBED:       "embed",
	TOKEN_WORKER:      "worker",
	TOKEN_ASK:         "ask",
	TOKEN_STREAM:      "stream",
	TOKEN_SCREEN:      "screen",
	TOKEN_NAVIGATE:    "navigate",
	TOKEN_EFFECT:      "effect",
	TOKEN_EVERY:       "every",
	TOKEN_DELETE:      "delete",
	TOKEN_ASSIGN:      "=",
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
