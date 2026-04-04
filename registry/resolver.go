package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Resolver handles dependency resolution for Quill packages.
type Resolver struct {
	client *RegistryClient
}

// NewResolver creates a new dependency resolver.
func NewResolver() *Resolver {
	return &Resolver{
		client: NewClient(),
	}
}

// Resolve takes a map of dependency name -> version constraint and returns
// the resolved PackageMeta for each dependency.
func (r *Resolver) Resolve(deps map[string]string) ([]PackageMeta, error) {
	var resolved []PackageMeta

	for name, constraint := range deps {
		versions, err := r.client.GetVersions(name)
		if err != nil || len(versions) == 0 {
			return nil, fmt.Errorf("package %q not found in registry", name)
		}

		matched := matchVersion(versions, constraint)
		if matched == "" {
			return nil, fmt.Errorf("no version of %q matches constraint %q", name, constraint)
		}

		meta, err := r.client.getVersionMeta(r.client.packageDir(name), matched)
		if err != nil {
			return nil, fmt.Errorf("could not read metadata for %s@%s: %w", name, matched, err)
		}

		resolved = append(resolved, *meta)
	}

	return resolved, nil
}

// Install reads the dependency map, resolves versions, downloads packages,
// and extracts them into quill_modules/ under the given project directory.
// It also creates a quill.lock file.
func (r *Resolver) Install(dir string, deps map[string]string) error {
	modulesDir := filepath.Join(dir, "quill_modules")
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		return fmt.Errorf("could not create quill_modules: %w", err)
	}

	lockEntries := make(map[string]string)

	for name, constraint := range deps {
		versions, err := r.client.GetVersions(name)
		if err != nil || len(versions) == 0 {
			fmt.Printf("  Warning: package %q not found in Quill registry, skipping\n", name)
			continue
		}

		matched := matchVersion(versions, constraint)
		if matched == "" {
			fmt.Printf("  Warning: no version of %q matches %q, skipping\n", name, constraint)
			continue
		}

		bundle, err := r.client.Download(name, matched)
		if err != nil {
			return fmt.Errorf("could not download %s@%s: %w", name, matched, err)
		}

		pkgDest := filepath.Join(modulesDir, name)
		if err := os.MkdirAll(pkgDest, 0755); err != nil {
			return err
		}

		if err := UnpackBundle(bundle, pkgDest); err != nil {
			return fmt.Errorf("could not extract %s@%s: %w", name, matched, err)
		}

		lockEntries[name] = matched
		fmt.Printf("  Installed %s@%s\n", name, matched)
	}

	// Write quill.lock
	if len(lockEntries) > 0 {
		lockData, err := json.MarshalIndent(lockEntries, "", "  ")
		if err != nil {
			return err
		}
		lockPath := filepath.Join(dir, "quill.lock")
		if err := os.WriteFile(lockPath, append(lockData, '\n'), 0644); err != nil {
			return fmt.Errorf("could not write quill.lock: %w", err)
		}
	}

	return nil
}

// matchVersion finds the best matching version from a list given a constraint.
// Supports: exact ("1.2.3"), caret ("^1.2.3"), tilde ("~1.2.3"), wildcard ("*").
func matchVersion(versions []string, constraint string) string {
	if len(versions) == 0 {
		return ""
	}

	constraint = strings.TrimSpace(constraint)

	// Wildcard: return latest
	if constraint == "*" || constraint == "latest" || constraint == "" {
		return versions[len(versions)-1]
	}

	// Caret: ^major.minor.patch — compatible with major version
	if strings.HasPrefix(constraint, "^") {
		base := strings.TrimPrefix(constraint, "^")
		return matchCaret(versions, base)
	}

	// Tilde: ~major.minor.patch — compatible with minor version
	if strings.HasPrefix(constraint, "~") {
		base := strings.TrimPrefix(constraint, "~")
		return matchTilde(versions, base)
	}

	// Exact match
	for _, v := range versions {
		if v == constraint {
			return v
		}
	}

	return ""
}

// matchCaret finds the latest version compatible with ^base.
// ^1.2.3 matches >=1.2.3 and <2.0.0
// ^0.2.3 matches >=0.2.3 and <0.3.0
func matchCaret(versions []string, base string) string {
	if !ValidateVersion(base) {
		return ""
	}

	baseParts := parseSemver(base)
	if baseParts == nil {
		return ""
	}

	var best string
	for _, v := range versions {
		vParts := parseSemver(v)
		if vParts == nil {
			continue
		}

		if baseParts[0] > 0 {
			// ^1.2.3: same major, >= minor.patch
			if vParts[0] == baseParts[0] && compareSemver(vParts, baseParts) >= 0 {
				best = v
			}
		} else if baseParts[1] > 0 {
			// ^0.2.3: same major.minor, >= patch
			if vParts[0] == 0 && vParts[1] == baseParts[1] && compareSemver(vParts, baseParts) >= 0 {
				best = v
			}
		} else {
			// ^0.0.3: exact match only
			if v == base {
				best = v
			}
		}
	}

	return best
}

// matchTilde finds the latest version compatible with ~base.
// ~1.2.3 matches >=1.2.3 and <1.3.0
func matchTilde(versions []string, base string) string {
	if !ValidateVersion(base) {
		return ""
	}

	baseParts := parseSemver(base)
	if baseParts == nil {
		return ""
	}

	var best string
	for _, v := range versions {
		vParts := parseSemver(v)
		if vParts == nil {
			continue
		}

		if vParts[0] == baseParts[0] && vParts[1] == baseParts[1] && compareSemver(vParts, baseParts) >= 0 {
			best = v
		}
	}

	return best
}

// parseSemver splits "1.2.3" into [1, 2, 3].
func parseSemver(v string) []int {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return nil
	}
	result := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		result[i] = n
	}
	return result
}

// compareSemver returns -1, 0, or 1 comparing a to b.
func compareSemver(a, b []int) int {
	for i := 0; i < 3; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}
