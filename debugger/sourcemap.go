package debugger

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// SourceMapData represents a parsed V3 source map.
type SourceMapData struct {
	Version    int      `json:"version"`
	File       string   `json:"file"`
	SourceRoot string   `json:"sourceRoot"`
	Sources    []string `json:"sources"`
	Names      []string `json:"names"`
	Mappings   string   `json:"mappings"`

	// Decoded mappings: genLine -> []MappingEntry
	entries [][]MappingEntry
}

// MappingEntry represents a single decoded mapping segment.
type MappingEntry struct {
	GenCol  int
	SrcIdx  int
	SrcLine int
	SrcCol  int
}

// LoadSourceMap parses a .js.map file from disk.
func LoadSourceMap(path string) (*SourceMapData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read source map %q: %w", path, err)
	}

	var sm SourceMapData
	if err := json.Unmarshal(data, &sm); err != nil {
		return nil, fmt.Errorf("could not parse source map: %w", err)
	}

	sm.entries = decodeMappings(sm.Mappings)
	return &sm, nil
}

// LoadSourceMapFromJSON parses source map data from a JSON string.
func LoadSourceMapFromJSON(jsonStr string) (*SourceMapData, error) {
	var sm SourceMapData
	if err := json.Unmarshal([]byte(jsonStr), &sm); err != nil {
		return nil, fmt.Errorf("could not parse source map JSON: %w", err)
	}
	sm.entries = decodeMappings(sm.Mappings)
	return &sm, nil
}

// JSLineToQuillLine maps a 0-based JS line number to a 1-based Quill source line.
// Returns 0 if no mapping is found.
func (sm *SourceMapData) JSLineToQuillLine(jsLine int) int {
	if jsLine < 0 || jsLine >= len(sm.entries) {
		return 0
	}

	segs := sm.entries[jsLine]
	if len(segs) == 0 {
		return 0
	}

	// Return the source line from the first segment on this generated line.
	// Source lines in the map are 0-based; Quill lines are 1-based.
	return segs[0].SrcLine + 1
}

// QuillLineToJSLine maps a 1-based Quill source line to a 0-based JS line number.
// Returns -1 if no mapping is found.
func (sm *SourceMapData) QuillLineToJSLine(quillLine int) int {
	srcLine := quillLine - 1 // convert to 0-based

	// Scan all generated lines for one that maps to this source line.
	for genLine, segs := range sm.entries {
		for _, seg := range segs {
			if seg.SrcLine == srcLine {
				return genLine
			}
		}
	}
	return -1
}

// decodeMappings decodes the VLQ-encoded mappings string into per-line segments.
func decodeMappings(mappings string) [][]MappingEntry {
	if mappings == "" {
		return nil
	}

	lines := strings.Split(mappings, ";")
	result := make([][]MappingEntry, len(lines))

	// These are relative and accumulate across the entire mappings string.
	prevGenCol := 0
	prevSrcIdx := 0
	prevSrcLine := 0
	prevSrcCol := 0

	for lineIdx, line := range lines {
		if line == "" {
			continue
		}

		segments := strings.Split(line, ",")
		prevGenCol = 0 // generated column resets per line

		for _, seg := range segments {
			values := decodeVLQSegment(seg)
			if len(values) < 4 {
				continue
			}

			genCol := prevGenCol + values[0]
			srcIdx := prevSrcIdx + values[1]
			srcLine := prevSrcLine + values[2]
			srcCol := prevSrcCol + values[3]

			prevGenCol = genCol
			prevSrcIdx = srcIdx
			prevSrcLine = srcLine
			prevSrcCol = srcCol

			result[lineIdx] = append(result[lineIdx], MappingEntry{
				GenCol:  genCol,
				SrcIdx:  srcIdx,
				SrcLine: srcLine,
				SrcCol:  srcCol,
			})
		}
	}

	return result
}

const base64Decode = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

// decodeVLQSegment decodes a single VLQ segment into a slice of integers.
func decodeVLQSegment(segment string) []int {
	var values []int
	i := 0

	for i < len(segment) {
		value := 0
		shift := 0

		for {
			if i >= len(segment) {
				break
			}

			ch := segment[i]
			i++

			idx := strings.IndexByte(base64Decode, ch)
			if idx < 0 {
				break
			}

			hasContinuation := idx&0x20 != 0
			digit := idx & 0x1f
			value |= digit << shift
			shift += 5

			if !hasContinuation {
				break
			}
		}

		// Convert from VLQ signed representation
		if value&1 == 1 {
			value = -(value >> 1)
		} else {
			value = value >> 1
		}

		values = append(values, value)
	}

	return values
}
