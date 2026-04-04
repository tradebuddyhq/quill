package codegen

import (
	"fmt"
	"quill/ast"
	"strings"
)

// genComponent generates JavaScript for a ComponentStatement.
// It creates a component definition object with initialState, methods, and render function,
// then uses the QuillComponent runtime to make it reactive.
func (g *Generator) genComponent(s *ast.ComponentStatement, prefix string) string {
	var out strings.Builder

	out.WriteString(fmt.Sprintf("%sconst %s = {\n", prefix, s.Name))
	g.indent++
	inner := strings.Repeat("  ", g.indent)

	// --- initialState ---
	out.WriteString(fmt.Sprintf("%sinitialState: function() {\n", inner))
	g.indent++
	stateInner := strings.Repeat("  ", g.indent)
	out.WriteString(fmt.Sprintf("%sreturn {\n", stateInner))
	g.indent++
	propInner := strings.Repeat("  ", g.indent)
	for _, st := range s.States {
		out.WriteString(fmt.Sprintf("%s%s: %s,\n", propInner, st.Name, g.genExpr(st.Value)))
	}
	g.indent--
	out.WriteString(fmt.Sprintf("%s};\n", stateInner))
	g.indent--
	out.WriteString(fmt.Sprintf("%s},\n", inner))

	// --- methods ---
	out.WriteString(fmt.Sprintf("%smethods: function(__comp) {\n", inner))
	g.indent++
	methOuter := strings.Repeat("  ", g.indent)
	out.WriteString(fmt.Sprintf("%sreturn {\n", methOuter))
	g.indent++
	methInner := strings.Repeat("  ", g.indent)
	for _, method := range s.Methods {
		params := strings.Join(method.Params, ", ")
		out.WriteString(fmt.Sprintf("%s%s: function(%s) {\n", methInner, method.Name, params))
		g.indent++
		for _, stmt := range method.Body {
			out.WriteString(g.genComponentStmt(stmt))
			out.WriteString("\n")
		}
		g.indent--
		out.WriteString(fmt.Sprintf("%s},\n", methInner))
	}
	g.indent--
	out.WriteString(fmt.Sprintf("%s};\n", methOuter))
	g.indent--
	out.WriteString(fmt.Sprintf("%s},\n", inner))

	// --- form actions ---
	if len(s.Actions) > 0 {
		out.WriteString(fmt.Sprintf("%sactions: {\n", inner))
		g.indent++
		actInner := strings.Repeat("  ", g.indent)
		for _, action := range s.Actions {
			out.WriteString(fmt.Sprintf("%s%s: function(__comp, __data) {\n", actInner, action.Name))
			g.indent++
			for _, stmt := range action.Body {
				out.WriteString(g.genComponentStmt(stmt))
				out.WriteString("\n")
			}
			g.indent--
			out.WriteString(fmt.Sprintf("%s},\n", actInner))
		}
		g.indent--
		out.WriteString(fmt.Sprintf("%s},\n", inner))
	}

	// --- render ---
	out.WriteString(fmt.Sprintf("%srender: function(__comp) {\n", inner))
	g.indent++
	renderInner := strings.Repeat("  ", g.indent)
	if len(s.RenderBody) == 1 {
		out.WriteString(fmt.Sprintf("%sreturn %s;\n", renderInner, g.genRenderElement(&s.RenderBody[0])))
	} else if len(s.RenderBody) > 1 {
		var children []string
		for i := range s.RenderBody {
			children = append(children, g.genRenderElement(&s.RenderBody[i]))
		}
		out.WriteString(fmt.Sprintf("%sreturn h(\"div\", {}, %s);\n", renderInner, strings.Join(children, ", ")))
	} else {
		out.WriteString(fmt.Sprintf("%sreturn h(\"div\", {});\n", renderInner))
	}
	g.indent--
	out.WriteString(fmt.Sprintf("%s}\n", inner))

	g.indent--
	out.WriteString(fmt.Sprintf("%s};\n", prefix))

	// --- scoped styles ---
	if s.Styles != nil {
		cssHash := ComponentHash(s.Name)
		scopedCSS := GenerateScopedCSS(s.Styles, cssHash)
		out.WriteString(fmt.Sprintf("%s// Scoped styles for %s\n", prefix, s.Name))
		out.WriteString(fmt.Sprintf("%s(function() {\n", prefix))
		out.WriteString(fmt.Sprintf("%s  var style = document.createElement('style');\n", prefix))
		out.WriteString(fmt.Sprintf("%s  style.textContent = %q;\n", prefix, scopedCSS))
		out.WriteString(fmt.Sprintf("%s  document.head.appendChild(style);\n", prefix))
		out.WriteString(fmt.Sprintf("%s})();\n", prefix))
	}

	// --- head management ---
	if s.Head != nil {
		out.WriteString(GenerateHeadJS(s.Head))
	}

	return out.String()
}

// genRenderElement converts a RenderElement AST node into an h() virtual DOM call.
func (g *Generator) genRenderElement(el *ast.RenderElement) string {
	// Handle conditional rendering
	if el.Condition != nil {
		var childCalls []string
		for _, child := range el.Children {
			if child.Element != nil {
				childCalls = append(childCalls, g.genRenderElement(child.Element))
			} else if child.Text != nil {
				childCalls = append(childCalls, g.genComponentExpr(child.Text))
			}
		}
		inner := strings.Join(childCalls, ", ")
		if len(childCalls) == 1 {
			return fmt.Sprintf("(%s ? %s : null)", g.genComponentExpr(el.Condition), inner)
		}
		return fmt.Sprintf("(%s ? h(\"div\", {}, %s) : null)", g.genComponentExpr(el.Condition), inner)
	}

	// Handle list rendering
	if el.Iterator != nil {
		var childCalls []string
		for _, child := range el.Children {
			if child.Element != nil {
				childCalls = append(childCalls, g.genRenderElement(child.Element))
			} else if child.Text != nil {
				childCalls = append(childCalls, g.genComponentExpr(child.Text))
			}
		}
		inner := childCalls[0]
		if len(childCalls) > 1 {
			inner = fmt.Sprintf("h(\"div\", {}, %s)", strings.Join(childCalls, ", "))
		}
		return fmt.Sprintf("(%s).map(function(%s, __idx) { return %s; })",
			g.genComponentExpr(el.Iterator.Iterable),
			el.Iterator.Variable,
			inner)
	}

	// Handle link element: generates <a> with client-side navigation
	if el.Tag == "__link" {
		to := `"/"`
		if toExpr, ok := el.Props["to"]; ok {
			to = g.genComponentExpr(toExpr)
		}
		var childCalls []string
		for _, child := range el.Children {
			if child.Text != nil {
				childCalls = append(childCalls, g.genComponentExpr(child.Text))
			} else if child.Element != nil {
				childCalls = append(childCalls, g.genRenderElement(child.Element))
			}
		}
		propsStr := fmt.Sprintf(`{href: %s, "data-quill-link": "true", onclick: function(e) { e.preventDefault(); if (typeof __quill_navigate === 'function') __quill_navigate(%s); }}`, to, to)
		if len(childCalls) > 0 {
			return fmt.Sprintf("h(\"a\", %s, %s)", propsStr, strings.Join(childCalls, ", "))
		}
		return fmt.Sprintf("h(\"a\", %s)", propsStr)
	}

	// Build props object
	var propParts []string
	for key, val := range el.Props {
		if strings.HasPrefix(key, "on") {
			// Event handler: onClick handler -> {onclick: function() { __comp.handler(); }}
			handlerName := ""
			if ident, ok := val.(*ast.Identifier); ok {
				handlerName = ident.Name
			}
			jsEvtName := strings.ToLower(key)
			propParts = append(propParts, fmt.Sprintf("%s: function(e) { __comp.%s(e); }", jsEvtName, handlerName))
		} else if strings.HasPrefix(key, "bind:") {
			// Two-way binding: bind:value stateVar
			boundField := strings.TrimPrefix(key, "bind:")
			if ident, ok := val.(*ast.Identifier); ok {
				propParts = append(propParts, fmt.Sprintf("value: __comp.state.%s", ident.Name))
				propParts = append(propParts, fmt.Sprintf("oninput: function(e) { __comp.state.%s = e.target.value; }", ident.Name))
			} else {
				propParts = append(propParts, fmt.Sprintf("%s: %s", boundField, g.genComponentExpr(val)))
			}
		} else {
			propParts = append(propParts, fmt.Sprintf("%s: %s", key, g.genComponentExpr(val)))
		}
	}

	propsStr := "{}"
	if len(propParts) > 0 {
		propsStr = "{" + strings.Join(propParts, ", ") + "}"
	}

	// Build children
	var childCalls []string
	for _, child := range el.Children {
		if child.Element != nil {
			childCalls = append(childCalls, g.genRenderElement(child.Element))
		} else if child.Text != nil {
			childCalls = append(childCalls, g.genComponentExpr(child.Text))
		}
	}

	if len(childCalls) > 0 {
		return fmt.Sprintf("h(\"%s\", %s, %s)", el.Tag, propsStr, strings.Join(childCalls, ", "))
	}
	return fmt.Sprintf("h(\"%s\", %s)", el.Tag, propsStr)
}

// genComponentExpr generates a JS expression for use inside a component,
// replacing bare identifiers that are state variables with __comp.state.X references.
func (g *Generator) genComponentExpr(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		if hasInterpolation(e.Value) {
			// Convert {var} to ${__comp.state.var}
			converted := e.Value
			converted = strings.ReplaceAll(converted, "`", "\\`")
			converted = convertComponentInterpolation(converted)
			return "`" + converted + "`"
		}
		escaped := strings.ReplaceAll(e.Value, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		return `"` + escaped + `"`

	case *ast.Identifier:
		return "__comp.state." + e.Name

	case *ast.DotExpr:
		return g.genComponentExpr(e.Object) + "." + e.Field

	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", g.genComponentExpr(e.Object), g.genComponentExpr(e.Index))

	case *ast.BinaryExpr:
		return fmt.Sprintf("(%s %s %s)", g.genComponentExpr(e.Left), e.Operator, g.genComponentExpr(e.Right))

	case *ast.ComparisonExpr:
		op := e.Operator
		if op == "==" {
			op = "==="
		}
		if op == "!=" {
			op = "!=="
		}
		return fmt.Sprintf("(%s %s %s)", g.genComponentExpr(e.Left), op, g.genComponentExpr(e.Right))

	case *ast.NotExpr:
		return fmt.Sprintf("(!%s)", g.genComponentExpr(e.Operand))

	case *ast.LogicalExpr:
		op := "&&"
		if e.Operator == "or" {
			op = "||"
		}
		return fmt.Sprintf("(%s %s %s)", g.genComponentExpr(e.Left), op, g.genComponentExpr(e.Right))

	case *ast.CallExpr:
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = g.genComponentExpr(a)
		}
		// For function calls, use the regular genExpr for the function name
		// since push/pop etc are globals, not state
		fnName := g.genExpr(e.Function)
		return fmt.Sprintf("%s(%s)", fnName, strings.Join(args, ", "))

	case *ast.NumberLiteral:
		return g.genExpr(expr)

	case *ast.BoolLiteral:
		return g.genExpr(expr)

	case *ast.ListLiteral:
		elems := make([]string, len(e.Elements))
		for i, el := range e.Elements {
			elems[i] = g.genComponentExpr(el)
		}
		return "[" + strings.Join(elems, ", ") + "]"

	case *ast.ObjectLiteral:
		if len(e.Keys) == 0 {
			return "{}"
		}
		pairs := make([]string, len(e.Keys))
		for i, key := range e.Keys {
			pairs[i] = fmt.Sprintf("%s: %s", key, g.genComponentExpr(e.Values[i]))
		}
		return "{ " + strings.Join(pairs, ", ") + " }"

	case *ast.NothingLiteral:
		return "null"

	default:
		return g.genExpr(expr)
	}
}

// genComponentStmt generates a statement inside a component method,
// where state variables are accessed through __comp.state.
func (g *Generator) genComponentStmt(stmt ast.Statement) string {
	prefix := strings.Repeat("  ", g.indent)

	switch s := stmt.(type) {
	case *ast.AssignStatement:
		return fmt.Sprintf("%s__comp.state.%s = %s;", prefix, s.Name, g.genComponentExpr(s.Value))

	case *ast.IfStatement:
		return g.genComponentIf(s, prefix)

	case *ast.ForEachStatement:
		g.indent++
		var body strings.Builder
		for _, st := range s.Body {
			body.WriteString(g.genComponentStmt(st))
			body.WriteString("\n")
		}
		g.indent--
		return fmt.Sprintf("%sfor (const %s of %s) {\n%s%s}",
			prefix, s.Variable, g.genComponentExpr(s.Iterable), body.String(), prefix)

	case *ast.ExprStatement:
		return fmt.Sprintf("%s%s;", prefix, g.genComponentExpr(s.Expr))

	case *ast.SayStatement:
		return fmt.Sprintf("%sconsole.log(%s);", prefix, g.genComponentExpr(s.Value))

	case *ast.ReturnStatement:
		return fmt.Sprintf("%sreturn %s;", prefix, g.genComponentExpr(s.Value))

	default:
		return g.genStmt(stmt)
	}
}

func (g *Generator) genComponentIf(s *ast.IfStatement, prefix string) string {
	var out strings.Builder

	g.indent++
	var body strings.Builder
	for _, st := range s.Body {
		body.WriteString(g.genComponentStmt(st))
		body.WriteString("\n")
	}
	g.indent--

	out.WriteString(fmt.Sprintf("%sif (%s) {\n%s%s}", prefix, g.genComponentExpr(s.Condition), body.String(), prefix))

	for _, elif := range s.ElseIfs {
		g.indent++
		var elifBody strings.Builder
		for _, st := range elif.Body {
			elifBody.WriteString(g.genComponentStmt(st))
			elifBody.WriteString("\n")
		}
		g.indent--
		out.WriteString(fmt.Sprintf(" else if (%s) {\n%s%s}", g.genComponentExpr(elif.Condition), elifBody.String(), prefix))
	}

	if len(s.Else) > 0 {
		g.indent++
		var elseBody strings.Builder
		for _, st := range s.Else {
			elseBody.WriteString(g.genComponentStmt(st))
			elseBody.WriteString("\n")
		}
		g.indent--
		out.WriteString(fmt.Sprintf(" else {\n%s%s}", elseBody.String(), prefix))
	}

	return out.String()
}

// genMount generates JavaScript for a MountStatement.
func (g *Generator) genMount(s *ast.MountStatement, prefix string) string {
	return fmt.Sprintf("%s__quill_mount(%s, %s);", prefix, s.Component, g.genExpr(s.Selector))
}

// convertComponentInterpolation converts {var} to ${__comp.state.var} for template literals
// inside component render functions.
func convertComponentInterpolation(s string) string {
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
						// Rewrite to __comp.state.X
						out.WriteString("${__comp.state.")
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
