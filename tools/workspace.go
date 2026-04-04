package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Package represents a single package in a workspace.
type Package struct {
	Name         string
	Path         string            // absolute path to the package directory
	Version      string
	Dependencies map[string]string // package name -> version
	Entry        string            // main .quill file
}

// WorkspaceScan finds all packages in a workspace based on the workspace config.
// It reads the root quill.toml for [workspace] members glob patterns.
func WorkspaceScan(root string) ([]Package, error) {
	configPath := filepath.Join(root, "quill.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", configPath, err)
	}

	members := parseWorkspaceMembers(string(data))
	if len(members) == 0 {
		return nil, fmt.Errorf("no workspace members defined in quill.toml")
	}

	var packages []Package
	for _, pattern := range members {
		fullPattern := filepath.Join(root, pattern)
		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			continue
		}
		for _, m := range matches {
			info, err := os.Stat(m)
			if err != nil || !info.IsDir() {
				continue
			}
			pkg, err := readPackage(m)
			if err != nil {
				continue
			}
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// parseWorkspaceMembers extracts member patterns from a TOML string.
func parseWorkspaceMembers(data string) []string {
	var members []string
	inWorkspace := false
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[workspace]" {
			inWorkspace = true
			continue
		}
		if strings.HasPrefix(trimmed, "[") && trimmed != "[workspace]" {
			inWorkspace = false
			continue
		}
		if inWorkspace && strings.HasPrefix(trimmed, "members") {
			// Parse members = ["packages/*", "apps/*"]
			eqIdx := strings.Index(trimmed, "=")
			if eqIdx < 0 {
				continue
			}
			val := strings.TrimSpace(trimmed[eqIdx+1:])
			val = strings.Trim(val, "[]")
			parts := strings.Split(val, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				p = strings.Trim(p, "\"' ")
				if p != "" {
					members = append(members, p)
				}
			}
		}
	}

	return members
}

// readPackage reads a package from a directory with a quill.toml file.
func readPackage(dir string) (Package, error) {
	configPath := filepath.Join(dir, "quill.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Package{}, fmt.Errorf("no quill.toml found in %s", dir)
	}

	pkg := Package{
		Path:         dir,
		Dependencies: make(map[string]string),
		Entry:        "main.quill",
	}

	section := ""
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section = strings.TrimSpace(trimmed[1 : len(trimmed)-1])
			continue
		}
		eqIdx := strings.Index(trimmed, "=")
		if eqIdx < 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:eqIdx])
		val := strings.TrimSpace(trimmed[eqIdx+1:])
		val = strings.Trim(val, "\"")

		switch section {
		case "project":
			switch key {
			case "name":
				pkg.Name = val
			case "version":
				pkg.Version = val
			}
		case "build":
			if key == "entry" {
				pkg.Entry = val
			}
		case "dependencies":
			pkg.Dependencies[key] = val
		}
	}

	if pkg.Name == "" {
		pkg.Name = filepath.Base(dir)
	}

	return pkg, nil
}

// ResolveBuildOrder sorts packages in dependency order using topological sort.
// Packages with no dependencies on other workspace packages come first.
func ResolveBuildOrder(packages []Package) []Package {
	// Build name -> package map
	nameMap := make(map[string]int)
	for i, p := range packages {
		nameMap[p.Name] = i
	}

	// Build adjacency list (dependency graph)
	adjList := make(map[int][]int)
	inDegree := make(map[int]int)
	for i := range packages {
		inDegree[i] = 0
	}

	for i, pkg := range packages {
		for dep := range pkg.Dependencies {
			if j, ok := nameMap[dep]; ok {
				adjList[j] = append(adjList[j], i)
				inDegree[i]++
			}
		}
	}

	// Topological sort (Kahn's algorithm)
	var queue []int
	for i := range packages {
		if inDegree[i] == 0 {
			queue = append(queue, i)
		}
	}

	var ordered []Package
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		ordered = append(ordered, packages[curr])

		for _, neighbor := range adjList[curr] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// If there's a cycle, append remaining packages
	if len(ordered) < len(packages) {
		for i, pkg := range packages {
			found := false
			for _, o := range ordered {
				if o.Name == packages[i].Name {
					found = true
					break
				}
			}
			if !found {
				ordered = append(ordered, pkg)
			}
		}
	}

	return ordered
}

// BuildWorkspace compiles all packages in dependency order.
// The buildFn callback is called for each package with its entry file path.
func BuildWorkspace(packages []Package, buildFn func(entryFile string) error) error {
	ordered := ResolveBuildOrder(packages)
	for _, pkg := range ordered {
		entry := filepath.Join(pkg.Path, pkg.Entry)
		fmt.Printf("Building %s (%s)...\n", pkg.Name, entry)
		if err := buildFn(entry); err != nil {
			return fmt.Errorf("failed to build %s: %w", pkg.Name, err)
		}
	}
	return nil
}

// TestWorkspace runs tests for all packages in the workspace.
// The testFn callback is called for each package with its directory path.
func TestWorkspace(packages []Package, testFn func(pkgDir string) error) error {
	for _, pkg := range packages {
		fmt.Printf("Testing %s...\n", pkg.Name)
		if err := testFn(pkg.Path); err != nil {
			return fmt.Errorf("tests failed for %s: %w", pkg.Name, err)
		}
	}
	return nil
}
