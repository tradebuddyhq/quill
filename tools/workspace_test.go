package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseWorkspaceMembers(t *testing.T) {
	toml := `[project]
name = "myapp"

[workspace]
members = ["packages/*", "apps/*"]

[build]
target = "js"
`
	members := parseWorkspaceMembers(toml)

	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
	if members[0] != "packages/*" {
		t.Errorf("expected packages/*, got %s", members[0])
	}
	if members[1] != "apps/*" {
		t.Errorf("expected apps/*, got %s", members[1])
	}
}

func TestResolveBuildOrder(t *testing.T) {
	packages := []Package{
		{Name: "app", Dependencies: map[string]string{"core": "0.1.0", "utils": "0.1.0"}},
		{Name: "core", Dependencies: map[string]string{}},
		{Name: "utils", Dependencies: map[string]string{"core": "0.1.0"}},
	}

	ordered := ResolveBuildOrder(packages)

	if len(ordered) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(ordered))
	}

	// core should come before utils and app
	coreIdx := -1
	utilsIdx := -1
	appIdx := -1
	for i, p := range ordered {
		switch p.Name {
		case "core":
			coreIdx = i
		case "utils":
			utilsIdx = i
		case "app":
			appIdx = i
		}
	}

	if coreIdx > utilsIdx {
		t.Error("core should be built before utils")
	}
	if coreIdx > appIdx {
		t.Error("core should be built before app")
	}
	if utilsIdx > appIdx {
		t.Error("utils should be built before app")
	}
}

func TestWorkspaceScan(t *testing.T) {
	// Create temp workspace
	tmpDir, err := os.MkdirTemp("", "quill-workspace-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create root config
	rootConfig := `[workspace]
members = ["packages/*"]
`
	os.WriteFile(filepath.Join(tmpDir, "quill.toml"), []byte(rootConfig), 0644)

	// Create package directories
	pkgDir := filepath.Join(tmpDir, "packages", "mylib")
	os.MkdirAll(pkgDir, 0755)

	pkgConfig := `[project]
name = "mylib"
version = "1.0.0"
`
	os.WriteFile(filepath.Join(pkgDir, "quill.toml"), []byte(pkgConfig), 0644)

	packages, err := WorkspaceScan(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}
	if packages[0].Name != "mylib" {
		t.Errorf("expected mylib, got %s", packages[0].Name)
	}
}

func TestResolveBuildOrderNoDeps(t *testing.T) {
	packages := []Package{
		{Name: "a", Dependencies: map[string]string{}},
		{Name: "b", Dependencies: map[string]string{}},
		{Name: "c", Dependencies: map[string]string{}},
	}

	ordered := ResolveBuildOrder(packages)

	if len(ordered) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(ordered))
	}
}
