package main

import (
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
)

const version = "0.1.0"

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
			fmt.Fprintln(os.Stderr, "Usage: quill build <file.quill>")
			os.Exit(1)
		}
		buildFile(os.Args[2])

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

func buildFile(filename string) {
	js := compile(filename)

	// Output .js file next to the source
	ext := filepath.Ext(filename)
	outFile := filename[:len(filename)-len(ext)] + ".js"

	if err := os.WriteFile(outFile, []byte(js), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not write output file: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Built %s -> %s\n", filename, outFile)
}

func compile(filename string) string {
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
	g := codegen.New()
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

	a := analyzer.New()
	diagnostics := a.Analyze(program)

	if len(diagnostics) == 0 {
		fmt.Printf("✓ %s — no issues found\n", filename)
		return
	}

	fmt.Printf("Found %d issue(s) in %s:\n\n", len(diagnostics), filename)
	for _, d := range diagnostics {
		fmt.Println(d.String())
	}
	fmt.Println()

	if analyzer.HasErrors(diagnostics) {
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Quill — a programming language for humans")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  quill run <file.quill>      Run a Quill program")
	fmt.Println("  quill build <file.quill>    Compile to JavaScript")
	fmt.Println("  quill repl                  Start interactive REPL")
	fmt.Println("  quill test [files...]       Run tests in .quill files")
	fmt.Println("  quill fmt <file.quill>      Format a Quill file")
	fmt.Println("  quill check <file.quill>    Check for common issues")
	fmt.Println("  quill version               Show version")
	fmt.Println("  quill help                  Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  quill run hello.quill")
	fmt.Println("  quill build script.quill")
	fmt.Println("  quill repl")
	fmt.Println("  quill test examples/test_example.quill")
	fmt.Println("  quill fmt script.quill")
	fmt.Println("  quill check script.quill")
}
