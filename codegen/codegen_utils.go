package codegen

import (
	"fmt"
	"quill/ast"
	"strings"
)

func escapeJS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, "\x00", "\\0")
	return s
}

// replaceMyInInterpolations replaces "my." with "this." only inside {interpolation} braces.
func replaceMyInInterpolations(s string) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '{' {
			// Find the matching closing brace
			j := i + 1
			for j < len(s) && s[j] != '}' {
				j++
			}
			if j < len(s) {
				content := s[i+1 : j]
				// Replace my. with this. only inside the interpolation
				content = strings.ReplaceAll(content, "my.", "this.")
				out.WriteByte('{')
				out.WriteString(content)
				out.WriteByte('}')
				i = j + 1
				continue
			}
		}
		out.WriteByte(s[i])
		i++
	}
	return out.String()
}

// convertInterpolation converts {expr} to ${expr} for JS template literals.
// Only converts {identifier} or {expr.field} patterns, not { css } blocks.
func convertInterpolation(s string) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '{' {
			// Check if it's already ${
			if i > 0 && s[i-1] == '$' {
				out.WriteByte(s[i])
				i++
				continue
			}
			// Look ahead to see if content looks like an interpolation expression
			// (starts with a letter/underscore, contains only valid identifier chars, dots, parens, etc.)
			j := i + 1
			if j < len(s) && isInterpolationStart(s[j]) {
				// Find the closing brace
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
			// Not an interpolation — keep the brace as-is but escape for template literal
			out.WriteString("\\{")
			i++
		} else {
			out.WriteByte(s[i])
			i++
		}
	}
	return out.String()
}

// hasInterpolation returns true if the string contains at least one {identifier} interpolation.
func hasInterpolation(s string) bool {
	i := 0
	for i < len(s) {
		if s[i] == '{' {
			j := i + 1
			if j < len(s) && isInterpolationStart(s[j]) {
				for j < len(s) && s[j] != '}' {
					j++
				}
				if j < len(s) {
					content := s[i+1 : j]
					if isInterpolationExpr(content) {
						return true
					}
				}
			}
		}
		i++
	}
	return false
}

// isInterpolationStart returns true if the character can start a Quill interpolation.
func isInterpolationStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// isInterpolationExpr returns true if the string looks like a valid interpolation expression
// (identifier, dot access, function call, etc.) rather than CSS or other content.
func isInterpolationExpr(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Must start with a letter or underscore
	if !isInterpolationStart(s[0]) {
		return false
	}
	// Should only contain valid identifier characters, dots, parens, commas, spaces
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '_' || c == '.' || c == '(' || c == ')' || c == ',' || c == ' ') {
			return false
		}
	}
	return true
}

// usesIdentifier checks if the generated user code references a given identifier
// followed by a dot (e.g., "Auth.", "DB.", "Validate.", "Log.").
func (g *Generator) usesIdentifier(code string, name string) bool {
	return strings.Contains(code, name+".")
}

// hasComponents checks if the program contains any ComponentStatement nodes.
func (g *Generator) hasComponents(program *ast.Program) bool {
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.ComponentStatement); ok {
			return true
		}
		if _, ok := stmt.(*ast.MountStatement); ok {
			return true
		}
	}
	return false
}

// bodyContainsYield checks if a function body contains any yield statements (including nested in loops).
func (g *Generator) bodyContainsYield(stmts []ast.Statement) bool {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.YieldStatement:
			return true
		case *ast.LoopStatement:
			if g.bodyContainsYield(s.Body) {
				return true
			}
		case *ast.IfStatement:
			if g.bodyContainsYield(s.Body) {
				return true
			}
			for _, elif := range s.ElseIfs {
				if g.bodyContainsYield(elif.Body) {
					return true
				}
			}
			if g.bodyContainsYield(s.Else) {
				return true
			}
		case *ast.WhileStatement:
			if g.bodyContainsYield(s.Body) {
				return true
			}
		case *ast.ForEachStatement:
			if g.bodyContainsYield(s.Body) {
				return true
			}
		case *ast.TryCatchStatement:
			if g.bodyContainsYield(s.TryBody) {
				return true
			}
			if g.bodyContainsYield(s.CatchBody) {
				return true
			}
		}
	}
	return false
}

// bodyContainsAwait checks if a function body contains any await expressions (recursively).
func (g *Generator) bodyContainsAwait(stmts []ast.Statement) bool {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.AssignStatement:
			if g.exprContainsAwait(s.Value) {
				return true
			}
		case *ast.SayStatement:
			if g.exprContainsAwait(s.Value) {
				return true
			}
		case *ast.ReturnStatement:
			if g.exprContainsAwait(s.Value) {
				return true
			}
		case *ast.ExprStatement:
			if g.exprContainsAwait(s.Expr) {
				return true
			}
		case *ast.IfStatement:
			if g.exprContainsAwait(s.Condition) {
				return true
			}
			if g.bodyContainsAwait(s.Body) {
				return true
			}
			for _, elif := range s.ElseIfs {
				if g.bodyContainsAwait(elif.Body) {
					return true
				}
			}
			if g.bodyContainsAwait(s.Else) {
				return true
			}
		case *ast.WhileStatement:
			if g.bodyContainsAwait(s.Body) {
				return true
			}
		case *ast.ForEachStatement:
			if g.bodyContainsAwait(s.Body) {
				return true
			}
		case *ast.LoopStatement:
			if g.bodyContainsAwait(s.Body) {
				return true
			}
		case *ast.TryCatchStatement:
			if g.bodyContainsAwait(s.TryBody) {
				return true
			}
			if g.bodyContainsAwait(s.CatchBody) {
				return true
			}
		case *ast.DotAssignStatement:
			if g.exprContainsAwait(s.Value) {
				return true
			}
		}
	}
	return false
}

// exprContainsAwait checks if an expression tree contains any await expression.
func (g *Generator) exprContainsAwait(expr ast.Expression) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.AwaitExpr:
		return true
	case *ast.AwaitExpression:
		return true
	case *ast.CallExpr:
		if g.exprContainsAwait(e.Function) {
			return true
		}
		for _, arg := range e.Args {
			if g.exprContainsAwait(arg) {
				return true
			}
		}
	case *ast.DotExpr:
		return g.exprContainsAwait(e.Object)
	case *ast.IndexExpr:
		return g.exprContainsAwait(e.Object) || g.exprContainsAwait(e.Index)
	case *ast.BinaryExpr:
		return g.exprContainsAwait(e.Left) || g.exprContainsAwait(e.Right)
	case *ast.UnaryMinusExpr:
		return g.exprContainsAwait(e.Operand)
	case *ast.PipeExpr:
		return g.exprContainsAwait(e.Left) || g.exprContainsAwait(e.Right)
	}
	return false
}

// genTypeCondition generates a JavaScript type check condition for pattern matching.
func (g *Generator) genTypeCondition(typeName string, matchVar string) string {
	switch typeName {
	case "text":
		return fmt.Sprintf("(typeof %s === \"string\")", matchVar)
	case "number":
		return fmt.Sprintf("(typeof %s === \"number\")", matchVar)
	case "boolean":
		return fmt.Sprintf("(typeof %s === \"boolean\")", matchVar)
	case "list":
		return fmt.Sprintf("(Array.isArray(%s))", matchVar)
	case "nothing":
		return fmt.Sprintf("(%s == null)", matchVar)
	default:
		return fmt.Sprintf("(typeof %s === \"%s\")", matchVar, typeName)
	}
}

// hasGenerators checks if the program uses generator/iterator features.
func (g *Generator) hasGenerators(program *ast.Program) bool {
	for _, stmt := range program.Statements {
		if g.stmtUsesGenerators(stmt) {
			return true
		}
	}
	return false
}

// needsIteratorRuntime checks if generated code uses lazy iterator features.
func (g *Generator) needsIteratorRuntime(code string) bool {
	keywords := []string{"__quill_lazy", "__quill_range", "__QuillLazy", "collect()", ".filter(", ".take(", ".skip("}
	for _, kw := range keywords {
		if strings.Contains(code, kw) {
			return true
		}
	}
	return false
}

func (g *Generator) stmtUsesGenerators(stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.YieldStatement:
		return true
	case *ast.LoopStatement:
		// Only return true if the loop body actually contains yield
		if g.bodyContainsYield(s.Body) {
			return true
		}
	case *ast.FuncDefinition:
		if g.bodyContainsYield(s.Body) {
			return true
		}
	}
	return false
}

// --- Type Alias Code Generation ---

func (g *Generator) genTypeAlias(s *ast.TypeAliasStatement, prefix string) string {
	args := ""
	if len(s.Args) > 0 {
		args = ", [" + strings.Join(s.Args, ", ") + "]"
	}
	return fmt.Sprintf("%s// type %s = %s<%s%s> (erased at runtime)", prefix, s.Name, s.Utility, s.BaseType, args)
}

// --- Decorator Code Generation ---

const decoratorRuntime = `function authenticated(req) {
  return req.headers && req.headers.authorization;
}
function rateLimit(max) {
  const counts = new Map();
  return function(req) {
    const ip = req.socket && req.socket.remoteAddress || 'unknown';
    const count = (counts.get(ip) || 0) + 1;
    counts.set(ip, count);
    return count <= max;
  };
}
function log(fn) {
  return function(...args) {
    console.log('[' + fn.name + '] called with', args);
    const result = fn.apply(this, args);
    console.log('[' + fn.name + '] returned', result);
    return result;
  };
}
`

func (g *Generator) genDecoratedFunc(s *ast.DecoratedFuncDefinition, prefix string) string {
	var out strings.Builder
	// Generate the function itself as a let-bound function expression
	params := strings.Join(s.Func.Params, ", ")
	g.indent++
	body := g.genBlock(s.Func.Body)
	g.indent--
	out.WriteString(fmt.Sprintf("%slet %s = function(%s) {\n%s%s};\n", prefix, s.Func.Name, params, body, prefix))
	g.declared[s.Func.Name] = true
	// Apply decorators in reverse order (innermost first)
	for i := len(s.Decorators) - 1; i >= 0; i-- {
		dec := s.Decorators[i]
		if len(dec.Args) > 0 {
			args := make([]string, len(dec.Args))
			for j, a := range dec.Args {
				args[j] = g.genExpr(a)
			}
			out.WriteString(fmt.Sprintf("%s%s = %s(%s)(%s);\n", prefix, s.Func.Name, dec.Name, strings.Join(args, ", "), s.Func.Name))
		} else {
			out.WriteString(fmt.Sprintf("%s%s = %s(%s);\n", prefix, s.Func.Name, dec.Name, s.Func.Name))
		}
	}
	return out.String()
}

func (g *Generator) genDecoratedRoute(s *ast.DecoratedRouteDefinition, prefix string) string {
	var out strings.Builder
	// Comment showing decorators
	decNames := make([]string, len(s.Decorators))
	for i, d := range s.Decorators {
		decNames[i] = "@" + d.Name
	}
	out.WriteString(fmt.Sprintf("%s// %s\n", prefix, strings.Join(decNames, " ")))
	out.WriteString(fmt.Sprintf("%sif (method === '%s' && url.pathname === '%s') {\n", prefix, s.Route.Method, s.Route.Path))
	g.indent++
	innerPrefix := strings.Repeat("  ", g.indent)
	// Generate middleware checks
	for _, dec := range s.Decorators {
		if len(dec.Args) > 0 {
			args := make([]string, len(dec.Args))
			for j, a := range dec.Args {
				args[j] = g.genExpr(a)
			}
			out.WriteString(fmt.Sprintf("%sif (!%s(%s)(req)) { res.writeHead(429); res.end('Too Many Requests'); return; }\n", innerPrefix, dec.Name, strings.Join(args, ", ")))
		} else {
			out.WriteString(fmt.Sprintf("%sif (!%s(req)) { res.writeHead(401); res.end('Unauthorized'); return; }\n", innerPrefix, dec.Name))
		}
	}
	// Handler body
	for _, stmt := range s.Route.Body {
		out.WriteString(g.genStmt(stmt))
		out.WriteString("\n")
	}
	g.indent--
	out.WriteString(fmt.Sprintf("%s}", prefix))
	return out.String()
}

// hasDecorators checks if the program uses any decorator features.
func (g *Generator) hasDecorators(program *ast.Program) bool {
	for _, stmt := range program.Statements {
		switch stmt.(type) {
		case *ast.DecoratedFuncDefinition, *ast.DecoratedRouteDefinition:
			return true
		}
	}
	return false
}

// --- WebSocket Code Generation ---

func (g *Generator) genWebSocket(s *ast.WebSocketBlock, prefix string) string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%sconst WebSocket = require('ws');\n", prefix))
	out.WriteString(fmt.Sprintf("%sconst __wss = new WebSocket.Server({ noServer: true });\n", prefix))
	out.WriteString(fmt.Sprintf("%sconst __ws_clients = new Set();\n", prefix))
	out.WriteString(fmt.Sprintf("%sserver.on('upgrade', (request, socket, head) => {\n", prefix))
	g.indent++
	innerPrefix := strings.Repeat("  ", g.indent)
	out.WriteString(fmt.Sprintf("%sif (request.url === '%s') {\n", innerPrefix, s.Path))
	g.indent++
	innerPrefix2 := strings.Repeat("  ", g.indent)
	out.WriteString(fmt.Sprintf("%s__wss.handleUpgrade(request, socket, head, (ws) => {\n", innerPrefix2))
	out.WriteString(fmt.Sprintf("%s  __wss.emit('connection', ws, request);\n", innerPrefix2))
	out.WriteString(fmt.Sprintf("%s});\n", innerPrefix2))
	g.indent--
	out.WriteString(fmt.Sprintf("%s}\n", innerPrefix))
	g.indent--
	out.WriteString(fmt.Sprintf("%s});\n", prefix))

	// Connection handler
	out.WriteString(fmt.Sprintf("%s__wss.on('connection', (%s) => {\n", prefix, s.ConnectVar))
	g.indent++
	innerPrefix = strings.Repeat("  ", g.indent)
	out.WriteString(fmt.Sprintf("%s__ws_clients.add(%s);\n", innerPrefix, s.ConnectVar))

	// OnConnect body
	if len(s.OnConnect) > 0 {
		for _, stmt := range s.OnConnect {
			out.WriteString(g.genStmt(stmt))
			out.WriteString("\n")
		}
	}

	// OnMessage handler
	if len(s.OnMessage) > 0 {
		out.WriteString(fmt.Sprintf("%s%s.on('message', (%s) => {\n", innerPrefix, s.ConnectVar, s.DataVar))
		g.indent++
		for _, stmt := range s.OnMessage {
			out.WriteString(g.genStmt(stmt))
			out.WriteString("\n")
		}
		g.indent--
		out.WriteString(fmt.Sprintf("%s});\n", innerPrefix))
	}

	// OnClose handler
	if len(s.OnClose) > 0 {
		out.WriteString(fmt.Sprintf("%s%s.on('close', () => {\n", innerPrefix, s.ConnectVar))
		g.indent++
		closePrefix := strings.Repeat("  ", g.indent)
		out.WriteString(fmt.Sprintf("%s__ws_clients.delete(%s);\n", closePrefix, s.ConnectVar))
		for _, stmt := range s.OnClose {
			out.WriteString(g.genStmt(stmt))
			out.WriteString("\n")
		}
		g.indent--
		out.WriteString(fmt.Sprintf("%s});\n", innerPrefix))
	}

	g.indent--
	out.WriteString(fmt.Sprintf("%s});\n", prefix))
	return out.String()
}

// hasWebSockets checks if the program uses WebSocket features.
func (g *Generator) hasWebSockets(program *ast.Program) bool {
	for _, stmt := range program.Statements {
		if srv, ok := stmt.(*ast.ServerBlockStatement); ok {
			if len(srv.WebSockets) > 0 {
				return true
			}
		}
		if _, ok := stmt.(*ast.WebSocketBlock); ok {
			return true
		}
	}
	return false
}
