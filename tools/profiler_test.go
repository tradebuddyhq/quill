package tools

import (
	"strings"
	"testing"
)

func TestProfilerInstrument(t *testing.T) {
	prof := NewProfiler()
	js := `function hello() {
  console.log("hello");
}
hello();`

	instrumented := prof.Instrument(js)

	if !strings.Contains(instrumented, "__profile_data") {
		t.Error("should inject profiling preamble")
	}
	if !strings.Contains(instrumented, "__perf_now") {
		t.Error("should inject performance timer")
	}
	if !strings.Contains(instrumented, "__PROFILE_DATA__") {
		t.Error("should inject profile data dump")
	}
}

func TestProfilerParseResults(t *testing.T) {
	prof := NewProfiler()
	output := `some output
__PROFILE_DATA__{"fetchUsers":{"calls":12,"total":245.3},"renderPage":{"calls":48,"total":189.0}}
`

	entries := prof.ParseResults(output)

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Should be sorted by total time descending
	if entries[0].Function != "fetchUsers" {
		t.Errorf("expected fetchUsers first, got %s", entries[0].Function)
	}
	if entries[0].Calls != 12 {
		t.Errorf("expected 12 calls, got %d", entries[0].Calls)
	}
	if entries[0].TotalMs != 245.3 {
		t.Errorf("expected 245.3ms total, got %f", entries[0].TotalMs)
	}
}

func TestProfilerFormatReport(t *testing.T) {
	prof := NewProfiler()
	entries := []ProfileEntry{
		{Function: "fetchUsers", Calls: 12, TotalMs: 245.3, AvgMs: 20.4},
		{Function: "renderPage", Calls: 48, TotalMs: 189.0, AvgMs: 3.9},
	}

	report := prof.FormatReport(entries)

	if !strings.Contains(report, "Profile Report") {
		t.Error("should contain header")
	}
	if !strings.Contains(report, "fetchUsers") {
		t.Error("should list fetchUsers")
	}
	if !strings.Contains(report, "renderPage") {
		t.Error("should list renderPage")
	}
	if !strings.Contains(report, "12") {
		t.Error("should show call count")
	}
}

func TestProfilerEmptyResults(t *testing.T) {
	prof := NewProfiler()
	entries := prof.ParseResults("no profiling data here")

	if entries != nil {
		t.Error("should return nil for output without profile data")
	}
}
