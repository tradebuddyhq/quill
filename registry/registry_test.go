package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"1.0.0", true},
		{"0.0.1", true},
		{"10.20.30", true},
		{"1.0", false},
		{"1", false},
		{"abc", false},
		{"1.0.0-beta", false},
		{"", false},
	}
	for _, tt := range tests {
		got := ValidateVersion(tt.input)
		if got != tt.want {
			t.Errorf("ValidateVersion(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestBumpVersion(t *testing.T) {
	tests := []struct {
		version string
		bump    string
		want    string
	}{
		{"1.0.0", "patch", "1.0.1"},
		{"1.0.0", "minor", "1.1.0"},
		{"1.0.0", "major", "2.0.0"},
		{"0.9.9", "patch", "0.9.10"},
		{"0.9.9", "minor", "0.10.0"},
		{"0.9.9", "major", "1.0.0"},
		{"invalid", "patch", "invalid"},
		{"1.0.0", "unknown", "1.0.0"},
	}
	for _, tt := range tests {
		got := BumpVersion(tt.version, tt.bump)
		if got != tt.want {
			t.Errorf("BumpVersion(%q, %q) = %q, want %q", tt.version, tt.bump, got, tt.want)
		}
	}
}

func TestReadWritePackageMeta(t *testing.T) {
	dir := t.TempDir()

	meta := &PackageMeta{
		Name:        "test-pkg",
		Version:     "1.0.0",
		Description: "A test package",
		Author:      "test",
		Main:        "main.quill",
	}

	if err := WritePackageMeta(dir, meta); err != nil {
		t.Fatalf("WritePackageMeta failed: %v", err)
	}

	got, err := ReadPackageMeta(dir)
	if err != nil {
		t.Fatalf("ReadPackageMeta failed: %v", err)
	}

	if got.Name != "test-pkg" {
		t.Errorf("Name = %q, want %q", got.Name, "test-pkg")
	}
	if got.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", got.Version, "1.0.0")
	}
	if got.Dependencies == nil {
		t.Error("Dependencies should not be nil")
	}
}

func TestReadPackageMetaMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := ReadPackageMeta(dir)
	if err == nil {
		t.Error("expected error for missing quill.json")
	}
}

func TestPackageBundleAndUnpack(t *testing.T) {
	src := t.TempDir()

	// Create a quill.json and a .quill file
	meta := &PackageMeta{Name: "bundle-test", Version: "0.1.0"}
	WritePackageMeta(src, meta)
	os.WriteFile(filepath.Join(src, "main.quill"), []byte("say \"hello\""), 0644)
	os.WriteFile(filepath.Join(src, "readme.txt"), []byte("ignored"), 0644)

	data, err := PackageBundle(src)
	if err != nil {
		t.Fatalf("PackageBundle failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty bundle")
	}

	// Unpack and verify
	dest := t.TempDir()
	if err := UnpackBundle(data, dest); err != nil {
		t.Fatalf("UnpackBundle failed: %v", err)
	}

	// quill.json should exist
	if _, err := os.Stat(filepath.Join(dest, "quill.json")); err != nil {
		t.Error("quill.json should exist after unpack")
	}
	// main.quill should exist
	if _, err := os.Stat(filepath.Join(dest, "main.quill")); err != nil {
		t.Error("main.quill should exist after unpack")
	}
	// readme.txt should NOT exist (not a .quill or quill.json)
	if _, err := os.Stat(filepath.Join(dest, "readme.txt")); err == nil {
		t.Error("readme.txt should NOT be in bundle")
	}
}
