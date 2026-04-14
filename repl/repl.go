package repl

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"quill/codegen"
	"quill/lexer"
	"quill/parser"
)

// Version is set by the caller (main.go) to keep the version in one place.
var Version = "0.10.1"

func findRuntime() string {
	for _, name := range []string{"node", "bun", "deno"} {
		if _, err := exec.LookPath(name); err == nil {
			return name
		}
	}
	return ""
}

func Start() {
	fmt.Printf("Quill REPL v%s\n", Version)
	fmt.Println("Type your code. Use 'exit' or Ctrl+C to quit.")
	fmt.Println("Commands: :help, :reset, :vars")
	fmt.Println()

	rl := newReadline()
	defer rl.close()

	var lines []string
	var prevOutputLines int

	for {
		prompt := "quill> "
		if isInBlock(lines) {
			prompt = "...   "
		}

		line, ok := rl.readLine(prompt)
		if !ok {
			break
		}

		// Handle commands
		trimmed := strings.TrimSpace(line)
		if trimmed == "exit" || trimmed == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		if trimmed == ":help" {
			fmt.Println("  :reset  — Clear all variables and start fresh")
			fmt.Println("  :vars   — Show defined variables and functions")
			fmt.Println("  :help   — Show this help")
			fmt.Println("  exit    — Quit the REPL")
			continue
		}

		if trimmed == ":reset" {
			lines = nil
			prevOutputLines = 0
			fmt.Println("  (reset)")
			continue
		}

		if trimmed == ":vars" {
			printVars(lines)
			continue
		}

		if trimmed == "" && !isInBlock(lines) {
			continue
		}

		lines = append(lines, line)

		// If line ends with : we're starting a block, continue
		if strings.HasSuffix(trimmed, ":") {
			continue
		}

		// If we're in a block and line is indented, continue
		if isInBlock(lines) && (strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")) {
			continue
		}

		// Try to compile and run, only showing new output
		source := strings.Join(lines, "\n")
		output, err := evalSource(source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s\n", err)
			// Remove the last line(s) that caused the error
			lines = lines[:len(lines)-1]
		} else {
			outputLines := strings.Split(output, "\n")
			// Remove trailing empty line from split
			if len(outputLines) > 0 && outputLines[len(outputLines)-1] == "" {
				outputLines = outputLines[:len(outputLines)-1]
			}
			// Only print lines we haven't seen before
			for i := prevOutputLines; i < len(outputLines); i++ {
				fmt.Println(outputLines[i])
			}
			prevOutputLines = len(outputLines)
		}
	}
}

func printVars(lines []string) {
	if len(lines) == 0 {
		fmt.Println("  (no definitions)")
		return
	}
	found := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Variable assignments: "name is value" or "things are value"
		if strings.Contains(trimmed, " is ") && !strings.HasPrefix(trimmed, "if ") && !strings.HasPrefix(trimmed, "otherwise ") {
			parts := strings.SplitN(trimmed, " is ", 2)
			if len(parts) == 2 && !strings.ContainsAny(parts[0], " \t") {
				fmt.Printf("  %s = %s\n", parts[0], parts[1])
				found = true
			}
		} else if strings.Contains(trimmed, " are ") {
			parts := strings.SplitN(trimmed, " are ", 2)
			if len(parts) == 2 && !strings.ContainsAny(parts[0], " \t") {
				fmt.Printf("  %s = %s\n", parts[0], parts[1])
				found = true
			}
		}
		// Function definitions: "to funcName ..."
		if strings.HasPrefix(trimmed, "to ") {
			name := strings.Fields(trimmed)[1]
			fmt.Printf("  fn %s\n", name)
			found = true
		}
	}
	if !found {
		fmt.Println("  (no definitions)")
	}
}

func isInBlock(lines []string) bool {
	if len(lines) == 0 {
		return false
	}
	last := strings.TrimSpace(lines[len(lines)-1])
	return strings.HasSuffix(last, ":") || strings.HasPrefix(lines[len(lines)-1], "  ") || strings.HasPrefix(lines[len(lines)-1], "\t")
}

func evalSource(source string) (string, error) {
	l := lexer.New(source)
	tokens, err := l.Tokenize()
	if err != nil {
		return "", err
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return "", err
	}

	g := codegen.New()
	js := g.Generate(program)

	// Run with available JS runtime
	runtime := findRuntime()
	if runtime == "" {
		return "", fmt.Errorf("no JavaScript runtime found (install Node.js, Bun, or Deno)")
	}
	cmd := exec.Command(runtime, "-e", js)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return "", fmt.Errorf("%s", errMsg)
		}
		return "", err
	}

	return stdout.String(), nil
}
