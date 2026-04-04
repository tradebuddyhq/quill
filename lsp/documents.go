package lsp

import (
	"strings"
	"sync"
)

// DocumentStore holds all open documents.
type DocumentStore struct {
	mu   sync.RWMutex
	docs map[string]*Document
}

// Document represents an open text document.
type Document struct {
	URI     string
	Version int
	Content string
	Lines   []string
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		docs: make(map[string]*Document),
	}
}

func (s *DocumentStore) Open(uri string, version int, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.docs[uri] = &Document{
		URI:     uri,
		Version: version,
		Content: content,
		Lines:   splitLines(content),
	}
}

func (s *DocumentStore) Update(uri string, version int, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if doc, ok := s.docs[uri]; ok {
		doc.Version = version
		doc.Content = content
		doc.Lines = splitLines(content)
	} else {
		s.docs[uri] = &Document{
			URI:     uri,
			Version: version,
			Content: content,
			Lines:   splitLines(content),
		}
	}
}

func (s *DocumentStore) Get(uri string) *Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.docs[uri]
}

func (s *DocumentStore) Close(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.docs, uri)
}

// GetWordAtPosition returns the word under the cursor at the given position.
func (doc *Document) GetWordAtPosition(pos Position) string {
	if pos.Line < 0 || pos.Line >= len(doc.Lines) {
		return ""
	}
	line := doc.Lines[pos.Line]
	if pos.Character < 0 || pos.Character > len(line) {
		return ""
	}

	// Find word boundaries
	start := pos.Character
	end := pos.Character

	for start > 0 && isWordChar(line[start-1]) {
		start--
	}
	for end < len(line) && isWordChar(line[end]) {
		end++
	}

	if start == end {
		return ""
	}
	return line[start:end]
}

// GetLineContent returns the content of a specific line.
func (doc *Document) GetLineContent(line int) string {
	if line < 0 || line >= len(doc.Lines) {
		return ""
	}
	return doc.Lines[line]
}

// OffsetToPosition converts a byte offset to an LSP position.
func (doc *Document) OffsetToPosition(offset int) Position {
	line := 0
	col := 0
	for i := 0; i < offset && i < len(doc.Content); i++ {
		if doc.Content[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return Position{Line: line, Character: col}
}

// LineToRange returns the range covering an entire line (1-based line number to 0-based LSP range).
func (doc *Document) LineToRange(line1Based int) Range {
	line := line1Based - 1
	if line < 0 {
		line = 0
	}
	endChar := 0
	if line < len(doc.Lines) {
		endChar = len(doc.Lines[line])
	}
	return Range{
		Start: Position{Line: line, Character: 0},
		End:   Position{Line: line, Character: endChar},
	}
}

func splitLines(content string) []string {
	return strings.Split(content, "\n")
}

func isWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') ||
		b == '_'
}
