package debugger

import (
	"encoding/json"
	"testing"
)

// --- Source map tests ---

func makeTestSourceMap() *SourceMapData {
	// Simulate a source map where:
	//   Generated line 5 (0-based) -> Quill source line 0 (0-based) = Quill line 1
	//   Generated line 6 -> Quill source line 1 = Quill line 2
	//   Generated line 7 -> Quill source line 2 = Quill line 3
	//   Generated line 10 -> Quill source line 5 = Quill line 6
	sm := &SourceMapData{
		Version:  3,
		File:     "test.js",
		Sources:  []string{"test.quill"},
		Mappings: "",
		entries: [][]MappingEntry{
			{}, // gen line 0 - no mapping
			{}, // gen line 1
			{}, // gen line 2
			{}, // gen line 3
			{}, // gen line 4
			{{GenCol: 0, SrcIdx: 0, SrcLine: 0, SrcCol: 0}}, // gen line 5 -> src line 0
			{{GenCol: 0, SrcIdx: 0, SrcLine: 1, SrcCol: 0}}, // gen line 6 -> src line 1
			{{GenCol: 0, SrcIdx: 0, SrcLine: 2, SrcCol: 0}}, // gen line 7 -> src line 2
			{}, // gen line 8
			{}, // gen line 9
			{{GenCol: 0, SrcIdx: 0, SrcLine: 5, SrcCol: 0}}, // gen line 10 -> src line 5
		},
	}
	return sm
}

func TestJSLineToQuillLine(t *testing.T) {
	sm := makeTestSourceMap()

	tests := []struct {
		jsLine    int
		quillLine int
	}{
		{5, 1},  // gen 5 -> src 0 -> Quill line 1
		{6, 2},  // gen 6 -> src 1 -> Quill line 2
		{7, 3},  // gen 7 -> src 2 -> Quill line 3
		{10, 6}, // gen 10 -> src 5 -> Quill line 6
		{0, 0},  // no mapping
		{4, 0},  // no mapping
		{-1, 0}, // out of range
		{99, 0}, // out of range
	}

	for _, tt := range tests {
		got := sm.JSLineToQuillLine(tt.jsLine)
		if got != tt.quillLine {
			t.Errorf("JSLineToQuillLine(%d) = %d, want %d", tt.jsLine, got, tt.quillLine)
		}
	}
}

func TestQuillLineToJSLine(t *testing.T) {
	sm := makeTestSourceMap()

	tests := []struct {
		quillLine int
		jsLine    int
	}{
		{1, 5},   // Quill 1 (src 0) -> gen 5
		{2, 6},   // Quill 2 (src 1) -> gen 6
		{3, 7},   // Quill 3 (src 2) -> gen 7
		{6, 10},  // Quill 6 (src 5) -> gen 10
		{4, -1},  // no mapping for Quill line 4
		{99, -1}, // no mapping
	}

	for _, tt := range tests {
		got := sm.QuillLineToJSLine(tt.quillLine)
		if got != tt.jsLine {
			t.Errorf("QuillLineToJSLine(%d) = %d, want %d", tt.quillLine, got, tt.jsLine)
		}
	}
}

func TestRoundTripMapping(t *testing.T) {
	sm := makeTestSourceMap()

	// For lines that have mappings, round-tripping should work
	quillLines := []int{1, 2, 3, 6}
	for _, ql := range quillLines {
		jsLine := sm.QuillLineToJSLine(ql)
		if jsLine < 0 {
			t.Errorf("QuillLineToJSLine(%d) returned -1", ql)
			continue
		}
		backToQuill := sm.JSLineToQuillLine(jsLine)
		if backToQuill != ql {
			t.Errorf("Round trip failed: Quill %d -> JS %d -> Quill %d", ql, jsLine, backToQuill)
		}
	}
}

// --- VLQ decode tests ---

func TestDecodeVLQSegment(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"AAAA", []int{0, 0, 0, 0}},
		{"AACA", []int{0, 0, 1, 0}},
		{"AAEA", []int{0, 0, 2, 0}},
		{"AADA", []int{0, 0, -1, 0}},
	}

	for _, tt := range tests {
		got := decodeVLQSegment(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("decodeVLQSegment(%q) = %v (len %d), want %v (len %d)",
				tt.input, got, len(got), tt.expected, len(tt.expected))
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("decodeVLQSegment(%q)[%d] = %d, want %d",
					tt.input, i, got[i], tt.expected[i])
			}
		}
	}
}

func TestDecodeMappingsRealFormat(t *testing.T) {
	// "AAAA;AACA;AACA" means:
	// Line 0: col 0 -> src 0, line 0, col 0
	// Line 1: col 0 -> src 0, line 1, col 0
	// Line 2: col 0 -> src 0, line 2, col 0
	entries := decodeMappings("AAAA;AACA;AACA")

	if len(entries) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(entries))
	}

	// Line 0 -> srcLine 0
	if len(entries[0]) != 1 || entries[0][0].SrcLine != 0 {
		t.Errorf("line 0: expected srcLine 0, got %v", entries[0])
	}
	// Line 1 -> srcLine 1
	if len(entries[1]) != 1 || entries[1][0].SrcLine != 1 {
		t.Errorf("line 1: expected srcLine 1, got %v", entries[1])
	}
	// Line 2 -> srcLine 2
	if len(entries[2]) != 1 || entries[2][0].SrcLine != 2 {
		t.Errorf("line 2: expected srcLine 2, got %v", entries[2])
	}
}

func TestLoadSourceMapFromJSON(t *testing.T) {
	smJSON := `{
		"version": 3,
		"file": "test.js",
		"sources": ["test.quill"],
		"names": [],
		"mappings": "AAAA;AACA;AACA"
	}`

	sm, err := LoadSourceMapFromJSON(smJSON)
	if err != nil {
		t.Fatalf("LoadSourceMapFromJSON failed: %v", err)
	}

	if sm.Version != 3 {
		t.Errorf("Version = %d, want 3", sm.Version)
	}

	// Line 0 -> Quill line 1
	if got := sm.JSLineToQuillLine(0); got != 1 {
		t.Errorf("JSLineToQuillLine(0) = %d, want 1", got)
	}
	// Line 1 -> Quill line 2
	if got := sm.JSLineToQuillLine(1); got != 2 {
		t.Errorf("JSLineToQuillLine(1) = %d, want 2", got)
	}
}

// --- Breakpoint management tests ---

func TestBreakpointManagement(t *testing.T) {
	dbg := New("test.quill")
	dbg.quillLines = []string{
		"x is 10",
		"y is 20",
		"say x + y",
	}
	dbg.sourceMap = makeTestSourceMap()

	// No breakpoints initially
	if len(dbg.Breakpoints()) != 0 {
		t.Errorf("expected 0 breakpoints, got %d", len(dbg.Breakpoints()))
	}

	// Add breakpoints (these will fail without CDP but we can test the map directly)
	dbg.breakpoints[1] = "bp-1"
	dbg.breakpoints[3] = "bp-3"

	if len(dbg.Breakpoints()) != 2 {
		t.Errorf("expected 2 breakpoints, got %d", len(dbg.Breakpoints()))
	}

	if _, exists := dbg.breakpoints[1]; !exists {
		t.Error("expected breakpoint on line 1")
	}
	if _, exists := dbg.breakpoints[3]; !exists {
		t.Error("expected breakpoint on line 3")
	}

	// Remove breakpoint
	delete(dbg.breakpoints, 1)
	if len(dbg.Breakpoints()) != 1 {
		t.Errorf("expected 1 breakpoint after removal, got %d", len(dbg.Breakpoints()))
	}
	if _, exists := dbg.breakpoints[1]; exists {
		t.Error("breakpoint on line 1 should be removed")
	}
}

func TestBreakpointOutOfRange(t *testing.T) {
	dbg := New("test.quill")
	dbg.quillLines = []string{"line1", "line2"}
	dbg.sourceMap = makeTestSourceMap()

	err := dbg.SetBreakpoint(0)
	if err == nil {
		t.Error("expected error for line 0")
	}

	err = dbg.SetBreakpoint(99)
	if err == nil {
		t.Error("expected error for line 99")
	}
}

// --- REPL command parsing tests ---

func TestParseREPLCommand(t *testing.T) {
	tests := []struct {
		input   string
		cmd     string
		arg     string
	}{
		{"break 5", "break", "5"},
		{"b 10", "b", "10"},
		{"continue", "continue", ""},
		{"c", "c", ""},
		{"step", "step", ""},
		{"print x + y", "print", "x + y"},
		{"p myVar", "p", "myVar"},
		{"quit", "quit", ""},
		{"", "", ""},
		{"  list  42  ", "list", "42"},
		{"delete 3", "delete", "3"},
	}

	for _, tt := range tests {
		cmd, arg := ParseREPLCommand(tt.input)
		if cmd != tt.cmd {
			t.Errorf("ParseREPLCommand(%q) cmd = %q, want %q", tt.input, cmd, tt.cmd)
		}
		if arg != tt.arg {
			t.Errorf("ParseREPLCommand(%q) arg = %q, want %q", tt.input, arg, tt.arg)
		}
	}
}

func TestResolveCommandAlias(t *testing.T) {
	tests := []struct {
		alias    string
		expected string
	}{
		{"b", "break"},
		{"d", "delete"},
		{"c", "continue"},
		{"s", "step"},
		{"i", "into"},
		{"o", "out"},
		{"p", "print"},
		{"l", "locals"},
		{"bt", "stack"},
		{"ls", "list"},
		{"h", "help"},
		{"q", "quit"},
		{"break", "break"},
		{"continue", "continue"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		got := ResolveCommandAlias(tt.alias)
		if got != tt.expected {
			t.Errorf("ResolveCommandAlias(%q) = %q, want %q", tt.alias, got, tt.expected)
		}
	}
}

// --- CDP message type tests ---

func TestCDPMessageMarshal(t *testing.T) {
	params := json.RawMessage(`{"lineNumber": 5}`)
	msg := CDPMessage{
		ID:     1,
		Method: "Debugger.setBreakpointByUrl",
		Params: &params,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded CDPMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.ID != 1 {
		t.Errorf("ID = %d, want 1", decoded.ID)
	}
	if decoded.Method != "Debugger.setBreakpointByUrl" {
		t.Errorf("Method = %q, want Debugger.setBreakpointByUrl", decoded.Method)
	}
}

func TestPausedEventUnmarshal(t *testing.T) {
	raw := `{
		"callFrames": [
			{
				"callFrameId": "frame-0",
				"functionName": "main",
				"location": {"scriptId": "42", "lineNumber": 10, "columnNumber": 0}
			}
		],
		"reason": "breakpoint"
	}`

	var evt PausedEvent
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if evt.Reason != "breakpoint" {
		t.Errorf("Reason = %q, want breakpoint", evt.Reason)
	}
	if len(evt.CallFrames) != 1 {
		t.Fatalf("expected 1 call frame, got %d", len(evt.CallFrames))
	}
	if evt.CallFrames[0].FunctionName != "main" {
		t.Errorf("FunctionName = %q, want main", evt.CallFrames[0].FunctionName)
	}
	if evt.CallFrames[0].Location.LineNumber != 10 {
		t.Errorf("LineNumber = %d, want 10", evt.CallFrames[0].Location.LineNumber)
	}
}

// --- formatRemoteObject tests ---

func TestFormatRemoteObject(t *testing.T) {
	tests := []struct {
		obj      RemoteObject
		expected string
	}{
		{RemoteObject{Type: "string", Value: "hello"}, `"hello"`},
		{RemoteObject{Type: "number", Description: "42"}, "42"},
		{RemoteObject{Type: "boolean", Description: "true"}, "true"},
		{RemoteObject{Type: "undefined"}, "undefined"},
		{RemoteObject{Type: "object", Subtype: "null"}, "nothing"},
		{RemoteObject{Type: "object", Description: "Array(3)"}, "Array(3)"},
		{RemoteObject{Type: "function", Description: "function foo() {}"}, "function foo() {}"},
	}

	for _, tt := range tests {
		got := formatRemoteObject(tt.obj)
		if got != tt.expected {
			t.Errorf("formatRemoteObject(%+v) = %q, want %q", tt.obj, got, tt.expected)
		}
	}
}

// --- Debugger state tests ---

func TestDebuggerInitialState(t *testing.T) {
	dbg := New("example.quill")

	if dbg.SourceFile() != "example.quill" {
		t.Errorf("SourceFile() = %q, want example.quill", dbg.SourceFile())
	}

	if dbg.IsPaused() {
		t.Error("should not be paused initially")
	}

	if dbg.CurrentLine() != 0 {
		t.Errorf("CurrentLine() = %d, want 0", dbg.CurrentLine())
	}

	if len(dbg.Breakpoints()) != 0 {
		t.Errorf("should have 0 breakpoints initially")
	}
}
