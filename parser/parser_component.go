package parser

import (
	"fmt"
	"quill/ast"
	"quill/lexer"
	"strings"
)

func (p *Parser) parseComponent() *ast.ComponentStatement {
	line := p.current().Line
	p.advance() // consume "component"
	name := p.expect(lexer.TOKEN_IDENT)

	// Parse optional props: component MyScreen with navigation route:
	var hasProps bool
	var propNames []string
	if p.check(lexer.TOKEN_WITH) {
		p.advance() // consume "with"
		hasProps = true
		for (p.check(lexer.TOKEN_IDENT) || isKeywordToken(p.current().Type)) && !p.check(lexer.TOKEN_COLON) {
			propNames = append(propNames, p.advance().Value)
		}
	}

	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	var states []ast.StateDeclaration
	var effects []ast.EffectDeclaration
	var contexts []ast.UseContextDeclaration
	var memos []ast.MemoDeclaration
	var callbacks []ast.CallbackDeclaration
	var methods []ast.FuncDefinition
	var renderBody []ast.RenderElement
	var preRenderStmts []ast.Statement
	var styles *ast.StyleBlock
	var nativeStyles *ast.NativeStyleBlock
	var loader *ast.LoadFunction
	var actions []ast.FormAction
	var head *ast.HeadBlock

	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		if p.check(lexer.TOKEN_STATE) {
			states = append(states, *p.parseStateDecl())
		} else if p.check(lexer.TOKEN_USE) && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_EFFECT {
			effect := p.parseEffectDecl()
			effects = append(effects, *effect)
		} else if p.check(lexer.TOKEN_USE) && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_IDENT && p.tokens[p.pos+1].Value == "context" {
			ctx := p.parseUseContextDecl()
			contexts = append(contexts, *ctx)
		} else if p.check(lexer.TOKEN_IDENT) && p.current().Value == "memo" {
			m := p.parseMemoDecl()
			memos = append(memos, *m)
		} else if p.check(lexer.TOKEN_IDENT) && p.current().Value == "callback" {
			cb := p.parseCallbackDecl()
			callbacks = append(callbacks, *cb)
		} else if p.check(lexer.TOKEN_STYLE) {
			// Check if next-next token is IDENT (native style) or selector (CSS style)
			// For Expo components: "native style:" block
			if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_IDENT && p.tokens[p.pos+1].Value == "native" {
				nativeStyles = p.parseNativeStyleBlock()
			} else {
				styles = p.parseStyleBlock()
			}
		} else if p.check(lexer.TOKEN_HEAD) {
			head = p.parseHeadBlock()
		} else if p.check(lexer.TOKEN_TO) {
			if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Value == "render" {
				renderBody, preRenderStmts = p.parseRenderMethod()
			} else if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_LOAD &&
				p.pos+2 < len(p.tokens) && p.tokens[p.pos+2].Type == lexer.TOKEN_IDENT {
				loader = p.parseLoadFunction()
			} else {
				method := p.parseFuncDef()
				methods = append(methods, *method)
			}
		} else if p.check(lexer.TOKEN_FORM) {
			action := p.parseFormAction()
			actions = append(actions, action)
		} else {
			p.error("expected state, style, head, method (to), form, effect, context, memo, callback, or render inside component block")
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.ComponentStatement{
		Name:                name.Value,
		HasProps:            hasProps,
		Props:               propNames,
		States:              states,
		Effects:             effects,
		Contexts:            contexts,
		Memos:               memos,
		Callbacks:           callbacks,
		Methods:             methods,
		RenderBody:          renderBody,
		PreRenderStatements: preRenderStmts,
		Styles:              styles,
		NativeStyles:        nativeStyles,
		Loader:              loader,
		Actions:             actions,
		Head:                head,
		Line:                line,
	}
}

// parseEffectDecl parses: use effect [when [dep1, dep2]]:
func (p *Parser) parseEffectDecl() *ast.EffectDeclaration {
	line := p.current().Line
	p.advance() // consume "use"
	p.advance() // consume "effect"

	var deps []string
	if p.check(lexer.TOKEN_WHEN) {
		p.advance() // consume "when"
		p.expect(lexer.TOKEN_LBRACKET)
		for !p.check(lexer.TOKEN_RBRACKET) && !p.isAtEnd() {
			dep := p.expect(lexer.TOKEN_IDENT)
			deps = append(deps, dep.Value)
			if p.check(lexer.TOKEN_COMMA) {
				p.advance()
			}
		}
		p.expect(lexer.TOKEN_RBRACKET)
	}

	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	var body []ast.Statement
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		body = append(body, p.parseStatement())
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.EffectDeclaration{Dependencies: deps, Body: body, Line: line}
}

// parseUseContextDecl parses: use context ThemeContext as theme
func (p *Parser) parseUseContextDecl() *ast.UseContextDeclaration {
	line := p.current().Line
	p.advance() // consume "use"
	p.advance() // consume "context" (ident)

	contextName := p.expect(lexer.TOKEN_IDENT)

	alias := contextName.Value
	if p.check(lexer.TOKEN_AS) {
		p.advance() // consume "as"
		aliasTok := p.expect(lexer.TOKEN_IDENT)
		alias = aliasTok.Value
	}

	p.consumeNewline()
	return &ast.UseContextDeclaration{ContextName: contextName.Value, Alias: alias, Line: line}
}

// parseMemoDecl parses: memo total is computeTotal(data) when [data]
func (p *Parser) parseMemoDecl() *ast.MemoDeclaration {
	line := p.current().Line
	p.advance() // consume "memo"
	name := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_IS)
	value := p.parseExpression()

	var deps []string
	if p.check(lexer.TOKEN_WHEN) {
		p.advance() // consume "when"
		p.expect(lexer.TOKEN_LBRACKET)
		for !p.check(lexer.TOKEN_RBRACKET) && !p.isAtEnd() {
			dep := p.expect(lexer.TOKEN_IDENT)
			deps = append(deps, dep.Value)
			if p.check(lexer.TOKEN_COMMA) {
				p.advance()
			}
		}
		p.expect(lexer.TOKEN_RBRACKET)
	}

	p.consumeNewline()
	return &ast.MemoDeclaration{Name: name.Value, Value: value, Dependencies: deps, Line: line}
}

// parseCallbackDecl parses: callback handlePress is with item: doSomething(item) when [items]
func (p *Parser) parseCallbackDecl() *ast.CallbackDeclaration {
	line := p.current().Line
	p.advance() // consume "callback"
	name := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_IS)

	// Parse the function: with param1 param2: body
	var params []string
	if p.check(lexer.TOKEN_WITH) {
		p.advance() // consume "with"
		for p.check(lexer.TOKEN_IDENT) && !p.check(lexer.TOKEN_COLON) {
			params = append(params, p.advance().Value)
			if p.check(lexer.TOKEN_COMMA) {
				p.advance()
			}
		}
	}

	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	var body []ast.Statement
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		body = append(body, p.parseStatement())
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	var deps []string
	if p.check(lexer.TOKEN_WHEN) {
		p.advance() // consume "when"
		p.expect(lexer.TOKEN_LBRACKET)
		for !p.check(lexer.TOKEN_RBRACKET) && !p.isAtEnd() {
			dep := p.expect(lexer.TOKEN_IDENT)
			deps = append(deps, dep.Value)
			if p.check(lexer.TOKEN_COMMA) {
				p.advance()
			}
		}
		p.expect(lexer.TOKEN_RBRACKET)
	}
	p.consumeNewline()

	return &ast.CallbackDeclaration{Name: name.Value, Params: params, Body: body, Dependencies: deps, Line: line}
}

// parseContextDecl parses top-level: context ThemeContext is defaultValue
func (p *Parser) parseContextDecl() *ast.ContextDeclaration {
	line := p.current().Line
	p.advance() // consume "context"
	name := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_IS)
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.ContextDeclaration{Name: name.Value, DefaultValue: value, Line: line}
}

// parseNativeStyleBlock parses: style native:
func (p *Parser) parseNativeStyleBlock() *ast.NativeStyleBlock {
	line := p.current().Line
	p.advance() // consume "style"
	p.advance() // consume "native"
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	styles := make(map[string][]ast.NativeStyleProp)

	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}

		// Parse style name (e.g., "container")
		styleName := p.expectIdentOrKeyword()
		p.expect(lexer.TOKEN_COLON)
		p.expect(lexer.TOKEN_NEWLINE)
		p.expect(lexer.TOKEN_INDENT)

		var props []ast.NativeStyleProp
		for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
			p.skipNewlines()
			if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
				break
			}

			// Parse property: "font size is 24" or "background color is '#fff'"
			var nameParts []string
			for !p.check(lexer.TOKEN_IS) && !p.check(lexer.TOKEN_NEWLINE) && !p.isAtEnd() {
				tok := p.advance()
				nameParts = append(nameParts, tok.Value)
			}

			if p.check(lexer.TOKEN_IS) {
				p.advance() // consume "is"
				// Read value - could be string, number, boolean, or expression
				var val string
				if p.check(lexer.TOKEN_STRING) {
					val = "\"" + p.advance().Value + "\""
				} else if p.check(lexer.TOKEN_NUMBER) {
					val = p.advance().Value
				} else if p.check(lexer.TOKEN_YES) {
					p.advance()
					val = "true"
				} else if p.check(lexer.TOKEN_NO) {
					p.advance()
					val = "false"
				} else if p.check(lexer.TOKEN_IDENT) {
					// Could be a variable reference or expression
					val = p.advance().Value
					// Handle dot access (e.g., theme.primary)
					for p.check(lexer.TOKEN_DOT) {
						p.advance() // consume "."
						if p.check(lexer.TOKEN_IDENT) {
							val += "." + p.advance().Value
						}
					}
					// Handle simple arithmetic: width - 20
					for p.check(lexer.TOKEN_PLUS) || p.check(lexer.TOKEN_MINUS) || p.check(lexer.TOKEN_STAR) || p.check(lexer.TOKEN_SLASH) {
						op := p.advance().Value
						if p.check(lexer.TOKEN_NUMBER) {
							val += " " + op + " " + p.advance().Value
						} else if p.check(lexer.TOKEN_IDENT) {
							val += " " + op + " " + p.advance().Value
						}
					}
					// Handle comparison and ternary expressions
					// Translate Quill comparisons to JS operators and consume until newline
					for !p.check(lexer.TOKEN_NEWLINE) && !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
						tok := p.advance()
						if tok.Type == lexer.TOKEN_STRING {
							val += " \"" + tok.Value + "\""
						} else if tok.Type == lexer.TOKEN_QUESTION {
							val += " ? "
						} else if tok.Type == lexer.TOKEN_COLON {
							val += " : "
						} else if tok.Type == lexer.TOKEN_GREATER {
							// "greater than" → ">"
							if p.check(lexer.TOKEN_THAN) {
								p.advance() // consume "than"
							}
							val += " >"
						} else if tok.Type == lexer.TOKEN_LESS {
							// "less than" → "<"
							if p.check(lexer.TOKEN_THAN) {
								p.advance() // consume "than"
							}
							val += " <"
						} else {
							val += " " + tok.Value
						}
					}
				} else {
					val = p.advance().Value
				}

				// Convert multi-word property to camelCase
				propName := toCamelCase(nameParts)
				props = append(props, ast.NativeStyleProp{Name: propName, Value: val})
			}
			p.consumeNewline()
		}
		if p.check(lexer.TOKEN_DEDENT) {
			p.advance()
		}

		styles[styleName.Value] = props
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.NativeStyleBlock{Styles: styles, Line: line}
}

// toCamelCase converts ["font", "size"] to "fontSize"
func toCamelCase(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return result
}

// parseNavigateStmt parses: navigate to "ScreenName" [with { params }]
func (p *Parser) parseNavigateStmt() *ast.NavigateStatement {
	line := p.current().Line
	p.advance() // consume "navigate"
	p.expect(lexer.TOKEN_TO) // consume "to"

	screen := p.expect(lexer.TOKEN_STRING)

	var params ast.Expression
	if p.check(lexer.TOKEN_WITH) {
		p.advance() // consume "with"
		params = p.parseExpression()
	}

	return &ast.NavigateStatement{Screen: screen.Value, Params: params, Line: line}
}

// parseEveryStatement parses: every 5 seconds/minutes/hours:
func (p *Parser) parseEveryStatement() *ast.EveryStatement {
	line := p.current().Line
	p.advance() // consume "every"

	numTok := p.expect(lexer.TOKEN_NUMBER)
	interval := int(0)
	fmt.Sscanf(numTok.Value, "%d", &interval)

	// Parse unit: seconds, minutes, hours, second, minute, hour
	unitTok := p.expectIdentOrKeyword()
	unit := unitTok.Value
	// Normalize to plural
	switch unit {
	case "second":
		unit = "seconds"
	case "minute":
		unit = "minutes"
	case "hour":
		unit = "hours"
	}

	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	var body []ast.Statement
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		body = append(body, p.parseStatement())
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.EveryStatement{Interval: interval, Unit: unit, Body: body, Line: line}
}

// parseNavigationBlock parses: app navigation: stack: screen "Home" component HomeScreen
func (p *Parser) parseNavigationBlock() *ast.NavigationBlock {
	line := p.current().Line
	p.advance() // consume "app"
	p.advance() // consume "navigation" (consumed as ident)
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	navType := "stack"
	var screens []ast.ScreenDefinition

	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}

		// Parse navigation type: stack: / tab: / drawer:
		if p.check(lexer.TOKEN_IDENT) {
			typeTok := p.advance()
			navType = typeTok.Value
			p.expect(lexer.TOKEN_COLON)
			p.expect(lexer.TOKEN_NEWLINE)
			p.expect(lexer.TOKEN_INDENT)

			for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
				p.skipNewlines()
				if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
					break
				}
				if p.check(lexer.TOKEN_SCREEN) {
					p.advance() // consume "screen"
					screenName := p.expect(lexer.TOKEN_STRING)
					p.expect(lexer.TOKEN_COMPONENT)
					compName := p.expect(lexer.TOKEN_IDENT)
					screens = append(screens, ast.ScreenDefinition{
						Name:      screenName.Value,
						Component: compName.Value,
					})
					p.consumeNewline()
				} else {
					p.advance() // skip unknown
				}
			}
			if p.check(lexer.TOKEN_DEDENT) {
				p.advance()
			}
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.NavigationBlock{Type: navType, Screens: screens, Line: line}
}

func (p *Parser) parseStyleBlock() *ast.StyleBlock {
	p.advance() // consume "style"
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)
	var rules []ast.CSSRule
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		rule := p.parseCSSRule()
		rules = append(rules, rule)
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
	return &ast.StyleBlock{Rules: rules}
}

func (p *Parser) parseCSSRule() ast.CSSRule {
	var selectorParts []string
	for !p.check(lexer.TOKEN_COLON) && !p.check(lexer.TOKEN_NEWLINE) && !p.isAtEnd() {
		tok := p.advance()
		selectorParts = append(selectorParts, tok.Value)
	}
	selector := ""
	for i, part := range selectorParts {
		if i > 0 && part != ":" && selectorParts[i-1] != "." {
			selector += " "
		}
		selector += part
	}
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)
	properties := make(map[string]string)
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		var propParts []string
		for !p.check(lexer.TOKEN_IS) && !p.check(lexer.TOKEN_NEWLINE) && !p.isAtEnd() {
			tok := p.advance()
			propParts = append(propParts, tok.Value)
		}
		propName := ""
		for i, part := range propParts {
			if i > 0 {
				propName += "-"
			}
			propName += part
		}
		if p.check(lexer.TOKEN_IS) {
			p.advance()
			val := p.expect(lexer.TOKEN_STRING)
			properties[propName] = val.Value
		}
		p.consumeNewline()
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
	return ast.CSSRule{Selector: selector, Properties: properties}
}

func (p *Parser) parseLoadFunction() *ast.LoadFunction {
	p.advance() // consume "to"
	p.advance() // consume "load"
	param := "request"
	if p.check(lexer.TOKEN_IDENT) {
		param = p.advance().Value
	}
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()
	return &ast.LoadFunction{Param: param, Body: body}
}

func (p *Parser) parseHeadBlock() *ast.HeadBlock {
	line := p.current().Line
	p.advance() // consume "head"
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)
	var entries []ast.HeadEntry
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		tagTok := p.expect(lexer.TOKEN_IDENT)
		tag := tagTok.Value
		entry := ast.HeadEntry{Tag: tag, Attrs: make(map[string]string)}
		if tag == "title" {
			text := p.expect(lexer.TOKEN_STRING)
			entry.Text = text.Value
		} else {
			for p.check(lexer.TOKEN_IDENT) {
				attrName := p.advance().Value
				if p.check(lexer.TOKEN_STRING) {
					attrVal := p.advance().Value
					entry.Attrs[attrName] = attrVal
				}
			}
		}
		p.consumeNewline()
		entries = append(entries, entry)
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
	return &ast.HeadBlock{Entries: entries, Line: line}
}

func (p *Parser) parseFormAction() ast.FormAction {
	p.advance()                                      // consume "form"
	p.expect(lexer.TOKEN_IDENT)                      // "action"
	handlerName := p.expect(lexer.TOKEN_IDENT).Value // handler name
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)
	var body []ast.Statement
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		body = append(body, p.parseStatement())
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
	return ast.FormAction{Name: handlerName, Body: body}
}

func (p *Parser) parseStateDecl() *ast.StateDeclaration {
	line := p.current().Line
	p.advance() // consume "state"
	name := p.expect(lexer.TOKEN_IDENT)
	p.advance() // consume "is" or "are"
	value := p.parseExpression()
	p.consumeNewline()
	return &ast.StateDeclaration{Name: name.Value, Value: value, Line: line}
}

func (p *Parser) parseRenderMethod() ([]ast.RenderElement, []ast.Statement) {
	p.advance() // consume "to"
	p.advance() // consume "render"
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	var elements []ast.RenderElement
	var preStatements []ast.Statement
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		// Check if this is an assignment (variable is/are expr) before render elements
		if (p.check(lexer.TOKEN_IDENT) || isKeywordToken(p.current().Type)) && p.checkNext(lexer.TOKEN_IS, lexer.TOKEN_ARE) {
			stmt := p.parseAssignment()
			preStatements = append(preStatements, stmt)
		} else {
			elements = append(elements, *p.parseRenderElement())
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
	return elements, preStatements
}

func (p *Parser) parseRenderElement() *ast.RenderElement {
	line := p.current().Line

	// Handle link element: link to="/path" "text"
	if p.check(lexer.TOKEN_LINK) {
		p.advance() // consume "link"
		props := make(map[string]ast.Expression)
		// Parse to="path"
		if p.check(lexer.TOKEN_TO) {
			p.advance() // consume "to"
			if p.check(lexer.TOKEN_STRING) {
				toVal := p.advance()
				props["to"] = &ast.StringLiteral{Value: toVal.Value}
			}
		}
		var children []ast.RenderNode
		if p.check(lexer.TOKEN_STRING) {
			text := p.advance()
			children = append(children, ast.RenderNode{Text: &ast.StringLiteral{Value: text.Value}})
		}
		p.consumeNewline()
		return &ast.RenderElement{
			Tag:      "__link",
			Props:    props,
			Children: children,
			Line:     line,
		}
	}

	// Handle conditional: if condition:
	if p.check(lexer.TOKEN_IF) {
		p.advance() // consume "if"
		condition := p.parseExpression()
		p.expect(lexer.TOKEN_COLON)
		p.expect(lexer.TOKEN_NEWLINE)
		p.expect(lexer.TOKEN_INDENT)
		var children []ast.RenderNode
		for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
			p.skipNewlines()
			if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
				break
			}
			el := p.parseRenderElement()
			children = append(children, ast.RenderNode{Element: el})
		}
		if p.check(lexer.TOKEN_DEDENT) {
			p.advance()
		}

		// Check for otherwise: branch
		if p.check(lexer.TOKEN_OTHERWISE) {
			p.advance() // consume "otherwise"
			p.expect(lexer.TOKEN_COLON)
			p.expect(lexer.TOKEN_NEWLINE)
			p.expect(lexer.TOKEN_INDENT)
			var elseChildren []ast.RenderNode
			for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
				p.skipNewlines()
				if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
					break
				}
				el := p.parseRenderElement()
				elseChildren = append(elseChildren, ast.RenderNode{Element: el})
			}
			if p.check(lexer.TOKEN_DEDENT) {
				p.advance()
			}
			return &ast.RenderElement{
				Tag:          "__fragment",
				Condition:    condition,
				Children:     children,
				ElseChildren: elseChildren,
				Props:        map[string]ast.Expression{},
				Line:         line,
			}
		}

		return &ast.RenderElement{
			Tag:       "__fragment",
			Condition: condition,
			Children:  children,
			Props:     map[string]ast.Expression{},
			Line:      line,
		}
	}

	// Handle for each: for each item in list:
	if p.check(lexer.TOKEN_FOR) {
		p.advance() // consume "for"
		p.expect(lexer.TOKEN_EACH)
		varTok := p.expect(lexer.TOKEN_IDENT)
		p.expect(lexer.TOKEN_IN)
		iterable := p.parseExpression()
		p.expect(lexer.TOKEN_COLON)
		p.expect(lexer.TOKEN_NEWLINE)
		p.expect(lexer.TOKEN_INDENT)
		var children []ast.RenderNode
		for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
			p.skipNewlines()
			if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
				break
			}
			el := p.parseRenderElement()
			children = append(children, ast.RenderNode{Element: el})
		}
		if p.check(lexer.TOKEN_DEDENT) {
			p.advance()
		}
		return &ast.RenderElement{
			Tag:      "__fragment",
			Iterator: &ast.RenderIterator{Variable: varTok.Value, Iterable: iterable},
			Children: children,
			Props:    map[string]ast.Expression{},
			Line:     line,
		}
	}

	// Handle variable assignments in render blocks (e.g., pct is score / total * 100)
	if p.check(lexer.TOKEN_IDENT) && p.pos+1 < len(p.tokens) && (p.tokens[p.pos+1].Type == lexer.TOKEN_IS || p.tokens[p.pos+1].Type == lexer.TOKEN_ARE) {
		name := p.advance() // consume identifier
		p.advance()          // consume "is" or "are"
		value := p.parseExpression()
		p.consumeNewline()
		return &ast.RenderElement{
			Tag: "__assignment",
			Props: map[string]ast.Expression{
				"__name":  &ast.StringLiteral{Value: name.Value},
				"__value": value,
			},
			Line: line,
		}
	}

	// Handle children placeholder
	if p.check(lexer.TOKEN_IDENT) && p.current().Value == "children" {
		p.advance() // consume "children"
		p.consumeNewline()
		return &ast.RenderElement{
			Tag:   "__children",
			Props: map[string]ast.Expression{},
			Line:  line,
		}
	}

	// Handle provide: provide ThemeContext with value:
	if p.check(lexer.TOKEN_IDENT) && p.current().Value == "provide" {
		p.advance() // consume "provide"
		contextName := p.expect(lexer.TOKEN_IDENT)
		var valueExpr ast.Expression
		if p.check(lexer.TOKEN_WITH) {
			p.advance() // consume "with"
			valueExpr = p.parseExpression()
		}
		p.expect(lexer.TOKEN_COLON)
		p.expect(lexer.TOKEN_NEWLINE)
		p.expect(lexer.TOKEN_INDENT)
		var children []ast.RenderNode
		for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
			p.skipNewlines()
			if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
				break
			}
			el := p.parseRenderElement()
			children = append(children, ast.RenderNode{Element: el})
		}
		if p.check(lexer.TOKEN_DEDENT) {
			p.advance()
		}
		props := map[string]ast.Expression{
			"__context": &ast.StringLiteral{Value: contextName.Value},
		}
		if valueExpr != nil {
			props["value"] = valueExpr
		}
		return &ast.RenderElement{
			Tag:      "__provider",
			Props:    props,
			Children: children,
			Line:     line,
		}
	}

	// Regular element: tag [props...] [: "text" | NEWLINE INDENT children DEDENT]
	// Supports dot-notation for component namespaces (e.g., Nav.NavigationContainer, Stack.Screen)
	tag := p.expectIdentOrKeyword()
	for p.check(lexer.TOKEN_DOT) {
		p.advance() // consume "."
		next := p.expectIdentOrKeyword()
		tag.Value = tag.Value + "." + next.Value
	}
	props := make(map[string]ast.Expression)

	// Parse props: onClick handler, bind:value ident, key value, etc.
	// Supports both space-separated (onClick increment) and = syntax (onClick=increment)
	// Also handles keyword tokens like "style" as prop names
	for (p.check(lexer.TOKEN_IDENT) || p.check(lexer.TOKEN_STYLE) || (isKeywordToken(p.current().Type) && !p.check(lexer.TOKEN_COLON) && !p.check(lexer.TOKEN_NEWLINE) && !p.check(lexer.TOKEN_IF) && !p.check(lexer.TOKEN_FOR) && !p.check(lexer.TOKEN_OTHERWISE))) && !p.isAtEnd() {
		propName := p.current().Value
		// Check for bind:value pattern
		if propName == "bind" && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_COLON {
			p.advance() // consume "bind"
			p.advance() // consume ":"
			bindTarget := p.expect(lexer.TOKEN_IDENT)
			props["bind:"+bindTarget.Value] = &ast.Identifier{Name: bindTarget.Value}
			continue
		}
		// Check for prop=value syntax (e.g., onClick=increment, class="my-class")
		if p.pos+2 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_ASSIGN {
			p.advance() // consume prop name
			p.advance() // consume "="
			if p.check(lexer.TOKEN_IDENT) {
				valTok := p.advance()
				props[propName] = &ast.Identifier{Name: valTok.Value}
			} else if p.check(lexer.TOKEN_STRING) {
				valTok := p.advance()
				props[propName] = &ast.StringLiteral{Value: valTok.Value}
			} else {
				break
			}
			continue
		}
		// Inline style object: style { flexDirection: "row", gap: 10 }
		if propName == "style" && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_LBRACE {
			p.advance() // consume "style"
			expr := p.parseExpression() // parses the object literal { ... }
			props["__inlineStyle"] = expr
			continue
		}
		// Special: style [a, b] for multiple styles
		if propName == "style" && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_LBRACKET {
			p.advance() // consume "style"
			p.advance() // consume "["
			var parts []string
			for !p.check(lexer.TOKEN_RBRACKET) && !p.isAtEnd() {
				if p.check(lexer.TOKEN_IDENT) || isKeywordToken(p.current().Type) {
					parts = append(parts, p.advance().Value)
				} else if p.check(lexer.TOKEN_COMMA) {
					p.advance()
				} else {
					break // avoid infinite loop on unexpected tokens
				}
			}
			if p.check(lexer.TOKEN_RBRACKET) {
				p.advance()
			}
			props["__multiStyle"] = &ast.StringLiteral{Value: strings.Join(parts, ",")}
			continue
		}
		// Check for inline style object: style { key: val }
		if propName == "style" && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_LBRACE {
			p.advance() // consume "style"
			expr := p.parseExpression()
			props["__inlineStyle"] = expr
			continue
		}
		// Check for object literal as prop value: propName { key: val }
		if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_LBRACE {
			p.advance() // consume prop name
			expr := p.parseExpression()
			props[propName] = expr
			continue
		}
		// Check if next token is an identifier (event handler or attr with value)
		if p.checkNext(lexer.TOKEN_IDENT) {
			p.advance() // consume prop name
			// Check if the identifier is followed by ( — if so, parse as full expression
			if p.check(lexer.TOKEN_IDENT) && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_LPAREN {
				expr := p.parseExpression()
				props[propName] = expr
			} else {
				valTok := p.advance() // consume prop value (identifier)
				props[propName] = &ast.Identifier{Name: valTok.Value}
			}
		} else if p.checkNext(lexer.TOKEN_STRING) {
			p.advance() // consume prop name
			valTok := p.advance() // consume string value
			props[propName] = &ast.StringLiteral{Value: valTok.Value}
		} else if p.checkNext(lexer.TOKEN_NUMBER) {
			p.advance() // consume prop name
			valTok := p.advance() // consume number value
			var numVal float64
			fmt.Sscanf(valTok.Value, "%f", &numVal)
			props[propName] = &ast.NumberLiteral{Value: numVal}
		} else if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_YES {
			p.advance() // consume prop name
			p.advance() // consume "yes"/"true"
			props[propName] = &ast.BoolLiteral{Value: true}
		} else if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_NO {
			p.advance() // consume prop name
			p.advance() // consume "no"/"false"
			props[propName] = &ast.BoolLiteral{Value: false}
		} else if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_NOTHING {
			p.advance() // consume prop name
			p.advance() // consume "nothing"
			props[propName] = &ast.Identifier{Name: "null"}
		} else {
			// Check if this is a boolean prop (no value, e.g., showsVerticalScrollIndicator)
			nextType := lexer.TOKEN_EOF
			if p.pos+1 < len(p.tokens) {
				nextType = p.tokens[p.pos+1].Type
			}
			if nextType == lexer.TOKEN_COLON || nextType == lexer.TOKEN_NEWLINE || nextType == lexer.TOKEN_EOF || nextType == lexer.TOKEN_IDENT {
				p.advance() // consume prop name as boolean
				props[propName] = &ast.Identifier{Name: "true"}
				continue
			}
			break
		}
	}

	var children []ast.RenderNode

	if p.check(lexer.TOKEN_COLON) {
		p.advance() // consume ":"
		if p.check(lexer.TOKEN_STRING) {
			// Inline text: tag: "text"
			text := p.advance()
			children = append(children, ast.RenderNode{Text: &ast.StringLiteral{Value: text.Value}})
			p.consumeNewline()
		} else if p.check(lexer.TOKEN_IDENT) {
			// Inline expression: tag: expr
			expr := p.parseExpression()
			children = append(children, ast.RenderNode{Text: expr})
			p.consumeNewline()
		} else if p.check(lexer.TOKEN_NEWLINE) {
			// Block children
			p.advance() // consume newline
			p.expect(lexer.TOKEN_INDENT)
			for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
				p.skipNewlines()
				if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
					break
				}
				el := p.parseRenderElement()
				children = append(children, ast.RenderNode{Element: el})
			}
			if p.check(lexer.TOKEN_DEDENT) {
				p.advance()
			}
		}
	} else {
		p.consumeNewline()
	}

	return &ast.RenderElement{
		Tag:      tag.Value,
		Props:    props,
		Children: children,
		Line:     line,
	}
}

