package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents a quill.toml project configuration.
type Config struct {
	Project         ProjectConfig
	Build           BuildConfig
	Dependencies    map[string]string
	DevDependencies map[string]string
	Test            TestConfig
	LSP             LSPConfig
}

// ProjectConfig holds project metadata.
type ProjectConfig struct {
	Name        string
	Version     string
	Description string
	Author      string
	License     string
}

// BuildConfig holds build-related settings.
type BuildConfig struct {
	Target    string // "js", "llvm", "browser"
	Entry     string // e.g., "src/main.quill"
	Output    string // e.g., "dist/"
	Minify    bool
	SourceMap bool
}

// TestConfig holds test runner settings.
type TestConfig struct {
	Pattern string // e.g., "tests/**/*.quill"
	Timeout int    // milliseconds
}

// LSPConfig holds LSP server settings.
type LSPConfig struct {
	Port int
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Project: ProjectConfig{
			Name:    "",
			Version: "0.1.0",
			License: "MIT",
		},
		Build: BuildConfig{
			Target:    "js",
			Entry:     "main.quill",
			Output:    "dist/",
			Minify:    false,
			SourceMap: true,
		},
		Dependencies:    map[string]string{},
		DevDependencies: map[string]string{},
		Test: TestConfig{
			Pattern: "tests/**/*.quill",
			Timeout: 30000,
		},
		LSP: LSPConfig{
			Port: 7998,
		},
	}
}

// LoadConfig finds and loads quill.toml from the given directory (or any parent).
func LoadConfig(dir string) (*Config, error) {
	path := findConfigFile(dir)
	if path == "" {
		return nil, fmt.Errorf("no quill.toml found in %s or parent directories", dir)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", path, err)
	}

	return ParseTOML(string(data))
}

// findConfigFile searches for quill.toml starting from dir and going up.
func findConfigFile(dir string) string {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}

	for {
		candidate := filepath.Join(absDir, "quill.toml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(absDir)
		if parent == absDir {
			break
		}
		absDir = parent
	}
	return ""
}

// ParseTOML parses a simple TOML string into a Config.
// This handles key=value pairs with [section] headers.
// Supports string, number, boolean values and simple tables.
func ParseTOML(data string) (*Config, error) {
	cfg := DefaultConfig()

	section := ""
	lines := strings.Split(data, "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}

		// Key = value
		eqIdx := strings.Index(line, "=")
		if eqIdx < 0 {
			return nil, fmt.Errorf("line %d: expected key = value, got %q", lineNum+1, line)
		}

		key := strings.TrimSpace(line[:eqIdx])
		val := strings.TrimSpace(line[eqIdx+1:])

		// Strip quotes from string values
		strVal := stripQuotes(val)
		boolVal := parseBoolValue(val)
		intVal := parseIntValue(val)

		switch section {
		case "project":
			switch key {
			case "name":
				cfg.Project.Name = strVal
			case "version":
				cfg.Project.Version = strVal
			case "description":
				cfg.Project.Description = strVal
			case "author":
				cfg.Project.Author = strVal
			case "license":
				cfg.Project.License = strVal
			}

		case "build":
			switch key {
			case "target":
				cfg.Build.Target = strVal
			case "entry":
				cfg.Build.Entry = strVal
			case "output":
				cfg.Build.Output = strVal
			case "minify":
				cfg.Build.Minify = boolVal
			case "sourcemap":
				cfg.Build.SourceMap = boolVal
			}

		case "dependencies":
			cfg.Dependencies[key] = strVal

		case "dev-dependencies":
			cfg.DevDependencies[key] = strVal

		case "test":
			switch key {
			case "pattern":
				cfg.Test.Pattern = strVal
			case "timeout":
				if intVal > 0 {
					cfg.Test.Timeout = intVal
				}
			}

		case "lsp":
			switch key {
			case "port":
				if intVal > 0 {
					cfg.LSP.Port = intVal
				}
			}
		}
	}

	return cfg, nil
}

// GenerateTOML creates a quill.toml string from a Config.
func GenerateTOML(cfg *Config) string {
	var out strings.Builder

	out.WriteString("[project]\n")
	out.WriteString(fmt.Sprintf("name = %q\n", cfg.Project.Name))
	out.WriteString(fmt.Sprintf("version = %q\n", cfg.Project.Version))
	out.WriteString(fmt.Sprintf("description = %q\n", cfg.Project.Description))
	if cfg.Project.Author != "" {
		out.WriteString(fmt.Sprintf("author = %q\n", cfg.Project.Author))
	}
	out.WriteString(fmt.Sprintf("license = %q\n", cfg.Project.License))

	out.WriteString("\n[build]\n")
	out.WriteString(fmt.Sprintf("target = %q\n", cfg.Build.Target))
	out.WriteString(fmt.Sprintf("entry = %q\n", cfg.Build.Entry))
	out.WriteString(fmt.Sprintf("output = %q\n", cfg.Build.Output))
	out.WriteString(fmt.Sprintf("minify = %s\n", boolToString(cfg.Build.Minify)))
	out.WriteString(fmt.Sprintf("sourcemap = %s\n", boolToString(cfg.Build.SourceMap)))

	out.WriteString("\n[dependencies]\n")
	for k, v := range cfg.Dependencies {
		out.WriteString(fmt.Sprintf("%s = %q\n", k, v))
	}

	out.WriteString("\n[dev-dependencies]\n")
	for k, v := range cfg.DevDependencies {
		out.WriteString(fmt.Sprintf("%s = %q\n", k, v))
	}

	out.WriteString("\n[test]\n")
	out.WriteString(fmt.Sprintf("pattern = %q\n", cfg.Test.Pattern))
	out.WriteString(fmt.Sprintf("timeout = %d\n", cfg.Test.Timeout))

	out.WriteString("\n[lsp]\n")
	out.WriteString(fmt.Sprintf("port = %d\n", cfg.LSP.Port))

	return out.String()
}

func stripQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func parseBoolValue(s string) bool {
	s = strings.TrimSpace(s)
	return s == "true"
}

func parseIntValue(s string) int {
	s = strings.TrimSpace(s)
	val := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0
		}
		val = val*10 + int(ch-'0')
	}
	return val
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
