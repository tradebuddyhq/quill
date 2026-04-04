package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilePathToRoute_Index(t *testing.T) {
	r := filePathToRoute("index")
	if r.Path != "/" {
		t.Errorf("expected '/', got %q", r.Path)
	}
	if len(r.Params) != 0 {
		t.Errorf("expected no params, got %v", r.Params)
	}
	if r.IsCatchAll {
		t.Error("expected not catch-all")
	}
}

func TestFilePathToRoute_Static(t *testing.T) {
	r := filePathToRoute("about")
	if r.Path != "/about" {
		t.Errorf("expected '/about', got %q", r.Path)
	}
}

func TestFilePathToRoute_Nested(t *testing.T) {
	r := filePathToRoute("blog/posts")
	if r.Path != "/blog/posts" {
		t.Errorf("expected '/blog/posts', got %q", r.Path)
	}
}

func TestFilePathToRoute_Dynamic(t *testing.T) {
	r := filePathToRoute("blog/[id]")
	if r.Path != "/blog/:id" {
		t.Errorf("expected '/blog/:id', got %q", r.Path)
	}
	if len(r.Params) != 1 || r.Params[0] != "id" {
		t.Errorf("expected params [id], got %v", r.Params)
	}
	if r.IsCatchAll {
		t.Error("expected not catch-all")
	}
}

func TestFilePathToRoute_CatchAll(t *testing.T) {
	r := filePathToRoute("blog/[...slug]")
	if r.Path != "/blog/*" {
		t.Errorf("expected '/blog/*', got %q", r.Path)
	}
	if len(r.Params) != 1 || r.Params[0] != "slug" {
		t.Errorf("expected params [slug], got %v", r.Params)
	}
	if !r.IsCatchAll {
		t.Error("expected catch-all")
	}
}

func TestFilePathToRoute_NestedIndex(t *testing.T) {
	r := filePathToRoute("blog/index")
	if r.Path != "/blog" {
		t.Errorf("expected '/blog', got %q", r.Path)
	}
}

func TestMatchRoute_Root(t *testing.T) {
	routes := []Route{
		{Path: "/", Params: nil},
		{Path: "/about", Params: nil},
	}
	r, params := MatchRoute(routes, "/")
	if r == nil {
		t.Fatal("expected match for /")
	}
	if r.Path != "/" {
		t.Errorf("expected '/', got %q", r.Path)
	}
	if len(params) != 0 {
		t.Errorf("expected no params, got %v", params)
	}
}

func TestMatchRoute_Static(t *testing.T) {
	routes := []Route{
		{Path: "/", Params: nil},
		{Path: "/about", Params: nil},
	}
	r, _ := MatchRoute(routes, "/about")
	if r == nil {
		t.Fatal("expected match for /about")
	}
	if r.Path != "/about" {
		t.Errorf("expected '/about', got %q", r.Path)
	}
}

func TestMatchRoute_Dynamic(t *testing.T) {
	routes := []Route{
		{Path: "/", Params: nil},
		{Path: "/blog/:id", Params: []string{"id"}},
	}
	r, params := MatchRoute(routes, "/blog/42")
	if r == nil {
		t.Fatal("expected match for /blog/42")
	}
	if params["id"] != "42" {
		t.Errorf("expected id=42, got %q", params["id"])
	}
}

func TestMatchRoute_CatchAll(t *testing.T) {
	routes := []Route{
		{Path: "/", Params: nil},
		{Path: "/docs/*", Params: []string{"slug"}, IsCatchAll: true},
	}
	r, params := MatchRoute(routes, "/docs/guide/getting-started")
	if r == nil {
		t.Fatal("expected match for /docs/guide/getting-started")
	}
	if params["slug"] != "guide/getting-started" {
		t.Errorf("expected slug=guide/getting-started, got %q", params["slug"])
	}
}

func TestMatchRoute_NoMatch(t *testing.T) {
	routes := []Route{
		{Path: "/", Params: nil},
		{Path: "/about", Params: nil},
	}
	r, _ := MatchRoute(routes, "/contact")
	if r != nil {
		t.Error("expected no match for /contact")
	}
}

func TestMatchRoute_TrailingSlash(t *testing.T) {
	routes := []Route{
		{Path: "/about", Params: nil},
	}
	r, _ := MatchRoute(routes, "/about/")
	if r == nil {
		t.Fatal("expected match for /about/ (trailing slash)")
	}
}

func TestScanPages(t *testing.T) {
	// Create temp directory structure
	dir := t.TempDir()
	pagesDir := filepath.Join(dir, "pages")
	os.MkdirAll(filepath.Join(pagesDir, "blog"), 0755)

	// Create test files
	os.WriteFile(filepath.Join(pagesDir, "index.quill"), []byte("-- index"), 0644)
	os.WriteFile(filepath.Join(pagesDir, "about.quill"), []byte("-- about"), 0644)
	os.WriteFile(filepath.Join(pagesDir, "blog", "[id].quill"), []byte("-- blog post"), 0644)
	os.WriteFile(filepath.Join(pagesDir, "blog", "[...slug].quill"), []byte("-- catch all"), 0644)

	routes := ScanPages(pagesDir)

	if len(routes) != 4 {
		t.Fatalf("expected 4 routes, got %d", len(routes))
	}

	// Check that static routes come before dynamic, catch-all last
	paths := make([]string, len(routes))
	for i, r := range routes {
		paths[i] = r.Path
	}

	// "/" and "/about" should be first (static), then "/blog/:id" (dynamic), then "/blog/*" (catch-all)
	foundCatchAll := false
	for _, r := range routes {
		if r.IsCatchAll {
			foundCatchAll = true
		}
		if foundCatchAll && !r.IsCatchAll {
			t.Error("catch-all route should be last")
		}
	}
}

func TestGenerateRouter(t *testing.T) {
	routes := []Route{
		{Path: "/", Params: nil},
		{Path: "/about", Params: nil},
		{Path: "/blog/:id", Params: []string{"id"}},
	}

	js := GenerateRouter(routes)

	if !strings.Contains(js, "__quill_routes") {
		t.Error("expected router to contain __quill_routes")
	}
	if !strings.Contains(js, "__quill_match") {
		t.Error("expected router to contain __quill_match")
	}
	if !strings.Contains(js, "__quill_navigate") {
		t.Error("expected router to contain __quill_navigate")
	}
	if !strings.Contains(js, "popstate") {
		t.Error("expected router to listen for popstate")
	}
	if !strings.Contains(js, "data-quill-link") {
		t.Error("expected router to handle quill links")
	}
	if !strings.Contains(js, "/blog/:id") {
		t.Error("expected router to contain /blog/:id route")
	}
}
