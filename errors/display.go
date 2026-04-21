package errors

import (
	"fmt"
	"strings"
)

// DisplayError represents a rich error for display purposes.
type DisplayError struct {
	Line    int
	Column  int
	Message string
	Hint    string
	Code    string // e.g., "E001"
}

// Color constants for terminal output.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// max0 returns n if n >= 0, else 0. Used to guard strings.Repeat counts
// against negative values when line numbers exceed the padding width or
// columns end up out of bounds.
func max0(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

// FormatError formats a single error with source context.
func FormatError(err DisplayError, sourceLines []string, filename string, useColor bool) string {
	var out strings.Builder

	red := ""
	cyan := ""
	gray := ""
	bold := ""
	reset := ""
	if useColor {
		red = colorRed
		cyan = colorCyan
		gray = colorGray
		bold = colorBold
		reset = colorReset
	}

	// Header: Error[E001]: message
	out.WriteString(fmt.Sprintf("%s%sError[%s]%s: %s\n", red, bold, err.Code, reset, err.Message))

	// Location: --> file:line:col
	col := err.Column
	if col <= 0 {
		col = 1
	}
	out.WriteString(fmt.Sprintf("  %s-->%s %s:%d:%d\n", cyan, reset, filename, err.Line, col))

	// Source context
	if err.Line > 0 && err.Line <= len(sourceLines) {
		out.WriteString(fmt.Sprintf("   %s|%s\n", cyan, reset))

		// Show up to 3 lines before
		startLine := err.Line - 2
		if startLine < 1 {
			startLine = 1
		}

		for i := startLine; i < err.Line; i++ {
			lineNum := fmt.Sprintf("%d", i)
			padding := strings.Repeat(" ", max0(3-len(lineNum)))
			out.WriteString(fmt.Sprintf("%s%s%s |%s %s\n", gray, padding, lineNum, reset, sourceLines[i-1]))
		}

		// Show the error line with highlight
		lineNum := fmt.Sprintf("%d", err.Line)
		padding := strings.Repeat(" ", max0(3-len(lineNum)))
		out.WriteString(fmt.Sprintf("%s%s%s |%s %s\n", red, padding, lineNum, reset, sourceLines[err.Line-1]))

		// Show caret (clamp to line length so pathological column values
		// don't blow up strings.Repeat or produce absurd padding).
		caretPos := col - 1
		if caretPos < 0 {
			caretPos = 0
		}
		if caretPos > len(sourceLines[err.Line-1]) {
			caretPos = len(sourceLines[err.Line-1])
		}
		caretPadding := strings.Repeat(" ", caretPos)
		out.WriteString(fmt.Sprintf("   %s|%s %s%s%s^^^%s\n", cyan, reset, caretPadding, red, bold, reset))

		// Show up to 1 line after
		if err.Line < len(sourceLines) {
			nextLineNum := fmt.Sprintf("%d", err.Line+1)
			nextPadding := strings.Repeat(" ", max0(3-len(nextLineNum)))
			out.WriteString(fmt.Sprintf("%s%s%s |%s %s\n", gray, nextPadding, nextLineNum, reset, sourceLines[err.Line]))
		}

		out.WriteString(fmt.Sprintf("   %s|%s\n", cyan, reset))
	}

	// Hint
	if err.Hint != "" {
		lines := strings.Split(err.Hint, "\n")
		out.WriteString(fmt.Sprintf("   %s= hint:%s %s\n", cyan, reset, lines[0]))
		for _, line := range lines[1:] {
			out.WriteString(fmt.Sprintf("   %s=%s %s\n", cyan, reset, line))
		}
	}

	return out.String()
}

// FormatErrors formats multiple errors with a summary line.
func FormatErrors(errs []DisplayError, sourceLines []string, filename string, useColor bool) string {
	if len(errs) == 0 {
		return ""
	}

	var out strings.Builder

	for i, err := range errs {
		out.WriteString(FormatError(err, sourceLines, filename, useColor))
		if i < len(errs)-1 {
			out.WriteString("\n")
		}
	}

	// Summary
	red := ""
	yellow := ""
	reset := ""
	if useColor {
		red = colorRed
		yellow = colorYellow
		reset = colorReset
	}

	errorCount := 0
	warningCount := 0
	for _, err := range errs {
		if strings.HasPrefix(err.Code, "W") {
			warningCount++
		} else {
			errorCount++
		}
	}

	out.WriteString("\n")
	parts := []string{}
	if errorCount > 0 {
		word := "error"
		if errorCount != 1 {
			word = "errors"
		}
		parts = append(parts, fmt.Sprintf("%s%d %s%s", red, errorCount, word, reset))
	}
	if warningCount > 0 {
		word := "warning"
		if warningCount != 1 {
			word = "warnings"
		}
		parts = append(parts, fmt.Sprintf("%s%d %s%s", yellow, warningCount, word, reset))
	}
	out.WriteString(fmt.Sprintf("Found %s\n", strings.Join(parts, " and ")))

	return out.String()
}
