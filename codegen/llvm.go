package codegen

import (
	"fmt"
	"math"
	"quill/ast"
	"strings"
)

// LLVMGenerator generates LLVM IR text from a Quill AST.
// It does NOT use LLVM C bindings; it emits .ll files as plain text.
type LLVMGenerator struct {
	output     strings.Builder
	tempCount  int
	strCount   int
	fmtCount   int
	labelCount int
	variables  map[string]llvmVar   // variable name -> alloca register + type
	strings    map[string]llvmStr   // string literal -> global name + length
	functions  map[string]llvmFunc  // Quill function name -> LLVM info
	indent     int
	// break/continue label stacks for loops
	breakLabels    []string
	continueLabels []string
}

type llvmVar struct {
	reg  string // e.g. %name
	typ  string // "double", "i8*", "i1"
}

type llvmStr struct {
	name string // e.g. @.str.0
	len  int    // length including null terminator
}

type llvmFunc struct {
	llName     string // e.g. @add
	retType    string
	paramTypes []string
}

// NewLLVM creates a new LLVM IR generator.
func NewLLVM() *LLVMGenerator {
	return &LLVMGenerator{
		variables: make(map[string]llvmVar),
		strings:   make(map[string]llvmStr),
		functions: make(map[string]llvmFunc),
	}
}

// Generate takes a Quill AST program and returns LLVM IR as a string.
func (g *LLVMGenerator) Generate(program *ast.Program) string {
	// First pass: collect string literals and function definitions
	g.collectStrings(program.Statements)
	g.collectFunctions(program.Statements)

	// Emit module header
	g.emitLine("; ModuleID = 'quill_module'")
	g.emitLine("source_filename = \"main.quill\"")
	g.emitLine("")

	// Emit string constants
	g.emitStringConstants()

	// Emit format strings
	g.emitLine("@.fmt.str = private constant [4 x i8] c\"%s\\0A\\00\"")
	g.emitLine("@.fmt.num = private constant [4 x i8] c\"%g\\0A\\00\"")
	g.emitLine("@.fmt.true = private constant [5 x i8] c\"true\\00\"")
	g.emitLine("@.fmt.false = private constant [6 x i8] c\"false\\00\"")
	g.emitLine("@.fmt.num_to_str = private constant [3 x i8] c\"%g\\00\"")
	g.emitLine("")

	// Emit external declarations
	g.emitLine("declare i32 @printf(i8*, ...)")
	g.emitLine("declare i32 @puts(i8*)")
	g.emitLine("declare i8* @malloc(i64)")
	g.emitLine("declare void @free(i8*)")
	g.emitLine("declare i64 @strlen(i8*)")
	g.emitLine("declare i8* @strcpy(i8*, i8*)")
	g.emitLine("declare i8* @strcat(i8*, i8*)")
	g.emitLine("declare i32 @strcmp(i8*, i8*)")
	g.emitLine("declare i32 @sprintf(i8*, i8*, ...)")
	g.emitLine("declare i32 @atoi(i8*)")
	g.emitLine("declare double @atof(i8*)")
	g.emitLine("")

	// Emit runtime helpers
	g.emitRuntimeHelpers()

	// Emit user-defined functions
	for _, stmt := range program.Statements {
		if fn, ok := stmt.(*ast.FuncDefinition); ok {
			g.emitFuncDefinition(fn)
		}
	}

	// Emit main function
	g.emitLine("define i32 @main() {")
	g.emitLine("entry:")
	g.indent = 1
	for _, stmt := range program.Statements {
		// Skip function definitions; they are emitted above
		if _, ok := stmt.(*ast.FuncDefinition); ok {
			continue
		}
		g.genStmt(stmt)
	}
	g.emitIndented("ret i32 0")
	g.emitLine("}")
	g.emitLine("")

	return g.output.String()
}

// --- String collection pass ---

func (g *LLVMGenerator) collectStrings(stmts []ast.Statement) {
	for _, stmt := range stmts {
		g.collectStringsInStmt(stmt)
	}
}

func (g *LLVMGenerator) collectStringsInStmt(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.AssignStatement:
		g.collectStringsInExpr(s.Value)
	case *ast.SayStatement:
		g.collectStringsInExpr(s.Value)
	case *ast.IfStatement:
		g.collectStringsInExpr(s.Condition)
		g.collectStrings(s.Body)
		for _, ei := range s.ElseIfs {
			g.collectStringsInExpr(ei.Condition)
			g.collectStrings(ei.Body)
		}
		g.collectStrings(s.Else)
	case *ast.WhileStatement:
		g.collectStringsInExpr(s.Condition)
		g.collectStrings(s.Body)
	case *ast.ForEachStatement:
		g.collectStringsInExpr(s.Iterable)
		g.collectStrings(s.Body)
	case *ast.FuncDefinition:
		g.collectStrings(s.Body)
	case *ast.ReturnStatement:
		if s.Value != nil {
			g.collectStringsInExpr(s.Value)
		}
	case *ast.ExprStatement:
		g.collectStringsInExpr(s.Expr)
	}
}

func (g *LLVMGenerator) collectStringsInExpr(expr ast.Expression) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.StringLiteral:
		g.registerString(e.Value)
	case *ast.BinaryExpr:
		g.collectStringsInExpr(e.Left)
		g.collectStringsInExpr(e.Right)
	case *ast.ComparisonExpr:
		g.collectStringsInExpr(e.Left)
		g.collectStringsInExpr(e.Right)
	case *ast.LogicalExpr:
		g.collectStringsInExpr(e.Left)
		g.collectStringsInExpr(e.Right)
	case *ast.NotExpr:
		g.collectStringsInExpr(e.Operand)
	case *ast.UnaryMinusExpr:
		g.collectStringsInExpr(e.Operand)
	case *ast.CallExpr:
		g.collectStringsInExpr(e.Function)
		for _, a := range e.Args {
			g.collectStringsInExpr(a)
		}
	case *ast.ListLiteral:
		for _, el := range e.Elements {
			g.collectStringsInExpr(el)
		}
	}
}

func (g *LLVMGenerator) registerString(s string) {
	if _, ok := g.strings[s]; ok {
		return
	}
	name := fmt.Sprintf("@.str.%d", g.strCount)
	g.strCount++
	g.strings[s] = llvmStr{name: name, len: len(s) + 1} // +1 for null
}

func (g *LLVMGenerator) emitStringConstants() {
	// We need deterministic ordering, so collect and sort by name
	type entry struct {
		literal string
		info    llvmStr
	}
	ordered := make([]entry, 0, len(g.strings))
	for lit, info := range g.strings {
		ordered = append(ordered, entry{lit, info})
	}
	// Sort by strCount embedded in the name
	for i := 0; i < len(ordered); i++ {
		for j := i + 1; j < len(ordered); j++ {
			if ordered[i].info.name > ordered[j].info.name {
				ordered[i], ordered[j] = ordered[j], ordered[i]
			}
		}
	}
	for _, e := range ordered {
		escaped := g.escapeLLVMString(e.literal)
		g.emitLine(fmt.Sprintf("%s = private constant [%d x i8] c\"%s\\00\"",
			e.info.name, e.info.len, escaped))
	}
	if len(ordered) > 0 {
		g.emitLine("")
	}
}

func (g *LLVMGenerator) escapeLLVMString(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\n':
			b.WriteString("\\0A")
		case '\r':
			b.WriteString("\\0D")
		case '\t':
			b.WriteString("\\09")
		case '"':
			b.WriteString("\\22")
		case '\\':
			b.WriteString("\\5C")
		default:
			if c < 32 || c > 126 {
				b.WriteString(fmt.Sprintf("\\%02X", c))
			} else {
				b.WriteByte(c)
			}
		}
	}
	return b.String()
}

// --- Function collection pass ---

func (g *LLVMGenerator) collectFunctions(stmts []ast.Statement) {
	for _, stmt := range stmts {
		if fn, ok := stmt.(*ast.FuncDefinition); ok {
			retType := g.mapQuillType(fn.ReturnType)
			if retType == "" {
				retType = "void"
			}
			paramTypes := make([]string, len(fn.Params))
			for i, pt := range fn.ParamTypes {
				t := g.mapQuillType(pt)
				if t == "" {
					// Default to double for untyped params
					t = "double"
				}
				paramTypes[i] = t
			}
			g.functions[fn.Name] = llvmFunc{
				llName:     "@" + fn.Name,
				retType:    retType,
				paramTypes: paramTypes,
			}
		}
	}
}

func (g *LLVMGenerator) mapQuillType(t string) string {
	switch t {
	case "number":
		return "double"
	case "text":
		return "i8*"
	case "boolean":
		return "i1"
	case "nothing", "":
		return "void"
	default:
		return "double"
	}
}

// --- Runtime helpers ---

func (g *LLVMGenerator) emitRuntimeHelpers() {
	// __quill_print_num
	g.emitLine("define void @__quill_print_num(double %n) {")
	g.emitLine("  %fmt = getelementptr [4 x i8], [4 x i8]* @.fmt.num, i32 0, i32 0")
	g.emitLine("  call i32 (i8*, ...) @printf(i8* %fmt, double %n)")
	g.emitLine("  ret void")
	g.emitLine("}")
	g.emitLine("")

	// __quill_print_str
	g.emitLine("define void @__quill_print_str(i8* %s) {")
	g.emitLine("  %fmt = getelementptr [4 x i8], [4 x i8]* @.fmt.str, i32 0, i32 0")
	g.emitLine("  call i32 (i8*, ...) @printf(i8* %fmt, i8* %s)")
	g.emitLine("  ret void")
	g.emitLine("}")
	g.emitLine("")

	// __quill_print_bool
	g.emitLine("define void @__quill_print_bool(i1 %b) {")
	g.emitLine("  br i1 %b, label %is_true, label %is_false")
	g.emitLine("is_true:")
	g.emitLine("  %t = getelementptr [5 x i8], [5 x i8]* @.fmt.true, i32 0, i32 0")
	g.emitLine("  %fmt1 = getelementptr [4 x i8], [4 x i8]* @.fmt.str, i32 0, i32 0")
	g.emitLine("  call i32 (i8*, ...) @printf(i8* %fmt1, i8* %t)")
	g.emitLine("  ret void")
	g.emitLine("is_false:")
	g.emitLine("  %f = getelementptr [6 x i8], [6 x i8]* @.fmt.false, i32 0, i32 0")
	g.emitLine("  %fmt2 = getelementptr [4 x i8], [4 x i8]* @.fmt.str, i32 0, i32 0")
	g.emitLine("  call i32 (i8*, ...) @printf(i8* %fmt2, i8* %f)")
	g.emitLine("  ret void")
	g.emitLine("}")
	g.emitLine("")

	// __quill_str_concat
	g.emitLine("define i8* @__quill_str_concat(i8* %a, i8* %b) {")
	g.emitLine("  %len_a = call i64 @strlen(i8* %a)")
	g.emitLine("  %len_b = call i64 @strlen(i8* %b)")
	g.emitLine("  %total = add i64 %len_a, %len_b")
	g.emitLine("  %total1 = add i64 %total, 1")
	g.emitLine("  %buf = call i8* @malloc(i64 %total1)")
	g.emitLine("  call i8* @strcpy(i8* %buf, i8* %a)")
	g.emitLine("  call i8* @strcat(i8* %buf, i8* %b)")
	g.emitLine("  ret i8* %buf")
	g.emitLine("}")
	g.emitLine("")

	// __quill_num_to_str
	g.emitLine("define i8* @__quill_num_to_str(double %n) {")
	g.emitLine("  %buf = call i8* @malloc(i64 64)")
	g.emitLine("  %fmt = getelementptr [3 x i8], [3 x i8]* @.fmt.num_to_str, i32 0, i32 0")
	g.emitLine("  call i32 (i8*, i8*, ...) @sprintf(i8* %buf, i8* %fmt, double %n)")
	g.emitLine("  ret i8* %buf")
	g.emitLine("}")
	g.emitLine("")
}

// --- Emit helpers ---

func (g *LLVMGenerator) emitLine(s string) {
	g.output.WriteString(s)
	g.output.WriteByte('\n')
}

func (g *LLVMGenerator) emitIndented(s string) {
	g.output.WriteString(strings.Repeat("  ", g.indent))
	g.output.WriteString(s)
	g.output.WriteByte('\n')
}

func (g *LLVMGenerator) nextTemp() string {
	g.tempCount++
	return fmt.Sprintf("%%t%d", g.tempCount)
}

func (g *LLVMGenerator) nextLabel(prefix string) string {
	g.labelCount++
	return fmt.Sprintf("%s.%d", prefix, g.labelCount)
}

// --- Statement generation ---

func (g *LLVMGenerator) genStmt(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.AssignStatement:
		g.genAssign(s)
	case *ast.TypedAssignStatement:
		g.genTypedAssign(s)
	case *ast.SayStatement:
		g.genSay(s)
	case *ast.IfStatement:
		g.genIf(s)
	case *ast.WhileStatement:
		g.genWhile(s)
	case *ast.ForEachStatement:
		g.genForEach(s)
	case *ast.ReturnStatement:
		g.genReturn(s)
	case *ast.ExprStatement:
		g.genExpr(s.Expr)
	case *ast.BreakStatement:
		if len(g.breakLabels) > 0 {
			g.emitIndented(fmt.Sprintf("br label %%%s", g.breakLabels[len(g.breakLabels)-1]))
		}
	case *ast.ContinueStatement:
		if len(g.continueLabels) > 0 {
			g.emitIndented(fmt.Sprintf("br label %%%s", g.continueLabels[len(g.continueLabels)-1]))
		}
	case *ast.FuncDefinition:
		// handled at top level, skip here

	// Concurrency stubs
	case *ast.SpawnStatement:
		g.emitIndented("; TODO: concurrency not yet supported in LLVM backend (spawn)")
	case *ast.ParallelBlock:
		g.emitIndented("; TODO: concurrency not yet supported in LLVM backend (parallel)")
	case *ast.RaceBlock:
		g.emitIndented("; TODO: concurrency not yet supported in LLVM backend (race)")
	case *ast.ChannelStatement:
		g.emitIndented("; TODO: concurrency not yet supported in LLVM backend (channel)")
	case *ast.SendStatement:
		g.emitIndented("; TODO: concurrency not yet supported in LLVM backend (send)")
	case *ast.SelectStatement:
		g.emitIndented("; TODO: concurrency not yet supported in LLVM backend (select)")
	}
}

func (g *LLVMGenerator) genAssign(s *ast.AssignStatement) {
	val := g.genExpr(s.Value)
	if existing, ok := g.variables[s.Name]; ok {
		// Re-assignment: store to existing alloca
		g.emitIndented(fmt.Sprintf("store %s %s, %s* %s", val.typ, val.reg, val.typ, existing.reg))
		return
	}
	// New variable: alloca + store
	reg := "%" + s.Name
	g.emitIndented(fmt.Sprintf("%s = alloca %s", reg, val.typ))
	g.emitIndented(fmt.Sprintf("store %s %s, %s* %s", val.typ, val.reg, val.typ, reg))
	g.variables[s.Name] = llvmVar{reg: reg, typ: val.typ}
}

func (g *LLVMGenerator) genTypedAssign(s *ast.TypedAssignStatement) {
	val := g.genExpr(s.Value)
	if existing, ok := g.variables[s.Name]; ok {
		g.emitIndented(fmt.Sprintf("store %s %s, %s* %s", val.typ, val.reg, val.typ, existing.reg))
		return
	}
	reg := "%" + s.Name
	g.emitIndented(fmt.Sprintf("%s = alloca %s", reg, val.typ))
	g.emitIndented(fmt.Sprintf("store %s %s, %s* %s", val.typ, val.reg, val.typ, reg))
	g.variables[s.Name] = llvmVar{reg: reg, typ: val.typ}
}

func (g *LLVMGenerator) genSay(s *ast.SayStatement) {
	val := g.genExpr(s.Value)
	switch val.typ {
	case "double":
		g.emitIndented(fmt.Sprintf("call void @__quill_print_num(double %s)", val.reg))
	case "i1":
		g.emitIndented(fmt.Sprintf("call void @__quill_print_bool(i1 %s)", val.reg))
	default: // i8*
		g.emitIndented(fmt.Sprintf("call void @__quill_print_str(i8* %s)", val.reg))
	}
}

func (g *LLVMGenerator) genIf(s *ast.IfStatement) {
	cond := g.genExpr(s.Condition)

	thenLabel := g.nextLabel("if.then")
	elseLabel := g.nextLabel("if.else")
	endLabel := g.nextLabel("if.end")

	if len(s.ElseIfs) == 0 && len(s.Else) == 0 {
		g.emitIndented(fmt.Sprintf("br i1 %s, label %%%s, label %%%s", cond.reg, thenLabel, endLabel))

		g.emitLine(thenLabel + ":")
		for _, st := range s.Body {
			g.genStmt(st)
		}
		g.emitIndented(fmt.Sprintf("br label %%%s", endLabel))

		g.emitLine(endLabel + ":")
		return
	}

	// Has else or else-ifs
	firstElseLabel := elseLabel
	if len(s.ElseIfs) > 0 {
		firstElseLabel = g.nextLabel("elif.cond")
	}
	g.emitIndented(fmt.Sprintf("br i1 %s, label %%%s, label %%%s", cond.reg, thenLabel, firstElseLabel))

	// Then block
	g.emitLine(thenLabel + ":")
	for _, st := range s.Body {
		g.genStmt(st)
	}
	g.emitIndented(fmt.Sprintf("br label %%%s", endLabel))

	// Else-if blocks
	for i, ei := range s.ElseIfs {
		condLabel := firstElseLabel
		if i > 0 {
			condLabel = g.nextLabel("elif.cond")
		}
		_ = condLabel // already emitted as branch target
		if i > 0 {
			// Need to re-derive since we used nextLabel above
		}
		g.emitLine(fmt.Sprintf("%s:", func() string {
			if i == 0 {
				return firstElseLabel
			}
			return condLabel
		}()))
		eiCond := g.genExpr(ei.Condition)
		eiThenLabel := g.nextLabel("elif.then")
		var eiNextLabel string
		if i+1 < len(s.ElseIfs) {
			eiNextLabel = g.nextLabel("elif.cond")
		} else if len(s.Else) > 0 {
			eiNextLabel = elseLabel
		} else {
			eiNextLabel = endLabel
		}
		g.emitIndented(fmt.Sprintf("br i1 %s, label %%%s, label %%%s", eiCond.reg, eiThenLabel, eiNextLabel))

		g.emitLine(eiThenLabel + ":")
		for _, st := range ei.Body {
			g.genStmt(st)
		}
		g.emitIndented(fmt.Sprintf("br label %%%s", endLabel))

		if i+1 < len(s.ElseIfs) {
			firstElseLabel = eiNextLabel
		}
	}

	// Else block
	if len(s.Else) > 0 {
		g.emitLine(elseLabel + ":")
		for _, st := range s.Else {
			g.genStmt(st)
		}
		g.emitIndented(fmt.Sprintf("br label %%%s", endLabel))
	}

	g.emitLine(endLabel + ":")
}

func (g *LLVMGenerator) genWhile(s *ast.WhileStatement) {
	condLabel := g.nextLabel("while.cond")
	bodyLabel := g.nextLabel("while.body")
	endLabel := g.nextLabel("while.end")

	g.breakLabels = append(g.breakLabels, endLabel)
	g.continueLabels = append(g.continueLabels, condLabel)

	g.emitIndented(fmt.Sprintf("br label %%%s", condLabel))
	g.emitLine(condLabel + ":")
	cond := g.genExpr(s.Condition)
	g.emitIndented(fmt.Sprintf("br i1 %s, label %%%s, label %%%s", cond.reg, bodyLabel, endLabel))

	g.emitLine(bodyLabel + ":")
	for _, st := range s.Body {
		g.genStmt(st)
	}
	g.emitIndented(fmt.Sprintf("br label %%%s", condLabel))

	g.emitLine(endLabel + ":")

	g.breakLabels = g.breakLabels[:len(g.breakLabels)-1]
	g.continueLabels = g.continueLabels[:len(g.continueLabels)-1]
}

func (g *LLVMGenerator) genForEach(s *ast.ForEachStatement) {
	// Simplified: treat iterable as a list-like structure.
	// For now, emit a comment and a simple counted loop pattern.
	// This is a simplified approach; real implementation would need array runtime.
	condLabel := g.nextLabel("foreach.cond")
	bodyLabel := g.nextLabel("foreach.body")
	endLabel := g.nextLabel("foreach.end")

	g.breakLabels = append(g.breakLabels, endLabel)
	g.continueLabels = append(g.continueLabels, condLabel)

	// Allocate loop index
	idxReg := g.nextTemp()
	g.emitIndented(fmt.Sprintf("%s = alloca i64", idxReg))
	g.emitIndented(fmt.Sprintf("store i64 0, i64* %s", idxReg))

	// For now, iterate 0 times (placeholder for list support)
	g.emitIndented(fmt.Sprintf("; foreach %s (simplified - list runtime not fully implemented)", s.Variable))
	g.emitIndented(fmt.Sprintf("br label %%%s", condLabel))

	g.emitLine(condLabel + ":")
	curIdx := g.nextTemp()
	g.emitIndented(fmt.Sprintf("%s = load i64, i64* %s", curIdx, idxReg))
	cmp := g.nextTemp()
	g.emitIndented(fmt.Sprintf("%s = icmp slt i64 %s, 0", cmp, curIdx)) // 0-length placeholder
	g.emitIndented(fmt.Sprintf("br i1 %s, label %%%s, label %%%s", cmp, bodyLabel, endLabel))

	g.emitLine(bodyLabel + ":")
	// Allocate loop variable
	loopVarReg := "%" + s.Variable
	g.emitIndented(fmt.Sprintf("%s = alloca double", loopVarReg))
	g.variables[s.Variable] = llvmVar{reg: loopVarReg, typ: "double"}

	for _, st := range s.Body {
		g.genStmt(st)
	}

	nextIdx := g.nextTemp()
	g.emitIndented(fmt.Sprintf("%s = add i64 %s, 1", nextIdx, curIdx))
	g.emitIndented(fmt.Sprintf("store i64 %s, i64* %s", nextIdx, idxReg))
	g.emitIndented(fmt.Sprintf("br label %%%s", condLabel))

	g.emitLine(endLabel + ":")

	g.breakLabels = g.breakLabels[:len(g.breakLabels)-1]
	g.continueLabels = g.continueLabels[:len(g.continueLabels)-1]
}

func (g *LLVMGenerator) genReturn(s *ast.ReturnStatement) {
	if s.Value == nil {
		g.emitIndented("ret void")
		return
	}
	val := g.genExpr(s.Value)
	g.emitIndented(fmt.Sprintf("ret %s %s", val.typ, val.reg))
}

func (g *LLVMGenerator) emitFuncDefinition(fn *ast.FuncDefinition) {
	info := g.functions[fn.Name]

	// Build param list
	params := make([]string, len(fn.Params))
	for i, p := range fn.Params {
		params[i] = fmt.Sprintf("%s %%%s", info.paramTypes[i], p)
	}

	g.emitLine(fmt.Sprintf("define %s @%s(%s) {", info.retType, fn.Name, strings.Join(params, ", ")))
	g.emitLine("entry:")

	// Save outer variable scope
	outerVars := g.variables
	g.variables = make(map[string]llvmVar)
	// Copy outer scope (for closures - simplified)
	for k, v := range outerVars {
		g.variables[k] = v
	}

	oldIndent := g.indent
	g.indent = 1

	// Alloca params so they can be re-assigned
	for i, p := range fn.Params {
		reg := "%" + p + ".addr"
		g.emitIndented(fmt.Sprintf("%s = alloca %s", reg, info.paramTypes[i]))
		g.emitIndented(fmt.Sprintf("store %s %%%s, %s* %s", info.paramTypes[i], p, info.paramTypes[i], reg))
		g.variables[p] = llvmVar{reg: reg, typ: info.paramTypes[i]}
	}

	for _, stmt := range fn.Body {
		g.genStmt(stmt)
	}

	// If the function is void and doesn't end with a return, add one
	if info.retType == "void" {
		g.emitIndented("ret void")
	}

	g.indent = oldIndent
	g.emitLine("}")
	g.emitLine("")

	g.variables = outerVars
}

// --- Expression generation ---

type llvmValue struct {
	reg string
	typ string
}

func (g *LLVMGenerator) genExpr(expr ast.Expression) llvmValue {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return g.genNumber(e)
	case *ast.StringLiteral:
		return g.genString(e)
	case *ast.BoolLiteral:
		return g.genBool(e)
	case *ast.NothingLiteral:
		return llvmValue{reg: "null", typ: "i8*"}
	case *ast.Identifier:
		return g.genIdentifier(e)
	case *ast.BinaryExpr:
		return g.genBinary(e)
	case *ast.ComparisonExpr:
		return g.genComparison(e)
	case *ast.LogicalExpr:
		return g.genLogical(e)
	case *ast.NotExpr:
		return g.genNot(e)
	case *ast.UnaryMinusExpr:
		return g.genUnaryMinus(e)
	case *ast.CallExpr:
		return g.genCall(e)
	case *ast.ListLiteral:
		// Simplified: return null pointer
		return llvmValue{reg: "null", typ: "i8*"}
	case *ast.AwaitExpression:
		g.emitIndented("; TODO: concurrency not yet supported in LLVM backend (await expression)")
		return llvmValue{reg: "0", typ: "i64"}
	case *ast.ReceiveExpression:
		g.emitIndented("; TODO: concurrency not yet supported in LLVM backend (receive)")
		return llvmValue{reg: "0", typ: "i64"}
	default:
		// Unsupported expression, emit a comment and return zero
		g.emitIndented(fmt.Sprintf("; unsupported expression type: %T", expr))
		return llvmValue{reg: "0", typ: "i64"}
	}
}

func (g *LLVMGenerator) genNumber(e *ast.NumberLiteral) llvmValue {
	// Format as LLVM double hex if needed, or use scientific notation
	if e.Value == float64(int64(e.Value)) && !math.IsInf(e.Value, 0) {
		return llvmValue{
			reg: fmt.Sprintf("%g.0", e.Value),
			typ: "double",
		}
	}
	return llvmValue{
		reg: fmt.Sprintf("%g", e.Value),
		typ: "double",
	}
}

func (g *LLVMGenerator) genString(e *ast.StringLiteral) llvmValue {
	info := g.strings[e.Value]
	tmp := g.nextTemp()
	g.emitIndented(fmt.Sprintf("%s = getelementptr [%d x i8], [%d x i8]* %s, i32 0, i32 0",
		tmp, info.len, info.len, info.name))
	return llvmValue{reg: tmp, typ: "i8*"}
}

func (g *LLVMGenerator) genBool(e *ast.BoolLiteral) llvmValue {
	if e.Value {
		return llvmValue{reg: "1", typ: "i1"}
	}
	return llvmValue{reg: "0", typ: "i1"}
}

func (g *LLVMGenerator) genIdentifier(e *ast.Identifier) llvmValue {
	if v, ok := g.variables[e.Name]; ok {
		tmp := g.nextTemp()
		g.emitIndented(fmt.Sprintf("%s = load %s, %s* %s", tmp, v.typ, v.typ, v.reg))
		return llvmValue{reg: tmp, typ: v.typ}
	}
	// Might be a function reference or unknown; return as i64 0
	g.emitIndented(fmt.Sprintf("; warning: unknown variable '%s'", e.Name))
	return llvmValue{reg: "0", typ: "double"}
}

func (g *LLVMGenerator) genBinary(e *ast.BinaryExpr) llvmValue {
	left := g.genExpr(e.Left)
	right := g.genExpr(e.Right)

	// String concatenation
	if left.typ == "i8*" && right.typ == "i8*" && e.Operator == "+" {
		tmp := g.nextTemp()
		g.emitIndented(fmt.Sprintf("%s = call i8* @__quill_str_concat(i8* %s, i8* %s)", tmp, left.reg, right.reg))
		return llvmValue{reg: tmp, typ: "i8*"}
	}

	// String + number or number + string: convert number to string then concat
	if left.typ == "i8*" && right.typ == "double" && e.Operator == "+" {
		convTmp := g.nextTemp()
		g.emitIndented(fmt.Sprintf("%s = call i8* @__quill_num_to_str(double %s)", convTmp, right.reg))
		tmp := g.nextTemp()
		g.emitIndented(fmt.Sprintf("%s = call i8* @__quill_str_concat(i8* %s, i8* %s)", tmp, left.reg, convTmp))
		return llvmValue{reg: tmp, typ: "i8*"}
	}
	if left.typ == "double" && right.typ == "i8*" && e.Operator == "+" {
		convTmp := g.nextTemp()
		g.emitIndented(fmt.Sprintf("%s = call i8* @__quill_num_to_str(double %s)", convTmp, left.reg))
		tmp := g.nextTemp()
		g.emitIndented(fmt.Sprintf("%s = call i8* @__quill_str_concat(i8* %s, i8* %s)", tmp, convTmp, right.reg))
		return llvmValue{reg: tmp, typ: "i8*"}
	}

	// Numeric operations
	tmp := g.nextTemp()
	var op string
	switch e.Operator {
	case "+":
		op = "fadd"
	case "-":
		op = "fsub"
	case "*":
		op = "fmul"
	case "/":
		op = "fdiv"
	case "%":
		op = "frem"
	default:
		op = "fadd"
	}
	g.emitIndented(fmt.Sprintf("%s = %s double %s, %s", tmp, op, left.reg, right.reg))
	return llvmValue{reg: tmp, typ: "double"}
}

func (g *LLVMGenerator) genComparison(e *ast.ComparisonExpr) llvmValue {
	left := g.genExpr(e.Left)
	right := g.genExpr(e.Right)

	// String comparison
	if left.typ == "i8*" && right.typ == "i8*" {
		cmpResult := g.nextTemp()
		g.emitIndented(fmt.Sprintf("%s = call i32 @strcmp(i8* %s, i8* %s)", cmpResult, left.reg, right.reg))
		tmp := g.nextTemp()
		switch e.Operator {
		case "==":
			g.emitIndented(fmt.Sprintf("%s = icmp eq i32 %s, 0", tmp, cmpResult))
		case "!=":
			g.emitIndented(fmt.Sprintf("%s = icmp ne i32 %s, 0", tmp, cmpResult))
		default:
			g.emitIndented(fmt.Sprintf("%s = icmp eq i32 %s, 0", tmp, cmpResult))
		}
		return llvmValue{reg: tmp, typ: "i1"}
	}

	// Numeric comparison (fcmp)
	tmp := g.nextTemp()
	var cond string
	switch e.Operator {
	case ">":
		cond = "ogt"
	case "<":
		cond = "olt"
	case ">=":
		cond = "oge"
	case "<=":
		cond = "ole"
	case "==":
		cond = "oeq"
	case "!=":
		cond = "one"
	default:
		cond = "oeq"
	}
	g.emitIndented(fmt.Sprintf("%s = fcmp %s double %s, %s", tmp, cond, left.reg, right.reg))
	return llvmValue{reg: tmp, typ: "i1"}
}

func (g *LLVMGenerator) genLogical(e *ast.LogicalExpr) llvmValue {
	left := g.genExpr(e.Left)
	right := g.genExpr(e.Right)

	tmp := g.nextTemp()
	switch e.Operator {
	case "and":
		g.emitIndented(fmt.Sprintf("%s = and i1 %s, %s", tmp, left.reg, right.reg))
	case "or":
		g.emitIndented(fmt.Sprintf("%s = or i1 %s, %s", tmp, left.reg, right.reg))
	default:
		g.emitIndented(fmt.Sprintf("%s = and i1 %s, %s", tmp, left.reg, right.reg))
	}
	return llvmValue{reg: tmp, typ: "i1"}
}

func (g *LLVMGenerator) genNot(e *ast.NotExpr) llvmValue {
	operand := g.genExpr(e.Operand)
	tmp := g.nextTemp()
	g.emitIndented(fmt.Sprintf("%s = xor i1 %s, 1", tmp, operand.reg))
	return llvmValue{reg: tmp, typ: "i1"}
}

func (g *LLVMGenerator) genUnaryMinus(e *ast.UnaryMinusExpr) llvmValue {
	operand := g.genExpr(e.Operand)
	tmp := g.nextTemp()
	g.emitIndented(fmt.Sprintf("%s = fsub double 0.0, %s", tmp, operand.reg))
	return llvmValue{reg: tmp, typ: "double"}
}

func (g *LLVMGenerator) genCall(e *ast.CallExpr) llvmValue {
	// Check for built-in functions
	if ident, ok := e.Function.(*ast.Identifier); ok {
		switch ident.Name {
		case "toText":
			if len(e.Args) == 1 {
				arg := g.genExpr(e.Args[0])
				if arg.typ == "double" {
					tmp := g.nextTemp()
					g.emitIndented(fmt.Sprintf("%s = call i8* @__quill_num_to_str(double %s)", tmp, arg.reg))
					return llvmValue{reg: tmp, typ: "i8*"}
				}
				return arg // already a string
			}
		case "toNumber":
			if len(e.Args) == 1 {
				arg := g.genExpr(e.Args[0])
				if arg.typ == "i8*" {
					tmp := g.nextTemp()
					g.emitIndented(fmt.Sprintf("%s = call double @atof(i8* %s)", tmp, arg.reg))
					return llvmValue{reg: tmp, typ: "double"}
				}
				return arg
			}
		case "length":
			if len(e.Args) == 1 {
				arg := g.genExpr(e.Args[0])
				if arg.typ == "i8*" {
					lenTmp := g.nextTemp()
					g.emitIndented(fmt.Sprintf("%s = call i64 @strlen(i8* %s)", lenTmp, arg.reg))
					tmp := g.nextTemp()
					g.emitIndented(fmt.Sprintf("%s = sitofp i64 %s to double", tmp, lenTmp))
					return llvmValue{reg: tmp, typ: "double"}
				}
			}
		}

		// User-defined function call
		if fn, ok := g.functions[ident.Name]; ok {
			args := make([]string, len(e.Args))
			for i, a := range e.Args {
				val := g.genExpr(a)
				argType := "double"
				if i < len(fn.paramTypes) {
					argType = fn.paramTypes[i]
				}
				args[i] = fmt.Sprintf("%s %s", argType, val.reg)
			}
			if fn.retType == "void" {
				g.emitIndented(fmt.Sprintf("call void @%s(%s)", ident.Name, strings.Join(args, ", ")))
				return llvmValue{reg: "0", typ: "double"}
			}
			tmp := g.nextTemp()
			g.emitIndented(fmt.Sprintf("%s = call %s @%s(%s)", tmp, fn.retType, ident.Name, strings.Join(args, ", ")))
			return llvmValue{reg: tmp, typ: fn.retType}
		}
	}

	// Fallback: unknown function call
	g.emitIndented("; warning: unknown function call")
	return llvmValue{reg: "0", typ: "double"}
}
