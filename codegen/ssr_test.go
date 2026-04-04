package codegen

import (
	"quill/ast"
	"strings"
	"testing"
)

func TestGenerateSSR_BasicComponent(t *testing.T) {
	comp := &ast.ComponentStatement{
		Name: "Page",
		States: []ast.StateDeclaration{
			{Name: "title", Value: &ast.StringLiteral{Value: "Hello"}},
		},
		RenderBody: []ast.RenderElement{
			{
				Tag:      "h1",
				Props:    map[string]ast.Expression{},
				Children: []ast.RenderNode{{Text: &ast.StringLiteral{Value: "Hello World"}}},
			},
		},
	}

	hash := ComponentHash("Page")
	result := GenerateSSR(comp, hash)

	if !strings.Contains(result, "__ssr_Page") {
		t.Error("expected SSR function name __ssr_Page")
	}
	if !strings.Contains(result, "__QUILL_DATA__") {
		t.Error("expected data embedding with __QUILL_DATA__")
	}
	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("expected full HTML page")
	}
	if !strings.Contains(result, "<h1") {
		t.Error("expected h1 tag in output")
	}
}

func TestGenerateSSR_WithLoader(t *testing.T) {
	comp := &ast.ComponentStatement{
		Name: "Page",
		Loader: &ast.LoadFunction{
			Param: "request",
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Value: &ast.ObjectLiteral{
						Keys:   []string{"title"},
						Values: []ast.Expression{&ast.StringLiteral{Value: "Test"}},
					},
				},
			},
		},
		RenderBody: []ast.RenderElement{
			{
				Tag:      "div",
				Props:    map[string]ast.Expression{},
				Children: nil,
			},
		},
	}

	hash := ComponentHash("Page")
	result := GenerateSSR(comp, hash)

	if !strings.Contains(result, "async function __ssr_Page") {
		t.Error("expected async SSR function")
	}
	if !strings.Contains(result, "request") {
		t.Error("expected request parameter in loader")
	}
}

func TestGenerateHydration(t *testing.T) {
	comp := &ast.ComponentStatement{
		Name: "App",
		RenderBody: []ast.RenderElement{
			{Tag: "div", Props: map[string]ast.Expression{}},
		},
	}

	hash := ComponentHash("App")
	result := GenerateHydration(comp, hash)

	if !strings.Contains(result, "__QUILL_DATA__") {
		t.Error("expected hydration to read __QUILL_DATA__")
	}
	if !strings.Contains(result, "QuillComponent") {
		t.Error("expected hydration to create QuillComponent")
	}
	if !strings.Contains(result, "__update") {
		t.Error("expected hydration to call __update")
	}
}

func TestGenerateScopedCSS(t *testing.T) {
	styles := &ast.StyleBlock{
		Rules: []ast.CSSRule{
			{
				Selector: ".card",
				Properties: map[string]string{
					"padding":       "1rem",
					"border-radius": "8px",
				},
			},
			{
				Selector: ".card:hover",
				Properties: map[string]string{
					"box-shadow": "0 2px 8px rgba(0,0,0,0.1)",
				},
			},
		},
	}

	hash := "abc123"
	css := GenerateScopedCSS(styles, hash)

	if !strings.Contains(css, "[data-q-abc123]") {
		t.Error("expected scoped attribute selector")
	}
	if !strings.Contains(css, ".card") {
		t.Error("expected .card selector")
	}
	if !strings.Contains(css, "padding: 1rem") {
		t.Error("expected padding property")
	}
}

func TestComponentHash(t *testing.T) {
	hash1 := ComponentHash("ComponentA")
	hash2 := ComponentHash("ComponentB")

	if hash1 == hash2 {
		t.Error("expected different hashes for different component names")
	}
	if len(hash1) != 6 {
		t.Errorf("expected 6-char hash, got %d chars: %s", len(hash1), hash1)
	}
}

func TestGenerateHeadHTML(t *testing.T) {
	head := &ast.HeadBlock{
		Entries: []ast.HeadEntry{
			{Tag: "title", Text: "My App"},
			{Tag: "meta", Attrs: map[string]string{"name": "description", "content": "A Quill app"}},
			{Tag: "link", Attrs: map[string]string{"rel": "stylesheet", "href": "/style.css"}},
		},
	}

	html := GenerateHeadHTML(head)

	if !strings.Contains(html, "<title>My App</title>") {
		t.Error("expected title tag")
	}
	if !strings.Contains(html, "<meta") {
		t.Error("expected meta tag")
	}
	if !strings.Contains(html, "description") {
		t.Error("expected description attribute")
	}
	if !strings.Contains(html, "<link") {
		t.Error("expected link tag")
	}
	if !strings.Contains(html, "stylesheet") {
		t.Error("expected stylesheet rel")
	}
}

func TestGenerateHeadJS(t *testing.T) {
	head := &ast.HeadBlock{
		Entries: []ast.HeadEntry{
			{Tag: "title", Text: "Test Page"},
			{Tag: "meta", Attrs: map[string]string{"name": "viewport", "content": "width=device-width"}},
		},
	}

	js := GenerateHeadJS(head)

	if !strings.Contains(js, "document.title") {
		t.Error("expected document.title assignment")
	}
	if !strings.Contains(js, "Test Page") {
		t.Error("expected title text")
	}
	if !strings.Contains(js, "createElement") {
		t.Error("expected meta element creation")
	}
}

func TestScopeSelector(t *testing.T) {
	tests := []struct {
		sel    string
		hash   string
		expect string
	}{
		{".card", "abc", ".card[data-q-abc]"},
		{".card:hover", "abc", ".card:hover[data-q-abc]"},
		{"h1", "abc", "h1[data-q-abc]"},
		{".parent .child", "abc", ".parent[data-q-abc] .child"},
	}

	for _, tt := range tests {
		result := scopeSelector(tt.sel, tt.hash)
		if result != tt.expect {
			t.Errorf("scopeSelector(%q, %q) = %q, want %q", tt.sel, tt.hash, result, tt.expect)
		}
	}
}
