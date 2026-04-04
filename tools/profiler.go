package tools

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// ProfileEntry holds profiling data for a single function.
type ProfileEntry struct {
	Function string
	Calls    int
	TotalMs  float64
	AvgMs    float64
}

// Profiler instruments compiled JavaScript with timing around function calls
// and collects profiling data.
type Profiler struct {
	entries []ProfileEntry
}

// NewProfiler creates a new Profiler.
func NewProfiler() *Profiler {
	return &Profiler{}
}

// Instrument wraps function bodies with performance timing instrumentation.
// It inserts timing code around function declarations in the compiled JS.
func (p *Profiler) Instrument(js string) string {
	var out strings.Builder

	// Inject profiling preamble
	out.WriteString("const __profile_data = {};\n")
	out.WriteString("function __profile_log(name, ms) {\n")
	out.WriteString("  if (!__profile_data[name]) __profile_data[name] = { calls: 0, total: 0 };\n")
	out.WriteString("  __profile_data[name].calls++;\n")
	out.WriteString("  __profile_data[name].total += ms;\n")
	out.WriteString("}\n")
	out.WriteString("const __perf_now = typeof performance !== 'undefined' ? () => performance.now() : () => Date.now();\n\n")

	// Instrument function bodies
	// Match: function name(params) {
	funcRe := regexp.MustCompile(`(function\s+(\w+)\s*\([^)]*\)\s*\{)`)

	instrumented := funcRe.ReplaceAllStringFunc(js, func(match string) string {
		sub := funcRe.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		fnName := sub[2]
		// Skip internal/runtime functions
		if strings.HasPrefix(fnName, "__") || fnName == "require" {
			return match
		}
		return match + fmt.Sprintf("\nconst __t_%s = __perf_now();", fnName)
	})

	// For each instrumented function, we need to add timing at the end.
	// We do this by adding a process.on('exit') handler that dumps timing data.
	out.WriteString(instrumented)
	out.WriteString("\n\n// Profile dump\n")
	out.WriteString("if (typeof process !== 'undefined') {\n")
	out.WriteString("  process.on('exit', () => {\n")
	out.WriteString("    console.log('__PROFILE_DATA__' + JSON.stringify(__profile_data));\n")
	out.WriteString("  });\n")
	out.WriteString("}\n")

	return out.String()
}

// InstrumentFunction wraps a single function's JS code with timing.
// Returns JS that times the function and reports to __profile_log.
func (p *Profiler) InstrumentFunction(fnName string, fnBody string) string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("const __t_%s = __perf_now();\n", fnName))
	out.WriteString(fnBody)
	out.WriteString(fmt.Sprintf("\n__profile_log(%q, __perf_now() - __t_%s);\n", fnName, fnName))
	return out.String()
}

// ParseResults extracts profiling data from program output containing __PROFILE_DATA__.
func (p *Profiler) ParseResults(output string) []ProfileEntry {
	re := regexp.MustCompile(`__PROFILE_DATA__({.*})`)
	matches := re.FindStringSubmatch(output)
	if len(matches) < 2 {
		return nil
	}

	data := matches[1]
	// Parse: {"fnName": {"calls": N, "total": M}, ...}
	fnRe := regexp.MustCompile(`"(\w+)":\s*\{\s*"calls":\s*(\d+),\s*"total":\s*([\d.]+)\s*\}`)
	fnMatches := fnRe.FindAllStringSubmatch(data, -1)

	var entries []ProfileEntry
	for _, fm := range fnMatches {
		name := fm[1]
		var calls int
		var total float64
		fmt.Sscanf(fm[2], "%d", &calls)
		fmt.Sscanf(fm[3], "%f", &total)
		avg := 0.0
		if calls > 0 {
			avg = total / float64(calls)
		}
		entries = append(entries, ProfileEntry{
			Function: name,
			Calls:    calls,
			TotalMs:  total,
			AvgMs:    avg,
		})
	}

	// Sort by total time descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].TotalMs > entries[j].TotalMs
	})

	p.entries = entries
	return entries
}

// FormatReport generates a text-based profiling report.
func (p *Profiler) FormatReport(entries []ProfileEntry) string {
	var out strings.Builder
	out.WriteString("Profile Report (top 10 by total time):\n")
	out.WriteString(strings.Repeat("\u2500", 60) + "\n")
	out.WriteString(fmt.Sprintf("%-25s %6s  %8s  %8s\n", "Function", "Calls", "Total", "Avg"))
	out.WriteString(strings.Repeat("\u2500", 60) + "\n")

	limit := len(entries)
	if limit > 10 {
		limit = 10
	}

	for i := 0; i < limit; i++ {
		e := entries[i]
		out.WriteString(fmt.Sprintf("%-25s %6d  %7.1fms  %7.1fms\n", e.Function, e.Calls, e.TotalMs, e.AvgMs))
	}

	out.WriteString(strings.Repeat("\u2500", 60) + "\n")
	return out.String()
}
