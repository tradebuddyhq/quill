package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareBundle(t *testing.T) {
	// Create a temp directory for the test
	tmpDir, err := os.MkdirTemp("", "quill-deploy-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	config := DeployConfig{
		AppName:   "test-app",
		Entry:     "main.quill",
		Port:      3000,
		Env:       "production",
		OutputDir: filepath.Join(tmpDir, "dist"),
	}

	deployer := NewDeployer(config)
	outDir, err := deployer.PrepareBundle("console.log('hello');")
	if err != nil {
		t.Fatalf("PrepareBundle failed: %s", err)
	}

	// Check output directory exists
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		t.Error("output directory was not created")
	}

	// Check server.js exists
	serverJS, err := os.ReadFile(filepath.Join(outDir, "server.js"))
	if err != nil {
		t.Error("server.js was not created")
	}
	if !strings.Contains(string(serverJS), "console.log('hello')") {
		t.Error("server.js does not contain compiled code")
	}
	if !strings.Contains(string(serverJS), "NODE_ENV") {
		t.Error("server.js does not set NODE_ENV")
	}

	// Check package.json exists
	if _, err := os.Stat(filepath.Join(outDir, "package.json")); os.IsNotExist(err) {
		t.Error("package.json was not created")
	}

	// Check Dockerfile exists
	if _, err := os.Stat(filepath.Join(outDir, "Dockerfile")); os.IsNotExist(err) {
		t.Error("Dockerfile was not created")
	}

	// Check public/ directory exists
	if _, err := os.Stat(filepath.Join(outDir, "public")); os.IsNotExist(err) {
		t.Error("public/ directory was not created")
	}
}

func TestGenerateDockerfile(t *testing.T) {
	config := DeployConfig{
		AppName: "my-app",
		Port:    8080,
		Env:     "production",
	}
	deployer := NewDeployer(config)
	dockerfile := deployer.GenerateDockerfile()

	if !strings.Contains(dockerfile, "FROM node:20-alpine") {
		t.Error("Dockerfile missing base image")
	}
	if !strings.Contains(dockerfile, "EXPOSE 8080") {
		t.Error("Dockerfile missing EXPOSE")
	}
	if !strings.Contains(dockerfile, "NODE_ENV=production") {
		t.Error("Dockerfile missing NODE_ENV")
	}
	if !strings.Contains(dockerfile, "npm install --production") {
		t.Error("Dockerfile missing npm install")
	}
	if !strings.Contains(dockerfile, `CMD ["npm", "start"]`) {
		t.Error("Dockerfile missing CMD")
	}
}

func TestGeneratePackageJSON(t *testing.T) {
	config := DeployConfig{
		AppName: "my-app",
		Port:    3000,
	}
	deployer := NewDeployer(config)
	pkg := deployer.GeneratePackageJSON()

	if !strings.Contains(pkg, `"name": "my-app"`) {
		t.Error("package.json missing app name")
	}
	if !strings.Contains(pkg, `"start": "node server.js"`) {
		t.Error("package.json missing start script")
	}
	if !strings.Contains(pkg, `"better-sqlite3"`) {
		t.Error("package.json missing dependencies")
	}
}

func TestDeployerDefaults(t *testing.T) {
	deployer := NewDeployer(DeployConfig{})

	if deployer.config.Port != 3000 {
		t.Errorf("expected default port 3000, got %d", deployer.config.Port)
	}
	if deployer.config.Env != "production" {
		t.Errorf("expected default env 'production', got %s", deployer.config.Env)
	}
	if deployer.config.OutputDir != "dist" {
		t.Errorf("expected default output dir 'dist', got %s", deployer.config.OutputDir)
	}
	if deployer.config.AppName != "quill-app" {
		t.Errorf("expected default app name 'quill-app', got %s", deployer.config.AppName)
	}
}
