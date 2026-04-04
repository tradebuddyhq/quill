package codegen

import (
	"quill/lexer"
	"quill/parser"
	"strings"
	"testing"
)

func compileLLVM(input string) (string, error) {
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		return "", err
	}
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		return "", err
	}
	gen := NewLLVM()
	return gen.Generate(prog), nil
}

func TestLLVM_SimpleAssignAndSay(t *testing.T) {
	input := "name is \"World\"\nsay name\n"
	ir, err := compileLLVM(input)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	checks := []string{
		"define i32 @main()",
		"@.str.",
		"alloca i8*",
		"store i8*",
		"call void @__quill_print_str",
		"ret i32 0",
	}
	for _, c := range checks {
		if !strings.Contains(ir, c) {
			t.Errorf("expected IR to contain %q\nGot:\n%s", c, ir)
		}
	}
}

func TestLLVM_ArithmeticExpressions(t *testing.T) {
	input := "x is 10 + 20\nsay x\n"
	ir, err := compileLLVM(input)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	checks := []string{
		"fadd double",
		"alloca double",
		"store double",
		"call void @__quill_print_num",
	}
	for _, c := range checks {
		if !strings.Contains(ir, c) {
			t.Errorf("expected IR to contain %q\nGot:\n%s", c, ir)
		}
	}
}

func TestLLVM_StringLiteralsAndConcat(t *testing.T) {
	input := "greeting is \"Hello, \" + \"World\"\nsay greeting\n"
	ir, err := compileLLVM(input)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	checks := []string{
		"@.str.",
		"private constant",
		"call i8* @__quill_str_concat",
		"call void @__quill_print_str",
	}
	for _, c := range checks {
		if !strings.Contains(ir, c) {
			t.Errorf("expected IR to contain %q\nGot:\n%s", c, ir)
		}
	}
}

func TestLLVM_IfElseBranching(t *testing.T) {
	input := "x is 10\nif x is greater than 5:\n  say \"big\"\notherwise:\n  say \"small\"\n"
	ir, err := compileLLVM(input)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	checks := []string{
		"fcmp ogt double",
		"br i1",
		"if.then",
		"if.else",
		"if.end",
	}
	for _, c := range checks {
		if !strings.Contains(ir, c) {
			t.Errorf("expected IR to contain %q\nGot:\n%s", c, ir)
		}
	}
}

func TestLLVM_WhileLoop(t *testing.T) {
	input := "x is 0\nwhile x is less than 10:\n  x is x + 1\n"
	ir, err := compileLLVM(input)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	checks := []string{
		"while.cond",
		"while.body",
		"while.end",
		"br i1",
		"fcmp olt double",
	}
	for _, c := range checks {
		if !strings.Contains(ir, c) {
			t.Errorf("expected IR to contain %q\nGot:\n%s", c, ir)
		}
	}
}

func TestLLVM_FunctionDefinitionAndCall(t *testing.T) {
	input := "to add a b:\n  give back a + b\n\nsay add(3, 4)\n"
	ir, err := compileLLVM(input)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	checks := []string{
		"define",
		"@add",
		"fadd double",
		"ret double",
		"call void @add",
	}
	for _, c := range checks {
		if !strings.Contains(ir, c) {
			t.Errorf("expected IR to contain %q\nGot:\n%s", c, ir)
		}
	}
}

func TestLLVM_BooleanExpressions(t *testing.T) {
	input := "a is true\nb is false\nc is a and b\nsay c\n"
	ir, err := compileLLVM(input)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	checks := []string{
		"alloca i1",
		"store i1",
		"and i1",
		"call void @__quill_print_bool",
	}
	for _, c := range checks {
		if !strings.Contains(ir, c) {
			t.Errorf("expected IR to contain %q\nGot:\n%s", c, ir)
		}
	}
}

func TestLLVM_ModuleStructure(t *testing.T) {
	input := "say \"hello\"\n"
	ir, err := compileLLVM(input)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	checks := []string{
		"; ModuleID = 'quill_module'",
		"source_filename",
		"declare i32 @printf",
		"declare i8* @malloc",
		"declare void @free",
		"declare i64 @strlen",
		"define void @__quill_print_num",
		"define void @__quill_print_str",
		"define void @__quill_print_bool",
		"define i8* @__quill_str_concat",
		"define i8* @__quill_num_to_str",
		"define i32 @main()",
	}
	for _, c := range checks {
		if !strings.Contains(ir, c) {
			t.Errorf("expected IR to contain %q\nGot:\n%s", c, ir)
		}
	}
}

func TestLLVM_NumberSay(t *testing.T) {
	input := "say 42\n"
	ir, err := compileLLVM(input)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	if !strings.Contains(ir, "call void @__quill_print_num(double") {
		t.Errorf("expected numeric print call\nGot:\n%s", ir)
	}
}

func TestLLVM_StringNumberConcat(t *testing.T) {
	input := "say \"value: \" + 42\n"
	ir, err := compileLLVM(input)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	checks := []string{
		"call i8* @__quill_num_to_str",
		"call i8* @__quill_str_concat",
	}
	for _, c := range checks {
		if !strings.Contains(ir, c) {
			t.Errorf("expected IR to contain %q\nGot:\n%s", c, ir)
		}
	}
}

func TestLLVM_Comparison(t *testing.T) {
	input := "x is 5\nif x is 5:\n  say \"yes\"\n"
	ir, err := compileLLVM(input)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	if !strings.Contains(ir, "fcmp oeq double") {
		t.Errorf("expected fcmp oeq\nGot:\n%s", ir)
	}
}
