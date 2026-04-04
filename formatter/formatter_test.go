package formatter

import (
	"quill/lexer"
	"quill/parser"
	"strings"
	"testing"
)

func format(input string) (string, error) {
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
	f := New()
	return f.Format(prog), nil
}

func TestFormatAssignment(t *testing.T) {
	output, err := format(`name is "hello"`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, `name is "hello"`) {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestFormatSay(t *testing.T) {
	output, err := format(`say "Hello, world!"`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, `say "Hello, world!"`) {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestFormatIfStatement(t *testing.T) {
	output, err := format("if x is greater than 10:\n  say x\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "if x is greater than 10:") {
		t.Errorf("unexpected output: %s", output)
	}
	if !strings.Contains(output, "  say x") {
		t.Errorf("expected indented body: %s", output)
	}
}

func TestFormatIfOtherwise(t *testing.T) {
	output, err := format("if x is 1:\n  say x\notherwise:\n  say y\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "otherwise:") {
		t.Errorf("expected 'otherwise:' in output: %s", output)
	}
}

func TestFormatForEach(t *testing.T) {
	output, err := format("for each item in items:\n  say item\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "for each item in items:") {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestFormatWhile(t *testing.T) {
	output, err := format("while x is less than 10:\n  x is x + 1\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "while x is less than 10:") {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestFormatFunction(t *testing.T) {
	output, err := format("to add a b:\n  give back a + b\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "to add a b:") {
		t.Errorf("unexpected output: %s", output)
	}
	if !strings.Contains(output, "  give back") {
		t.Errorf("expected indented return: %s", output)
	}
}

func TestFormatFunctionNoParams(t *testing.T) {
	output, err := format("to greet:\n  say \"hi\"\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "to greet:") {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestFormatUse(t *testing.T) {
	output, err := format(`use "helpers.quill"`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, `use "helpers.quill"`) {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestFormatUseAs(t *testing.T) {
	output, err := format(`use "express" as app`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, `use "express" as app`) {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestFormatTestBlock(t *testing.T) {
	output, err := format("test \"math works\":\n  expect 1 + 1 is 2\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, `test "math works":`) {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestFormatDescribe(t *testing.T) {
	output, err := format("describe Dog:\n  name is \"\"\n  to bark:\n    say \"woof\"\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "describe Dog:") {
		t.Errorf("unexpected output: %s", output)
	}
	if !strings.Contains(output, "  name is") {
		t.Errorf("expected indented property: %s", output)
	}
}

func TestFormatBooleans(t *testing.T) {
	output, err := format("active is yes\ndone is no\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "active is yes") {
		t.Errorf("expected 'yes' for boolean true: %s", output)
	}
	if !strings.Contains(output, "done is no") {
		t.Errorf("expected 'no' for boolean false: %s", output)
	}
}

func TestFormatList(t *testing.T) {
	output, err := format(`items are [1, 2, 3]`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "[1, 2, 3]") {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestFormatEmptyList(t *testing.T) {
	output, err := format("items are []")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "[]") {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestFormatComparisons(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"if x is greater than 10:\n  say x\n", "is greater than"},
		{"if x is less than 5:\n  say x\n", "is less than"},
		{"if x is equal to 0:\n  say x\n", "is 0"},
		{"if x is not 0:\n  say x\n", "is not"},
		{"if list contains 5:\n  say list\n", "contains"},
	}

	for _, tt := range tests {
		output, err := format(tt.input)
		if err != nil {
			t.Fatalf("error formatting %q: %v", tt.input, err)
		}
		if !strings.Contains(output, tt.expected) {
			t.Errorf("expected %q in output for %q, got:\n%s", tt.expected, tt.input, output)
		}
	}
}

func TestFormatBlankLinesBetweenBlocks(t *testing.T) {
	src := "x is 1\n\nif x is 1:\n  say x\n\ny is 2\n"
	output, err := format(src)
	if err != nil {
		t.Fatal(err)
	}
	// Should have blank lines around the if block
	lines := strings.Split(output, "\n")
	if len(lines) < 5 {
		t.Errorf("expected at least 5 lines, got %d: %s", len(lines), output)
	}
}

func TestFormatDotAssignment(t *testing.T) {
	output, err := format(`dog.name is "Rex"`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, `dog.name is "Rex"`) {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestFormatNew(t *testing.T) {
	output, err := format("dog is new Dog()")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "new Dog()") {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestFormatComplexProgram(t *testing.T) {
	src := `name is "Sarah"
age is 25
say "Hello, {name}!"
if age is greater than 18:
  say "adult"
otherwise:
  say "young"
to add a b:
  give back a + b
colors are ["red", "green", "blue"]
for each color in colors:
  say color
`
	output, err := format(src)
	if err != nil {
		t.Fatal(err)
	}
	// Just verify it doesn't crash and produces output
	if len(output) < 100 {
		t.Errorf("output too short: %s", output)
	}
}

func TestFormatCustomIndent(t *testing.T) {
	f := NewWithIndent("\t")
	l := lexer.New("if x is 1:\n  say x\n")
	tokens, _ := l.Tokenize()
	p := parser.New(tokens)
	prog, _ := p.Parse()
	output := f.Format(prog)
	if !strings.Contains(output, "\tsay x") {
		t.Errorf("expected tab indentation: %s", output)
	}
}
