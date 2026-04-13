package typechecker

import (
	"fmt"
	"quill/ast"
	"strings"
)

// Type represents a Quill type.
type Type struct {
	Name     string // "text", "number", "boolean", "list", "object", "nothing", "any", "function", or custom
	Inner    string // for generics: "list of number" -> Inner="number"
	Union    []Type // for union types: number | text
	Nullable bool   // for nullable: ?number
}

func (t Type) String() string {
	if len(t.Union) > 0 {
		parts := make([]string, len(t.Union))
		for i, u := range t.Union {
			parts[i] = u.String()
		}
		return strings.Join(parts, " | ")
	}
	if t.Nullable {
		if t.Inner != "" {
			return "?" + t.Name + " of " + t.Inner
		}
		return "?" + t.Name
	}
	if t.Inner != "" {
		return t.Name + " of " + t.Inner
	}
	return t.Name
}

// TypeDiagnostic represents a type error or warning.
type TypeDiagnostic struct {
	Line     int
	Severity string // "error", "warning"
	Message  string
}

func (d TypeDiagnostic) String() string {
	return fmt.Sprintf("  line %d [type %s]: %s", d.Line, d.Severity, d.Message)
}

// TraitInfo stores information about a registered trait.
type TraitInfo struct {
	Name    string
	Methods []TraitMethodSig
}

// TraitMethodSig represents a method signature in a trait.
type TraitMethodSig struct {
	Name       string
	ParamTypes []Type
	ReturnType Type
}

// EnumInfo stores information about a defined enum/algebraic type.
type EnumInfo struct {
	Name     string
	Variants []string
}

// ClassInfo stores information about a described class.
type ClassInfo struct {
	Name    string
	Methods map[string]FuncSig
}

// TypeAliasInfo stores information about a type alias.
type TypeAliasInfo struct {
	Name     string
	BaseType string
	Utility  string
	Args     []string
}

// Scope represents a lexical scope with its own variable bindings.
type Scope struct {
	variables map[string]Type
}

// Checker performs type checking on Quill ASTs.
type Checker struct {
	diagnostics []TypeDiagnostic
	scopes      []*Scope           // scope stack; last element is the current scope
	functions   map[string]FuncSig // function name -> signature
	types       map[string]bool    // defined type names (from describe/define)
	traits      map[string]TraitInfo  // trait name -> trait info
	enums       map[string]EnumInfo   // enum name -> enum info
	classes     map[string]ClassInfo  // class name -> class info
	narrowings  map[string]Type       // temporary type narrowings (for type narrowing in if blocks)
	typeAliases map[string]TypeAliasInfo // type alias name -> info
}

// FuncSig represents a function signature.
type FuncSig struct {
	Params     []Type
	ParamTypes []string // for generic params
	ReturnType Type
	TypeParams []ast.TypeParam
}

// New creates a new type checker.
func New() *Checker {
	c := &Checker{
		scopes: []*Scope{{variables: make(map[string]Type)}},
		functions: make(map[string]FuncSig),
		types: map[string]bool{
			"text": true, "number": true, "boolean": true,
			"list": true, "object": true, "nothing": true, "any": true,
		},
		traits:      make(map[string]TraitInfo),
		enums:       make(map[string]EnumInfo),
		classes:     make(map[string]ClassInfo),
		narrowings:  make(map[string]Type),
		typeAliases: make(map[string]TypeAliasInfo),
	}
	return c
}

// pushScope pushes a new scope onto the scope stack.
func (c *Checker) pushScope() {
	c.scopes = append(c.scopes, &Scope{variables: make(map[string]Type)})
}

// popScope removes the top scope from the scope stack.
func (c *Checker) popScope() {
	if len(c.scopes) > 1 {
		c.scopes = c.scopes[:len(c.scopes)-1]
	}
}

// defineVar defines a variable in the current (top) scope.
func (c *Checker) defineVar(name string, typ Type) {
	c.scopes[len(c.scopes)-1].variables[name] = typ
}

// lookupVar looks up a variable by walking the scope chain from top to bottom.
func (c *Checker) lookupVar(name string) (Type, bool) {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if t, ok := c.scopes[i].variables[name]; ok {
			return t, true
		}
	}
	return Type{}, false
}

// Check runs type checking and returns diagnostics.
func (c *Checker) Check(program *ast.Program) []TypeDiagnostic {
	c.diagnostics = nil

	// First pass: collect function signatures and type definitions
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.FuncDefinition:
			sig := FuncSig{}
			for i, param := range s.Params {
				t := Type{Name: "any"}
				if i < len(s.ParamTypes) && s.ParamTypes[i] != "" {
					t = parseType(s.ParamTypes[i])
				}
				sig.Params = append(sig.Params, t)
				c.defineVar(param, t)
			}
			if s.ReturnType != "" {
				sig.ReturnType = parseType(s.ReturnType)
			} else {
				sig.ReturnType = Type{Name: "any"}
			}
			c.functions[s.Name] = sig

		case *ast.DescribeStatement:
			c.types[s.Name] = true
			// Register class methods for trait checking
			classInfo := ClassInfo{Name: s.Name, Methods: make(map[string]FuncSig)}
			for _, method := range s.Methods {
				sig := FuncSig{}
				for i, _ := range method.Params {
					t := Type{Name: "any"}
					if i < len(method.ParamTypes) && method.ParamTypes[i] != "" {
						t = parseType(method.ParamTypes[i])
					}
					sig.Params = append(sig.Params, t)
				}
				if method.ReturnType != "" {
					sig.ReturnType = parseType(method.ReturnType)
				} else {
					sig.ReturnType = Type{Name: "any"}
				}
				classInfo.Methods[method.Name] = sig
			}
			c.classes[s.Name] = classInfo

		case *ast.DefineStatement:
			c.types[s.Name] = true
			variants := []string{}
			for _, v := range s.Variants {
				c.types[v.Name] = true
				variants = append(variants, v.Name)
			}
			c.enums[s.Name] = EnumInfo{Name: s.Name, Variants: variants}

		case *ast.TraitDeclaration:
			c.types[s.Name] = true
			traitInfo := TraitInfo{Name: s.Name}
			for _, m := range s.Methods {
				methodSig := TraitMethodSig{Name: m.Name}
				for _, p := range m.Params {
					if p.TypeHint != "" {
						methodSig.ParamTypes = append(methodSig.ParamTypes, parseType(p.TypeHint))
					} else {
						methodSig.ParamTypes = append(methodSig.ParamTypes, Type{Name: "any"})
					}
				}
				if m.ReturnType != "" {
					methodSig.ReturnType = parseType(m.ReturnType)
				} else {
					methodSig.ReturnType = Type{Name: "any"}
				}
				traitInfo.Methods = append(traitInfo.Methods, methodSig)
			}
			c.traits[s.Name] = traitInfo

		case *ast.TypeAliasStatement:
			c.types[s.Name] = true
			c.typeAliases[s.Name] = TypeAliasInfo{
				Name:     s.Name,
				BaseType: s.BaseType,
				Utility:  s.Utility,
				Args:     s.Args,
			}
			// Validate that base type is known
			if !c.types[s.BaseType] && s.Utility != "Record" {
				c.addWarning(s.Line, fmt.Sprintf("type alias %q references unknown base type %q", s.Name, s.BaseType))
			}
			// For Omit/Pick, validate field names could be checked if we had class info
			// For Record, register as a known type without base type check
		}
	}

	// Second pass: type check statements
	for _, stmt := range program.Statements {
		c.checkStmt(stmt)
	}

	return c.diagnostics
}

func (c *Checker) addError(line int, msg string) {
	c.diagnostics = append(c.diagnostics, TypeDiagnostic{Line: line, Severity: "error", Message: msg})
}

func (c *Checker) addWarning(line int, msg string) {
	c.diagnostics = append(c.diagnostics, TypeDiagnostic{Line: line, Severity: "warning", Message: msg})
}

func (c *Checker) checkStmt(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.AssignStatement:
		inferredType := c.inferType(s.Value)
		c.defineVar(s.Name, inferredType)

	case *ast.SayStatement:
		c.inferType(s.Value) // just check for errors in the expression

	case *ast.IfStatement:
		condType := c.inferType(s.Condition)
		if condType.Name != "boolean" && condType.Name != "any" {
			c.addWarning(s.Line, fmt.Sprintf("condition should be boolean, got %s", condType))
		}

		// Type narrowing: if the condition is a type check, narrow the variable
		if tc, ok := s.Condition.(*ast.TypeCheckExpr); ok {
			if ident, ok := tc.Expr.(*ast.Identifier); ok {
				// Save old narrowing and apply new one
				oldNarrowing, hadOldNarrowing := c.narrowings[ident.Name]
				c.narrowings[ident.Name] = Type{Name: tc.TypeName}
				c.pushScope()
				for _, stmt := range s.Body {
					c.checkStmt(stmt)
				}
				c.popScope()
				// Restore narrowing
				if hadOldNarrowing {
					c.narrowings[ident.Name] = oldNarrowing
				} else {
					delete(c.narrowings, ident.Name)
				}
				// Check else-ifs and else without narrowing
				for _, elif := range s.ElseIfs {
					c.pushScope()
					for _, stmt := range elif.Body {
						c.checkStmt(stmt)
					}
					c.popScope()
				}
				if len(s.Else) > 0 {
					c.pushScope()
					for _, stmt := range s.Else {
						c.checkStmt(stmt)
					}
					c.popScope()
				}
				break
			}
		}

		c.pushScope()
		for _, stmt := range s.Body {
			c.checkStmt(stmt)
		}
		c.popScope()
		for _, elif := range s.ElseIfs {
			c.pushScope()
			for _, stmt := range elif.Body {
				c.checkStmt(stmt)
			}
			c.popScope()
		}
		if len(s.Else) > 0 {
			c.pushScope()
			for _, stmt := range s.Else {
				c.checkStmt(stmt)
			}
			c.popScope()
		}

	case *ast.ForEachStatement:
		iterType := c.inferType(s.Iterable)
		c.pushScope()
		if iterType.Name == "list" && iterType.Inner != "" {
			c.defineVar(s.Variable, Type{Name: iterType.Inner})
		} else {
			c.defineVar(s.Variable, Type{Name: "any"})
		}
		for _, stmt := range s.Body {
			c.checkStmt(stmt)
		}
		c.popScope()

	case *ast.WhileStatement:
		c.inferType(s.Condition)
		c.pushScope()
		for _, stmt := range s.Body {
			c.checkStmt(stmt)
		}
		c.popScope()

	case *ast.FuncDefinition:
		c.pushScope()
		// Register param types
		for i, param := range s.Params {
			if i < len(s.ParamTypes) && s.ParamTypes[i] != "" {
				c.defineVar(param, parseType(s.ParamTypes[i]))
			} else {
				c.defineVar(param, Type{Name: "any"})
			}
		}
		// Check body
		for _, stmt := range s.Body {
			c.checkStmt(stmt)
		}
		// Check return type
		if s.ReturnType != "" {
			expectedReturn := parseType(s.ReturnType)
			for _, stmt := range s.Body {
				if ret, ok := stmt.(*ast.ReturnStatement); ok {
					actualType := c.inferType(ret.Value)
					if !c.typeCompatible(expectedReturn, actualType) {
						c.addError(ret.Line, fmt.Sprintf("function %q should return %s, but returns %s",
							s.Name, expectedReturn, actualType))
					}
				}
			}
		}
		c.popScope()

	case *ast.ReturnStatement:
		c.inferType(s.Value)

	case *ast.ExprStatement:
		c.inferType(s.Expr)

	case *ast.TryCatchStatement:
		c.pushScope()
		for _, stmt := range s.TryBody {
			c.checkStmt(stmt)
		}
		c.popScope()
		c.pushScope()
		if s.ErrorVar != "" {
			c.defineVar(s.ErrorVar, Type{Name: "text"})
		}
		for _, stmt := range s.CatchBody {
			c.checkStmt(stmt)
		}
		c.popScope()

	case *ast.MatchStatement:
		matchType := c.inferType(s.Value)
		hasOtherwise := false
		matchedVariants := make(map[string]bool)

		for _, mc := range s.Cases {
			if mc.Pattern == nil {
				hasOtherwise = true
			} else {
				patternType := c.inferType(mc.Pattern)
				if !c.typeCompatible(matchType, patternType) && matchType.Name != "any" && patternType.Name != "any" {
					c.addWarning(s.Line, fmt.Sprintf("match value is %s but case pattern is %s", matchType, patternType))
				}
				// Track matched variant names for exhaustiveness
				if ident, ok := mc.Pattern.(*ast.Identifier); ok {
					matchedVariants[ident.Name] = true
				}
				// Also check DotExpr like Shape.Circle
				if dot, ok := mc.Pattern.(*ast.DotExpr); ok {
					if ident, ok := dot.Object.(*ast.Identifier); ok {
						matchedVariants[ident.Name+"."+dot.Field] = true
						matchedVariants[dot.Field] = true
					}
				}
			}
			for _, stmt := range mc.Body {
				c.checkStmt(stmt)
			}
		}

		// Exhaustive match checking: if the match value is a known enum type
		if !hasOtherwise {
			// Try to find the enum type being matched
			for enumName, enumInfo := range c.enums {
				// Check if the match value is of this enum type
				if matchType.Name == enumName || c.isEnumVariantMatch(s.Value, enumName) {
					for _, variant := range enumInfo.Variants {
						if !matchedVariants[variant] && !matchedVariants[enumName+"."+variant] {
							c.addError(s.Line, fmt.Sprintf("Non-exhaustive match: missing variant '%s'", variant))
						}
					}
					break
				}
			}
		}

	case *ast.DescribeStatement:
		for _, method := range s.Methods {
			for _, stmt := range method.Body {
				c.checkStmt(stmt)
			}
		}

	case *ast.TestBlock:
		for _, stmt := range s.Body {
			c.checkStmt(stmt)
		}

	case *ast.MockStatement:
		for _, stmt := range s.Body {
			c.checkStmt(stmt)
		}

	case *ast.ExpectStatement:
		c.inferType(s.Expr)

	case *ast.DotAssignStatement:
		c.inferType(s.Value)

	case *ast.TypedAssignStatement:
		inferredType := c.inferType(s.Value)
		if s.TypeHint != "" {
			expected := parseType(s.TypeHint)
			if !c.typeCompatible(expected, inferredType) {
				c.addError(s.Line, fmt.Sprintf("variable %q declared as %s but assigned %s",
					s.Name, expected, inferredType))
			}
			c.defineVar(s.Name, expected)
		} else {
			c.defineVar(s.Name, inferredType)
		}

	case *ast.TraitDeclaration:
		// Already registered in first pass

	case *ast.DestructureStatement:
		c.inferType(s.Value)

	case *ast.UseStatement:
		// nothing to type check
	case *ast.FromUseStatement:
		// nothing to type check
	case *ast.DefineStatement:
		// nothing to type check
	case *ast.BreakStatement:
		// nothing to type check
	case *ast.ContinueStatement:
		// nothing to type check
	case *ast.TypeAliasStatement:
		// Already registered in first pass
	case *ast.DecoratedFuncDefinition:
		// Check the inner function
		c.checkStmt(s.Func)
	case *ast.DecoratedRouteDefinition:
		for _, stmt := range s.Route.Body {
			c.checkStmt(stmt)
		}
	case *ast.WebSocketBlock:
		for _, stmt := range s.OnConnect {
			c.checkStmt(stmt)
		}
		for _, stmt := range s.OnMessage {
			c.checkStmt(stmt)
		}
		for _, stmt := range s.OnClose {
			c.checkStmt(stmt)
		}
	case *ast.BroadcastStatement:
		c.inferType(s.Value)
	}
}

// isEnumVariantMatch checks if an expression refers to a variant of the given enum.
func (c *Checker) isEnumVariantMatch(expr ast.Expression, enumName string) bool {
	if ident, ok := expr.(*ast.Identifier); ok {
		if enumInfo, ok := c.enums[enumName]; ok {
			for _, v := range enumInfo.Variants {
				if ident.Name == v {
					return true
				}
			}
		}
	}
	// Check for enum.Variant dot expression
	if dot, ok := expr.(*ast.DotExpr); ok {
		if obj, ok := dot.Object.(*ast.Identifier); ok {
			return obj.Name == enumName
		}
	}
	// Check if the variable type matches the enum name
	if ident, ok := expr.(*ast.Identifier); ok {
		if t, ok := c.lookupVar(ident.Name); ok {
			return t.Name == enumName
		}
	}
	return false
}

// inferType infers the type of an expression.
func (c *Checker) inferType(expr ast.Expression) Type {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		return Type{Name: "text"}
	case *ast.NumberLiteral:
		return Type{Name: "number"}
	case *ast.BoolLiteral:
		return Type{Name: "boolean"}
	case *ast.NothingLiteral:
		return Type{Name: "nothing"}
	case *ast.Identifier:
		// Check narrowings first (from type narrowing in if blocks)
		if t, ok := c.narrowings[e.Name]; ok {
			return t
		}
		if t, ok := c.lookupVar(e.Name); ok {
			return t
		}
		return Type{Name: "any"}
	case *ast.TypeCheckExpr:
		return Type{Name: "boolean"}
	case *ast.ListLiteral:
		if len(e.Elements) > 0 {
			firstType := c.inferType(e.Elements[0])
			// Check all elements have same type
			for i := 1; i < len(e.Elements); i++ {
				elemType := c.inferType(e.Elements[i])
				if elemType.Name != firstType.Name && firstType.Name != "any" && elemType.Name != "any" {
					// Mixed list
					return Type{Name: "list", Inner: "any"}
				}
			}
			return Type{Name: "list", Inner: firstType.Name}
		}
		return Type{Name: "list"}
	case *ast.ObjectLiteral:
		return Type{Name: "object"}
	case *ast.BinaryExpr:
		leftType := c.inferType(e.Left)
		rightType := c.inferType(e.Right)
		if e.Operator == "+" {
			if leftType.Name == "text" || rightType.Name == "text" {
				return Type{Name: "text"}
			}
			if leftType.Name == "number" && rightType.Name == "number" {
				return Type{Name: "number"}
			}
			if leftType.Name == "number" && rightType.Name != "number" && rightType.Name != "any" {
				c.addWarning(0, fmt.Sprintf("adding number to %s may produce unexpected results", rightType))
			}
		}
		if e.Operator == "-" || e.Operator == "*" || e.Operator == "/" || e.Operator == "%" {
			if leftType.Name != "number" && leftType.Name != "any" {
				c.addWarning(0, fmt.Sprintf("operator %q expects numbers, got %s on the left", e.Operator, leftType))
			}
			if rightType.Name != "number" && rightType.Name != "any" {
				c.addWarning(0, fmt.Sprintf("operator %q expects numbers, got %s on the right", e.Operator, rightType))
			}
			return Type{Name: "number"}
		}
		return Type{Name: "any"}
	case *ast.ComparisonExpr:
		return Type{Name: "boolean"}
	case *ast.LogicalExpr:
		return Type{Name: "boolean"}
	case *ast.NotExpr:
		return Type{Name: "boolean"}
	case *ast.UnaryMinusExpr:
		return Type{Name: "number"}
	case *ast.CallExpr:
		if ident, ok := e.Function.(*ast.Identifier); ok {
			if sig, ok := c.functions[ident.Name]; ok {
				// Check argument types
				for i, arg := range e.Args {
					if i < len(sig.Params) {
						argType := c.inferType(arg)
						expected := sig.Params[i]
						if !c.typeCompatible(expected, argType) {
							c.addError(0, fmt.Sprintf("argument %d of %s() expects %s, got %s",
								i+1, ident.Name, expected, argType))
						}
					}
				}
				return sig.ReturnType
			}
			// Known stdlib return types
			return c.stdlibReturnType(ident.Name)
		}
		return Type{Name: "any"}
	case *ast.DotExpr:
		// Check if this is an enum variant access like Shape.Circle
		if ident, ok := e.Object.(*ast.Identifier); ok {
			if _, ok := c.enums[ident.Name]; ok {
				return Type{Name: ident.Name}
			}
		}
		return Type{Name: "any"}
	case *ast.IndexExpr:
		objType := c.inferType(e.Object)
		if objType.Inner != "" {
			return Type{Name: objType.Inner}
		}
		return Type{Name: "any"}
	case *ast.NewExpr:
		return Type{Name: e.ClassName}
	case *ast.AwaitExpr:
		return c.inferType(e.Expr)
	case *ast.LambdaExpr:
		return Type{Name: "function"}
	case *ast.PipeExpr:
		return Type{Name: "any"}
	case *ast.SpreadExpr:
		return c.inferType(e.Expr)
	case *ast.MockAssertionExpr:
		return Type{Name: "boolean"}
	default:
		return Type{Name: "any"}
	}
}

// typeCompatible checks if actual is compatible with expected.
func (c *Checker) typeCompatible(expected, actual Type) bool {
	if expected.Name == "any" || actual.Name == "any" {
		return true
	}

	// Union type on expected side: actual must match at least one member
	if len(expected.Union) > 0 {
		for _, u := range expected.Union {
			if c.typeCompatible(u, actual) {
				return true
			}
		}
		return false
	}

	// Union type on actual side: all members must be compatible with expected
	if len(actual.Union) > 0 {
		for _, u := range actual.Union {
			if !c.typeCompatible(expected, u) {
				return false
			}
		}
		return true
	}

	// Nullable expected: compatible with the base type or nothing
	if expected.Nullable {
		base := Type{Name: expected.Name, Inner: expected.Inner}
		if actual.Name == "nothing" {
			return true
		}
		return c.typeCompatible(base, actual)
	}

	// Nullable actual: compatible only if expected is also nullable or accepts nothing
	if actual.Nullable {
		base := Type{Name: actual.Name, Inner: actual.Inner}
		if expected.Nullable {
			expectedBase := Type{Name: expected.Name, Inner: expected.Inner}
			return c.typeCompatible(expectedBase, base)
		}
		// Non-nullable expected: the base must match and nothing must also be acceptable
		return false
	}

	if expected.Name == actual.Name {
		if expected.Inner == "" || actual.Inner == "" {
			return true
		}
		return expected.Inner == actual.Inner || expected.Inner == "any" || actual.Inner == "any"
	}
	// nothing is compatible with any type (like null)
	if actual.Name == "nothing" {
		return true
	}
	return false
}

// stdlibReturnType returns known return types for stdlib functions.
func (c *Checker) stdlibReturnType(name string) Type {
	switch name {
	case "length", "toNumber", "round", "floor", "ceil", "abs", "randomInt", "indexOf", "timestamp", "diffDays":
		return Type{Name: "number"}
	case "toText", "join", "trim", "upper", "lower", "replace_text", "now", "today", "uuid", "hash",
		"capitalize", "encodeBase64", "decodeBase64", "encodeURL", "decodeURL", "formatDate", "env",
		"read", "truncate":
		return Type{Name: "text"}
	case "split", "keys", "values", "range", "sort", "reverse", "unique", "filter", "map_list",
		"flat", "concat", "slice", "push", "words", "lines", "matches", "entries", "args":
		return Type{Name: "list"}
	case "includes", "startsWith", "endsWith", "exists", "fileExists", "isText", "isNumber",
		"isList", "isObject", "isNothing", "isFunction", "hasKey", "some", "every",
		"matchesPattern":
		return Type{Name: "boolean"}
	case "random":
		return Type{Name: "number"}
	case "typeOf", "platform":
		return Type{Name: "text"}
	case "parseJSON", "readJSON", "merge", "pick", "omit", "fromEntries", "deepCopy", "fileInfo", "memory":
		return Type{Name: "object"}
	default:
		return Type{Name: "any"}
	}
}

// parseType parses a type string like "number", "list of number", "number | text", or "?number".
func parseType(s string) Type {
	s = strings.TrimSpace(s)

	// Union type: "number | text"
	if strings.Contains(s, " | ") {
		parts := strings.Split(s, " | ")
		union := make([]Type, len(parts))
		for i, p := range parts {
			union[i] = parseType(strings.TrimSpace(p))
		}
		return Type{Union: union}
	}

	// Nullable type: "?number" is shorthand for "number | nothing"
	if strings.HasPrefix(s, "?") {
		inner := parseType(s[1:])
		inner.Nullable = true
		return inner
	}

	if strings.Contains(s, " of ") {
		parts := strings.SplitN(s, " of ", 2)
		return Type{Name: strings.TrimSpace(parts[0]), Inner: strings.TrimSpace(parts[1])}
	}
	return Type{Name: s}
}

// HasErrors returns true if there are type errors.
func HasErrors(diagnostics []TypeDiagnostic) bool {
	for _, d := range diagnostics {
		if d.Severity == "error" {
			return true
		}
	}
	return false
}
