package formatter

import (
	"fmt"
	"quill/ast"
	"strings"
)

// Formatter takes a Quill AST and produces clean, consistently formatted Quill source code.
type Formatter struct {
	indent     int
	indentStr  string
	output     strings.Builder
}

// New creates a new Formatter with the given indent string (default: 2 spaces).
func New() *Formatter {
	return &Formatter{
		indentStr: "  ",
	}
}

// NewWithIndent creates a Formatter with a custom indent string.
func NewWithIndent(indent string) *Formatter {
	return &Formatter{
		indentStr: indent,
	}
}

// Format formats a Quill program into clean source code.
func (f *Formatter) Format(program *ast.Program) string {
	f.output.Reset()

	prevWasBlock := false
	for i, stmt := range program.Statements {
		isBlock := isBlockStatement(stmt)

		// Add blank line before block statements (if, for, while, to, describe, test)
		// but not at the very beginning
		if i > 0 && (isBlock || prevWasBlock) {
			f.output.WriteString("\n")
		}

		f.formatStmt(stmt)
		f.output.WriteString("\n")
		prevWasBlock = isBlock
	}

	return strings.TrimRight(f.output.String(), "\n") + "\n"
}

func isBlockStatement(stmt ast.Statement) bool {
	switch stmt.(type) {
	case *ast.IfStatement, *ast.ForEachStatement, *ast.WhileStatement,
		*ast.FuncDefinition, *ast.DescribeStatement, *ast.TestBlock:
		return true
	}
	return false
}

func (f *Formatter) prefix() string {
	return strings.Repeat(f.indentStr, f.indent)
}

func (f *Formatter) formatStmt(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.AssignStatement:
		f.output.WriteString(fmt.Sprintf("%s%s is %s", f.prefix(), s.Name, f.formatExpr(s.Value)))

	case *ast.SayStatement:
		f.output.WriteString(fmt.Sprintf("%ssay %s", f.prefix(), f.formatExpr(s.Value)))

	case *ast.IfStatement:
		f.formatIf(s)

	case *ast.ForEachStatement:
		f.output.WriteString(fmt.Sprintf("%sfor each %s in %s:", f.prefix(), s.Variable, f.formatExpr(s.Iterable)))
		f.formatBlock(s.Body)

	case *ast.WhileStatement:
		f.output.WriteString(fmt.Sprintf("%swhile %s:", f.prefix(), f.formatExpr(s.Condition)))
		f.formatBlock(s.Body)

	case *ast.FuncDefinition:
		params := ""
		if len(s.Params) > 0 {
			params = " " + strings.Join(s.Params, " ")
		}
		f.output.WriteString(fmt.Sprintf("%sto %s%s:", f.prefix(), s.Name, params))
		f.formatBlock(s.Body)

	case *ast.ReturnStatement:
		f.output.WriteString(fmt.Sprintf("%sgive back %s", f.prefix(), f.formatExpr(s.Value)))

	case *ast.UseStatement:
		if s.Alias != "" {
			f.output.WriteString(fmt.Sprintf("%suse \"%s\" as %s", f.prefix(), s.Path, s.Alias))
		} else {
			f.output.WriteString(fmt.Sprintf("%suse \"%s\"", f.prefix(), s.Path))
		}

	case *ast.ExprStatement:
		f.output.WriteString(fmt.Sprintf("%s%s", f.prefix(), f.formatExpr(s.Expr)))

	case *ast.DotAssignStatement:
		f.output.WriteString(fmt.Sprintf("%s%s.%s is %s", f.prefix(), s.Object, s.Field, f.formatExpr(s.Value)))

	case *ast.DescribeStatement:
		f.output.WriteString(fmt.Sprintf("%sdescribe %s:", f.prefix(), s.Name))
		f.indent++
		for _, prop := range s.Properties {
			f.output.WriteString("\n")
			f.output.WriteString(fmt.Sprintf("%s%s is %s", f.prefix(), prop.Name, f.formatExpr(prop.Value)))
		}
		if len(s.Properties) > 0 && len(s.Methods) > 0 {
			f.output.WriteString("\n")
		}
		for _, method := range s.Methods {
			f.output.WriteString("\n")
			params := ""
			if len(method.Params) > 0 {
				params = " " + strings.Join(method.Params, " ")
			}
			f.output.WriteString(fmt.Sprintf("%sto %s%s:", f.prefix(), method.Name, params))
			f.formatBlock(method.Body)
		}
		f.indent--

	case *ast.TestBlock:
		f.output.WriteString(fmt.Sprintf("%stest \"%s\":", f.prefix(), s.Name))
		f.formatBlock(s.Body)

	case *ast.ExpectStatement:
		f.output.WriteString(fmt.Sprintf("%sexpect %s", f.prefix(), f.formatExpr(s.Expr)))

	default:
		f.output.WriteString(f.prefix() + "-- unknown statement")
	}
}

func (f *Formatter) formatIf(s *ast.IfStatement) {
	f.output.WriteString(fmt.Sprintf("%sif %s:", f.prefix(), f.formatExpr(s.Condition)))
	f.formatBlock(s.Body)

	for _, elif := range s.ElseIfs {
		f.output.WriteString("\n")
		f.output.WriteString(fmt.Sprintf("%sotherwise if %s:", f.prefix(), f.formatExpr(elif.Condition)))
		f.formatBlock(elif.Body)
	}

	if len(s.Else) > 0 {
		f.output.WriteString("\n")
		f.output.WriteString(fmt.Sprintf("%sotherwise:", f.prefix()))
		f.formatBlock(s.Else)
	}
}

func (f *Formatter) formatBlock(stmts []ast.Statement) {
	f.indent++
	for _, stmt := range stmts {
		f.output.WriteString("\n")
		f.formatStmt(stmt)
	}
	f.indent--
}

func (f *Formatter) formatExpr(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		return fmt.Sprintf("\"%s\"", e.Value)

	case *ast.NumberLiteral:
		if e.Value == float64(int64(e.Value)) {
			return fmt.Sprintf("%d", int64(e.Value))
		}
		return fmt.Sprintf("%g", e.Value)

	case *ast.BoolLiteral:
		if e.Value {
			return "yes"
		}
		return "no"

	case *ast.Identifier:
		return e.Name

	case *ast.ListLiteral:
		if len(e.Elements) == 0 {
			return "[]"
		}
		elems := make([]string, len(e.Elements))
		for i, el := range e.Elements {
			elems[i] = f.formatExpr(el)
		}
		return "[" + strings.Join(elems, ", ") + "]"

	case *ast.BinaryExpr:
		return fmt.Sprintf("%s %s %s", f.formatExpr(e.Left), e.Operator, f.formatExpr(e.Right))

	case *ast.ComparisonExpr:
		return f.formatComparison(e)

	case *ast.LogicalExpr:
		return fmt.Sprintf("%s %s %s", f.formatExpr(e.Left), e.Operator, f.formatExpr(e.Right))

	case *ast.NotExpr:
		return fmt.Sprintf("not %s", f.formatExpr(e.Operand))

	case *ast.UnaryMinusExpr:
		return fmt.Sprintf("-%s", f.formatExpr(e.Operand))

	case *ast.CallExpr:
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = f.formatExpr(a)
		}
		return fmt.Sprintf("%s(%s)", f.formatExpr(e.Function), strings.Join(args, ", "))

	case *ast.DotExpr:
		obj := f.formatExpr(e.Object)
		if obj == "this" {
			return "my." + e.Field
		}
		return obj + "." + e.Field

	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", f.formatExpr(e.Object), f.formatExpr(e.Index))

	case *ast.NewExpr:
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = f.formatExpr(a)
		}
		return fmt.Sprintf("new %s(%s)", e.ClassName, strings.Join(args, ", "))

	case *ast.AwaitExpr:
		return fmt.Sprintf("await %s", f.formatExpr(e.Expr))

	default:
		return "???"
	}
}

func (f *Formatter) formatComparison(e *ast.ComparisonExpr) string {
	left := f.formatExpr(e.Left)
	right := f.formatExpr(e.Right)

	switch e.Operator {
	case ">":
		return fmt.Sprintf("%s is greater than %s", left, right)
	case "<":
		return fmt.Sprintf("%s is less than %s", left, right)
	case ">=":
		return fmt.Sprintf("%s is greater than or equal to %s", left, right)
	case "<=":
		return fmt.Sprintf("%s is less than or equal to %s", left, right)
	case "==":
		return fmt.Sprintf("%s is %s", left, right)
	case "!=":
		return fmt.Sprintf("%s is not %s", left, right)
	case "contains":
		return fmt.Sprintf("%s contains %s", left, right)
	default:
		return fmt.Sprintf("%s %s %s", left, e.Operator, right)
	}
}
