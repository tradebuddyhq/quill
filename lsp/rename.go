package lsp

import (
	"quill/ast"
	"strings"
)

// RenameProvider handles textDocument/rename requests.
type RenameProvider struct{}

func NewRenameProvider() *RenameProvider {
	return &RenameProvider{}
}

// GetRename computes a WorkspaceEdit that renames the symbol at the given position.
func (rp *RenameProvider) GetRename(doc *Document, pos Position, newName string, program *ast.Program, uri string) *WorkspaceEdit {
	word := doc.GetWordAtPosition(pos)
	if word == "" {
		return nil
	}

	// Don't rename keywords or stdlib functions
	if _, ok := keywordDocs[word]; ok {
		return nil
	}
	if _, ok := stdlibDocs[word]; ok {
		return nil
	}

	// Find all occurrences of the word in the document
	var edits []TextEdit
	for i, line := range doc.Lines {
		offset := 0
		for {
			idx := strings.Index(line[offset:], word)
			if idx < 0 {
				break
			}
			col := offset + idx

			// Check word boundaries
			before := col > 0 && isIdentChar(line[col-1])
			after := col+len(word) < len(line) && isIdentChar(line[col+len(word)])
			if !before && !after {
				edits = append(edits, TextEdit{
					Range: Range{
						Start: Position{Line: i, Character: col},
						End:   Position{Line: i, Character: col + len(word)},
					},
					NewText: newName,
				})
			}
			offset = col + len(word)
		}
	}

	if len(edits) == 0 {
		return nil
	}

	return &WorkspaceEdit{
		Changes: map[string][]TextEdit{
			uri: edits,
		},
	}
}
