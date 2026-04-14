package lsp

import (
	"fmt"
	"quill/ast"
	"strings"
)

// SignatureHelpProvider provides function signature help when typing arguments.
type SignatureHelpProvider struct{}

func NewSignatureHelpProvider() *SignatureHelpProvider {
	return &SignatureHelpProvider{}
}

// GetSignatureHelp returns signature help for the function being called at the given position.
func (sh *SignatureHelpProvider) GetSignatureHelp(doc *Document, pos Position, program *ast.Program) *SignatureHelp {
	line := doc.GetLineContent(pos.Line)
	if line == "" {
		return nil
	}

	// Find the function name and parameter index by scanning backwards from cursor
	funcName, activeParam := parseFunctionCall(line, pos.Character)
	if funcName == "" {
		return nil
	}

	// Check stdlib docs first
	if info, ok := stdlibDocs[funcName]; ok {
		sig := SignatureInformation{
			Label:         info.Signature,
			Documentation: info.Doc,
		}
		// Extract parameter names from signature
		sig.Parameters = extractParams(info.Signature)
		return &SignatureHelp{
			Signatures:      []SignatureInformation{sig},
			ActiveSignature: 0,
			ActiveParameter: activeParam,
		}
	}

	// Check user-defined functions
	if program != nil {
		for _, stmt := range program.Statements {
			if fn, ok := stmt.(*ast.FuncDefinition); ok {
				if fn.Name == funcName {
					label := fmt.Sprintf("to %s %s", fn.Name, strings.Join(fn.Params, ", "))
					sig := SignatureInformation{
						Label: label,
					}
					for _, p := range fn.Params {
						sig.Parameters = append(sig.Parameters, ParameterInformation{Label: p})
					}
					return &SignatureHelp{
						Signatures:      []SignatureInformation{sig},
						ActiveSignature: 0,
						ActiveParameter: activeParam,
					}
				}
			}
		}
	}

	return nil
}

// parseFunctionCall extracts the function name and active parameter index
// from a line of code at the given cursor position.
func parseFunctionCall(line string, cursorCol int) (string, int) {
	if cursorCol > len(line) {
		cursorCol = len(line)
	}

	// Walk backwards from cursor to find the opening paren
	depth := 0
	parenPos := -1
	for i := cursorCol - 1; i >= 0; i-- {
		ch := line[i]
		if ch == ')' {
			depth++
		} else if ch == '(' {
			if depth == 0 {
				parenPos = i
				break
			}
			depth--
		}
	}

	if parenPos < 0 {
		return "", 0
	}

	// Extract function name before the paren
	nameEnd := parenPos
	nameStart := nameEnd
	for nameStart > 0 && isIdentChar(line[nameStart-1]) {
		nameStart--
	}
	if nameStart == nameEnd {
		return "", 0
	}
	funcName := line[nameStart:nameEnd]

	// Count commas between paren and cursor to get active parameter
	activeParam := 0
	innerDepth := 0
	for i := parenPos + 1; i < cursorCol && i < len(line); i++ {
		ch := line[i]
		if ch == '(' {
			innerDepth++
		} else if ch == ')' {
			innerDepth--
		} else if ch == ',' && innerDepth == 0 {
			activeParam++
		}
	}

	return funcName, activeParam
}

func isIdentChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}

// extractParams parses parameter names from a signature string like "func(a, b, c) -> ret"
func extractParams(sig string) []ParameterInformation {
	start := strings.Index(sig, "(")
	end := strings.Index(sig, ")")
	if start < 0 || end < 0 || end <= start+1 {
		return nil
	}
	inner := sig[start+1 : end]
	parts := strings.Split(inner, ", ")
	var params []ParameterInformation
	for _, p := range parts {
		params = append(params, ParameterInformation{Label: strings.TrimSpace(p)})
	}
	return params
}
