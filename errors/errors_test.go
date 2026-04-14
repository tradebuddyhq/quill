package errors

import (
	"strings"
	"testing"
)

func TestQuillErrorFormatting(t *testing.T) {
	source := "name is \"Alice\"\nsay name\nfoo bar baz"
	err := NewError(2, 0, source, "unexpected token", "did you mean 'say'?")

	msg := err.Error()

	if !strings.Contains(msg, "Error on line 2") {
		t.Errorf("expected 'Error on line 2' in output, got:\n%s", msg)
	}
	if !strings.Contains(msg, "unexpected token") {
		t.Errorf("expected error message in output, got:\n%s", msg)
	}
	if !strings.Contains(msg, "say name") {
		t.Errorf("expected source line in output, got:\n%s", msg)
	}
	if !strings.Contains(msg, "Hint: did you mean 'say'?") {
		t.Errorf("expected hint in output, got:\n%s", msg)
	}
}

func TestQuillErrorNoHint(t *testing.T) {
	source := "x is 5"
	err := NewError(1, 3, source, "type mismatch", "")
	msg := err.Error()

	if strings.Contains(msg, "Hint") {
		t.Errorf("expected no hint, got:\n%s", msg)
	}
	if !strings.Contains(msg, "^ here") {
		t.Errorf("expected column pointer, got:\n%s", msg)
	}
}

func TestQuillErrorContextLines(t *testing.T) {
	source := "line1\nline2\nline3\nline4\nline5"
	err := NewError(3, 0, source, "bad line", "")
	msg := err.Error()

	// Should show line 2 (before), line 3 (error), line 4 (after)
	if !strings.Contains(msg, "2 | line2") {
		t.Errorf("expected context line before, got:\n%s", msg)
	}
	if !strings.Contains(msg, "3 | line3") {
		t.Errorf("expected error line, got:\n%s", msg)
	}
	if !strings.Contains(msg, "4 | line4") {
		t.Errorf("expected context line after, got:\n%s", msg)
	}
}

func TestNewError(t *testing.T) {
	err := NewError(5, 10, "source", "msg", "hint")
	if err.Line != 5 {
		t.Errorf("Line = %d, want 5", err.Line)
	}
	if err.Column != 10 {
		t.Errorf("Column = %d, want 10", err.Column)
	}
	if err.Message != "msg" {
		t.Errorf("Message = %q, want %q", err.Message, "msg")
	}
	if err.Hint != "hint" {
		t.Errorf("Hint = %q, want %q", err.Hint, "hint")
	}
}
