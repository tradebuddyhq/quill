package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyMigrationRules(t *testing.T) {
	migration := Migration{
		FromVersion: "v0.1",
		ToVersion:   "v0.2",
		Rules: []MigrationRule{
			{Description: "rename print to say", Pattern: `\bprint\b`, Replace: "say"},
			{Description: "rename func to to", Pattern: `\bfunc\b`, Replace: "to"},
		},
	}

	source := `func greet name:
    print "Hello {name}"

print "world"`

	result, changes := ApplyMigration(source, migration)

	if changes != 3 {
		t.Errorf("expected 3 changes, got %d", changes)
	}
	if strings.Contains(result, "print") {
		t.Error("should have replaced all 'print' with 'say'")
	}
	if strings.Contains(result, "func") {
		t.Error("should have replaced all 'func' with 'to'")
	}
	if !strings.Contains(result, "say") {
		t.Error("should contain 'say'")
	}
	if !strings.Contains(result, "to") {
		t.Error("should contain 'to'")
	}
}

func TestGetMigrationChain(t *testing.T) {
	migrations := BuiltinMigrations()

	chain := GetMigrationChain("v0.1", "v0.3", migrations)
	if chain == nil {
		t.Fatal("expected a migration chain from v0.1 to v0.3")
	}
	if len(chain) != 2 {
		t.Fatalf("expected 2 migrations in chain, got %d", len(chain))
	}
	if chain[0].FromVersion != "v0.1" || chain[0].ToVersion != "v0.2" {
		t.Error("first migration should be v0.1 -> v0.2")
	}
	if chain[1].FromVersion != "v0.2" || chain[1].ToVersion != "v0.3" {
		t.Error("second migration should be v0.2 -> v0.3")
	}
}

func TestGetMigrationChainNotFound(t *testing.T) {
	migrations := BuiltinMigrations()

	chain := GetMigrationChain("v0.1", "v9.9", migrations)
	if chain != nil {
		t.Error("should return nil when no migration path exists")
	}
}

func TestDryRunMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "quill-migration-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	source := "print \"hello\"\n"
	filePath := filepath.Join(tmpDir, "test.quill")
	os.WriteFile(filePath, []byte(source), 0644)

	migrations := []Migration{
		{
			FromVersion: "v0.1",
			ToVersion:   "v0.2",
			Rules: []MigrationRule{
				{Pattern: `\bprint\b`, Replace: "say"},
			},
		},
	}

	result, err := MigrateFile(filePath, migrations, true)
	if err != nil {
		t.Fatal(err)
	}

	if result.Changes != 1 {
		t.Errorf("expected 1 change, got %d", result.Changes)
	}

	// File should not be modified in dry run
	data, _ := os.ReadFile(filePath)
	if string(data) != source {
		t.Error("dry run should not modify the file")
	}
}

func TestMigrateFileCreatesBackup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "quill-migration-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	source := "print \"hello\"\n"
	filePath := filepath.Join(tmpDir, "test.quill")
	os.WriteFile(filePath, []byte(source), 0644)

	migrations := []Migration{
		{
			FromVersion: "v0.1",
			ToVersion:   "v0.2",
			Rules: []MigrationRule{
				{Pattern: `\bprint\b`, Replace: "say"},
			},
		},
	}

	_, err = MigrateFile(filePath, migrations, false)
	if err != nil {
		t.Fatal(err)
	}

	// Check backup exists
	backupPath := filePath + ".bak"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("should create a backup file")
	}

	// Check file was modified
	data, _ := os.ReadFile(filePath)
	if !strings.Contains(string(data), "say") {
		t.Error("file should be migrated")
	}
}

func TestFormatMigrationReport(t *testing.T) {
	results := []MigrationResult{
		{File: "main.quill", Changes: 3},
		{File: "utils.quill", Changes: 0},
	}

	report := FormatMigrationReport(results, false)

	if !strings.Contains(report, "Migration Report") {
		t.Error("should contain header")
	}
	if !strings.Contains(report, "main.quill") {
		t.Error("should list main.quill")
	}
	if !strings.Contains(report, "3 change(s) applied") {
		t.Error("should show changes applied")
	}
	if !strings.Contains(report, "unchanged") {
		t.Error("should show unchanged for utils.quill")
	}
}

func TestFormatMigrationReportDryRun(t *testing.T) {
	results := []MigrationResult{
		{File: "main.quill", Changes: 2},
	}

	report := FormatMigrationReport(results, true)

	if !strings.Contains(report, "dry run") {
		t.Error("should indicate dry run")
	}
	if !strings.Contains(report, "would be made") {
		t.Error("should use 'would be made' in dry run")
	}
}
