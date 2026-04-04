package stdlib

import (
	"strings"
	"testing"
)

func TestAuthRuntimeNonEmpty(t *testing.T) {
	runtime := GetAuthRuntime()
	if len(runtime) == 0 {
		t.Fatal("Auth runtime should not be empty")
	}
}

func TestAuthRuntimeContainsAuthObject(t *testing.T) {
	runtime := GetAuthRuntime()
	if !strings.Contains(runtime, "var Auth = {}") {
		t.Error("Auth runtime should define Auth object")
	}
}

func TestAuthRuntimeContainsHash(t *testing.T) {
	runtime := GetAuthRuntime()
	if !strings.Contains(runtime, "Auth.hash") {
		t.Error("Auth runtime should contain Auth.hash function")
	}
}

func TestAuthRuntimeContainsVerify(t *testing.T) {
	runtime := GetAuthRuntime()
	if !strings.Contains(runtime, "Auth.verify") {
		t.Error("Auth runtime should contain Auth.verify function")
	}
}

func TestAuthRuntimeContainsCreateToken(t *testing.T) {
	runtime := GetAuthRuntime()
	if !strings.Contains(runtime, "Auth.createToken") {
		t.Error("Auth runtime should contain Auth.createToken function")
	}
}

func TestAuthRuntimeContainsVerifyToken(t *testing.T) {
	runtime := GetAuthRuntime()
	if !strings.Contains(runtime, "Auth.verifyToken") {
		t.Error("Auth runtime should contain Auth.verifyToken function")
	}
}

func TestAuthRuntimeContainsSession(t *testing.T) {
	runtime := GetAuthRuntime()
	if !strings.Contains(runtime, "Auth.session") {
		t.Error("Auth runtime should contain Auth.session function")
	}
}

func TestAuthRuntimeContainsDeriveKey(t *testing.T) {
	runtime := GetAuthRuntime()
	if !strings.Contains(runtime, "_deriveKey") {
		t.Error("Auth runtime should contain _deriveKey for password hashing")
	}
}

func TestAuthRuntimeContainsBase64(t *testing.T) {
	runtime := GetAuthRuntime()
	if !strings.Contains(runtime, "_base64Encode") {
		t.Error("Auth runtime should contain _base64Encode for JWT tokens")
	}
}

func TestAuthRuntimeContainsHMAC(t *testing.T) {
	runtime := GetAuthRuntime()
	if !strings.Contains(runtime, "_hmacSign") {
		t.Error("Auth runtime should contain _hmacSign for token signatures")
	}
}
