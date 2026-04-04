package tools

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// CoverageInstrumenter instruments compiled JavaScript with line counters
// and generates coverage reports after test execution.
type CoverageInstrumenter struct {
	counters map[string]map[int]bool // file -> line -> covered
	stmts    map[string]int          // file -> total statement count
}

// NewCoverageInstrumenter creates a new CoverageInstrumenter.
func NewCoverageInstrumenter() *CoverageInstrumenter {
	return &CoverageInstrumenter{
		counters: make(map[string]map[int]bool),
		stmts:    make(map[string]int),
	}
}

// Instrument inserts coverage counters into compiled JavaScript.
// Before each statement line, it inserts: __cov["file"][line]++;
func (c *CoverageInstrumenter) Instrument(js string, filename string) string {
	lines := strings.Split(js, "\n")
	var out strings.Builder

	// Inject coverage preamble
	out.WriteString("if (typeof __cov === 'undefined') { var __cov = {}; }\n")
	out.WriteString(fmt.Sprintf("__cov[%q] = __cov[%q] || {};\n", filename, filename))

	stmtCount := 0
	if c.counters[filename] == nil {
		c.counters[filename] = make(map[int]bool)
	}

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip empty lines, comments, and braces-only lines
		if isInstrumentableLine(trimmed) {
			out.WriteString(fmt.Sprintf("__cov[%q][%d] = (__cov[%q][%d] || 0) + 1;\n", filename, lineNum, filename, lineNum))
			stmtCount++
			c.counters[filename][lineNum] = false // mark as uncovered initially
		}

		out.WriteString(line)
		if i < len(lines)-1 {
			out.WriteString("\n")
		}
	}

	c.stmts[filename] = stmtCount

	// Append coverage dump code
	out.WriteString("\n// Coverage dump\n")
	out.WriteString("if (typeof process !== 'undefined') {\n")
	out.WriteString("  process.on('exit', () => {\n")
	out.WriteString("    console.log('__COVERAGE_DATA__' + JSON.stringify(__cov));\n")
	out.WriteString("  });\n")
	out.WriteString("}\n")

	return out.String()
}

// isInstrumentableLine returns true if the line should be instrumented.
func isInstrumentableLine(trimmed string) bool {
	if trimmed == "" {
		return false
	}
	if strings.HasPrefix(trimmed, "//") {
		return false
	}
	if trimmed == "{" || trimmed == "}" || trimmed == "};" || trimmed == "});" {
		return false
	}
	if trimmed == "})();" || trimmed == "'use strict';" {
		return false
	}
	// Skip function/class declarations (they are structural, not executable)
	if strings.HasPrefix(trimmed, "function ") && strings.HasSuffix(trimmed, "{") {
		return false
	}
	return true
}

// MarkCovered marks a specific line as covered (called when parsing output).
func (c *CoverageInstrumenter) MarkCovered(filename string, line int) {
	if c.counters[filename] == nil {
		c.counters[filename] = make(map[int]bool)
	}
	c.counters[filename][line] = true
}

// ParseCoverageOutput extracts coverage data from program output containing __COVERAGE_DATA__.
func (c *CoverageInstrumenter) ParseCoverageOutput(output string) {
	re := regexp.MustCompile(`__COVERAGE_DATA__({.*})`)
	matches := re.FindStringSubmatch(output)
	if len(matches) < 2 {
		return
	}

	// Simple JSON parsing for coverage data: {"file": {line: count, ...}, ...}
	// We mark lines as covered if count > 0
	data := matches[1]
	// Parse file sections
	fileRe := regexp.MustCompile(`"([^"]+)":\s*\{([^}]*)\}`)
	fileMatches := fileRe.FindAllStringSubmatch(data, -1)
	for _, fm := range fileMatches {
		filename := fm[1]
		lineData := fm[2]
		if c.counters[filename] == nil {
			c.counters[filename] = make(map[int]bool)
		}
		// Parse line:count pairs
		lineRe := regexp.MustCompile(`"?(\d+)"?\s*:\s*(\d+)`)
		lineMatches := lineRe.FindAllStringSubmatch(lineData, -1)
		for _, lm := range lineMatches {
			var lineNum int
			fmt.Sscanf(lm[1], "%d", &lineNum)
			var count int
			fmt.Sscanf(lm[2], "%d", &count)
			c.counters[filename][lineNum] = count > 0
		}
	}
}

// CoverageResult holds coverage data for a single file.
type CoverageResult struct {
	Filename   string
	TotalStmts int
	Covered    int
	Percentage float64
}

// GetResults computes coverage results for all instrumented files.
func (c *CoverageInstrumenter) GetResults() []CoverageResult {
	var results []CoverageResult

	for filename, lines := range c.counters {
		total := len(lines)
		covered := 0
		for _, isCovered := range lines {
			if isCovered {
				covered++
			}
		}
		pct := 0.0
		if total > 0 {
			pct = float64(covered) / float64(total) * 100.0
		}
		results = append(results, CoverageResult{
			Filename:   filename,
			TotalStmts: total,
			Covered:    covered,
			Percentage: pct,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Filename < results[j].Filename
	})

	return results
}

// TotalCoverage returns the aggregate coverage percentage.
func (c *CoverageInstrumenter) TotalCoverage() float64 {
	totalStmts := 0
	totalCovered := 0
	for _, lines := range c.counters {
		for _, isCovered := range lines {
			totalStmts++
			if isCovered {
				totalCovered++
			}
		}
	}
	if totalStmts == 0 {
		return 0
	}
	return float64(totalCovered) / float64(totalStmts) * 100.0
}

// GenerateReport generates a text-based coverage report.
func (c *CoverageInstrumenter) GenerateReport() string {
	results := c.GetResults()

	var out strings.Builder
	out.WriteString("Coverage Report:\n")
	out.WriteString(strings.Repeat("\u2500", 55) + "\n")
	out.WriteString(fmt.Sprintf("%-30s %6s  %7s\n", "File", "Stmts", "Cover"))
	out.WriteString(strings.Repeat("\u2500", 55) + "\n")

	totalStmts := 0
	totalCovered := 0

	for _, r := range results {
		out.WriteString(fmt.Sprintf("%-30s %6d  %6.1f%%\n", r.Filename, r.TotalStmts, r.Percentage))
		totalStmts += r.TotalStmts
		totalCovered += r.Covered
	}

	totalPct := 0.0
	if totalStmts > 0 {
		totalPct = float64(totalCovered) / float64(totalStmts) * 100.0
	}

	out.WriteString(strings.Repeat("\u2500", 55) + "\n")
	out.WriteString(fmt.Sprintf("%-30s %6d  %6.1f%%\n", "Total", totalStmts, totalPct))

	return out.String()
}

// GenerateHTML generates an HTML coverage report.
func (c *CoverageInstrumenter) GenerateHTML() string {
	results := c.GetResults()
	totalPct := c.TotalCoverage()

	var out strings.Builder
	out.WriteString("<!DOCTYPE html>\n<html><head>\n")
	out.WriteString("<meta charset=\"UTF-8\">\n")
	out.WriteString("<title>Quill Coverage Report</title>\n")
	out.WriteString("<style>\n")
	out.WriteString("body { font-family: -apple-system, sans-serif; max-width: 800px; margin: 0 auto; padding: 40px 20px; background: #0D1117; color: #E6EDF3; }\n")
	out.WriteString("h1 { color: #1EB969; }\n")
	out.WriteString("table { width: 100%; border-collapse: collapse; margin-top: 20px; }\n")
	out.WriteString("th, td { padding: 10px 16px; text-align: left; border-bottom: 1px solid #30363D; }\n")
	out.WriteString("th { color: #8B949E; font-size: 12px; text-transform: uppercase; }\n")
	out.WriteString(".good { color: #1EB969; }\n")
	out.WriteString(".warn { color: #D29922; }\n")
	out.WriteString(".bad { color: #F85149; }\n")
	out.WriteString(".bar { height: 8px; border-radius: 4px; background: #30363D; }\n")
	out.WriteString(".bar-fill { height: 100%; border-radius: 4px; }\n")
	out.WriteString("footer { margin-top: 40px; color: #6E7681; font-size: 13px; }\n")
	out.WriteString("</style>\n</head><body>\n")
	out.WriteString(fmt.Sprintf("<h1>Coverage Report — %.1f%%</h1>\n", totalPct))
	out.WriteString("<table>\n")
	out.WriteString("<tr><th>File</th><th>Statements</th><th>Coverage</th><th></th></tr>\n")

	for _, r := range results {
		colorClass := "good"
		if r.Percentage < 50 {
			colorClass = "bad"
		} else if r.Percentage < 80 {
			colorClass = "warn"
		}
		barColor := "#1EB969"
		if r.Percentage < 50 {
			barColor = "#F85149"
		} else if r.Percentage < 80 {
			barColor = "#D29922"
		}
		out.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%d</td><td class=\"%s\">%.1f%%</td>", r.Filename, r.TotalStmts, colorClass, r.Percentage))
		out.WriteString(fmt.Sprintf("<td><div class=\"bar\"><div class=\"bar-fill\" style=\"width:%.0f%%;background:%s\"></div></div></td></tr>\n", r.Percentage, barColor))
	}

	out.WriteString("</table>\n")
	out.WriteString("<footer>Generated by <code>quill test --coverage</code></footer>\n")
	out.WriteString("</body></html>\n")

	return out.String()
}

// CheckThreshold returns an error if total coverage is below the given minimum percentage.
func (c *CoverageInstrumenter) CheckThreshold(min float64) error {
	total := c.TotalCoverage()
	if total < min {
		return fmt.Errorf("coverage %.1f%% is below minimum threshold %.1f%%", total, min)
	}
	return nil
}
