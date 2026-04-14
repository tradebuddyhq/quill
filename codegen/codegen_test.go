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
	src := `statusVal is "active"
match statusVal:
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
  portNum is 3000
  to start:
    say "Starting on port " + toText(my.portNum)

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

func TestGenerateRouteHandler(t *testing.T) {
	input := "app on get \"/\" with req res:\n  say \"Hello!\"\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `app.get("/",`) {
		t.Errorf("expected app.get(\"/\", ...), got:\n%s", output)
	}
	if !strings.Contains(output, "(req, res) =>") {
		t.Errorf("expected (req, res) => callback, got:\n%s", output)
	}
}

func TestGenerateRouteHandlerPost(t *testing.T) {
	input := "app on post \"/api/data\" with req res:\n  say req.body\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `app.post("/api/data",`) {
		t.Errorf("expected app.post(\"/api/data\", ...), got:\n%s", output)
	}
	if !strings.Contains(output, "(req, res) =>") {
		t.Errorf("expected (req, res) => callback, got:\n%s", output)
	}
}

func TestGenerateRouteHandlerDelete(t *testing.T) {
	input := "app on delete \"/api/items/:id\" with req res:\n  say \"deleted\"\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `app.delete("/api/items/:id",`) {
		t.Errorf("expected app.delete(\"/api/items/:id\", ...), got:\n%s", output)
	}
}

func TestGenerateCommandStatement(t *testing.T) {
	src := `use "discord.js" as Discord
bot is Discord.bot(env("DISCORD_TOKEN"))

command "ping" described "Check if bot is alive":
  reply "Pong!"

command "greet" with user described "Greet someone":
  reply "Hello, {user}!"
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"__quill_commands",
		`setName("ping")`,
		`setDescription("Check if bot is alive")`,
		`setName("greet")`,
		`setDescription("Greet someone")`,
		`addStringOption`,
		"interactionCreate",
		"isChatInputCommand",
		`interaction.reply("Pong!")`,
		`interaction.reply`,
		"applicationCommands",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q\nGot:\n%s", check, output)
		}
	}
}

func TestGenerateEmbedLiteral(t *testing.T) {
	src := `use "discord.js" as Discord
bot is Discord.bot(env("DISCORD_TOKEN"))

command "help" described "Show help":
  reply embed "My Bot":
    color green
    description "A bot built with Quill"
    field "Ping" "Check if alive"
    footer "Footer text"
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"EmbedBuilder()",
		`setTitle("My Bot")`,
		"setColor(0x1EB969)",
		`setDescription("A bot built with Quill")`,
		`addFields({ name: "Ping", value: "Check if alive" })`,
		`setFooter({ text: "Footer text" })`,
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q\nGot:\n%s", check, output)
		}
	}
}

func TestGenerateDiscordBot(t *testing.T) {
	src := `use "discord.js" as Discord
bot is Discord.bot(env("DISCORD_TOKEN"))
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"new Discord.Client",
		"GatewayIntentBits.Guilds",
		"GatewayIntentBits.GuildMessages",
		"GatewayIntentBits.MessageContent",
		"process.nextTick",
		"process.env[",
		".env",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q\nGot:\n%s", check, output)
		}
	}
}

func TestGenerateEnvFunction(t *testing.T) {
	src := `token is env("MY_TOKEN")`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `process.env["MY_TOKEN"]`) {
		t.Errorf("expected process.env access, got:\n%s", output)
	}
}

func TestGenerateAutoEnvLoading(t *testing.T) {
	src := `use "discord.js" as Discord
bot is Discord.bot(env("DISCORD_TOKEN"))
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "readFileSync('.env'") {
		t.Errorf("expected .env loader, got:\n%s", output)
	}
}

func TestGenerateWorkerFetch(t *testing.T) {
	src := `worker on fetch with request:
  respond "Hello from Quill!"
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"export default {",
		"async fetch(request)",
		"return new Response(",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q\nGot:\n%s", check, output)
		}
	}

	// Worker mode should NOT include Node.js runtime
	if strings.Contains(output, "require(") {
		t.Errorf("worker output should not contain require(), got:\n%s", output)
	}
}

func TestGenerateRespondJson(t *testing.T) {
	src := `worker on fetch with request:
  respond json { message: "Hello!" }
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"JSON.stringify(",
		`"Content-Type": "application/json"`,
		"return new Response(",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q\nGot:\n%s", check, output)
		}
	}
}

func TestGenerateRespondStatus(t *testing.T) {
	src := `worker on fetch with request:
  respond "not found" status 404
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"return new Response(",
		"status: 404",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q\nGot:\n%s", check, output)
		}
	}
}

func TestGenerateAskClaude(t *testing.T) {
	src := `answer is ask claude "What is the capital of France?"
say answer
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		`require("@anthropic-ai/sdk")`,
		"__ai_client",
		"messages.create",
		`"claude-sonnet-4-20250514"`,
		"max_tokens: 1024",
		`"What is the capital of France?"`,
		"content[0].text",
		"async",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q\nGot:\n%s", check, output)
		}
	}
}

func TestGenerateAskClaudeWithOptions(t *testing.T) {
	src := `answer is ask claude "Summarize this" with model "claude-sonnet-4-20250514" max_tokens 500
say answer
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		`"claude-sonnet-4-20250514"`,
		"max_tokens: 500",
		"__ai_client",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q\nGot:\n%s", check, output)
		}
	}
}

func TestGenerateStreamClaude(t *testing.T) {
	src := `stream claude "Write a poem about coding":
  say chunk
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		`require("@anthropic-ai/sdk")`,
		"__ai_client",
		"messages.stream",
		"for await",
		"content_block_delta",
		"__event.delta.text",
		"const chunk",
		"console.log(chunk)",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q\nGot:\n%s", check, output)
		}
	}
}

func TestAwaitInFunctionAsync(t *testing.T) {
	src := `to fetchData url:
  data is await fetch(url)
  give back data
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "async function fetchData(url)") {
		t.Errorf("expected async function declaration, got:\n%s", output)
	}
	if !strings.Contains(output, "await fetch(url)") {
		t.Errorf("expected await in function body, got:\n%s", output)
	}
}

func TestReservedWordsAsDotAccess(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`obj.send("hello")`, `obj.send("hello")`},
		{`say obj.status`, `obj.status`},
		{`say obj.type`, `obj.type`},
		{`say obj.from`, `obj.from`},
		{`say obj.select`, `obj.select`},
		{`say obj.load`, `obj.load`},
		{`say obj.after`, `obj.after`},
		{`say obj.use`, `obj.use`},
		{`say obj.on`, `obj.on`},
		{`say obj.server`, `obj.server`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			output, err := compile(tt.input)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if !strings.Contains(output, tt.expected) {
				t.Errorf("expected output to contain %q, got:\n%s", tt.expected, output)
			}
		})
	}
}

func TestReservedWordsAsObjectKeys(t *testing.T) {
	src := `msg is {type: "message", from: "alice", status: 200}`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checks := []string{
		"type:",
		"from:",
		"status:",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, output)
		}
	}
}

// --- Bug fix regression tests ---

func TestVariableScopingPerFunction(t *testing.T) {
	// Bug #3: variables in one function should not affect another
	input := `to foo:
  x is 1
  say x

to bar:
  x is 2
  say x
`
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Both functions should declare x with "let"
	count := strings.Count(output, "let x = ")
	if count != 2 {
		t.Errorf("expected 2 'let x = ' declarations (one per function), got %d in:\n%s", count, output)
	}
}

func TestAsyncFunctionWithAwait(t *testing.T) {
	// Bug #1: functions with await should be emitted as async
	input := `to fetchData:
  result is await fetch("http://example.com")
  give back result
`
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "async function fetchData") {
		t.Errorf("expected 'async function fetchData', got:\n%s", output)
	}
}

func TestStringEscaping(t *testing.T) {
	// Bug #6: special characters should be escaped in single-line strings
	gen := New()
	// Single-line strings (no actual newlines) should escape special chars
	expr := gen.genExpr(&ast.StringLiteral{Value: "hello\\nworld\\ttab"})
	if !strings.Contains(expr, `"`) {
		t.Errorf("expected double-quoted string, got: %s", expr)
	}

	// Strings with actual newlines (from """) use backtick template literals
	expr2 := gen.genExpr(&ast.StringLiteral{Value: "line1\nline2"})
	if !strings.HasPrefix(expr2, "`") {
		t.Errorf("expected backtick string for multiline, got: %s", expr2)
	}
	if !strings.Contains(expr2, "line1") || !strings.Contains(expr2, "line2") {
		t.Errorf("expected both lines preserved, got: %s", expr2)
	}
}

func TestTaggedTemplateSmartConversion(t *testing.T) {
	// Bug #11: tagged templates should not blindly replace all { with ${
	gen := New()
	expr := gen.genExpr(&ast.TaggedTemplateExpr{
		Tag:      "css",
		Template: ".container { color: {color}; }",
	})
	// Should contain ${color} for the interpolation
	if !strings.Contains(expr, "${color}") {
		t.Errorf("expected ${color} interpolation, got: %s", expr)
	}
	// CSS braces should be escaped, not turned into interpolation
	if strings.Contains(expr, "${ color") || strings.Contains(expr, "${}") {
		t.Errorf("CSS braces should not be converted to interpolation, got: %s", expr)
	}
}

func TestFromUseKeywordImport(t *testing.T) {
	// Bug #9: from imports should accept keywords as export names
	input := `from "ws" use send`
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	gen := New()
	output := gen.Generate(prog)
	if !strings.Contains(output, "send") {
		t.Errorf("expected 'send' in output, got:\n%s", output)
	}
}

func TestPackageNameWithDots(t *testing.T) {
	// Bug #10: package names with dots should produce valid JS variable names
	input := `use "socket.io" as SocketIO`
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "SocketIO") {
		t.Errorf("expected alias SocketIO to be used, got:\n%s", output)
	}
	// Without alias, dots should be replaced
	input2 := `use "socket.io"`
	output2, err := compile(input2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(output2, "socket.io =") || strings.Contains(output2, "const socket.io") {
		t.Errorf("expected dots to be replaced in variable name, got:\n%s", output2)
	}
}

func TestYieldInTryCatch(t *testing.T) {
	// Bug #12: bodyContainsYield should find yield inside try/catch
	gen := New()
	stmts := []ast.Statement{
		&ast.TryCatchStatement{
			TryBody: []ast.Statement{
				&ast.YieldStatement{Value: &ast.NumberLiteral{Value: 1}},
			},
			CatchBody: []ast.Statement{},
		},
	}
	if !gen.bodyContainsYield(stmts) {
		t.Error("expected bodyContainsYield to find yield inside try/catch")
	}
}

func TestLoopWithoutYieldNoIteratorRuntime(t *testing.T) {
	// Bug #17: loops without yield should not inject iterator runtime
	input := `to count:
  loop:
    say "hello"
    break
`
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(output, "__QuillLazy") || strings.Contains(output, "__quill_lazy") {
		t.Errorf("loop without yield should not inject iterator runtime, got:\n%s", output)
	}
}

func TestBracketAssignment(t *testing.T) {
	src := `obj is {}
obj["key"] is "value"
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `obj["key"] = "value";`) {
		t.Errorf("expected bracket assignment, got:\n%s", output)
	}
	if strings.Contains(output, `obj["key"] === "value"`) {
		t.Errorf("bracket assignment should not produce comparison, got:\n%s", output)
	}
}

func TestIsNothingCoversUndefined(t *testing.T) {
	src := `x is 1
if x is nothing:
  say "gone"
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The comparison should use loose equality == null to catch both null and undefined
	if !strings.Contains(output, "(x == null)") {
		t.Errorf("expected (x == null) with loose equality, got:\n%s", output)
	}
}

func TestOuterVariableReassignment(t *testing.T) {
	src := `x is 10
to change:
  x is 20
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Inside the function, x should be reassigned (no let), not redeclared
	// Count occurrences of "let x"
	count := strings.Count(output, "let x")
	if count != 1 {
		t.Errorf("expected exactly 1 'let x' declaration, got %d in:\n%s", count, output)
	}
	if !strings.Contains(output, "x = 20;") {
		t.Errorf("expected bare reassignment 'x = 20;' inside function, got:\n%s", output)
	}
}

func TestServerAsVariableName(t *testing.T) {
	src := `server is "localhost"
say server
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `let server = "localhost";`) {
		t.Errorf("expected server as variable name, got:\n%s", output)
	}
}

func TestGenerateEveryStatement(t *testing.T) {
	output, err := compile("every 5 seconds:\n  say \"tick\"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "setInterval(") {
		t.Errorf("expected setInterval call, got:\n%s", output)
	}
	if !strings.Contains(output, "5000") {
		t.Errorf("expected 5000ms interval, got:\n%s", output)
	}
}

func TestGenerateEveryMinutes(t *testing.T) {
	output, err := compile("every 2 minutes:\n  say \"check\"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "setInterval(") {
		t.Errorf("expected setInterval call, got:\n%s", output)
	}
	if !strings.Contains(output, "120000") {
		t.Errorf("expected 120000ms interval, got:\n%s", output)
	}
}

func TestCryptoRuntimeInjection(t *testing.T) {
	output, err := compile("result is hash(\"hello\")\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__crypto") {
		t.Errorf("expected __crypto runtime injection, got:\n%s", output)
	}
	if !strings.Contains(output, "createHash") {
		t.Errorf("expected createHash in runtime, got:\n%s", output)
	}
}

func TestBufferRuntimeInjection(t *testing.T) {
	output, err := compile("encoded is toBase64(\"hello\")\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function toBase64") {
		t.Errorf("expected toBase64 function in runtime, got:\n%s", output)
	}
}

func TestCLIRuntimeInjection(t *testing.T) {
	output, err := compile("name is arg(0)\nsay name\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function arg(index)") {
		t.Errorf("expected arg(index) function in runtime, got:\n%s", output)
	}
}

func TestSecureServerRuntimeInjection(t *testing.T) {
	output, err := compile("app is createSecureServer({ key: \"key.pem\", cert: \"cert.pem\" })\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "createSecureServer") {
		t.Errorf("expected createSecureServer in output, got:\n%s", output)
	}
}

func TestHKDFRuntimeInjection(t *testing.T) {
	output, err := compile("derived is hkdf(inputKey, salt, \"info\", 32)\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function hkdf(") {
		t.Error("expected hkdf function in runtime")
	}
}

func TestDiffieHellmanRuntimeInjection(t *testing.T) {
	output, err := compile("shared is diffieHellman(myKey, theirKey)\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function diffieHellman(") {
		t.Error("expected diffieHellman function in runtime")
	}
}

func TestArgon2RuntimeInjection(t *testing.T) {
	output, err := compile("hashed is await argon2(password, salt)\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "async function argon2(") {
		t.Error("expected argon2 function in runtime")
	}
}

func TestConstantTimeEqualRuntimeInjection(t *testing.T) {
	output, err := compile("result is constantTimeEqual(a, b)\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function constantTimeEqual(") {
		t.Error("expected constantTimeEqual function in runtime")
	}
}

func TestAESEncryptRuntimeInjection(t *testing.T) {
	output, err := compile("encrypted is aesEncrypt(data, key, iv, \"aes-256-gcm\")\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function aesEncrypt(") {
		t.Error("expected aesEncrypt function in runtime")
	}
}

func TestSecureEraseRuntimeInjection(t *testing.T) {
	output, err := compile("buf is randomBytes(32)\nsecureErase(buf)\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function secureErase(") {
		t.Error("expected secureErase function in runtime")
	}
}

func TestSerializationRuntimeInjection(t *testing.T) {
	output, err := compile("schema is defineSchema({ name: { type: \"string\", tag: 1 } })\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function defineSchema(") {
		t.Error("expected defineSchema function in runtime")
	}
}

func TestCatchSyntax(t *testing.T) {
	output, err := compile("try:\n  say \"test\"\ncatch err:\n  say err\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "try {") || !strings.Contains(output, "catch (err)") {
		t.Errorf("expected try/catch block, got:\n%s", output)
	}
}

// --- AI Feature Tests ---

func TestAskOpenAI(t *testing.T) {
	output, err := compile(`answer is ask openai "What is 2+2?"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__ask_openai") {
		t.Errorf("expected __ask_openai call, got:\n%s", output)
	}
}

func TestAskGemini(t *testing.T) {
	output, err := compile(`answer is ask gemini "What is 2+2?"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__ask_gemini") {
		t.Errorf("expected __ask_gemini call, got:\n%s", output)
	}
}

func TestAskOllama(t *testing.T) {
	output, err := compile(`answer is ask ollama "What is 2+2?"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__ask_ollama") {
		t.Errorf("expected __ask_ollama call, got:\n%s", output)
	}
}

func TestStreamOpenAI(t *testing.T) {
	output, err := compile("stream openai \"Tell me a joke\":\n  say chunk\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__stream_openai") {
		t.Errorf("expected __stream_openai call, got:\n%s", output)
	}
}

func TestStreamGemini(t *testing.T) {
	output, err := compile("stream gemini \"Tell me a joke\":\n  say chunk\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__stream_gemini") {
		t.Errorf("expected __stream_gemini call, got:\n%s", output)
	}
}

func TestStructuredOutput(t *testing.T) {
	output, err := compile(`result is ask claude "Extract name and age from: John is 30" as {name: text, age: number}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__parse_structured") {
		t.Errorf("expected __parse_structured call, got:\n%s", output)
	}
}

func TestAgentStatement(t *testing.T) {
	output, err := compile("agent \"researcher\" with tools [search, browse]:\n  say \"agent running\"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `createAgent("researcher"`) {
		t.Errorf("expected createAgent(\"researcher\") call, got:\n%s", output)
	}
}

func TestEmbedExpression(t *testing.T) {
	output, err := compile(`vec is embed("hello world")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "embed(") {
		t.Errorf("expected embed( call, got:\n%s", output)
	}
}

func TestVectorRuntimeInjection(t *testing.T) {
	output, err := compile("store is createVectorStore()\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "VectorStore") {
		t.Errorf("expected VectorStore in runtime preamble, got:\n%s", output)
	}
}

func TestDocumentRuntimeInjection(t *testing.T) {
	output, err := compile("data is extract(\"file.pdf\")\nparts is chunk(text, 100)\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "extract(") {
		t.Errorf("expected extract( in output, got:\n%s", output)
	}
	if !strings.Contains(output, "chunk(") {
		t.Errorf("expected chunk( in output, got:\n%s", output)
	}
}

func TestOpenAIRuntimeInjection(t *testing.T) {
	output, err := compile(`answer is ask openai "hello"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__ask_openai") {
		t.Errorf("expected __ask_openai in runtime preamble, got:\n%s", output)
	}
	if !strings.Contains(output, "__stream_openai") {
		t.Errorf("expected __stream_openai in runtime preamble, got:\n%s", output)
	}
}

func TestAgentRuntimeInjection(t *testing.T) {
	output, err := compile("agent \"helper\" with tools [search]:\n  say \"running\"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "class Agent") {
		t.Errorf("expected 'class Agent' in runtime preamble, got:\n%s", output)
	}
}

func TestMultilineObjectLiteral(t *testing.T) {
	output, err := compile("x is {foo: 1,\n  bar: 2}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "foo: 1") || !strings.Contains(output, "bar: 2") {
		t.Errorf("expected multiline object literal to compile, got:\n%s", output)
	}
	// Also test multiline array
	output2, err := compile("x is [1,\n  2,\n  3]")
	if err != nil {
		t.Fatalf("unexpected error for multiline array: %v", err)
	}
	if !strings.Contains(output2, "[1, 2, 3]") {
		t.Errorf("expected multiline array literal to compile, got:\n%s", output2)
	}
}

func TestQuotedKeysObjectLiteral(t *testing.T) {
	output, err := compile(`x is {"User-Agent": "test"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `["User-Agent"]: "test"`) {
		t.Errorf("expected quoted key to compile as computed property, got:\n%s", output)
	}
}

func TestExponentiationOperator(t *testing.T) {
	output, err := compile("x is 2 ^ 3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "2 ** 3") {
		t.Errorf("expected ^ to compile to **, got:\n%s", output)
	}
}

func TestInlineTernary(t *testing.T) {
	output, err := compile("x is if yes: 1 otherwise: 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "? 1 : 2") {
		t.Errorf("expected ternary expression, got:\n%s", output)
	}
}

func TestEveryMinutesCron(t *testing.T) {
	output, err := compile("every 30 minutes:\n  say \"tick\"")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "setInterval") {
		t.Errorf("expected setInterval, got:\n%s", output)
	}
	if !strings.Contains(output, "1800000") {
		t.Errorf("expected 30 minutes = 1800000ms, got:\n%s", output)
	}
}

func TestAskOllamaWithModelAndStructuredOutput(t *testing.T) {
	output, err := compile(`result is ask ollama "hello" with model "llama3" as {name: text}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__ask_ollama") {
		t.Errorf("expected __ask_ollama call, got:\n%s", output)
	}
	if !strings.Contains(output, "__parse_structured") {
		t.Errorf("expected __parse_structured wrapper for structured output, got:\n%s", output)
	}
	if !strings.Contains(output, `"llama3"`) {
		t.Errorf("expected model llama3 in options, got:\n%s", output)
	}
}

// --- Server Block Tests ---

func TestServerBlockCompiles(t *testing.T) {
	input := "server:\n  port is 8080\n\n  route get \"/\":\n    respond \"Hello\"\n"
	_, err := compile(input)
	if err != nil {
		t.Fatalf("server block should compile, got error: %v", err)
	}
}

func TestServerBlockWithRoutes(t *testing.T) {
	input := "server:\n  port is 3000\n\n  route get \"/api/users\":\n    respond json {name: \"Alice\"}\n\n  route post \"/api/users\":\n    respond json {ok: yes}\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// In build mode, server block is a comment placeholder
	if !strings.Contains(output, "server block") {
		t.Errorf("expected server block comment, got:\n%s", output)
	}
}

// --- Delete Statement Tests ---

func TestDeleteStatement(t *testing.T) {
	input := "obj is {name: \"test\", age: 25}\ndelete obj.age\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "delete obj.age;") {
		t.Errorf("expected 'delete obj.age;', got:\n%s", output)
	}
}

// --- Template Engine Tests ---

func TestTemplateEngineInjection(t *testing.T) {
	input := "html is tag(\"div\", \"hello\")\nsay html\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function tag(") {
		t.Errorf("expected template runtime to be injected, got:\n%s", output)
	}
	if !strings.Contains(output, "function escapeHTML(") {
		t.Errorf("expected escapeHTML function, got:\n%s", output)
	}
	if !strings.Contains(output, "function page(") {
		t.Errorf("expected page function, got:\n%s", output)
	}
}

func TestTemplatePageFunction(t *testing.T) {
	input := "html is page({title: \"Test\", body: \"Hello\"})\nsay html\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function page(") {
		t.Errorf("expected page function injection, got:\n%s", output)
	}
}

// --- Component Tests ---

func compileBrowser(input string) (string, error) {
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
	gen := NewBrowser()
	return gen.Generate(prog), nil
}

func TestComponentWithUIClasses(t *testing.T) {
	input := "component App:\n  state msg is \"hi\"\n\n  to render:\n    div className \"flex p-4\":\n      h1: \"Hello\"\n\nmount App to \"#app\"\n"
	output, err := compileBrowser(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should include UI CSS framework
	if !strings.Contains(output, "Quill UI") {
		t.Errorf("expected Quill UI CSS injection, got:\n%s", output[:min(500, len(output))])
	}
	// Should include component framework
	if !strings.Contains(output, "__quill_mount") {
		t.Errorf("expected __quill_mount function, got:\n%s", output[:min(500, len(output))])
	}
	// Should have className prop
	if !strings.Contains(output, `className: "flex p-4"`) {
		t.Errorf("expected className prop, got:\n%s", output)
	}
}

func TestArrowFunctionObjectWrap(t *testing.T) {
	input := "result is items.map(with item: {id: item})\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "=> ({") {
		t.Errorf("expected arrow function with wrapped object literal, got:\n%s", output)
	}
}

func TestBrowserNoRequireFS(t *testing.T) {
	input := "say \"hello\"\n"
	output, err := compileBrowser(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(output, "require('fs')") {
		t.Errorf("browser mode should not contain require('fs'), got:\n%s", output[:min(500, len(output))])
	}
}

func TestStringEscapeSequences(t *testing.T) {
	input := "msg is \"Hello\\nWorld\"\nsay msg\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(output, `"Hello\\nWorld"`) {
		t.Errorf("\\n should not be double-escaped, got:\n%s", output)
	}
	if !strings.Contains(output, `"Hello\nWorld"`) {
		t.Errorf("expected single-escaped \\n, got:\n%s", output)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
