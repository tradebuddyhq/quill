package analyzer

import (
	"quill/lexer"
	"quill/parser"
	"strings"
	"testing"
)

func analyze(input string) []Diagnostic {
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		return nil
	}
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		return nil
	}
	a := New()
	return a.Analyze(prog)
}

func hasDiagnostic(diagnostics []Diagnostic, severity Severity, substr string) bool {
	for _, d := range diagnostics {
		if d.Severity == severity && strings.Contains(d.Message, substr) {
			return true
		}
	}
	return false
}

func TestNoWarningsForCleanCode(t *testing.T) {
	src := `name is "Sarah"
say name
`
	diags := analyze(src)
	if len(diags) > 0 {
		t.Errorf("expected no diagnostics for clean code, got %d: %v", len(diags), diags)
	}
}

func TestUnusedVariable(t *testing.T) {
	src := `x is 5
y is 10
say y
`
	diags := analyze(src)
	if !hasDiagnostic(diags, Warning, "\"x\" is defined but never used") {
		t.Errorf("expected unused variable warning for x, got: %v", diags)
	}
}

func TestUnderscoreVariableNoWarning(t *testing.T) {
	src := `_temp is 5
x is 10
say x
`
	diags := analyze(src)
	if hasDiagnostic(diags, Warning, "_temp") {
		t.Error("should not warn about variables starting with _")
	}
}

func TestDivisionByZero(t *testing.T) {
	src := `result is 10 / 0
say result
`
	diags := analyze(src)
	if !hasDiagnostic(diags, Warning, "division by zero") {
		t.Errorf("expected division by zero warning, got: %v", diags)
	}
}

func TestSelfComparison(t *testing.T) {
	src := "if x is x:\n  say x\n"
	diags := analyze(src)
	if !hasDiagnostic(diags, Warning, "comparing") {
		t.Errorf("expected self-comparison warning, got: %v", diags)
	}
}

func TestAlwaysTrueCondition(t *testing.T) {
	src := "if yes:\n  say \"always\"\n"
	diags := analyze(src)
	if !hasDiagnostic(diags, Warning, "always true") {
		t.Errorf("expected always-true warning, got: %v", diags)
	}
}

func TestAlwaysFalseCondition(t *testing.T) {
	src := "if no:\n  say \"never\"\n"
	diags := analyze(src)
	if !hasDiagnostic(diags, Warning, "always false") {
		t.Errorf("expected always-false warning, got: %v", diags)
	}
}

func TestInfiniteLoop(t *testing.T) {
	src := "while yes:\n  say \"loop\"\n"
	diags := analyze(src)
	if !hasDiagnostic(diags, Warning, "infinite loop") {
		t.Errorf("expected infinite loop warning, got: %v", diags)
	}
}

func TestReturnOutsideFunction(t *testing.T) {
	src := "give back 5\n"
	diags := analyze(src)
	if !hasDiagnostic(diags, Error, "outside of a function") {
		t.Errorf("expected return-outside-function error, got: %v", diags)
	}
}

func TestTestWithNoExpect(t *testing.T) {
	src := "test \"empty test\":\n  say \"nothing\"\n"
	diags := analyze(src)
	if !hasDiagnostic(diags, Warning, "no expectations") {
		t.Errorf("expected no-expectations warning, got: %v", diags)
	}
}

func TestExpectOutsideTest(t *testing.T) {
	src := "expect 1 is 1\n"
	diags := analyze(src)
	if !hasDiagnostic(diags, Warning, "outside of a test") {
		t.Errorf("expected expect-outside-test warning, got: %v", diags)
	}
}

func TestLowercaseClassName(t *testing.T) {
	src := "describe dog:\n  name is \"\"\n"
	diags := analyze(src)
	if !hasDiagnostic(diags, Info, "should start with uppercase") {
		t.Errorf("expected lowercase class name info, got: %v", diags)
	}
}

func TestUppercaseClassNameNoWarning(t *testing.T) {
	src := "describe Dog:\n  name is \"\"\n"
	diags := analyze(src)
	if hasDiagnostic(diags, Info, "should start with uppercase") {
		t.Error("should not warn about uppercase class names")
	}
}

func TestWrongArgCount(t *testing.T) {
	src := `to add a b:
  give back a + b

result is add(1, 2, 3)
say result
`
	diags := analyze(src)
	if !hasDiagnostic(diags, Error, "expects 2 arguments but got 3") {
		t.Errorf("expected wrong arg count error, got: %v", diags)
	}
}

func TestCorrectArgCount(t *testing.T) {
	src := `to add a b:
  give back a + b

result is add(1, 2)
say result
`
	diags := analyze(src)
	if hasDiagnostic(diags, Error, "expects") {
		t.Error("should not warn about correct arg count")
	}
}

func TestNegativeIndex(t *testing.T) {
	src := "say items[-1]\n"
	diags := analyze(src)
	if !hasDiagnostic(diags, Warning, "negative array index") {
		t.Errorf("expected negative index warning, got: %v", diags)
	}
}

func TestHasErrors(t *testing.T) {
	diags := []Diagnostic{
		{Severity: Warning, Message: "test"},
		{Severity: Error, Message: "test"},
	}
	if !HasErrors(diags) {
		t.Error("expected HasErrors to return true")
	}
}

func TestHasWarnings(t *testing.T) {
	diags := []Diagnostic{
		{Severity: Warning, Message: "test"},
	}
	if !HasWarnings(diags) {
		t.Error("expected HasWarnings to return true")
	}
}

func TestNoErrors(t *testing.T) {
	diags := []Diagnostic{
		{Severity: Warning, Message: "test"},
	}
	if HasErrors(diags) {
		t.Error("expected HasErrors to return false")
	}
}

func TestFilterBySeverity(t *testing.T) {
	diags := []Diagnostic{
		{Severity: Warning, Message: "w1"},
		{Severity: Error, Message: "e1"},
		{Severity: Warning, Message: "w2"},
		{Severity: Info, Message: "i1"},
	}

	warnings := FilterBySeverity(diags, Warning)
	if len(warnings) != 2 {
		t.Errorf("expected 2 warnings, got %d", len(warnings))
	}

	errors := FilterBySeverity(diags, Error)
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}

	infos := FilterBySeverity(diags, Info)
	if len(infos) != 1 {
		t.Errorf("expected 1 info, got %d", len(infos))
	}
}

func TestSeverityString(t *testing.T) {
	if Warning.String() != "warning" {
		t.Errorf("expected 'warning', got %q", Warning.String())
	}
	if Error.String() != "error" {
		t.Errorf("expected 'error', got %q", Error.String())
	}
	if Info.String() != "info" {
		t.Errorf("expected 'info', got %q", Info.String())
	}
}

func TestDiagnosticString(t *testing.T) {
	d := Diagnostic{Line: 5, Severity: Warning, Message: "test warning", Hint: "fix it"}
	s := d.String()
	if !strings.Contains(s, "line 5") {
		t.Error("expected line number in string")
	}
	if !strings.Contains(s, "warning") {
		t.Error("expected severity in string")
	}
	if !strings.Contains(s, "fix it") {
		t.Error("expected hint in string")
	}
}

func TestDiagnosticStringNoHint(t *testing.T) {
	d := Diagnostic{Line: 1, Severity: Error, Message: "bad"}
	s := d.String()
	if strings.Contains(s, "hint") {
		t.Error("should not contain hint when empty")
	}
}

func TestCleanProgram(t *testing.T) {
	src := `name is "Sarah"
age is 25
say "Hello, {name}!"

if age is greater than 18:
  say "adult"

to greet person:
  say "Hi, {person}!"

greet(name)

test "age check":
  expect age is 25
`
	diags := analyze(src)
	errors := FilterBySeverity(diags, Error)
	if len(errors) > 0 {
		t.Errorf("expected no errors for clean program, got: %v", errors)
	}
}

func TestDuplicateProperty(t *testing.T) {
	src := "describe Car:\n  color is \"\"\n  color is \"red\"\n"
	diags := analyze(src)
	if !hasDiagnostic(diags, Warning, "duplicate property") {
		t.Errorf("expected duplicate property warning, got: %v", diags)
	}
}

func TestDuplicateMethod(t *testing.T) {
	src := "describe Car:\n  to drive:\n    say \"vroom\"\n  to drive:\n    say \"zoom\"\n"
	diags := analyze(src)
	if !hasDiagnostic(diags, Warning, "duplicate method") {
		t.Errorf("expected duplicate method warning, got: %v", diags)
	}
}
