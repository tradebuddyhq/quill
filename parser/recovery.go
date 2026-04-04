package parser

import (
	"fmt"
	"quill/ast"
	"quill/lexer"
	"strings"
)

// ParseErrorDetail represents a single parse error with rich context.
type ParseErrorDetail struct {
	Line    int
	Column  int
	Message string
	Hint    string // helpful suggestion
	Code    string // error code like "E001"
}

func (e *ParseErrorDetail) Error() string {
	s := fmt.Sprintf("Error[%s]: line %d, col %d: %s", e.Code, e.Line, e.Column, e.Message)
	if e.Hint != "" {
		s += fmt.Sprintf("\n  hint: %s", e.Hint)
	}
	return s
}

// ErrorRecovery collects parse errors and enables recovery.
type ErrorRecovery struct {
	errors    []ParseErrorDetail
	maxErrors int // stop after this many (default 50)
}

// NewErrorRecovery creates a new ErrorRecovery instance.
func NewErrorRecovery() *ErrorRecovery {
	return &ErrorRecovery{
		maxErrors: 50,
	}
}

// Errors returns the collected errors.
func (r *ErrorRecovery) Errors() []ParseErrorDetail {
	return r.errors
}

// HasErrors returns true if any errors were collected.
func (r *ErrorRecovery) HasErrors() bool {
	return len(r.errors) > 0
}

// ErrorCount returns the number of collected errors.
func (r *ErrorRecovery) ErrorCount() int {
	return len(r.errors)
}

// TooManyErrors returns true if the max error threshold has been reached.
func (r *ErrorRecovery) TooManyErrors() bool {
	return len(r.errors) >= r.maxErrors
}

// addError adds an error to the collection.
func (r *ErrorRecovery) addError(line, col int, code, message, hint string) {
	r.errors = append(r.errors, ParseErrorDetail{
		Line:    line,
		Column:  col,
		Message: message,
		Hint:    hint,
		Code:    code,
	})
}

// synchronize skips tokens until a known statement boundary is found.
// This allows the parser to continue after encountering an error.
func (p *Parser) synchronize() {
	for !p.isAtEnd() {
		// If we hit a newline followed by a statement keyword, stop after the newline
		if p.check(lexer.TOKEN_NEWLINE) {
			p.advance()
			p.skipNewlines()
			if p.isAtEnd() {
				return
			}
			if isStatementStart(p.current().Type) {
				return
			}
			continue
		}
		// If we hit a DEDENT, we're exiting a block, good place to stop
		if p.check(lexer.TOKEN_DEDENT) {
			return
		}
		p.advance()
	}
}

// isStatementStart returns true if the token type can begin a statement.
func isStatementStart(t lexer.TokenType) bool {
	switch t {
	case lexer.TOKEN_SAY, lexer.TOKEN_IF, lexer.TOKEN_FOR, lexer.TOKEN_WHILE,
		lexer.TOKEN_TO, lexer.TOKEN_GIVE, lexer.TOKEN_USE, lexer.TOKEN_TEST,
		lexer.TOKEN_EXPECT, lexer.TOKEN_DESCRIBE, lexer.TOKEN_TRY,
		lexer.TOKEN_BREAK, lexer.TOKEN_CONTINUE, lexer.TOKEN_FROM,
		lexer.TOKEN_MATCH, lexer.TOKEN_DEFINE, lexer.TOKEN_COMPONENT,
		lexer.TOKEN_MOUNT, lexer.TOKEN_IDENT:
		return true
	}
	return false
}

// --- Hint generation helpers ---

// hintForMissingColon returns a hint about missing colons.
func hintForMissingColon() string {
	return "did you forget a ':' at the end?"
}

// hintForMissingIs returns a hint about using 'is' for assignment.
func hintForMissingIs() string {
	return "use 'is' to assign values, e.g., 'x is 5'"
}

// hintForUnknownKeyword tries to suggest a known keyword based on Levenshtein distance.
func hintForUnknownKeyword(word string) string {
	knownKeywords := []string{
		"say", "if", "otherwise", "for", "each", "in", "to", "give", "back",
		"and", "or", "not", "greater", "less", "than", "equal", "contains",
		"while", "use", "test", "expect", "describe", "new", "my", "await",
		"as", "try", "fails", "extends", "from", "with", "nothing", "break",
		"continue", "match", "when", "define", "component", "state", "mount",
		"set", "repeat", "is", "are", "yes", "no", "true", "false",
	}

	bestMatch := ""
	bestDist := 3 // only suggest if distance is <= 2

	for _, kw := range knownKeywords {
		d := levenshtein(strings.ToLower(word), kw)
		if d < bestDist {
			bestDist = d
			bestMatch = kw
		}
	}

	if bestMatch != "" {
		return fmt.Sprintf("did you mean '%s'?", bestMatch)
	}
	return ""
}

// hintForIndentation returns a hint about indentation mismatch.
func hintForIndentation(expected, found int) string {
	return fmt.Sprintf("expected %d spaces of indentation, found %d", expected, found)
}

// Levenshtein computes the edit distance between two strings.
// Exported for testing.
func Levenshtein(a, b string) int {
	return levenshtein(a, b)
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	la := len(a)
	lb := len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Use two rows instead of full matrix
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			curr[j] = min3(del, ins, sub)
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// --- Error codes ---
const (
	// Syntax errors E001-E099
	ErrUnexpectedToken     = "E001"
	ErrMissingColon        = "E002"
	ErrMissingNewline      = "E003"
	ErrMissingIndent       = "E004"
	ErrUnknownKeyword      = "E005"
	ErrMissingExpression   = "E006"
	ErrUnclosedParen       = "E007"
	ErrUnclosedBracket     = "E008"
	ErrUnclosedBrace       = "E009"
	ErrMissingIs           = "E010"
	ErrIndentationMismatch = "E011"
	ErrUnexpectedEOF       = "E012"
	ErrInvalidNumber       = "E013"
	ErrBadAssignment       = "E014"
	ErrMissingBody         = "E015"
)

// ParseWithRecovery parses tokens with error recovery, collecting all errors
// instead of stopping at the first one. Returns the program (possibly partial)
// and all collected errors.
func ParseWithRecovery(tokens []lexer.Token) (*ast.Program, []ParseErrorDetail) {
	p := New(tokens)
	p.recovery = NewErrorRecovery()

	stmts := []ast.Statement{}
	for !p.isAtEnd() {
		p.skipNewlines()
		if p.isAtEnd() {
			break
		}
		if p.recovery.TooManyErrors() {
			p.recovery.addError(p.current().Line, p.current().Column,
				ErrUnexpectedEOF, "too many errors, stopping", "fix the errors above and try again")
			break
		}

		stmt := p.parseStatementRecover()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}

	program := &ast.Program{Statements: stmts}
	return program, p.recovery.Errors()
}

// parseStatementRecover wraps parseStatement with panic recovery.
func (p *Parser) parseStatementRecover() (stmt ast.Statement) {
	defer func() {
		if r := recover(); r != nil {
			if pe, ok := r.(*ParseError); ok {
				// Convert the panic error to a detailed error
				hint := ""
				code := ErrUnexpectedToken

				// Try to generate a useful hint
				msg := pe.Message
				if strings.Contains(msg, "expected :") || strings.Contains(msg, "expected colon") {
					hint = hintForMissingColon()
					code = ErrMissingColon
				} else if strings.Contains(msg, "expected is") {
					hint = hintForMissingIs()
					code = ErrMissingIs
				} else if strings.Contains(msg, "didn't expect") {
					tok := p.current()
					if tok.Type == lexer.TOKEN_IDENT {
						kwHint := hintForUnknownKeyword(tok.Value)
						if kwHint != "" {
							hint = kwHint
							code = ErrUnknownKeyword
						}
					}
				}

				if p.recovery != nil {
					p.recovery.addError(pe.Line, 0, code, pe.Message, hint)
				}

				// Synchronize to continue parsing
				p.synchronize()
				stmt = nil
			} else {
				panic(r) // re-panic for non-parse errors
			}
		}
	}()

	return p.parseStatement()
}
