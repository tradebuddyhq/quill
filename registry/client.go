package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RegistryClient talks to the Quill package registry.
// Since no real server exists yet, it uses a local file-based fallback
// at ~/.quill/registry/.
type RegistryClient struct {
	BaseURL  string
	localDir string
}

// NewClient creates a new RegistryClient with defaults.
func NewClient() *RegistryClient {
	home, _ := os.UserHomeDir()
	return &RegistryClient{
		BaseURL:  "https://registry.quill.tradebuddy.dev",
		localDir: filepath.Join(home, ".quill", "registry"),
	}
}

// ensureLocalDir creates the local registry directory if it does not exist.
func (c *RegistryClient) ensureLocalDir() error {
	return os.MkdirAll(c.localDir, 0755)
}

// packageDir returns the local directory for a specific package.
func (c *RegistryClient) packageDir(name string) string {
	return filepath.Join(c.localDir, name)
}

// versionDir returns the local directory for a specific package version.
func (c *RegistryClient) versionDir(name, version string) string {
	return filepath.Join(c.localDir, name, version)
}

// Search searches the local registry for packages matching the query string.
func (c *RegistryClient) Search(query string) ([]PackageMeta, error) {
	if err := c.ensureLocalDir(); err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []PackageMeta

	entries, err := os.ReadDir(c.localDir)
	if err != nil {
		return results, nil // empty registry
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pkgName := entry.Name()

		// Check name match
		if !strings.Contains(strings.ToLower(pkgName), query) {
			// Also check metadata for keyword/description match
			meta, err := c.GetPackage(pkgName)
			if err != nil {
				continue
			}
			matched := strings.Contains(strings.ToLower(meta.Description), query)
			if !matched {
				for _, kw := range meta.Keywords {
					if strings.Contains(strings.ToLower(kw), query) {
						matched = true
						break
					}
				}
			}
			if !matched {
				continue
			}
			results = append(results, *meta)
			continue
		}

		meta, err := c.GetPackage(pkgName)
		if err != nil {
			continue
		}
		results = append(results, *meta)
	}

	return results, nil
}

// GetPackage returns the latest version metadata for a package from the local registry.
func (c *RegistryClient) GetPackage(name string) (*PackageMeta, error) {
	pkgDir := c.packageDir(name)

	versions, err := c.GetVersions(name)
	if err != nil || len(versions) == 0 {
		return nil, fmt.Errorf("package %q not found in registry", name)
	}

	// Use the last (latest) version
	latest := versions[len(versions)-1]
	return c.getVersionMeta(pkgDir, latest)
}

// GetVersions returns all available versions for a package, sorted by directory listing order.
func (c *RegistryClient) GetVersions(name string) ([]string, error) {
	pkgDir := c.packageDir(name)

	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, fmt.Errorf("package %q not found in registry", name)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() && ValidateVersion(entry.Name()) {
			versions = append(versions, entry.Name())
		}
	}

	return versions, nil
}

// getVersionMeta reads the meta.json file for a specific version.
func (c *RegistryClient) getVersionMeta(pkgDir, version string) (*PackageMeta, error) {
	metaPath := filepath.Join(pkgDir, version, "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("could not read package metadata: %w", err)
	}

	var meta PackageMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("could not parse package metadata: %w", err)
	}

	return &meta, nil
}

// Publish saves a package and its bundle to the local registry.
// The token parameter is accepted for future remote registry auth but is unused locally.
func (c *RegistryClient) Publish(meta *PackageMeta, bundle []byte, token string) error {
	if meta.Name == "" {
		return fmt.Errorf("package name is required")
	}
	if !ValidateVersion(meta.Version) {
		return fmt.Errorf("invalid version %q - must be semver (e.g. 1.0.0)", meta.Version)
	}

	vDir := c.versionDir(meta.Name, meta.Version)

	// Check if version already exists
	if _, err := os.Stat(vDir); err == nil {
		return fmt.Errorf("version %s of %s is already published", meta.Version, meta.Name)
	}

	if err := os.MkdirAll(vDir, 0755); err != nil {
		return fmt.Errorf("could not create registry directory: %w", err)
	}

	// Write metadata
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(vDir, "meta.json"), metaData, 0644); err != nil {
		return err
	}

	// Write bundle
	if err := os.WriteFile(filepath.Join(vDir, "package.tar.gz"), bundle, 0644); err != nil {
		return err
	}

	return nil
}

// Download retrieves the package bundle bytes for a specific version from the local registry.
// If version is empty or "*", the latest version is used.
func (c *RegistryClient) Download(name, version string) ([]byte, error) {
	if version == "" || version == "*" {
		versions, err := c.GetVersions(name)
		if err != nil || len(versions) == 0 {
			return nil, fmt.Errorf("package %q not found in registry", name)
		}
		version = versions[len(versions)-1]
	}

	// Strip semver prefixes for exact lookup
	cleanVersion := strings.TrimLeft(version, "^~>=<")
	if !ValidateVersion(cleanVersion) {
		return nil, fmt.Errorf("invalid version %q", version)
	}

	bundlePath := filepath.Join(c.versionDir(name, cleanVersion), "package.tar.gz")
	data, err := os.ReadFile(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("could not download %s@%s: %w", name, cleanVersion, err)
	}

	return data, nil
}
