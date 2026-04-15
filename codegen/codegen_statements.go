package codegen

import (
	"fmt"
	"os"
	"quill/ast"
	"quill/lexer"
	"quill/parser"
	"strings"
)

func (g *Generator) genStmt(stmt ast.Statement) string {
	prefix := strings.Repeat("  ", g.indent)

	switch s := stmt.(type) {
	case *ast.AssignStatement:
		g.addStmtMapping(s.Line)
		// Detect Discord.bot() or new Discord.Client() assignment to track the client variable name
		if call, ok := s.Value.(*ast.CallExpr); ok {
			if dot, ok := call.Function.(*ast.DotExpr); ok {
				if ident, ok := dot.Object.(*ast.Identifier); ok && ident.Name == "Discord" && dot.Field == "bot" {
					g.discordClientVar = s.Name
				}
			}
		}
		if newExpr, ok := s.Value.(*ast.NewExpr); ok {
			if newExpr.ClassName == "Discord.Client" {
				g.discordClientVar = s.Name
			}
		}
		value := g.genExpr(s.Value)
		if g.declared[s.Name] {
			return fmt.Sprintf("%s%s = %s;", prefix, s.Name, value)
		}
		g.declared[s.Name] = true
		return fmt.Sprintf("%slet %s = %s;", prefix, s.Name, value)

	case *ast.SayStatement:
		g.addStmtMapping(s.Line)
		return fmt.Sprintf("%sconsole.log(%s);", prefix, g.genExpr(s.Value))

	case *ast.IfStatement:
		g.addStmtMapping(s.Line)
		return g.genIf(s, prefix)

	case *ast.ForEachStatement:
		g.addStmtMapping(s.Line)
		g.indent++
		body := g.genBlock(s.Body)
		g.indent--
		varPart := s.Variable
		if s.DestructurePattern != nil {
			varPart = g.genPattern(s.DestructurePattern)
		}
		awaitPart := ""
		if s.IsAsync {
			awaitPart = "await "
		}
		return fmt.Sprintf("%sfor %s(const %s of %s) {\n%s%s}", prefix, awaitPart, varPart, g.genExpr(s.Iterable), body, prefix)

	case *ast.WhileStatement:
		g.addStmtMapping(s.Line)
		g.indent++
		body := g.genBlock(s.Body)
		g.indent--
		return fmt.Sprintf("%swhile (%s) {\n%s%s}", prefix, g.genExpr(s.Condition), body, prefix)

	case *ast.FuncDefinition:
		g.addStmtMapping(s.Line)
		// Save the current declared map and create a new scope for the function
		// Copy the outer scope so the function knows about already-declared variables
		outerDeclared := g.declared
		g.declared = make(map[string]bool)
		for k, v := range outerDeclared {
			g.declared[k] = v
		}
		// Mark function params as declared in the new scope
		for _, param := range s.Params {
			g.declared[param] = true
		}
		g.indent++
		body := g.genBlock(s.Body)
		g.indent--
		// Restore the outer scope
		g.declared = outerDeclared
		params := strings.Join(s.Params, ", ")
		if g.bodyContainsYield(s.Body) {
			return fmt.Sprintf("%sfunction* %s(%s) {\n%s%s}", prefix, s.Name, params, body, prefix)
		}
		if g.bodyContainsAwait(s.Body) {
			return fmt.Sprintf("%sasync function %s(%s) {\n%s%s}", prefix, s.Name, params, body, prefix)
		}
		return fmt.Sprintf("%sfunction %s(%s) {\n%s%s}", prefix, s.Name, params, body, prefix)

	case *ast.YieldStatement:
		g.addStmtMapping(s.Line)
		return fmt.Sprintf("%syield %s;", prefix, g.genExpr(s.Value))

	case *ast.LoopStatement:
		g.addStmtMapping(s.Line)
		g.indent++
		body := g.genBlock(s.Body)
		g.indent--
		return fmt.Sprintf("%swhile (true) {\n%s%s}", prefix, body, prefix)

	case *ast.ReturnStatement:
		g.addStmtMapping(s.Line)
		if s.Value == nil {
			return fmt.Sprintf("%sreturn;", prefix)
		}
		return fmt.Sprintf("%sreturn %s;", prefix, g.genExpr(s.Value))

	case *ast.ExprStatement:
		g.addStmtMapping(s.Line)
		return fmt.Sprintf("%s%s;", prefix, g.genExpr(s.Expr))

	case *ast.DotAssignStatement:
		return fmt.Sprintf("%s%s.%s = %s;", prefix, s.Object, s.Field, g.genExpr(s.Value))

	case *ast.IndexAssignStatement:
		return fmt.Sprintf("%s%s[%s] = %s;", prefix, g.genExpr(s.Object), g.genExpr(s.Index), g.genExpr(s.Value))

	case *ast.DescribeStatement:
		var dout strings.Builder
		if s.Extends != "" {
			dout.WriteString(fmt.Sprintf("%sclass %s extends %s {\n", prefix, s.Name, s.Extends))
		} else {
			dout.WriteString(fmt.Sprintf("%sclass %s {\n", prefix, s.Name))
		}
		g.indent++
		innerPrefix := strings.Repeat("  ", g.indent)

		// Emit private field declarations (JS requires # fields to be declared)
		for i, prop := range s.Properties {
			vis := ""
			if i < len(s.PropertyVisibilities) {
				vis = s.PropertyVisibilities[i]
			}
			if vis == "private" {
				dout.WriteString(fmt.Sprintf("%s#%s;\n", innerPrefix, prop.Name))
			}
		}

		// Constructor
		dout.WriteString(fmt.Sprintf("%sconstructor() {\n", innerPrefix))
		if s.Extends != "" {
			g.indent++
			superPrefix := strings.Repeat("  ", g.indent)
			dout.WriteString(fmt.Sprintf("%ssuper();\n", superPrefix))
			g.indent--
		}
		g.indent++
		propPrefix := strings.Repeat("  ", g.indent)
		for i, prop := range s.Properties {
			vis := ""
			if i < len(s.PropertyVisibilities) {
				vis = s.PropertyVisibilities[i]
			}
			if vis == "private" {
				dout.WriteString(fmt.Sprintf("%sthis.#%s = %s;\n", propPrefix, prop.Name, g.genExpr(prop.Value)))
			} else {
				dout.WriteString(fmt.Sprintf("%sthis.%s = %s;\n", propPrefix, prop.Name, g.genExpr(prop.Value)))
			}
		}
		g.indent--
		dout.WriteString(fmt.Sprintf("%s}\n", innerPrefix))
		// Methods
		for i, method := range s.Methods {
			vis := ""
			if i < len(s.MethodVisibilities) {
				vis = s.MethodVisibilities[i]
			}
			params := strings.Join(method.Params, ", ")
			asyncPrefix := ""
			if g.bodyContainsAwait(method.Body) {
				asyncPrefix = "async "
			}
			if vis == "private" {
				dout.WriteString(fmt.Sprintf("%s%s#%s(%s) {\n", innerPrefix, asyncPrefix, method.Name, params))
			} else {
				dout.WriteString(fmt.Sprintf("%s%s%s(%s) {\n", innerPrefix, asyncPrefix, method.Name, params))
			}
			// Save and reset declared scope for each method
			savedDeclared := g.declared
			g.declared = make(map[string]bool)
			for _, p := range method.Params {
				g.declared[p] = true
			}
			g.indent++
			for _, stmt := range method.Body {
				dout.WriteString(g.genStmt(stmt))
				dout.WriteString("\n")
			}
			g.indent--
			g.declared = savedDeclared
			dout.WriteString(fmt.Sprintf("%s}\n", innerPrefix))
		}
		g.indent--
		dout.WriteString(fmt.Sprintf("%s}", prefix))
		return dout.String()

	case *ast.UseStatement:
		if strings.HasSuffix(s.Path, ".quill") {
			// Circular import protection
			if g.importedFiles[s.Path] {
				return fmt.Sprintf("%s// skipped circular import of %q", prefix, s.Path)
			}
			g.importedFiles[s.Path] = true
			imported, err := os.ReadFile(s.Path)
			if err != nil {
				return fmt.Sprintf("%s// Error: could not import %q", prefix, s.Path)
			}
			l := lexer.New(string(imported))
			tokens, err := l.Tokenize()
			if err != nil {
				return fmt.Sprintf("%s// Error: could not tokenize %q: %s", prefix, s.Path, err)
			}
			pr := parser.New(tokens)
			prog, err := pr.Parse()
			if err != nil || prog == nil {
				return fmt.Sprintf("%s// Error: could not parse %q", prefix, s.Path)
			}
			var importGen *Generator
			if g.browser {
				importGen = NewBrowser()
			} else {
				importGen = New()
			}
			importGen.indent = g.indent
			// Share the importedFiles map to prevent circular imports across transitive imports
			importGen.importedFiles = g.importedFiles
			code := importGen.GenerateBody(prog)
			return fmt.Sprintf("%s// imported from %q\n%s", prefix, s.Path, code)
		}
		// NPM package - use import for worker mode, require for Node
		varName := strings.ReplaceAll(s.Path, "-", "_")
		varName = strings.ReplaceAll(varName, "/", "_")
		varName = strings.ReplaceAll(varName, "@", "")
		varName = strings.ReplaceAll(varName, ".", "_")
		if s.Alias != "" {
			varName = s.Alias
		}
		if g.workerMode {
			return fmt.Sprintf("%simport %s from \"%s\";", prefix, varName, s.Path)
		}
		return fmt.Sprintf("%sconst %s = require(\"%s\");", prefix, varName, s.Path)

	case *ast.TestBlock:
		g.indent++
		body := g.genBlock(s.Body)
		g.indent--
		return fmt.Sprintf(`%s(() => {
%s  try {
%s%s    __test_passed++;
%s    console.log("  \u2713 %s");
%s  } catch(e) {
%s    __test_failed++;
%s    console.log("  \u2717 %s:", e.message);
%s  }
%s})();`, prefix, prefix, prefix, body, prefix, s.Name, prefix, prefix, prefix, s.Name, prefix, prefix)

	case *ast.MockStatement:
		var mout strings.Builder
		mout.WriteString(fmt.Sprintf("%sconst __mock_%s_calls = [];\n", prefix, s.FuncName))
		mout.WriteString(fmt.Sprintf("%sconst __original_%s = typeof %s !== 'undefined' ? %s : undefined;\n", prefix, s.FuncName, s.FuncName, s.FuncName))
		params := strings.Join(s.Params, ", ")
		mout.WriteString(fmt.Sprintf("%svar %s = function(%s) {\n", prefix, s.FuncName, params))
		args := s.Params
		argsArray := "[]"
		if len(args) > 0 {
			argsArray = "[" + strings.Join(args, ", ") + "]"
		}
		mout.WriteString(fmt.Sprintf("%s  __mock_%s_calls.push({args: %s});\n", prefix, s.FuncName, argsArray))
		g.indent += 2
		for _, stmt := range s.Body {
			mout.WriteString(g.genStmt(stmt))
			mout.WriteString("\n")
		}
		g.indent -= 2
		mout.WriteString(fmt.Sprintf("%s};\n", prefix))
		return mout.String()

	case *ast.ExpectStatement:
		if ma, ok := s.Expr.(*ast.MockAssertionExpr); ok {
			return fmt.Sprintf(`%sif (__mock_%s_calls.length !== %d) throw new Error("Expected %s to be called %d time(s), was called " + __mock_%s_calls.length + " time(s)");`,
				prefix, ma.FuncName, ma.Count, ma.FuncName, ma.Count, ma.FuncName)
		}
		exprStr := g.genExpr(s.Expr)
		return fmt.Sprintf(`%sif (!(%s)) throw new Error("Expected %s to be true");`, prefix, exprStr, escapeJS(exprStr))

	case *ast.TryCatchStatement:
		g.indent++
		tryBody := g.genBlock(s.TryBody)
		g.indent--
		g.indent++
		catchBody := g.genBlock(s.CatchBody)
		g.indent--
		errorVar := s.ErrorVar
		if errorVar == "" {
			errorVar = "error"
		}
		return fmt.Sprintf("%stry {\n%s%s} catch (%s) {\n%s%s}", prefix, tryBody, prefix, errorVar, catchBody, prefix)

	case *ast.RaiseStatement:
		g.addStmtMapping(s.Line)
		expr := g.genExpr(s.Value)
		// If the expression is a plain string literal, wrap in new Error()
		if strings.HasPrefix(expr, "\"") || strings.HasPrefix(expr, "`") {
			return fmt.Sprintf("%sthrow new Error(%s);", prefix, expr)
		}
		// Otherwise throw the value directly (could be an Error object already)
		return fmt.Sprintf("%sthrow %s;", prefix, expr)

	case *ast.DeleteStatement:
		return fmt.Sprintf("%sdelete %s;", prefix, g.genExpr(s.Target))

	case *ast.BreakStatement:
		return fmt.Sprintf("%sbreak;", prefix)

	case *ast.ContinueStatement:
		return fmt.Sprintf("%scontinue;", prefix)

	case *ast.MatchStatement:
		return g.genMatch(s, prefix)

	case *ast.DefineStatement:
		return g.genDefine(s, prefix)

	case *ast.TypedAssignStatement:
		value := g.genExpr(s.Value)
		if g.declared[s.Name] {
			return fmt.Sprintf("%s%s = %s;", prefix, s.Name, value)
		}
		g.declared[s.Name] = true
		return fmt.Sprintf("%slet %s = %s;", prefix, s.Name, value)

	case *ast.FromUseStatement:
		// Build import names with optional aliases
		namesParts := make([]string, len(s.Names))
		for i, n := range s.Names {
			if i < len(s.Aliases) && s.Aliases[i] != "" {
				namesParts[i] = n + ": " + s.Aliases[i]
			} else {
				namesParts[i] = n
			}
		}
		names := strings.Join(namesParts, ", ")
		if strings.HasSuffix(s.Path, ".quill") {
			// Circular import protection
			if g.importedFiles[s.Path] {
				return fmt.Sprintf("%s// skipped circular import of %q", prefix, s.Path)
			}
			g.importedFiles[s.Path] = true
			imported, err := os.ReadFile(s.Path)
			if err != nil {
				return fmt.Sprintf("%s// Error: could not import %q", prefix, s.Path)
			}
			l := lexer.New(string(imported))
			tokens, err := l.Tokenize()
			if err != nil {
				return fmt.Sprintf("%s// Error: could not tokenize %q: %s", prefix, s.Path, err)
			}
			pr := parser.New(tokens)
			prog, err := pr.Parse()
			if err != nil || prog == nil {
				return fmt.Sprintf("%s// Error: could not parse %q", prefix, s.Path)
			}
			var importGen *Generator
			if g.browser {
				importGen = NewBrowser()
			} else {
				importGen = New()
			}
			importGen.indent = g.indent
			importGen.importedFiles = g.importedFiles
			code := importGen.GenerateBody(prog)
			return fmt.Sprintf("%s// imported from %q\n%s", prefix, s.Path, code)
		}
		if g.workerMode {
			return fmt.Sprintf("%simport { %s } from \"%s\";", prefix, names, s.Path)
		}
		return fmt.Sprintf("%sconst { %s } = require(\"%s\");", prefix, names, s.Path)

	case *ast.ComponentStatement:
		return g.genComponent(s, prefix)

	case *ast.MountStatement:
		return g.genMount(s, prefix)

	case *ast.CancelStatement:
		g.addStmtMapping(s.Line)
		return fmt.Sprintf("%s__abort_%s.abort();", prefix, s.Target)

	case *ast.SpawnStatement:
		g.addStmtMapping(s.Line)
		return g.genSpawn(s, prefix)

	case *ast.ParallelBlock:
		g.addStmtMapping(s.Line)
		return g.genParallel(s, prefix)

	case *ast.RaceBlock:
		g.addStmtMapping(s.Line)
		return g.genRaceBlock(s, prefix)

	case *ast.ChannelStatement:
		g.addStmtMapping(s.Line)
		return g.genChannelStmt(s, prefix)

	case *ast.SendStatement:
		g.addStmtMapping(s.Line)
		return fmt.Sprintf("%sawait %s.send(%s);", prefix, s.Channel, g.genExpr(s.Value))

	case *ast.SelectStatement:
		g.addStmtMapping(s.Line)
		return g.genSelectStmt(s, prefix)

	case *ast.TraitDeclaration:
		return g.genTrait(s, prefix)

	case *ast.DestructureStatement:
		return g.genDestructure(s, prefix)

	case *ast.RespondStatement:
		g.addStmtMapping(s.Line)
		if g.workerMode {
			return g.genWorkerRespond(s, prefix)
		}
		statusCode := s.StatusCode
		if statusCode == 0 {
			statusCode = 200
		}
		return fmt.Sprintf("%sres.writeHead(%d, {'Content-Type': 'application/json'}); res.end(JSON.stringify(%s));", prefix, statusCode, g.genExpr(s.Value))

	case *ast.EveryStatement:
		g.addStmtMapping(s.Line)
		ms := s.Interval
		switch s.Unit {
		case "seconds":
			ms = s.Interval * 1000
		case "minutes":
			ms = s.Interval * 60000
		case "hours":
			ms = s.Interval * 3600000
		}
		g.indent++
		body := g.genBlock(s.Body)
		g.indent--
		return fmt.Sprintf("%ssetInterval(() => {\n%s%s}, %d);", prefix, body, prefix, ms)

	case *ast.AgentStatement:
		g.addStmtMapping(s.Line)
		g.needsAgentRuntime = true
		// Also need provider runtime for the agent
		switch s.Provider {
		case "openai":
			g.needsOpenAIRuntime = true
		case "gemini":
			g.needsGeminiRuntime = true
		case "ollama":
			g.needsOllamaRuntime = true
		default:
			g.needsAIRuntime = true
		}
		var out strings.Builder
		agentVar := fmt.Sprintf("__agent_%s", strings.ReplaceAll(s.Name, " ", "_"))
		providerStr := s.Provider
		if providerStr == "" {
			providerStr = "claude"
		}
		out.WriteString(fmt.Sprintf("%sconst %s = createAgent(%q, { provider: %q", prefix, agentVar, s.Name, providerStr))
		if m, ok := s.Options["model"]; ok {
			out.WriteString(fmt.Sprintf(", model: %s", g.genExpr(m)))
		}
		if sys, ok := s.Options["system"]; ok {
			out.WriteString(fmt.Sprintf(", system: %s", g.genExpr(sys)))
		}
		out.WriteString(" });\n")
		// Register tools
		for _, tool := range s.Tools {
			out.WriteString(fmt.Sprintf("%s%s.addTool(%q, %q, %s);\n", prefix, agentVar, tool, tool+" tool", tool))
		}
		// Generate body (agent.run call and handling)
		g.indent++
		out.WriteString(g.genBlock(s.Body))
		g.indent--
		return out.String()

	case *ast.NavigateStatement:
		g.addStmtMapping(s.Line)
		if s.Params != nil {
			return fmt.Sprintf("%snavigation.navigate(\"%s\", %s);", prefix, s.Screen, g.genExpr(s.Params))
		}
		return fmt.Sprintf("%snavigation.navigate(\"%s\");", prefix, s.Screen)

	case *ast.NavigationBlock:
		return prefix + "/* navigation block — use 'quill build --expo' for Expo mode */"

	case *ast.WorkerHandler:
		g.addStmtMapping(s.Line)
		return g.genWorkerHandler(s, prefix)

	case *ast.ServerBlockStatement:
		return prefix + "/* server block — use 'quill run' for full-stack mode */"

	case *ast.DatabaseBlockStatement:
		return prefix + "/* database block — use 'quill run' for full-stack mode */"

	case *ast.AuthBlockStatement:
		return prefix + "/* auth block — use 'quill run' for full-stack mode */"

	case *ast.TypeAliasStatement:
		return g.genTypeAlias(s, prefix)

	case *ast.DecoratedFuncDefinition:
		return g.genDecoratedFunc(s, prefix)

	case *ast.DecoratedRouteDefinition:
		return g.genDecoratedRoute(s, prefix)

	case *ast.BroadcastStatement:
		return fmt.Sprintf("%sfor (const __c of __ws_clients) { if (__c !== client && __c.readyState === 1) { __c.send(%s); } }", prefix, g.genExpr(s.Value))

	case *ast.WebSocketBlock:
		return g.genWebSocket(s, prefix)

	case *ast.CommandStatement:
		// Command statements are collected and handled in Generate()
		// This should not normally be reached directly
		return prefix + "/* command statement — handled at top level */"

	case *ast.ReplyStatement:
		g.addStmtMapping(s.Line)
		return fmt.Sprintf("%sinteraction.reply(%s);", prefix, g.genExpr(s.Value))

	case *ast.StreamStatement:
		g.addStmtMapping(s.Line)
		var out strings.Builder
		prompt := g.genExpr(s.Prompt)

		// Build options for non-claude providers
		optParts := []string{}
		if m, ok := s.Options["model"]; ok {
			optParts = append(optParts, fmt.Sprintf("model: %s", g.genExpr(m)))
		}
		if mt, ok := s.Options["max_tokens"]; ok {
			optParts = append(optParts, fmt.Sprintf("max_tokens: %s", g.genExpr(mt)))
		}
		if sys, ok := s.Options["system"]; ok {
			optParts = append(optParts, fmt.Sprintf("system: %s", g.genExpr(sys)))
		}
		if temp, ok := s.Options["temperature"]; ok {
			optParts = append(optParts, fmt.Sprintf("temperature: %s", g.genExpr(temp)))
		}
		opts := "{" + strings.Join(optParts, ", ") + "}"

		switch s.Provider {
		case "openai", "gemini", "ollama":
			switch s.Provider {
			case "openai":
				g.needsOpenAIRuntime = true
			case "gemini":
				g.needsGeminiRuntime = true
			case "ollama":
				g.needsOllamaRuntime = true
			}
			streamFn := fmt.Sprintf("__stream_%s", s.Provider)
			out.WriteString(fmt.Sprintf("%sfor await (const %s of %s(%s, %s)) {\n", prefix, s.ChunkVar, streamFn, prompt, opts))
			g.indent++
			out.WriteString(g.genBlock(s.Body))
			g.indent--
			out.WriteString(fmt.Sprintf("%s}", prefix))
		default: // claude
			g.needsAIRuntime = true
			model := `"claude-sonnet-4-20250514"`
			maxTokens := "1024"
			if m, ok := s.Options["model"]; ok {
				model = g.genExpr(m)
			}
			if mt, ok := s.Options["max_tokens"]; ok {
				maxTokens = g.genExpr(mt)
			}
			out.WriteString(fmt.Sprintf("%sconst __stream = await __ai_client.messages.stream({\n", prefix))
			out.WriteString(fmt.Sprintf("%s  model: %s,\n", prefix, model))
			out.WriteString(fmt.Sprintf("%s  max_tokens: %s,\n", prefix, maxTokens))
			if sys, ok := s.Options["system"]; ok {
				out.WriteString(fmt.Sprintf("%s  system: %s,\n", prefix, g.genExpr(sys)))
			}
			if temp, ok := s.Options["temperature"]; ok {
				out.WriteString(fmt.Sprintf("%s  temperature: %s,\n", prefix, g.genExpr(temp)))
			}
			out.WriteString(fmt.Sprintf("%s  messages: [{ role: \"user\", content: %s }]\n", prefix, g.genExpr(s.Prompt)))
			out.WriteString(fmt.Sprintf("%s});\n", prefix))
			out.WriteString(fmt.Sprintf("%sfor await (const __event of __stream) {\n", prefix))
			out.WriteString(fmt.Sprintf("%s  if (__event.type === \"content_block_delta\") {\n", prefix))
			out.WriteString(fmt.Sprintf("%s    const %s = __event.delta.text;\n", prefix, s.ChunkVar))
			g.indent += 2
			out.WriteString(g.genBlock(s.Body))
			g.indent -= 2
			out.WriteString(fmt.Sprintf("%s  }\n", prefix))
			out.WriteString(fmt.Sprintf("%s}", prefix))
		}
		return out.String()

	case *ast.OnStatement:
		var out strings.Builder
		params := strings.Join(s.Params, ", ")
		if s.Method != "" {
			// Route handler: app.get("/path", (req, res) => { ... })
			out.WriteString(fmt.Sprintf("%s%s.%s(\"%s\", ", prefix, g.genExpr(s.Object), s.Method, s.Path))
		} else {
			// Event handler: obj.on("event", ...)
			out.WriteString(fmt.Sprintf("%s%s.on(\"%s\", ", prefix, g.genExpr(s.Object), s.Event))
		}
		if len(s.Params) > 0 {
			out.WriteString(fmt.Sprintf("(%s) => {\n", params))
		} else {
			out.WriteString("() => {\n")
		}
		g.indent++
		out.WriteString(g.genBlock(s.Body))
		g.indent--
		out.WriteString(fmt.Sprintf("%s});", prefix))
		return out.String()

	default:
		return prefix + "/* unknown statement */"
	}
}

func (g *Generator) genIf(s *ast.IfStatement, prefix string) string {
	var out strings.Builder

	g.indent++
	body := g.genBlock(s.Body)
	g.indent--

	out.WriteString(fmt.Sprintf("%sif (%s) {\n%s%s}", prefix, g.genExpr(s.Condition), body, prefix))

	for _, elif := range s.ElseIfs {
		g.indent++
		elifBody := g.genBlock(elif.Body)
		g.indent--
		out.WriteString(fmt.Sprintf(" else if (%s) {\n%s%s}", g.genExpr(elif.Condition), elifBody, prefix))
	}

	if len(s.Else) > 0 {
		g.indent++
		elseBody := g.genBlock(s.Else)
		g.indent--
		out.WriteString(fmt.Sprintf(" else {\n%s%s}", elseBody, prefix))
	}

	return out.String()
}

func (g *Generator) genMatch(s *ast.MatchStatement, prefix string) string {
	var out strings.Builder
	matchVar := "__match_val"
	out.WriteString(fmt.Sprintf("%s{\n", prefix))
	g.indent++
	innerPrefix := strings.Repeat("  ", g.indent)
	out.WriteString(fmt.Sprintf("%sconst %s = %s;\n", innerPrefix, matchVar, g.genExpr(s.Value)))

	for i, c := range s.Cases {
		// Build body with possible binding prefix
		var bodyStr string
		if c.TypePattern != "" && c.Binding != "" {
			// Inject binding variable at the top of the body
			g.indent++
			bindPrefix := strings.Repeat("  ", g.indent)
			bodyStr = fmt.Sprintf("%slet %s = %s;\n", bindPrefix, c.Binding, matchVar)
			bodyStr += g.genBlock(c.Body)
			g.indent--
		} else {
			g.indent++
			bodyStr = g.genBlock(c.Body)
			g.indent--
		}

		if c.TypePattern != "" {
			// Type-based pattern matching
			condition := g.genTypeCondition(c.TypePattern, matchVar)
			if c.Guard != nil {
				condition = fmt.Sprintf("(%s && %s)", condition, g.genExpr(c.Guard))
			}
			if i == 0 {
				out.WriteString(fmt.Sprintf("%sif %s {\n%s%s}", innerPrefix, condition, bodyStr, innerPrefix))
			} else {
				out.WriteString(fmt.Sprintf(" else if %s {\n%s%s}", condition, bodyStr, innerPrefix))
			}
		} else if objPat, ok := c.Pattern.(*ast.ObjectMatchPattern); ok {
			// Object destructuring pattern
			var conditions []string
			var bindings []string
			for _, f := range objPat.Fields {
				if f.Value != nil {
					conditions = append(conditions, fmt.Sprintf("%s.%s === %s", matchVar, f.Key, g.genExpr(f.Value)))
				} else {
					// Just bind the field
					bindings = append(bindings, f.Key)
				}
			}
			condition := "true"
			if len(conditions) > 0 {
				condition = strings.Join(conditions, " && ")
			}
			if c.Guard != nil {
				condition = fmt.Sprintf("(%s && %s)", condition, g.genExpr(c.Guard))
			}

			// Inject bindings at the top of the body
			g.indent++
			bindPrefix := strings.Repeat("  ", g.indent)
			var bindStr string
			for _, b := range bindings {
				bindStr += fmt.Sprintf("%sconst %s = %s.%s;\n", bindPrefix, b, matchVar, b)
			}
			bodyStr = bindStr + bodyStr
			g.indent--

			if i == 0 {
				out.WriteString(fmt.Sprintf("%sif (%s) {\n%s%s}", innerPrefix, condition, bodyStr, innerPrefix))
			} else {
				out.WriteString(fmt.Sprintf(" else if (%s) {\n%s%s}", condition, bodyStr, innerPrefix))
			}
		} else if c.Pattern == nil {
			// otherwise case
			if i == 0 {
				out.WriteString(fmt.Sprintf("%sif (true) {\n%s%s}", innerPrefix, bodyStr, innerPrefix))
			} else {
				out.WriteString(fmt.Sprintf(" else {\n%s%s}", bodyStr, innerPrefix))
			}
		} else {
			// Value-based pattern
			condition := fmt.Sprintf("(%s === %s)", matchVar, g.genExpr(c.Pattern))
			if c.Guard != nil {
				condition = fmt.Sprintf("(%s === %s && %s)", matchVar, g.genExpr(c.Pattern), g.genExpr(c.Guard))
			}

			if i == 0 {
				out.WriteString(fmt.Sprintf("%sif %s {\n%s%s}", innerPrefix, condition, bodyStr, innerPrefix))
			} else {
				out.WriteString(fmt.Sprintf(" else if %s {\n%s%s}", condition, bodyStr, innerPrefix))
			}
		}
	}
	out.WriteString("\n")
	g.indent--
	out.WriteString(fmt.Sprintf("%s}", prefix))
	return out.String()
}

func (g *Generator) genDefine(s *ast.DefineStatement, prefix string) string {
	var out strings.Builder

	// Check if any variant has an associated value (enum with values + methods)
	hasValues := false
	for _, v := range s.Variants {
		if v.Value != nil {
			hasValues = true
			break
		}
	}

	if hasValues && len(s.Methods) > 0 {
		// Generate class-based enum with methods
		enumClassName := s.Name + "Enum"
		out.WriteString(fmt.Sprintf("%sconst %s = (() => {\n", prefix, s.Name))
		g.indent++
		ip := strings.Repeat("  ", g.indent)
		out.WriteString(fmt.Sprintf("%sclass %s {\n", ip, enumClassName))
		g.indent++
		ip2 := strings.Repeat("  ", g.indent)
		out.WriteString(fmt.Sprintf("%sconstructor(name, value) { this.name = name; this.value = value; Object.freeze(this); }\n", ip2))
		for _, method := range s.Methods {
			params := strings.Join(method.Params, ", ")
			out.WriteString(fmt.Sprintf("%s%s(%s) {\n", ip2, method.Name, params))
			g.indent++
			for _, stmt := range method.Body {
				out.WriteString(g.genStmt(stmt))
				out.WriteString("\n")
			}
			g.indent--
			out.WriteString(fmt.Sprintf("%s}\n", ip2))
		}
		g.indent--
		out.WriteString(fmt.Sprintf("%s}\n", ip))
		out.WriteString(fmt.Sprintf("%sreturn Object.freeze({\n", ip))
		g.indent++
		ip3 := strings.Repeat("  ", g.indent)
		for i, variant := range s.Variants {
			valStr := "undefined"
			if variant.Value != nil {
				valStr = g.genExpr(variant.Value)
			}
			out.WriteString(fmt.Sprintf("%s%s: new %s(\"%s\", %s)", ip3, variant.Name, enumClassName, variant.Name, valStr))
			if i < len(s.Variants)-1 {
				out.WriteString(",")
			}
			out.WriteString("\n")
		}
		g.indent--
		out.WriteString(fmt.Sprintf("%s});\n", ip))
		g.indent--
		out.WriteString(fmt.Sprintf("%s})();\n", prefix))
	} else if hasValues {
		// Enum with values but no methods
		out.WriteString(fmt.Sprintf("%sconst %s = Object.freeze({\n", prefix, s.Name))
		g.indent++
		innerPrefix := strings.Repeat("  ", g.indent)
		for i, variant := range s.Variants {
			valStr := "undefined"
			if variant.Value != nil {
				valStr = g.genExpr(variant.Value)
			}
			out.WriteString(fmt.Sprintf("%s%s: Object.freeze({ type: \"%s\", variant: \"%s\", name: \"%s\", value: %s, toString() { return \"%s.%s\"; } })",
				innerPrefix, variant.Name, s.Name, variant.Name, variant.Name, valStr, s.Name, variant.Name))
			if i < len(s.Variants)-1 {
				out.WriteString(",")
			}
			out.WriteString("\n")
		}
		g.indent--
		out.WriteString(fmt.Sprintf("%s});\n", prefix))
	} else {
		// Original behavior: frozen object with variant constructors
		out.WriteString(fmt.Sprintf("%sconst %s = Object.freeze({\n", prefix, s.Name))
		g.indent++
		innerPrefix := strings.Repeat("  ", g.indent)

		for i, variant := range s.Variants {
			if len(variant.Fields) == 0 {
				out.WriteString(fmt.Sprintf("%s%s: Object.freeze({ type: \"%s\", variant: \"%s\", toString() { return \"%s.%s\"; } })", innerPrefix, variant.Name, s.Name, variant.Name, s.Name, variant.Name))
			} else {
				params := strings.Join(variant.Fields, ", ")
				fieldAssignments := make([]string, len(variant.Fields))
				for j, field := range variant.Fields {
					fieldAssignments[j] = fmt.Sprintf("%s: %s", field, field)
				}
				out.WriteString(fmt.Sprintf("%s%s: (%s) => Object.freeze({ type: \"%s\", variant: \"%s\", %s, toString() { return \"%s.%s(\" + [%s].join(\", \") + \")\"; } })",
					innerPrefix, variant.Name, params, s.Name, variant.Name, strings.Join(fieldAssignments, ", "), s.Name, variant.Name, params))
			}
			if i < len(s.Variants)-1 {
				out.WriteString(",")
			}
			out.WriteString("\n")
		}

		g.indent--
		out.WriteString(fmt.Sprintf("%s});\n", prefix))
	}

	// Add helper: is<Variant> functions
	for _, variant := range s.Variants {
		out.WriteString(fmt.Sprintf("%sconst is%s = (v) => v && v.variant === \"%s\";\n", prefix, variant.Name, variant.Name))
	}

	return out.String()
}

