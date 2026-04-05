package codegen

import (
	"quill/ast"
	"strings"
	"testing"
)

func TestMockStatementGeneratesMockFunction(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.MockStatement{
				FuncName: "fetchJSON",
				Params:   []string{"url"},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.ObjectLiteral{
							Keys:   []string{"id", "name"},
							Values: []ast.Expression{&ast.NumberLiteral{Value: 1}, &ast.StringLiteral{Value: "Alice"}},
						},
						Line: 3,
					},
				},
				Line: 1,
			},
		},
	}

	gen := New()
	js := gen.Generate(program)

	if !strings.Contains(js, "const __mock_fetchJSON_calls = []") {
		t.Error("mock should create call tracking array")
	}
	if !strings.Contains(js, "const __original_fetchJSON = typeof fetchJSON !== 'undefined' ? fetchJSON : undefined") {
		t.Error("mock should save original function")
	}
	if !strings.Contains(js, "var fetchJSON = function(url)") {
		t.Error("mock should create mock function")
	}
}

func TestMockCallTracking(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.MockStatement{
				FuncName: "fetchJSON",
				Params:   []string{"url"},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.NumberLiteral{Value: 42},
						Line:  3,
					},
				},
				Line: 1,
			},
		},
	}

	gen := New()
	js := gen.Generate(program)

	if !strings.Contains(js, "__mock_fetchJSON_calls.push({args: [url]})") {
		t.Error("mock should track calls with arguments")
	}
}

func TestMockAssertionCalledNTimes(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpectStatement{
				Expr: &ast.MockAssertionExpr{
					FuncName:   "fetchJSON",
					AssertType: "called",
					Count:      2,
				},
				Line: 5,
			},
		},
	}

	gen := New()
	js := gen.Generate(program)

	if !strings.Contains(js, "__mock_fetchJSON_calls.length !== 2") {
		t.Error("assertion should check call count")
	}
	if !strings.Contains(js, "throw new Error") {
		t.Error("assertion should throw on failure")
	}
}

func TestMockRestoreAfterTest(t *testing.T) {
	// The mock generates __original_<func> which can be used to restore
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.MockStatement{
				FuncName: "myFunc",
				Params:   []string{"x"},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.NumberLiteral{Value: 0},
						Line:  3,
					},
				},
				Line: 1,
			},
		},
	}

	gen := New()
	js := gen.Generate(program)

	if !strings.Contains(js, "__original_myFunc") {
		t.Error("mock should save original for restoration")
	}
}

func TestMockWithThrow(t *testing.T) {
	// In Quill, "throw" in a mock body compiles to a throw statement.
	// We simulate this via an ExprStatement with a string for simplicity.
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.MockStatement{
				FuncName: "fetchJSON",
				Params:   []string{"url"},
				Body: []ast.Statement{
					// "throw" in Quill is parsed as an expression statement
					// that generates: throw "Network error"
					// We approximate with a ReturnStatement that we check
					// the mock structure is generated correctly
					&ast.ExprStatement{
						Expr: &ast.CallExpr{
							Function: &ast.Identifier{Name: "throw"},
							Args:     []ast.Expression{&ast.StringLiteral{Value: "Network error"}},
						},
						Line: 3,
					},
				},
				Line: 1,
			},
		},
	}

	gen := New()
	js := gen.Generate(program)

	// The mock should still generate the wrapper with call tracking
	if !strings.Contains(js, "__mock_fetchJSON_calls") {
		t.Error("mock with throw should still track calls")
	}
	if !strings.Contains(js, "var fetchJSON = function(url)") {
		t.Error("mock with throw should create mock function")
	}
}

func TestMockNoParams(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.MockStatement{
				FuncName: "getData",
				Params:   []string{},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.NumberLiteral{Value: 42},
						Line:  3,
					},
				},
				Line: 1,
			},
		},
	}

	gen := New()
	js := gen.Generate(program)

	if !strings.Contains(js, "var getData = function()") {
		t.Error("mock with no params should have empty param list")
	}
	if !strings.Contains(js, "__mock_getData_calls.push({args: []})") {
		t.Error("mock with no params should push empty args array")
	}
}
