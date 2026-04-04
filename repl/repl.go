package repl

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"quill/codegen"
	"quill/lexer"
	"quill/parser"
)

func Start() {
	fmt.Println("Quill REPL v0.1.0")
	fmt.Println("Type your code. Use 'exit' or Ctrl+C to quit.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	var lines []string

	for {
		if isInBlock(lines) {
			fmt.Print("...   ")
		} else {
			fmt.Print("quill> ")
		}

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()

		if line == "exit" || line == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		if line == "" && !isInBlock(lines) {
			continue
		}

		lines = append(lines, line)

		// If line ends with : we're starting a block, continue
		trimmed := strings.TrimSpace(line)
		if strings.HasSuffix(trimmed, ":") {
			continue
		}

		// If we're in a block and line is indented, continue
		if isInBlock(lines) && (strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")) {
			continue
		}

		// Try to compile and run
		source := strings.Join(lines, "\n")
		err := evalSource(source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s\n", err)
			// Remove the last batch of lines
			lines = lines[:len(lines)-1]
		}
	}
}

func isInBlock(lines []string) bool {
	if len(lines) == 0 {
		return false
	}
	last := strings.TrimSpace(lines[len(lines)-1])
	return strings.HasSuffix(last, ":") || strings.HasPrefix(lines[len(lines)-1], "  ") || strings.HasPrefix(lines[len(lines)-1], "\t")
}

func evalSource(source string) error {
	l := lexer.New(source)
	tokens, err := l.Tokenize()
	if err != nil {
		return err
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return err
	}

	g := codegen.New()
	js := g.Generate(program)

	// Run with node
	cmd := exec.Command("node", "-e", js)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
