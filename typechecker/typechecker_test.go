package typechecker

import (
	"quill/lexer"
	"quill/parser"
	"testing"
)

func check(input string) []TypeDiagnostic {
	l := lexer.New(input)
	tokens, _ := l.Tokenize()
	p := parser.New(tokens)
	prog, _ := p.Parse()
	tc := New()
	return tc.Check(prog)
}

func TestNoErrors(t *testing.T) {
	diags := check(`x is 5
y is "hello"
say y
`)
	for _, d := range diags {
		if d.Severity == "error" {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}
}

func TestInferNumberType(t *testing.T) {
	diags := check(`x is 5
y is x + 3
`)
	for _, d := range diags {
		if d.Severity == "error" {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}
}

func TestInferStringType(t *testing.T) {
	diags := check(`name is "hello"
greeting is "hi " + name
`)
	for _, d := range diags {
		if d.Severity == "error" {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}
}

func TestFunctionReturnTypeCheck(t *testing.T) {
	diags := check(`to add a as number, b as number -> number:
  give back "oops"
`)
	hasError := false
	for _, d := range diags {
		if d.Severity == "error" {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected type error for returning text from number function")
	}
}

func TestFunctionArgTypeCheck(t *testing.T) {
	diags := check(`to double x as number -> number:
  give back x * 2
double("hello")
`)
	hasError := false
	for _, d := range diags {
		if d.Severity == "error" {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected type error for passing text to number param")
	}
}

func TestConditionWarning(t *testing.T) {
	diags := check(`if 5:
  say "hello"
`)
	hasWarning := false
	for _, d := range diags {
		if d.Severity == "warning" {
			hasWarning = true
		}
	}
	if !hasWarning {
		t.Error("expected warning for non-boolean condition")
	}
}

func TestListTypeInference(t *testing.T) {
	diags := check(`nums are [1, 2, 3]
say nums
`)
	for _, d := range diags {
		if d.Severity == "error" {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}
}

func TestStdlibReturnTypes(t *testing.T) {
	tc := New()
	numFuncs := []string{"length", "round", "floor", "ceil", "abs"}
	for _, fn := range numFuncs {
		ret := tc.stdlibReturnType(fn)
		if ret.Name != "number" {
			t.Errorf("expected %s to return number, got %s", fn, ret.Name)
		}
	}

	textFuncs := []string{"toText", "trim", "upper", "lower"}
	for _, fn := range textFuncs {
		ret := tc.stdlibReturnType(fn)
		if ret.Name != "text" {
			t.Errorf("expected %s to return text, got %s", fn, ret.Name)
		}
	}

	boolFuncs := []string{"includes", "startsWith", "endsWith", "isText", "isNumber"}
	for _, fn := range boolFuncs {
		ret := tc.stdlibReturnType(fn)
		if ret.Name != "boolean" {
			t.Errorf("expected %s to return boolean, got %s", fn, ret.Name)
		}
	}
}

func TestTypeCompatibility(t *testing.T) {
	tc := New()

	// any is compatible with everything
	if !tc.typeCompatible(Type{Name: "any"}, Type{Name: "number"}) {
		t.Error("any should be compatible with number")
	}

	// nothing is compatible with any type
	if !tc.typeCompatible(Type{Name: "text"}, Type{Name: "nothing"}) {
		t.Error("nothing should be compatible with text")
	}

	// same types are compatible
	if !tc.typeCompatible(Type{Name: "number"}, Type{Name: "number"}) {
		t.Error("number should be compatible with number")
	}

	// different types are not compatible
	if tc.typeCompatible(Type{Name: "number"}, Type{Name: "text"}) {
		t.Error("number should not be compatible with text")
	}
}

func TestParseType(t *testing.T) {
	simple := parseType("number")
	if simple.Name != "number" || simple.Inner != "" {
		t.Errorf("expected number, got %v", simple)
	}

	generic := parseType("list of number")
	if generic.Name != "list" || generic.Inner != "number" {
		t.Errorf("expected list of number, got %v", generic)
	}
}

func TestUnionTypeParsing(t *testing.T) {
	ut := parseType("number | text")
	if len(ut.Union) != 2 {
		t.Fatalf("expected 2 union members, got %d", len(ut.Union))
	}
	if ut.Union[0].Name != "number" {
		t.Errorf("expected first union member to be number, got %s", ut.Union[0].Name)
	}
	if ut.Union[1].Name != "text" {
		t.Errorf("expected second union member to be text, got %s", ut.Union[1].Name)
	}

	// String representation
	if ut.String() != "number | text" {
		t.Errorf("expected 'number | text', got %q", ut.String())
	}
}

func TestNullableTypeParsing(t *testing.T) {
	nt := parseType("?number")
	if nt.Name != "number" {
		t.Errorf("expected name to be number, got %s", nt.Name)
	}
	if !nt.Nullable {
		t.Error("expected nullable to be true")
	}
	if nt.String() != "?number" {
		t.Errorf("expected '?number', got %q", nt.String())
	}

	// Nullable generic
	ng := parseType("?list of number")
	if ng.Name != "list" || ng.Inner != "number" || !ng.Nullable {
		t.Errorf("expected ?list of number, got %v", ng)
	}
}

func TestUnionTypeCompatibility(t *testing.T) {
	tc := New()

	unionType := parseType("number | text")

	// number is compatible with number | text
	if !tc.typeCompatible(unionType, Type{Name: "number"}) {
		t.Error("number should be compatible with number | text")
	}

	// text is compatible with number | text
	if !tc.typeCompatible(unionType, Type{Name: "text"}) {
		t.Error("text should be compatible with number | text")
	}

	// boolean is NOT compatible with number | text
	if tc.typeCompatible(unionType, Type{Name: "boolean"}) {
		t.Error("boolean should not be compatible with number | text")
	}
}

func TestNullableTypeCompatibility(t *testing.T) {
	tc := New()

	nullableNum := parseType("?number")

	// number is compatible with ?number
	if !tc.typeCompatible(nullableNum, Type{Name: "number"}) {
		t.Error("number should be compatible with ?number")
	}

	// nothing is compatible with ?number
	if !tc.typeCompatible(nullableNum, Type{Name: "nothing"}) {
		t.Error("nothing should be compatible with ?number")
	}

	// text is NOT compatible with ?number
	if tc.typeCompatible(nullableNum, Type{Name: "text"}) {
		t.Error("text should not be compatible with ?number")
	}
}

func TestHasErrors(t *testing.T) {
	diags := []TypeDiagnostic{
		{Line: 1, Severity: "warning", Message: "test"},
	}
	if HasErrors(diags) {
		t.Error("expected no errors")
	}

	diags = append(diags, TypeDiagnostic{Line: 2, Severity: "error", Message: "test"})
	if !HasErrors(diags) {
		t.Error("expected errors")
	}
}
