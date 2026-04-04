package parser

import (
	"quill/ast"
	"testing"
)

func TestParseYieldStatement(t *testing.T) {
	prog, err := parse("to gen:\n    yield 42\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	fn, ok := prog.Statements[0].(*ast.FuncDefinition)
	if !ok {
		t.Fatalf("expected FuncDefinition, got %T", prog.Statements[0])
	}
	if fn.Name != "gen" {
		t.Errorf("expected function name 'gen', got %q", fn.Name)
	}
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(fn.Body))
	}
	yield, ok := fn.Body[0].(*ast.YieldStatement)
	if !ok {
		t.Fatalf("expected YieldStatement, got %T", fn.Body[0])
	}
	num, ok := yield.Value.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral in yield, got %T", yield.Value)
	}
	if num.Value != 42 {
		t.Errorf("expected yield value 42, got %v", num.Value)
	}
}

func TestParseLoopBlock(t *testing.T) {
	prog, err := parse("to gen:\n    loop:\n        yield 1\n        break\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fn, ok := prog.Statements[0].(*ast.FuncDefinition)
	if !ok {
		t.Fatalf("expected FuncDefinition, got %T", prog.Statements[0])
	}
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(fn.Body))
	}
	loop, ok := fn.Body[0].(*ast.LoopStatement)
	if !ok {
		t.Fatalf("expected LoopStatement, got %T", fn.Body[0])
	}
	if len(loop.Body) != 2 {
		t.Fatalf("expected 2 loop body statements, got %d", len(loop.Body))
	}
	_, ok = loop.Body[0].(*ast.YieldStatement)
	if !ok {
		t.Errorf("expected YieldStatement in loop body, got %T", loop.Body[0])
	}
	_, ok = loop.Body[1].(*ast.BreakStatement)
	if !ok {
		t.Errorf("expected BreakStatement in loop body, got %T", loop.Body[1])
	}
}

func TestParseGeneratorFunction(t *testing.T) {
	src := `to fibonacci:
    a is 0
    b is 1
    loop:
        yield a
        next is a + b
        a is b
        b is next
`
	prog, err := parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	fn, ok := prog.Statements[0].(*ast.FuncDefinition)
	if !ok {
		t.Fatalf("expected FuncDefinition, got %T", prog.Statements[0])
	}
	if fn.Name != "fibonacci" {
		t.Errorf("expected function name 'fibonacci', got %q", fn.Name)
	}
	// Should have: a is 0, b is 1, loop:
	if len(fn.Body) != 3 {
		t.Fatalf("expected 3 body statements, got %d", len(fn.Body))
	}
	loop, ok := fn.Body[2].(*ast.LoopStatement)
	if !ok {
		t.Fatalf("expected LoopStatement as 3rd statement, got %T", fn.Body[2])
	}
	// Loop body: yield a, next is a + b, a is b, b is next
	if len(loop.Body) != 4 {
		t.Fatalf("expected 4 loop body statements, got %d", len(loop.Body))
	}
}

func TestParseYieldExpression(t *testing.T) {
	prog, err := parse("to gen:\n    yield a + b\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fn := prog.Statements[0].(*ast.FuncDefinition)
	yield, ok := fn.Body[0].(*ast.YieldStatement)
	if !ok {
		t.Fatalf("expected YieldStatement, got %T", fn.Body[0])
	}
	_, ok = yield.Value.(*ast.BinaryExpr)
	if !ok {
		t.Errorf("expected BinaryExpr in yield value, got %T", yield.Value)
	}
}
