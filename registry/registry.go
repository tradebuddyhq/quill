package registry

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// PackageMeta holds all metadata about a Quill package.
type PackageMeta struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Author       string            `json:"author"`
	License      string            `json:"license"`
	Main         string            `json:"main"`
	Dependencies map[string]string `json:"dependencies"`
	Repository   string            `json:"repository"`
	Keywords     []string          `json:"keywords"`
}

var semverRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// ReadPackageMeta reads and parses the quill.json file in the given directory.
func ReadPackageMeta(dir string) (*PackageMeta, error) {
	path := filepath.Join(dir, "quill.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read quill.json: %w", err)
	}

	var meta PackageMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("could not parse quill.json: %w", err)
	}

	if meta.Dependencies == nil {
		meta.Dependencies = make(map[string]string)
	}

	return &meta, nil
}

// WritePackageMeta writes a PackageMeta struct to quill.json in the given directory.
func WritePackageMeta(dir string, meta *PackageMeta) error {
	if meta.Dependencies == nil {
		meta.Dependencies = make(map[string]string)
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal package meta: %w", err)
	}

	path := filepath.Join(dir, "quill.json")
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("could not write quill.json: %w", err)
	}

	return nil
}

// ValidateVersion checks if a version string is valid semver (major.minor.patch).
func ValidateVersion(v string) bool {
	return semverRegex.MatchString(v)
}

// BumpVersion bumps a semver version string by the given component (major, minor, or patch).
func BumpVersion(v string, bump string) string {
	if !ValidateVersion(v) {
		return v
	}

	parts := strings.Split(v, ".")
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	switch bump {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return v
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

// PackageBundle creates a tar.gz archive of the package source files in the given directory.
// It includes .quill files and quill.json, excluding node_modules and quill_modules.
func PackageBundle(dir string) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Skip hidden dirs, node_modules, quill_modules
		if info.IsDir() {
			base := filepath.Base(relPath)
			if base == "node_modules" || base == "quill_modules" || (strings.HasPrefix(base, ".") && base != ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Only include relevant files
		ext := filepath.Ext(relPath)
		base := filepath.Base(relPath)
		if ext != ".quill" && base != "quill.json" {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not bundle package: %w", err)
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnpackBundle extracts a tar.gz archive into the destination directory.
func UnpackBundle(data []byte, destDir string) error {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("could not read gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("could not read tar: %w", err)
		}

		target := filepath.Join(destDir, header.Name)

		// Prevent path traversal
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return fmt.Errorf("invalid file path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	return nil
}
