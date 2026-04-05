package codegen

import (
	"strings"
	"testing"
)

// --- Feature 4: Computed Properties ---

func TestComputedPropertyCodegen(t *testing.T) {
	output, err := compile(`key is "name"
obj is {[key]: "Alice"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `[key]: "Alice"`) {
		t.Errorf("expected computed property syntax [key]: \"Alice\", got:\n%s", output)
	}
}

func TestMixedObjectCodegen(t *testing.T) {
	output, err := compile(`obj is {name: "Bob", [key]: "Alice"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `name: "Bob"`) {
		t.Errorf("expected regular property, got:\n%s", output)
	}
	if !strings.Contains(output, `[key]: "Alice"`) {
		t.Errorf("expected computed property, got:\n%s", output)
	}
}

func TestDynamicPropertyAccessCodegen(t *testing.T) {
	output, err := compile(`x is obj[key]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "obj[key]") {
		t.Errorf("expected bracket notation access, got:\n%s", output)
	}
}

// --- Feature 5: Tagged Templates ---

func TestTaggedTemplateQueryCodegen(t *testing.T) {
	output, err := compile("sql is query`SELECT * FROM users WHERE age > {minAge}`")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "query`SELECT * FROM users WHERE age > ${minAge}`") {
		t.Errorf("expected tagged template with ${} interpolation, got:\n%s", output)
	}
	// Should inject query runtime
	if !strings.Contains(output, "function query(strings, ...values)") {
		t.Errorf("expected query runtime function to be injected, got:\n%s", output)
	}
}

func TestTaggedTemplateHtmlCodegen(t *testing.T) {
	output, err := compile("result is html`<div>{content}</div>`")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "html`<div>${content}</div>`") {
		t.Errorf("expected tagged template with ${} interpolation, got:\n%s", output)
	}
	if !strings.Contains(output, "function html(strings, ...values)") {
		t.Errorf("expected html runtime function to be injected, got:\n%s", output)
	}
}

func TestTaggedTemplateCssCodegen(t *testing.T) {
	output, err := compile("styles is css`color: {color}; font-size: {size}px`")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "css`color: ${color}; font-size: ${size}px`") {
		t.Errorf("expected tagged template with ${} interpolation, got:\n%s", output)
	}
	if !strings.Contains(output, "function css(strings, ...values)") {
		t.Errorf("expected css runtime function to be injected, got:\n%s", output)
	}
}

// --- Feature 6: Private Fields ---

func TestPrivateFieldsCodegen(t *testing.T) {
	input := `describe User:
    private password is "secret"
    public name is "Alice"
    private to hashPassword:
        give back "hashed"
    public to getName:
        give back "test"
`
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Private field declaration
	if !strings.Contains(output, "#password;") {
		t.Errorf("expected #password field declaration, got:\n%s", output)
	}
	// Private field in constructor
	if !strings.Contains(output, "this.#password") {
		t.Errorf("expected this.#password in constructor, got:\n%s", output)
	}
	// Public field (no #)
	if !strings.Contains(output, "this.name") {
		t.Errorf("expected this.name (public), got:\n%s", output)
	}
	// Private method
	if !strings.Contains(output, "#hashPassword(") {
		t.Errorf("expected #hashPassword private method, got:\n%s", output)
	}
	// Public method (no #)
	if !strings.Contains(output, "getName(") {
		t.Errorf("expected getName public method, got:\n%s", output)
	}
}

func TestDescribeWithoutVisibilityCodegen(t *testing.T) {
	input := `describe Animal:
    name is "dog"
    to speak:
        give back "woof"
`
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should not have # prefixes
	if strings.Contains(output, "#name") {
		t.Errorf("expected no # prefix for default visibility, got:\n%s", output)
	}
	if strings.Contains(output, "#speak") {
		t.Errorf("expected no # prefix for default visibility method, got:\n%s", output)
	}
}

// --- Feature 7: Enum Methods ---

func TestEnumMethodsCodegen(t *testing.T) {
	input := `define HttpStatus:
    OK is 200
    NotFound is 404
    ServerError is 500
    to isSuccess:
        give back "yes"
`
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should use class-based enum pattern
	if !strings.Contains(output, "HttpStatusEnum") {
		t.Errorf("expected HttpStatusEnum class, got:\n%s", output)
	}
	if !strings.Contains(output, "constructor(name, value)") {
		t.Errorf("expected constructor(name, value), got:\n%s", output)
	}
	if !strings.Contains(output, "isSuccess(") {
		t.Errorf("expected isSuccess method, got:\n%s", output)
	}
	if !strings.Contains(output, `new HttpStatusEnum("OK", 200)`) {
		t.Errorf("expected new HttpStatusEnum(\"OK\", 200), got:\n%s", output)
	}
	if !strings.Contains(output, `new HttpStatusEnum("NotFound", 404)`) {
		t.Errorf("expected new HttpStatusEnum(\"NotFound\", 404), got:\n%s", output)
	}
	if !strings.Contains(output, "Object.freeze") {
		t.Errorf("expected Object.freeze, got:\n%s", output)
	}
}

func TestEnumWithValuesNoMethodsCodegen(t *testing.T) {
	input := `define Priority:
    Low is 1
    Medium is 2
    High is 3
`
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be simple frozen object, not class-based
	if strings.Contains(output, "PriorityEnum") {
		t.Errorf("expected simple frozen object (no class), got:\n%s", output)
	}
	if !strings.Contains(output, "Object.freeze") {
		t.Errorf("expected Object.freeze, got:\n%s", output)
	}
	if !strings.Contains(output, "value: 1") {
		t.Errorf("expected value: 1 for Low, got:\n%s", output)
	}
}

func TestSimpleEnumCodegen(t *testing.T) {
	input := `define Color:
    Red
    Green
    Blue
`
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Color") {
		t.Errorf("expected Color enum, got:\n%s", output)
	}
	if !strings.Contains(output, "Red") {
		t.Errorf("expected Red variant, got:\n%s", output)
	}
}
