package tools

import (
	"strings"
	"testing"
)

func TestInstrumentInsertsCounters(t *testing.T) {
	cov := NewCoverageInstrumenter()
	js := `let x = 1;
console.log(x);
if (x > 0) {
  x = 2;
}`
	instrumented := cov.Instrument(js, "test.quill")

	if !strings.Contains(instrumented, `__cov["test.quill"]`) {
		t.Error("expected coverage counter initialization for test.quill")
	}
	if !strings.Contains(instrumented, `__cov["test.quill"][1]`) {
		t.Error("expected counter for line 1")
	}
	if !strings.Contains(instrumented, `__cov["test.quill"][2]`) {
		t.Error("expected counter for line 2")
	}
}

func TestInstrumentSkipsComments(t *testing.T) {
	cov := NewCoverageInstrumenter()
	js := `// this is a comment
let x = 1;`
	instrumented := cov.Instrument(js, "test.quill")

	// Line 1 is a comment, should not be instrumented
	if strings.Contains(instrumented, `__cov["test.quill"][1]`) {
		t.Error("should not instrument comment lines")
	}
	// Line 2 is real code, should be instrumented
	if !strings.Contains(instrumented, `__cov["test.quill"][2]`) {
		t.Error("should instrument code lines")
	}
}

func TestCoverageReportFormatting(t *testing.T) {
	cov := NewCoverageInstrumenter()

	// Simulate some coverage data
	cov.counters["src/auth.quill"] = map[int]bool{1: true, 2: true, 3: false, 4: true}
	cov.counters["src/main.quill"] = map[int]bool{1: true, 2: true}

	report := cov.GenerateReport()

	if !strings.Contains(report, "Coverage Report:") {
		t.Error("report should contain header")
	}
	if !strings.Contains(report, "src/auth.quill") {
		t.Error("report should list auth.quill")
	}
	if !strings.Contains(report, "src/main.quill") {
		t.Error("report should list main.quill")
	}
	if !strings.Contains(report, "Total") {
		t.Error("report should contain total line")
	}
}

func TestCoveragePercentage(t *testing.T) {
	cov := NewCoverageInstrumenter()
	cov.counters["test.quill"] = map[int]bool{1: true, 2: true, 3: false, 4: false}

	total := cov.TotalCoverage()
	if total != 50.0 {
		t.Errorf("expected 50%% coverage, got %.1f%%", total)
	}
}

func TestCoverageThreshold(t *testing.T) {
	cov := NewCoverageInstrumenter()
	cov.counters["test.quill"] = map[int]bool{1: true, 2: false}

	err := cov.CheckThreshold(80)
	if err == nil {
		t.Error("expected error when coverage is below threshold")
	}

	err = cov.CheckThreshold(40)
	if err != nil {
		t.Error("should not error when coverage meets threshold")
	}
}

func TestCoverageHTMLGeneration(t *testing.T) {
	cov := NewCoverageInstrumenter()
	cov.counters["test.quill"] = map[int]bool{1: true, 2: true, 3: false}

	html := cov.GenerateHTML()

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("should generate valid HTML")
	}
	if !strings.Contains(html, "Coverage Report") {
		t.Error("should contain title")
	}
	if !strings.Contains(html, "test.quill") {
		t.Error("should list the file")
	}
}
