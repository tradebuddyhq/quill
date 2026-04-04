package lexer

import (
	"testing"
)

func TestTokenizeEmpty(t *testing.T) {
	l := New("")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) == 0 {
		t.Fatal("expected at least EOF token")
	}
	if tokens[len(tokens)-1].Type != TOKEN_EOF {
		t.Error("last token should be EOF")
	}
}

func TestTokenizeAssignment(t *testing.T) {
	l := New(`name is "hello"`)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []struct {
		typ TokenType
		val string
	}{
		{TOKEN_IDENT, "name"},
		{TOKEN_IS, "is"},
		{TOKEN_STRING, "hello"},
		{TOKEN_NEWLINE, "\\n"},
		{TOKEN_EOF, ""},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.typ {
			t.Errorf("token %d: expected type %s, got %s", i, exp.typ, tokens[i].Type)
		}
		if tokens[i].Value != exp.val {
			t.Errorf("token %d: expected value %q, got %q", i, exp.val, tokens[i].Value)
		}
	}
}

func TestTokenizeNumber(t *testing.T) {
	l := New("age is 25")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Type != TOKEN_IDENT || tokens[0].Value != "age" {
		t.Errorf("expected IDENT 'age', got %s %q", tokens[0].Type, tokens[0].Value)
	}
	if tokens[2].Type != TOKEN_NUMBER || tokens[2].Value != "25" {
		t.Errorf("expected NUMBER '25', got %s %q", tokens[2].Type, tokens[2].Value)
	}
}

func TestTokenizeDecimalNumber(t *testing.T) {
	l := New("pi is 3.14")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[2].Type != TOKEN_NUMBER || tokens[2].Value != "3.14" {
		t.Errorf("expected NUMBER '3.14', got %s %q", tokens[2].Type, tokens[2].Value)
	}
}

func TestTokenizeKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"is", TOKEN_IS},
		{"are", TOKEN_ARE},
		{"say", TOKEN_SAY},
		{"if", TOKEN_IF},
		{"otherwise", TOKEN_OTHERWISE},
		{"for", TOKEN_FOR},
		{"each", TOKEN_EACH},
		{"in", TOKEN_IN},
		{"to", TOKEN_TO},
		{"give", TOKEN_GIVE},
		{"back", TOKEN_BACK},
		{"and", TOKEN_AND},
		{"or", TOKEN_OR},
		{"not", TOKEN_NOT},
		{"greater", TOKEN_GREATER},
		{"less", TOKEN_LESS},
		{"than", TOKEN_THAN},
		{"equal", TOKEN_EQUAL},
		{"contains", TOKEN_CONTAINS},
		{"while", TOKEN_WHILE},
		{"use", TOKEN_USE},
		{"test", TOKEN_TEST},
		{"expect", TOKEN_EXPECT},
		{"yes", TOKEN_YES},
		{"no", TOKEN_NO},
		{"true", TOKEN_YES},
		{"false", TOKEN_NO},
		{"describe", TOKEN_DESCRIBE},
		{"new", TOKEN_NEW},
		{"my", TOKEN_MY},
		{"await", TOKEN_AWAIT},
		{"as", TOKEN_AS},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %s for %q, got %s", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeOperators(t *testing.T) {
	l := New("1 + 2 - 3 * 4 / 5 % 6")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedOps := []struct {
		typ TokenType
		val string
	}{
		{TOKEN_NUMBER, "1"},
		{TOKEN_PLUS, "+"},
		{TOKEN_NUMBER, "2"},
		{TOKEN_MINUS, "-"},
		{TOKEN_NUMBER, "3"},
		{TOKEN_STAR, "*"},
		{TOKEN_NUMBER, "4"},
		{TOKEN_SLASH, "/"},
		{TOKEN_NUMBER, "5"},
		{TOKEN_MODULO, "%"},
		{TOKEN_NUMBER, "6"},
	}

	for i, exp := range expectedOps {
		if tokens[i].Type != exp.typ {
			t.Errorf("token %d: expected %s, got %s", i, exp.typ, tokens[i].Type)
		}
	}
}

func TestTokenizeDelimiters(t *testing.T) {
	l := New("fn(a, b) [1, 2]")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedTypes := []TokenType{
		TOKEN_IDENT, TOKEN_LPAREN, TOKEN_IDENT, TOKEN_COMMA,
		TOKEN_IDENT, TOKEN_RPAREN, TOKEN_LBRACKET, TOKEN_NUMBER,
		TOKEN_COMMA, TOKEN_NUMBER, TOKEN_RBRACKET,
	}

	for i, exp := range expectedTypes {
		if i >= len(tokens) {
			t.Fatalf("ran out of tokens at index %d", i)
		}
		if tokens[i].Type != exp {
			t.Errorf("token %d: expected %s, got %s (%q)", i, exp, tokens[i].Type, tokens[i].Value)
		}
	}
}

func TestTokenizeString(t *testing.T) {
	l := New(`say "Hello, world!"`)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Type != TOKEN_SAY {
		t.Errorf("expected SAY, got %s", tokens[0].Type)
	}
	if tokens[1].Type != TOKEN_STRING || tokens[1].Value != "Hello, world!" {
		t.Errorf("expected STRING 'Hello, world!', got %s %q", tokens[1].Type, tokens[1].Value)
	}
}

func TestTokenizeStringWithEscape(t *testing.T) {
	l := New(`say "He said \"hi\""`)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[1].Type != TOKEN_STRING {
		t.Errorf("expected STRING, got %s", tokens[1].Type)
	}
}

func TestTokenizeUnterminatedString(t *testing.T) {
	l := New(`say "hello`)
	_, err := l.Tokenize()
	if err == nil {
		t.Error("expected error for unterminated string")
	}
}

func TestTokenizeComment(t *testing.T) {
	l := New("x is 5 -- this is a comment\ny is 10")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not contain any comment tokens - comments are skipped
	for _, tok := range tokens {
		if tok.Value == "this" || tok.Value == "comment" {
			t.Error("comment content should not appear in tokens")
		}
	}

	// Should have both assignments
	identCount := 0
	for _, tok := range tokens {
		if tok.Type == TOKEN_IDENT {
			identCount++
		}
	}
	if identCount != 2 {
		t.Errorf("expected 2 identifiers (x, y), got %d", identCount)
	}
}

func TestTokenizeIndentation(t *testing.T) {
	src := "if x:\n  say x\n  say y\nz is 1\n"
	l := New(src)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasIndent := false
	hasDedent := false
	for _, tok := range tokens {
		if tok.Type == TOKEN_INDENT {
			hasIndent = true
		}
		if tok.Type == TOKEN_DEDENT {
			hasDedent = true
		}
	}

	if !hasIndent {
		t.Error("expected INDENT token")
	}
	if !hasDedent {
		t.Error("expected DEDENT token")
	}
}

func TestTokenizeNestedIndentation(t *testing.T) {
	src := "if x:\n  if y:\n    say z\n"
	l := New(src)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	indentCount := 0
	dedentCount := 0
	for _, tok := range tokens {
		if tok.Type == TOKEN_INDENT {
			indentCount++
		}
		if tok.Type == TOKEN_DEDENT {
			dedentCount++
		}
	}

	if indentCount != 2 {
		t.Errorf("expected 2 INDENT tokens, got %d", indentCount)
	}
	if dedentCount != 2 {
		t.Errorf("expected 2 DEDENT tokens, got %d", dedentCount)
	}
}

func TestTokenizeBooleans(t *testing.T) {
	l := New("active is yes\ndone is no")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, tok := range tokens {
		if tok.Type == TOKEN_YES && tok.Value == "yes" {
			found = true
		}
	}
	if !found {
		t.Error("expected YES token")
	}
}

func TestTokenizeDot(t *testing.T) {
	l := New("obj.field")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Type != TOKEN_IDENT || tokens[0].Value != "obj" {
		t.Errorf("expected IDENT 'obj', got %s %q", tokens[0].Type, tokens[0].Value)
	}
	if tokens[1].Type != TOKEN_DOT {
		t.Errorf("expected DOT, got %s", tokens[1].Type)
	}
	if tokens[2].Type != TOKEN_IDENT || tokens[2].Value != "field" {
		t.Errorf("expected IDENT 'field', got %s %q", tokens[2].Type, tokens[2].Value)
	}
}

func TestTokenizeListLiteral(t *testing.T) {
	l := New(`items are [1, "two", 3]`)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundBrackets := false
	for _, tok := range tokens {
		if tok.Type == TOKEN_LBRACKET {
			foundBrackets = true
		}
	}
	if !foundBrackets {
		t.Error("expected bracket tokens for list literal")
	}
}

func TestTokenizeUnexpectedCharacter(t *testing.T) {
	l := New("x is @invalid")
	_, err := l.Tokenize()
	if err == nil {
		t.Error("expected error for unexpected character '@'")
	}
}

func TestTokenizeCRLFLineEndings(t *testing.T) {
	l := New("x is 1\r\ny is 2\r\n")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	identCount := 0
	for _, tok := range tokens {
		if tok.Type == TOKEN_IDENT {
			identCount++
		}
	}
	if identCount != 2 {
		t.Errorf("expected 2 identifiers with CRLF line endings, got %d", identCount)
	}
}

func TestTokenizeLineNumbers(t *testing.T) {
	l := New("x is 1\ny is 2\nz is 3\n")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First token 'x' should be on line 1
	if tokens[0].Line != 1 {
		t.Errorf("expected 'x' on line 1, got line %d", tokens[0].Line)
	}

	// Find 'y' - should be on line 2
	for _, tok := range tokens {
		if tok.Type == TOKEN_IDENT && tok.Value == "y" {
			if tok.Line != 2 {
				t.Errorf("expected 'y' on line 2, got line %d", tok.Line)
			}
		}
		if tok.Type == TOKEN_IDENT && tok.Value == "z" {
			if tok.Line != 3 {
				t.Errorf("expected 'z' on line 3, got line %d", tok.Line)
			}
		}
	}
}

func TestTokenizeBlankLines(t *testing.T) {
	l := New("x is 1\n\n\ny is 2\n")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	identCount := 0
	for _, tok := range tokens {
		if tok.Type == TOKEN_IDENT {
			identCount++
		}
	}
	if identCount != 2 {
		t.Errorf("expected 2 identifiers (blank lines should be skipped), got %d", identCount)
	}
}

func TestTokenizeForEach(t *testing.T) {
	l := New("for each item in items:\n  say item\n")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Type != TOKEN_FOR {
		t.Errorf("expected FOR, got %s", tokens[0].Type)
	}
	if tokens[1].Type != TOKEN_EACH {
		t.Errorf("expected EACH, got %s", tokens[1].Type)
	}
}

func TestTokenizeFunction(t *testing.T) {
	l := New("to add a b:\n  give back a + b\n")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Type != TOKEN_TO {
		t.Errorf("expected TO, got %s", tokens[0].Type)
	}
}

func TestTokenizeDescribe(t *testing.T) {
	l := New("describe Dog:\n  name is \"\"\n")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Type != TOKEN_DESCRIBE {
		t.Errorf("expected DESCRIBE, got %s", tokens[0].Type)
	}
}

func TestTokenizeAwait(t *testing.T) {
	l := New(`data is await fetchJSON("url")`)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundAwait := false
	for _, tok := range tokens {
		if tok.Type == TOKEN_AWAIT {
			foundAwait = true
		}
	}
	if !foundAwait {
		t.Error("expected AWAIT token")
	}
}

func TestTokenizeUseAs(t *testing.T) {
	l := New(`use "express" as app`)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Type != TOKEN_USE {
		t.Errorf("expected USE, got %s", tokens[0].Type)
	}
	if tokens[1].Type != TOKEN_STRING || tokens[1].Value != "express" {
		t.Errorf("expected STRING 'express', got %s %q", tokens[1].Type, tokens[1].Value)
	}
	if tokens[2].Type != TOKEN_AS {
		t.Errorf("expected AS, got %s", tokens[2].Type)
	}
	if tokens[3].Type != TOKEN_IDENT || tokens[3].Value != "app" {
		t.Errorf("expected IDENT 'app', got %s %q", tokens[3].Type, tokens[3].Value)
	}
}

func TestTokenizeTestExpect(t *testing.T) {
	src := "test \"math works\":\n  expect 1 + 1 is 2\n"
	l := New(src)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Type != TOKEN_TEST {
		t.Errorf("expected TEST, got %s", tokens[0].Type)
	}
}

func TestTokenizeMyKeyword(t *testing.T) {
	l := New("my.name")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Type != TOKEN_MY {
		t.Errorf("expected MY, got %s", tokens[0].Type)
	}
	if tokens[1].Type != TOKEN_DOT {
		t.Errorf("expected DOT, got %s", tokens[1].Type)
	}
}

func TestTokenTypeString(t *testing.T) {
	if TOKEN_IS.String() != "is" {
		t.Errorf("expected 'is', got %q", TOKEN_IS.String())
	}
	if TOKEN_SAY.String() != "say" {
		t.Errorf("expected 'say', got %q", TOKEN_SAY.String())
	}
	if TOKEN_EOF.String() != "end of file" {
		t.Errorf("expected 'end of file', got %q", TOKEN_EOF.String())
	}
}

func TestTokenString(t *testing.T) {
	tok := Token{Type: TOKEN_IDENT, Value: "hello", Line: 5}
	s := tok.String()
	if s != `Token(name, "hello", line 5)` {
		t.Errorf("unexpected token string: %s", s)
	}
}

func TestTokenizeIdentifiersWithUnderscores(t *testing.T) {
	l := New("my_var is 10\n_private is 20\n")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Type != TOKEN_IDENT || tokens[0].Value != "my_var" {
		t.Errorf("expected IDENT 'my_var', got %s %q", tokens[0].Type, tokens[0].Value)
	}
}

func TestTokenizeComplexProgram(t *testing.T) {
	src := `name is "Sarah"
age is 25

if age is greater than 18:
  say "You are an adult"
otherwise:
  say "You are young"

to add a b:
  give back a + b

result is add(10, 20)
say "10 + 20 = {result}"
`
	l := New(src)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error tokenizing complex program: %v", err)
	}

	// Just verify it tokenizes successfully and has reasonable token count
	if len(tokens) < 30 {
		t.Errorf("expected at least 30 tokens for complex program, got %d", len(tokens))
	}

	// Verify EOF
	if tokens[len(tokens)-1].Type != TOKEN_EOF {
		t.Error("last token should be EOF")
	}
}
