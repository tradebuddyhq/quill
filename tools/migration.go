package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// MigrationRule defines a single source code transformation.
type MigrationRule struct {
	Description string
	Pattern     string // regex to match
	Replace     string // replacement string (supports $1, $2 etc.)
}

// Migration defines a set of rules to apply when migrating between versions.
type Migration struct {
	FromVersion string
	ToVersion   string
	Rules       []MigrationRule
}

// MigrationResult tracks what changes were made to a file.
type MigrationResult struct {
	File     string
	Changes  int
	Original string
	Migrated string
}

// BuiltinMigrations returns the set of built-in migration rules for Quill.
func BuiltinMigrations() []Migration {
	return []Migration{
		{
			FromVersion: "v0.1",
			ToVersion:   "v0.2",
			Rules: []MigrationRule{
				{
					Description: "Rename 'print' to 'say'",
					Pattern:     `\bprint\b`,
					Replace:     "say",
				},
				{
					Description: "Rename 'func' to 'to'",
					Pattern:     `\bfunc\b`,
					Replace:     "to",
				},
				{
					Description: "Rename 'var' to 'set' (assignment keyword)",
					Pattern:     `\bvar\b`,
					Replace:     "set",
				},
			},
		},
		{
			FromVersion: "v0.2",
			ToVersion:   "v0.3",
			Rules: []MigrationRule{
				{
					Description: "Update match syntax: 'case X:' to 'when X:'",
					Pattern:     `\bcase\s+`,
					Replace:     "when ",
				},
				{
					Description: "Update define syntax: 'enum' to 'define'",
					Pattern:     `\benum\b`,
					Replace:     "define",
				},
			},
		},
	}
}

// GetMigrationChain returns the ordered list of migrations needed to go
// from fromVersion to toVersion.
func GetMigrationChain(from, to string, migrations []Migration) []Migration {
	// Build version graph
	type edge struct {
		migration Migration
	}
	graph := make(map[string][]edge)
	for _, m := range migrations {
		graph[m.FromVersion] = append(graph[m.FromVersion], edge{migration: m})
	}

	// BFS to find path from -> to
	type pathNode struct {
		version    string
		migrations []Migration
	}

	visited := make(map[string]bool)
	queue := []pathNode{{version: from, migrations: nil}}
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.version == to {
			return current.migrations
		}

		for _, e := range graph[current.version] {
			if !visited[e.migration.ToVersion] {
				visited[e.migration.ToVersion] = true
				newPath := make([]Migration, len(current.migrations)+1)
				copy(newPath, current.migrations)
				newPath[len(current.migrations)] = e.migration
				queue = append(queue, pathNode{version: e.migration.ToVersion, migrations: newPath})
			}
		}
	}

	return nil
}

// ApplyMigration applies a single migration's rules to source code.
// Returns the migrated code and the number of changes made.
func ApplyMigration(source string, migration Migration) (string, int) {
	result := source
	totalChanges := 0

	for _, rule := range migration.Rules {
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			continue
		}

		matches := re.FindAllStringIndex(result, -1)
		totalChanges += len(matches)
		result = re.ReplaceAllString(result, rule.Replace)
	}

	return result, totalChanges
}

// MigrateFile applies a chain of migrations to a single file.
func MigrateFile(filePath string, migrations []Migration, dryRun bool) (*MigrationResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", filePath, err)
	}

	original := string(data)
	current := original
	totalChanges := 0

	for _, m := range migrations {
		migrated, changes := ApplyMigration(current, m)
		totalChanges += changes
		current = migrated
	}

	result := &MigrationResult{
		File:     filePath,
		Changes:  totalChanges,
		Original: original,
		Migrated: current,
	}

	if !dryRun && totalChanges > 0 {
		// Backup original file
		backupPath := filePath + ".bak"
		if err := os.WriteFile(backupPath, data, 0644); err != nil {
			return result, fmt.Errorf("could not create backup %s: %w", backupPath, err)
		}

		// Write migrated file
		if err := os.WriteFile(filePath, []byte(current), 0644); err != nil {
			return result, fmt.Errorf("could not write %s: %w", filePath, err)
		}
	}

	return result, nil
}

// MigrateDirectory applies migrations to all .quill files in a directory.
func MigrateDirectory(dir string, from, to string, dryRun bool) ([]MigrationResult, error) {
	migrations := GetMigrationChain(from, to, BuiltinMigrations())
	if migrations == nil {
		return nil, fmt.Errorf("no migration path from %s to %s", from, to)
	}

	var results []MigrationResult

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".quill") {
			return nil
		}

		result, err := MigrateFile(path, migrations, dryRun)
		if err != nil {
			return err
		}
		results = append(results, *result)
		return nil
	})

	sort.Slice(results, func(i, j int) bool {
		return results[i].File < results[j].File
	})

	return results, err
}

// FormatMigrationReport generates a human-readable report of migration results.
func FormatMigrationReport(results []MigrationResult, dryRun bool) string {
	var out strings.Builder

	if dryRun {
		out.WriteString("Migration Preview (dry run):\n")
	} else {
		out.WriteString("Migration Report:\n")
	}
	out.WriteString(strings.Repeat("\u2500", 50) + "\n")

	totalChanges := 0
	for _, r := range results {
		status := "unchanged"
		if r.Changes > 0 {
			if dryRun {
				status = fmt.Sprintf("%d change(s) would be made", r.Changes)
			} else {
				status = fmt.Sprintf("%d change(s) applied", r.Changes)
			}
		}
		out.WriteString(fmt.Sprintf("  %s: %s\n", r.File, status))
		totalChanges += r.Changes
	}

	out.WriteString(strings.Repeat("\u2500", 50) + "\n")
	out.WriteString(fmt.Sprintf("Total: %d file(s), %d change(s)\n", len(results), totalChanges))

	return out.String()
}
