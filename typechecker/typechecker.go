package typechecker

import (
	"fmt"
	"quill/ast"
	"strings"
)

// Type represents a Quill type.
type Type struct {
	Name    string // "text", "number", "boolean", "list", "object", "nothing", "any", "function", or custom
	Inner   string // for generics: "list of number" -> Inner="number"
}

func (t Type) String() string {
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

// Checker performs type checking on Quill ASTs.
type Checker struct {
	diagnostics []TypeDiagnostic
	variables   map[string]Type   // variable name -> type
	functions   map[string]FuncSig // function name -> signature
	types       map[string]bool   // defined type names (from describe/define)
}

// FuncSig represents a function signature.
type FuncSig struct {
	Params     []Type
	ReturnType Type
}

// New creates a new type checker.
func New() *Checker {
	return &Checker{
		variables: make(map[string]Type),
		functions: make(map[string]FuncSig),
		types:     map[string]bool{
			"text": true, "number": true, "boolean": true,
			"list": true, "object": true, "nothing": true, "any": true,
		},
	}
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
				c.variables[param] = t
			}
			if s.ReturnType != "" {
				sig.ReturnType = parseType(s.ReturnType)
			} else {
				sig.ReturnType = Type{Name: "any"}
			}
			c.functions[s.Name] = sig

		case *ast.DescribeStatement:
			c.types[s.Name] = true

		case *ast.DefineStatement:
			c.types[s.Name] = true
			for _, v := range s.Variants {
				c.types[v.Name] = true
			}
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
		c.variables[s.Name] = inferredType

	case *ast.SayStatement:
		c.inferType(s.Value) // just check for errors in the expression

	case *ast.IfStatement:
		condType := c.inferType(s.Condition)
		if condType.Name != "boolean" && condType.Name != "any" {
			c.addWarning(s.Line, fmt.Sprintf("condition should be boolean, got %s", condType))
		}
		for _, stmt := range s.Body {
			c.checkStmt(stmt)
		}
		for _, elif := range s.ElseIfs {
			for _, stmt := range elif.Body {
				c.checkStmt(stmt)
			}
		}
		for _, stmt := range s.Else {
			c.checkStmt(stmt)
		}

	case *ast.ForEachStatement:
		iterType := c.inferType(s.Iterable)
		if iterType.Name == "list" && iterType.Inner != "" {
			c.variables[s.Variable] = Type{Name: iterType.Inner}
		} else {
			c.variables[s.Variable] = Type{Name: "any"}
		}
		for _, stmt := range s.Body {
			c.checkStmt(stmt)
		}

	case *ast.WhileStatement:
		c.inferType(s.Condition)
		for _, stmt := range s.Body {
			c.checkStmt(stmt)
		}

	case *ast.FuncDefinition:
		// Register param types
		for i, param := range s.Params {
			if i < len(s.ParamTypes) && s.ParamTypes[i] != "" {
				c.variables[param] = parseType(s.ParamTypes[i])
			} else {
				c.variables[param] = Type{Name: "any"}
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

	case *ast.ReturnStatement:
		c.inferType(s.Value)

	case *ast.ExprStatement:
		c.inferType(s.Expr)

	case *ast.TryCatchStatement:
		for _, stmt := range s.TryBody {
			c.checkStmt(stmt)
		}
		if s.ErrorVar != "" {
			c.variables[s.ErrorVar] = Type{Name: "text"}
		}
		for _, stmt := range s.CatchBody {
			c.checkStmt(stmt)
		}

	case *ast.MatchStatement:
		matchType := c.inferType(s.Value)
		for _, mc := range s.Cases {
			if mc.Pattern != nil {
				patternType := c.inferType(mc.Pattern)
				if !c.typeCompatible(matchType, patternType) && matchType.Name != "any" && patternType.Name != "any" {
					c.addWarning(s.Line, fmt.Sprintf("match value is %s but case pattern is %s", matchType, patternType))
				}
			}
			for _, stmt := range mc.Body {
				c.checkStmt(stmt)
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

	case *ast.ExpectStatement:
		c.inferType(s.Expr)

	case *ast.DotAssignStatement:
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
	}
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
		if t, ok := c.variables[e.Name]; ok {
			return t
		}
		return Type{Name: "any"}
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
	default:
		return Type{Name: "any"}
	}
}

// typeCompatible checks if actual is compatible with expected.
func (c *Checker) typeCompatible(expected, actual Type) bool {
	if expected.Name == "any" || actual.Name == "any" {
		return true
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

// parseType parses a type string like "number" or "list of number".
func parseType(s string) Type {
	s = strings.TrimSpace(s)
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
