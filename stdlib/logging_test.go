package stdlib

import (
	"strings"
	"testing"
)

func TestLoggingRuntimeNonEmpty(t *testing.T) {
	runtime := GetLoggingRuntime()
	if len(runtime) == 0 {
		t.Fatal("Logging runtime should not be empty")
	}
}

func TestLoggingRuntimeContainsLogObject(t *testing.T) {
	runtime := GetLoggingRuntime()
	if !strings.Contains(runtime, "var Log = {}") {
		t.Error("Logging runtime should define Log object")
	}
}

func TestLoggingRuntimeContainsCreate(t *testing.T) {
	runtime := GetLoggingRuntime()
	if !strings.Contains(runtime, "Log.create") {
		t.Error("Logging runtime should contain Log.create function")
	}
}

func TestLoggingRuntimeContainsMiddleware(t *testing.T) {
	runtime := GetLoggingRuntime()
	if !strings.Contains(runtime, "Log.middleware") {
		t.Error("Logging runtime should contain Log.middleware function")
	}
}

func TestLoggingRuntimeContainsLoggerClass(t *testing.T) {
	runtime := GetLoggingRuntime()
	if !strings.Contains(runtime, "function Logger(") {
		t.Error("Logging runtime should contain Logger class")
	}
}

func TestLoggingRuntimeContainsDebug(t *testing.T) {
	runtime := GetLoggingRuntime()
	if !strings.Contains(runtime, "Logger.prototype.debug") {
		t.Error("Logging runtime should contain Logger.prototype.debug")
	}
}

func TestLoggingRuntimeContainsInfo(t *testing.T) {
	runtime := GetLoggingRuntime()
	if !strings.Contains(runtime, "Logger.prototype.info") {
		t.Error("Logging runtime should contain Logger.prototype.info")
	}
}

func TestLoggingRuntimeContainsWarn(t *testing.T) {
	runtime := GetLoggingRuntime()
	if !strings.Contains(runtime, "Logger.prototype.warn") {
		t.Error("Logging runtime should contain Logger.prototype.warn")
	}
}

func TestLoggingRuntimeContainsError(t *testing.T) {
	runtime := GetLoggingRuntime()
	if !strings.Contains(runtime, "Logger.prototype.error") {
		t.Error("Logging runtime should contain Logger.prototype.error")
	}
}

func TestLoggingRuntimeContainsFatal(t *testing.T) {
	runtime := GetLoggingRuntime()
	if !strings.Contains(runtime, "Logger.prototype.fatal") {
		t.Error("Logging runtime should contain Logger.prototype.fatal")
	}
}

func TestLoggingRuntimeContainsChild(t *testing.T) {
	runtime := GetLoggingRuntime()
	if !strings.Contains(runtime, "Logger.prototype.child") {
		t.Error("Logging runtime should contain Logger.prototype.child")
	}
}

func TestLoggingRuntimeContainsLevels(t *testing.T) {
	runtime := GetLoggingRuntime()
	for _, level := range []string{"debug", "info", "warn", "error", "fatal"} {
		if !strings.Contains(runtime, level) {
			t.Errorf("Logging runtime should reference level %q", level)
		}
	}
}

func TestLoggingRuntimeContainsJSONFormat(t *testing.T) {
	runtime := GetLoggingRuntime()
	if !strings.Contains(runtime, "_outputJSON") {
		t.Error("Logging runtime should contain JSON output mode")
	}
}

func TestLoggingRuntimeContainsPrettyFormat(t *testing.T) {
	runtime := GetLoggingRuntime()
	if !strings.Contains(runtime, "_outputPretty") {
		t.Error("Logging runtime should contain pretty output mode")
	}
}

func TestLoggingRuntimeContainsColors(t *testing.T) {
	runtime := GetLoggingRuntime()
	if !strings.Contains(runtime, "COLORS") {
		t.Error("Logging runtime should contain COLORS for colored output")
	}
}
