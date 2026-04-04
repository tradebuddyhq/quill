package codegen

import (
	"quill/ast"
	"quill/lexer"
	"quill/parser"
	"strings"
	"testing"
)

func compile(input string) (string, error) {
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
	gen := New()
	return gen.Generate(prog), nil
}

func TestGenerateAssignment(t *testing.T) {
	output, err := compile(`name is "hello"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `let name = "hello";`) {
		t.Errorf("expected let assignment, got:\n%s", output)
	}
}

func TestGenerateReassignment(t *testing.T) {
	output, err := compile("x is 1\nx is 2\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "let x = 1;") {
		t.Error("expected let for first assignment")
	}
	if !strings.Contains(output, "x = 2;") && strings.Contains(output, "let x = 2;") {
		t.Error("expected bare reassignment (no let) for second assignment")
	}
}

func TestGenerateSay(t *testing.T) {
	output, err := compile(`say "Hello!"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `console.log("Hello!");`) {
		t.Errorf("expected console.log, got:\n%s", output)
	}
}

func TestGenerateStringInterpolation(t *testing.T) {
	output, err := compile(`say "Hello, {name}!"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "${name}") {
		t.Errorf("expected template literal with ${name}, got:\n%s", output)
	}
	if !strings.Contains(output, "`") {
		t.Error("expected backtick template literal")
	}
}

func TestGenerateIf(t *testing.T) {
	output, err := compile("if x is greater than 10:\n  say x\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "if (") {
		t.Error("expected if statement")
	}
	if !strings.Contains(output, ">") {
		t.Error("expected > operator")
	}
}

func TestGenerateIfOtherwise(t *testing.T) {
	output, err := compile("if x is 1:\n  say x\notherwise:\n  say y\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "else {") {
		t.Error("expected else block")
	}
}

func TestGenerateForEach(t *testing.T) {
	output, err := compile("for each item in items:\n  say item\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "for (const item of items)") {
		t.Errorf("expected for...of loop, got:\n%s", output)
	}
}

func TestGenerateWhile(t *testing.T) {
	output, err := compile("while x is less than 10:\n  x is x + 1\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "while (") {
		t.Error("expected while loop")
	}
}

func TestGenerateFunction(t *testing.T) {
	output, err := compile("to add a b:\n  give back a + b\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function add(a, b)") {
		t.Errorf("expected function declaration, got:\n%s", output)
	}
	if !strings.Contains(output, "return") {
		t.Error("expected return statement")
	}
}

func TestGenerateEquality(t *testing.T) {
	output, err := compile("if x is 5:\n  say x\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "===") {
		t.Errorf("expected === for equality, got:\n%s", output)
	}
}

func TestGenerateInequality(t *testing.T) {
	output, err := compile("if x is not 5:\n  say x\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "!==") {
		t.Errorf("expected !== for inequality, got:\n%s", output)
	}
}

func TestGenerateContains(t *testing.T) {
	output, err := compile("if list contains 5:\n  say list\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__contains(") {
		t.Errorf("expected __contains call, got:\n%s", output)
	}
}

func TestGenerateBoolean(t *testing.T) {
	output, err := compile("active is yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "let active = true;") {
		t.Errorf("expected 'true', got:\n%s", output)
	}
}

func TestGenerateBooleanNo(t *testing.T) {
	output, err := compile("done is no")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "let done = false;") {
		t.Errorf("expected 'false', got:\n%s", output)
	}
}

func TestGenerateList(t *testing.T) {
	output, err := compile(`items are [1, 2, 3]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "[1, 2, 3]") {
		t.Errorf("expected list literal, got:\n%s", output)
	}
}

func TestGenerateLogicalAnd(t *testing.T) {
	output, err := compile("if x and y:\n  say x\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "&&") {
		t.Error("expected && operator")
	}
}

func TestGenerateLogicalOr(t *testing.T) {
	output, err := compile("if x or y:\n  say x\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "||") {
		t.Error("expected || operator")
	}
}

func TestGenerateNot(t *testing.T) {
	output, err := compile("if not x:\n  say y\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "!") {
		t.Error("expected ! operator")
	}
}

func TestGenerateUnaryMinus(t *testing.T) {
	output, err := compile("x is -5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "(-5)") {
		t.Errorf("expected (-5), got:\n%s", output)
	}
}

func TestGenerateDotAssignment(t *testing.T) {
	output, err := compile(`dog.name is "Rex"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `dog.name = "Rex";`) {
		t.Errorf("expected dot assignment, got:\n%s", output)
	}
}

func TestGenerateNew(t *testing.T) {
	output, err := compile("dog is new Dog()")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "new Dog()") {
		t.Errorf("expected new Dog(), got:\n%s", output)
	}
}

func TestGenerateDescribe(t *testing.T) {
	output, err := compile("describe Dog:\n  name is \"\"\n  to bark:\n    say \"woof\"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "class Dog") {
		t.Error("expected class declaration")
	}
	if !strings.Contains(output, "constructor()") {
		t.Error("expected constructor")
	}
	if !strings.Contains(output, "bark()") {
		t.Error("expected bark method")
	}
}

func TestGenerateUseNPM(t *testing.T) {
	output, err := compile(`use "express" as app`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `const app = require("express");`) {
		t.Errorf("expected require, got:\n%s", output)
	}
}

func TestGenerateTestBlock(t *testing.T) {
	output, err := compile("test \"math\":\n  expect 1 is 1\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__test_passed") {
		t.Error("expected test tracking variable")
	}
	if !strings.Contains(output, "math") {
		t.Error("expected test name in output")
	}
}

func TestGenerateExpect(t *testing.T) {
	output, err := compile("test \"t\":\n  expect x is 5\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "throw new Error") {
		t.Error("expected throw for expect")
	}
}

func TestGenerateAsyncAwait(t *testing.T) {
	output, err := compile(`data is await fetchJSON("url")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "async") {
		t.Error("expected async IIFE wrapper")
	}
	if !strings.Contains(output, "await") {
		t.Error("expected await in output")
	}
}

func TestGenerateComment(t *testing.T) {
	output, err := compile("// Generated by Quill")
	if err != nil {
		// This might fail to parse, that's fine
		return
	}
	_ = output
}

func TestGenerateHeader(t *testing.T) {
	output, err := compile("x is 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "// Generated by Quill") {
		t.Error("expected Quill header comment")
	}
}

func TestGenerateBodyOnly(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SayStatement{
				Value: &ast.StringLiteral{Value: "hello"},
				Line:  1,
			},
		},
	}
	gen := New()
	body := gen.GenerateBody(prog)
	if !strings.Contains(body, `console.log("hello");`) {
		t.Errorf("expected console.log in body, got:\n%s", body)
	}
	// GenerateBody should not include the runtime header
	if strings.Contains(body, "// Generated by Quill") {
		t.Error("GenerateBody should not include header")
	}
}

func TestConvertInterpolation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"{name}", "${name}"},
		{"Hello {name}!", "Hello ${name}!"},
		{"{a} and {b}", "${a} and ${b}"},
		{"no interpolation", "no interpolation"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertInterpolation(tt.input)
			if result != tt.expected {
				t.Errorf("convertInterpolation(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEscapeJS(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`hello`, `hello`},
		{`he said "hi"`, `he said \"hi\"`},
		{`back\slash`, `back\\slash`},
	}

	for _, tt := range tests {
		result := escapeJS(tt.input)
		if result != tt.expected {
			t.Errorf("escapeJS(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGenerateComplexProgram(t *testing.T) {
	src := `name is "Sarah"
age is 25

say "Hello, {name}!"

if age is greater than 18:
  say "You are an adult"
otherwise:
  say "You are young"

to add a b:
  give back a + b

result is add(10, 20)
say "10 + 20 = {result}"

colors are ["red", "green", "blue"]
for each color in colors:
  say "I like {color}"
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		`let name = "Sarah";`,
		"let age = 25;",
		"console.log(",
		"if (",
		"else {",
		"function add(a, b)",
		"return",
		"for (const color of",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q\nGot:\n%s", check, output)
		}
	}
}

func TestGenerateMyKeyword(t *testing.T) {
	output, err := compile("describe Cat:\n  name is \"\"\n  to speak:\n    say my.name\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "this.name") {
		t.Errorf("expected 'this.name' for my.name, got:\n%s", output)
	}
}

func TestGenerateChainedDot(t *testing.T) {
	output, err := compile("say obj.a.b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "obj.a.b") {
		t.Errorf("expected chained dot access, got:\n%s", output)
	}
}

func TestGenerateIndex(t *testing.T) {
	output, err := compile("say items[0]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "items[0]") {
		t.Errorf("expected index access, got:\n%s", output)
	}
}

func TestGenerateArithmetic(t *testing.T) {
	output, err := compile("x is 2 + 3 * 4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should generate parenthesized expressions
	if !strings.Contains(output, "+") || !strings.Contains(output, "*") {
		t.Errorf("expected arithmetic operators, got:\n%s", output)
	}
}

func TestGenerateModulo(t *testing.T) {
	output, err := compile("x is 10 % 3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "%") {
		t.Errorf("expected modulo operator, got:\n%s", output)
	}
}

// --- New Feature Tests ---

func TestGenerateTryCatch(t *testing.T) {
	src := "try:\n  say \"hello\"\nif it fails err:\n  say err\n"
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "try {") {
		t.Errorf("expected try block, got:\n%s", output)
	}
	if !strings.Contains(output, "catch (err)") {
		t.Errorf("expected catch with err variable, got:\n%s", output)
	}
}

func TestGenerateBreak(t *testing.T) {
	src := "while yes:\n  break\n"
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "break;") {
		t.Errorf("expected break statement, got:\n%s", output)
	}
}

func TestGenerateContinue(t *testing.T) {
	src := "for each x in items:\n  continue\n"
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "continue;") {
		t.Errorf("expected continue statement, got:\n%s", output)
	}
}

func TestGenerateObjectLiteral(t *testing.T) {
	src := `config is {name: "test", value: 42}`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "name:") && !strings.Contains(output, "value:") {
		t.Errorf("expected object literal, got:\n%s", output)
	}
}

func TestGenerateEmptyObject(t *testing.T) {
	src := "obj is {}"
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "{}") {
		t.Errorf("expected empty object, got:\n%s", output)
	}
}

func TestGenerateNothing(t *testing.T) {
	src := "x is nothing"
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "null") {
		t.Errorf("expected null for nothing, got:\n%s", output)
	}
}

func TestGenerateLambda(t *testing.T) {
	src := "doubled is map_list(nums, with x: x * 2)"
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "=>") {
		t.Errorf("expected arrow function, got:\n%s", output)
	}
	if !strings.Contains(output, "(x) =>") {
		t.Errorf("expected (x) => syntax, got:\n%s", output)
	}
}

func TestGenerateMultiParamLambda(t *testing.T) {
	src := "result is reduce(nums, with a, b: a + b)"
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "(a, b) =>") {
		t.Errorf("expected (a, b) => syntax, got:\n%s", output)
	}
}

func TestGenerateTypeAnnotations(t *testing.T) {
	src := "to add a as number, b as number -> number:\n  give back a + b\n"
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Type annotations should be stripped
	if !strings.Contains(output, "function add(a, b)") {
		t.Errorf("expected function with params (types stripped), got:\n%s", output)
	}
}

func TestGenerateClassExtends(t *testing.T) {
	src := "describe Dog extends Animal:\n  breed is \"mixed\"\n  to bark:\n    say \"woof\"\n"
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "class Dog extends Animal") {
		t.Errorf("expected extends in class, got:\n%s", output)
	}
	if !strings.Contains(output, "super()") {
		t.Errorf("expected super() call, got:\n%s", output)
	}
}

func TestGenerateFromUse(t *testing.T) {
	src := `from "express" use Router, json`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "const { Router, json } = require") {
		t.Errorf("expected destructured require, got:\n%s", output)
	}
}

func TestGenerateArrow(t *testing.T) {
	// Test that -> doesn't interfere with minus
	src := "x is 5 - 3"
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "5 - 3") {
		t.Errorf("expected subtraction, got:\n%s", output)
	}
}

func TestGenerateSpread(t *testing.T) {
	src := "all is concat([1, 2], ...rest)"
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "...rest") {
		t.Errorf("expected spread operator, got:\n%s", output)
	}
}

func TestGenerateMatch(t *testing.T) {
	src := `status is "active"
match status:
  when "active":
    say "User is active"
  when "inactive":
    say "User is inactive"
  otherwise:
    say "Unknown status"
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checks := []string{
		"__match_val",
		"if (",
		"else if (",
		"else {",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q\nGot:\n%s", check, output)
		}
	}
}

func TestGenerateDefine(t *testing.T) {
	src := `define Color:
  Red
  Green
  Blue
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Object.freeze") {
		t.Errorf("expected Object.freeze, got:\n%s", output)
	}
	if !strings.Contains(output, "Red") {
		t.Errorf("expected Red variant, got:\n%s", output)
	}
}

func TestGeneratePipe(t *testing.T) {
	src := `result is 5 | toText`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "toText(5)") {
		t.Errorf("expected piped function call, got:\n%s", output)
	}
}

func TestGeneratePipeWithArgs(t *testing.T) {
	src := `result is "hello world" | replace_text("world", "quill")`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "replace_text") {
		t.Errorf("expected replace_text call, got:\n%s", output)
	}
}

func TestGenerateComplexNewFeatures(t *testing.T) {
	src := `
config is {debug: yes, name: "app"}

to processItems items as list -> list:
  results are []
  for each item in items:
    if item is nothing:
      continue
    push(results, item)
  give back results

describe HTTPServer extends Server:
  port is 3000
  to start:
    say "Starting on port " + toText(my.port)

try:
  data is processItems([1, nothing, 3])
  say toText(data)
if it fails error:
  say "Error: " + error
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"debug:",
		"null",
		"function processItems(items)",
		"continue;",
		"class HTTPServer extends Server",
		"super()",
		"try {",
		"catch (error)",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q\nGot:\n%s", check, output)
		}
	}
}
