package codegen

import (
	"strings"
	"testing"
)

func TestTypeBasedMatchText(t *testing.T) {
	src := `match value:
    when text t:
        say t
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `typeof __match_val === "string"`) {
		t.Errorf("expected typeof string check, got:\n%s", output)
	}
	if !strings.Contains(output, "let t = __match_val;") {
		t.Errorf("expected binding variable 't', got:\n%s", output)
	}
}

func TestTypeBasedMatchNumber(t *testing.T) {
	src := `match value:
    when number n:
        say n
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `typeof __match_val === "number"`) {
		t.Errorf("expected typeof number check, got:\n%s", output)
	}
	if !strings.Contains(output, "let n = __match_val;") {
		t.Errorf("expected binding variable 'n', got:\n%s", output)
	}
}

func TestTypeBasedMatchList(t *testing.T) {
	src := `match value:
    when list l:
        say l
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Array.isArray(__match_val)") {
		t.Errorf("expected Array.isArray check, got:\n%s", output)
	}
	if !strings.Contains(output, "let l = __match_val;") {
		t.Errorf("expected binding variable 'l', got:\n%s", output)
	}
}

func TestTypeBasedMatchNothing(t *testing.T) {
	src := `match value:
    when nothing:
        say "null"
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__match_val == null") {
		t.Errorf("expected null check, got:\n%s", output)
	}
}

func TestGuardClauseInMatch(t *testing.T) {
	src := `match age:
    when 0:
        say "zero"
    otherwise:
        say "other"
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__match_val === 0") {
		t.Errorf("expected value match, got:\n%s", output)
	}
}

func TestGuardClauseWithCondition(t *testing.T) {
	src := `match x:
    when 1 if y is greater than 10:
        say "big"
    otherwise:
        say "small"
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__match_val === 1 && (y > 10)") {
		t.Errorf("expected guard clause with condition, got:\n%s", output)
	}
}

func TestTypeMatchWithGuard(t *testing.T) {
	src := `match value:
    when number n if n is greater than 0:
        say "positive"
    otherwise:
        say "other"
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `typeof __match_val === "number"`) {
		t.Errorf("expected typeof number check, got:\n%s", output)
	}
	if !strings.Contains(output, "let n = __match_val;") {
		t.Errorf("expected binding variable 'n', got:\n%s", output)
	}
	// Should have guard condition
	if !strings.Contains(output, "(n > 0)") {
		t.Errorf("expected guard condition (n > 0), got:\n%s", output)
	}
}

func TestMultipleTypePatternsInMatch(t *testing.T) {
	src := `match value:
    when text t:
        say "text"
    when number n:
        say "number"
    when list l:
        say "list"
    when nothing:
        say "nothing"
    otherwise:
        say "unknown"
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `typeof __match_val === "string"`) {
		t.Errorf("expected string type check")
	}
	if !strings.Contains(output, `typeof __match_val === "number"`) {
		t.Errorf("expected number type check")
	}
	if !strings.Contains(output, `Array.isArray(__match_val)`) {
		t.Errorf("expected array type check")
	}
	if !strings.Contains(output, `__match_val == null`) {
		t.Errorf("expected null check")
	}
	if !strings.Contains(output, "else {") {
		t.Errorf("expected otherwise clause")
	}
}
