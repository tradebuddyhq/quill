package parser

import (
	"quill/ast"
	"quill/lexer"
	"testing"
)

func parseInput(input string) (*ast.Program, error) {
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		return nil, err
	}
	p := New(tokens)
	return p.Parse()
}

// --- Trait Parsing Tests ---

func TestParseTraitDeclaration(t *testing.T) {
	input := `describe trait Printable:
    to toString -> text
`
	prog, err := parseInput(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}

	trait, ok := prog.Statements[0].(*ast.TraitDeclaration)
	if !ok {
		t.Fatalf("expected TraitDeclaration, got %T", prog.Statements[0])
	}
	if trait.Name != "Printable" {
		t.Errorf("expected trait name 'Printable', got %q", trait.Name)
	}
	if len(trait.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(trait.Methods))
	}
	if trait.Methods[0].Name != "toString" {
		t.Errorf("expected method name 'toString', got %q", trait.Methods[0].Name)
	}
	if trait.Methods[0].ReturnType != "text" {
		t.Errorf("expected return type 'text', got %q", trait.Methods[0].ReturnType)
	}
}

func TestParseTraitMultipleMethods(t *testing.T) {
	input := `describe trait Serializable:
    to toJSON -> text
    to fromJSON data as text -> self
`
	prog, err := parseInput(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	trait, ok := prog.Statements[0].(*ast.TraitDeclaration)
	if !ok {
		t.Fatalf("expected TraitDeclaration, got %T", prog.Statements[0])
	}
	if len(trait.Methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(trait.Methods))
	}

	// First method: toJSON
	if trait.Methods[0].Name != "toJSON" {
		t.Errorf("expected 'toJSON', got %q", trait.Methods[0].Name)
	}
	if trait.Methods[0].ReturnType != "text" {
		t.Errorf("expected return type 'text', got %q", trait.Methods[0].ReturnType)
	}

	// Second method: fromJSON with param
	if trait.Methods[1].Name != "fromJSON" {
		t.Errorf("expected 'fromJSON', got %q", trait.Methods[1].Name)
	}
	if len(trait.Methods[1].Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(trait.Methods[1].Params))
	}
	if trait.Methods[1].Params[0].Name != "data" {
		t.Errorf("expected param name 'data', got %q", trait.Methods[1].Params[0].Name)
	}
	if trait.Methods[1].Params[0].TypeHint != "text" {
		t.Errorf("expected param type 'text', got %q", trait.Methods[1].Params[0].TypeHint)
	}
	if trait.Methods[1].ReturnType != "self" {
		t.Errorf("expected return type 'self', got %q", trait.Methods[1].ReturnType)
	}
}

// --- Generic Functions with Where Clauses ---

func TestParseGenericFunctionWithWhere(t *testing.T) {
	input := `to sort items as list of T where T is Comparable -> list of T:
    give back items
`
	prog, err := parseInput(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	funcDef, ok := prog.Statements[0].(*ast.FuncDefinition)
	if !ok {
		t.Fatalf("expected FuncDefinition, got %T", prog.Statements[0])
	}
	if funcDef.Name != "sort" {
		t.Errorf("expected function name 'sort', got %q", funcDef.Name)
	}
	if len(funcDef.TypeParams) != 1 {
		t.Fatalf("expected 1 type param, got %d", len(funcDef.TypeParams))
	}
	if funcDef.TypeParams[0].Name != "T" {
		t.Errorf("expected type param 'T', got %q", funcDef.TypeParams[0].Name)
	}
	if funcDef.TypeParams[0].Constraint != "Comparable" {
		t.Errorf("expected constraint 'Comparable', got %q", funcDef.TypeParams[0].Constraint)
	}
	if funcDef.ReturnType != "list of T" {
		t.Errorf("expected return type 'list of T', got %q", funcDef.ReturnType)
	}
}

func TestParseGenericFunctionWithoutWhere(t *testing.T) {
	input := `to first items as list of T -> T:
    give back items
`
	prog, err := parseInput(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	funcDef, ok := prog.Statements[0].(*ast.FuncDefinition)
	if !ok {
		t.Fatalf("expected FuncDefinition, got %T", prog.Statements[0])
	}
	if funcDef.Name != "first" {
		t.Errorf("expected function name 'first', got %q", funcDef.Name)
	}
	if len(funcDef.TypeParams) != 0 {
		t.Errorf("expected 0 type params (no where clause), got %d", len(funcDef.TypeParams))
	}
}

// --- Object Destructuring ---

func TestParseObjectDestructuring(t *testing.T) {
	input := `{name, age} is person
`
	prog, err := parseInput(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	ds, ok := prog.Statements[0].(*ast.DestructureStatement)
	if !ok {
		t.Fatalf("expected DestructureStatement, got %T", prog.Statements[0])
	}

	objPat, ok := ds.Pattern.(*ast.ObjectPattern)
	if !ok {
		t.Fatalf("expected ObjectPattern, got %T", ds.Pattern)
	}
	if len(objPat.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(objPat.Fields))
	}
	if objPat.Fields[0].Key != "name" {
		t.Errorf("expected field 'name', got %q", objPat.Fields[0].Key)
	}
	if objPat.Fields[1].Key != "age" {
		t.Errorf("expected field 'age', got %q", objPat.Fields[1].Key)
	}

	ident, ok := ds.Value.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier value, got %T", ds.Value)
	}
	if ident.Name != "person" {
		t.Errorf("expected value 'person', got %q", ident.Name)
	}
}

func TestParseObjectDestructuringWithRest(t *testing.T) {
	input := `{host, address, ...rest} is config
`
	prog, err := parseInput(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	ds, ok := prog.Statements[0].(*ast.DestructureStatement)
	if !ok {
		t.Fatalf("expected DestructureStatement, got %T", prog.Statements[0])
	}
	objPat, ok := ds.Pattern.(*ast.ObjectPattern)
	if !ok {
		t.Fatalf("expected ObjectPattern, got %T", ds.Pattern)
	}
	if len(objPat.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(objPat.Fields))
	}
	if objPat.Rest != "rest" {
		t.Errorf("expected rest 'rest', got %q", objPat.Rest)
	}
}

// --- Array Destructuring ---

func TestParseArrayDestructuring(t *testing.T) {
	input := `[first, second] are items
`
	prog, err := parseInput(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	ds, ok := prog.Statements[0].(*ast.DestructureStatement)
	if !ok {
		t.Fatalf("expected DestructureStatement, got %T", prog.Statements[0])
	}
	arrPat, ok := ds.Pattern.(*ast.ArrayPattern)
	if !ok {
		t.Fatalf("expected ArrayPattern, got %T", ds.Pattern)
	}
	if len(arrPat.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arrPat.Elements))
	}
	if arrPat.Elements[0].Name != "first" {
		t.Errorf("expected element 'first', got %q", arrPat.Elements[0].Name)
	}
	if arrPat.Elements[1].Name != "second" {
		t.Errorf("expected element 'second', got %q", arrPat.Elements[1].Name)
	}
}

func TestParseArrayDestructuringWithRest(t *testing.T) {
	input := `[first, ...rest] are items
`
	prog, err := parseInput(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	ds, ok := prog.Statements[0].(*ast.DestructureStatement)
	if !ok {
		t.Fatalf("expected DestructureStatement, got %T", prog.Statements[0])
	}
	arrPat, ok := ds.Pattern.(*ast.ArrayPattern)
	if !ok {
		t.Fatalf("expected ArrayPattern, got %T", ds.Pattern)
	}
	if len(arrPat.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(arrPat.Elements))
	}
	if arrPat.Rest != "rest" {
		t.Errorf("expected rest 'rest', got %q", arrPat.Rest)
	}
}

// --- Nested Destructuring ---

func TestParseNestedDestructuring(t *testing.T) {
	input := `{user: {name, email}} is response
`
	prog, err := parseInput(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	ds, ok := prog.Statements[0].(*ast.DestructureStatement)
	if !ok {
		t.Fatalf("expected DestructureStatement, got %T", prog.Statements[0])
	}
	objPat, ok := ds.Pattern.(*ast.ObjectPattern)
	if !ok {
		t.Fatalf("expected ObjectPattern, got %T", ds.Pattern)
	}
	if len(objPat.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(objPat.Fields))
	}
	if objPat.Fields[0].Key != "user" {
		t.Errorf("expected field key 'user', got %q", objPat.Fields[0].Key)
	}

	// Check nested pattern
	nested, ok := objPat.Fields[0].Nested.(*ast.ObjectPattern)
	if !ok {
		t.Fatalf("expected nested ObjectPattern, got %T", objPat.Fields[0].Nested)
	}
	if len(nested.Fields) != 2 {
		t.Fatalf("expected 2 nested fields, got %d", len(nested.Fields))
	}
	if nested.Fields[0].Key != "name" {
		t.Errorf("expected nested field 'name', got %q", nested.Fields[0].Key)
	}
	if nested.Fields[1].Key != "email" {
		t.Errorf("expected nested field 'email', got %q", nested.Fields[1].Key)
	}
}

// --- Type Check Expression ---

func TestParseTypeCheckExpression(t *testing.T) {
	input := `if x is text:
    say "it is text"
`
	prog, err := parseInput(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	ifStmt, ok := prog.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", prog.Statements[0])
	}

	tc, ok := ifStmt.Condition.(*ast.TypeCheckExpr)
	if !ok {
		t.Fatalf("expected TypeCheckExpr condition, got %T", ifStmt.Condition)
	}
	if tc.TypeName != "text" {
		t.Errorf("expected type name 'text', got %q", tc.TypeName)
	}
	ident, ok := tc.Expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier in TypeCheckExpr, got %T", tc.Expr)
	}
	if ident.Name != "x" {
		t.Errorf("expected identifier 'x', got %q", ident.Name)
	}
}
