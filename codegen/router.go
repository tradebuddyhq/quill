package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Route represents a file-based route mapping.
type Route struct {
	Path       string   // URL path pattern e.g. "/blog/:id"
	FilePath   string   // File system path e.g. "pages/blog/[id].quill"
	Params     []string // Dynamic segment names e.g. ["id"]
	IsCatchAll bool     // True for [...slug] routes
}

// ScanPages walks a pages/ directory and builds a route table.
func ScanPages(dir string) []Route {
	var routes []Route

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".quill" {
			return nil
		}

		// Get relative path from the pages dir
		relPath, _ := filepath.Rel(dir, path)
		relPath = strings.TrimSuffix(relPath, ".quill")
		relPath = filepath.ToSlash(relPath)

		// Convert file path to URL route
		route := filePathToRoute(relPath)
		route.FilePath = path
		routes = append(routes, route)
		return nil
	})

	// Sort routes: static routes before dynamic, catch-all last
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].IsCatchAll != routes[j].IsCatchAll {
			return !routes[i].IsCatchAll
		}
		if len(routes[i].Params) != len(routes[j].Params) {
			return len(routes[i].Params) < len(routes[j].Params)
		}
		return routes[i].Path < routes[j].Path
	})

	return routes
}

// filePathToRoute converts a file path like "blog/[id]" to a Route.
func filePathToRoute(relPath string) Route {
	// Handle index routes
	if relPath == "index" {
		return Route{Path: "/"}
	}
	if strings.HasSuffix(relPath, "/index") {
		relPath = strings.TrimSuffix(relPath, "/index")
	}

	parts := strings.Split(relPath, "/")
	var urlParts []string
	var params []string
	isCatchAll := false

	for _, part := range parts {
		if strings.HasPrefix(part, "[...") && strings.HasSuffix(part, "]") {
			// Catch-all: [...slug]
			paramName := part[4 : len(part)-1]
			params = append(params, paramName)
			urlParts = append(urlParts, "*")
			isCatchAll = true
		} else if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			// Dynamic: [id]
			paramName := part[1 : len(part)-1]
			params = append(params, paramName)
			urlParts = append(urlParts, ":"+paramName)
		} else {
			urlParts = append(urlParts, part)
		}
	}

	return Route{
		Path:       "/" + strings.Join(urlParts, "/"),
		Params:     params,
		IsCatchAll: isCatchAll,
	}
}

// MatchRoute matches a URL path against the route table and returns the matching route
// and extracted parameters. Returns nil if no match.
func MatchRoute(routes []Route, urlPath string) (*Route, map[string]string) {
	urlPath = strings.TrimSuffix(urlPath, "/")
	if urlPath == "" {
		urlPath = "/"
	}

	for i := range routes {
		route := &routes[i]
		params := matchRoutePath(route, urlPath)
		if params != nil {
			return route, params
		}
	}
	return nil, nil
}

func matchRoutePath(route *Route, urlPath string) map[string]string {
	if route.Path == "/" && urlPath == "/" {
		return map[string]string{}
	}

	routeParts := strings.Split(strings.TrimPrefix(route.Path, "/"), "/")
	urlParts := strings.Split(strings.TrimPrefix(urlPath, "/"), "/")

	if route.IsCatchAll {
		// Catch-all: match prefix and capture rest
		if len(urlParts) < len(routeParts)-1 {
			return nil
		}
		params := map[string]string{}
		for i, rp := range routeParts {
			if rp == "*" {
				// Capture remaining parts
				catchAllParam := route.Params[len(route.Params)-1]
				params[catchAllParam] = strings.Join(urlParts[i:], "/")
				return params
			}
			if i >= len(urlParts) {
				return nil
			}
			if strings.HasPrefix(rp, ":") {
				paramName := rp[1:]
				params[paramName] = urlParts[i]
			} else if rp != urlParts[i] {
				return nil
			}
		}
		return params
	}

	if len(routeParts) != len(urlParts) {
		return nil
	}

	params := map[string]string{}
	for i, rp := range routeParts {
		if strings.HasPrefix(rp, ":") {
			paramName := rp[1:]
			params[paramName] = urlParts[i]
		} else if rp != urlParts[i] {
			return nil
		}
	}
	return params
}

// GenerateRouter generates JavaScript client-side router code from a route table.
func GenerateRouter(routes []Route) string {
	var out strings.Builder

	out.WriteString("// Quill Client-Side Router\n")
	out.WriteString("(function(global) {\n")
	out.WriteString("  'use strict';\n\n")

	// Route table
	out.WriteString("  var __quill_routes = [\n")
	for _, r := range routes {
		paramsJSON := "[]"
		if len(r.Params) > 0 {
			quoted := make([]string, len(r.Params))
			for i, p := range r.Params {
				quoted[i] = fmt.Sprintf("%q", p)
			}
			paramsJSON = "[" + strings.Join(quoted, ", ") + "]"
		}
		out.WriteString(fmt.Sprintf("    { path: %q, params: %s, catchAll: %v },\n",
			r.Path, paramsJSON, r.IsCatchAll))
	}
	out.WriteString("  ];\n\n")

	// Route matching function
	out.WriteString(`  function __quill_match(url) {
    var urlParts = url.replace(/\/$/, '').split('/').filter(Boolean);
    for (var i = 0; i < __quill_routes.length; i++) {
      var route = __quill_routes[i];
      var routeParts = route.path.replace(/\/$/, '').split('/').filter(Boolean);
      var params = {};
      var matched = true;

      if (route.catchAll) {
        for (var j = 0; j < routeParts.length; j++) {
          if (routeParts[j] === '*') {
            params[route.params[route.params.length - 1]] = urlParts.slice(j).join('/');
            return { route: route, params: params };
          } else if (routeParts[j].charAt(0) === ':') {
            if (j >= urlParts.length) { matched = false; break; }
            params[routeParts[j].slice(1)] = urlParts[j];
          } else if (j >= urlParts.length || routeParts[j] !== urlParts[j]) {
            matched = false; break;
          }
        }
      } else {
        if (route.path === '/' && url === '/') return { route: route, params: {} };
        if (routeParts.length !== urlParts.length) continue;
        for (var j = 0; j < routeParts.length; j++) {
          if (routeParts[j].charAt(0) === ':') {
            params[routeParts[j].slice(1)] = urlParts[j];
          } else if (routeParts[j] !== urlParts[j]) {
            matched = false; break;
          }
        }
      }
      if (matched) return { route: route, params: params };
    }
    return null;
  }

`)

	// Navigation function
	out.WriteString(`  function __quill_navigate(url, push) {
    if (push !== false) {
      history.pushState({}, '', url);
    }
    var match = __quill_match(url);
    if (match && global.__quill_on_navigate) {
      global.__quill_on_navigate(match.route, match.params);
    }
  }

  // Listen for popstate (back/forward)
  window.addEventListener('popstate', function() {
    __quill_navigate(location.pathname, false);
  });

  // Intercept link clicks
  document.addEventListener('click', function(e) {
    var el = e.target;
    while (el && el.tagName !== 'A') el = el.parentElement;
    if (el && el.hasAttribute('data-quill-link')) {
      e.preventDefault();
      __quill_navigate(el.getAttribute('href'));
    }
  });

  global.__quill_routes = __quill_routes;
  global.__quill_match = __quill_match;
  global.__quill_navigate = __quill_navigate;
`)

	out.WriteString("})(typeof window !== 'undefined' ? window : global);\n")
	return out.String()
}
