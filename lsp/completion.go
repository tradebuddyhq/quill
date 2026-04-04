package lsp

import (
	"quill/ast"
	"strings"
)

// CompletionProvider generates completion suggestions.
type CompletionProvider struct{}

func NewCompletionProvider() *CompletionProvider {
	return &CompletionProvider{}
}

// GetCompletions returns a list of completion items for the given position.
func (c *CompletionProvider) GetCompletions(doc *Document, pos Position, program *ast.Program) CompletionList {
	items := []CompletionItem{}

	// Get partial word for filtering
	prefix := getPrefix(doc, pos)

	// Keyword completions
	for _, kw := range keywordCompletions {
		if prefix == "" || strings.HasPrefix(kw.Label, prefix) {
			items = append(items, kw)
		}
	}

	// Stdlib function completions
	for name, info := range stdlibDocs {
		if prefix == "" || strings.HasPrefix(name, prefix) {
			items = append(items, CompletionItem{
				Label:         name,
				Kind:          CompletionKindFunction,
				Detail:        info.Signature,
				Documentation: info.Doc,
				InsertText:    name,
			})
		}
	}

	// User-defined completions from AST
	if program != nil {
		seen := make(map[string]bool)

		for _, stmt := range program.Statements {
			switch s := stmt.(type) {
			case *ast.FuncDefinition:
				if !seen[s.Name] && (prefix == "" || strings.HasPrefix(s.Name, prefix)) {
					sig := buildFuncSignature(s)
					items = append(items, CompletionItem{
						Label:         s.Name,
						Kind:          CompletionKindFunction,
						Detail:        sig,
						Documentation: "User-defined function",
						InsertText:    s.Name,
					})
					seen[s.Name] = true
				}

			case *ast.AssignStatement:
				if !seen[s.Name] && (prefix == "" || strings.HasPrefix(s.Name, prefix)) {
					typeLabel := inferExprTypeLabel(s.Value)
					items = append(items, CompletionItem{
						Label:         s.Name,
						Kind:          CompletionKindVariable,
						Detail:        typeLabel,
						Documentation: "User-defined variable",
						InsertText:    s.Name,
					})
					seen[s.Name] = true
				}

			case *ast.DescribeStatement:
				if !seen[s.Name] && (prefix == "" || strings.HasPrefix(s.Name, prefix)) {
					items = append(items, CompletionItem{
						Label:         s.Name,
						Kind:          CompletionKindVariable,
						Detail:        "class",
						Documentation: "User-defined class",
						InsertText:    s.Name,
					})
					seen[s.Name] = true
				}

			case *ast.DefineStatement:
				if !seen[s.Name] && (prefix == "" || strings.HasPrefix(s.Name, prefix)) {
					items = append(items, CompletionItem{
						Label:         s.Name,
						Kind:          CompletionKindVariable,
						Detail:        "type",
						Documentation: "User-defined type",
						InsertText:    s.Name,
					})
					seen[s.Name] = true
				}
				for _, v := range s.Variants {
					if !seen[v.Name] && (prefix == "" || strings.HasPrefix(v.Name, prefix)) {
						items = append(items, CompletionItem{
							Label:         v.Name,
							Kind:          CompletionKindVariable,
							Detail:        "variant of " + s.Name,
							Documentation: "Enum variant",
							InsertText:    v.Name,
						})
						seen[v.Name] = true
					}
				}

			case *ast.ForEachStatement:
				if !seen[s.Variable] && (prefix == "" || strings.HasPrefix(s.Variable, prefix)) {
					items = append(items, CompletionItem{
						Label:      s.Variable,
						Kind:       CompletionKindVariable,
						Detail:     "loop variable",
						InsertText: s.Variable,
					})
					seen[s.Variable] = true
				}
			}
		}
	}

	// Snippet completions
	for _, snip := range snippetCompletions {
		if prefix == "" || strings.HasPrefix(snip.Label, prefix) {
			items = append(items, snip)
		}
	}

	return CompletionList{
		IsIncomplete: false,
		Items:        items,
	}
}

func getPrefix(doc *Document, pos Position) string {
	if pos.Line < 0 || pos.Line >= len(doc.Lines) {
		return ""
	}
	line := doc.Lines[pos.Line]
	end := pos.Character
	if end > len(line) {
		end = len(line)
	}
	start := end
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}
	return line[start:end]
}

var keywordCompletions = []CompletionItem{
	{Label: "is", Kind: CompletionKindKeyword, Detail: "assignment", Documentation: "Assigns a value to a variable", InsertText: "is "},
	{Label: "are", Kind: CompletionKindKeyword, Detail: "assignment (plural)", Documentation: "Assigns a value to a variable (plural form)", InsertText: "are "},
	{Label: "say", Kind: CompletionKindKeyword, Detail: "output", Documentation: "Prints a value to the console", InsertText: "say "},
	{Label: "if", Kind: CompletionKindKeyword, Detail: "conditional", Documentation: "Starts a conditional block", InsertText: "if "},
	{Label: "otherwise", Kind: CompletionKindKeyword, Detail: "else branch", Documentation: "Alternative branch of an if statement", InsertText: "otherwise"},
	{Label: "for", Kind: CompletionKindKeyword, Detail: "loop", Documentation: "Used with 'each' for iteration", InsertText: "for "},
	{Label: "each", Kind: CompletionKindKeyword, Detail: "iteration", Documentation: "Used with 'for' to iterate", InsertText: "each "},
	{Label: "in", Kind: CompletionKindKeyword, Detail: "membership", Documentation: "Used in for each loops", InsertText: "in "},
	{Label: "to", Kind: CompletionKindKeyword, Detail: "function def", Documentation: "Defines a new function", InsertText: "to "},
	{Label: "give", Kind: CompletionKindKeyword, Detail: "return (part 1)", Documentation: "Used as 'give back' to return a value", InsertText: "give "},
	{Label: "back", Kind: CompletionKindKeyword, Detail: "return (part 2)", Documentation: "Used as 'give back' to return a value", InsertText: "back "},
	{Label: "while", Kind: CompletionKindKeyword, Detail: "loop", Documentation: "Repeats while a condition is true", InsertText: "while "},
	{Label: "try", Kind: CompletionKindKeyword, Detail: "error handling", Documentation: "Starts a try block", InsertText: "try"},
	{Label: "fails", Kind: CompletionKindKeyword, Detail: "error handling", Documentation: "Catches errors from a try block", InsertText: "fails "},
	{Label: "match", Kind: CompletionKindKeyword, Detail: "pattern matching", Documentation: "Matches a value against patterns", InsertText: "match "},
	{Label: "when", Kind: CompletionKindKeyword, Detail: "pattern case", Documentation: "Defines a case in a match block", InsertText: "when "},
	{Label: "define", Kind: CompletionKindKeyword, Detail: "type definition", Documentation: "Defines an enum/algebraic type", InsertText: "define "},
	{Label: "describe", Kind: CompletionKindKeyword, Detail: "class definition", Documentation: "Defines a class", InsertText: "describe "},
	{Label: "use", Kind: CompletionKindKeyword, Detail: "import", Documentation: "Imports a module", InsertText: "use "},
	{Label: "from", Kind: CompletionKindKeyword, Detail: "import", Documentation: "Selective imports", InsertText: "from "},
	{Label: "new", Kind: CompletionKindKeyword, Detail: "constructor", Documentation: "Creates a new instance", InsertText: "new "},
	{Label: "my", Kind: CompletionKindKeyword, Detail: "self-reference", Documentation: "Refers to the current instance", InsertText: "my "},
	{Label: "await", Kind: CompletionKindKeyword, Detail: "async", Documentation: "Waits for an async operation", InsertText: "await "},
	{Label: "as", Kind: CompletionKindKeyword, Detail: "alias", Documentation: "Gives an import an alias", InsertText: "as "},
	{Label: "with", Kind: CompletionKindKeyword, Detail: "error binding", Documentation: "Binds error variable in fails block", InsertText: "with "},
	{Label: "test", Kind: CompletionKindKeyword, Detail: "testing", Documentation: "Defines a test block", InsertText: "test "},
	{Label: "expect", Kind: CompletionKindKeyword, Detail: "assertion", Documentation: "Asserts a condition in a test", InsertText: "expect "},
	{Label: "and", Kind: CompletionKindKeyword, Detail: "logical AND", Documentation: "Logical AND operator", InsertText: "and "},
	{Label: "or", Kind: CompletionKindKeyword, Detail: "logical OR", Documentation: "Logical OR operator", InsertText: "or "},
	{Label: "not", Kind: CompletionKindKeyword, Detail: "logical NOT", Documentation: "Logical NOT operator", InsertText: "not "},
	{Label: "greater", Kind: CompletionKindKeyword, Detail: "comparison", Documentation: "Used as 'greater than'", InsertText: "greater "},
	{Label: "less", Kind: CompletionKindKeyword, Detail: "comparison", Documentation: "Used as 'less than'", InsertText: "less "},
	{Label: "than", Kind: CompletionKindKeyword, Detail: "comparison", Documentation: "Used after greater/less", InsertText: "than "},
	{Label: "equal", Kind: CompletionKindKeyword, Detail: "comparison", Documentation: "Tests equality", InsertText: "equal "},
	{Label: "contains", Kind: CompletionKindKeyword, Detail: "membership", Documentation: "Checks if collection contains an element", InsertText: "contains "},
	{Label: "extends", Kind: CompletionKindKeyword, Detail: "inheritance", Documentation: "A class extends another", InsertText: "extends "},
	{Label: "nothing", Kind: CompletionKindKeyword, Detail: "null value", Documentation: "Represents the absence of a value", InsertText: "nothing"},
	{Label: "yes", Kind: CompletionKindKeyword, Detail: "boolean true", Documentation: "Boolean true literal", InsertText: "yes"},
	{Label: "no", Kind: CompletionKindKeyword, Detail: "boolean false", Documentation: "Boolean false literal", InsertText: "no"},
	{Label: "break", Kind: CompletionKindKeyword, Detail: "loop control", Documentation: "Exits the current loop", InsertText: "break"},
	{Label: "continue", Kind: CompletionKindKeyword, Detail: "loop control", Documentation: "Skips to the next iteration", InsertText: "continue"},
	{Label: "of", Kind: CompletionKindKeyword, Detail: "type annotation", Documentation: "Used in generics like 'list of number'", InsertText: "of "},
}

var snippetCompletions = []CompletionItem{
	{
		Label:         "if-block",
		Kind:          CompletionKindSnippet,
		Detail:        "if ... otherwise",
		Documentation: "If/otherwise block",
		InsertText:    "if ${1:condition}\n\t${2:body}\notherwise\n\t${3:else_body}",
		InsertTextFmt: InsertTextFormatSnippet,
	},
	{
		Label:         "for-each",
		Kind:          CompletionKindSnippet,
		Detail:        "for each ... in ...",
		Documentation: "For each loop",
		InsertText:    "for each ${1:item} in ${2:list}\n\t${3:body}",
		InsertTextFmt: InsertTextFormatSnippet,
	},
	{
		Label:         "while-loop",
		Kind:          CompletionKindSnippet,
		Detail:        "while ...",
		Documentation: "While loop",
		InsertText:    "while ${1:condition}\n\t${2:body}",
		InsertTextFmt: InsertTextFormatSnippet,
	},
	{
		Label:         "function",
		Kind:          CompletionKindSnippet,
		Detail:        "to ... -> ...",
		Documentation: "Function definition",
		InsertText:    "to ${1:name} ${2:params}\n\t${3:body}\n\tgive back ${4:result}",
		InsertTextFmt: InsertTextFormatSnippet,
	},
	{
		Label:         "try-fails",
		Kind:          CompletionKindSnippet,
		Detail:        "try ... fails with ...",
		Documentation: "Try/catch block",
		InsertText:    "try\n\t${1:body}\nfails with ${2:error}\n\t${3:handler}",
		InsertTextFmt: InsertTextFormatSnippet,
	},
	{
		Label:         "match-block",
		Kind:          CompletionKindSnippet,
		Detail:        "match ... when ...",
		Documentation: "Pattern matching block",
		InsertText:    "match ${1:value}\n\twhen ${2:pattern}\n\t\t${3:body}\n\totherwise\n\t\t${4:default}",
		InsertTextFmt: InsertTextFormatSnippet,
	},
	{
		Label:         "describe-class",
		Kind:          CompletionKindSnippet,
		Detail:        "describe ClassName",
		Documentation: "Class definition",
		InsertText:    "describe ${1:ClassName}\n\t${2:property} is ${3:value}\n\n\tto ${4:method}\n\t\t${5:body}",
		InsertTextFmt: InsertTextFormatSnippet,
	},
	{
		Label:         "define-enum",
		Kind:          CompletionKindSnippet,
		Detail:        "define TypeName",
		Documentation: "Enum/algebraic type definition",
		InsertText:    "define ${1:TypeName}\n\twhen ${2:Variant1}\n\twhen ${3:Variant2}",
		InsertTextFmt: InsertTextFormatSnippet,
	},
	{
		Label:         "test-block",
		Kind:          CompletionKindSnippet,
		Detail:        "test \"...\"",
		Documentation: "Test block",
		InsertText:    "test \"${1:description}\"\n\t${2:setup}\n\texpect ${3:condition}",
		InsertTextFmt: InsertTextFormatSnippet,
	},
	{
		Label:         "use-import",
		Kind:          CompletionKindSnippet,
		Detail:        "use \"module\"",
		Documentation: "Import a module",
		InsertText:    "use \"${1:module}\"",
		InsertTextFmt: InsertTextFormatSnippet,
	},
}
