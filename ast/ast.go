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
	Variable   string
	Iterable   Expression
	Body       []Statement
	IsAsync    bool
	DestructurePattern DestructurePattern // optional: for destructuring in for-each
	Line       int
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
	Name       string
	Params     []string
	ParamTypes []string    // parallel to Params, empty string = no annotation
	ReturnType string      // optional return type annotation
	TypeParams []TypeParam // generic type parameters with optional constraints
	Body       []Statement
	Line       int
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
	Path  string
	Alias string
	Line  int
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

// MockStatement represents a mock block in a test: mock <func> with <params>: <body>
type MockStatement struct {
	FuncName string
	Params   []string
	Body     []Statement
	Line     int
}

func (s *MockStatement) nodeType() string { return "Mock" }
func (s *MockStatement) stmtNode()        {}

// MockAssertionExpr represents an assertion like: fetchJSON was called 1 time
type MockAssertionExpr struct {
	FuncName   string
	AssertType string // "called"
	Count      int
}

func (e *MockAssertionExpr) nodeType() string { return "MockAssertion" }
func (e *MockAssertionExpr) exprNode()        {}

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

// --- New nodes for classes, async, etc. ---

type DescribeStatement struct {
	Name                 string
	Extends              string // parent class name (optional)
	Properties           []AssignStatement
	Methods              []FuncDefinition
	PropertyVisibilities []string // parallel to Properties: "public", "private", or ""
	MethodVisibilities   []string // parallel to Methods: "public", "private", or ""
	Line                 int
}

func (s *DescribeStatement) nodeType() string { return "Describe" }
func (s *DescribeStatement) stmtNode()        {}

type DotAssignStatement struct {
	Object string
	Field  string
	Value  Expression
	Line   int
}

func (s *DotAssignStatement) nodeType() string { return "DotAssign" }
func (s *DotAssignStatement) stmtNode()        {}

type NewExpr struct {
	ClassName string
	Args      []Expression
}

func (e *NewExpr) nodeType() string { return "New" }
func (e *NewExpr) exprNode()        {}

type AwaitExpr struct {
	Expr Expression
}

func (e *AwaitExpr) nodeType() string { return "Await" }
func (e *AwaitExpr) exprNode()        {}

// --- Try/Catch ---

type TryCatchStatement struct {
	TryBody   []Statement
	ErrorVar  string
	CatchBody []Statement
	Line      int
}

func (s *TryCatchStatement) nodeType() string { return "TryCatch" }
func (s *TryCatchStatement) stmtNode()        {}

// --- Break/Continue ---

type BreakStatement struct {
	Line int
}

func (s *BreakStatement) nodeType() string { return "Break" }
func (s *BreakStatement) stmtNode()        {}

type ContinueStatement struct {
	Line int
}

func (s *ContinueStatement) nodeType() string { return "Continue" }
func (s *ContinueStatement) stmtNode()        {}

// --- Nothing literal ---

type NothingLiteral struct{}

func (e *NothingLiteral) nodeType() string { return "Nothing" }
func (e *NothingLiteral) exprNode()        {}

// --- Object/Map literal ---

// ComputedProperty represents a computed property in an object literal: {[expr]: value}
type ComputedProperty struct {
	KeyExpr Expression
	Value   Expression
}

type ObjectLiteral struct {
	Keys               []string
	Values             []Expression
	ComputedProperties []ComputedProperty
}

func (e *ObjectLiteral) nodeType() string { return "Object" }
func (e *ObjectLiteral) exprNode()        {}

// --- Tagged template expression ---

type TaggedTemplateExpr struct {
	Tag         string
	Template    string
	Expressions []Expression
	Line        int
}

func (e *TaggedTemplateExpr) nodeType() string { return "TaggedTemplate" }
func (e *TaggedTemplateExpr) exprNode()        {}

// --- Lambda/Arrow function ---

type LambdaExpr struct {
	Params []string
	Body   Expression
}

func (e *LambdaExpr) nodeType() string { return "Lambda" }
func (e *LambdaExpr) exprNode()        {}

// --- Spread expression ---

type SpreadExpr struct {
	Expr Expression
}

func (e *SpreadExpr) nodeType() string { return "Spread" }
func (e *SpreadExpr) exprNode()        {}

// --- Type annotation (optional, informational) ---

type TypeAnnotation struct {
	Name string // "text", "number", "list", "object", "boolean", custom
}

// --- Enhanced FuncDefinition with types and async ---

type TypedParam struct {
	Name     string
	TypeHint string // optional type annotation
}

type FuncDefTyped struct {
	Name       string
	Params     []TypedParam
	ReturnType string // optional
	IsAsync    bool
	Body       []Statement
	Line       int
}

func (s *FuncDefTyped) nodeType() string { return "FuncDefTyped" }
func (s *FuncDefTyped) stmtNode()        {}

// --- From import ---

type FromUseStatement struct {
	Names []string
	Path  string
	Line  int
}

func (s *FromUseStatement) nodeType() string { return "FromUse" }
func (s *FromUseStatement) stmtNode()        {}

// --- Pattern Matching ---

type MatchCase struct {
	Pattern     Expression // the value to match against (or nil for otherwise)
	Guard       Expression // optional "if" guard condition
	TypePattern string     // type-based pattern: "text", "number", "list", "nothing"
	Binding     string     // variable binding for type pattern
	Body        []Statement
}

type MatchStatement struct {
	Value Expression
	Cases []MatchCase
	Line  int
}

func (s *MatchStatement) nodeType() string { return "Match" }
func (s *MatchStatement) stmtNode()        {}

// --- Enum / Algebraic Types ---

type EnumVariant struct {
	Name   string
	Fields []string     // optional fields for algebraic data types
	Value  Expression   // optional associated value (e.g., OK is 200)
}

type DefineStatement struct {
	Name     string
	Variants []EnumVariant
	Methods  []FuncDefinition // methods defined on the enum
	Line     int
}

func (s *DefineStatement) nodeType() string { return "Define" }
func (s *DefineStatement) stmtNode()        {}

// --- Pipe expression ---

type PipeExpr struct {
	Left  Expression
	Right Expression // must be a CallExpr or Identifier
}

func (e *PipeExpr) nodeType() string { return "Pipe" }
func (e *PipeExpr) exprNode()        {}

// --- Enhanced type info on FuncDefinition for type checker ---

type FuncParamType struct {
	Name     string
	TypeHint string // empty if no annotation
}

// TypedAssignStatement is like AssignStatement but with type info
type TypedAssignStatement struct {
	Name     string
	TypeHint string // e.g. "number", "text", "list of number"
	Value    Expression
	Line     int
}

func (s *TypedAssignStatement) nodeType() string { return "TypedAssign" }
func (s *TypedAssignStatement) stmtNode()        {}

// --- Error Propagation ---

// PropagateExpr represents an expression followed by ? (auto-propagate errors).
type PropagateExpr struct {
	Expr Expression
}

func (e *PropagateExpr) nodeType() string { return "Propagate" }
func (e *PropagateExpr) exprNode()        {}

// TryExpression wraps an expression with try semantics.
type TryExpression struct {
	Expr Expression
}

func (e *TryExpression) nodeType() string { return "TryExpr" }
func (e *TryExpression) exprNode()        {}

// --- Destructured Match Pattern ---

// ObjectMatchPattern represents a {key: value, ...} pattern in match/when.
type ObjectMatchPattern struct {
	Fields []ObjectMatchField
}

func (e *ObjectMatchPattern) nodeType() string { return "ObjectMatchPattern" }
func (e *ObjectMatchPattern) exprNode()        {}

// ObjectMatchField represents a single field in an object match pattern.
type ObjectMatchField struct {
	Key   string
	Value Expression // if non-nil, checks equality; if nil, just binds the key
}

// --- Reactive UI Framework ---

// CSSRule represents a single CSS rule in a scoped style block.
type CSSRule struct {
	Selector   string
	Properties map[string]string
}

// StyleBlock represents scoped styles within a component.
type StyleBlock struct {
	Rules []CSSRule
}

// LoadFunction represents a server-side data loader in a component.
type LoadFunction struct {
	Param string // e.g. "request"
	Body  []Statement
}

// FormAction represents a form action handler in a component.
type FormAction struct {
	Name  string
	Param string
	Body  []Statement
}

// HeadEntry represents a single element in a head block (title, meta, link).
type HeadEntry struct {
	Tag   string            // "title", "meta", "link"
	Text  string            // for title text content
	Attrs map[string]string // for meta/link attributes
}

// HeadBlock represents a head management block.
type HeadBlock struct {
	Entries []HeadEntry
	Line    int
}

func (s *HeadBlock) nodeType() string { return "Head" }
func (s *HeadBlock) stmtNode()       {}

// LinkElement represents a client-side navigation link.
type LinkElement struct {
	To   string // path expression e.g. "/about" or "/blog/{post.id}"
	Text Expression
	Line int
}

// ComponentStatement represents a reactive UI component definition.
type ComponentStatement struct {
	Name       string
	States     []StateDeclaration
	Methods    []FuncDefinition
	RenderBody []RenderElement
	Styles     *StyleBlock
	Loader     *LoadFunction
	Actions    []FormAction
	Head       *HeadBlock
	Line       int
}

func (s *ComponentStatement) nodeType() string { return "Component" }
func (s *ComponentStatement) stmtNode()        {}

// StateDeclaration represents a reactive state variable inside a component.
type StateDeclaration struct {
	Name  string
	Value Expression
	Line  int
}

func (s *StateDeclaration) nodeType() string { return "StateDecl" }
func (s *StateDeclaration) stmtNode()        {}

// RenderElement represents an HTML element in the render block.
type RenderElement struct {
	Tag       string
	Props     map[string]Expression // attributes and event handlers
	Children  []RenderNode          // child elements or text
	Condition Expression            // for conditional rendering (if)
	Iterator  *RenderIterator       // for list rendering (for each)
	Line      int
}

// RenderNode is either an element or text content in a render block.
type RenderNode struct {
	Element *RenderElement // either an element...
	Text    Expression     // ...or text content (string interpolation)
}

// RenderIterator represents a for-each loop in a render block.
type RenderIterator struct {
	Variable string
	Iterable Expression
}

// MountStatement represents mounting a component to a DOM selector.
type MountStatement struct {
	Component string
	Selector  Expression
	Line      int
}

func (s *MountStatement) nodeType() string { return "Mount" }
func (s *MountStatement) stmtNode()        {}

// --- Concurrency ---

// SpawnStatement represents spawning a concurrent task.
type SpawnStatement struct {
	Name string
	Body []Statement
	Line int
}

func (s *SpawnStatement) nodeType() string { return "Spawn" }
func (s *SpawnStatement) stmtNode()        {}

// AwaitExpression represents awaiting a task, "all", or "first".
type AwaitExpression struct {
	Target Expression // variable or identifier "all" / "first"
	Line   int
}

func (e *AwaitExpression) nodeType() string { return "AwaitExpr" }
func (e *AwaitExpression) exprNode()        {}

// CancelStatement represents cancelling a spawned task.
type CancelStatement struct {
	Target string
	Line   int
}

func (s *CancelStatement) nodeType() string { return "Cancel" }
func (s *CancelStatement) stmtNode()        {}

// ParallelBlock represents a parallel: block with multiple task assignments.
type ParallelBlock struct {
	Tasks     []Statement // assignments inside the block
	IsSettled bool        // if true, use Promise.allSettled
	Line      int
}

func (s *ParallelBlock) nodeType() string { return "Parallel" }
func (s *ParallelBlock) stmtNode()        {}

// RaceBlock represents a race: block with multiple task assignments.
type RaceBlock struct {
	Tasks []Statement
	Line  int
}

func (s *RaceBlock) nodeType() string { return "Race" }
func (s *RaceBlock) stmtNode()        {}

// ChannelStatement represents a channel declaration.
type ChannelStatement struct {
	Name       string
	BufferSize Expression // optional
	Line       int
}

func (s *ChannelStatement) nodeType() string { return "Channel" }
func (s *ChannelStatement) stmtNode()        {}

// SendStatement represents sending a value to a channel.
type SendStatement struct {
	Value   Expression
	Channel string
	Line    int
}

func (s *SendStatement) nodeType() string { return "Send" }
func (s *SendStatement) stmtNode()        {}

// ReceiveExpression represents receiving a value from a channel.
type ReceiveExpression struct {
	Channel string
	Line    int
}

func (e *ReceiveExpression) nodeType() string { return "Receive" }
func (e *ReceiveExpression) exprNode()        {}

// SelectStatement represents a select: block with channel cases and optional timeout.
type SelectStatement struct {
	Cases     []SelectCase
	AfterMs   Expression  // optional timeout in ms
	AfterBody []Statement // body to execute on timeout
	Line      int
}

func (s *SelectStatement) nodeType() string { return "Select" }
func (s *SelectStatement) stmtNode()        {}

// SelectCase represents a single case in a select statement.
type SelectCase struct {
	Channel string
	Body    []Statement
}

// --- Traits / Interfaces ---

// TraitMethod represents a method signature in a trait declaration.
type TraitMethod struct {
	Name       string
	Params     []TypedParam
	ReturnType string
}

// TraitDeclaration represents a trait (interface) definition.
type TraitDeclaration struct {
	Name    string
	Methods []TraitMethod
	Line    int
}

func (s *TraitDeclaration) nodeType() string { return "Trait" }
func (s *TraitDeclaration) stmtNode()        {}

// --- Generic Type Parameters ---

// TypeParam represents a generic type parameter with optional constraint.
type TypeParam struct {
	Name       string
	Constraint string // optional trait name
}

// --- Destructuring ---

// DestructurePattern is the interface for destructuring patterns.
type DestructurePattern interface {
	patternNode()
}

// ObjectPattern represents {name, age, ...rest} destructuring.
type ObjectPattern struct {
	Fields  []ObjectPatternField
	Rest    string // optional rest variable name
}

func (p *ObjectPattern) patternNode() {}

// ObjectPatternField represents a single field in an object destructuring.
type ObjectPatternField struct {
	Key      string             // the key name
	Nested   DestructurePattern // optional nested pattern (for {user: {name}})
}

// ArrayPattern represents [first, second, ...rest] destructuring.
type ArrayPattern struct {
	Elements []ArrayPatternElement
	Rest     string // optional rest variable name
}

func (p *ArrayPattern) patternNode() {}

// ArrayPatternElement represents a single element in an array destructuring.
type ArrayPatternElement struct {
	Name   string             // variable name (empty if nested)
	Nested DestructurePattern // optional nested pattern
}

// DestructureStatement represents a destructuring assignment.
type DestructureStatement struct {
	Pattern DestructurePattern
	Value   Expression
	Line    int
}

func (s *DestructureStatement) nodeType() string { return "Destructure" }
func (s *DestructureStatement) stmtNode()        {}

// --- Type Check Expression (for type narrowing) ---

// TypeCheckExpr represents "x is text", "x is number" type checks.
type TypeCheckExpr struct {
	Expr     Expression
	TypeName string // "text", "number", "nothing", "boolean"
}

func (e *TypeCheckExpr) nodeType() string { return "TypeCheck" }
func (e *TypeCheckExpr) exprNode()        {}

// --- Generators / Iterators ---

// YieldStatement represents a yield expression inside a generator function.
type YieldStatement struct {
	Value Expression
	Line  int
}

func (s *YieldStatement) nodeType() string { return "Yield" }
func (s *YieldStatement) stmtNode()        {}

// LoopStatement represents an infinite loop (loop:).
type LoopStatement struct {
	Body []Statement
	Line int
}

func (s *LoopStatement) nodeType() string { return "Loop" }
func (s *LoopStatement) stmtNode()        {}

// --- Full-Stack App Blocks ---

// RouteDefinition represents a route handler inside a server or auth block.
type RouteDefinition struct {
	Method string // get, post, put, delete
	Path   string
	Body   []Statement
	Line   int
}

// ModelFieldDef represents a field in a model definition.
type ModelFieldDef struct {
	Name string
	Type string
}

// ModelDef represents a model definition inside a database block.
type ModelDef struct {
	Name   string
	Fields []ModelFieldDef
}

// ServerBlockStatement represents a server: top-level block.
type ServerBlockStatement struct {
	Port       int
	Routes     []RouteDefinition
	WebSockets []WebSocketBlock
	Line       int
}

func (s *ServerBlockStatement) nodeType() string { return "ServerBlock" }
func (s *ServerBlockStatement) stmtNode()        {}

// DatabaseBlockStatement represents a database: top-level block.
type DatabaseBlockStatement struct {
	ConnectString string
	Models        []ModelDef
	Line          int
}

func (s *DatabaseBlockStatement) nodeType() string { return "DatabaseBlock" }
func (s *DatabaseBlockStatement) stmtNode()        {}

// AuthBlockStatement represents an auth: top-level block.
type AuthBlockStatement struct {
	Secret string
	Routes []RouteDefinition
	Line   int
}

func (s *AuthBlockStatement) nodeType() string { return "AuthBlock" }
func (s *AuthBlockStatement) stmtNode()        {}

// RespondStatement represents "respond with <expr> [status <code>]".
type RespondStatement struct {
	Value      Expression
	StatusCode int // default 200
	Line       int
}

func (s *RespondStatement) nodeType() string { return "Respond" }
func (s *RespondStatement) stmtNode()        {}

// --- Type Utilities ---

// TypeAliasStatement represents "type X is Partial of Y" etc.
type TypeAliasStatement struct {
	Name     string
	BaseType string
	Utility  string   // "Partial", "Omit", "Pick", "Record", "Readonly", "Required"
	Args     []string // extra args like field names for Omit/Pick, or key/value types for Record
	Line     int
}

func (s *TypeAliasStatement) nodeType() string { return "TypeAlias" }
func (s *TypeAliasStatement) stmtNode()        {}

// --- Decorators ---

// Decorator represents a @name or @name(args) annotation.
type Decorator struct {
	Name string
	Args []Expression
	Line int
}

// DecoratedFuncDefinition is a function with decorators.
type DecoratedFuncDefinition struct {
	Decorators []Decorator
	Func       *FuncDefinition
	Line       int
}

func (s *DecoratedFuncDefinition) nodeType() string { return "DecoratedFunc" }
func (s *DecoratedFuncDefinition) stmtNode()        {}

// DecoratedRouteDefinition is a route with decorators.
type DecoratedRouteDefinition struct {
	Decorators []Decorator
	Route      RouteDefinition
	Line       int
}

func (s *DecoratedRouteDefinition) nodeType() string { return "DecoratedRoute" }
func (s *DecoratedRouteDefinition) stmtNode()        {}

// --- WebSocket ---

// WebSocketBlock represents a websocket handler block inside a server.
type WebSocketBlock struct {
	Path       string
	OnConnect  []Statement
	OnMessage  []Statement
	OnClose    []Statement
	ConnectVar string
	MessageVar string
	DataVar    string
	CloseVar   string
	Line       int
}

func (s *WebSocketBlock) nodeType() string { return "WebSocket" }
func (s *WebSocketBlock) stmtNode()        {}

// BroadcastStatement represents "broadcast <expr>".
type BroadcastStatement struct {
	Value Expression
	Line  int
}

func (s *BroadcastStatement) nodeType() string { return "Broadcast" }
func (s *BroadcastStatement) stmtNode()        {}

// OnStatement represents "object on "event" with [params]:" event handler syntax.
type OnStatement struct {
	Object Expression
	Event  string
	Params []string
	Body   []Statement
	Line   int
}

func (s *OnStatement) nodeType() string { return "On" }
func (s *OnStatement) stmtNode()        {}
