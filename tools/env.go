package tools

import (
	"fmt"
	"os"
	"strings"
)

// EnvManager handles loading and managing environment variables for Quill apps.
type EnvManager struct {
	vars     map[string]string
	required []string
}

// NewEnvManager creates a new EnvManager.
func NewEnvManager() *EnvManager {
	return &EnvManager{
		vars: make(map[string]string),
	}
}

// LoadEnv loads environment variables from .env files based on the environment name.
// Loading order (later files override earlier):
//  1. .env , always loaded
//  2. .env.local , local overrides (gitignored)
//  3. .env.{envName} , environment-specific
//  4. .env.{envName}.local , environment + local overrides
func LoadEnv(envName string) (map[string]string, error) {
	vars := make(map[string]string)

	// Load in order of priority (last wins)
	files := []string{
		".env",
		".env.local",
	}
	if envName != "" {
		files = append(files, ".env."+envName)
		files = append(files, ".env."+envName+".local")
	}

	for _, file := range files {
		parsed, err := ParseEnvFile(file)
		if err != nil {
			// Ignore missing files
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("error loading %s: %w", file, err)
		}
		for k, v := range parsed {
			vars[k] = v
		}
	}

	return vars, nil
}

// ParseEnvFile parses a single .env file and returns its key-value pairs.
// Supports: KEY=value, KEY="quoted value", # comments, export KEY=value
func ParseEnvFile(filename string) (map[string]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseEnvString(string(data)), nil
}

// ParseEnvString parses .env file content and returns key-value pairs.
func ParseEnvString(content string) map[string]string {
	vars := make(map[string]string)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Strip optional "export " prefix
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
			line = strings.TrimSpace(line)
		}

		// Find the = separator
		eqIdx := strings.Index(line, "=")
		if eqIdx < 0 {
			continue
		}

		key := strings.TrimSpace(line[:eqIdx])
		value := strings.TrimSpace(line[eqIdx+1:])

		// Handle quoted values
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		if key != "" {
			vars[key] = value
		}
	}

	return vars
}

// ValidateRequired checks that all required keys are present.
// Returns a list of missing keys.
func ValidateRequired(vars map[string]string, required []string) []string {
	var missing []string
	for _, key := range required {
		if _, ok := vars[key]; !ok {
			missing = append(missing, key)
		}
	}
	return missing
}

// GenerateEnvInjection generates JavaScript code that injects env vars into the runtime.
func GenerateEnvInjection(vars map[string]string) string {
	var out strings.Builder
	out.WriteString("const __env = {\n")

	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	// Sort for deterministic output
	sortStrings(keys)

	for _, k := range keys {
		out.WriteString(fmt.Sprintf("  %s: %q,\n", k, vars[k]))
	}

	out.WriteString("  require(key) {\n")
	out.WriteString("    if (!(key in this)) throw new Error(`Missing required env: ${key}`);\n")
	out.WriteString("    return this[key];\n")
	out.WriteString("  },\n")
	out.WriteString("  get(key, defaultValue) {\n")
	out.WriteString("    return this[key] || defaultValue;\n")
	out.WriteString("  }\n")
	out.WriteString("};\n")
	out.WriteString("const env = __env;\n")

	return out.String()
}

// sortStrings sorts a slice of strings in place.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
