package typechecker

import (
	"quill/ast"
	"testing"
)

func TestTypeAliasRegistration(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.DescribeStatement{
				Name:       "User",
				Properties: []ast.AssignStatement{},
				Methods:    []ast.FuncDefinition{},
				Line:       1,
			},
			&ast.TypeAliasStatement{
				Name:     "UserUpdate",
				BaseType: "User",
				Utility:  "Partial",
				Args:     nil,
				Line:     5,
			},
		},
	}

	tc := New()
	diags := tc.Check(program)

	// Should register the type alias
	if !tc.types["UserUpdate"] {
		t.Error("expected UserUpdate to be registered as a type")
	}

	// Should have no errors
	for _, d := range diags {
		if d.Severity == "error" {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}

	// Verify alias info stored
	alias, ok := tc.typeAliases["UserUpdate"]
	if !ok {
		t.Fatal("expected type alias info to be stored")
	}
	if alias.Utility != "Partial" {
		t.Errorf("expected utility 'Partial', got %q", alias.Utility)
	}
	if alias.BaseType != "User" {
		t.Errorf("expected base type 'User', got %q", alias.BaseType)
	}
}

func TestPartialUtilityType(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.DescribeStatement{
				Name:       "User",
				Properties: []ast.AssignStatement{},
				Methods:    []ast.FuncDefinition{},
				Line:       1,
			},
			&ast.TypeAliasStatement{
				Name:     "PartialUser",
				BaseType: "User",
				Utility:  "Partial",
				Args:     nil,
				Line:     3,
			},
		},
	}

	tc := New()
	diags := tc.Check(program)

	// Partial of a known type should not produce errors
	for _, d := range diags {
		if d.Severity == "error" {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}

	if !tc.types["PartialUser"] {
		t.Error("expected PartialUser to be a known type")
	}
}

func TestOmitUtilityType(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.DescribeStatement{
				Name:       "User",
				Properties: []ast.AssignStatement{},
				Methods:    []ast.FuncDefinition{},
				Line:       1,
			},
			&ast.TypeAliasStatement{
				Name:     "PublicUser",
				BaseType: "User",
				Utility:  "Omit",
				Args:     []string{"password", "email"},
				Line:     3,
			},
		},
	}

	tc := New()
	diags := tc.Check(program)

	for _, d := range diags {
		if d.Severity == "error" {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}

	if !tc.types["PublicUser"] {
		t.Error("expected PublicUser to be a known type")
	}

	alias := tc.typeAliases["PublicUser"]
	if alias.Utility != "Omit" {
		t.Errorf("expected utility 'Omit', got %q", alias.Utility)
	}
	if len(alias.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(alias.Args))
	}
	if alias.Args[0] != "password" || alias.Args[1] != "email" {
		t.Errorf("expected args [password, email], got %v", alias.Args)
	}
}

func TestRecordUtilityType(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.TypeAliasStatement{
				Name:     "StringMap",
				BaseType: "text",
				Utility:  "Record",
				Args:     []string{"number"},
				Line:     1,
			},
		},
	}

	tc := New()
	diags := tc.Check(program)

	// Record of text, number should not produce errors (base type is a primitive)
	for _, d := range diags {
		if d.Severity == "error" {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}

	if !tc.types["StringMap"] {
		t.Error("expected StringMap to be a known type")
	}

	alias := tc.typeAliases["StringMap"]
	if alias.Utility != "Record" {
		t.Errorf("expected utility 'Record', got %q", alias.Utility)
	}
	if alias.BaseType != "text" {
		t.Errorf("expected base type 'text', got %q", alias.BaseType)
	}
	if len(alias.Args) != 1 || alias.Args[0] != "number" {
		t.Errorf("expected args [number], got %v", alias.Args)
	}
}

func TestUnknownBaseTypeWarning(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.TypeAliasStatement{
				Name:     "BadAlias",
				BaseType: "NonExistent",
				Utility:  "Partial",
				Args:     nil,
				Line:     1,
			},
		},
	}

	tc := New()
	diags := tc.Check(program)

	hasWarning := false
	for _, d := range diags {
		if d.Severity == "warning" && d.Line == 1 {
			hasWarning = true
		}
	}

	if !hasWarning {
		t.Error("expected warning about unknown base type")
	}
}
