package stdlib

import (
	"strings"
	"testing"
)

func TestValidateRuntimeNonEmpty(t *testing.T) {
	runtime := GetValidateRuntime()
	if len(runtime) == 0 {
		t.Fatal("Validate runtime should not be empty")
	}
}

func TestValidateRuntimeContainsValidateObject(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "var Validate = {}") {
		t.Error("Validate runtime should define Validate object")
	}
}

func TestValidateRuntimeContainsSchema(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.schema") {
		t.Error("Validate runtime should contain Validate.schema function")
	}
}

func TestValidateRuntimeContainsRequired(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.required") {
		t.Error("Validate runtime should contain Validate.required rule")
	}
}

func TestValidateRuntimeContainsMinLength(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.minLength") {
		t.Error("Validate runtime should contain Validate.minLength rule")
	}
}

func TestValidateRuntimeContainsMaxLength(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.maxLength") {
		t.Error("Validate runtime should contain Validate.maxLength rule")
	}
}

func TestValidateRuntimeContainsMin(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.min") {
		t.Error("Validate runtime should contain Validate.min rule")
	}
}

func TestValidateRuntimeContainsMax(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.max") {
		t.Error("Validate runtime should contain Validate.max rule")
	}
}

func TestValidateRuntimeContainsEmail(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.email") {
		t.Error("Validate runtime should contain Validate.email rule")
	}
}

func TestValidateRuntimeContainsURL(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.url") {
		t.Error("Validate runtime should contain Validate.url rule")
	}
}

func TestValidateRuntimeContainsPattern(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.pattern") {
		t.Error("Validate runtime should contain Validate.pattern rule")
	}
}

func TestValidateRuntimeContainsOneOf(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.oneOf") {
		t.Error("Validate runtime should contain Validate.oneOf rule")
	}
}

func TestValidateRuntimeContainsCustom(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.custom") {
		t.Error("Validate runtime should contain Validate.custom rule")
	}
}

func TestValidateRuntimeContainsArrayOf(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.arrayOf") {
		t.Error("Validate runtime should contain Validate.arrayOf rule")
	}
}

func TestValidateRuntimeContainsIsEmail(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.isEmail") {
		t.Error("Validate runtime should contain Validate.isEmail quick validator")
	}
}

func TestValidateRuntimeContainsIsURL(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.isURL") {
		t.Error("Validate runtime should contain Validate.isURL quick validator")
	}
}

func TestValidateRuntimeContainsIsNumber(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "Validate.isNumber") {
		t.Error("Validate runtime should contain Validate.isNumber quick validator")
	}
}

func TestValidateRuntimeContainsValidationSchema(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "ValidationSchema") {
		t.Error("Validate runtime should contain ValidationSchema class")
	}
}

func TestValidateRuntimeContainsCheck(t *testing.T) {
	runtime := GetValidateRuntime()
	if !strings.Contains(runtime, "ValidationSchema.prototype.check") {
		t.Error("Validate runtime should contain ValidationSchema.prototype.check")
	}
}
