package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// DBMigration represents a single database migration with up and down SQL.
type DBMigration struct {
	Version   string
	Name      string
	UpSQL     string
	DownSQL   string
	Timestamp time.Time
}

// MigrationManager handles database migration operations.
type MigrationManager struct {
	Dir string // migrations directory, default "migrations/"
}

// ModelDef represents a model definition extracted from .quill files.
type ModelDef struct {
	Name   string
	Fields []ModelDefField
}

// ModelDefField represents a field in a model definition.
type ModelDefField struct {
	Name string
	Type string
}

// NewMigrationManager creates a new MigrationManager with the given directory.
func NewMigrationManager(dir string) *MigrationManager {
	if dir == "" {
		dir = "migrations"
	}
	return &MigrationManager{Dir: dir}
}

// ScanMigrations reads migration files from the migrations directory.
func (m *MigrationManager) ScanMigrations() ([]DBMigration, error) {
	return ScanMigrations(m.Dir)
}

// ScanMigrations reads migration files from the given directory.
func ScanMigrations(dir string) ([]DBMigration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("could not read migrations directory: %w", err)
	}

	// Match up.sql files and find their corresponding down.sql
	upPattern := regexp.MustCompile(`^(\d{8}_\d{6})_(.+)\.up\.sql$`)

	var migrations []DBMigration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := upPattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}
		version := matches[1]
		name := matches[2]

		upPath := filepath.Join(dir, entry.Name())
		downPath := filepath.Join(dir, fmt.Sprintf("%s_%s.down.sql", version, name))

		upSQL, err := os.ReadFile(upPath)
		if err != nil {
			return nil, fmt.Errorf("could not read %s: %w", upPath, err)
		}

		var downSQL []byte
		if data, err := os.ReadFile(downPath); err == nil {
			downSQL = data
		}

		ts, _ := time.Parse("20060102_150405", version)

		migrations = append(migrations, DBMigration{
			Version:   version,
			Name:      name,
			UpSQL:     string(upSQL),
			DownSQL:   string(downSQL),
			Timestamp: ts,
		})
	}

	// Sort by version (timestamp)
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// GetApplied returns the list of applied migration versions by querying the tracking table.
func GetApplied(queryFunc func(sql string) ([]string, error)) ([]string, error) {
	return queryFunc("SELECT version FROM __quill_migrations ORDER BY version")
}

var safeMigrationName = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)

func sanitizeMigrationField(s string) string {
	if !safeMigrationName.MatchString(s) {
		// Strip anything that isn't alphanumeric, underscore, or hyphen
		clean := regexp.MustCompile(`[^a-zA-Z0-9_\-]`).ReplaceAllString(s, "")
		return clean
	}
	return s
}

// ApplyDBMigration runs the up SQL and records it in the tracking table.
func ApplyDBMigration(migration DBMigration, execFunc func(sql string) error) error {
	if err := execFunc(migration.UpSQL); err != nil {
		return fmt.Errorf("failed to apply migration %s_%s: %w", migration.Version, migration.Name, err)
	}
	version := sanitizeMigrationField(migration.Version)
	name := sanitizeMigrationField(migration.Name)
	trackSQL := fmt.Sprintf("INSERT INTO __quill_migrations (version, name, applied_at) VALUES ('%s', '%s', datetime('now'))",
		version, name)
	if err := execFunc(trackSQL); err != nil {
		return fmt.Errorf("failed to record migration %s_%s: %w", migration.Version, migration.Name, err)
	}
	return nil
}

// RollbackMigration runs the down SQL and removes the tracking record.
func RollbackMigration(migration DBMigration, execFunc func(sql string) error) error {
	if migration.DownSQL == "" {
		return fmt.Errorf("no down migration for %s_%s", migration.Version, migration.Name)
	}
	if err := execFunc(migration.DownSQL); err != nil {
		return fmt.Errorf("failed to rollback migration %s_%s: %w", migration.Version, migration.Name, err)
	}
	version := sanitizeMigrationField(migration.Version)
	trackSQL := fmt.Sprintf("DELETE FROM __quill_migrations WHERE version = '%s'", version)
	if err := execFunc(trackSQL); err != nil {
		return fmt.Errorf("failed to remove migration record %s_%s: %w", migration.Version, migration.Name, err)
	}
	return nil
}

// GenerateMigration creates a new migration file pair and returns the up and down file paths.
func GenerateMigration(dir string, name string) (string, string, error) {
	if dir == "" {
		dir = "migrations"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", fmt.Errorf("could not create migrations directory: %w", err)
	}

	ts := time.Now().Format("20060102_150405")
	safeName := strings.ReplaceAll(strings.ToLower(name), " ", "_")

	upFile := filepath.Join(dir, fmt.Sprintf("%s_%s.up.sql", ts, safeName))
	downFile := filepath.Join(dir, fmt.Sprintf("%s_%s.down.sql", ts, safeName))

	upContent := fmt.Sprintf("-- Migration: %s (up)\n-- Created: %s\n\n", name, time.Now().Format(time.RFC3339))
	downContent := fmt.Sprintf("-- Migration: %s (down)\n-- Created: %s\n\n", name, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		return "", "", fmt.Errorf("could not write up migration: %w", err)
	}
	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		return "", "", fmt.Errorf("could not write down migration: %w", err)
	}

	return upFile, downFile, nil
}

// ShowStatus formats a status table showing applied and pending migrations.
func ShowStatus(migrations []DBMigration, applied []string) string {
	if len(migrations) == 0 {
		return "No migrations found."
	}

	appliedSet := make(map[string]bool)
	for _, v := range applied {
		appliedSet[v] = true
	}

	var out strings.Builder
	out.WriteString(fmt.Sprintf("%-20s %-30s %s\n", "VERSION", "NAME", "STATUS"))
	out.WriteString(strings.Repeat("-", 60) + "\n")

	for _, m := range migrations {
		status := "pending"
		if appliedSet[m.Version] {
			status = "applied"
		}
		out.WriteString(fmt.Sprintf("%-20s %-30s %s\n", m.Version, m.Name, status))
	}

	return out.String()
}

// GenerateFromModels auto-generates CREATE TABLE SQL from model definitions.
func GenerateFromModels(models []ModelDef) string {
	var out strings.Builder

	for i, model := range models {
		tableName := strings.ToLower(model.Name) + "s"
		out.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", tableName))
		out.WriteString("  id INTEGER PRIMARY KEY AUTOINCREMENT")

		for _, field := range model.Fields {
			sqlType := quillTypeToSQL(field.Type)
			out.WriteString(fmt.Sprintf(",\n  %s %s", field.Name, sqlType))
		}

		out.WriteString("\n);\n")
		if i < len(models)-1 {
			out.WriteString("\n")
		}
	}

	return out.String()
}

func quillTypeToSQL(t string) string {
	switch strings.ToLower(t) {
	case "string", "text":
		return "TEXT"
	case "number", "float":
		return "REAL"
	case "integer", "int":
		return "INTEGER"
	case "boolean", "bool":
		return "INTEGER"
	case "date", "datetime":
		return "TEXT"
	case "json":
		return "TEXT"
	default:
		return "TEXT"
	}
}
