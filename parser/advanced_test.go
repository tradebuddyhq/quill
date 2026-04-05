package parser

import (
	"quill/ast"
	"testing"
)

// --- Feature 4: Computed Properties ---

func TestComputedPropertyParsing(t *testing.T) {
	prog, err := parse(`obj is {[key]: "Alice"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	assign, ok := prog.Statements[0].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", prog.Statements[0])
	}
	obj, ok := assign.Value.(*ast.ObjectLiteral)
	if !ok {
		t.Fatalf("expected ObjectLiteral, got %T", assign.Value)
	}
	if len(obj.ComputedProperties) != 1 {
		t.Fatalf("expected 1 computed property, got %d", len(obj.ComputedProperties))
	}
	keyExpr, ok := obj.ComputedProperties[0].KeyExpr.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier key expr, got %T", obj.ComputedProperties[0].KeyExpr)
	}
	if keyExpr.Name != "key" {
		t.Errorf("expected key expr 'key', got %q", keyExpr.Name)
	}
	valStr, ok := obj.ComputedProperties[0].Value.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expected StringLiteral value, got %T", obj.ComputedProperties[0].Value)
	}
	if valStr.Value != "Alice" {
		t.Errorf("expected value 'Alice', got %q", valStr.Value)
	}
}

func TestMixedObjectPropertiesParsing(t *testing.T) {
	prog, err := parse(`obj is {name: "Bob", [key]: "Alice"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	obj := assign.Value.(*ast.ObjectLiteral)
	if len(obj.Keys) != 1 || obj.Keys[0] != "name" {
		t.Errorf("expected 1 regular key 'name', got %v", obj.Keys)
	}
	if len(obj.ComputedProperties) != 1 {
		t.Errorf("expected 1 computed property, got %d", len(obj.ComputedProperties))
	}
}

func TestDynamicPropertyAccess(t *testing.T) {
	prog, err := parse(`x is obj[key]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	idx, ok := assign.Value.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr, got %T", assign.Value)
	}
	objIdent, ok := idx.Object.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier for object, got %T", idx.Object)
	}
	if objIdent.Name != "obj" {
		t.Errorf("expected 'obj', got %q", objIdent.Name)
	}
	keyIdent, ok := idx.Index.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier for index, got %T", idx.Index)
	}
	if keyIdent.Name != "key" {
		t.Errorf("expected 'key', got %q", keyIdent.Name)
	}
}

// --- Feature 5: Tagged Templates ---

func TestTaggedTemplateParsing(t *testing.T) {
	prog, err := parse("sql is query`SELECT * FROM users WHERE age > {minAge}`")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	assign, ok := prog.Statements[0].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", prog.Statements[0])
	}
	tagged, ok := assign.Value.(*ast.TaggedTemplateExpr)
	if !ok {
		t.Fatalf("expected TaggedTemplateExpr, got %T", assign.Value)
	}
	if tagged.Tag != "query" {
		t.Errorf("expected tag 'query', got %q", tagged.Tag)
	}
	if len(tagged.Expressions) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(tagged.Expressions))
	}
	ident, ok := tagged.Expressions[0].(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier expression, got %T", tagged.Expressions[0])
	}
	if ident.Name != "minAge" {
		t.Errorf("expected 'minAge', got %q", ident.Name)
	}
}

func TestTaggedTemplateMultipleExpressions(t *testing.T) {
	prog, err := parse("result is html`<div>{title}</div><p>{content}</p>`")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	tagged := assign.Value.(*ast.TaggedTemplateExpr)
	if tagged.Tag != "html" {
		t.Errorf("expected tag 'html', got %q", tagged.Tag)
	}
	if len(tagged.Expressions) != 2 {
		t.Fatalf("expected 2 expressions, got %d", len(tagged.Expressions))
	}
}

// --- Feature 6: Module Visibility ---

func TestPrivatePublicParsingInDescribe(t *testing.T) {
	input := `describe User:
    private password is ""
    public name is ""
    private to hashPassword:
        give back "hashed"
    public to getName:
        give back "test"
`
	prog, err := parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	desc, ok := prog.Statements[0].(*ast.DescribeStatement)
	if !ok {
		t.Fatalf("expected DescribeStatement, got %T", prog.Statements[0])
	}
	if desc.Name != "User" {
		t.Errorf("expected class name 'User', got %q", desc.Name)
	}
	// Check properties
	if len(desc.Properties) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(desc.Properties))
	}
	if len(desc.PropertyVisibilities) != 2 {
		t.Fatalf("expected 2 property visibilities, got %d", len(desc.PropertyVisibilities))
	}
	if desc.PropertyVisibilities[0] != "private" {
		t.Errorf("expected first property visibility 'private', got %q", desc.PropertyVisibilities[0])
	}
	if desc.PropertyVisibilities[1] != "public" {
		t.Errorf("expected second property visibility 'public', got %q", desc.PropertyVisibilities[1])
	}
	// Check methods
	if len(desc.Methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(desc.Methods))
	}
	if len(desc.MethodVisibilities) != 2 {
		t.Fatalf("expected 2 method visibilities, got %d", len(desc.MethodVisibilities))
	}
	if desc.MethodVisibilities[0] != "private" {
		t.Errorf("expected first method visibility 'private', got %q", desc.MethodVisibilities[0])
	}
	if desc.MethodVisibilities[1] != "public" {
		t.Errorf("expected second method visibility 'public', got %q", desc.MethodVisibilities[1])
	}
}

func TestDescribeWithoutVisibility(t *testing.T) {
	input := `describe Animal:
    name is "dog"
    to speak:
        give back "woof"
`
	prog, err := parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	desc := prog.Statements[0].(*ast.DescribeStatement)
	if len(desc.PropertyVisibilities) != 1 || desc.PropertyVisibilities[0] != "" {
		t.Errorf("expected empty visibility for property, got %v", desc.PropertyVisibilities)
	}
	if len(desc.MethodVisibilities) != 1 || desc.MethodVisibilities[0] != "" {
		t.Errorf("expected empty visibility for method, got %v", desc.MethodVisibilities)
	}
}

// --- Feature 7: Enums with Methods ---

func TestEnumWithMethodsParsing(t *testing.T) {
	input := `define HttpStatus:
    OK is 200
    NotFound is 404
    ServerError is 500
    to isSuccess:
        give back "yes"
`
	prog, err := parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	def, ok := prog.Statements[0].(*ast.DefineStatement)
	if !ok {
		t.Fatalf("expected DefineStatement, got %T", prog.Statements[0])
	}
	if def.Name != "HttpStatus" {
		t.Errorf("expected enum name 'HttpStatus', got %q", def.Name)
	}
	if len(def.Variants) != 3 {
		t.Fatalf("expected 3 variants, got %d", len(def.Variants))
	}
	// Check variant values
	if def.Variants[0].Name != "OK" {
		t.Errorf("expected variant 'OK', got %q", def.Variants[0].Name)
	}
	if def.Variants[0].Value == nil {
		t.Fatal("expected variant OK to have a value")
	}
	okVal, ok := def.Variants[0].Value.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral for OK value, got %T", def.Variants[0].Value)
	}
	if okVal.Value != 200 {
		t.Errorf("expected OK value 200, got %v", okVal.Value)
	}
	// Check methods
	if len(def.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(def.Methods))
	}
	if def.Methods[0].Name != "isSuccess" {
		t.Errorf("expected method name 'isSuccess', got %q", def.Methods[0].Name)
	}
}

func TestEnumWithoutMethods(t *testing.T) {
	input := `define Color:
    Red
    Green
    Blue
`
	prog, err := parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	def := prog.Statements[0].(*ast.DefineStatement)
	if len(def.Variants) != 3 {
		t.Fatalf("expected 3 variants, got %d", len(def.Variants))
	}
	if len(def.Methods) != 0 {
		t.Errorf("expected 0 methods, got %d", len(def.Methods))
	}
	for _, v := range def.Variants {
		if v.Value != nil {
			t.Errorf("expected no value for variant %q", v.Name)
		}
	}
}
