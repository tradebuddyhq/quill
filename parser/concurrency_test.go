package parser

import (
	"quill/ast"
	"testing"
)

func TestParseSpawnTask(t *testing.T) {
	prog, err := parse("spawn task fetchData:\n    result is 42\n    give back result\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	spawn, ok := prog.Statements[0].(*ast.SpawnStatement)
	if !ok {
		t.Fatalf("expected SpawnStatement, got %T", prog.Statements[0])
	}
	if spawn.Name != "fetchData" {
		t.Errorf("expected name 'fetchData', got %q", spawn.Name)
	}
	if len(spawn.Body) != 2 {
		t.Errorf("expected 2 body statements, got %d", len(spawn.Body))
	}
}

func TestParseParallelBlock(t *testing.T) {
	input := "parallel:\n    task1 is fetch(\"/users\")\n    task2 is fetch(\"/posts\")\n"
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
	if len(par.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(par.Tasks))
	}
}

func TestParseRaceBlock(t *testing.T) {
	input := "race:\n    fast is fetch(\"/cdn1\")\n    backup is fetch(\"/cdn2\")\n"
	prog, err := parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	race, ok := prog.Statements[0].(*ast.RaceBlock)
	if !ok {
		t.Fatalf("expected RaceBlock, got %T", prog.Statements[0])
	}
	if len(race.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(race.Tasks))
	}
}

func TestParseChannelDeclaration(t *testing.T) {
	prog, err := parse("channel messages with buffer 10\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	ch, ok := prog.Statements[0].(*ast.ChannelStatement)
	if !ok {
		t.Fatalf("expected ChannelStatement, got %T", prog.Statements[0])
	}
	if ch.Name != "messages" {
		t.Errorf("expected channel name 'messages', got %q", ch.Name)
	}
	if ch.BufferSize == nil {
		t.Error("expected buffer size expression")
	}
}

func TestParseChannelNoBuffer(t *testing.T) {
	prog, err := parse("channel events\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ch, ok := prog.Statements[0].(*ast.ChannelStatement)
	if !ok {
		t.Fatalf("expected ChannelStatement, got %T", prog.Statements[0])
	}
	if ch.Name != "events" {
		t.Errorf("expected channel name 'events', got %q", ch.Name)
	}
	if ch.BufferSize != nil {
		t.Error("expected no buffer size")
	}
}

func TestParseSendStatement(t *testing.T) {
	prog, err := parse("send \"hello\" to messages\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	send, ok := prog.Statements[0].(*ast.SendStatement)
	if !ok {
		t.Fatalf("expected SendStatement, got %T", prog.Statements[0])
	}
	if send.Channel != "messages" {
		t.Errorf("expected channel 'messages', got %q", send.Channel)
	}
}

func TestParseSelectStatement(t *testing.T) {
	input := `select:
    when receive from messages:
        say "got message"
    when receive from errors:
        say "got error"
    after 5000:
        say "timeout"
`
	prog, err := parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sel, ok := prog.Statements[0].(*ast.SelectStatement)
	if !ok {
		t.Fatalf("expected SelectStatement, got %T", prog.Statements[0])
	}
	if len(sel.Cases) != 2 {
		t.Errorf("expected 2 cases, got %d", len(sel.Cases))
	}
	if sel.Cases[0].Channel != "messages" {
		t.Errorf("expected first case channel 'messages', got %q", sel.Cases[0].Channel)
	}
	if sel.Cases[1].Channel != "errors" {
		t.Errorf("expected second case channel 'errors', got %q", sel.Cases[1].Channel)
	}
	if sel.AfterMs == nil {
		t.Error("expected after timeout expression")
	}
}

func TestParseReceiveExpression(t *testing.T) {
	prog, err := parse("msg is receive from messages\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign, ok := prog.Statements[0].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", prog.Statements[0])
	}
	recv, ok := assign.Value.(*ast.ReceiveExpression)
	if !ok {
		t.Fatalf("expected ReceiveExpression, got %T", assign.Value)
	}
	if recv.Channel != "messages" {
		t.Errorf("expected channel 'messages', got %q", recv.Channel)
	}
}

func TestParseAwaitAll(t *testing.T) {
	prog, err := parse("results is await all\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign, ok := prog.Statements[0].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", prog.Statements[0])
	}
	awaitExpr, ok := assign.Value.(*ast.AwaitExpression)
	if !ok {
		t.Fatalf("expected AwaitExpression, got %T", assign.Value)
	}
	ident, ok := awaitExpr.Target.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier target, got %T", awaitExpr.Target)
	}
	if ident.Name != "all" {
		t.Errorf("expected 'all', got %q", ident.Name)
	}
}

func TestParseAwaitFirst(t *testing.T) {
	prog, err := parse("winner is await first\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign, ok := prog.Statements[0].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", prog.Statements[0])
	}
	awaitExpr, ok := assign.Value.(*ast.AwaitExpression)
	if !ok {
		t.Fatalf("expected AwaitExpression, got %T", assign.Value)
	}
	ident, ok := awaitExpr.Target.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier target, got %T", awaitExpr.Target)
	}
	if ident.Name != "first" {
		t.Errorf("expected 'first', got %q", ident.Name)
	}
}
