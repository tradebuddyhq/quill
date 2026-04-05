package tools

import (
	"strings"
	"testing"
)

func TestParsePromptBlog(t *testing.T) {
	templateType, _ := ParsePrompt("a blog with auth and comments")
	if templateType != "blog" {
		t.Errorf("expected 'blog', got %q", templateType)
	}
}

func TestParsePromptAPI(t *testing.T) {
	templateType, _ := ParsePrompt("a REST API for my app")
	if templateType != "api" {
		t.Errorf("expected 'api', got %q", templateType)
	}
}

func TestParsePromptChat(t *testing.T) {
	templateType, _ := ParsePrompt("a realtime chat application")
	if templateType != "chat" {
		t.Errorf("expected 'chat', got %q", templateType)
	}
}

func TestParsePromptCRUD(t *testing.T) {
	templateType, param := ParsePrompt("crud product")
	if templateType != "crud" {
		t.Errorf("expected 'crud', got %q", templateType)
	}
	if param != "product" {
		t.Errorf("expected param 'product', got %q", param)
	}
}

func TestParsePromptAuth(t *testing.T) {
	templateType, _ := ParsePrompt("authentication system")
	if templateType != "auth" {
		t.Errorf("expected 'auth', got %q", templateType)
	}
}

func TestParsePromptDashboard(t *testing.T) {
	templateType, _ := ParsePrompt("admin dashboard")
	if templateType != "dashboard" {
		t.Errorf("expected 'dashboard', got %q", templateType)
	}
}

func TestParsePromptDefault(t *testing.T) {
	templateType, _ := ParsePrompt("something unknown")
	if templateType != "api" {
		t.Errorf("expected default 'api', got %q", templateType)
	}
}

func TestGenerateBlog(t *testing.T) {
	template := GenerateBlog()

	if template.Name != "blog" {
		t.Errorf("expected name 'blog', got %q", template.Name)
	}
	if len(template.Files) == 0 {
		t.Error("expected at least one file")
	}

	// Check app.quill content
	found := false
	for _, f := range template.Files {
		if f.Path == "app.quill" {
			found = true
			if !strings.Contains(f.Content, "server:") {
				t.Error("blog template missing server block")
			}
			if !strings.Contains(f.Content, "model Post:") {
				t.Error("blog template missing Post model")
			}
			if !strings.Contains(f.Content, "model Comment:") {
				t.Error("blog template missing Comment model")
			}
		}
	}
	if !found {
		t.Error("blog template missing app.quill")
	}
}

func TestGenerateAPI(t *testing.T) {
	template := GenerateAPI()

	if template.Name != "api" {
		t.Errorf("expected name 'api', got %q", template.Name)
	}
	if len(template.Files) == 0 {
		t.Error("expected at least one file")
	}

	for _, f := range template.Files {
		if f.Path == "app.quill" {
			if !strings.Contains(f.Content, "route get") {
				t.Error("API template missing GET route")
			}
			if !strings.Contains(f.Content, "route post") {
				t.Error("API template missing POST route")
			}
		}
	}
}

func TestGenerateCRUD(t *testing.T) {
	template := GenerateCRUD("product")

	if template.Name != "crud-product" {
		t.Errorf("expected name 'crud-product', got %q", template.Name)
	}
	if len(template.Files) == 0 {
		t.Error("expected at least one file")
	}

	for _, f := range template.Files {
		if f.Path == "app.quill" {
			if !strings.Contains(f.Content, "model Product:") {
				t.Error("CRUD template missing Product model")
			}
			if !strings.Contains(f.Content, "/api/products") {
				t.Error("CRUD template missing products route")
			}
		}
	}
}

func TestGenerateAppFromPrompt(t *testing.T) {
	gen := NewAppGenerator()
	template := gen.Generate("build me a blog")

	if template.Name != "blog" {
		t.Errorf("expected 'blog' template, got %q", template.Name)
	}
}
