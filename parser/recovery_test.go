package parser

import (
	"quill/lexer"
	"testing"
)

func tokenize(input string) []lexer.Token {
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		// Return whatever tokens we got
		return nil
	}
	return tokens
}

func TestRecoveryCollectsMultipleErrors(t *testing.T) {
	// Two bad statements in a row
	source := `say
say "ok"
say
say "also ok"
`
	tokens := tokenize(source)
	if tokens == nil {
		t.Fatal("tokenization failed")
	}

	program, errs := ParseWithRecovery(tokens)
	if len(errs) == 0 {
		t.Fatal("expected errors but got none")
	}
	if len(errs) < 2 {
		t.Errorf("expected at least 2 errors, got %d", len(errs))
	}
	// Should still get some valid statements
	if program == nil {
		t.Fatal("expected partial program, got nil")
	}
}

func TestRecoverySynchronizesToNextStatement(t *testing.T) {
	// First statement is broken, second is valid
	source := `say
x is 5
`
	tokens := tokenize(source)
	if tokens == nil {
		t.Fatal("tokenization failed")
	}

	program, errs := ParseWithRecovery(tokens)
	if len(errs) == 0 {
		t.Fatal("expected at least one error")
	}
	if program == nil {
		t.Fatal("expected partial program, got nil")
	}
	// The valid "x is 5" should still be parsed
	if len(program.Statements) < 1 {
		t.Errorf("expected at least 1 valid statement, got %d", len(program.Statements))
	}
}

func TestRecoveryValidProgramHasNoErrors(t *testing.T) {
	source := `x is 5
say x
`
	tokens := tokenize(source)
	if tokens == nil {
		t.Fatal("tokenization failed")
	}

	program, errs := ParseWithRecovery(tokens)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
	if program == nil {
		t.Fatal("expected program, got nil")
	}
	if len(program.Statements) != 2 {
		t.Errorf("expected 2 statements, got %d", len(program.Statements))
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a, b string
		dist int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"say", "sa", 1},
		{"say", "day", 1},
		{"say", "says", 1},
		{"say", "xyz", 3},
		{"while", "whle", 1},
		{"describe", "descirbe", 2},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			d := Levenshtein(tt.a, tt.b)
			if d != tt.dist {
				t.Errorf("levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, d, tt.dist)
			}
		})
	}
}

func TestHintForUnknownKeyword(t *testing.T) {
	tests := []struct {
		word     string
		contains string
	}{
		{"sya", "say"},
		{"whle", "while"},
		{"fi", "if"},
		{"sey", "say"},
		{"tset", "set"},
		{"xyzabc", ""}, // too far from any keyword
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			hint := hintForUnknownKeyword(tt.word)
			if tt.contains == "" {
				if hint != "" {
					t.Errorf("expected no hint for %q, got %q", tt.word, hint)
				}
			} else {
				if hint == "" {
					t.Errorf("expected hint containing %q for %q, got empty", tt.contains, tt.word)
				} else if !contains(hint, tt.contains) {
					t.Errorf("expected hint containing %q for %q, got %q", tt.contains, tt.word, hint)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestErrorRecoveryMaxErrors(t *testing.T) {
	r := NewErrorRecovery()
	r.maxErrors = 3

	r.addError(1, 1, "E001", "error 1", "")
	r.addError(2, 1, "E001", "error 2", "")
	r.addError(3, 1, "E001", "error 3", "")

	if !r.TooManyErrors() {
		t.Error("expected TooManyErrors to be true after reaching max")
	}
	if r.ErrorCount() != 3 {
		t.Errorf("expected 3 errors, got %d", r.ErrorCount())
	}
}

func TestIsStatementStart(t *testing.T) {
	starters := []lexer.TokenType{
		lexer.TOKEN_SAY, lexer.TOKEN_IF, lexer.TOKEN_FOR, lexer.TOKEN_WHILE,
		lexer.TOKEN_TO, lexer.TOKEN_IDENT, lexer.TOKEN_MATCH,
	}
	for _, tt := range starters {
		if !isStatementStart(tt) {
			t.Errorf("expected %v to be a statement start", tt)
		}
	}

	nonStarters := []lexer.TokenType{
		lexer.TOKEN_PLUS, lexer.TOKEN_MINUS, lexer.TOKEN_STRING,
		lexer.TOKEN_NUMBER, lexer.TOKEN_COLON,
	}
	for _, tt := range nonStarters {
		if isStatementStart(tt) {
			t.Errorf("expected %v to NOT be a statement start", tt)
		}
	}
}
