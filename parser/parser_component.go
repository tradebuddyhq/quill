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
	var methods []ast.FuncDefinition
	var renderBody []ast.RenderElement
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
				renderBody = p.parseRenderMethod()
			} else if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TOKEN_LOAD {
				loader = p.parseLoadFunction()
			} else {
				method := p.parseFuncDef()
				methods = append(methods, *method)
			}
		} else if p.check(lexer.TOKEN_FORM) {
			action := p.parseFormAction()
			actions = append(actions, action)
		} else {
			p.error("expected state, style, head, method (to), form, effect, or render inside component block")
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.ComponentStatement{
		Name:         name.Value,
		HasProps:     hasProps,
		Props:        propNames,
		States:       states,
		Effects:      effects,
		Methods:      methods,
		RenderBody:   renderBody,
		Styles:       styles,
		NativeStyles: nativeStyles,
		Loader:       loader,
		Actions:      actions,
		Head:         head,
		Line:         line,
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
				// Read value - could be string or number
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

func (p *Parser) parseRenderMethod() []ast.RenderElement {
	p.advance() // consume "to"
	p.advance() // consume "render"
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	var elements []ast.RenderElement
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		elements = append(elements, *p.parseRenderElement())
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}
	return elements
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

	// Regular element: tag [props...] [: "text" | NEWLINE INDENT children DEDENT]
	tag := p.expect(lexer.TOKEN_IDENT)
	props := make(map[string]ast.Expression)

	// Parse props: onClick handler, bind:value ident, key value, etc.
	// Supports both space-separated (onClick increment) and = syntax (onClick=increment)
	for p.check(lexer.TOKEN_IDENT) && !p.isAtEnd() {
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
		// Check if next token is an identifier (event handler or attr with value)
		if p.checkNext(lexer.TOKEN_IDENT) {
			p.advance() // consume prop name
			valTok := p.advance() // consume prop value (identifier)
			props[propName] = &ast.Identifier{Name: valTok.Value}
		} else if p.checkNext(lexer.TOKEN_STRING) {
			p.advance() // consume prop name
			valTok := p.advance() // consume string value
			props[propName] = &ast.StringLiteral{Value: valTok.Value}
		} else {
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

