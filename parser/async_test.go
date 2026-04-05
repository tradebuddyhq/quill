package parser

import (
	"quill/ast"
	"testing"
)

func TestParseCancelStatement(t *testing.T) {
	prog, err := parse("cancel myTask\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	cancel, ok := prog.Statements[0].(*ast.CancelStatement)
	if !ok {
		t.Fatalf("expected CancelStatement, got %T", prog.Statements[0])
	}
	if cancel.Target != "myTask" {
		t.Errorf("expected target 'myTask', got %q", cancel.Target)
	}
}

func TestParseForAwaitEach(t *testing.T) {
	input := "for await each chunk in stream:\n    say chunk\n"
	prog, err := parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	forEach, ok := prog.Statements[0].(*ast.ForEachStatement)
	if !ok {
		t.Fatalf("expected ForEachStatement, got %T", prog.Statements[0])
	}
	if !forEach.IsAsync {
		t.Error("expected IsAsync to be true")
	}
	if forEach.Variable != "chunk" {
		t.Errorf("expected variable 'chunk', got %q", forEach.Variable)
	}
}

func TestParseParallelSettled(t *testing.T) {
	input := "parallel settled:\n    task1 is fetch(\"/a\")\n    task2 is fetch(\"/b\")\n"
	prog, err := parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	par, ok := prog.Statements[0].(*ast.ParallelBlock)
	if !ok {
		t.Fatalf("expected ParallelBlock, got %T", prog.Statements[0])
	}
	if !par.IsSettled {
		t.Error("expected IsSettled to be true")
	}
	if len(par.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(par.Tasks))
	}
}

func TestParsePropagateOperator(t *testing.T) {
	input := "data is fetchJSON(\"/api\")?\n"
	prog, err := parse(input)
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
	propagate, ok := assign.Value.(*ast.PropagateExpr)
	if !ok {
		t.Fatalf("expected PropagateExpr as value, got %T", assign.Value)
	}
	_, ok = propagate.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr inside PropagateExpr, got %T", propagate.Expr)
	}
}

func TestParseDestructuringInForEach(t *testing.T) {
	input := "for each {name, age} in users:\n    say name\n"
	prog, err := parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	forEach, ok := prog.Statements[0].(*ast.ForEachStatement)
	if !ok {
		t.Fatalf("expected ForEachStatement, got %T", prog.Statements[0])
	}
	if forEach.Variable != "" {
		t.Errorf("expected empty variable name for destructuring, got %q", forEach.Variable)
	}
	objPat, ok := forEach.DestructurePattern.(*ast.ObjectPattern)
	if !ok {
		t.Fatalf("expected ObjectPattern, got %T", forEach.DestructurePattern)
	}
	if len(objPat.Fields) != 2 {
		t.Errorf("expected 2 fields in pattern, got %d", len(objPat.Fields))
	}
	if objPat.Fields[0].Key != "name" {
		t.Errorf("expected first field 'name', got %q", objPat.Fields[0].Key)
	}
	if objPat.Fields[1].Key != "age" {
		t.Errorf("expected second field 'age', got %q", objPat.Fields[1].Key)
	}
}

func TestParseDestructuringInMatch(t *testing.T) {
	input := "match response:\n    when {status: 200, body}:\n        say body\n    when {status: 404}:\n        say \"not found\"\n"
	prog, err := parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	match, ok := prog.Statements[0].(*ast.MatchStatement)
	if !ok {
		t.Fatalf("expected MatchStatement, got %T", prog.Statements[0])
	}
	if len(match.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(match.Cases))
	}

	// First case: when {status: 200, body}
	case1 := match.Cases[0]
	objPat1, ok := case1.Pattern.(*ast.ObjectMatchPattern)
	if !ok {
		t.Fatalf("expected ObjectMatchPattern for case 1, got %T", case1.Pattern)
	}
	if len(objPat1.Fields) != 2 {
		t.Fatalf("expected 2 fields in case 1, got %d", len(objPat1.Fields))
	}
	if objPat1.Fields[0].Key != "status" {
		t.Errorf("expected field key 'status', got %q", objPat1.Fields[0].Key)
	}
	if objPat1.Fields[0].Value == nil {
		t.Error("expected value for 'status' field")
	}
	if objPat1.Fields[1].Key != "body" {
		t.Errorf("expected field key 'body', got %q", objPat1.Fields[1].Key)
	}
	if objPat1.Fields[1].Value != nil {
		t.Error("expected nil value for 'body' field (binding only)")
	}

	// Second case: when {status: 404}
	case2 := match.Cases[1]
	objPat2, ok := case2.Pattern.(*ast.ObjectMatchPattern)
	if !ok {
		t.Fatalf("expected ObjectMatchPattern for case 2, got %T", case2.Pattern)
	}
	if len(objPat2.Fields) != 1 {
		t.Fatalf("expected 1 field in case 2, got %d", len(objPat2.Fields))
	}
}
