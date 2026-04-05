package analyzer

import (
	"fmt"
	"quill/ast"
	"strings"
)

// Severity represents the severity of a diagnostic.
type Severity int

const (
	Warning Severity = iota
	Error
	Info
)

func (s Severity) String() string {
	switch s {
	case Warning:
		return "warning"
	case Error:
		return "error"
	case Info:
		return "info"
	}
	return "unknown"
}

// Diagnostic represents a single issue found by the analyzer.
type Diagnostic struct {
	Line     int
	Severity Severity
	Message  string
	Hint     string
}

func (d Diagnostic) String() string {
	prefix := d.Severity.String()
	s := fmt.Sprintf("  line %d [%s]: %s", d.Line, prefix, d.Message)
	if d.Hint != "" {
		s += fmt.Sprintf("\n    hint: %s", d.Hint)
	}
	return s
}

// Analyzer performs static analysis on Quill ASTs.
type Analyzer struct {
	diagnostics []Diagnostic
	defined     map[string]int    // variable name -> line defined
	used        map[string]bool   // variable name -> was it used
	functions   map[string]int    // function name -> param count
	inFunction  bool
	inTest      bool
	inDescribe  bool
}

// New creates a new Analyzer.
func New() *Analyzer {
	return &Analyzer{
		defined:   make(map[string]int),
		used:      make(map[string]bool),
		functions: make(map[string]int),
	}
}

// Analyze runs all checks on a program and returns diagnostics.
func (a *Analyzer) Analyze(program *ast.Program) []Diagnostic {
	a.diagnostics = nil

	// First pass: collect function definitions
	for _, stmt := range program.Statements {
		if fn, ok := stmt.(*ast.FuncDefinition); ok {
			a.functions[fn.Name] = len(fn.Params)
		}
	}

	// Second pass: analyze statements
	for _, stmt := range program.Statements {
		a.analyzeStmt(stmt)
	}

	// Check for unused variables
	for name, line := range a.defined {
		if !a.used[name] {
			// Don't warn about variables starting with _
			if !strings.HasPrefix(name, "_") {
				a.addDiagnostic(line, Warning, fmt.Sprintf("variable %q is defined but never used", name),
					"prefix with _ to suppress this warning")
			}
		}
	}

	return a.diagnostics
}

func (a *Analyzer) addDiagnostic(line int, severity Severity, message, hint string) {
	a.diagnostics = append(a.diagnostics, Diagnostic{
		Line:     line,
		Severity: severity,
		Message:  message,
		Hint:     hint,
	})
}

func (a *Analyzer) analyzeStmt(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.AssignStatement:
		a.analyzeExpr(s.Value, s.Line)
		if _, exists := a.defined[s.Name]; exists {
			// Reassignment is fine, just note it
		}
		a.defined[s.Name] = s.Line

	case *ast.SayStatement:
		a.analyzeExpr(s.Value, s.Line)

	case *ast.IfStatement:
		a.analyzeExpr(s.Condition, s.Line)
		a.checkAlwaysTrueCondition(s.Condition, s.Line)
		for _, stmt := range s.Body {
			a.analyzeStmt(stmt)
		}
		for _, elif := range s.ElseIfs {
			a.analyzeExpr(elif.Condition, s.Line)
			for _, stmt := range elif.Body {
				a.analyzeStmt(stmt)
			}
		}
		for _, stmt := range s.Else {
			a.analyzeStmt(stmt)
		}

	case *ast.ForEachStatement:
		a.analyzeExpr(s.Iterable, s.Line)
		a.defined[s.Variable] = s.Line
		for _, stmt := range s.Body {
			a.analyzeStmt(stmt)
		}

	case *ast.WhileStatement:
		a.analyzeExpr(s.Condition, s.Line)
		a.checkInfiniteLoop(s)
		for _, stmt := range s.Body {
			a.analyzeStmt(stmt)
		}

	case *ast.FuncDefinition:
		a.inFunction = true
		// Register params as defined
		for _, param := range s.Params {
			a.defined[param] = s.Line
			a.used[param] = true // don't warn about unused params
		}
		hasReturn := false
		for _, stmt := range s.Body {
			a.analyzeStmt(stmt)
			if _, ok := stmt.(*ast.ReturnStatement); ok {
				hasReturn = true
			}
		}
		// Check if function has no return (might be intentional for side-effect functions)
		if !hasReturn && len(s.Body) > 0 {
			// Check if last statement is an expression (potential missing return)
			if _, ok := s.Body[len(s.Body)-1].(*ast.ExprStatement); ok {
				a.addDiagnostic(s.Line, Info, fmt.Sprintf("function %q doesn't return a value", s.Name),
					"use 'give back' to return a value")
			}
		}
		a.inFunction = false

	case *ast.ReturnStatement:
		if !a.inFunction {
			a.addDiagnostic(s.Line, Error, "'give back' used outside of a function", "")
		}
		a.analyzeExpr(s.Value, s.Line)

	case *ast.ExprStatement:
		a.analyzeExpr(s.Expr, s.Line)

	case *ast.UseStatement:
		if s.Path == "" {
			a.addDiagnostic(s.Line, Error, "empty import path", "")
		}

	case *ast.TestBlock:
		a.inTest = true
		if s.Name == "" {
			a.addDiagnostic(s.Line, Warning, "test has no description", "add a descriptive name for your test")
		}
		hasExpect := false
		for _, stmt := range s.Body {
			a.analyzeStmt(stmt)
			if _, ok := stmt.(*ast.ExpectStatement); ok {
				hasExpect = true
			}
		}
		if !hasExpect {
			a.addDiagnostic(s.Line, Warning, "test has no expectations", "add 'expect' statements to verify behavior")
		}
		a.inTest = false

	case *ast.MockStatement:
		if !a.inTest {
			a.addDiagnostic(s.Line, Warning, "'mock' used outside of a test block", "wrap in a 'test' block")
		}
		for _, stmt := range s.Body {
			a.analyzeStmt(stmt)
		}

	case *ast.ExpectStatement:
		if !a.inTest {
			a.addDiagnostic(s.Line, Warning, "'expect' used outside of a test block", "wrap in a 'test' block")
		}
		a.analyzeExpr(s.Expr, s.Line)

	case *ast.DescribeStatement:
		a.inDescribe = true
		if len(s.Name) > 0 && s.Name[0] >= 'a' && s.Name[0] <= 'z' {
			a.addDiagnostic(s.Line, Info, fmt.Sprintf("class name %q should start with uppercase", s.Name),
				fmt.Sprintf("rename to %q", strings.Title(s.Name)))
		}
		// Check for duplicate property names
		propNames := make(map[string]bool)
		for _, prop := range s.Properties {
			if propNames[prop.Name] {
				a.addDiagnostic(prop.Line, Warning, fmt.Sprintf("duplicate property %q in class %s", prop.Name, s.Name), "")
			}
			propNames[prop.Name] = true
		}
		// Check for duplicate method names
		methodNames := make(map[string]bool)
		for _, method := range s.Methods {
			if methodNames[method.Name] {
				a.addDiagnostic(method.Line, Warning, fmt.Sprintf("duplicate method %q in class %s", method.Name, s.Name), "")
			}
			methodNames[method.Name] = true
			for _, stmt := range method.Body {
				a.analyzeStmt(stmt)
			}
		}
		a.inDescribe = false

	case *ast.DotAssignStatement:
		a.analyzeExpr(s.Value, s.Line)

	case *ast.TryCatchStatement:
		for _, stmt := range s.TryBody {
			a.analyzeStmt(stmt)
		}
		if s.ErrorVar != "" {
			a.defined[s.ErrorVar] = s.Line
			a.used[s.ErrorVar] = true
		}
		for _, stmt := range s.CatchBody {
			a.analyzeStmt(stmt)
		}

	case *ast.BreakStatement:
		// Could check if we're inside a loop

	case *ast.ContinueStatement:
		// Could check if we're inside a loop

	case *ast.FromUseStatement:
		if s.Path == "" {
			a.addDiagnostic(s.Line, Error, "empty import path", "")
		}
		for _, name := range s.Names {
			a.defined[name] = s.Line
			a.used[name] = true
		}

	case *ast.MatchStatement:
		a.analyzeExpr(s.Value, s.Line)
		for _, mc := range s.Cases {
			if mc.Pattern != nil {
				a.analyzeExpr(mc.Pattern, s.Line)
			}
			if mc.Guard != nil {
				a.analyzeExpr(mc.Guard, s.Line)
			}
			for _, stmt := range mc.Body {
				a.analyzeStmt(stmt)
			}
		}

	case *ast.DefineStatement:
		// Register enum type and variant names
		a.defined[s.Name] = s.Line
		a.used[s.Name] = true
		for _, v := range s.Variants {
			a.defined[v.Name] = s.Line
			a.used[v.Name] = true
		}
	}
}

func (a *Analyzer) analyzeExpr(expr ast.Expression, line int) {
	switch e := expr.(type) {
	case *ast.Identifier:
		a.used[e.Name] = true

	case *ast.BinaryExpr:
		a.analyzeExpr(e.Left, line)
		a.analyzeExpr(e.Right, line)
		// Check for division by zero
		if e.Operator == "/" {
			if num, ok := e.Right.(*ast.NumberLiteral); ok && num.Value == 0 {
				a.addDiagnostic(line, Warning, "division by zero", "")
			}
		}

	case *ast.ComparisonExpr:
		a.analyzeExpr(e.Left, line)
		a.analyzeExpr(e.Right, line)
		// Check for self-comparison
		if leftId, ok := e.Left.(*ast.Identifier); ok {
			if rightId, ok := e.Right.(*ast.Identifier); ok {
				if leftId.Name == rightId.Name {
					a.addDiagnostic(line, Warning, fmt.Sprintf("comparing %q to itself", leftId.Name),
						"this is always true or always false")
				}
			}
		}

	case *ast.LogicalExpr:
		a.analyzeExpr(e.Left, line)
		a.analyzeExpr(e.Right, line)

	case *ast.NotExpr:
		a.analyzeExpr(e.Operand, line)

	case *ast.UnaryMinusExpr:
		a.analyzeExpr(e.Operand, line)

	case *ast.CallExpr:
		a.analyzeExpr(e.Function, line)
		for _, arg := range e.Args {
			a.analyzeExpr(arg, line)
		}
		// Check function call argument count
		if ident, ok := e.Function.(*ast.Identifier); ok {
			if paramCount, defined := a.functions[ident.Name]; defined {
				if len(e.Args) != paramCount {
					a.addDiagnostic(line, Error,
						fmt.Sprintf("%s() expects %d arguments but got %d", ident.Name, paramCount, len(e.Args)),
						"")
				}
			}
		}

	case *ast.DotExpr:
		a.analyzeExpr(e.Object, line)

	case *ast.IndexExpr:
		a.analyzeExpr(e.Object, line)
		a.analyzeExpr(e.Index, line)
		// Check for negative index (handles both -1 literal and UnaryMinus(1))
		if num, ok := e.Index.(*ast.NumberLiteral); ok && num.Value < 0 {
			a.addDiagnostic(line, Warning, "negative array index", "JavaScript doesn't support negative indexing like Python")
		}
		if unary, ok := e.Index.(*ast.UnaryMinusExpr); ok {
			if _, ok := unary.Operand.(*ast.NumberLiteral); ok {
				a.addDiagnostic(line, Warning, "negative array index", "JavaScript doesn't support negative indexing like Python")
			}
		}

	case *ast.ListLiteral:
		for _, el := range e.Elements {
			a.analyzeExpr(el, line)
		}

	case *ast.NewExpr:
		for _, arg := range e.Args {
			a.analyzeExpr(arg, line)
		}

	case *ast.AwaitExpr:
		a.analyzeExpr(e.Expr, line)

	case *ast.ObjectLiteral:
		for _, val := range e.Values {
			a.analyzeExpr(val, line)
		}

	case *ast.LambdaExpr:
		a.analyzeExpr(e.Body, line)

	case *ast.SpreadExpr:
		a.analyzeExpr(e.Expr, line)

	case *ast.NothingLiteral:
		// nothing to analyze

	case *ast.PipeExpr:
		a.analyzeExpr(e.Left, line)
		a.analyzeExpr(e.Right, line)
	}
}

func (a *Analyzer) checkAlwaysTrueCondition(expr ast.Expression, line int) {
	switch e := expr.(type) {
	case *ast.BoolLiteral:
		if e.Value {
			a.addDiagnostic(line, Warning, "condition is always true", "")
		} else {
			a.addDiagnostic(line, Warning, "condition is always false — this code will never run", "")
		}
	}
}

func (a *Analyzer) checkInfiniteLoop(s *ast.WhileStatement) {
	// Check if condition is a literal true
	if b, ok := s.Condition.(*ast.BoolLiteral); ok && b.Value {
		// Check if body contains a break
		hasBreak := false
		for _, stmt := range s.Body {
			if _, ok := stmt.(*ast.BreakStatement); ok {
				hasBreak = true
				break
			}
			// Check inside if blocks
			if ifStmt, ok := stmt.(*ast.IfStatement); ok {
				for _, bodyStmt := range ifStmt.Body {
					if _, ok := bodyStmt.(*ast.BreakStatement); ok {
						hasBreak = true
						break
					}
				}
			}
		}
		if !hasBreak {
			a.addDiagnostic(s.Line, Warning, "infinite loop — condition is always true", "make sure you have a way to exit")
		}
	}
}

// HasErrors returns true if any diagnostics are errors.
func HasErrors(diagnostics []Diagnostic) bool {
	for _, d := range diagnostics {
		if d.Severity == Error {
			return true
		}
	}
	return false
}

// HasWarnings returns true if any diagnostics are warnings.
func HasWarnings(diagnostics []Diagnostic) bool {
	for _, d := range diagnostics {
		if d.Severity == Warning {
			return true
		}
	}
	return false
}

// FilterBySeverity returns diagnostics of a specific severity.
func FilterBySeverity(diagnostics []Diagnostic, severity Severity) []Diagnostic {
	var result []Diagnostic
	for _, d := range diagnostics {
		if d.Severity == severity {
			result = append(result, d)
		}
	}
	return result
}
