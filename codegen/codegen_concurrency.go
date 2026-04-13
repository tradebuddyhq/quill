package codegen

import (
	"fmt"
	"quill/ast"
	"strings"
)

func (g *Generator) genTrait(s *ast.TraitDeclaration, prefix string) string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%s// Trait: %s\n", prefix, s.Name))
	for _, m := range s.Methods {
		params := make([]string, len(m.Params))
		for i, p := range m.Params {
			if p.TypeHint != "" {
				params[i] = fmt.Sprintf("%s: %s", p.Name, p.TypeHint)
			} else {
				params[i] = p.Name
			}
		}
		ret := ""
		if m.ReturnType != "" {
			ret = " -> " + m.ReturnType
		}
		out.WriteString(fmt.Sprintf("%s//   %s(%s)%s\n", prefix, m.Name, strings.Join(params, ", "), ret))
	}
	// Generate runtime __implements checker
	out.WriteString(fmt.Sprintf("%sfunction __implements_%s(obj) {\n", prefix, s.Name))
	for _, m := range s.Methods {
		out.WriteString(fmt.Sprintf("%s  if (typeof obj.%s !== 'function') return false;\n", prefix, m.Name))
	}
	out.WriteString(fmt.Sprintf("%s  return true;\n", prefix))
	out.WriteString(fmt.Sprintf("%s}\n", prefix))
	return out.String()
}

// --- Destructuring code generation ---

func (g *Generator) genDestructure(s *ast.DestructureStatement, prefix string) string {
	patternCode := g.genPattern(s.Pattern)
	valueCode := g.genExpr(s.Value)
	return fmt.Sprintf("%sconst %s = %s;", prefix, patternCode, valueCode)
}

func (g *Generator) genPattern(p ast.DestructurePattern) string {
	switch pat := p.(type) {
	case *ast.ObjectPattern:
		parts := []string{}
		for _, f := range pat.Fields {
			if f.Nested != nil {
				parts = append(parts, fmt.Sprintf("%s: %s", f.Key, g.genPattern(f.Nested)))
			} else {
				parts = append(parts, f.Key)
			}
		}
		if pat.Rest != "" {
			parts = append(parts, "..."+pat.Rest)
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *ast.ArrayPattern:
		parts := []string{}
		for _, e := range pat.Elements {
			if e.Nested != nil {
				parts = append(parts, g.genPattern(e.Nested))
			} else {
				parts = append(parts, e.Name)
			}
		}
		if pat.Rest != "" {
			parts = append(parts, "..."+pat.Rest)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return "/* unknown pattern */"
	}
}

// --- Concurrency code generation ---

func (g *Generator) genSpawn(s *ast.SpawnStatement, prefix string) string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%sconst __abort_%s = new AbortController();\n", prefix, s.Name))
	g.indent++
	body := g.genBlock(s.Body)
	g.indent--
	out.WriteString(fmt.Sprintf("%sconst __task_%s = (async (signal) => {\n%s%s})(__abort_%s.signal);", prefix, s.Name, body, prefix, s.Name))
	return out.String()
}

func (g *Generator) genParallel(s *ast.ParallelBlock, prefix string) string {
	var out strings.Builder
	// Collect variable names and their value expressions
	var names []string
	var exprs []string
	for _, task := range s.Tasks {
		if assign, ok := task.(*ast.AssignStatement); ok {
			names = append(names, assign.Name)
			exprs = append(exprs, g.genExpr(assign.Value))
			g.declared[assign.Name] = true
		}
	}
	promiseMethod := "Promise.all"
	if s.IsSettled {
		promiseMethod = "Promise.allSettled"
	}
	out.WriteString(fmt.Sprintf("%slet [%s] = await %s([\n", prefix, strings.Join(names, ", "), promiseMethod))
	for i, expr := range exprs {
		comma := ","
		if i == len(exprs)-1 {
			comma = ""
		}
		out.WriteString(fmt.Sprintf("%s  %s%s\n", prefix, expr, comma))
	}
	out.WriteString(fmt.Sprintf("%s]);", prefix))
	return out.String()
}

func (g *Generator) genRaceBlock(s *ast.RaceBlock, prefix string) string {
	var out strings.Builder
	var exprs []string
	for _, task := range s.Tasks {
		if assign, ok := task.(*ast.AssignStatement); ok {
			exprs = append(exprs, g.genExpr(assign.Value))
			g.declared[assign.Name] = true
		}
	}
	out.WriteString(fmt.Sprintf("%slet __race_result = await Promise.race([\n", prefix))
	for i, expr := range exprs {
		comma := ","
		if i == len(exprs)-1 {
			comma = ""
		}
		out.WriteString(fmt.Sprintf("%s  %s%s\n", prefix, expr, comma))
	}
	out.WriteString(fmt.Sprintf("%s]);", prefix))
	return out.String()
}

func (g *Generator) genChannelStmt(s *ast.ChannelStatement, prefix string) string {
	if s.BufferSize != nil {
		return fmt.Sprintf("%slet %s = new __QuillChannel(%s);", prefix, s.Name, g.genExpr(s.BufferSize))
	}
	return fmt.Sprintf("%slet %s = new __QuillChannel();", prefix, s.Name)
}

func (g *Generator) genSelectStmt(s *ast.SelectStatement, prefix string) string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%sawait Promise.race([\n", prefix))
	for _, c := range s.Cases {
		g.indent++
		body := g.genBlock(c.Body)
		g.indent--
		out.WriteString(fmt.Sprintf("%s  %s.receive().then(val => {\n%s%s  }),\n", prefix, c.Channel, body, prefix))
	}
	if s.AfterMs != nil {
		g.indent++
		afterBody := g.genBlock(s.AfterBody)
		g.indent--
		out.WriteString(fmt.Sprintf("%s  new Promise(resolve => setTimeout(() => {\n%s%s    resolve();\n%s  }, %s)),\n",
			prefix, afterBody, prefix, prefix, g.genExpr(s.AfterMs)))
	}
	out.WriteString(fmt.Sprintf("%s]);", prefix))
	return out.String()
}

// channelRuntime is the __QuillChannel class injected when channel/send/receive/select is used.
const channelRuntime = `class __QuillChannel {
  constructor(bufferSize = 0) {
    this.buffer = [];
    this.bufferSize = bufferSize;
    this.waiting = [];
    this.senders = [];
  }
  send(value) {
    return new Promise(resolve => {
      if (this.waiting.length > 0) {
        this.waiting.shift()(value);
        resolve();
      } else if (this.buffer.length < this.bufferSize) {
        this.buffer.push(value);
        resolve();
      } else {
        this.senders.push({ value, resolve });
      }
    });
  }
  receive() {
    return new Promise(resolve => {
      if (this.buffer.length > 0) {
        resolve(this.buffer.shift());
        if (this.senders.length > 0) {
          let s = this.senders.shift();
          this.buffer.push(s.value);
          s.resolve();
        }
      } else if (this.senders.length > 0) {
        let s = this.senders.shift();
        resolve(s.value);
        s.resolve();
      } else {
        this.waiting.push(resolve);
      }
    });
  }
}
`

// resultRuntime provides Success/Error Result type helpers and propagation.
const resultRuntime = `function Success(value) { return { __isOk: true, value }; }
function __QuillError(message) { return { __isErr: true, message }; }
function __propagate(val) {
  if (val instanceof Error || val?.__isErr) return val;
  return val?.__isOk ? val.value : val;
}
function __tryResult(fn) {
  try { const r = fn; if (r?.__isErr) return r; return r; } catch(e) { return __QuillError(e.message); }
}
`

// programContainsAwait checks if any top-level statement in the program contains an await expression.
func (g *Generator) programContainsAwait(program *ast.Program) bool {
	return g.bodyContainsAwait(program.Statements)
}

// hasAsyncConstructs checks if the program uses any concurrency constructs that need async IIFE.
func (g *Generator) hasAsyncConstructs(program *ast.Program) bool {
	return g.stmtsHaveAsyncConstructs(program.Statements)
}

// stmtsHaveAsyncConstructs recursively checks statements for async constructs.
func (g *Generator) stmtsHaveAsyncConstructs(stmts []ast.Statement) bool {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.SpawnStatement, *ast.ParallelBlock, *ast.RaceBlock,
			*ast.ChannelStatement, *ast.SendStatement, *ast.SelectStatement,
			*ast.CancelStatement, *ast.StreamStatement:
			return true
		case *ast.IfStatement:
			if g.stmtsHaveAsyncConstructs(s.Body) {
				return true
			}
			for _, elif := range s.ElseIfs {
				if g.stmtsHaveAsyncConstructs(elif.Body) {
					return true
				}
			}
			if g.stmtsHaveAsyncConstructs(s.Else) {
				return true
			}
		case *ast.WhileStatement:
			if g.stmtsHaveAsyncConstructs(s.Body) {
				return true
			}
		case *ast.ForEachStatement:
			if g.stmtsHaveAsyncConstructs(s.Body) {
				return true
			}
		case *ast.LoopStatement:
			if g.stmtsHaveAsyncConstructs(s.Body) {
				return true
			}
		case *ast.TryCatchStatement:
			if g.stmtsHaveAsyncConstructs(s.TryBody) {
				return true
			}
			if g.stmtsHaveAsyncConstructs(s.CatchBody) {
				return true
			}
		}
	}
	return g.needsAIRuntime
}

// hasConcurrency checks if the program uses channel/send/receive/select features.
func (g *Generator) hasConcurrency(program *ast.Program) bool {
	for _, stmt := range program.Statements {
		switch stmt.(type) {
		case *ast.ChannelStatement, *ast.SendStatement, *ast.SelectStatement:
			return true
		}
	}
	return false
}

// embedColorMap maps color names to hex values for the embed builder.
var embedColorMap = map[string]string{
	"red":    "0xFF0000",
	"green":  "0x1EB969",
	"blue":   "0x3498DB",
	"yellow": "0xF1C40F",
	"purple": "0x9B59B6",
	"orange": "0xE67E22",
	"white":  "0xFFFFFF",
	"black":  "0x000000",
}

