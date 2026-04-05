package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanMigrations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "quill-migrations-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test migration files
	os.WriteFile(filepath.Join(tmpDir, "20260101_120000_create_users.up.sql"), []byte("CREATE TABLE users (id INTEGER);"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "20260101_120000_create_users.down.sql"), []byte("DROP TABLE users;"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "20260102_120000_add_posts.up.sql"), []byte("CREATE TABLE posts (id INTEGER);"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "20260102_120000_add_posts.down.sql"), []byte("DROP TABLE posts;"), 0644)

	migrations, err := ScanMigrations(tmpDir)
	if err != nil {
		t.Fatalf("ScanMigrations failed: %s", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("expected 2 migrations, got %d", len(migrations))
	}

	if migrations[0].Name != "create_users" {
		t.Errorf("expected first migration name 'create_users', got %s", migrations[0].Name)
	}
	if migrations[1].Name != "add_posts" {
		t.Errorf("expected second migration name 'add_posts', got %s", migrations[1].Name)
	}
}

func TestMigrationOrdering(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "quill-migrations-order-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create migrations out of order
	os.WriteFile(filepath.Join(tmpDir, "20260301_120000_third.up.sql"), []byte("-- third"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "20260101_120000_first.up.sql"), []byte("-- first"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "20260201_120000_second.up.sql"), []byte("-- second"), 0644)

	migrations, err := ScanMigrations(tmpDir)
	if err != nil {
		t.Fatalf("ScanMigrations failed: %s", err)
	}

	if len(migrations) != 3 {
		t.Fatalf("expected 3 migrations, got %d", len(migrations))
	}

	if migrations[0].Version != "20260101_120000" {
		t.Errorf("expected first migration version '20260101_120000', got %s", migrations[0].Version)
	}
	if migrations[1].Version != "20260201_120000" {
		t.Errorf("expected second migration version '20260201_120000', got %s", migrations[1].Version)
	}
	if migrations[2].Version != "20260301_120000" {
		t.Errorf("expected third migration version '20260301_120000', got %s", migrations[2].Version)
	}
}

func TestGenerateMigration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "quill-migrations-gen-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	upFile, downFile, err := GenerateMigration(tmpDir, "create users")
	if err != nil {
		t.Fatalf("GenerateMigration failed: %s", err)
	}

	if !strings.HasSuffix(upFile, "_create_users.up.sql") {
		t.Errorf("up file has unexpected name: %s", upFile)
	}
	if !strings.HasSuffix(downFile, "_create_users.down.sql") {
		t.Errorf("down file has unexpected name: %s", downFile)
	}

	// Check files exist
	if _, err := os.Stat(upFile); os.IsNotExist(err) {
		t.Error("up migration file was not created")
	}
	if _, err := os.Stat(downFile); os.IsNotExist(err) {
		t.Error("down migration file was not created")
	}
}

func TestShowStatus(t *testing.T) {
	migrations := []DBMigration{
		{Version: "20260101_120000", Name: "create_users"},
		{Version: "20260102_120000", Name: "add_posts"},
		{Version: "20260103_120000", Name: "add_comments"},
	}
	applied := []string{"20260101_120000", "20260102_120000"}

	status := ShowStatus(migrations, applied)

	if !strings.Contains(status, "create_users") {
		t.Error("status missing create_users")
	}
	if !strings.Contains(status, "applied") {
		t.Error("status missing 'applied' label")
	}
	if !strings.Contains(status, "pending") {
		t.Error("status missing 'pending' label")
	}
}

func TestShowStatusEmpty(t *testing.T) {
	status := ShowStatus(nil, nil)
	if status != "No migrations found." {
		t.Errorf("expected 'No migrations found.', got %q", status)
	}
}

func TestGenerateFromModels(t *testing.T) {
	models := []ModelDef{
		{
			Name: "User",
			Fields: []ModelDefField{
				{Name: "email", Type: "string"},
				{Name: "age", Type: "integer"},
			},
		},
	}

	sql := GenerateFromModels(models)

	if !strings.Contains(sql, "CREATE TABLE IF NOT EXISTS users") {
		t.Error("missing CREATE TABLE statement")
	}
	if !strings.Contains(sql, "email TEXT") {
		t.Error("missing email field")
	}
	if !strings.Contains(sql, "age INTEGER") {
		t.Error("missing age field")
	}
	if !strings.Contains(sql, "id INTEGER PRIMARY KEY AUTOINCREMENT") {
		t.Error("missing primary key")
	}
}

func TestScanMigrationsEmptyDir(t *testing.T) {
	migrations, err := ScanMigrations("/nonexistent/path")
	if err != nil {
		t.Fatalf("expected nil error for nonexistent dir, got: %s", err)
	}
	if len(migrations) != 0 {
		t.Errorf("expected 0 migrations, got %d", len(migrations))
	}
}
