package codegen

import (
	"encoding/json"
	"strings"
)

// SourceMap generates a V3 source map.
type SourceMap struct {
	Version    int      `json:"version"`
	File       string   `json:"file"`
	SourceRoot string   `json:"sourceRoot"`
	Sources    []string `json:"sources"`
	Names      []string `json:"names"`
	Mappings   string   `json:"mappings"`
	// internal tracking
	segments []segment
	genLine  int
	genCol   int
	srcLine  int
	srcCol   int
}

type segment struct {
	genLine int
	genCol  int
	srcLine int
	srcCol  int
}

// NewSourceMap creates a new source map for the given source and output file.
func NewSourceMap(sourceFile, outputFile string) *SourceMap {
	return &SourceMap{
		Version:    3,
		File:       outputFile,
		SourceRoot: "",
		Sources:    []string{sourceFile},
		Names:      []string{},
	}
}

// AddMapping adds a source-to-generated mapping.
func (sm *SourceMap) AddMapping(genLine, genCol, srcLine, srcCol int) {
	sm.segments = append(sm.segments, segment{
		genLine: genLine,
		genCol:  genCol,
		srcLine: srcLine,
		srcCol:  srcCol,
	})
}

// ToJSON serializes the source map to JSON with VLQ-encoded mappings.
func (sm *SourceMap) ToJSON() string {
	sm.Mappings = sm.encodeMappings()
	data, _ := json.MarshalIndent(sm, "", "  ")
	return string(data)
}

// encodeMappings encodes all segments into VLQ-based mappings string.
func (sm *SourceMap) encodeMappings() string {
	if len(sm.segments) == 0 {
		return ""
	}

	// Group segments by generated line
	maxLine := 0
	for _, seg := range sm.segments {
		if seg.genLine > maxLine {
			maxLine = seg.genLine
		}
	}

	lineSegments := make([][]segment, maxLine+1)
	for _, seg := range sm.segments {
		lineSegments[seg.genLine] = append(lineSegments[seg.genLine], seg)
	}

	var lines []string
	prevGenCol := 0
	prevSrcLine := 0
	prevSrcCol := 0

	for _, segs := range lineSegments {
		var parts []string
		prevGenCol = 0 // reset gen col per line

		for _, seg := range segs {
			var vlq strings.Builder
			// Field 1: generated column (relative to previous in this line)
			vlq.WriteString(encodeVLQ(seg.genCol - prevGenCol))
			// Field 2: source index (always 0, relative)
			vlq.WriteString(encodeVLQ(0))
			// Field 3: source line (relative)
			vlq.WriteString(encodeVLQ(seg.srcLine - prevSrcLine))
			// Field 4: source column (relative)
			vlq.WriteString(encodeVLQ(seg.srcCol - prevSrcCol))

			prevGenCol = seg.genCol
			prevSrcLine = seg.srcLine
			prevSrcCol = seg.srcCol

			parts = append(parts, vlq.String())
		}
		lines = append(lines, strings.Join(parts, ","))
	}

	return strings.Join(lines, ";")
}

const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

// encodeVLQ encodes an integer as a Base64 VLQ string per the source map spec.
func encodeVLQ(value int) string {
	var result strings.Builder

	// Convert to VLQ signed representation
	vlq := value << 1
	if value < 0 {
		vlq = ((-value) << 1) | 1
	}

	for {
		digit := vlq & 0x1f
		vlq >>= 5
		if vlq > 0 {
			digit |= 0x20 // set continuation bit
		}
		result.WriteByte(base64Chars[digit])
		if vlq == 0 {
			break
		}
	}

	return result.String()
}
