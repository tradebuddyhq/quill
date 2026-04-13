package lsp

import (
	"quill/ast"
	"strings"
	"testing"
)

// ============================================================
// Document Store Tests
// ============================================================

func TestDocumentStore_OpenAndGet(t *testing.T) {
	store := NewDocumentStore()
	uri := "file:///test.quill"
	content := "name is \"Alice\"\nsay name"

	store.Open(uri, 1, content)
	doc := store.Get(uri)

	if doc == nil {
		t.Fatal("expected document, got nil")
	}
	if doc.URI != uri {
		t.Errorf("URI = %q, want %q", doc.URI, uri)
	}
	if doc.Version != 1 {
		t.Errorf("Version = %d, want 1", doc.Version)
	}
	if doc.Content != content {
		t.Errorf("Content = %q, want %q", doc.Content, content)
	}
	if len(doc.Lines) != 2 {
		t.Errorf("Lines count = %d, want 2", len(doc.Lines))
	}
}

func TestDocumentStore_GetNonexistent(t *testing.T) {
	store := NewDocumentStore()
	doc := store.Get("file:///nonexistent.quill")
	if doc != nil {
		t.Errorf("expected nil for nonexistent document, got %v", doc)
	}
}

func TestDocumentStore_Update(t *testing.T) {
	store := NewDocumentStore()
	uri := "file:///test.quill"

	store.Open(uri, 1, "original content")
	store.Update(uri, 2, "updated content\nwith two lines")

	doc := store.Get(uri)
	if doc.Version != 2 {
		t.Errorf("Version = %d, want 2", doc.Version)
	}
	if doc.Content != "updated content\nwith two lines" {
		t.Errorf("Content not updated correctly")
	}
	if len(doc.Lines) != 2 {
		t.Errorf("Lines count = %d, want 2", len(doc.Lines))
	}
}

func TestDocumentStore_UpdateCreatesIfNotExists(t *testing.T) {
	store := NewDocumentStore()
	uri := "file:///new.quill"

	store.Update(uri, 1, "new content")
	doc := store.Get(uri)
	if doc == nil {
		t.Fatal("Update should create document if it does not exist")
	}
	if doc.Content != "new content" {
		t.Errorf("Content = %q, want %q", doc.Content, "new content")
	}
}

func TestDocumentStore_Close(t *testing.T) {
	store := NewDocumentStore()
	uri := "file:///test.quill"

	store.Open(uri, 1, "content")
	store.Close(uri)

	doc := store.Get(uri)
	if doc != nil {
		t.Errorf("expected nil after Close, got %v", doc)
	}
}

// ============================================================
// GetWordAtPosition Tests
// ============================================================

func TestGetWordAtPosition_MiddleOfWord(t *testing.T) {
	doc := makeDoc("name is \"Alice\"")
	// Cursor on 'a' of "name" (position 1)
	word := doc.GetWordAtPosition(Position{Line: 0, Character: 1})
	if word != "name" {
		t.Errorf("word = %q, want %q", word, "name")
	}
}

func TestGetWordAtPosition_StartOfWord(t *testing.T) {
	doc := makeDoc("name is \"Alice\"")
	word := doc.GetWordAtPosition(Position{Line: 0, Character: 0})
	if word != "name" {
		t.Errorf("word = %q, want %q", word, "name")
	}
}

func TestGetWordAtPosition_JustPastWord(t *testing.T) {
	doc := makeDoc("name is \"Alice\"")
	// Character 4 is just past "name" (on space). The cursor walks back to
	// find the preceding word, so it still returns "name".
	word := doc.GetWordAtPosition(Position{Line: 0, Character: 4})
	if word != "name" {
		t.Errorf("word = %q, want %q", word, "name")
	}
}

func TestGetWordAtPosition_OnKeyword(t *testing.T) {
	doc := makeDoc("name is \"Alice\"")
	// "is" starts at char 5
	word := doc.GetWordAtPosition(Position{Line: 0, Character: 5})
	if word != "is" {
		t.Errorf("word = %q, want %q", word, "is")
	}
}

func TestGetWordAtPosition_SecondLine(t *testing.T) {
	doc := makeDoc("name is \"Alice\"\nsay name")
	word := doc.GetWordAtPosition(Position{Line: 1, Character: 0})
	if word != "say" {
		t.Errorf("word = %q, want %q", word, "say")
	}
}

func TestGetWordAtPosition_OutOfBounds(t *testing.T) {
	doc := makeDoc("hello")
	// Negative line
	word := doc.GetWordAtPosition(Position{Line: -1, Character: 0})
	if word != "" {
		t.Errorf("negative line: word = %q, want empty", word)
	}
	// Line beyond end
	word = doc.GetWordAtPosition(Position{Line: 5, Character: 0})
	if word != "" {
		t.Errorf("line beyond end: word = %q, want empty", word)
	}
	// Character beyond end
	word = doc.GetWordAtPosition(Position{Line: 0, Character: 100})
	if word != "" {
		t.Errorf("char beyond end: word = %q, want empty", word)
	}
}

func TestGetWordAtPosition_Underscore(t *testing.T) {
	doc := makeDoc("my_var is 42")
	word := doc.GetWordAtPosition(Position{Line: 0, Character: 3})
	if word != "my_var" {
		t.Errorf("word = %q, want %q", word, "my_var")
	}
}

func TestGetWordAtPosition_EmptyLine(t *testing.T) {
	doc := makeDoc("")
	word := doc.GetWordAtPosition(Position{Line: 0, Character: 0})
	if word != "" {
		t.Errorf("word = %q, want empty", word)
	}
}

// ============================================================
// OffsetToPosition Tests
// ============================================================

func TestOffsetToPosition_Beginning(t *testing.T) {
	doc := makeDoc("hello\nworld")
	pos := doc.OffsetToPosition(0)
	if pos.Line != 0 || pos.Character != 0 {
		t.Errorf("pos = %+v, want {0, 0}", pos)
	}
}

func TestOffsetToPosition_MiddleOfFirstLine(t *testing.T) {
	doc := makeDoc("hello\nworld")
	pos := doc.OffsetToPosition(3)
	if pos.Line != 0 || pos.Character != 3 {
		t.Errorf("pos = %+v, want {0, 3}", pos)
	}
}

func TestOffsetToPosition_StartOfSecondLine(t *testing.T) {
	doc := makeDoc("hello\nworld")
	// offset 6 = after "hello\n" = start of "world"
	pos := doc.OffsetToPosition(6)
	if pos.Line != 1 || pos.Character != 0 {
		t.Errorf("pos = %+v, want {1, 0}", pos)
	}
}

func TestOffsetToPosition_MiddleOfSecondLine(t *testing.T) {
	doc := makeDoc("hello\nworld")
	pos := doc.OffsetToPosition(8)
	if pos.Line != 1 || pos.Character != 2 {
		t.Errorf("pos = %+v, want {1, 2}", pos)
	}
}

func TestOffsetToPosition_BeyondContent(t *testing.T) {
	doc := makeDoc("hi")
	pos := doc.OffsetToPosition(100)
	// Should stop at end of content
	if pos.Line != 0 || pos.Character != 2 {
		t.Errorf("pos = %+v, want {0, 2}", pos)
	}
}

// ============================================================
// LineToRange Tests
// ============================================================

func TestLineToRange(t *testing.T) {
	doc := makeDoc("first line\nsecond line")
	// LineToRange is 1-based
	r := doc.LineToRange(1)
	if r.Start.Line != 0 || r.Start.Character != 0 {
		t.Errorf("Start = %+v, want {0, 0}", r.Start)
	}
	if r.End.Line != 0 || r.End.Character != 10 {
		t.Errorf("End = %+v, want {0, 10}", r.End)
	}

	r2 := doc.LineToRange(2)
	if r2.Start.Line != 1 || r2.End.Character != 11 {
		t.Errorf("Line 2 range = %+v - %+v, want {1,0}-{1,11}", r2.Start, r2.End)
	}
}

// ============================================================
// Hover Provider Tests
// ============================================================

func TestHover_Keyword(t *testing.T) {
	doc := makeDoc("if x greater than 5\n\tsay x")
	hover := NewHoverProvider()

	result := hover.GetHover(doc, Position{Line: 0, Character: 0}, nil)
	if result == nil {
		t.Fatal("expected hover for keyword 'if', got nil")
	}
	if !strings.Contains(result.Contents.Value, "(keyword)") {
		t.Errorf("expected keyword marker, got %q", result.Contents.Value)
	}
	if !strings.Contains(result.Contents.Value, "**if**") {
		t.Errorf("expected keyword name 'if', got %q", result.Contents.Value)
	}
}

func TestHover_MultipleKeywords(t *testing.T) {
	doc := makeDoc("for each item in list\n\tsay item")
	hover := NewHoverProvider()

	keywords := []struct {
		char int
		word string
	}{
		{0, "for"},
		{4, "each"},
		{14, "in"},
	}

	for _, kw := range keywords {
		result := hover.GetHover(doc, Position{Line: 0, Character: kw.char}, nil)
		if result == nil {
			t.Errorf("expected hover for keyword %q, got nil", kw.word)
			continue
		}
		if !strings.Contains(result.Contents.Value, "(keyword)") {
			t.Errorf("keyword %q: expected keyword marker, got %q", kw.word, result.Contents.Value)
		}
	}
}

func TestHover_StdlibFunction(t *testing.T) {
	doc := makeDoc("result is length items")
	hover := NewHoverProvider()

	result := hover.GetHover(doc, Position{Line: 0, Character: 10}, nil)
	if result == nil {
		t.Fatal("expected hover for stdlib function 'length', got nil")
	}
	if !strings.Contains(result.Contents.Value, "(stdlib)") {
		t.Errorf("expected stdlib marker, got %q", result.Contents.Value)
	}
	if !strings.Contains(result.Contents.Value, "length(value) -> number") {
		t.Errorf("expected signature, got %q", result.Contents.Value)
	}
}

func TestHover_MultipleStdlibFunctions(t *testing.T) {
	hover := NewHoverProvider()

	tests := []struct {
		source    string
		char      int
		funcName  string
		signature string
	}{
		{"x is round 3.7", 5, "round", "round(number) -> number"},
		{"x is trim text", 5, "trim", "trim(text) -> text"},
		{"x is upper text", 5, "upper", "upper(text) -> text"},
		{"x is push list item", 5, "push", "push(list, item) -> list"},
		{"x is keys obj", 5, "keys", "keys(object) -> list"},
		{"x is typeOf val", 5, "typeOf", "typeOf(value) -> text"},
	}

	for _, tt := range tests {
		doc := makeDoc(tt.source)
		result := hover.GetHover(doc, Position{Line: 0, Character: tt.char}, nil)
		if result == nil {
			t.Errorf("expected hover for %q, got nil", tt.funcName)
			continue
		}
		if !strings.Contains(result.Contents.Value, tt.signature) {
			t.Errorf("%s: expected signature %q in %q", tt.funcName, tt.signature, result.Contents.Value)
		}
	}
}

func TestHover_UserDefinedVariable(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.AssignStatement{
				Name:  "greeting",
				Value: &ast.StringLiteral{Value: "hello"},
				Line:  1,
			},
			&ast.SayStatement{
				Value: &ast.Identifier{Name: "greeting"},
				Line:  2,
			},
		},
	}

	doc := makeDoc("greeting is \"hello\"\nsay greeting")
	hover := NewHoverProvider()

	// Hover on "greeting" in "say greeting"
	result := hover.GetHover(doc, Position{Line: 1, Character: 5}, program)
	if result == nil {
		t.Fatal("expected hover for user variable 'greeting', got nil")
	}
	if !strings.Contains(result.Contents.Value, "(variable)") {
		t.Errorf("expected variable marker, got %q", result.Contents.Value)
	}
	if !strings.Contains(result.Contents.Value, "text") {
		t.Errorf("expected type 'text' for string variable, got %q", result.Contents.Value)
	}
}

func TestHover_UserDefinedVariableTypes(t *testing.T) {
	hover := NewHoverProvider()

	tests := []struct {
		name     string
		value    ast.Expression
		expected string
	}{
		{"strVar", &ast.StringLiteral{Value: "hi"}, "text"},
		{"numVar", &ast.NumberLiteral{Value: 42}, "number"},
		{"boolVar", &ast.BoolLiteral{Value: true}, "boolean"},
		{"listVar", &ast.ListLiteral{Elements: nil}, "list"},
		{"objVar", &ast.ObjectLiteral{}, "object"},
		{"nothingVar", &ast.NothingLiteral{}, "nothing"},
		{"funcVar", &ast.LambdaExpr{Params: []string{"x"}}, "function"},
		{"unknownVar", &ast.Identifier{Name: "other"}, "any"},
	}

	for _, tt := range tests {
		program := &ast.Program{
			Statements: []ast.Statement{
				&ast.AssignStatement{
					Name:  tt.name,
					Value: tt.value,
					Line:  1,
				},
			},
		}
		doc := makeDoc(tt.name + " is something")
		result := hover.GetHover(doc, Position{Line: 0, Character: 0}, program)
		if result == nil {
			t.Errorf("%s: expected hover, got nil", tt.name)
			continue
		}
		if !strings.Contains(result.Contents.Value, tt.expected) {
			t.Errorf("%s: expected type %q in %q", tt.name, tt.expected, result.Contents.Value)
		}
	}
}

func TestHover_UserDefinedFunction(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.FuncDefinition{
				Name:       "greet",
				Params:     []string{"name"},
				ParamTypes: []string{"text"},
				ReturnType: "text",
				Body:       nil,
				Line:       1,
			},
		},
	}

	doc := makeDoc("greet \"Alice\"")
	hover := NewHoverProvider()

	result := hover.GetHover(doc, Position{Line: 0, Character: 0}, program)
	if result == nil {
		t.Fatal("expected hover for user function 'greet', got nil")
	}
	if !strings.Contains(result.Contents.Value, "(function)") {
		t.Errorf("expected function marker, got %q", result.Contents.Value)
	}
	if !strings.Contains(result.Contents.Value, "to greet name: text") {
		t.Errorf("expected signature with typed param, got %q", result.Contents.Value)
	}
	if !strings.Contains(result.Contents.Value, "-> text") {
		t.Errorf("expected return type, got %q", result.Contents.Value)
	}
}

func TestHover_UserDefinedFunctionNoTypes(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.FuncDefinition{
				Name:       "add",
				Params:     []string{"a", "b"},
				ParamTypes: []string{},
				ReturnType: "",
				Line:       1,
			},
		},
	}

	doc := makeDoc("add 1 2")
	hover := NewHoverProvider()

	result := hover.GetHover(doc, Position{Line: 0, Character: 0}, program)
	if result == nil {
		t.Fatal("expected hover for function 'add', got nil")
	}
	if !strings.Contains(result.Contents.Value, "to add a, b") {
		t.Errorf("expected untyped signature, got %q", result.Contents.Value)
	}
}

func TestHover_Class(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.DescribeStatement{
				Name:    "Animal",
				Extends: "LivingThing",
				Properties: []ast.AssignStatement{
					{Name: "name", Value: &ast.StringLiteral{Value: ""}},
					{Name: "age", Value: &ast.NumberLiteral{Value: 0}},
				},
				Methods: []ast.FuncDefinition{
					{Name: "speak", Params: nil, ParamTypes: nil, ReturnType: "text"},
				},
				Line: 1,
			},
		},
	}

	doc := makeDoc("Animal something")
	hover := NewHoverProvider()

	result := hover.GetHover(doc, Position{Line: 0, Character: 0}, program)
	if result == nil {
		t.Fatal("expected hover for class 'Animal', got nil")
	}
	if !strings.Contains(result.Contents.Value, "(class)") {
		t.Errorf("expected class marker, got %q", result.Contents.Value)
	}
	if !strings.Contains(result.Contents.Value, "Extends: `LivingThing`") {
		t.Errorf("expected extends info, got %q", result.Contents.Value)
	}
	if !strings.Contains(result.Contents.Value, "name") {
		t.Errorf("expected property 'name', got %q", result.Contents.Value)
	}
	if !strings.Contains(result.Contents.Value, "speak") {
		t.Errorf("expected method 'speak', got %q", result.Contents.Value)
	}
}

func TestHover_EmptyWord(t *testing.T) {
	doc := makeDoc("= + -")
	hover := NewHoverProvider()

	// Position on non-word characters — should return nil
	result := hover.GetHover(doc, Position{Line: 0, Character: 0}, nil)
	if result != nil {
		t.Errorf("expected nil hover on non-word char, got %+v", result)
	}
}

func TestHover_UnknownWord(t *testing.T) {
	doc := makeDoc("xyzzy foobar")
	hover := NewHoverProvider()

	result := hover.GetHover(doc, Position{Line: 0, Character: 0}, nil)
	if result != nil {
		t.Errorf("expected nil hover for unknown word, got %+v", result)
	}
}

func TestHover_NilProgram(t *testing.T) {
	doc := makeDoc("say \"hello\"")
	hover := NewHoverProvider()

	// "say" is a keyword, should work even with nil program
	result := hover.GetHover(doc, Position{Line: 0, Character: 0}, nil)
	if result == nil {
		t.Fatal("expected hover for keyword 'say' with nil program")
	}
	if !strings.Contains(result.Contents.Value, "(keyword)") {
		t.Errorf("expected keyword marker, got %q", result.Contents.Value)
	}
}

// ============================================================
// Completion Provider Tests
// ============================================================

func TestCompletion_Keywords(t *testing.T) {
	doc := makeDoc("i")
	comp := NewCompletionProvider()

	list := comp.GetCompletions(doc, Position{Line: 0, Character: 1}, nil)

	// Should include keywords starting with "i": if, in, is
	found := map[string]bool{}
	for _, item := range list.Items {
		if item.Kind == CompletionKindKeyword {
			found[item.Label] = true
		}
	}

	for _, kw := range []string{"if", "in", "is"} {
		if !found[kw] {
			t.Errorf("expected keyword completion %q in results", kw)
		}
	}
}

func TestCompletion_AllKeywordsWithEmptyPrefix(t *testing.T) {
	doc := makeDoc("")
	comp := NewCompletionProvider()

	list := comp.GetCompletions(doc, Position{Line: 0, Character: 0}, nil)

	kwCount := 0
	for _, item := range list.Items {
		if item.Kind == CompletionKindKeyword {
			kwCount++
		}
	}

	// Should have all keyword completions
	if kwCount != len(keywordCompletions) {
		t.Errorf("keyword count = %d, want %d", kwCount, len(keywordCompletions))
	}
}

func TestCompletion_StdlibFunctions(t *testing.T) {
	doc := makeDoc("le")
	comp := NewCompletionProvider()

	list := comp.GetCompletions(doc, Position{Line: 0, Character: 2}, nil)

	found := false
	for _, item := range list.Items {
		if item.Label == "length" && item.Kind == CompletionKindFunction {
			found = true
			if item.Detail != "length(value) -> number" {
				t.Errorf("length detail = %q, want signature", item.Detail)
			}
			break
		}
	}
	if !found {
		t.Error("expected stdlib completion for 'length'")
	}
}

func TestCompletion_StdlibFiltering(t *testing.T) {
	doc := makeDoc("tri")
	comp := NewCompletionProvider()

	list := comp.GetCompletions(doc, Position{Line: 0, Character: 3}, nil)

	found := false
	for _, item := range list.Items {
		if item.Label == "trim" && item.Kind == CompletionKindFunction {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'trim' in completions for prefix 'tri'")
	}

	// Should NOT include "push" for prefix "tri"
	for _, item := range list.Items {
		if item.Label == "push" {
			t.Error("did not expect 'push' in completions for prefix 'tri'")
		}
	}
}

func TestCompletion_UserDefinedFunction(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.FuncDefinition{
				Name:       "greet",
				Params:     []string{"name"},
				ParamTypes: []string{"text"},
				ReturnType: "text",
				Line:       1,
			},
		},
	}

	doc := makeDoc("gr")
	comp := NewCompletionProvider()

	list := comp.GetCompletions(doc, Position{Line: 0, Character: 2}, program)

	found := false
	for _, item := range list.Items {
		if item.Label == "greet" && item.Kind == CompletionKindFunction {
			found = true
			if !strings.Contains(item.Detail, "to greet name: text") {
				t.Errorf("expected function signature in detail, got %q", item.Detail)
			}
			if item.Documentation != "User-defined function" {
				t.Errorf("documentation = %q, want 'User-defined function'", item.Documentation)
			}
			break
		}
	}
	if !found {
		t.Error("expected user-defined function 'greet' in completions")
	}
}

func TestCompletion_UserDefinedVariable(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.AssignStatement{
				Name:  "counter",
				Value: &ast.NumberLiteral{Value: 0},
				Line:  1,
			},
		},
	}

	doc := makeDoc("co")
	comp := NewCompletionProvider()

	list := comp.GetCompletions(doc, Position{Line: 0, Character: 2}, program)

	found := false
	for _, item := range list.Items {
		if item.Label == "counter" && item.Kind == CompletionKindVariable {
			found = true
			if item.Detail != "number" {
				t.Errorf("detail = %q, want 'number'", item.Detail)
			}
			break
		}
	}
	if !found {
		t.Error("expected user-defined variable 'counter' in completions")
	}
}

func TestCompletion_UserDefinedClass(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.DescribeStatement{
				Name: "Animal",
				Line: 1,
			},
		},
	}

	doc := makeDoc("An")
	comp := NewCompletionProvider()

	list := comp.GetCompletions(doc, Position{Line: 0, Character: 2}, program)

	found := false
	for _, item := range list.Items {
		if item.Label == "Animal" {
			found = true
			if item.Detail != "class" {
				t.Errorf("detail = %q, want 'class'", item.Detail)
			}
			break
		}
	}
	if !found {
		t.Error("expected 'Animal' class in completions")
	}
}

func TestCompletion_EnumAndVariants(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefineStatement{
				Name: "Color",
				Variants: []ast.EnumVariant{
					{Name: "Red"},
					{Name: "Green"},
					{Name: "Blue"},
				},
				Line: 1,
			},
		},
	}

	doc := makeDoc("")
	comp := NewCompletionProvider()

	list := comp.GetCompletions(doc, Position{Line: 0, Character: 0}, program)

	names := map[string]bool{}
	for _, item := range list.Items {
		names[item.Label] = true
	}

	for _, name := range []string{"Color", "Red", "Green", "Blue"} {
		if !names[name] {
			t.Errorf("expected %q in completions", name)
		}
	}
}

func TestCompletion_ForEachVariable(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ForEachStatement{
				Variable: "item",
				Iterable: &ast.Identifier{Name: "items"},
				Line:     1,
			},
		},
	}

	doc := makeDoc("it")
	comp := NewCompletionProvider()

	list := comp.GetCompletions(doc, Position{Line: 0, Character: 2}, program)

	found := false
	for _, item := range list.Items {
		if item.Label == "item" && item.Kind == CompletionKindVariable {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected loop variable 'item' in completions")
	}
}

func TestCompletion_Snippets(t *testing.T) {
	doc := makeDoc("if")
	comp := NewCompletionProvider()

	list := comp.GetCompletions(doc, Position{Line: 0, Character: 2}, nil)

	found := false
	for _, item := range list.Items {
		if item.Label == "if-block" && item.Kind == CompletionKindSnippet {
			found = true
			if item.InsertTextFmt != InsertTextFormatSnippet {
				t.Errorf("snippet InsertTextFmt = %d, want %d", item.InsertTextFmt, InsertTextFormatSnippet)
			}
			break
		}
	}
	if !found {
		t.Error("expected 'if-block' snippet in completions")
	}
}

func TestCompletion_NoDuplicateUserSymbols(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.AssignStatement{Name: "x", Value: &ast.NumberLiteral{Value: 1}, Line: 1},
			&ast.AssignStatement{Name: "x", Value: &ast.NumberLiteral{Value: 2}, Line: 2},
		},
	}

	doc := makeDoc("x")
	comp := NewCompletionProvider()

	list := comp.GetCompletions(doc, Position{Line: 0, Character: 1}, program)

	count := 0
	for _, item := range list.Items {
		if item.Label == "x" && item.Kind == CompletionKindVariable {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 completion for 'x', got %d", count)
	}
}

func TestCompletion_EmptyPrefixIncludesAll(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.FuncDefinition{Name: "myFunc", Params: nil, Line: 1},
			&ast.AssignStatement{Name: "myVar", Value: &ast.NumberLiteral{Value: 1}, Line: 2},
		},
	}

	doc := makeDoc("")
	comp := NewCompletionProvider()

	list := comp.GetCompletions(doc, Position{Line: 0, Character: 0}, program)

	foundFunc := false
	foundVar := false
	for _, item := range list.Items {
		if item.Label == "myFunc" {
			foundFunc = true
		}
		if item.Label == "myVar" {
			foundVar = true
		}
	}
	if !foundFunc {
		t.Error("expected 'myFunc' in empty prefix completions")
	}
	if !foundVar {
		t.Error("expected 'myVar' in empty prefix completions")
	}
}

func TestCompletion_OutOfBoundsPosition(t *testing.T) {
	doc := makeDoc("hello")
	comp := NewCompletionProvider()

	// Should not panic on out-of-bounds
	list := comp.GetCompletions(doc, Position{Line: 99, Character: 0}, nil)
	if list.Items == nil {
		t.Error("expected non-nil items slice")
	}
}

// ============================================================
// GetLineContent Tests
// ============================================================

func TestGetLineContent(t *testing.T) {
	doc := makeDoc("first\nsecond\nthird")
	if doc.GetLineContent(0) != "first" {
		t.Errorf("line 0 = %q, want 'first'", doc.GetLineContent(0))
	}
	if doc.GetLineContent(1) != "second" {
		t.Errorf("line 1 = %q, want 'second'", doc.GetLineContent(1))
	}
	if doc.GetLineContent(2) != "third" {
		t.Errorf("line 2 = %q, want 'third'", doc.GetLineContent(2))
	}
	if doc.GetLineContent(5) != "" {
		t.Errorf("out of range line = %q, want empty", doc.GetLineContent(5))
	}
	if doc.GetLineContent(-1) != "" {
		t.Errorf("negative line = %q, want empty", doc.GetLineContent(-1))
	}
}

// ============================================================
// Helpers
// ============================================================

func makeDoc(content string) *Document {
	return &Document{
		URI:     "file:///test.quill",
		Version: 1,
		Content: content,
		Lines:   strings.Split(content, "\n"),
	}
}
