package parser

import (
	"quill/ast"
	"quill/lexer"
	"testing"
)

func parse(input string) (*ast.Program, error) {
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		return nil, err
	}
	p := New(tokens)
	return p.Parse()
}

func TestParseAssignment(t *testing.T) {
	prog, err := parse(`name is "hello"`)
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
	if assign.Name != "name" {
		t.Errorf("expected name 'name', got %q", assign.Name)
	}
	str, ok := assign.Value.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expected StringLiteral, got %T", assign.Value)
	}
	if str.Value != "hello" {
		t.Errorf("expected value 'hello', got %q", str.Value)
	}
}

func TestParseNumberAssignment(t *testing.T) {
	prog, err := parse("age is 25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	num, ok := assign.Value.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral, got %T", assign.Value)
	}
	if num.Value != 25 {
		t.Errorf("expected 25, got %f", num.Value)
	}
}

func TestParseBoolAssignment(t *testing.T) {
	prog, err := parse("active is yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	b, ok := assign.Value.(*ast.BoolLiteral)
	if !ok {
		t.Fatalf("expected BoolLiteral, got %T", assign.Value)
	}
	if !b.Value {
		t.Error("expected true")
	}
}

func TestParseAreAssignment(t *testing.T) {
	prog, err := parse(`items are [1, 2, 3]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	if assign.Name != "items" {
		t.Errorf("expected name 'items', got %q", assign.Name)
	}
	list, ok := assign.Value.(*ast.ListLiteral)
	if !ok {
		t.Fatalf("expected ListLiteral, got %T", assign.Value)
	}
	if len(list.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(list.Elements))
	}
}

func TestParseSay(t *testing.T) {
	prog, err := parse(`say "Hello, world!"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	say, ok := prog.Statements[0].(*ast.SayStatement)
	if !ok {
		t.Fatalf("expected SayStatement, got %T", prog.Statements[0])
	}
	str := say.Value.(*ast.StringLiteral)
	if str.Value != "Hello, world!" {
		t.Errorf("expected 'Hello, world!', got %q", str.Value)
	}
}

func TestParseIfStatement(t *testing.T) {
	prog, err := parse("if x is greater than 10:\n  say x\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ifStmt, ok := prog.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", prog.Statements[0])
	}
	if len(ifStmt.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(ifStmt.Body))
	}
}

func TestParseIfOtherwise(t *testing.T) {
	prog, err := parse("if x is 1:\n  say x\notherwise:\n  say y\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ifStmt := prog.Statements[0].(*ast.IfStatement)
	if len(ifStmt.Else) != 1 {
		t.Errorf("expected 1 else statement, got %d", len(ifStmt.Else))
	}
}

func TestParseIfOtherwiseIf(t *testing.T) {
	prog, err := parse("if x is 1:\n  say x\notherwise if x is 2:\n  say y\notherwise:\n  say z\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ifStmt := prog.Statements[0].(*ast.IfStatement)
	if len(ifStmt.ElseIfs) != 1 {
		t.Errorf("expected 1 else-if clause, got %d", len(ifStmt.ElseIfs))
	}
	if len(ifStmt.Else) != 1 {
		t.Errorf("expected 1 else statement, got %d", len(ifStmt.Else))
	}
}

func TestParseForEach(t *testing.T) {
	prog, err := parse("for each item in items:\n  say item\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	forStmt, ok := prog.Statements[0].(*ast.ForEachStatement)
	if !ok {
		t.Fatalf("expected ForEachStatement, got %T", prog.Statements[0])
	}
	if forStmt.Variable != "item" {
		t.Errorf("expected variable 'item', got %q", forStmt.Variable)
	}
}

func TestParseWhile(t *testing.T) {
	prog, err := parse("while x is less than 10:\n  x is x + 1\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	whileStmt, ok := prog.Statements[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("expected WhileStatement, got %T", prog.Statements[0])
	}
	if len(whileStmt.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(whileStmt.Body))
	}
}

func TestParseFuncDef(t *testing.T) {
	prog, err := parse("to add a b:\n  give back a + b\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fn, ok := prog.Statements[0].(*ast.FuncDefinition)
	if !ok {
		t.Fatalf("expected FuncDefinition, got %T", prog.Statements[0])
	}
	if fn.Name != "add" {
		t.Errorf("expected name 'add', got %q", fn.Name)
	}
	if len(fn.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(fn.Params))
	}
	if fn.Params[0] != "a" || fn.Params[1] != "b" {
		t.Errorf("expected params [a, b], got %v", fn.Params)
	}
}

func TestParseFuncNoParams(t *testing.T) {
	prog, err := parse("to greet:\n  say \"hi\"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fn := prog.Statements[0].(*ast.FuncDefinition)
	if len(fn.Params) != 0 {
		t.Errorf("expected 0 params, got %d", len(fn.Params))
	}
}

func TestParseReturn(t *testing.T) {
	prog, err := parse("to double x:\n  give back x * 2\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fn := prog.Statements[0].(*ast.FuncDefinition)
	ret, ok := fn.Body[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected ReturnStatement, got %T", fn.Body[0])
	}
	_ = ret
}

func TestParseUse(t *testing.T) {
	prog, err := parse(`use "helpers.quill"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	use, ok := prog.Statements[0].(*ast.UseStatement)
	if !ok {
		t.Fatalf("expected UseStatement, got %T", prog.Statements[0])
	}
	if use.Path != "helpers.quill" {
		t.Errorf("expected path 'helpers.quill', got %q", use.Path)
	}
	if use.Alias != "" {
		t.Errorf("expected no alias, got %q", use.Alias)
	}
}

func TestParseUseAs(t *testing.T) {
	prog, err := parse(`use "express" as app`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	use := prog.Statements[0].(*ast.UseStatement)
	if use.Path != "express" {
		t.Errorf("expected path 'express', got %q", use.Path)
	}
	if use.Alias != "app" {
		t.Errorf("expected alias 'app', got %q", use.Alias)
	}
}

func TestParseTestBlock(t *testing.T) {
	prog, err := parse("test \"math works\":\n  expect 1 + 1 is 2\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	test, ok := prog.Statements[0].(*ast.TestBlock)
	if !ok {
		t.Fatalf("expected TestBlock, got %T", prog.Statements[0])
	}
	if test.Name != "math works" {
		t.Errorf("expected name 'math works', got %q", test.Name)
	}
	if len(test.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(test.Body))
	}
}

func TestParseExpect(t *testing.T) {
	prog, err := parse("test \"t\":\n  expect x is 5\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	test := prog.Statements[0].(*ast.TestBlock)
	expect, ok := test.Body[0].(*ast.ExpectStatement)
	if !ok {
		t.Fatalf("expected ExpectStatement, got %T", test.Body[0])
	}
	_ = expect
}

func TestParseDescribe(t *testing.T) {
	prog, err := parse("describe Dog:\n  name is \"\"\n  to bark:\n    say \"woof\"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	desc, ok := prog.Statements[0].(*ast.DescribeStatement)
	if !ok {
		t.Fatalf("expected DescribeStatement, got %T", prog.Statements[0])
	}
	if desc.Name != "Dog" {
		t.Errorf("expected name 'Dog', got %q", desc.Name)
	}
	if len(desc.Properties) != 1 {
		t.Errorf("expected 1 property, got %d", len(desc.Properties))
	}
	if len(desc.Methods) != 1 {
		t.Errorf("expected 1 method, got %d", len(desc.Methods))
	}
}

func TestParseNewExpr(t *testing.T) {
	prog, err := parse("dog is new Dog()")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	newExpr, ok := assign.Value.(*ast.NewExpr)
	if !ok {
		t.Fatalf("expected NewExpr, got %T", assign.Value)
	}
	if newExpr.ClassName != "Dog" {
		t.Errorf("expected class 'Dog', got %q", newExpr.ClassName)
	}
}

func TestParseDotAssignment(t *testing.T) {
	prog, err := parse("dog.name is \"Rex\"")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dot, ok := prog.Statements[0].(*ast.DotAssignStatement)
	if !ok {
		t.Fatalf("expected DotAssignStatement, got %T", prog.Statements[0])
	}
	if dot.Object != "dog" {
		t.Errorf("expected object 'dog', got %q", dot.Object)
	}
	if dot.Field != "name" {
		t.Errorf("expected field 'name', got %q", dot.Field)
	}
}

func TestParseArithmetic(t *testing.T) {
	prog, err := parse("result is 1 + 2 * 3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	// Should be 1 + (2 * 3) due to precedence
	binExpr, ok := assign.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", assign.Value)
	}
	if binExpr.Operator != "+" {
		t.Errorf("expected top-level +, got %q", binExpr.Operator)
	}
	// Right side should be 2 * 3
	right, ok := binExpr.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr on right, got %T", binExpr.Right)
	}
	if right.Operator != "*" {
		t.Errorf("expected *, got %q", right.Operator)
	}
}

func TestParseComparisons(t *testing.T) {
	tests := []struct {
		input string
		op    string
	}{
		{"x is greater than 10", ">"},
		{"x is less than 5", "<"},
		{"x is equal to 0", "=="},
		{"x is not 0", "!="},
		{"x is 5", "=="},
		{"x is greater than or equal to 5", ">="},
		{"x is less than or equal to 5", "<="},
		{"list contains 5", "contains"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			prog, err := parse("if " + tt.input + ":\n  say x\n")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			ifStmt := prog.Statements[0].(*ast.IfStatement)
			cmp, ok := ifStmt.Condition.(*ast.ComparisonExpr)
			if !ok {
				t.Fatalf("expected ComparisonExpr, got %T", ifStmt.Condition)
			}
			if cmp.Operator != tt.op {
				t.Errorf("expected operator %q, got %q", tt.op, cmp.Operator)
			}
		})
	}
}

func TestParseLogicalExpr(t *testing.T) {
	prog, err := parse("if x and y:\n  say x\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ifStmt := prog.Statements[0].(*ast.IfStatement)
	logical, ok := ifStmt.Condition.(*ast.LogicalExpr)
	if !ok {
		t.Fatalf("expected LogicalExpr, got %T", ifStmt.Condition)
	}
	if logical.Operator != "and" {
		t.Errorf("expected 'and', got %q", logical.Operator)
	}
}

func TestParseOrExpr(t *testing.T) {
	prog, err := parse("if x or y:\n  say x\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ifStmt := prog.Statements[0].(*ast.IfStatement)
	logical, ok := ifStmt.Condition.(*ast.LogicalExpr)
	if !ok {
		t.Fatalf("expected LogicalExpr, got %T", ifStmt.Condition)
	}
	if logical.Operator != "or" {
		t.Errorf("expected 'or', got %q", logical.Operator)
	}
}

func TestParseNotExpr(t *testing.T) {
	prog, err := parse("if not x:\n  say y\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ifStmt := prog.Statements[0].(*ast.IfStatement)
	_, ok := ifStmt.Condition.(*ast.NotExpr)
	if !ok {
		t.Fatalf("expected NotExpr, got %T", ifStmt.Condition)
	}
}

func TestParseUnaryMinus(t *testing.T) {
	prog, err := parse("x is -5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	_, ok := assign.Value.(*ast.UnaryMinusExpr)
	if !ok {
		t.Fatalf("expected UnaryMinusExpr, got %T", assign.Value)
	}
}

func TestParseFunctionCall(t *testing.T) {
	prog, err := parse("result is add(1, 2)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	call, ok := assign.Value.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", assign.Value)
	}
	if len(call.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(call.Args))
	}
}

func TestParseFunctionCallNoArgs(t *testing.T) {
	prog, err := parse("result is random()")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	call, ok := assign.Value.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", assign.Value)
	}
	if len(call.Args) != 0 {
		t.Errorf("expected 0 args, got %d", len(call.Args))
	}
}

func TestParseDotExpr(t *testing.T) {
	prog, err := parse("say obj.field")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sayStmt := prog.Statements[0].(*ast.SayStatement)
	dot, ok := sayStmt.Value.(*ast.DotExpr)
	if !ok {
		t.Fatalf("expected DotExpr, got %T", sayStmt.Value)
	}
	if dot.Field != "field" {
		t.Errorf("expected field 'field', got %q", dot.Field)
	}
}

func TestParseIndexExpr(t *testing.T) {
	prog, err := parse("say items[0]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sayStmt := prog.Statements[0].(*ast.SayStatement)
	idx, ok := sayStmt.Value.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr, got %T", sayStmt.Value)
	}
	_ = idx
}

func TestParseMethodCall(t *testing.T) {
	prog, err := parse("dog.bark()")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	exprStmt := prog.Statements[0].(*ast.ExprStatement)
	call, ok := exprStmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
	}
	dot, ok := call.Function.(*ast.DotExpr)
	if !ok {
		t.Fatalf("expected DotExpr as function, got %T", call.Function)
	}
	if dot.Field != "bark" {
		t.Errorf("expected method 'bark', got %q", dot.Field)
	}
}

func TestParseAwait(t *testing.T) {
	prog, err := parse(`data is await fetchJSON("url")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	awaitExpr, ok := assign.Value.(*ast.AwaitExpr)
	if !ok {
		t.Fatalf("expected AwaitExpr, got %T", assign.Value)
	}
	_, ok = awaitExpr.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr inside await, got %T", awaitExpr.Expr)
	}
}

func TestParseMyDot(t *testing.T) {
	prog, err := parse("to bark:\n  say my.name\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fn := prog.Statements[0].(*ast.FuncDefinition)
	sayStmt := fn.Body[0].(*ast.SayStatement)
	dot, ok := sayStmt.Value.(*ast.DotExpr)
	if !ok {
		t.Fatalf("expected DotExpr, got %T", sayStmt.Value)
	}
	ident, ok := dot.Object.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier, got %T", dot.Object)
	}
	if ident.Name != "this" {
		t.Errorf("expected 'this' (from my), got %q", ident.Name)
	}
}

func TestParseListLiteral(t *testing.T) {
	prog, err := parse(`items are ["a", "b", "c"]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	list, ok := assign.Value.(*ast.ListLiteral)
	if !ok {
		t.Fatalf("expected ListLiteral, got %T", assign.Value)
	}
	if len(list.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(list.Elements))
	}
}

func TestParseEmptyList(t *testing.T) {
	prog, err := parse("items are []")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	list := assign.Value.(*ast.ListLiteral)
	if len(list.Elements) != 0 {
		t.Errorf("expected 0 elements, got %d", len(list.Elements))
	}
}

func TestParseTrailingCommaInList(t *testing.T) {
	prog, err := parse("items are [1, 2, 3,]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	list := assign.Value.(*ast.ListLiteral)
	if len(list.Elements) != 3 {
		t.Errorf("expected 3 elements with trailing comma, got %d", len(list.Elements))
	}
}

func TestParseParenExpr(t *testing.T) {
	prog, err := parse("x is (1 + 2) * 3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign := prog.Statements[0].(*ast.AssignStatement)
	bin, ok := assign.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", assign.Value)
	}
	if bin.Operator != "*" {
		t.Errorf("expected *, got %q", bin.Operator)
	}
}

func TestParseMultipleStatements(t *testing.T) {
	prog, err := parse("x is 1\ny is 2\nz is 3\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Statements) != 3 {
		t.Errorf("expected 3 statements, got %d", len(prog.Statements))
	}
}

func TestParseError(t *testing.T) {
	_, err := parse("if :\n  say x\n")
	if err == nil {
		t.Error("expected parse error")
	}
}

func TestParseLineNumbers(t *testing.T) {
	prog, err := parse("x is 1\ny is 2\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a1 := prog.Statements[0].(*ast.AssignStatement)
	a2 := prog.Statements[1].(*ast.AssignStatement)
	if a1.Line != 1 {
		t.Errorf("expected line 1, got %d", a1.Line)
	}
	if a2.Line != 2 {
		t.Errorf("expected line 2, got %d", a2.Line)
	}
}

func TestParseComplexProgram(t *testing.T) {
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

test "math works":
  expect add(2, 3) is 5
  expect add(-1, 1) is 0
`
	prog, err := parse(src)
	if err != nil {
		t.Fatalf("unexpected error parsing complex program: %v", err)
	}
	if len(prog.Statements) < 8 {
		t.Errorf("expected at least 8 statements, got %d", len(prog.Statements))
	}
}
