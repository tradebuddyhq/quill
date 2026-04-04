package codegen

import (
	"crypto/sha256"
	"fmt"
	"quill/ast"
	"strings"
)

// GenerateSSR generates Node.js server-side rendering code for a component.
// It produces a function that:
// 1. Runs the load function on the server
// 2. Renders the component to an HTML string
// 3. Embeds the data for client hydration
func GenerateSSR(comp *ast.ComponentStatement, cssHash string) string {
	var out strings.Builder

	out.WriteString(fmt.Sprintf("// SSR for component: %s\n", comp.Name))
	out.WriteString(fmt.Sprintf("async function __ssr_%s(request) {\n", comp.Name))

	// Run load function if present
	out.WriteString("  var __data = {};\n")
	if comp.Loader != nil {
		out.WriteString(fmt.Sprintf("  __data = await (async function(%s) {\n", comp.Loader.Param))
		g := New()
		g.indent = 2
		for _, stmt := range comp.Loader.Body {
			out.WriteString(g.genStmt(stmt))
			out.WriteString("\n")
		}
		out.WriteString("  })(request);\n")
		out.WriteString("  if (!__data) __data = {};\n")
	}

	// Generate initial state
	out.WriteString("  var __state = {};\n")
	for _, st := range comp.States {
		g := New()
		out.WriteString(fmt.Sprintf("  __state.%s = %s;\n", st.Name, g.genExpr(st.Value)))
	}

	// Render to HTML string
	out.WriteString("  var __html = '';\n")
	for _, el := range comp.RenderBody {
		out.WriteString(renderElementToString(&el, "  ", cssHash))
	}

	// Generate scoped CSS
	scopedCSS := ""
	if comp.Styles != nil {
		scopedCSS = GenerateScopedCSS(comp.Styles, cssHash)
	}

	// Generate head tags
	headHTML := ""
	if comp.Head != nil {
		headHTML = GenerateHeadHTML(comp.Head)
	}

	// Build full HTML page
	out.WriteString("  var __page = '<!DOCTYPE html><html><head><meta charset=\"utf-8\">';\n")
	if headHTML != "" {
		out.WriteString(fmt.Sprintf("  __page += %q;\n", headHTML))
	}
	if scopedCSS != "" {
		out.WriteString(fmt.Sprintf("  __page += '<style>%s</style>';\n", escapeJSString(scopedCSS)))
	}
	out.WriteString("  __page += '</head><body>';\n")
	out.WriteString(fmt.Sprintf("  __page += '<div id=\"app\" data-q-%s>';\n", cssHash))
	out.WriteString("  __page += __html;\n")
	out.WriteString("  __page += '</div>';\n")
	out.WriteString("  __page += '<script>window.__QUILL_DATA__ = ' + JSON.stringify(__data) + ';</script>';\n")
	out.WriteString("  __page += '</body></html>';\n")
	out.WriteString("  return __page;\n")
	out.WriteString("}\n")

	return out.String()
}

// GenerateHydration generates client-side hydration code.
func GenerateHydration(comp *ast.ComponentStatement, cssHash string) string {
	var out strings.Builder

	out.WriteString("// Client-side hydration\n")
	out.WriteString("(function() {\n")
	out.WriteString("  var __data = window.__QUILL_DATA__ || {};\n")
	out.WriteString(fmt.Sprintf("  var __app = document.getElementById('app');\n"))
	out.WriteString("  if (__app) {\n")
	out.WriteString(fmt.Sprintf("    var __comp = new QuillComponent(%s);\n", comp.Name))
	out.WriteString("    // Merge server data into state\n")
	out.WriteString("    for (var k in __data) { __comp.state[k] = __data[k]; }\n")
	out.WriteString("    __comp.__rootEl = __app;\n")
	out.WriteString("    __comp.__mounted = true;\n")
	out.WriteString("    // Attach event listeners (hydration)\n")
	out.WriteString("    __comp.__update();\n")
	out.WriteString("  }\n")
	out.WriteString("})();\n")

	return out.String()
}

// renderElementToString generates code that builds an HTML string from a render element.
func renderElementToString(el *ast.RenderElement, prefix string, cssHash string) string {
	var out strings.Builder

	if el.Condition != nil {
		out.WriteString(fmt.Sprintf("%sif (__data.%s) {\n", prefix, exprToSSRString(el.Condition)))
		for _, child := range el.Children {
			if child.Element != nil {
				out.WriteString(renderElementToString(child.Element, prefix+"  ", cssHash))
			} else if child.Text != nil {
				out.WriteString(fmt.Sprintf("%s  __html += %s;\n", prefix, ssrTextExpr(child.Text)))
			}
		}
		out.WriteString(fmt.Sprintf("%s}\n", prefix))
		return out.String()
	}

	if el.Iterator != nil {
		out.WriteString(fmt.Sprintf("%sif (__data.%s) {\n", prefix, exprToSSRString(el.Iterator.Iterable)))
		out.WriteString(fmt.Sprintf("%s  for (var __i = 0; __i < __data.%s.length; __i++) {\n",
			prefix, exprToSSRString(el.Iterator.Iterable)))
		out.WriteString(fmt.Sprintf("%s    var %s = __data.%s[__i];\n",
			prefix, el.Iterator.Variable, exprToSSRString(el.Iterator.Iterable)))
		for _, child := range el.Children {
			if child.Element != nil {
				out.WriteString(renderElementToString(child.Element, prefix+"    ", cssHash))
			} else if child.Text != nil {
				out.WriteString(fmt.Sprintf("%s    __html += %s;\n", prefix, ssrTextExpr(child.Text)))
			}
		}
		out.WriteString(fmt.Sprintf("%s  }\n", prefix))
		out.WriteString(fmt.Sprintf("%s}\n", prefix))
		return out.String()
	}

	if el.Tag == "__link" {
		// Render link as <a> tag with data-quill-link attribute
		to := ""
		if toExpr, ok := el.Props["to"]; ok {
			if sl, ok := toExpr.(*ast.StringLiteral); ok {
				to = sl.Value
			}
		}
		out.WriteString(fmt.Sprintf("%s__html += '<a href=\"%s\" data-quill-link>';\n", prefix, to))
		for _, child := range el.Children {
			if child.Text != nil {
				out.WriteString(fmt.Sprintf("%s__html += %s;\n", prefix, ssrTextExpr(child.Text)))
			}
		}
		out.WriteString(fmt.Sprintf("%s__html += '</a>';\n", prefix))
		return out.String()
	}

	// Regular element
	tag := el.Tag
	if tag == "__fragment" {
		tag = "div"
	}

	// Build attributes
	attrs := ""
	if cssHash != "" {
		attrs += fmt.Sprintf(" data-q-%s", cssHash)
	}
	for key, val := range el.Props {
		if strings.HasPrefix(key, "on") || strings.HasPrefix(key, "bind:") {
			continue // Skip event handlers and bindings in SSR
		}
		if sl, ok := val.(*ast.StringLiteral); ok {
			attrs += fmt.Sprintf(` %s="%s"`, key, sl.Value)
		}
	}

	out.WriteString(fmt.Sprintf("%s__html += '<%s%s>';\n", prefix, tag, attrs))

	for _, child := range el.Children {
		if child.Element != nil {
			out.WriteString(renderElementToString(child.Element, prefix, cssHash))
		} else if child.Text != nil {
			out.WriteString(fmt.Sprintf("%s__html += %s;\n", prefix, ssrTextExpr(child.Text)))
		}
	}

	// Self-closing tags
	selfClosing := map[string]bool{"input": true, "br": true, "hr": true, "img": true, "meta": true, "link": true}
	if !selfClosing[tag] {
		out.WriteString(fmt.Sprintf("%s__html += '</%s>';\n", prefix, tag))
	}

	return out.String()
}

func exprToSSRString(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name
	case *ast.DotExpr:
		return exprToSSRString(e.Object) + "." + e.Field
	default:
		return "undefined"
	}
}

func ssrTextExpr(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		if hasInterpolation(e.Value) {
			// For SSR, interpolation references come from __data or local vars
			converted := e.Value
			converted = strings.ReplaceAll(converted, "`", "\\`")
			converted = convertSSRInterpolation(converted)
			return "`" + converted + "`"
		}
		return fmt.Sprintf("%q", e.Value)
	case *ast.Identifier:
		return e.Name
	default:
		return "''"
	}
}

func convertSSRInterpolation(s string) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '{' {
			if i > 0 && s[i-1] == '$' {
				out.WriteByte(s[i])
				i++
				continue
			}
			j := i + 1
			if j < len(s) && isInterpolationStart(s[j]) {
				for j < len(s) && s[j] != '}' {
					j++
				}
				if j < len(s) {
					content := s[i+1 : j]
					if isInterpolationExpr(content) {
						out.WriteString("${")
						out.WriteString(content)
						out.WriteByte('}')
						i = j + 1
						continue
					}
				}
			}
			out.WriteString("\\{")
			i++
		} else {
			out.WriteByte(s[i])
			i++
		}
	}
	return out.String()
}

// GenerateScopedCSS generates CSS with scoped selectors.
func GenerateScopedCSS(styles *ast.StyleBlock, hash string) string {
	var out strings.Builder
	for _, rule := range styles.Rules {
		// Add scope attribute selector to each selector
		selector := scopeSelector(rule.Selector, hash)
		out.WriteString(selector + " {\n")
		for prop, val := range rule.Properties {
			out.WriteString(fmt.Sprintf("  %s: %s;\n", prop, val))
		}
		out.WriteString("}\n")
	}
	return out.String()
}

// scopeSelector adds [data-q-HASH] to a CSS selector for scoping.
func scopeSelector(sel string, hash string) string {
	attr := fmt.Sprintf("[data-q-%s]", hash)
	// Handle compound selectors like ".card:hover"
	parts := strings.Split(sel, ",")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		// Insert scope attribute after the first simple selector
		spaceIdx := strings.Index(part, " ")
		if spaceIdx == -1 {
			parts[i] = part + attr
		} else {
			parts[i] = part[:spaceIdx] + attr + part[spaceIdx:]
		}
	}
	return strings.Join(parts, ", ")
}

// GenerateHeadHTML generates HTML string for head entries.
func GenerateHeadHTML(head *ast.HeadBlock) string {
	var out strings.Builder
	for _, entry := range head.Entries {
		switch entry.Tag {
		case "title":
			out.WriteString(fmt.Sprintf("<title>%s</title>", entry.Text))
		case "meta":
			out.WriteString("<meta")
			for k, v := range entry.Attrs {
				out.WriteString(fmt.Sprintf(` %s="%s"`, k, v))
			}
			out.WriteString(">")
		case "link":
			out.WriteString("<link")
			for k, v := range entry.Attrs {
				out.WriteString(fmt.Sprintf(` %s="%s"`, k, v))
			}
			out.WriteString(">")
		}
	}
	return out.String()
}

// ComponentHash generates a unique hash for a component (used for CSS scoping).
func ComponentHash(name string) string {
	h := sha256.Sum256([]byte(name))
	return fmt.Sprintf("%x", h[:3]) // 6-char hex hash
}

// GenerateHeadJS generates client-side JS to manage document.head.
func GenerateHeadJS(head *ast.HeadBlock) string {
	var out strings.Builder
	out.WriteString("// Quill Head Management\n")
	out.WriteString("(function() {\n")
	for _, entry := range head.Entries {
		switch entry.Tag {
		case "title":
			out.WriteString(fmt.Sprintf("  document.title = %q;\n", entry.Text))
		case "meta":
			// Find or create meta tag
			var selector string
			if name, ok := entry.Attrs["name"]; ok {
				selector = fmt.Sprintf("meta[name=\"%s\"]", name)
			} else if prop, ok := entry.Attrs["property"]; ok {
				selector = fmt.Sprintf("meta[property=\"%s\"]", prop)
			}
			if selector != "" {
				out.WriteString(fmt.Sprintf("  (function() {\n"))
				out.WriteString(fmt.Sprintf("    var el = document.querySelector('%s');\n", selector))
				out.WriteString("    if (!el) { el = document.createElement('meta'); document.head.appendChild(el); }\n")
				for k, v := range entry.Attrs {
					out.WriteString(fmt.Sprintf("    el.setAttribute(%q, %q);\n", k, v))
				}
				out.WriteString("  })();\n")
			}
		case "link":
			out.WriteString("  (function() {\n")
			out.WriteString("    var el = document.createElement('link');\n")
			for k, v := range entry.Attrs {
				out.WriteString(fmt.Sprintf("    el.setAttribute(%q, %q);\n", k, v))
			}
			out.WriteString("    document.head.appendChild(el);\n")
			out.WriteString("  })();\n")
		}
	}
	out.WriteString("})();\n")
	return out.String()
}

func escapeJSString(s string) string {
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

// RenderToString is a runtime helper name - the actual implementation is in the framework runtime.
const renderToStringHelper = `
function __quill_renderToString(vnode) {
  if (typeof vnode === 'string' || typeof vnode === 'number') return String(vnode);
  if (!vnode || !vnode.tag) return '';
  var selfClosing = {input:1, br:1, hr:1, img:1, meta:1, link:1};
  var html = '<' + vnode.tag;
  for (var key in vnode.props) {
    if (key.slice(0,2) === 'on' || key.indexOf('bind:') === 0) continue;
    html += ' ' + key + '="' + vnode.props[key] + '"';
  }
  if (selfClosing[vnode.tag]) return html + '>';
  html += '>';
  for (var i = 0; i < vnode.children.length; i++) {
    html += __quill_renderToString(vnode.children[i]);
  }
  html += '</' + vnode.tag + '>';
  return html;
}
`
