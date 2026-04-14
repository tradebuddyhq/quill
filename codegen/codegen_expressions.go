package codegen

import (
	"fmt"
	"quill/ast"
	"strings"
)

func (g *Generator) genExpr(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		if hasInterpolation(e.Value) {
			// Convert {var} to ${var} for JS template literals
			converted := e.Value
			converted = strings.ReplaceAll(converted, "`", "\\`")
			// Convert my.field to this.field only inside interpolation braces
			converted = replaceMyInInterpolations(converted)
			converted = convertInterpolation(converted)
			return "`" + converted + "`"
		}
		// Multiline strings (from """) — use JS template literals to preserve newlines
		if strings.Contains(e.Value, "\n") {
			escaped := strings.ReplaceAll(e.Value, "`", "\\`")
			escaped = strings.ReplaceAll(escaped, "${", "\\${")
			return "`" + escaped + "`"
		}
		// The lexer preserves escape sequences as-is (e.g. \n stays as backslash + n),
		// which are already valid JS escape sequences. Only escape unescaped double quotes.
		escaped := strings.ReplaceAll(e.Value, "\n", "\\n")  // actual newline bytes (shouldn't occur, but safety)
		escaped = strings.ReplaceAll(escaped, "\r", "\\r")
		escaped = strings.ReplaceAll(escaped, "\t", "\\t")
		escaped = strings.ReplaceAll(escaped, "\x00", "\\0")
		return `"` + escaped + `"`

	case *ast.NumberLiteral:
		if e.Value == float64(int64(e.Value)) {
			return fmt.Sprintf("%d", int64(e.Value))
		}
		return fmt.Sprintf("%g", e.Value)

	case *ast.BoolLiteral:
		if e.Value {
			return "true"
		}
		return "false"

	case *ast.Identifier:
		return e.Name

	case *ast.ListLiteral:
		elems := make([]string, len(e.Elements))
		for i, el := range e.Elements {
			elems[i] = g.genExpr(el)
		}
		return "[" + strings.Join(elems, ", ") + "]"

	case *ast.BinaryExpr:
		if e.Operator == "^" {
			return fmt.Sprintf("(%s ** %s)", g.genExpr(e.Left), g.genExpr(e.Right))
		}
		return fmt.Sprintf("(%s %s %s)", g.genExpr(e.Left), e.Operator, g.genExpr(e.Right))

	case *ast.TernaryExpression:
		return fmt.Sprintf("(%s ? %s : %s)", g.genExpr(e.Condition), g.genExpr(e.Then), g.genExpr(e.Else))

	case *ast.ComparisonExpr:
		if e.Operator == "contains" {
			return fmt.Sprintf("__contains(%s, %s)", g.genExpr(e.Left), g.genExpr(e.Right))
		}
		op := e.Operator
		// For null comparisons, use == / != to catch both null and undefined
		_, leftIsNothing := e.Left.(*ast.NothingLiteral)
		_, rightIsNothing := e.Right.(*ast.NothingLiteral)
		isNullComparison := leftIsNothing || rightIsNothing
		if op == "==" && !isNullComparison {
			op = "==="
		}
		if op == "!=" && !isNullComparison {
			op = "!=="
		}
		return fmt.Sprintf("(%s %s %s)", g.genExpr(e.Left), op, g.genExpr(e.Right))

	case *ast.LogicalExpr:
		op := "&&"
		if e.Operator == "or" {
			op = "||"
		}
		return fmt.Sprintf("(%s %s %s)", g.genExpr(e.Left), op, g.genExpr(e.Right))

	case *ast.NotExpr:
		return fmt.Sprintf("(!%s)", g.genExpr(e.Operand))

	case *ast.UnaryMinusExpr:
		return fmt.Sprintf("(-%s)", g.genExpr(e.Operand))

	case *ast.CallExpr:
		// Check for Discord.bot(token) shorthand
		if dot, ok := e.Function.(*ast.DotExpr); ok {
			if ident, ok := dot.Object.(*ast.Identifier); ok && ident.Name == "Discord" && dot.Field == "bot" {
				tokenArg := "undefined"
				if len(e.Args) > 0 {
					tokenArg = g.genExpr(e.Args[0])
				}
				return fmt.Sprintf("(() => { const c = new Discord.Client({ intents: [Discord.GatewayIntentBits.Guilds, Discord.GatewayIntentBits.GuildMessages, Discord.GatewayIntentBits.MessageContent] }); c.__token = %s; process.nextTick(() => c.login(%s)); return c; })()", tokenArg, tokenArg)
			}
		}
		// Check for env("KEY") shorthand
		if ident, ok := e.Function.(*ast.Identifier); ok && ident.Name == "env" && len(e.Args) == 1 {
			return fmt.Sprintf("process.env[%s]", g.genExpr(e.Args[0]))
		}
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = g.genExpr(a)
		}
		return fmt.Sprintf("%s(%s)", g.genExpr(e.Function), strings.Join(args, ", "))

	case *ast.DotExpr:
		return fmt.Sprintf("%s.%s", g.genExpr(e.Object), e.Field)

	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", g.genExpr(e.Object), g.genExpr(e.Index))

	case *ast.NewExpr:
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = g.genExpr(a)
		}
		return fmt.Sprintf("new %s(%s)", e.ClassName, strings.Join(args, ", "))

	case *ast.AwaitExpr:
		return fmt.Sprintf("await %s", g.genExpr(e.Expr))

	case *ast.AwaitExpression:
		// "await all" or "await first" — handled contextually by parallel/race
		if ident, ok := e.Target.(*ast.Identifier); ok {
			if ident.Name == "all" || ident.Name == "first" {
				return fmt.Sprintf("await %s", ident.Name)
			}
		}
		return fmt.Sprintf("await __task_%s", g.genExpr(e.Target))

	case *ast.ReceiveExpression:
		return fmt.Sprintf("await %s.receive()", e.Channel)

	case *ast.NothingLiteral:
		return "null"

	case *ast.ObjectLiteral:
		if len(e.Keys) == 0 && len(e.ComputedProperties) == 0 {
			return "{}"
		}
		var pairs []string
		for i, key := range e.Keys {
			pairs = append(pairs, fmt.Sprintf("%s: %s", key, g.genExpr(e.Values[i])))
		}
		for _, cp := range e.ComputedProperties {
			pairs = append(pairs, fmt.Sprintf("[%s]: %s", g.genExpr(cp.KeyExpr), g.genExpr(cp.Value)))
		}
		return "{ " + strings.Join(pairs, ", ") + " }"

	case *ast.LambdaExpr:
		params := strings.Join(e.Params, ", ")
		if len(e.BodyStatements) > 0 {
			g.indent++
			body := g.genBlock(e.BodyStatements)
			g.indent--
			prefix := strings.Repeat("  ", g.indent)
			return fmt.Sprintf("(%s) => {\n%s%s}", params, body, prefix)
		}
		body := g.genExpr(e.Body)
		// Wrap object literals in parens so JS doesn't treat { as a block
		if _, isObj := e.Body.(*ast.ObjectLiteral); isObj {
			return fmt.Sprintf("(%s) => (%s)", params, body)
		}
		return fmt.Sprintf("(%s) => %s", params, body)

	case *ast.EmbedLiteral:
		return g.genEmbedLiteral(e)

	case *ast.SpreadExpr:
		return fmt.Sprintf("...%s", g.genExpr(e.Expr))

	case *ast.TypeCheckExpr:
		exprCode := g.genExpr(e.Expr)
		switch e.TypeName {
		case "text":
			return fmt.Sprintf("(typeof %s === \"string\")", exprCode)
		case "number":
			return fmt.Sprintf("(typeof %s === \"number\")", exprCode)
		case "boolean":
			return fmt.Sprintf("(typeof %s === \"boolean\")", exprCode)
		case "nothing":
			return fmt.Sprintf("(%s === null || %s === undefined)", exprCode, exprCode)
		default:
			return fmt.Sprintf("(typeof %s === \"%s\")", exprCode, e.TypeName)
		}

	case *ast.PropagateExpr:
		inner := g.genExpr(e.Expr)
		g.needsResultRuntime = true
		return fmt.Sprintf("__propagate(%s)", inner)

	case *ast.TryExpression:
		inner := g.genExpr(e.Expr)
		g.needsResultRuntime = true
		return fmt.Sprintf("__tryResult(%s)", inner)

	case *ast.ObjectMatchPattern:
		// This should not be reached directly in genExpr, handled in genMatch
		return "/* object match pattern */"

	case *ast.PipeExpr:
		// x | fn  becomes  fn(x)
		// x | fn(a, b)  becomes  fn(x, a, b)
		leftCode := g.genExpr(e.Left)
		switch right := e.Right.(type) {
		case *ast.CallExpr:
			// Insert left as first argument
			args := []string{leftCode}
			for _, a := range right.Args {
				args = append(args, g.genExpr(a))
			}
			return fmt.Sprintf("%s(%s)", g.genExpr(right.Function), strings.Join(args, ", "))
		case *ast.Identifier:
			return fmt.Sprintf("%s(%s)", right.Name, leftCode)
		default:
			return fmt.Sprintf("%s(%s)", g.genExpr(e.Right), leftCode)
		}

	case *ast.TaggedTemplateExpr:
		// Convert Quill {x} interpolations to JS ${x} in the template
		// Use the same smart conversion as regular string interpolation
		converted := convertInterpolation(e.Template)
		g.taggedTemplatesUsed = append(g.taggedTemplatesUsed, e.Tag)
		return fmt.Sprintf("%s`%s`", e.Tag, converted)

	case *ast.MockAssertionExpr:
		// This is handled at the ExpectStatement level, but provide a fallback
		return fmt.Sprintf("__mock_%s_calls.length === %d", e.FuncName, e.Count)

	case *ast.AskExpression:
		var out strings.Builder
		prompt := g.genExpr(e.Prompt)

		// Build options object
		optParts := []string{}
		if m, ok := e.Options["model"]; ok {
			optParts = append(optParts, fmt.Sprintf("model: %s", g.genExpr(m)))
		}
		if mt, ok := e.Options["max_tokens"]; ok {
			optParts = append(optParts, fmt.Sprintf("max_tokens: %s", g.genExpr(mt)))
		}
		if sys, ok := e.Options["system"]; ok {
			optParts = append(optParts, fmt.Sprintf("system: %s", g.genExpr(sys)))
		}
		if temp, ok := e.Options["temperature"]; ok {
			optParts = append(optParts, fmt.Sprintf("temperature: %s", g.genExpr(temp)))
		}
		opts := "{" + strings.Join(optParts, ", ") + "}"

		// Build structured output schema if present
		hasStructured := len(e.StructuredOutput) > 0
		schemaStr := ""
		if hasStructured {
			schemaParts := []string{}
			for k, v := range e.StructuredOutput {
				schemaParts = append(schemaParts, fmt.Sprintf(`"%s": "%s"`, k, v))
			}
			schemaStr = "{" + strings.Join(schemaParts, ", ") + "}"
		}

		switch e.Provider {
		case "openai":
			g.needsOpenAIRuntime = true
			out.WriteString(fmt.Sprintf("(await __ask_openai(%s, %s))", prompt, opts))
		case "gemini":
			g.needsGeminiRuntime = true
			out.WriteString(fmt.Sprintf("(await __ask_gemini(%s, %s))", prompt, opts))
		case "ollama":
			g.needsOllamaRuntime = true
			out.WriteString(fmt.Sprintf("(await __ask_ollama(%s, %s))", prompt, opts))
		default: // "claude"
			g.needsAIRuntime = true
			model := `"claude-sonnet-4-20250514"`
			maxTokens := "1024"
			if m, ok := e.Options["model"]; ok {
				model = g.genExpr(m)
			}
			if mt, ok := e.Options["max_tokens"]; ok {
				maxTokens = g.genExpr(mt)
			}
			if e.IsMessages {
				out.WriteString(fmt.Sprintf("(await (async () => { const __ai_resp = await __ai_client.messages.create({ model: %s, max_tokens: %s", model, maxTokens))
			} else {
				out.WriteString(fmt.Sprintf("(await (async () => { const __ai_resp = await __ai_client.messages.create({ model: %s, max_tokens: %s", model, maxTokens))
			}
			if sys, ok := e.Options["system"]; ok {
				out.WriteString(fmt.Sprintf(", system: %s", g.genExpr(sys)))
			}
			if temp, ok := e.Options["temperature"]; ok {
				out.WriteString(fmt.Sprintf(", temperature: %s", g.genExpr(temp)))
			}
			if e.IsMessages {
				out.WriteString(fmt.Sprintf(", messages: %s", prompt))
			} else {
				out.WriteString(fmt.Sprintf(", messages: [{ role: \"user\", content: %s }]", prompt))
			}
			out.WriteString(" }); return __ai_resp.content[0].text; })())")
		}

		// Wrap with structured output parsing if needed
		if hasStructured {
			g.needsStructuredOutput = true
			return fmt.Sprintf("__parse_structured(%s, %s)", out.String(), schemaStr)
		}
		return out.String()

	case *ast.EmbedExpression:
		g.needsVectorRuntime = true
		provider := e.Provider
		if provider == "" {
			provider = "openai"
		}
		model := e.Model
		if model == "" {
			return fmt.Sprintf("(await embed(%s, %q))", g.genExpr(e.Text), provider)
		}
		return fmt.Sprintf("(await embed(%s, %q, %q))", g.genExpr(e.Text), provider, model)

	default:
		return "undefined"
	}
}

// --- Tagged template runtimes ---

func (g *Generator) injectTagRuntimes(out *strings.Builder) {
	injected := map[string]bool{}
	for _, tag := range g.taggedTemplatesUsed {
		if injected[tag] {
			continue
		}
		injected[tag] = true
		switch tag {
		case "query":
			out.WriteString(`function query(strings, ...values) {
  let text = strings[0];
  const params = [];
  for (let i = 0; i < values.length; i++) {
    params.push(values[i]);
    text += '$' + (i + 1) + strings[i + 1];
  }
  return { text, values: params };
}
`)
		case "html":
			out.WriteString(`function html(strings, ...values) {
  return strings.reduce((r, s, i) => r + (values[i-1] || '') + s);
}
`)
		case "css":
			out.WriteString(`function css(strings, ...values) {
  return strings.reduce((r, s, i) => r + (values[i-1] || '') + s);
}
`)
		}
	}
}

// --- Trait code generation ---

