package codegen

import (
	"quill/stdlib"
	"strings"
	"testing"
)

func TestGeneratorFunctionCompilesToFunctionStar(t *testing.T) {
	output, err := compile("to gen:\n    yield 42\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function* gen()") {
		t.Errorf("expected function* gen(), got:\n%s", output)
	}
}

func TestYieldCompilesCorrectly(t *testing.T) {
	output, err := compile("to gen:\n    yield 42\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "yield 42;") {
		t.Errorf("expected 'yield 42;', got:\n%s", output)
	}
}

func TestLoopCompilesToWhileTrue(t *testing.T) {
	output, err := compile("to gen:\n    loop:\n        yield 1\n        break\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "while (true)") {
		t.Errorf("expected 'while (true)', got:\n%s", output)
	}
}

func TestGeneratorWithParams(t *testing.T) {
	src := `to range_gen start end:
    i is start
    loop:
        yield i
        i is i + 1
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function* range_gen(start, end)") {
		t.Errorf("expected function* range_gen(start, end), got:\n%s", output)
	}
}

func TestNonGeneratorFunctionNotStar(t *testing.T) {
	output, err := compile("to add a b:\n    give back a + b\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(output, "function*") {
		t.Errorf("non-generator function should not be function*, got:\n%s", output)
	}
	if !strings.Contains(output, "function add(a, b)") {
		t.Errorf("expected function add(a, b), got:\n%s", output)
	}
}

func TestLazyRuntimeContainsMethods(t *testing.T) {
	runtime := stdlib.GetIteratorRuntime()

	methods := []string{"filter(", "map(", "take(", "skip(", "takeWhile(", "skipWhile(",
		"zip(", "enumerate(", "flatten(", "collect()", "reduce(", "forEach(",
		"count()", "first()", "last()", "any(", "every("}
	for _, method := range methods {
		if !strings.Contains(runtime, method) {
			t.Errorf("iterator runtime missing method %q", method)
		}
	}
}

func TestLazyRuntimeContainsQuillLazy(t *testing.T) {
	runtime := stdlib.GetIteratorRuntime()
	if !strings.Contains(runtime, "class __QuillLazy") {
		t.Error("iterator runtime missing __QuillLazy class")
	}
	if !strings.Contains(runtime, "function __quill_lazy") {
		t.Error("iterator runtime missing __quill_lazy function")
	}
}

func TestRangeFunctionGeneration(t *testing.T) {
	runtime := stdlib.GetIteratorRuntime()
	if !strings.Contains(runtime, "function* __quill_range") {
		t.Error("iterator runtime missing __quill_range generator function")
	}
}

func TestIteratorRuntimeInjection(t *testing.T) {
	output, err := compile("to gen:\n    yield 1\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__QuillLazy") {
		t.Errorf("expected iterator runtime to be injected for generator function, got:\n%s", output)
	}
}

func TestFibonacciGenerator(t *testing.T) {
	src := `to fibonacci:
    a is 0
    b is 1
    loop:
        yield a
        next is a + b
        a is b
        b is next
`
	output, err := compile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function* fibonacci()") {
		t.Errorf("expected function* fibonacci(), got:\n%s", output)
	}
	if !strings.Contains(output, "yield a;") {
		t.Errorf("expected yield a;, got:\n%s", output)
	}
	if !strings.Contains(output, "while (true)") {
		t.Errorf("expected while (true), got:\n%s", output)
	}
}
