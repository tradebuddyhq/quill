package codegen

import (
	"quill/ast"
	"strings"
	"testing"
)

func TestDecoratorFunctionWrapping(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.DecoratedFuncDefinition{
				Decorators: []ast.Decorator{
					{Name: "log", Line: 1},
				},
				Func: &ast.FuncDefinition{
					Name:   "processOrder",
					Params: []string{"order"},
					Body: []ast.Statement{
						&ast.SayStatement{
							Value: &ast.StringLiteral{Value: "Processing"},
							Line:  3,
						},
					},
					Line: 2,
				},
				Line: 1,
			},
		},
	}

	gen := New()
	output := gen.Generate(program)

	if !strings.Contains(output, "let processOrder = function(order)") {
		t.Errorf("expected function expression, got:\n%s", output)
	}
	if !strings.Contains(output, "processOrder = log(processOrder)") {
		t.Errorf("expected decorator wrapping, got:\n%s", output)
	}
}

func TestDecoratorRouteMiddleware(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.DecoratedRouteDefinition{
				Decorators: []ast.Decorator{
					{Name: "authenticated", Line: 1},
				},
				Route: ast.RouteDefinition{
					Method: "get",
					Path:   "/api/admin",
					Body: []ast.Statement{
						&ast.RespondStatement{
							Value:      &ast.ObjectLiteral{Keys: []string{"data"}, Values: []ast.Expression{&ast.StringLiteral{Value: "secret"}}},
							StatusCode: 200,
							Line:       3,
						},
					},
					Line: 2,
				},
				Line: 1,
			},
		},
	}

	gen := New()
	output := gen.Generate(program)

	if !strings.Contains(output, "method === 'get'") {
		t.Errorf("expected method check, got:\n%s", output)
	}
	if !strings.Contains(output, "/api/admin") {
		t.Errorf("expected path check, got:\n%s", output)
	}
	if !strings.Contains(output, "authenticated(req)") {
		t.Errorf("expected authenticated middleware check, got:\n%s", output)
	}
}

func TestDecoratorWithArguments(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.DecoratedRouteDefinition{
				Decorators: []ast.Decorator{
					{Name: "rateLimit", Args: []ast.Expression{&ast.NumberLiteral{Value: 100}}, Line: 1},
				},
				Route: ast.RouteDefinition{
					Method: "get",
					Path:   "/api/data",
					Body: []ast.Statement{
						&ast.RespondStatement{
							Value:      &ast.ObjectLiteral{Keys: []string{"ok"}, Values: []ast.Expression{&ast.BoolLiteral{Value: true}}},
							StatusCode: 200,
							Line:       3,
						},
					},
					Line: 2,
				},
				Line: 1,
			},
		},
	}

	gen := New()
	output := gen.Generate(program)

	if !strings.Contains(output, "rateLimit(100)(req)") {
		t.Errorf("expected rateLimit with argument, got:\n%s", output)
	}
	if !strings.Contains(output, "429") {
		t.Errorf("expected 429 status code for rate limit, got:\n%s", output)
	}
}

func TestMultipleDecoratorsChain(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.DecoratedFuncDefinition{
				Decorators: []ast.Decorator{
					{Name: "log", Line: 1},
					{Name: "memoize", Line: 2},
				},
				Func: &ast.FuncDefinition{
					Name:   "compute",
					Params: []string{"x"},
					Body: []ast.Statement{
						&ast.ReturnStatement{
							Value: &ast.Identifier{Name: "x"},
							Line:  4,
						},
					},
					Line: 3,
				},
				Line: 1,
			},
		},
	}

	gen := New()
	output := gen.Generate(program)

	// Decorators should be applied in reverse order (innermost first)
	memoizeIdx := strings.Index(output, "compute = memoize(compute)")
	logIdx := strings.Index(output, "compute = log(compute)")
	if memoizeIdx == -1 {
		t.Error("expected memoize decorator wrapping")
	}
	if logIdx == -1 {
		t.Error("expected log decorator wrapping")
	}
	if memoizeIdx > logIdx {
		t.Error("expected memoize applied before log (reverse order)")
	}
}

func TestDecoratorRuntimeInjection(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.DecoratedFuncDefinition{
				Decorators: []ast.Decorator{
					{Name: "log", Line: 1},
				},
				Func: &ast.FuncDefinition{
					Name:   "test",
					Params: []string{},
					Body: []ast.Statement{
						&ast.SayStatement{Value: &ast.StringLiteral{Value: "hello"}, Line: 3},
					},
					Line: 2,
				},
				Line: 1,
			},
		},
	}

	gen := New()
	output := gen.Generate(program)

	if !strings.Contains(output, "function authenticated(req)") {
		t.Error("expected authenticated runtime function")
	}
	if !strings.Contains(output, "function rateLimit(max)") {
		t.Error("expected rateLimit runtime function")
	}
	if !strings.Contains(output, "function log(fn)") {
		t.Error("expected log runtime function")
	}
}
