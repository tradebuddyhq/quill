package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"quill/codegen"
	"quill/lexer"
	"quill/parser"
	"quill/stdlib"
	"strings"
	"time"
)

// DevServer is a development server for Quill applications.
type DevServer struct {
	Port      int
	PagesDir  string
	PublicDir string
	Routes    []codegen.Route
}

// NewDevServer creates a new development server.
func NewDevServer(port int) *DevServer {
	return &DevServer{
		Port:      port,
		PagesDir:  "pages",
		PublicDir: "public",
	}
}

// Start starts the dev server.
func (s *DevServer) Start() error {
	// Initial route scan
	s.scanRoutes()

	mux := http.NewServeMux()

	// Serve static files from public/
	if info, err := os.Stat(s.PublicDir); err == nil && info.IsDir() {
		mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(s.PublicDir))))
	}

	// Handle all other routes
	mux.HandleFunc("/", s.handleRequest)

	addr := fmt.Sprintf(":%d", s.Port)
	fmt.Printf("Quill dev server running at http://localhost:%d\n", s.Port)
	fmt.Printf("  Pages directory: %s/\n", s.PagesDir)
	fmt.Printf("  Public directory: %s/\n", s.PublicDir)
	fmt.Println("  Press Ctrl+C to stop")

	return http.ListenAndServe(addr, mux)
}

func (s *DevServer) scanRoutes() {
	if info, err := os.Stat(s.PagesDir); err == nil && info.IsDir() {
		s.Routes = codegen.ScanPages(s.PagesDir)
		fmt.Printf("  Found %d route(s):\n", len(s.Routes))
		for _, r := range s.Routes {
			fmt.Printf("    %s -> %s\n", r.Path, r.FilePath)
		}
	} else {
		fmt.Println("  Warning: no pages/ directory found")
	}
}

func (s *DevServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Re-scan routes on every request (dev mode)
	s.scanRoutesQuiet()

	path := r.URL.Path

	// Try static files first
	staticPath := filepath.Join(s.PublicDir, path)
	if info, err := os.Stat(staticPath); err == nil && !info.IsDir() {
		http.ServeFile(w, r, staticPath)
		return
	}

	// Match route
	route, params := codegen.MatchRoute(s.Routes, path)
	if route == nil {
		http.NotFound(w, r)
		return
	}

	// Compile and render
	html, err := s.compileAndRender(route, params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Compilation error: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *DevServer) scanRoutesQuiet() {
	if info, err := os.Stat(s.PagesDir); err == nil && info.IsDir() {
		s.Routes = codegen.ScanPages(s.PagesDir)
	}
}

func (s *DevServer) compileAndRender(route *codegen.Route, params map[string]string) (string, error) {
	source, err := os.ReadFile(route.FilePath)
	if err != nil {
		return "", fmt.Errorf("could not read %s: %w", route.FilePath, err)
	}

	l := lexer.New(string(source))
	tokens, err := l.Tokenize()
	if err != nil {
		return "", fmt.Errorf("lexer error: %w", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	// Generate JavaScript (browser mode for client-side rendering)
	g := codegen.NewBrowser()
	js := g.Generate(program)

	// Build HTML page with embedded JS
	var html strings.Builder
	html.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	html.WriteString("  <meta charset=\"utf-8\">\n")
	html.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")

	// Add route params as data
	html.WriteString("  <script>\n")
	html.WriteString("    window.__QUILL_PARAMS__ = {")
	first := true
	for k, v := range params {
		if !first {
			html.WriteString(", ")
		}
		html.WriteString(fmt.Sprintf("%q: %q", k, v))
		first = false
	}
	html.WriteString("};\n")
	html.WriteString("  </script>\n")

	html.WriteString("</head>\n<body>\n")
	html.WriteString("  <div id=\"app\"></div>\n")
	html.WriteString("  <script>\n")
	html.WriteString(js)
	html.WriteString("\n  </script>\n")
	html.WriteString("</body>\n</html>\n")

	return html.String(), nil
}

// WatchAndReload sets up basic file watching (polling-based).
func (s *DevServer) WatchAndReload(interval time.Duration) {
	modTimes := make(map[string]time.Time)

	for {
		time.Sleep(interval)
		changed := false
		filepath.Walk(s.PagesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".quill" {
				return nil
			}
			if prev, ok := modTimes[path]; ok {
				if info.ModTime().After(prev) {
					changed = true
				}
			}
			modTimes[path] = info.ModTime()
			return nil
		})
		if changed {
			s.scanRoutes()
		}
	}
}

// GenerateFullPageHTML generates a full HTML page with the framework runtime and component JS.
func GenerateFullPageHTML(js string, title string) string {
	var html strings.Builder
	html.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	html.WriteString("  <meta charset=\"utf-8\">\n")
	html.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	if title != "" {
		html.WriteString(fmt.Sprintf("  <title>%s</title>\n", title))
	}
	html.WriteString("</head>\n<body>\n")
	html.WriteString("  <div id=\"app\"></div>\n")
	html.WriteString("  <script>\n")
	html.WriteString(stdlib.FrameworkRuntime)
	html.WriteString("\n")
	html.WriteString(js)
	html.WriteString("\n  </script>\n")
	html.WriteString("</body>\n</html>\n")
	return html.String()
}
