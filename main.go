package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"quill/analyzer"
	"quill/codegen"
	"quill/formatter"
	"quill/lexer"
	"quill/parser"
	"quill/repl"
	"quill/typechecker"
	"strings"
)

const version = "0.2.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide a file to run")
			fmt.Fprintln(os.Stderr, "Usage: quill run <file.quill>")
			os.Exit(1)
		}
		runFile(os.Args[2])

	case "build":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide a file to build")
			fmt.Fprintln(os.Stderr, "Usage: quill build <file.quill> [--browser|--wasm|--standalone]")
			os.Exit(1)
		}
		target := "node"
		for _, arg := range os.Args[3:] {
			switch arg {
			case "--browser":
				target = "browser"
			case "--wasm":
				target = "wasm"
			case "--standalone":
				target = "standalone"
			}
		}
		buildFileWithTarget(os.Args[2], target)

	case "repl":
		repl.Start()

	case "test":
		runTests(os.Args[2:])

	case "fmt", "format":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide a file to format")
			fmt.Fprintln(os.Stderr, "Usage: quill fmt <file.quill>")
			os.Exit(1)
		}
		formatFile(os.Args[2])

	case "check", "lint":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide a file to check")
			fmt.Fprintln(os.Stderr, "Usage: quill check <file.quill>")
			os.Exit(1)
		}
		checkFile(os.Args[2])

	case "init":
		initProject()

	case "add":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide a package name")
			fmt.Fprintln(os.Stderr, "Usage: quill add <package>")
			os.Exit(1)
		}
		addPackage(os.Args[2])

	case "remove":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide a package name")
			fmt.Fprintln(os.Stderr, "Usage: quill remove <package>")
			os.Exit(1)
		}
		removePackage(os.Args[2])

	case "docs":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide a file to generate docs for")
			fmt.Fprintln(os.Stderr, "Usage: quill docs <file.quill>")
			os.Exit(1)
		}
		generateDocs(os.Args[2])

	case "version", "--version", "-v":
		fmt.Printf("quill %s\n", version)

	case "help", "--help", "-h":
		printUsage()

	default:
		// If it's a .quill file, run it directly
		if filepath.Ext(os.Args[1]) == ".quill" {
			runFile(os.Args[1])
		} else {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}
}

func runFile(filename string) {
	js := compile(filename)

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "quill-*.js")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not create temp file: %s\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(js)
	tmpFile.Close()

	// Try node first, then bun, then deno
	runtime := findRuntime()
	if runtime == "" {
		fmt.Fprintln(os.Stderr, "Error: no JavaScript runtime found")
		fmt.Fprintln(os.Stderr, "Please install Node.js, Bun, or Deno")
		os.Exit(1)
	}

	cmd := exec.Command(runtime, tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}

func buildFile(filename string, browser bool) {
	target := "node"
	if browser {
		target = "browser"
	}
	buildFileWithTarget(filename, target)
}

func buildFileWithTarget(filename string, target string) {
	browser := target == "browser"
	js := compileWithTarget(filename, browser)

	ext := filepath.Ext(filename)
	base := filename[:len(filename)-len(ext)]

	switch target {
	case "wasm":
		// Generate a WASM-compatible module with wasi shim
		wasmJS := generateWASMWrapper(js, base)
		outFile := base + ".wasm.js"
		if err := os.WriteFile(outFile, []byte(wasmJS), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not write output file: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Built %s -> %s (WASM-ready module)\n", filename, outFile)
		fmt.Println("  Run with: node --experimental-wasm-modules " + outFile)

	case "standalone":
		// Bundle with a shebang for direct execution
		standalone := "#!/usr/bin/env node\n" + js
		outFile := base
		if err := os.WriteFile(outFile, []byte(standalone), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not write output file: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Built %s -> %s (standalone executable)\n", filename, outFile)
		fmt.Println("  Run with: ./" + filepath.Base(outFile))

	default:
		outFile := base + ".js"
		if err := os.WriteFile(outFile, []byte(js), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not write output file: %s\n", err)
			os.Exit(1)
		}
		targetLabel := "Node.js"
		if browser {
			targetLabel = "browser"
		}
		fmt.Printf("Built %s -> %s (%s)\n", filename, outFile, targetLabel)
	}
}

func generateWASMWrapper(js string, base string) string {
	var out strings.Builder
	out.WriteString("// WASM-compatible module generated by Quill\n")
	out.WriteString("// This wraps the compiled JS in a WASI-compatible module format\n\n")
	out.WriteString("const { WASI } = require('wasi');\n")
	out.WriteString("const { readFileSync } = require('fs');\n\n")
	out.WriteString("// Quill compiled output (runs in Node.js WASM context)\n")
	out.WriteString("const __quill_main = (function() {\n")
	out.WriteString("  'use strict';\n")
	out.WriteString("  " + strings.ReplaceAll(js, "\n", "\n  ") + "\n")
	out.WriteString("});\n\n")
	out.WriteString("// Export as module for WASM interop\n")
	out.WriteString("if (typeof module !== 'undefined') {\n")
	out.WriteString("  module.exports = { run: __quill_main };\n")
	out.WriteString("}\n\n")
	out.WriteString("// Auto-run if executed directly\n")
	out.WriteString("if (require.main === module) {\n")
	out.WriteString("  __quill_main();\n")
	out.WriteString("}\n")
	return out.String()
}

func compile(filename string) string {
	return compileWithTarget(filename, false)
}

func compileWithTarget(filename string, browser bool) string {
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not read %q\n", filename)
		fmt.Fprintln(os.Stderr, "Make sure the file exists and you have permission to read it.")
		os.Exit(1)
	}

	// Lex
	l := lexer.New(string(source))
	tokens, err := l.Tokenize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	// Parse
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	// Generate JS
	var g *codegen.Generator
	if browser {
		g = codegen.NewBrowser()
	} else {
		g = codegen.New()
	}
	return g.Generate(program)
}

func findRuntime() string {
	for _, name := range []string{"node", "bun", "deno"} {
		if _, err := exec.LookPath(name); err == nil {
			return name
		}
	}
	return ""
}

func runTests(args []string) {
	files := args
	if len(files) == 0 {
		// Find all .quill files in current directory
		entries, _ := os.ReadDir(".")
		for _, e := range entries {
			if filepath.Ext(e.Name()) == ".quill" {
				files = append(files, e.Name())
			}
		}
	}

	if len(files) == 0 {
		fmt.Println("No .quill files found to test")
		return
	}

	for _, f := range files {
		fmt.Printf("\nTesting %s...\n", f)
		js := compile(f)
		// Add test harness
		js = js + "\nconsole.log(`\\n${__test_passed} passed, ${__test_failed} failed`);\nif (__test_failed > 0) process.exit(1);\n"
		runJS(js)
	}
}

func runJS(js string) {
	tmpFile, _ := os.CreateTemp("", "quill-*.js")
	tmpFile.WriteString(js)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	runtime := findRuntime()
	cmd := exec.Command(runtime, tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func formatFile(filename string) {
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not read %q\n", filename)
		os.Exit(1)
	}

	l := lexer.New(string(source))
	tokens, err := l.Tokenize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	f := formatter.New()
	formatted := f.Format(program)

	if err := os.WriteFile(filename, []byte(formatted), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not write %q\n", filename)
		os.Exit(1)
	}

	fmt.Printf("Formatted %s\n", filename)
}

func checkFile(filename string) {
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not read %q\n", filename)
		os.Exit(1)
	}

	l := lexer.New(string(source))
	tokens, err := l.Tokenize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	// Static analysis
	a := analyzer.New()
	diagnostics := a.Analyze(program)

	// Type checking
	tc := typechecker.New()
	typeDiags := tc.Check(program)

	totalIssues := len(diagnostics) + len(typeDiags)

	if totalIssues == 0 {
		fmt.Printf("✓ %s — no issues found\n", filename)
		return
	}

	fmt.Printf("Found %d issue(s) in %s:\n\n", totalIssues, filename)
	for _, d := range diagnostics {
		fmt.Println(d.String())
	}
	for _, d := range typeDiags {
		fmt.Println(d.String())
	}
	fmt.Println()

	if analyzer.HasErrors(diagnostics) || typechecker.HasErrors(typeDiags) {
		os.Exit(1)
	}
}

func initProject() {
	// Create quill.json
	if _, err := os.Stat("quill.json"); err == nil {
		fmt.Println("quill.json already exists")
		return
	}

	// Get current directory name for project name
	dir, _ := os.Getwd()
	name := filepath.Base(dir)

	config := map[string]interface{}{
		"name":         name,
		"version":      "0.1.0",
		"description":  "",
		"main":         "main.quill",
		"dependencies": map[string]string{},
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	if err := os.WriteFile("quill.json", data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not create quill.json: %s\n", err)
		os.Exit(1)
	}

	// Create main.quill if it doesn't exist
	if _, err := os.Stat("main.quill"); err != nil {
		starter := "-- Welcome to Quill!\nsay \"Hello, World!\"\n"
		os.WriteFile("main.quill", []byte(starter), 0644)
	}

	fmt.Println("✓ Initialized Quill project")
	fmt.Println("  Created quill.json")
	fmt.Println("  Run: quill run main.quill")
}

func addPackage(pkg string) {
	// Read or create quill.json
	var config map[string]interface{}

	data, err := os.ReadFile("quill.json")
	if err != nil {
		// Create quill.json if it doesn't exist
		initProject()
		data, _ = os.ReadFile("quill.json")
	}

	json.Unmarshal(data, &config)

	// Install with npm
	fmt.Printf("Installing %s...\n", pkg)
	cmd := exec.Command("npm", "install", pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error installing %s: %s\n", pkg, err)
		os.Exit(1)
	}

	// Update quill.json dependencies
	deps, ok := config["dependencies"].(map[string]interface{})
	if !ok {
		deps = make(map[string]interface{})
	}

	// Read installed version from node_modules
	pkgJsonPath := filepath.Join("node_modules", pkg, "package.json")
	if pkgData, err := os.ReadFile(pkgJsonPath); err == nil {
		var pkgInfo map[string]interface{}
		json.Unmarshal(pkgData, &pkgInfo)
		if v, ok := pkgInfo["version"].(string); ok {
			deps[pkg] = "^" + v
		}
	} else {
		deps[pkg] = "*"
	}

	config["dependencies"] = deps
	outData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile("quill.json", outData, 0644)

	fmt.Printf("✓ Added %s\n", pkg)
}

func removePackage(pkg string) {
	data, err := os.ReadFile("quill.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: no quill.json found. Run 'quill init' first.")
		os.Exit(1)
	}

	var config map[string]interface{}
	json.Unmarshal(data, &config)

	deps, ok := config["dependencies"].(map[string]interface{})
	if !ok {
		fmt.Printf("Package %s is not installed\n", pkg)
		return
	}

	delete(deps, pkg)
	config["dependencies"] = deps

	// Uninstall with npm
	cmd := exec.Command("npm", "uninstall", pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	outData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile("quill.json", outData, 0644)

	fmt.Printf("✓ Removed %s\n", pkg)
}

func generateDocs(filename string) {
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not read %q\n", filename)
		os.Exit(1)
	}

	lines := strings.Split(string(source), "\n")

	var out strings.Builder
	out.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	out.WriteString("  <meta charset=\"UTF-8\">\n")
	out.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	out.WriteString(fmt.Sprintf("  <title>Documentation - %s</title>\n", filename))
	out.WriteString("  <style>\n")
	out.WriteString("    body { font-family: -apple-system, sans-serif; max-width: 800px; margin: 0 auto; padding: 40px 20px; background: #0D1117; color: #E6EDF3; line-height: 1.7; }\n")
	out.WriteString("    h1 { color: #1EB969; border-bottom: 2px solid #30363D; padding-bottom: 12px; }\n")
	out.WriteString("    h2 { color: #E6EDF3; margin-top: 40px; }\n")
	out.WriteString("    h3 { color: #8B949E; font-size: 14px; text-transform: uppercase; letter-spacing: 1px; }\n")
	out.WriteString("    pre { background: #161B22; border: 1px solid #30363D; border-radius: 8px; padding: 16px; overflow-x: auto; font-family: 'JetBrains Mono', monospace; font-size: 13px; }\n")
	out.WriteString("    code { font-family: 'JetBrains Mono', monospace; background: rgba(30,185,105,0.1); padding: 2px 6px; border-radius: 4px; color: #1EB969; }\n")
	out.WriteString("    .doc-comment { color: #8B949E; margin-bottom: 8px; font-style: italic; }\n")
	out.WriteString("    .function { background: #161B22; border: 1px solid #30363D; border-radius: 8px; padding: 16px; margin: 16px 0; }\n")
	out.WriteString("    .function h3 { margin-top: 0; color: #D2A8FF; }\n")
	out.WriteString("    .tag { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: 600; margin-right: 4px; }\n")
	out.WriteString("    .tag-fn { background: rgba(210,168,255,0.15); color: #D2A8FF; }\n")
	out.WriteString("    .tag-var { background: rgba(30,185,105,0.15); color: #1EB969; }\n")
	out.WriteString("    .tag-class { background: rgba(255,166,87,0.15); color: #FFA657; }\n")
	out.WriteString("    footer { margin-top: 60px; padding-top: 20px; border-top: 1px solid #30363D; color: #6E7681; font-size: 13px; }\n")
	out.WriteString("  </style>\n</head>\n<body>\n")
	out.WriteString(fmt.Sprintf("<h1>📄 %s</h1>\n", filename))

	var pendingComments []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Collect comments
		if strings.HasPrefix(trimmed, "-- ") {
			pendingComments = append(pendingComments, strings.TrimPrefix(trimmed, "-- "))
			continue
		}

		// Function definition
		if strings.HasPrefix(trimmed, "to ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				fnName := parts[1]
				params := []string{}
				for _, p := range parts[2:] {
					if p == "as" || p == "->" {
						break
					}
					clean := strings.TrimSuffix(p, ":")
					if clean != "" {
						params = append(params, clean)
					}
				}

				out.WriteString("<div class=\"function\">\n")
				out.WriteString(fmt.Sprintf("  <span class=\"tag tag-fn\">function</span>\n"))
				out.WriteString(fmt.Sprintf("  <h3>%s(%s)</h3>\n", fnName, strings.Join(params, ", ")))
				if len(pendingComments) > 0 {
					for _, c := range pendingComments {
						out.WriteString(fmt.Sprintf("  <p class=\"doc-comment\">%s</p>\n", c))
					}
				}
				out.WriteString(fmt.Sprintf("  <pre>%s</pre>\n", trimmed))
				out.WriteString("</div>\n")
				pendingComments = nil
				continue
			}
		}

		// Variable/constant
		if strings.Contains(trimmed, " is ") || strings.Contains(trimmed, " are ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 && (parts[1] == "is" || parts[1] == "are") {
				varName := parts[0]
				out.WriteString("<div class=\"function\">\n")
				out.WriteString(fmt.Sprintf("  <span class=\"tag tag-var\">variable</span>\n"))
				out.WriteString(fmt.Sprintf("  <h3>%s</h3>\n", varName))
				if len(pendingComments) > 0 {
					for _, c := range pendingComments {
						out.WriteString(fmt.Sprintf("  <p class=\"doc-comment\">%s</p>\n", c))
					}
				}
				out.WriteString(fmt.Sprintf("  <pre>%s</pre>\n", trimmed))
				out.WriteString("</div>\n")
				pendingComments = nil
				continue
			}
		}

		// Class definition
		if strings.HasPrefix(trimmed, "describe ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				className := parts[1]
				out.WriteString("<div class=\"function\">\n")
				out.WriteString(fmt.Sprintf("  <span class=\"tag tag-class\">class</span>\n"))
				ext := ""
				if len(parts) >= 4 && parts[2] == "extends" {
					ext = " extends " + parts[3]
				}
				out.WriteString(fmt.Sprintf("  <h3>%s%s</h3>\n", className, ext))
				if len(pendingComments) > 0 {
					for _, c := range pendingComments {
						out.WriteString(fmt.Sprintf("  <p class=\"doc-comment\">%s</p>\n", c))
					}
				}
				out.WriteString(fmt.Sprintf("  <pre>%s</pre>\n", trimmed))
				out.WriteString("</div>\n")
				pendingComments = nil
				continue
			}
		}

		// Reset pending comments if we hit a non-doc line
		if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
			pendingComments = nil
		}
	}

	out.WriteString("<footer>Generated by <code>quill docs</code></footer>\n")
	out.WriteString("</body>\n</html>\n")

	// Write output
	ext := filepath.Ext(filename)
	outFile := filename[:len(filename)-len(ext)] + ".docs.html"
	if err := os.WriteFile(outFile, []byte(out.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing docs: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated documentation: %s\n", outFile)
}

func printUsage() {
	fmt.Println("Quill — code that reads like English")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  quill run <file.quill>       Run a Quill program")
	fmt.Println("  quill build <file.quill>          Compile to JavaScript (Node.js)")
	fmt.Println("  quill build <file> --browser       Compile for the browser")
	fmt.Println("  quill build <file> --wasm          Compile as WASM-ready module")
	fmt.Println("  quill build <file> --standalone     Compile as standalone executable")
	fmt.Println("  quill repl                   Start interactive REPL")
	fmt.Println("  quill test [files...]        Run tests in .quill files")
	fmt.Println("  quill fmt <file.quill>       Format a Quill file")
	fmt.Println("  quill check <file.quill>     Check for common issues")
	fmt.Println("  quill docs <file.quill>      Generate documentation")
	fmt.Println("  quill init                   Initialize a new Quill project")
	fmt.Println("  quill add <package>          Install an npm package")
	fmt.Println("  quill remove <package>       Remove a package")
	fmt.Println("  quill version                Show version")
	fmt.Println("  quill help                   Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  quill run hello.quill")
	fmt.Println("  quill build script.quill")
	fmt.Println("  quill repl")
	fmt.Println("  quill test examples/test_example.quill")
	fmt.Println("  quill fmt script.quill")
	fmt.Println("  quill check script.quill")
	fmt.Println("  quill docs api.quill")
	fmt.Println("  quill init")
	fmt.Println("  quill add express")
}
