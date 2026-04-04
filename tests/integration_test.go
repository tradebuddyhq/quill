package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"quill/codegen"
	"quill/lexer"
	"quill/parser"
	"strings"
	"testing"
)

// compileQuill compiles Quill source code to JavaScript.
func compileQuill(source string) (string, error) {
	l := lexer.New(source)
	tokens, err := l.Tokenize()
	if err != nil {
		return "", err
	}
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return "", err
	}
	g := codegen.New()
	return g.Generate(program), nil
}

// runJS runs JavaScript code using Node.js and returns stdout.
func runJS(t *testing.T, js string) string {
	t.Helper()

	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not found, skipping integration test")
	}

	tmpFile, err := os.CreateTemp("", "quill-test-*.js")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(js)
	tmpFile.Close()

	cmd := exec.Command("node", tmpFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("node execution failed: %v\nOutput: %s\nJS:\n%s", err, string(output), js)
	}
	return strings.TrimSpace(string(output))
}

// compileAndRun compiles Quill source and runs the resulting JS.
func compileAndRun(t *testing.T, source string) string {
	t.Helper()
	js, err := compileQuill(source)
	if err != nil {
		t.Fatalf("compilation failed: %v\nSource:\n%s", err, source)
	}
	return runJS(t, js)
}

func TestVariableAssignment(t *testing.T) {
	output := compileAndRun(t, `name is "Alice"
say name`)
	if output != "Alice" {
		t.Errorf("expected 'Alice', got %q", output)
	}
}

func TestNumberAssignment(t *testing.T) {
	output := compileAndRun(t, `x is 42
say x`)
	if output != "42" {
		t.Errorf("expected '42', got %q", output)
	}
}

func TestArithmetic(t *testing.T) {
	tests := []struct {
		name   string
		source string
		expect string
	}{
		{"addition", "say 2 + 3", "5"},
		{"subtraction", "say 10 - 4", "6"},
		{"multiplication", "say 3 * 7", "21"},
		{"division", "say 15 / 3", "5"},
		{"modulo", "say 10 % 3", "1"},
		{"precedence", "say 2 + 3 * 4", "14"},
		{"parentheses", "say (2 + 3) * 4", "20"},
		{"negative", "say -5 + 10", "5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := compileAndRun(t, tt.source)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

func TestStringOperations(t *testing.T) {
	output := compileAndRun(t, `greeting is "Hello" + " " + "World"
say greeting`)
	if output != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", output)
	}
}

func TestIfElse(t *testing.T) {
	tests := []struct {
		name   string
		source string
		expect string
	}{
		{
			"if true",
			"x is 10\nif x is greater than 5:\n  say \"big\"\n",
			"big",
		},
		{
			"if false with otherwise",
			"x is 3\nif x is greater than 5:\n  say \"big\"\notherwise:\n  say \"small\"\n",
			"small",
		},
		{
			"otherwise if chain",
			"x is 5\nif x is greater than 10:\n  say \"large\"\notherwise if x is greater than 3:\n  say \"medium\"\notherwise:\n  say \"small\"\n",
			"medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := compileAndRun(t, tt.source)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

func TestWhileLoop(t *testing.T) {
	source := `i is 0
result is ""
while i is less than 5:
  result is result + "x"
  i is i + 1
say result`
	output := compileAndRun(t, source)
	if output != "xxxxx" {
		t.Errorf("expected 'xxxxx', got %q", output)
	}
}

func TestForEachLoop(t *testing.T) {
	source := `items are [1, 2, 3]
total is 0
for each item in items:
  total is total + item
say total`
	output := compileAndRun(t, source)
	if output != "6" {
		t.Errorf("expected '6', got %q", output)
	}
}

func TestFunctions(t *testing.T) {
	source := `to add a, b:
  give back a + b

result is add(3, 4)
say result`
	output := compileAndRun(t, source)
	if output != "7" {
		t.Errorf("expected '7', got %q", output)
	}
}

func TestFunctionNoParams(t *testing.T) {
	source := `to greet:
  give back "hello"

say greet()`
	output := compileAndRun(t, source)
	if output != "hello" {
		t.Errorf("expected 'hello', got %q", output)
	}
}

func TestMatchStatement(t *testing.T) {
	source := `x is 2
match x:
  when 1:
    say "one"
  when 2:
    say "two"
  otherwise:
    say "other"
`
	output := compileAndRun(t, source)
	if output != "two" {
		t.Errorf("expected 'two', got %q", output)
	}
}

func TestBooleans(t *testing.T) {
	source := `active is yes
if active:
  say "active"
otherwise:
  say "inactive"
`
	output := compileAndRun(t, source)
	if output != "active" {
		t.Errorf("expected 'active', got %q", output)
	}
}

func TestListLiteral(t *testing.T) {
	source := `items are [10, 20, 30]
say items[1]`
	output := compileAndRun(t, source)
	if output != "20" {
		t.Errorf("expected '20', got %q", output)
	}
}

func TestPipeOperator(t *testing.T) {
	source := `to double x:
  give back x * 2

to addOne x:
  give back x + 1

result is 5 | double | addOne
say result`
	// Pipe operator passes left as first arg to right function
	output := compileAndRun(t, source)
	if output != "11" {
		t.Errorf("expected '11', got %q", output)
	}
}

func TestTryCatch(t *testing.T) {
	source := `try:
  x is nothing.field
if it fails err:
  say "caught"
`
	output := compileAndRun(t, source)
	if output != "caught" {
		t.Errorf("expected 'caught', got %q", output)
	}
}

func TestDescribeClass(t *testing.T) {
	source := `describe Dog:
  name is ""
  to bark:
    say "woof"

d is new Dog()
d.name is "Rex"
d.bark()
say d.name`
	output := compileAndRun(t, source)
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d: %q", len(lines), output)
	}
	if lines[0] != "woof" {
		t.Errorf("expected first line 'woof', got %q", lines[0])
	}
	if lines[1] != "Rex" {
		t.Errorf("expected second line 'Rex', got %q", lines[1])
	}
}

func TestBreakContinue(t *testing.T) {
	source := `i is 0
result is ""
while yes:
  if i is 5:
    break
  i is i + 1
  if i is 3:
    continue
  result is result + "x"
say result`
	output := compileAndRun(t, source)
	// i goes 0->1 (x), 1->2 (x), 2->3 (continue), 3->4 (x), 4->5 (x), 5 (break)
	if output != "xxxx" {
		t.Errorf("expected 'xxxx', got %q", output)
	}
}

func TestCompileToFile(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not found, skipping integration test")
	}

	// Write a .quill file
	tmpDir := t.TempDir()
	quillFile := filepath.Join(tmpDir, "test.quill")
	err := os.WriteFile(quillFile, []byte(`say "file test"`), 0644)
	if err != nil {
		t.Fatalf("failed to write quill file: %v", err)
	}

	// Compile it
	js, err := compileQuill(`say "file test"`)
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	jsFile := filepath.Join(tmpDir, "test.js")
	err = os.WriteFile(jsFile, []byte(js), 0644)
	if err != nil {
		t.Fatalf("failed to write js file: %v", err)
	}

	// Run the JS file
	cmd := exec.Command("node", jsFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("node execution failed: %v\nOutput: %s", err, output)
	}

	if strings.TrimSpace(string(output)) != "file test" {
		t.Errorf("expected 'file test', got %q", strings.TrimSpace(string(output)))
	}
}

func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		name   string
		source string
		expect string
	}{
		{
			"and true",
			"if yes and yes:\n  say \"both\"\n",
			"both",
		},
		{
			"or true",
			"if no or yes:\n  say \"one\"\n",
			"one",
		},
		{
			"not",
			"if not no:\n  say \"negated\"\n",
			"negated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := compileAndRun(t, tt.source)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

func TestComparisonOperators(t *testing.T) {
	tests := []struct {
		name   string
		source string
		expect string
	}{
		{
			"greater than",
			"if 10 is greater than 5:\n  say \"yes\"\n",
			"yes",
		},
		{
			"less than",
			"if 3 is less than 5:\n  say \"yes\"\n",
			"yes",
		},
		{
			"equal to",
			"if 5 is equal to 5:\n  say \"yes\"\n",
			"yes",
		},
		{
			"is not",
			"if 5 is not 3:\n  say \"yes\"\n",
			"yes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := compileAndRun(t, tt.source)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}
