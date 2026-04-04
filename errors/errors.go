package errors

import (
	"fmt"
	"strings"
)

type QuillError struct {
	Line    int
	Column  int
	Message string
	Hint    string
	Source  string // the full source code
}

func (e *QuillError) Error() string {
	var out strings.Builder
	lines := strings.Split(e.Source, "\n")

	out.WriteString(fmt.Sprintf("\n  Error on line %d: %s\n\n", e.Line, e.Message))

	// Show the line with context
	if e.Line > 0 && e.Line <= len(lines) {
		if e.Line > 1 {
			out.WriteString(fmt.Sprintf("    %d | %s\n", e.Line-1, lines[e.Line-2]))
		}
		out.WriteString(fmt.Sprintf("  > %d | %s\n", e.Line, lines[e.Line-1]))
		if e.Column > 0 {
			padding := strings.Repeat(" ", e.Column+5)
			out.WriteString(fmt.Sprintf("    %s^ here\n", padding))
		}
		if e.Line < len(lines) {
			out.WriteString(fmt.Sprintf("    %d | %s\n", e.Line+1, lines[e.Line]))
		}
	}

	if e.Hint != "" {
		out.WriteString(fmt.Sprintf("\n  Hint: %s\n", e.Hint))
	}

	return out.String()
}

func NewError(line, col int, source, message, hint string) *QuillError {
	return &QuillError{Line: line, Column: col, Source: source, Message: message, Hint: hint}
}
