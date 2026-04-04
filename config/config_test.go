package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTOMLBasic(t *testing.T) {
	toml := `
[project]
name = "my-app"
version = "1.0.0"
description = "My Quill application"
author = "Developer Name"
license = "MIT"

[build]
target = "js"
entry = "src/main.quill"
output = "dist/"
minify = false
sourcemap = true

[dependencies]
http-server = "^1.0.0"
json-parser = "~2.1.0"

[dev-dependencies]
test-runner = "*"

[test]
pattern = "tests/**/*.quill"
timeout = 30000

[lsp]
port = 7998
`
	cfg, err := ParseTOML(toml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Project.Name != "my-app" {
		t.Errorf("expected name 'my-app', got %q", cfg.Project.Name)
	}
	if cfg.Project.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", cfg.Project.Version)
	}
	if cfg.Project.Description != "My Quill application" {
		t.Errorf("expected description 'My Quill application', got %q", cfg.Project.Description)
	}
	if cfg.Project.Author != "Developer Name" {
		t.Errorf("expected author 'Developer Name', got %q", cfg.Project.Author)
	}
	if cfg.Project.License != "MIT" {
		t.Errorf("expected license 'MIT', got %q", cfg.Project.License)
	}

	if cfg.Build.Target != "js" {
		t.Errorf("expected target 'js', got %q", cfg.Build.Target)
	}
	if cfg.Build.Entry != "src/main.quill" {
		t.Errorf("expected entry 'src/main.quill', got %q", cfg.Build.Entry)
	}
	if cfg.Build.Output != "dist/" {
		t.Errorf("expected output 'dist/', got %q", cfg.Build.Output)
	}
	if cfg.Build.Minify != false {
		t.Error("expected minify false")
	}
	if cfg.Build.SourceMap != true {
		t.Error("expected sourcemap true")
	}

	if v, ok := cfg.Dependencies["http-server"]; !ok || v != "^1.0.0" {
		t.Errorf("expected dependency http-server='^1.0.0', got %q", v)
	}
	if v, ok := cfg.Dependencies["json-parser"]; !ok || v != "~2.1.0" {
		t.Errorf("expected dependency json-parser='~2.1.0', got %q", v)
	}

	if v, ok := cfg.DevDependencies["test-runner"]; !ok || v != "*" {
		t.Errorf("expected dev-dependency test-runner='*', got %q", v)
	}

	if cfg.Test.Pattern != "tests/**/*.quill" {
		t.Errorf("expected test pattern 'tests/**/*.quill', got %q", cfg.Test.Pattern)
	}
	if cfg.Test.Timeout != 30000 {
		t.Errorf("expected timeout 30000, got %d", cfg.Test.Timeout)
	}

	if cfg.LSP.Port != 7998 {
		t.Errorf("expected lsp port 7998, got %d", cfg.LSP.Port)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Project.Version != "0.1.0" {
		t.Errorf("expected default version '0.1.0', got %q", cfg.Project.Version)
	}
	if cfg.Build.Target != "js" {
		t.Errorf("expected default target 'js', got %q", cfg.Build.Target)
	}
	if cfg.Build.SourceMap != true {
		t.Error("expected default sourcemap true")
	}
	if cfg.Test.Timeout != 30000 {
		t.Errorf("expected default timeout 30000, got %d", cfg.Test.Timeout)
	}
	if cfg.LSP.Port != 7998 {
		t.Errorf("expected default lsp port 7998, got %d", cfg.LSP.Port)
	}
}

func TestParseTOMLMissingOptionalFields(t *testing.T) {
	toml := `
[project]
name = "minimal"
version = "0.1.0"
`
	cfg, err := ParseTOML(toml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Project.Name != "minimal" {
		t.Errorf("expected name 'minimal', got %q", cfg.Project.Name)
	}
	// Should use defaults for missing fields
	if cfg.Build.Target != "js" {
		t.Errorf("expected default target 'js', got %q", cfg.Build.Target)
	}
	if cfg.Build.Entry != "main.quill" {
		t.Errorf("expected default entry 'main.quill', got %q", cfg.Build.Entry)
	}
}

func TestParseTOMLComments(t *testing.T) {
	toml := `
# This is a comment
[project]
name = "test"
# Another comment
version = "1.0.0"
`
	cfg, err := ParseTOML(toml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Project.Name != "test" {
		t.Errorf("expected name 'test', got %q", cfg.Project.Name)
	}
}

func TestParseTOMLInvalidLine(t *testing.T) {
	toml := `
[project]
this is not valid
`
	_, err := ParseTOML(toml)
	if err == nil {
		t.Error("expected error for invalid TOML line")
	}
}

func TestGenerateTOML(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Project.Name = "roundtrip-test"
	cfg.Project.Version = "2.0.0"
	cfg.Dependencies["some-lib"] = "^1.0.0"

	toml := GenerateTOML(cfg)

	// Parse it back
	parsed, err := ParseTOML(toml)
	if err != nil {
		t.Fatalf("failed to parse generated TOML: %v", err)
	}
	if parsed.Project.Name != "roundtrip-test" {
		t.Errorf("roundtrip failed: expected name 'roundtrip-test', got %q", parsed.Project.Name)
	}
	if parsed.Project.Version != "2.0.0" {
		t.Errorf("roundtrip failed: expected version '2.0.0', got %q", parsed.Project.Version)
	}
	if v, ok := parsed.Dependencies["some-lib"]; !ok || v != "^1.0.0" {
		t.Errorf("roundtrip failed: expected dependency 'some-lib=^1.0.0', got %q", v)
	}
}

func TestLoadConfigFromDirectory(t *testing.T) {
	// Create a temp dir with a quill.toml
	tmpDir := t.TempDir()
	tomlContent := `[project]
name = "load-test"
version = "0.5.0"
`
	err := os.WriteFile(filepath.Join(tmpDir, "quill.toml"), []byte(tomlContent), 0644)
	if err != nil {
		t.Fatalf("failed to write quill.toml: %v", err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.Project.Name != "load-test" {
		t.Errorf("expected name 'load-test', got %q", cfg.Project.Name)
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := LoadConfig(tmpDir)
	if err == nil {
		t.Error("expected error when quill.toml not found")
	}
}

func TestParseTOMLBooleanValues(t *testing.T) {
	toml := `
[build]
minify = true
sourcemap = false
`
	cfg, err := ParseTOML(toml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Build.Minify != true {
		t.Error("expected minify true")
	}
	if cfg.Build.SourceMap != false {
		t.Error("expected sourcemap false")
	}
}
