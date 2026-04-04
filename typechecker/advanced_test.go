package typechecker

import (
	"quill/ast"
	"quill/codegen"
	"quill/lexer"
	"quill/parser"
	"strings"
	"testing"
)

// --- Trait Tests ---

func TestTraitDeclarationRegistered(t *testing.T) {
	input := `describe trait Printable:
    to toString -> text
`
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tc := New()
	diags := tc.Check(prog)

	// Should have no errors
	for _, d := range diags {
		if d.Severity == "error" {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}

	// Trait should be registered
	if _, ok := tc.traits["Printable"]; !ok {
		t.Error("expected trait 'Printable' to be registered")
	}

	// Check that the trait has the correct method
	trait := tc.traits["Printable"]
	if len(trait.Methods) != 1 {
		t.Fatalf("expected 1 method in trait, got %d", len(trait.Methods))
	}
	if trait.Methods[0].Name != "toString" {
		t.Errorf("expected method 'toString', got %q", trait.Methods[0].Name)
	}
	if trait.Methods[0].ReturnType.Name != "text" {
		t.Errorf("expected return type 'text', got %q", trait.Methods[0].ReturnType.Name)
	}
}

func TestTraitMultipleMethods(t *testing.T) {
	input := `describe trait Serializable:
    to toJSON -> text
    to fromJSON data as text -> self
`
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tc := New()
	tc.Check(prog)

	trait, ok := tc.traits["Serializable"]
	if !ok {
		t.Fatal("expected trait 'Serializable' to be registered")
	}
	if len(trait.Methods) != 2 {
		t.Fatalf("expected 2 methods in trait, got %d", len(trait.Methods))
	}
	if trait.Methods[0].Name != "toJSON" {
		t.Errorf("expected first method 'toJSON', got %q", trait.Methods[0].Name)
	}
	if trait.Methods[1].Name != "fromJSON" {
		t.Errorf("expected second method 'fromJSON', got %q", trait.Methods[1].Name)
	}
}

// --- Generic Constraint Tests ---

func TestGenericFunctionWithWhereClause(t *testing.T) {
	input := `to sort items as list of T where T is Comparable -> list of T:
    give back items
`
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Verify the function has type params
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	funcDef, ok := prog.Statements[0].(*ast.FuncDefinition)
	if !ok {
		t.Fatal("expected FuncDefinition statement")
	}
	if len(funcDef.TypeParams) != 1 {
		t.Fatalf("expected 1 type param, got %d", len(funcDef.TypeParams))
	}
	if funcDef.TypeParams[0].Name != "T" {
		t.Errorf("expected type param name 'T', got %q", funcDef.TypeParams[0].Name)
	}
	if funcDef.TypeParams[0].Constraint != "Comparable" {
		t.Errorf("expected constraint 'Comparable', got %q", funcDef.TypeParams[0].Constraint)
	}

	tc := New()
	diags := tc.Check(prog)
	for _, d := range diags {
		if d.Severity == "error" {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}
}

// --- Exhaustive Match Checking Tests ---

func TestExhaustiveMatchMissingVariant(t *testing.T) {
	input := `define Shape:
    Circle
    Square
    Triangle

shape is Shape.Circle
match shape:
    when Circle:
        say "circle"
    when Square:
        say "square"
`
	diags := check(input)
	hasExhaustiveError := false
	for _, d := range diags {
		if d.Severity == "error" && strings.Contains(d.Message, "Non-exhaustive match") && strings.Contains(d.Message, "Triangle") {
			hasExhaustiveError = true
		}
	}
	if !hasExhaustiveError {
		t.Error("expected non-exhaustive match error for missing 'Triangle' variant")
	}
}

func TestExhaustiveMatchAllVariantsCovered(t *testing.T) {
	input := `define Color:
    Red
    Green
    Blue

color is Color.Red
match color:
    when Red:
        say "red"
    when Green:
        say "green"
    when Blue:
        say "blue"
`
	diags := check(input)
	for _, d := range diags {
		if d.Severity == "error" && strings.Contains(d.Message, "Non-exhaustive") {
			t.Errorf("unexpected exhaustive match error: %s", d.Message)
		}
	}
}

func TestExhaustiveMatchWithOtherwise(t *testing.T) {
	input := `define Direction:
    North
    South
    East
    West

dir is Direction.North
match dir:
    when North:
        say "north"
    otherwise:
        say "other"
`
	diags := check(input)
	for _, d := range diags {
		if d.Severity == "error" && strings.Contains(d.Message, "Non-exhaustive") {
			t.Errorf("should not have exhaustive error when 'otherwise' is present: %s", d.Message)
		}
	}
}

// --- Type Narrowing Tests ---

func TestTypeNarrowingInIfBlock(t *testing.T) {
	input := `to process x as any -> text:
    if x is text:
        give back x
    give back "not text"
`
	diags := check(input)
	// Inside the if block, x should be narrowed to text, so returning x should not
	// produce an error about returning a non-text type
	for _, d := range diags {
		if d.Severity == "error" {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}
}

func TestTypeCheckExpressionInfersBooleanType(t *testing.T) {
	tc := New()
	expr := &ast.TypeCheckExpr{
		Expr:     &ast.Identifier{Name: "x"},
		TypeName: "text",
	}
	result := tc.inferType(expr)
	if result.Name != "boolean" {
		t.Errorf("expected TypeCheckExpr to infer boolean, got %s", result.Name)
	}
}

// --- Destructuring Codegen Tests ---

func TestDestructureObjectCodegen(t *testing.T) {
	input := `{name, age} is person
`
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := codegen.New()
	output := gen.Generate(prog)
	if !strings.Contains(output, "const {name, age} = person;") {
		t.Errorf("expected object destructuring output, got:\n%s", output)
	}
}

func TestDestructureArrayCodegen(t *testing.T) {
	input := `[first, second] are items
`
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := codegen.New()
	output := gen.Generate(prog)
	if !strings.Contains(output, "const [first, second] = items;") {
		t.Errorf("expected array destructuring output, got:\n%s", output)
	}
}

func TestDestructureObjectWithRestCodegen(t *testing.T) {
	input := `{host, address, ...rest} is config
`
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := codegen.New()
	output := gen.Generate(prog)
	if !strings.Contains(output, "const {host, address, ...rest} = config;") {
		t.Errorf("expected object destructuring with rest, got:\n%s", output)
	}
}

func TestDestructureArrayWithRestCodegen(t *testing.T) {
	input := `[first, ...rest] are items
`
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := codegen.New()
	output := gen.Generate(prog)
	if !strings.Contains(output, "const [first, ...rest] = items;") {
		t.Errorf("expected array destructuring with rest, got:\n%s", output)
	}
}

func TestDestructureNestedObjectCodegen(t *testing.T) {
	input := `{user: {name, email}} is response
`
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := codegen.New()
	output := gen.Generate(prog)
	if !strings.Contains(output, "const {user: {name, email}} = response;") {
		t.Errorf("expected nested destructuring output, got:\n%s", output)
	}
}

// --- Trait Codegen Test ---

func TestTraitCodegen(t *testing.T) {
	input := `describe trait Printable:
    to toString -> text
`
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := codegen.New()
	output := gen.Generate(prog)
	if !strings.Contains(output, "__implements_Printable") {
		t.Errorf("expected __implements_Printable function in output, got:\n%s", output)
	}
	if !strings.Contains(output, "typeof obj.toString") {
		t.Errorf("expected typeof check for toString method, got:\n%s", output)
	}
}
