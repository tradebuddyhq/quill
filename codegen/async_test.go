package codegen

import (
	"strings"
	"testing"
)

func TestCancelGeneratesAbortController(t *testing.T) {
	input := "spawn task fetcher:\n    result is 42\ncancel fetcher\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__abort_fetcher = new AbortController()") {
		t.Errorf("expected AbortController for spawned task, got:\n%s", output)
	}
	if !strings.Contains(output, "__abort_fetcher.abort()") {
		t.Errorf("expected abort() call, got:\n%s", output)
	}
}

func TestAsyncIterationGeneratesForAwait(t *testing.T) {
	input := "for await each chunk in stream:\n    say chunk\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "for await (const chunk of stream)") {
		t.Errorf("expected 'for await' loop, got:\n%s", output)
	}
}

func TestAllSettledGeneratesPromiseAllSettled(t *testing.T) {
	input := "parallel settled:\n    task1 is fetch(\"/a\")\n    task2 is fetch(\"/b\")\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Promise.allSettled") {
		t.Errorf("expected Promise.allSettled, got:\n%s", output)
	}
}

func TestPropagateGeneratesErrorCheck(t *testing.T) {
	input := "data is fetchJSON(\"/api\")?\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__propagate(") {
		t.Errorf("expected __propagate() call, got:\n%s", output)
	}
}

func TestResultRuntimeInjection(t *testing.T) {
	input := "data is fetchJSON(\"/api\")?\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function Success(") {
		t.Errorf("expected Success runtime helper, got:\n%s", output)
	}
	if !strings.Contains(output, "function __QuillError(") {
		t.Errorf("expected __QuillError runtime helper, got:\n%s", output)
	}
	if !strings.Contains(output, "function __propagate(") {
		t.Errorf("expected __propagate runtime helper, got:\n%s", output)
	}
}

func TestDestructureInLoopGeneratesCorrectJS(t *testing.T) {
	input := "for each {name, age} in users:\n    say name\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "for (const {name, age} of users)") {
		t.Errorf("expected destructured for-of loop, got:\n%s", output)
	}
}
