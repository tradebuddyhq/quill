package lsp

import (
	"quill/ast"
)

// DefinitionProvider finds the definition location of a symbol.
type DefinitionProvider struct{}

func NewDefinitionProvider() *DefinitionProvider {
	return &DefinitionProvider{}
}

// GetDefinition returns the location where the symbol at the given position is defined.
func (d *DefinitionProvider) GetDefinition(doc *Document, pos Position, program *ast.Program, uri string) *Location {
	word := doc.GetWordAtPosition(pos)
	if word == "" {
		return nil
	}

	if program == nil {
		return nil
	}

	// Search the AST for the definition of this symbol.
	// We search nested scopes (function bodies, for-loop bodies, etc.) as well as top-level.
	if loc := d.searchStatements(program.Statements, word, uri, doc); loc != nil {
		return loc
	}

	return nil
}

// searchStatements searches a slice of statements for a definition of the given word.
// It returns the first definition found.
func (d *DefinitionProvider) searchStatements(stmts []ast.Statement, word string, uri string, doc *Document) *Location {
	for _, stmt := range stmts {
		if loc := d.searchStatement(stmt, word, uri, doc); loc != nil {
			return loc
		}
	}
	return nil
}

// searchStatement checks a single statement for a definition of the given word,
// including searching nested bodies.
func (d *DefinitionProvider) searchStatement(stmt ast.Statement, word string, uri string, doc *Document) *Location {
	switch s := stmt.(type) {

	case *ast.AssignStatement:
		if s.Name == word {
			return d.makeLocation(uri, s.Line, s.Name, doc)
		}

	case *ast.TypedAssignStatement:
		if s.Name == word {
			return d.makeLocation(uri, s.Line, s.Name, doc)
		}

	case *ast.FuncDefinition:
		if s.Name == word {
			return d.makeLocation(uri, s.Line, s.Name, doc)
		}
		// Search parameters
		for _, param := range s.Params {
			if param == word {
				return d.makeLocation(uri, s.Line, param, doc)
			}
		}
		// Search function body
		if loc := d.searchStatements(s.Body, word, uri, doc); loc != nil {
			return loc
		}

	case *ast.DecoratedFuncDefinition:
		if s.Func != nil {
			if loc := d.searchStatement(s.Func, word, uri, doc); loc != nil {
				return loc
			}
		}

	case *ast.ComponentStatement:
		if s.Name == word {
			return d.makeLocation(uri, s.Line, s.Name, doc)
		}
		// Search component states
		for _, state := range s.States {
			if state.Name == word {
				return d.makeLocation(uri, state.Line, state.Name, doc)
			}
		}
		// Search component methods
		for _, method := range s.Methods {
			if method.Name == word {
				return d.makeLocation(uri, method.Line, method.Name, doc)
			}
			for _, param := range method.Params {
				if param == word {
					return d.makeLocation(uri, method.Line, param, doc)
				}
			}
			if loc := d.searchStatements(method.Body, word, uri, doc); loc != nil {
				return loc
			}
		}

	case *ast.DescribeStatement:
		if s.Name == word {
			return d.makeLocation(uri, s.Line, s.Name, doc)
		}
		// Search properties
		for _, prop := range s.Properties {
			if prop.Name == word {
				return d.makeLocation(uri, prop.Line, prop.Name, doc)
			}
		}
		// Search methods
		for _, method := range s.Methods {
			if method.Name == word {
				return d.makeLocation(uri, method.Line, method.Name, doc)
			}
			for _, param := range method.Params {
				if param == word {
					return d.makeLocation(uri, method.Line, param, doc)
				}
			}
			if loc := d.searchStatements(method.Body, word, uri, doc); loc != nil {
				return loc
			}
		}

	case *ast.DefineStatement:
		if s.Name == word {
			return d.makeLocation(uri, s.Line, s.Name, doc)
		}
		for _, v := range s.Variants {
			if v.Name == word {
				return d.makeLocation(uri, s.Line, v.Name, doc)
			}
		}
		for _, method := range s.Methods {
			if method.Name == word {
				return d.makeLocation(uri, method.Line, method.Name, doc)
			}
			if loc := d.searchStatements(method.Body, word, uri, doc); loc != nil {
				return loc
			}
		}

	case *ast.ForEachStatement:
		if s.Variable == word {
			return d.makeLocation(uri, s.Line, s.Variable, doc)
		}
		if loc := d.searchStatements(s.Body, word, uri, doc); loc != nil {
			return loc
		}

	case *ast.WhileStatement:
		if loc := d.searchStatements(s.Body, word, uri, doc); loc != nil {
			return loc
		}

	case *ast.IfStatement:
		if loc := d.searchStatements(s.Body, word, uri, doc); loc != nil {
			return loc
		}
		for _, elseIf := range s.ElseIfs {
			if loc := d.searchStatements(elseIf.Body, word, uri, doc); loc != nil {
				return loc
			}
		}
		if loc := d.searchStatements(s.Else, word, uri, doc); loc != nil {
			return loc
		}

	case *ast.TryCatchStatement:
		if s.ErrorVar == word {
			return d.makeLocation(uri, s.Line, s.ErrorVar, doc)
		}
		if loc := d.searchStatements(s.TryBody, word, uri, doc); loc != nil {
			return loc
		}
		if loc := d.searchStatements(s.CatchBody, word, uri, doc); loc != nil {
			return loc
		}

	case *ast.UseStatement:
		if s.Alias != "" && s.Alias == word {
			return d.makeLocation(uri, s.Line, s.Alias, doc)
		} else if s.Path == word {
			return d.makeLocation(uri, s.Line, s.Path, doc)
		}

	case *ast.FromUseStatement:
		for _, name := range s.Names {
			if name == word {
				return d.makeLocation(uri, s.Line, name, doc)
			}
		}

	case *ast.TestBlock:
		if loc := d.searchStatements(s.Body, word, uri, doc); loc != nil {
			return loc
		}

	case *ast.MatchStatement:
		for _, c := range s.Cases {
			if c.Binding == word {
				return d.makeLocation(uri, s.Line, c.Binding, doc)
			}
			if loc := d.searchStatements(c.Body, word, uri, doc); loc != nil {
				return loc
			}
		}

	case *ast.LoopStatement:
		if loc := d.searchStatements(s.Body, word, uri, doc); loc != nil {
			return loc
		}

	case *ast.TraitDeclaration:
		if s.Name == word {
			return d.makeLocation(uri, s.Line, s.Name, doc)
		}

	case *ast.TypeAliasStatement:
		if s.Name == word {
			return d.makeLocation(uri, s.Line, s.Name, doc)
		}

	case *ast.DestructureStatement:
		if loc := d.searchDestructurePattern(s.Pattern, word, uri, s.Line, doc); loc != nil {
			return loc
		}
	}

	return nil
}

// searchDestructurePattern searches a destructure pattern for a binding that matches word.
func (d *DefinitionProvider) searchDestructurePattern(pattern ast.DestructurePattern, word string, uri string, line int, doc *Document) *Location {
	if pattern == nil {
		return nil
	}
	switch p := pattern.(type) {
	case *ast.ObjectPattern:
		for _, f := range p.Fields {
			if f.Key == word {
				return d.makeLocation(uri, line, f.Key, doc)
			}
			if f.Nested != nil {
				if loc := d.searchDestructurePattern(f.Nested, word, uri, line, doc); loc != nil {
					return loc
				}
			}
		}
		if p.Rest == word {
			return d.makeLocation(uri, line, p.Rest, doc)
		}
	case *ast.ArrayPattern:
		for _, e := range p.Elements {
			if e.Name == word {
				return d.makeLocation(uri, line, e.Name, doc)
			}
			if e.Nested != nil {
				if loc := d.searchDestructurePattern(e.Nested, word, uri, line, doc); loc != nil {
					return loc
				}
			}
		}
		if p.Rest == word {
			return d.makeLocation(uri, line, p.Rest, doc)
		}
	}
	return nil
}

// makeLocation creates an LSP Location for a definition.
// line is 1-based (from the AST). We find the column of the word on that line.
func (d *DefinitionProvider) makeLocation(uri string, line1Based int, name string, doc *Document) *Location {
	line := line1Based - 1 // convert to 0-based
	if line < 0 {
		line = 0
	}

	col := 0
	if line < len(doc.Lines) {
		// Find the column where the name appears on this line
		lineText := doc.Lines[line]
		idx := findWordInLine(lineText, name)
		if idx >= 0 {
			col = idx
		}
	}

	return &Location{
		URI: uri,
		Range: Range{
			Start: Position{Line: line, Character: col},
			End:   Position{Line: line, Character: col + len(name)},
		},
	}
}

// findWordInLine finds the byte offset of `word` in `line` as a whole word.
// Returns -1 if not found.
func findWordInLine(line string, word string) int {
	start := 0
	for {
		idx := indexOf(line[start:], word)
		if idx < 0 {
			return -1
		}
		pos := start + idx
		// Check word boundaries
		before := pos == 0 || !isWordChar(line[pos-1])
		after := pos+len(word) >= len(line) || !isWordChar(line[pos+len(word)])
		if before && after {
			return pos
		}
		start = pos + 1
		if start >= len(line) {
			return -1
		}
	}
}

// indexOf returns the index of substr in s, or -1.
func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
