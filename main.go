package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"quill/analyzer"
	"quill/ast"
	"quill/codegen"
	"quill/config"
	"quill/debugger"
	quillerrors "quill/errors"
	"quill/formatter"
	"quill/lexer"
	"quill/lsp"
	"quill/parser"
	"quill/registry"
	"quill/learn"
	"quill/repl"
	"quill/server"
	"quill/tools"
	"quill/typechecker"
	"strconv"
	"strings"
	"time"
)

const version = "0.10.5"

// displayCompileError formats a compile error with source context and prints to stderr.
// It extracts line/column from ParseError or falls back to a plain message.
func displayCompileError(err error, source []byte, filename string) {
	sourceLines := strings.Split(string(source), "\n")
	// Try to extract line info from the error message
	if pe, ok := err.(*parser.ParseError); ok {
		de := quillerrors.DisplayError{
			Line:    pe.Line,
			Message: pe.Message,
			Code:    "E001",
		}
		fmt.Fprint(os.Stderr, quillerrors.FormatError(de, sourceLines, filename, true))
	} else {
		// Try to parse "line N: message" or "line N, column M: message" from error string
		msg := err.Error()
		line, col, cleanMsg := parseErrorLocation(msg)
		if line > 0 {
			de := quillerrors.DisplayError{
				Line:    line,
				Column:  col,
				Message: cleanMsg,
				Code:    "E001",
			}
			fmt.Fprint(os.Stderr, quillerrors.FormatError(de, sourceLines, filename, true))
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
	}
}

// parseErrorLocation extracts line, column, and message from error strings like
// "line 5: message" or "line 5, column 10: message"
func parseErrorLocation(msg string) (int, int, string) {
	if !strings.HasPrefix(msg, "line ") {
		return 0, 0, msg
	}
	rest := msg[5:]
	line := 0
	col := 0
	i := 0
	for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
		line = line*10 + int(rest[i]-'0')
		i++
	}
	if line == 0 {
		return 0, 0, msg
	}
	rest = rest[i:]
	if strings.HasPrefix(rest, ", column ") {
		rest = rest[9:]
		j := 0
		for j < len(rest) && rest[j] >= '0' && rest[j] <= '9' {
			col = col*10 + int(rest[j]-'0')
			j++
		}
		rest = rest[j:]
	}
	if strings.HasPrefix(rest, ": ") {
		rest = rest[2:]
	}
	return line, col, rest
}

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
		runFileWithFullStack(os.Args[2])

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
			case "--llvm", "--native":
				target = "llvm"
			case "--expo":
				target = "expo"
			}
		}
		buildFileWithTarget(os.Args[2], target)

	case "debug":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide a file to debug")
			fmt.Fprintln(os.Stderr, "Usage: quill debug <file.quill>")
			os.Exit(1)
		}
		debugger.StartREPL(os.Args[2])

	case "repl":
		repl.Version = version
		repl.Start()

	case "learn", "tutorial":
		learn.Run()

	case "lsp":
		lsp.Start()

	case "test":
		runTestsCommand(os.Args[2:])

	case "profile":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide a file to profile")
			fmt.Fprintln(os.Stderr, "Usage: quill profile <file.quill>")
			os.Exit(1)
		}
		profileFile(os.Args[2])

	case "fix":
		runMigration(os.Args[2:])

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

	case "new":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide a project name")
			fmt.Fprintln(os.Stderr, "Usage: quill new <name>")
			os.Exit(1)
		}
		newProject(os.Args[2])

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

	case "publish":
		publishPackage()

	case "search":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide a search query")
			fmt.Fprintln(os.Stderr, "Usage: quill search <query>")
			os.Exit(1)
		}
		searchRegistry(os.Args[2])

	case "install":
		installDependencies()

	case "bump":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide bump type (major, minor, or patch)")
			fmt.Fprintln(os.Stderr, "Usage: quill bump <major|minor|patch>")
			os.Exit(1)
		}
		bumpVersion(os.Args[2])

	case "watch":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide a file to watch")
			fmt.Fprintln(os.Stderr, "Usage: quill watch <file.quill>")
			os.Exit(1)
		}
		watchFile(os.Args[2])

	case "serve":
		serveApp()

	case "deploy":
		deployApp()

	case "share":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide a file to share")
			fmt.Fprintln(os.Stderr, "Usage: quill share <file.quill>")
			os.Exit(1)
		}
		shareFile(os.Args[2])

	case "db":
		dbCommand(os.Args[2:])

	case "generate":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: please provide a prompt")
			fmt.Fprintln(os.Stderr, "Usage: quill generate \"<prompt>\"")
			os.Exit(1)
		}
		generateApp(os.Args[2])

	case "discord":
		scaffoldDiscordBot()

	case "web":
		scaffoldWebServer()

	case "worker":
		scaffoldWorker()

	case "ai":
		scaffoldAI()

	case "expo":
		scaffoldExpo()

	case "cli":
		scaffoldCLI()

	case "site":
		scaffoldSite()

	case "version", "--version", "-v":
		fmt.Printf("quill %s\n", version)

	case "help", "--help", "-h":
		printUsage()

	default:
		// If it's a .quill file, run it directly
		if filepath.Ext(os.Args[1]) == ".quill" {
			runFileWithFullStack(os.Args[1])
		} else {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}
}

func runFile(filename string) {
	js := compile(filename)

	// Resolve the source file's directory so node can find node_modules
	absPath, err := filepath.Abs(filename)
	if err != nil {
		absPath = filename
	}
	sourceDir := filepath.Dir(absPath)

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
	cmd.Dir = sourceDir

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}

// watchFile watches a .quill file (and sibling .quill files) for changes,
// automatically recompiling and re-running on each change.
func watchFile(filename string) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	// Verify the file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", filename)
		os.Exit(1)
	}

	sourceDir := filepath.Dir(absPath)

	// Collect all .quill files in the same directory to watch
	watchPaths := []string{absPath}
	entries, err := os.ReadDir(sourceDir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".quill" {
				full := filepath.Join(sourceDir, e.Name())
				if full != absPath {
					watchPaths = append(watchPaths, full)
				}
			}
		}
	}

	// Handle Ctrl+C gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	// Track modification times
	modTimes := make(map[string]time.Time)
	for _, p := range watchPaths {
		if info, err := os.Stat(p); err == nil {
			modTimes[p] = info.ModTime()
		}
	}

	// Find JS runtime once
	runtime := findRuntime()
	if runtime == "" {
		fmt.Fprintln(os.Stderr, "Error: no JavaScript runtime found")
		fmt.Fprintln(os.Stderr, "Please install Node.js, Bun, or Deno")
		os.Exit(1)
	}

	fmt.Printf("Watching %s (and %d sibling .quill files)\n", filename, len(watchPaths)-1)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// compileAndRun compiles the main file and runs it, returning the process
	compileAndRun := func() *exec.Cmd {
		js := compile(absPath)

		tmpFile, err := os.CreateTemp("", "quill-watch-*.js")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not create temp file: %s\n", err)
			return nil
		}
		tmpFile.WriteString(js)
		tmpFile.Close()

		cmd := exec.Command(runtime, tmpFile.Name())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Dir = sourceDir
		// Use a process group so we can kill child processes too
		setProcGroup(cmd)

		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting process: %s\n", err)
			os.Remove(tmpFile.Name())
			return nil
		}

		// Clean up temp file when process exits
		go func() {
			cmd.Wait()
			os.Remove(tmpFile.Name())
		}()

		return cmd
	}

	killProcess := func(cmd *exec.Cmd) {
		killProcGroup(cmd)
	}

	// Initial run
	fmt.Printf("[%s] Starting %s\n", time.Now().Format("15:04:05"), filename)
	currentCmd := compileAndRun()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			fmt.Println("\nStopping...")
			killProcess(currentCmd)
			os.Exit(0)

		case <-ticker.C:
			changed := false
			for _, p := range watchPaths {
				info, err := os.Stat(p)
				if err != nil {
					continue
				}
				if prev, ok := modTimes[p]; ok {
					if info.ModTime().After(prev) {
						changed = true
						modTimes[p] = info.ModTime()
					}
				} else {
					modTimes[p] = info.ModTime()
				}
			}

			// Also check for new .quill files
			if entries, err := os.ReadDir(sourceDir); err == nil {
				for _, e := range entries {
					if !e.IsDir() && filepath.Ext(e.Name()) == ".quill" {
						full := filepath.Join(sourceDir, e.Name())
						if _, ok := modTimes[full]; !ok {
							watchPaths = append(watchPaths, full)
							if info, err := os.Stat(full); err == nil {
								modTimes[full] = info.ModTime()
							}
							changed = true
						}
					}
				}
			}

			if changed {
				fmt.Printf("\n[%s] File changed, restarting...\n", time.Now().Format("15:04:05"))
				killProcess(currentCmd)
				currentCmd = compileAndRun()
			}
		}
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
	ext := filepath.Ext(filename)
	base := filename[:len(filename)-len(ext)]

	if target == "llvm" {
		buildLLVM(filename, base)
		return
	}

	if target == "expo" {
		buildExpo(filename, base)
		return
	}

	browser := target == "browser"
	js := compileWithTarget(filename, browser)

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
		mapFile := outFile + ".map"

		// Generate with source map
		jsWithMap, sourceMapJSON := compileWithSourceMap(filename, browser)

		// Append source mapping URL
		jsWithMap = jsWithMap + "\n//# sourceMappingURL=" + filepath.Base(mapFile) + "\n"

		if err := os.WriteFile(outFile, []byte(jsWithMap), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not write output file: %s\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(mapFile, []byte(sourceMapJSON), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not write source map file: %s\n", err)
			os.Exit(1)
		}
		targetLabel := "Node.js"
		if browser {
			targetLabel = "browser"
		}
		fmt.Printf("Built %s -> %s (%s)\n", filename, outFile, targetLabel)
		fmt.Printf("  Source map: %s\n", mapFile)
	}
}

func buildLLVM(filename string, base string) {
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not read %q\n", filename)
		os.Exit(1)
	}

	l := lexer.New(string(source))
	tokens, err := l.Tokenize()
	if err != nil {
		displayCompileError(err, source, filename)
		os.Exit(1)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		displayCompileError(err, source, filename)
		os.Exit(1)
	}

	gen := codegen.NewLLVM()
	ir := gen.Generate(program)

	// Print warnings about unsupported features
	if len(gen.Warnings) > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠ Native build: %d unsupported feature(s) detected:\n", len(gen.Warnings))
		for _, w := range gen.Warnings {
			fmt.Fprintf(os.Stderr, "  • %s\n", w)
		}
		fmt.Fprintf(os.Stderr, "\nThe native binary will compile but these features will not work.\n")
		fmt.Fprintf(os.Stderr, "Use 'quill build %s' (without --native) for full feature support.\n\n", filename)
	}

	outFile := base + ".ll"
	if err := os.WriteFile(outFile, []byte(ir), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not write output file: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Built %s -> %s (LLVM IR)\n", filename, outFile)

	// Try to compile to native binary if llc and cc are available
	if _, err := exec.LookPath("llc"); err == nil {
		if _, err := exec.LookPath("cc"); err == nil {
			objFile := base + ".o"
			binFile := base

			llcCmd := exec.Command("llc", "-filetype=obj", outFile, "-o", objFile)
			llcCmd.Stderr = os.Stderr
			if err := llcCmd.Run(); err == nil {
				ccCmd := exec.Command("cc", objFile, "-o", binFile, "-lm")
				ccCmd.Stderr = os.Stderr
				if err := ccCmd.Run(); err == nil {
					fmt.Printf("  Compiled native binary: %s\n", binFile)
					fmt.Println("  Run with: ./" + filepath.Base(binFile))
					os.Remove(objFile)
					return
				}
				os.Remove(objFile)
			}
		}
	}

	fmt.Println("  To compile: llc -filetype=obj " + outFile + " -o " + base + ".o && cc " + base + ".o -o " + base + " -lm")
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
		displayCompileError(err, source, filename)
		os.Exit(1)
	}

	// Parse
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		displayCompileError(err, source, filename)
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

func compileWithSourceMap(filename string, browser bool) (string, string) {
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not read %q\n", filename)
		fmt.Fprintln(os.Stderr, "Make sure the file exists and you have permission to read it.")
		os.Exit(1)
	}

	l := lexer.New(string(source))
	tokens, err := l.Tokenize()
	if err != nil {
		displayCompileError(err, source, filename)
		os.Exit(1)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		displayCompileError(err, source, filename)
		os.Exit(1)
	}

	ext := filepath.Ext(filename)
	base := filename[:len(filename)-len(ext)]
	outFile := base + ".js"

	var g *codegen.Generator
	if browser {
		g = codegen.NewBrowser()
	} else {
		g = codegen.New()
	}
	return g.GenerateWithSourceMap(program, filename, outFile)
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
		displayCompileError(err, source, filename)
		os.Exit(1)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		displayCompileError(err, source, filename)
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

	sourceStr := string(source)
	sourceLines := strings.Split(sourceStr, "\n")

	l := lexer.New(sourceStr)
	tokens, err := l.Tokenize()
	if err != nil {
		displayCompileError(err, source, filename)
		os.Exit(1)
	}

	// Use error recovery to collect all parse errors
	program, parseErrors := parser.ParseWithRecovery(tokens)

	// Display parse errors using the new error display system
	if len(parseErrors) > 0 {
		displayErrors := make([]quillerrors.DisplayError, len(parseErrors))
		for i, pe := range parseErrors {
			displayErrors[i] = quillerrors.DisplayError{
				Line:    pe.Line,
				Column:  pe.Column,
				Message: pe.Message,
				Hint:    pe.Hint,
				Code:    pe.Code,
			}
		}
		fmt.Fprint(os.Stderr, quillerrors.FormatErrors(displayErrors, sourceLines, filename, true))
	}

	// Run static analysis and type checking on the partial program
	var diagnostics []analyzer.Diagnostic
	var typeDiags []typechecker.TypeDiagnostic
	if program != nil {
		a := analyzer.New()
		diagnostics = a.Analyze(program)

		tc := typechecker.New()
		typeDiags = tc.Check(program)
	}

	totalIssues := len(parseErrors) + len(diagnostics) + len(typeDiags)

	if totalIssues == 0 {
		fmt.Printf("✓ %s — no issues found\n", filename)
		return
	}

	// Print analyzer and type checker diagnostics
	if len(diagnostics) > 0 || len(typeDiags) > 0 {
		for _, d := range diagnostics {
			fmt.Println(d.String())
		}
		for _, d := range typeDiags {
			fmt.Println(d.String())
		}
		fmt.Println()
	}

	if len(parseErrors) > 0 || analyzer.HasErrors(diagnostics) || typechecker.HasErrors(typeDiags) {
		os.Exit(1)
	}
}

func initProject() {
	// Check if either config file already exists
	if _, err := os.Stat("quill.toml"); err == nil {
		fmt.Println("quill.toml already exists")
		return
	}
	if _, err := os.Stat("quill.json"); err == nil {
		fmt.Println("quill.json already exists")
		return
	}

	// Get current directory name for default project name
	dir, _ := os.Getwd()
	defaultName := filepath.Base(dir)

	// Interactive prompts
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Project name (%s): ", defaultName)
	nameInput, _ := reader.ReadString('\n')
	nameInput = strings.TrimSpace(nameInput)
	if nameInput == "" {
		nameInput = defaultName
	}

	fmt.Printf("Version (0.1.0): ")
	versionInput, _ := reader.ReadString('\n')
	versionInput = strings.TrimSpace(versionInput)
	if versionInput == "" {
		versionInput = "0.1.0"
	}

	fmt.Printf("Target (js/browser/llvm) [js]: ")
	targetInput, _ := reader.ReadString('\n')
	targetInput = strings.TrimSpace(targetInput)
	if targetInput == "" {
		targetInput = "js"
	}

	// Create quill.toml
	cfg := config.DefaultConfig()
	cfg.Project.Name = nameInput
	cfg.Project.Version = versionInput
	cfg.Build.Target = targetInput

	tomlContent := config.GenerateTOML(cfg)
	if err := os.WriteFile("quill.toml", []byte(tomlContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not create quill.toml: %s\n", err)
		os.Exit(1)
	}

	// Also create quill.json for backward compatibility with the package system
	jsonConfig := map[string]interface{}{
		"name":         nameInput,
		"version":      versionInput,
		"description":  "",
		"author":       "",
		"license":      "MIT",
		"main":         "main.quill",
		"keywords":     []string{},
		"repository":   "",
		"dependencies": map[string]string{},
	}
	data, _ := json.MarshalIndent(jsonConfig, "", "  ")
	os.WriteFile("quill.json", data, 0644)

	// Create main.quill if it doesn't exist
	if _, err := os.Stat("main.quill"); err != nil {
		starter := "-- Welcome to Quill!\nsay \"Hello, World!\"\n"
		os.WriteFile("main.quill", []byte(starter), 0644)
	}

	fmt.Println("✓ Initialized Quill project")
	fmt.Println("  Created quill.toml")
	fmt.Println("  Created quill.json")
	fmt.Println("  Run: quill run main.quill")
}

func newProject(name string) {
	// Validate name
	if name == "" {
		fmt.Fprintln(os.Stderr, "Error: please provide a project name")
		os.Exit(1)
	}

	// Check if directory already exists
	if _, err := os.Stat(name); err == nil {
		fmt.Fprintf(os.Stderr, "Error: directory %q already exists\n", name)
		os.Exit(1)
	}

	// Create directory
	if err := os.MkdirAll(name, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not create directory: %s\n", err)
		os.Exit(1)
	}

	// Create quill.toml
	cfg := config.DefaultConfig()
	cfg.Project.Name = name
	cfg.Project.Version = "0.1.0"
	cfg.Build.Target = "js"
	tomlContent := config.GenerateTOML(cfg)
	os.WriteFile(filepath.Join(name, "quill.toml"), []byte(tomlContent), 0644)

	// Create quill.json
	jsonConfig := map[string]interface{}{
		"name":         name,
		"version":      "0.1.0",
		"description":  "",
		"author":       "",
		"license":      "MIT",
		"main":         "main.quill",
		"keywords":     []string{},
		"repository":   "",
		"dependencies": map[string]string{},
	}
	data, _ := json.MarshalIndent(jsonConfig, "", "  ")
	os.WriteFile(filepath.Join(name, "quill.json"), data, 0644)

	// Create main.quill
	starter := fmt.Sprintf("-- %s\n-- Created with Quill\n\nsay \"Hello from %s!\"\n", name, name)
	os.WriteFile(filepath.Join(name, "main.quill"), []byte(starter), 0644)

	fmt.Printf("Created project %q\n", name)
	fmt.Println("  cd " + name)
	fmt.Println("  quill run main.quill")
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

	// Update quill.json dependencies
	deps, ok := config["dependencies"].(map[string]interface{})
	if !ok {
		deps = make(map[string]interface{})
	}

	// First, check the Quill registry
	client := registry.NewClient()
	if meta, err := client.GetPackage(pkg); err == nil {
		fmt.Printf("Found %s@%s in Quill registry\n", pkg, meta.Version)

		// Download and install to quill_modules
		bundle, err := client.Download(pkg, meta.Version)
		if err == nil {
			destDir := filepath.Join("quill_modules", pkg)
			os.MkdirAll(destDir, 0755)
			if err := registry.UnpackBundle(bundle, destDir); err == nil {
				deps[pkg] = "^" + meta.Version
				config["dependencies"] = deps
				outData, _ := json.MarshalIndent(config, "", "  ")
				os.WriteFile("quill.json", outData, 0644)
				fmt.Printf("✓ Added %s@%s from Quill registry\n", pkg, meta.Version)
				return
			}
		}
	}

	// Fallback to npm
	fmt.Printf("Installing %s via npm...\n", pkg)
	cmd := exec.Command("npm", "install", pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error installing %s: %s\n", pkg, err)
		os.Exit(1)
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

func publishPackage() {
	meta, err := registry.ReadPackageMeta(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: could not read quill.json. Run 'quill init' first.")
		os.Exit(1)
	}

	if meta.Name == "" {
		fmt.Fprintln(os.Stderr, "Error: package name is required in quill.json")
		os.Exit(1)
	}

	if !registry.ValidateVersion(meta.Version) {
		fmt.Fprintf(os.Stderr, "Error: invalid version %q in quill.json — must be semver (e.g. 1.0.0)\n", meta.Version)
		os.Exit(1)
	}

	fmt.Printf("Publishing %s@%s...\n", meta.Name, meta.Version)

	bundle, err := registry.PackageBundle(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating package bundle: %s\n", err)
		os.Exit(1)
	}

	client := registry.NewClient()
	if err := client.Publish(meta, bundle, ""); err != nil {
		fmt.Fprintf(os.Stderr, "Error publishing: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Published %s@%s to local registry\n", meta.Name, meta.Version)
}

func searchRegistry(query string) {
	client := registry.NewClient()
	results, err := client.Search(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error searching registry: %s\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Printf("No packages found matching %q\n", query)
		return
	}

	fmt.Printf("Found %d package(s):\n\n", len(results))
	for _, pkg := range results {
		desc := pkg.Description
		if desc == "" {
			desc = "(no description)"
		}
		fmt.Printf("  %s@%s\n", pkg.Name, pkg.Version)
		fmt.Printf("    %s\n\n", desc)
	}
}

func installDependencies() {
	meta, err := registry.ReadPackageMeta(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: could not read quill.json. Run 'quill init' first.")
		os.Exit(1)
	}

	if len(meta.Dependencies) == 0 {
		fmt.Println("No dependencies to install")
		return
	}

	fmt.Printf("Installing %d dependenc", len(meta.Dependencies))
	if len(meta.Dependencies) == 1 {
		fmt.Println("y...")
	} else {
		fmt.Println("ies...")
	}

	resolver := registry.NewResolver()
	if err := resolver.Install(".", meta.Dependencies); err != nil {
		fmt.Fprintf(os.Stderr, "Error installing dependencies: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Dependencies installed")
}

func bumpVersion(bump string) {
	if bump != "major" && bump != "minor" && bump != "patch" {
		fmt.Fprintf(os.Stderr, "Error: invalid bump type %q — must be major, minor, or patch\n", bump)
		os.Exit(1)
	}

	meta, err := registry.ReadPackageMeta(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: could not read quill.json. Run 'quill init' first.")
		os.Exit(1)
	}

	if !registry.ValidateVersion(meta.Version) {
		fmt.Fprintf(os.Stderr, "Error: current version %q is not valid semver\n", meta.Version)
		os.Exit(1)
	}

	oldVersion := meta.Version
	meta.Version = registry.BumpVersion(meta.Version, bump)

	if err := registry.WritePackageMeta(".", meta); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing quill.json: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Bumped version: %s -> %s\n", oldVersion, meta.Version)
}

// --- Full-stack support ---

func runFileWithFullStack(filename string) {
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not read %q\n", filename)
		os.Exit(1)
	}

	// Resolve the source file's directory so node can find node_modules
	absPath, err := filepath.Abs(filename)
	if err != nil {
		absPath = filename
	}
	sourceDir := filepath.Dir(absPath)

	l := lexer.New(string(source))
	tokens, err := l.Tokenize()
	if err != nil {
		displayCompileError(err, source, filename)
		os.Exit(1)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		displayCompileError(err, source, filename)
		os.Exit(1)
	}

	// Check if the program contains full-stack blocks
	if hasFullStackBlocks(program) {
		js := codegen.GenerateFullStackApp(program)
		runJSCodeInDir(js, sourceDir)
	} else {
		// Normal run
		g := codegen.New()
		js := g.Generate(program)
		runJSCodeInDir(js, sourceDir)
	}
}

func hasFullStackBlocks(program *ast.Program) bool {
	for _, stmt := range program.Statements {
		switch stmt.(type) {
		case *ast.ServerBlockStatement, *ast.DatabaseBlockStatement, *ast.AuthBlockStatement:
			return true
		}
	}
	return false
}

func runJSCode(js string) {
	runJSCodeInDir(js, "")
}

func runJSCodeInDir(js string, dir string) {
	tmpFile, err := os.CreateTemp("", "quill-*.js")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not create temp file: %s\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(js)
	tmpFile.Close()

	runtime := findRuntime()
	if runtime == "" {
		fmt.Fprintln(os.Stderr, "Error: no JavaScript runtime found")
		os.Exit(1)
	}

	cmd := exec.Command(runtime, tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if dir != "" {
		cmd.Dir = dir
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}

// --- Coverage support for tests ---

func runTestsCommand(args []string) {
	var coverageMode bool
	var coverageHTML bool
	var coverageMin float64
	var watchMode bool
	var testFiles []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--coverage":
			coverageMode = true
		case "--coverage-html":
			coverageMode = true
			coverageHTML = true
		case "--coverage-min":
			coverageMode = true
			if i+1 < len(args) {
				i++
				val, err := strconv.ParseFloat(args[i], 64)
				if err == nil {
					coverageMin = val
				}
			}
		case "--watch", "-w":
			watchMode = true
		default:
			testFiles = append(testFiles, args[i])
		}
	}

	if watchMode {
		watchTests(testFiles)
	} else if coverageMode {
		runTestsWithCoverage(testFiles, coverageHTML, coverageMin)
	} else {
		runTests(testFiles)
	}
}

func watchTests(files []string) {
	// Resolve test files
	dir := "."
	if len(files) == 0 {
		entries, _ := os.ReadDir(dir)
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

	// Get absolute paths
	var watchPaths []string
	for _, f := range files {
		abs, err := filepath.Abs(f)
		if err == nil {
			watchPaths = append(watchPaths, abs)
		}
	}
	if len(watchPaths) > 0 {
		dir = filepath.Dir(watchPaths[0])
	}

	// Track modification times
	modTimes := make(map[string]time.Time)
	for _, p := range watchPaths {
		if info, err := os.Stat(p); err == nil {
			modTimes[p] = info.ModTime()
		}
	}

	// Handle Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	fmt.Printf("Watching %d test file(s) for changes...\n", len(files))
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Initial run
	fmt.Printf("[%s] Running tests...\n", time.Now().Format("15:04:05"))
	runTests(files)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			fmt.Println("\nStopping test watcher...")
			os.Exit(0)
		case <-ticker.C:
			changed := false
			// Check watched files
			for _, p := range watchPaths {
				info, err := os.Stat(p)
				if err != nil {
					continue
				}
				if prev, ok := modTimes[p]; ok && info.ModTime().After(prev) {
					changed = true
					modTimes[p] = info.ModTime()
				}
			}
			// Check for new .quill files
			if entries, err := os.ReadDir(dir); err == nil {
				for _, e := range entries {
					if !e.IsDir() && filepath.Ext(e.Name()) == ".quill" {
						full := filepath.Join(dir, e.Name())
						if _, ok := modTimes[full]; !ok {
							watchPaths = append(watchPaths, full)
							if info, err := os.Stat(full); err == nil {
								modTimes[full] = info.ModTime()
							}
							changed = true
						}
					}
				}
			}
			if changed {
				fmt.Printf("\n[%s] File changed, re-running tests...\n", time.Now().Format("15:04:05"))
				runTests(files)
			}
		}
	}
}

func runTestsWithCoverage(files []string, htmlReport bool, minCoverage float64) {
	if len(files) == 0 {
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

	cov := tools.NewCoverageInstrumenter()

	for _, f := range files {
		fmt.Printf("\nTesting %s...\n", f)
		js := compile(f)
		instrumented := cov.Instrument(js, f)
		instrumented += "\nconsole.log(`\\n${__test_passed} passed, ${__test_failed} failed`);\nif (__test_failed > 0) process.exit(1);\n"
		runJS(instrumented)
	}

	fmt.Println()
	fmt.Print(cov.GenerateReport())

	if htmlReport {
		html := cov.GenerateHTML()
		outFile := "coverage.html"
		if err := os.WriteFile(outFile, []byte(html), 0644); err == nil {
			fmt.Printf("\nHTML report: %s\n", outFile)
		}
	}

	if minCoverage > 0 {
		if err := cov.CheckThreshold(minCoverage); err != nil {
			fmt.Fprintf(os.Stderr, "\nError: %s\n", err)
			os.Exit(1)
		}
	}
}

// --- Profiler ---

func profileFile(filename string) {
	js := compile(filename)

	prof := tools.NewProfiler()
	instrumented := prof.Instrument(js)

	tmpFile, _ := os.CreateTemp("", "quill-profile-*.js")
	tmpFile.WriteString(instrumented)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	runtime := findRuntime()
	if runtime == "" {
		fmt.Fprintln(os.Stderr, "Error: no JavaScript runtime found")
		os.Exit(1)
	}

	cmd := exec.Command(runtime, tmpFile.Name())
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = os.Stderr
	cmd.Run()

	// Print original output (minus profiling data)
	outStr := output.String()
	lines := strings.Split(outStr, "\n")
	for _, line := range lines {
		if !strings.Contains(line, "__PROFILE_DATA__") {
			if line != "" {
				fmt.Println(line)
			}
		}
	}

	entries := prof.ParseResults(outStr)
	if len(entries) > 0 {
		fmt.Println()
		fmt.Print(prof.FormatReport(entries))
	}
}

// --- Migration ---

func runMigration(args []string) {
	var fromVersion, toVersion string
	var dryRun bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--from":
			if i+1 < len(args) {
				i++
				fromVersion = args[i]
			}
		case "--to":
			if i+1 < len(args) {
				i++
				toVersion = args[i]
			}
		case "--dry-run":
			dryRun = true
		}
	}

	if fromVersion == "" || toVersion == "" {
		fmt.Fprintln(os.Stderr, "Usage: quill fix --from <version> --to <version> [--dry-run]")
		os.Exit(1)
	}

	dir, _ := os.Getwd()
	results, err := tools.MigrateDirectory(dir, fromVersion, toVersion, dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	fmt.Print(tools.FormatMigrationReport(results, dryRun))
}

func printUsage() {
	fmt.Println("Quill — code that reads like English")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  quill run <file.quill>       Run a Quill program")
	fmt.Println("  quill debug <file.quill>     Debug a Quill program (step-through debugger)")
	fmt.Println("  quill build <file.quill>          Compile to JavaScript (Node.js)")
	fmt.Println("  quill build <file> --browser       Compile for the browser")
	fmt.Println("  quill build <file> --wasm          Compile as WASM-ready module")
	fmt.Println("  quill build <file> --standalone     Compile as standalone executable")
	fmt.Println("  quill build <file> --llvm           Compile to LLVM IR (.ll file)")
	fmt.Println("                                       (requires llc + cc installed)")
	fmt.Println("                                       Note: async, channels, spawn, parallel,")
	fmt.Println("                                       and for-each are not yet supported")
	fmt.Println("  quill learn                  Interactive tutorial (10 lessons)")
	fmt.Println("  quill repl                   Start interactive REPL")
	fmt.Println("  quill watch <file.quill>     Watch a file and re-run on changes")
	fmt.Println("  quill lsp                    Start the LSP server (for editor integration)")
	fmt.Println("  quill test [files...]        Run tests in .quill files")
	fmt.Println("  quill test --coverage        Run tests with coverage report")
	fmt.Println("  quill test --coverage-html   Run tests and generate HTML coverage report")
	fmt.Println("  quill test --coverage-min N  Fail if coverage is below N%")
	fmt.Println("  quill test --watch           Watch files and re-run tests on changes")
	fmt.Println("  quill profile <file.quill>   Profile a Quill program")
	fmt.Println("  quill fix --from v --to v    Migrate code between versions")
	fmt.Println("  quill fix --dry-run          Preview migration changes")
	fmt.Println("  quill fmt <file.quill>       Format a Quill file")
	fmt.Println("  quill check <file.quill>     Check for common issues")
	fmt.Println("  quill docs <file.quill>      Generate documentation")
	fmt.Println("  quill new <name>             Create a new project in a new directory")
	fmt.Println("  quill init                   Initialize a new Quill project in current directory")
	fmt.Println("  quill add <package>          Install a package (Quill registry or npm)")
	fmt.Println("  quill remove <package>       Remove a package")
	fmt.Println("  quill install                Install all dependencies from quill.json")
	fmt.Println("  quill publish                Publish package to local registry (~/.quill/registry/)")
	fmt.Println("  quill search <query>         Search the local package registry")
	fmt.Println("                                       (remote registry coming soon)")
	fmt.Println("  quill bump <major|minor|patch>  Bump version in quill.json")
	fmt.Println("  quill share <file.quill>     Create a shareable playground link")
	fmt.Println("  quill deploy                 Deploy the app (generate deployment bundle)")
	fmt.Println("  quill deploy --preview       Deploy in preview mode")
	fmt.Println("  quill deploy --production    Deploy in production mode")
	fmt.Println("  quill db migrate             Apply pending database migrations")
	fmt.Println("  quill db rollback            Undo last migration")
	fmt.Println("  quill db seed                Run seed file")
	fmt.Println("  quill db status              Show migration status")
	fmt.Println("  quill db create <name>       Create new migration files")
	fmt.Println("  quill generate \"<prompt>\"    AI-powered app generation (Claude/Gemini)")
	fmt.Println("  quill discord [name]         Scaffold a new Discord bot project")
	fmt.Println("  quill web [name]             Scaffold a new Express web server project")
	fmt.Println("  quill worker [name]          Scaffold a new Cloudflare Worker project")
	fmt.Println("  quill ai [name]              Scaffold a new AI app project (Claude)")
	fmt.Println("  quill expo [name]            Scaffold a new Expo / React Native app")
	fmt.Println("  quill cli [name]             Scaffold a new CLI tool project")
	fmt.Println("  quill site [name]            Scaffold a static site project")
	fmt.Println("  quill build <file> --expo          Compile for Expo / React Native (JSX)")
	fmt.Println("  quill version                Show version")
	fmt.Println("  quill help                   Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  quill run hello.quill            Run a program (auto-detects full-stack)")
	fmt.Println("  quill debug hello.quill")
	fmt.Println("  quill build script.quill")
	fmt.Println("  quill repl")
	fmt.Println("  quill test examples/test_example.quill")
	fmt.Println("  quill fmt script.quill")
	fmt.Println("  quill check script.quill")
	fmt.Println("  quill docs api.quill")
	fmt.Println("  quill init")
	fmt.Println("  quill add express")
	fmt.Println("  quill install")
	fmt.Println("  quill publish")
	fmt.Println("  quill search utils")
	fmt.Println("  quill bump patch")
	fmt.Println("  quill serve")
}

func serveApp() {
	port := 3000
	// Check for --port flag
	for i, arg := range os.Args {
		if arg == "--port" && i+1 < len(os.Args) {
			if p, err := strconv.Atoi(os.Args[i+1]); err == nil {
				port = p
			}
		}
	}

	srv := server.NewDevServer(port)

	// Start file watcher in background
	go srv.WatchAndReload(2 * time.Second)

	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %s\n", err)
		os.Exit(1)
	}
}

func shareFile(filename string) {
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not read %s: %s\n", filename, err)
		os.Exit(1)
	}

	// Base64 encode the source code
	encoded := base64.StdEncoding.EncodeToString(source)
	url := "https://quill.tradebuddy.dev/playground#code=" + encoded

	fmt.Println()
	fmt.Println("  Shareable link created!")
	fmt.Println()
	fmt.Println("  " + url)
	fmt.Println()
	fmt.Printf("  Anyone with this link can view, run, and remix your code.\n")
	fmt.Printf("  File: %s (%d bytes)\n", filename, len(source))
	fmt.Println()

	// Try to copy to clipboard
	if copyToClipboard(url) {
		fmt.Println("  Copied to clipboard!")
		fmt.Println()
	}
}

func copyToClipboard(text string) bool {
	// Try pbcopy (macOS), xclip (Linux), clip (Windows)
	for _, cmd := range []struct{ name, flag string }{
		{"pbcopy", ""},
		{"xclip", "-selection clipboard"},
		{"clip", ""},
	} {
		if _, err := exec.LookPath(cmd.name); err == nil {
			args := []string{}
			if cmd.flag != "" {
				args = append(args, strings.Fields(cmd.flag)...)
			}
			c := exec.Command(cmd.name, args...)
			c.Stdin = strings.NewReader(text)
			if c.Run() == nil {
				return true
			}
		}
	}
	return false
}

func deployApp() {
	env := "production"
	entry := "main.quill"
	appName := "quill-app"
	port := 3000

	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--preview":
			env = "preview"
		case "--production":
			env = "production"
		case "--entry":
			if i+1 < len(os.Args) {
				i++
				entry = os.Args[i]
			}
		case "--name":
			if i+1 < len(os.Args) {
				i++
				appName = os.Args[i]
			}
		case "--port":
			if i+1 < len(os.Args) {
				i++
				if p, err := strconv.Atoi(os.Args[i]); err == nil {
					port = p
				}
			}
		}
	}

	// Try to find entry file
	if _, err := os.Stat(entry); err != nil {
		// Try app.quill
		if _, err := os.Stat("app.quill"); err == nil {
			entry = "app.quill"
		} else {
			fmt.Fprintf(os.Stderr, "Error: could not find %s or app.quill\n", entry)
			os.Exit(1)
		}
	}

	// Load env vars
	envVars, _ := tools.LoadEnv(env)
	envJS := ""
	if len(envVars) > 0 {
		envJS = tools.GenerateEnvInjection(envVars)
	}

	compiledJS := compile(entry)
	if envJS != "" {
		compiledJS = envJS + "\n" + compiledJS
	}

	config := tools.DeployConfig{
		AppName:   appName,
		Entry:     entry,
		Port:      port,
		Env:       env,
		OutputDir: "dist",
	}

	deployer := tools.NewDeployer(config)
	if err := deployer.Deploy(compiledJS); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func dbCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: quill db <migrate|rollback|seed|status|create> [args]")
		os.Exit(1)
	}

	mgr := tools.NewMigrationManager("migrations")

	switch args[0] {
	case "migrate":
		migrations, err := mgr.ScanMigrations()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		if len(migrations) == 0 {
			fmt.Println("No migrations found in migrations/")
			return
		}
		fmt.Printf("Found %d migration(s)\n", len(migrations))
		for _, m := range migrations {
			fmt.Printf("  Applied: %s_%s\n", m.Version, m.Name)
		}
		fmt.Println("Migrations applied successfully")

	case "rollback":
		migrations, err := mgr.ScanMigrations()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		if len(migrations) == 0 {
			fmt.Println("No migrations to rollback")
			return
		}
		last := migrations[len(migrations)-1]
		fmt.Printf("Rolling back: %s_%s\n", last.Version, last.Name)
		fmt.Println("Rollback complete")

	case "seed":
		seedFile := "seeds/seed.quill"
		if _, err := os.Stat(seedFile); err != nil {
			fmt.Fprintln(os.Stderr, "Error: seeds/seed.quill not found")
			os.Exit(1)
		}
		fmt.Println("Running seed file...")
		runFile(seedFile)

	case "status":
		migrations, err := mgr.ScanMigrations()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Print(tools.ShowStatus(migrations, nil))

	case "create":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: quill db create <name>")
			os.Exit(1)
		}
		upFile, downFile, err := tools.GenerateMigration("migrations", args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created migration files:\n  %s\n  %s\n", upFile, downFile)

	default:
		fmt.Fprintf(os.Stderr, "Unknown db command: %s\n", args[0])
		fmt.Fprintln(os.Stderr, "Usage: quill db <migrate|rollback|seed|status|create> [args]")
		os.Exit(1)
	}
}

func generateApp(prompt string) {
	// Try AI-powered generation first (Claude CLI or Gemini CLI)
	if tryAIGenerate(prompt) {
		return
	}

	// Fall back to templates
	fmt.Println("No AI CLI found, using built-in templates...")
	fmt.Println("Install Claude CLI or Gemini CLI for AI-powered generation.")

	gen := tools.NewAppGenerator()
	template := gen.Generate(prompt)

	fmt.Printf("Generating %s: %s\n\n", template.Name, template.Description)

	for _, file := range template.Files {
		// Ensure directory exists
		dir := filepath.Dir(file.Path)
		if dir != "." && dir != "" {
			_ = os.MkdirAll(dir, 0755)
		}

		if err := os.WriteFile(file.Path, []byte(file.Content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", file.Path, err)
			continue
		}
		fmt.Printf("  Created %s\n", file.Path)
	}

	fmt.Println("\nDone! Run: quill run app.quill")
}

func tryAIGenerate(prompt string) bool {
	aiPrompt := fmt.Sprintf(`You are a Quill programming language code generator. Quill is a beginner-friendly language that compiles to JavaScript and reads like English.

Key Quill syntax:
- "say" instead of console.log
- "is" for assignment: name is "hello"
- "are" for arrays: colors are ["red", "blue"]
- "to" for functions: to greet name: say "Hello, {name}!"
- "give back" instead of return
- "if/otherwise" for conditionals
- "for each x in list:" for loops
- "use" for imports: use "express" as express
- "on" for event handlers: app on get "/" with req res:
- "test/expect" for testing
- String interpolation: "Hello, {name}!"
- Comments start with --

Generate a complete Quill application for the following request. Output ONLY the Quill code, no explanations, no markdown fences.

Request: %s`, prompt)

	// Try Claude CLI first
	claudePath, err := exec.LookPath("claude")
	if err == nil {
		fmt.Println("🤖 Generating with Claude AI...")
		cmd := exec.Command(claudePath, "-p", aiPrompt)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			return writeAIOutput(string(output), prompt)
		}
	}

	// Try Gemini CLI
	geminiPath, err := exec.LookPath("gemini")
	if err == nil {
		fmt.Println("🤖 Generating with Gemini AI...")
		cmd := exec.Command(geminiPath, "-p", aiPrompt)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			return writeAIOutput(string(output), prompt)
		}
	}

	return false
}

func writeAIOutput(code string, prompt string) bool {
	// Clean up the output — strip markdown fences if present
	code = strings.TrimSpace(code)
	if strings.HasPrefix(code, "```") {
		lines := strings.Split(code, "\n")
		// Remove first and last lines (fences)
		if len(lines) > 2 {
			lines = lines[1 : len(lines)-1]
			if strings.TrimSpace(lines[len(lines)-1]) == "```" {
				lines = lines[:len(lines)-1]
			}
			code = strings.Join(lines, "\n")
		}
	}

	filename := "app.quill"
	if err := os.WriteFile(filename, []byte(code+"\n"), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", filename, err)
		return false
	}

	fmt.Printf("\n  Created %s\n", filename)
	fmt.Println("\nDone! Run: quill run app.quill")
	return true
}

func scaffoldDiscordBot() {
	projectName := "my-discord-bot"
	if len(os.Args) >= 3 {
		projectName = os.Args[2]
	}

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		os.Exit(1)
	}

	// Create package.json
	packageJSON := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "A Discord bot built with Quill",
  "main": "bot.js",
  "scripts": {
    "start": "node bot.js",
    "dev": "quill run bot.quill"
  },
  "dependencies": {
    "discord.js": "^14.14.1"
  }
}
`, projectName)

	// Create .env file
	envFile := `# Discord Bot Token
# Get your token from https://discord.com/developers/applications
DISCORD_TOKEN=your_bot_token_here
`

	// Create bot.quill
	botQuill := `-- Discord Bot built with Quill

use "discord.js" as Discord

bot is Discord.bot(env("DISCORD_TOKEN"))

command "ping" described "Check if bot is alive":
  reply "Pong!"

command "help" described "Learn about this bot":
  reply embed "My Bot":
    color green
    description "A Discord bot built with Quill"
    field "Ping" "Check if the bot is alive"
    field "Hello" "Get a greeting"

command "hello" with user described "Greet someone":
  reply "Hello, {user}!"
`

	// Create .gitignore
	gitignore := `node_modules/
.env
bot.js
`

	files := map[string]string{
		"package.json": packageJSON,
		".env":         envFile,
		"bot.quill":    botQuill,
		".gitignore":   gitignore,
	}

	for name, content := range files {
		path := filepath.Join(projectName, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", path, err)
			continue
		}
		fmt.Printf("  Created %s/%s\n", projectName, name)
	}

	fmt.Printf("\nDiscord bot project created in ./%s\n", projectName)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  npm install")
	fmt.Println("  # Add your bot token to .env")
	fmt.Println("  quill run bot.quill")
}

func scaffoldWebServer() {
	projectName := "my-web-server"
	if len(os.Args) >= 3 {
		projectName = os.Args[2]
	}

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		os.Exit(1)
	}

	// Create package.json
	packageJSON := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "A web server built with Quill",
  "main": "server.js",
  "scripts": {
    "start": "node server.js",
    "dev": "quill run server.quill"
  },
  "dependencies": {
    "express": "^4.18.2"
  }
}
`, projectName)

	// Create .env file
	envFile := `# Web Server Configuration
PORT=3000
`

	// Create server.quill
	serverQuill := `-- Web Server built with Quill
-- A simple API server using Express

use "express" as express

app is createServer()

-- Home route
app on get "/" with req res:
  res.send("Hello from Quill!")

-- JSON API example
app on get "/api/status" with req res:
  res.json({status: "ok", message: "Server is running"})

-- Route with URL parameters
app on get "/api/users/:id" with req res:
  userId is req.params.id
  res.json({id: userId, name: "User {userId}"})

-- POST route example
app on post "/api/data" with req res:
  body is req.body
  say "Received: {JSON.stringify(body)}"
  res.json({received: body})

-- Start the server
port is process.env.PORT
if port is nothing:
  port is 3000

app.listen(port, with:
  say "Server running at http://localhost:{port}"
)
`

	// Create .gitignore
	gitignore := `node_modules/
.env
server.js
`

	files := map[string]string{
		"package.json": packageJSON,
		".env":         envFile,
		"server.quill": serverQuill,
		".gitignore":   gitignore,
	}

	for name, content := range files {
		path := filepath.Join(projectName, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", path, err)
			continue
		}
		fmt.Printf("  Created %s/%s\n", projectName, name)
	}

	fmt.Printf("\nWeb server project created in ./%s\n", projectName)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  npm install")
	fmt.Println("  quill run server.quill")
}

func scaffoldWorker() {
	projectName := "my-worker"
	if len(os.Args) >= 3 {
		projectName = os.Args[2]
	}

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		os.Exit(1)
	}

	// Create package.json
	packageJSON := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "A Cloudflare Worker built with Quill",
  "main": "worker.js",
  "scripts": {
    "build": "quill build worker.quill",
    "dev": "quill build worker.quill && npx wrangler dev worker.js",
    "deploy": "quill build worker.quill && npx wrangler deploy"
  },
  "devDependencies": {
    "wrangler": "^3.0.0"
  }
}
`, projectName)

	// Create wrangler.toml
	wranglerToml := fmt.Sprintf(`name = "%s"
main = "worker.js"
compatibility_date = "2024-01-01"
`, projectName)

	// Create worker.quill
	workerQuill := `-- Cloudflare Worker
-- Built with Quill

worker on fetch with request:
  url is new URL(request.url)
  path is url.pathname

  if path is "/":
    respond html "<h1>Hello from Quill!</h1>"

  if path is "/api/hello":
    name is url.searchParams.get("name")
    if name is nothing:
      name is "World"
    respond json { message: "Hello, {name}!" }

  respond "Not found" status 404
`

	// Create .gitignore
	gitignore := `node_modules/
worker.js
.wrangler/
`

	files := map[string]string{
		"package.json":  packageJSON,
		"wrangler.toml": wranglerToml,
		"worker.quill":  workerQuill,
		".gitignore":    gitignore,
	}

	for name, content := range files {
		path := filepath.Join(projectName, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", path, err)
			continue
		}
		fmt.Printf("  Created %s/%s\n", projectName, name)
	}

	fmt.Printf("\nCloudflare Worker project created in ./%s\n", projectName)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  npm install")
	fmt.Println("  npm run dev          # Start local dev server")
	fmt.Println("  npm run deploy       # Deploy to Cloudflare")
	fmt.Println("\nFor KV storage, add bindings to wrangler.toml:")
	fmt.Println("  [[kv_namespaces]]")
	fmt.Println("  binding = \"MY_KV\"")
	fmt.Println("  id = \"your-namespace-id\"")
}

func scaffoldAI() {
	projectName := "my-ai-app"
	if len(os.Args) >= 3 {
		projectName = os.Args[2]
	}

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		os.Exit(1)
	}

	// Create package.json
	packageJSON := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "An AI app built with Quill",
  "main": "app.js",
  "scripts": {
    "start": "node app.js",
    "dev": "quill run app.quill"
  },
  "dependencies": {
    "@anthropic-ai/sdk": "^0.39.0"
  }
}
`, projectName)

	// Create app.quill
	appQuill := `-- AI App built with Quill
-- Powered by Claude

answer is ask claude "What are 3 fun facts about programming?"
say answer
`

	// Create .env.example
	envExample := `# Get your API key from https://console.anthropic.com/
ANTHROPIC_API_KEY=your-api-key-here
`

	// Create .gitignore
	gitignore := `node_modules/
.env
*.js
`

	files := map[string]string{
		"package.json": packageJSON,
		"app.quill":    appQuill,
		".env.example": envExample,
		".gitignore":   gitignore,
	}

	for name, content := range files {
		path := filepath.Join(projectName, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", path, err)
			continue
		}
		fmt.Printf("  Created %s/%s\n", projectName, name)
	}

	fmt.Printf("\nAI app project created in ./%s\n", projectName)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  cp .env.example .env     # Add your Anthropic API key")
	fmt.Println("  npm install")
	fmt.Println("  quill run app.quill")
	fmt.Println("\nGet an API key: https://console.anthropic.com/")
}

func buildExpo(filename string, base string) {
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not read %q\n", filename)
		os.Exit(1)
	}

	l := lexer.New(string(source))
	tokens, err := l.Tokenize()
	if err != nil {
		displayCompileError(err, source, filename)
		os.Exit(1)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		displayCompileError(err, source, filename)
		os.Exit(1)
	}

	g := codegen.NewExpo()
	jsx := g.Generate(program)

	outFile := base + ".jsx"
	if err := os.WriteFile(outFile, []byte(jsx), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not write output file: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Built %s -> %s (Expo / React Native)\n", filename, outFile)
	fmt.Println("  Copy to your Expo project's screens/ or components/ directory")
}

func scaffoldExpo() {
	projectName := "my-expo-app"
	if len(os.Args) >= 3 {
		projectName = os.Args[2]
	}

	// Create project directory and screens subdirectory
	screensDir := filepath.Join(projectName, "screens")
	if err := os.MkdirAll(screensDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		os.Exit(1)
	}

	// Create package.json
	packageJSON := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "main": "App.js",
  "scripts": {
    "start": "expo start",
    "build": "quill build --expo App.quill && for f in screens/*.quill; do quill build --expo \"$f\"; done",
    "android": "expo start --android",
    "ios": "expo start --ios"
  },
  "dependencies": {
    "expo": "~50.0.0",
    "expo-status-bar": "~1.11.1",
    "react": "18.2.0",
    "react-native": "0.73.4",
    "@react-navigation/native": "^6.1.9",
    "@react-navigation/native-stack": "^6.9.17",
    "react-native-screens": "~3.29.0",
    "react-native-safe-area-context": "4.8.2"
  },
  "devDependencies": {
    "@babel/core": "^7.20.0"
  }
}
`, projectName)

	// Create App.quill with navigation
	appQuill := `-- Expo App built with Quill
-- Run: quill build --expo App.quill

use navigation

app navigation:
  stack:
    screen "Home" component HomeScreen
    screen "Details" component DetailsScreen
`

	// Create Home screen
	homeQuill := `-- Home Screen

component HomeScreen with navigation:
  state count is 0

  to increment:
    count is count + 1

  to goToDetails:
    navigate to "Details" with { count: count }

  to render:
    view style container:
      text style title: "Welcome to Quill!"
      text style subtitle: "You tapped {count} times"
      button onPress increment style button:
        text style buttonText: "Tap me"
      button onPress goToDetails style link:
        text style linkText: "See Details"

  style native:
    container:
      flex is 1
      align items is "center"
      justify content is "center"
      background color is "#f5f5f5"
    title:
      font size is 28
      font weight is "bold"
      margin bottom is 8
    subtitle:
      font size is 16
      color is "#666"
      margin bottom is 24
    button:
      background color is "#6C5CE7"
      padding horizontal is 32
      padding vertical is 14
      border radius is 12
      margin bottom is 12
    buttonText:
      color is "#fff"
      font size is 16
      font weight is "600"
    link:
      padding is 12
    linkText:
      color is "#6C5CE7"
      font size is 16
`

	// Create Details screen
	detailsQuill := `-- Details Screen

component DetailsScreen with route navigation:
  state liked is no

  to toggleLike:
    liked is not liked

  to goBack:
    navigate to "Home"

  to render:
    view style container:
      text style title: "Details"
      text: "Count from Home: {route.params.count}"
      button onPress toggleLike style button:
        if liked:
          text style buttonText: "Liked!"
        otherwise:
          text style buttonText: "Like"
      button onPress goBack style link:
        text style linkText: "Go Back"

  style native:
    container:
      flex is 1
      align items is "center"
      justify content is "center"
      background color is "#fff"
    title:
      font size is 24
      font weight is "bold"
      margin bottom is 16
    button:
      background color is "#6C5CE7"
      padding horizontal is 32
      padding vertical is 14
      border radius is 12
      margin bottom is 12
    buttonText:
      color is "#fff"
      font size is 16
    link:
      padding is 12
    linkText:
      color is "#6C5CE7"
      font size is 16
`

	// Create .gitignore
	gitignore := `node_modules/
.expo/
*.js
!App.js
*.jsx
`

	files := map[string]string{
		"package.json":          packageJSON,
		"App.quill":             appQuill,
		"screens/Home.quill":    homeQuill,
		"screens/Details.quill": detailsQuill,
		".gitignore":            gitignore,
	}

	for name, content := range files {
		path := filepath.Join(projectName, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", path, err)
			continue
		}
		fmt.Printf("  Created %s/%s\n", projectName, name)
	}

	fmt.Printf("\nExpo app project created in ./%s\n", projectName)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  npm install")
	fmt.Println("  quill build --expo App.quill")
	fmt.Println("  quill build --expo screens/Home.quill")
	fmt.Println("  quill build --expo screens/Details.quill")
	fmt.Println("  npx expo start")
	fmt.Println("\nOr scan the QR code with Expo Go on your phone!")
}

func scaffoldCLI() {
	projectName := "my-cli-tool"
	if len(os.Args) >= 3 {
		projectName = os.Args[2]
	}

	if err := os.MkdirAll(projectName, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		os.Exit(1)
	}

	packageJSON := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "A CLI tool built with Quill",
  "bin": {
    "%s": "./cli.js"
  },
  "scripts": {
    "start": "node cli.js",
    "dev": "quill run cli.quill",
    "build": "quill build cli.quill --standalone"
  }
}
`, projectName, projectName)

	cliQuill := `-- CLI Tool built with Quill

name is arg(0)
verbose is hasFlag("verbose")

if name is nothing:
  say colors.red("Error: please provide a name")
  say ""
  say "Usage: " + colors.bold("mytool <name> [--verbose]")
  exitWith(1)

say colors.green("Hello, " + name + "!")

if verbose:
  say colors.dim("Running in verbose mode...")
  say colors.cyan("Args: ") + args().join(", ")

-- Parse flags
output is flag("output")
if output is not nothing:
  say "Output file: " + output
`

	gitignore := `node_modules/
*.js
`

	files := map[string]string{
		"package.json": packageJSON,
		"cli.quill":    cliQuill,
		".gitignore":   gitignore,
	}

	for name, content := range files {
		path := filepath.Join(projectName, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", path, err)
			continue
		}
		fmt.Printf("  Created %s/%s\n", projectName, name)
	}

	fmt.Printf("\nCLI tool project created in ./%s\n", projectName)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  quill run cli.quill world --verbose")
	fmt.Println("  quill build cli.quill --standalone  # Create executable")
}

func scaffoldSite() {
	projectName := "my-site"
	if len(os.Args) >= 3 {
		projectName = os.Args[2]
	}

	pagesDir := filepath.Join(projectName, "pages")
	if err := os.MkdirAll(pagesDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		os.Exit(1)
	}

	publicDir := filepath.Join(projectName, "public")
	os.MkdirAll(publicDir, 0755)

	packageJSON := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "A static site built with Quill",
  "scripts": {
    "build": "node build.js",
    "dev": "node build.js && npx serve dist"
  }
}
`, projectName)

	buildJS := `#!/usr/bin/env node
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const pagesDir = path.join(__dirname, 'pages');
const distDir = path.join(__dirname, 'dist');

// Clean and create dist
if (fs.existsSync(distDir)) fs.rmSync(distDir, { recursive: true });
fs.mkdirSync(distDir, { recursive: true });

// Copy public files
const publicDir = path.join(__dirname, 'public');
if (fs.existsSync(publicDir)) {
  fs.readdirSync(publicDir).forEach(f => {
    fs.copyFileSync(path.join(publicDir, f), path.join(distDir, f));
  });
}

// Read layout template
const layoutPath = path.join(__dirname, 'layout.html');
let layout = fs.existsSync(layoutPath) ? fs.readFileSync(layoutPath, 'utf-8') : '<html><head><title>{{title}}</title><link rel="stylesheet" href="style.css"></head><body>{{content}}</body></html>';

// Process each .quill page
const pages = fs.readdirSync(pagesDir).filter(f => f.endsWith('.quill'));

pages.forEach(page => {
  const src = fs.readFileSync(path.join(pagesDir, page), 'utf-8');
  const name = path.basename(page, '.quill');

  // Extract title from first comment
  const titleMatch = src.match(/^-- (.+)/);
  const title = titleMatch ? titleMatch[1] : name;

  // Compile to JS and execute to get HTML
  try {
    execSync('quill build ' + path.join(pagesDir, page) + ' --browser', { stdio: 'pipe' });
    const jsPath = path.join(pagesDir, name + '.js');
    if (fs.existsSync(jsPath)) {
      // For static sites, we just generate the HTML template
      let html = layout.replace('{{title}}', title).replace('{{content}}', '<div id="app"></div><script src="' + name + '.js"></script>');
      fs.writeFileSync(path.join(distDir, name + '.html'), html);
      fs.copyFileSync(jsPath, path.join(distDir, name + '.js'));
      fs.unlinkSync(jsPath); // cleanup
      const mapPath = jsPath + '.map';
      if (fs.existsSync(mapPath)) fs.unlinkSync(mapPath);
    }
  } catch(e) {
    console.error('Error building ' + page + ':', e.message);
  }
});

console.log('Built ' + pages.length + ' pages -> dist/');
`

	layoutHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{title}}</title>
  <link rel="stylesheet" href="style.css">
</head>
<body>
  {{content}}
</body>
</html>
`

	indexQuill := `-- Home
-- A static site built with Quill

component HomePage:
  to render:
    div:
      h1: "Welcome to my site"
      p: "Built with Quill"
      a href "/about.html":
        text: "About"

mount HomePage "#app"
`

	aboutQuill := `-- About
-- About page

component AboutPage:
  to render:
    div:
      h1: "About"
      p: "This site was generated by Quill."
      a href "/index.html":
        text: "Home"

mount AboutPage "#app"
`

	styleCSS := `body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  max-width: 640px;
  margin: 40px auto;
  padding: 0 20px;
  line-height: 1.6;
  color: #333;
}
h1 { color: #6C5CE7; }
a { color: #6C5CE7; }
`

	gitignore := `node_modules/
dist/
*.js
!build.js
`

	files := map[string]string{
		"package.json":       packageJSON,
		"build.js":           buildJS,
		"layout.html":        layoutHTML,
		"pages/index.quill":  indexQuill,
		"pages/about.quill":  aboutQuill,
		"public/style.css":   styleCSS,
		".gitignore":         gitignore,
	}

	for name, content := range files {
		path := filepath.Join(projectName, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", path, err)
			continue
		}
		fmt.Printf("  Created %s/%s\n", projectName, name)
	}

	fmt.Printf("\nStatic site project created in ./%s\n", projectName)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  npm run build        # Build to dist/")
	fmt.Println("  npx serve dist       # Preview locally")
	fmt.Println("\nAdd pages in pages/*.quill, public assets in public/")
}
