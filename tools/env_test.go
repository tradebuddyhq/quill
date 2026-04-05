package tools

import (
	"strings"
	"testing"
)

func TestParseEnvStringBasic(t *testing.T) {
	content := `
KEY1=value1
KEY2=value2
`
	vars := ParseEnvString(content)

	if vars["KEY1"] != "value1" {
		t.Errorf("expected KEY1=value1, got %q", vars["KEY1"])
	}
	if vars["KEY2"] != "value2" {
		t.Errorf("expected KEY2=value2, got %q", vars["KEY2"])
	}
}

func TestParseEnvStringQuoted(t *testing.T) {
	content := `
KEY1="quoted value"
KEY2='single quoted'
`
	vars := ParseEnvString(content)

	if vars["KEY1"] != "quoted value" {
		t.Errorf("expected KEY1='quoted value', got %q", vars["KEY1"])
	}
	if vars["KEY2"] != "single quoted" {
		t.Errorf("expected KEY2='single quoted', got %q", vars["KEY2"])
	}
}

func TestParseEnvStringComments(t *testing.T) {
	content := `
# This is a comment
KEY1=value1
# Another comment
KEY2=value2
`
	vars := ParseEnvString(content)

	if len(vars) != 2 {
		t.Errorf("expected 2 vars, got %d", len(vars))
	}
	if vars["KEY1"] != "value1" {
		t.Errorf("expected KEY1=value1, got %q", vars["KEY1"])
	}
}

func TestParseEnvStringExport(t *testing.T) {
	content := `
export KEY1=value1
export KEY2="value2"
`
	vars := ParseEnvString(content)

	if vars["KEY1"] != "value1" {
		t.Errorf("expected KEY1=value1, got %q", vars["KEY1"])
	}
	if vars["KEY2"] != "value2" {
		t.Errorf("expected KEY2=value2, got %q", vars["KEY2"])
	}
}

func TestValidateRequired(t *testing.T) {
	vars := map[string]string{
		"DATABASE_URL": "postgres://localhost",
		"PORT":         "3000",
	}

	// All present
	missing := ValidateRequired(vars, []string{"DATABASE_URL", "PORT"})
	if len(missing) != 0 {
		t.Errorf("expected no missing keys, got %v", missing)
	}

	// Some missing
	missing = ValidateRequired(vars, []string{"DATABASE_URL", "API_KEY", "SECRET"})
	if len(missing) != 2 {
		t.Errorf("expected 2 missing keys, got %v", missing)
	}
}

func TestGenerateEnvInjection(t *testing.T) {
	vars := map[string]string{
		"DATABASE_URL": "postgres://localhost",
		"API_KEY":      "secret123",
	}

	js := GenerateEnvInjection(vars)

	if !strings.Contains(js, "const __env = {") {
		t.Error("missing __env declaration")
	}
	if !strings.Contains(js, `DATABASE_URL: "postgres://localhost"`) {
		t.Error("missing DATABASE_URL")
	}
	if !strings.Contains(js, `API_KEY: "secret123"`) {
		t.Error("missing API_KEY")
	}
	if !strings.Contains(js, "require(key)") {
		t.Error("missing require method")
	}
	if !strings.Contains(js, "get(key, defaultValue)") {
		t.Error("missing get method")
	}
	if !strings.Contains(js, "const env = __env;") {
		t.Error("missing env alias")
	}
}

func TestParseEnvStringEmpty(t *testing.T) {
	vars := ParseEnvString("")
	if len(vars) != 0 {
		t.Errorf("expected 0 vars from empty string, got %d", len(vars))
	}
}

func TestParseEnvStringNoEquals(t *testing.T) {
	content := `
INVALID_LINE
KEY=value
`
	vars := ParseEnvString(content)
	if len(vars) != 1 {
		t.Errorf("expected 1 var, got %d", len(vars))
	}
	if vars["KEY"] != "value" {
		t.Errorf("expected KEY=value, got %q", vars["KEY"])
	}
}
