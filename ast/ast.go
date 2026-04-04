package ast

// Node is the base interface for all AST nodes.
type Node interface {
	nodeType() string
}

// Statement nodes
type Statement interface {
	Node
	stmtNode()
}

// Expression nodes
type Expression interface {
	Node
	exprNode()
}

// Program is the root node of every Quill program.
type Program struct {
	Statements []Statement
}

func (p *Program) nodeType() string { return "Program" }

// --- Statements ---

type AssignStatement struct {
	Name  string
	Value Expression
	Line  int
}

func (s *AssignStatement) nodeType() string { return "Assign" }
func (s *AssignStatement) stmtNode()        {}

type SayStatement struct {
	Value Expression
	Line  int
}

func (s *SayStatement) nodeType() string { return "Say" }
func (s *SayStatement) stmtNode()        {}

type IfStatement struct {
	Condition Expression
	Body      []Statement
	ElseIfs   []ElseIfClause
	Else      []Statement
	Line      int
}

type ElseIfClause struct {
	Condition Expression
	Body      []Statement
}

func (s *IfStatement) nodeType() string { return "If" }
func (s *IfStatement) stmtNode()        {}

type ForEachStatement struct {
	Variable string
	Iterable Expression
	Body     []Statement
	Line     int
}

func (s *ForEachStatement) nodeType() string { return "ForEach" }
func (s *ForEachStatement) stmtNode()        {}

type WhileStatement struct {
	Condition Expression
	Body      []Statement
	Line      int
}

func (s *WhileStatement) nodeType() string { return "While" }
func (s *WhileStatement) stmtNode()        {}

type FuncDefinition struct {
	Name   string
	Params []string
	Body   []Statement
	Line   int
}

func (s *FuncDefinition) nodeType() string { return "FuncDef" }
func (s *FuncDefinition) stmtNode()        {}

type ReturnStatement struct {
	Value Expression
	Line  int
}

func (s *ReturnStatement) nodeType() string { return "Return" }
func (s *ReturnStatement) stmtNode()        {}

type ExprStatement struct {
	Expr Expression
	Line int
}

func (s *ExprStatement) nodeType() string { return "ExprStmt" }
func (s *ExprStatement) stmtNode()        {}

type UseStatement struct {
	Path string
	Line int
}

func (s *UseStatement) nodeType() string { return "Use" }
func (s *UseStatement) stmtNode()        {}

type TestBlock struct {
	Name string
	Body []Statement
	Line int
}

func (s *TestBlock) nodeType() string { return "Test" }
func (s *TestBlock) stmtNode()        {}

type ExpectStatement struct {
	Expr Expression
	Line int
}

func (s *ExpectStatement) nodeType() string { return "Expect" }
func (s *ExpectStatement) stmtNode()        {}

// --- Expressions ---

type StringLiteral struct {
	Value string
}

func (e *StringLiteral) nodeType() string { return "String" }
func (e *StringLiteral) exprNode()        {}

type NumberLiteral struct {
	Value float64
}

func (e *NumberLiteral) nodeType() string { return "Number" }
func (e *NumberLiteral) exprNode()        {}

type BoolLiteral struct {
	Value bool
}

func (e *BoolLiteral) nodeType() string { return "Bool" }
func (e *BoolLiteral) exprNode()        {}

type ListLiteral struct {
	Elements []Expression
}

func (e *ListLiteral) nodeType() string { return "List" }
func (e *ListLiteral) exprNode()        {}

type Identifier struct {
	Name string
}

func (e *Identifier) nodeType() string { return "Ident" }
func (e *Identifier) exprNode()        {}

type BinaryExpr struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (e *BinaryExpr) nodeType() string { return "Binary" }
func (e *BinaryExpr) exprNode()        {}

type ComparisonExpr struct {
	Left     Expression
	Operator string // ">", "<", "==", "!=", ">=", "<=", "contains"
	Right    Expression
}

func (e *ComparisonExpr) nodeType() string { return "Comparison" }
func (e *ComparisonExpr) exprNode()        {}

type LogicalExpr struct {
	Left     Expression
	Operator string // "and", "or"
	Right    Expression
}

func (e *LogicalExpr) nodeType() string { return "Logical" }
func (e *LogicalExpr) exprNode()        {}

type NotExpr struct {
	Operand Expression
}

func (e *NotExpr) nodeType() string { return "Not" }
func (e *NotExpr) exprNode()        {}

type UnaryMinusExpr struct {
	Operand Expression
}

func (e *UnaryMinusExpr) nodeType() string { return "UnaryMinus" }
func (e *UnaryMinusExpr) exprNode()        {}

type CallExpr struct {
	Function Expression
	Args     []Expression
}

func (e *CallExpr) nodeType() string { return "Call" }
func (e *CallExpr) exprNode()        {}

type DotExpr struct {
	Object Expression
	Field  string
}

func (e *DotExpr) nodeType() string { return "Dot" }
func (e *DotExpr) exprNode()        {}

type IndexExpr struct {
	Object Expression
	Index  Expression
}

func (e *IndexExpr) nodeType() string { return "Index" }
func (e *IndexExpr) exprNode()        {}
