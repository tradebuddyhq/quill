# Quill

A beginner-friendly language that compiles to JavaScript. English-like syntax with real type checking, pattern matching, and a growing standard library.

[Website](https://quill.tradebuddy.dev) · [Playground](https://quill.tradebuddy.dev/playground) · [Docs](https://quill.tradebuddy.dev/docs/) · [VS Code Extension](https://marketplace.visualstudio.com/items?itemName=tradebuddyhq.quill-lang)

## Quick Start

```bash
# Clone and build
git clone https://github.com/tradebuddyhq/quill.git
cd quill
go build -o quill .

# Run a program
./quill run examples/hello.quill
```

## What it looks like

```
-- Variables with optional type annotations
name is "Sarah"
age is 25
config is {theme: "dark", debug: yes}

-- Print with interpolation
say "Hello, {name}!"

-- Functions with types
to add a as number, b as number -> number:
  give back a + b

say "10 + 20 = {add(10, 20)}"

-- Pattern matching
match age:
  when 18:
    say "Just turned 18!"
  when 25:
    say "Quarter century"
  otherwise:
    say "Age is {age}"

-- Pipe operator
result is "hello world" | upper | trim

-- Lambdas
evens is filter([1, 2, 3, 4], with x: x % 2 is 0)

-- Classes with inheritance
describe Dog extends Animal:
  breed is "mixed"
  to bark:
    say my.name + " says woof!"

-- Error handling
try:
  data is parseJSON(raw)
if it fails err:
  say "Parse error: " + err

-- Algebraic data types
define Color:
  Red
  Green
  Blue

-- Testing
test "math works":
  expect add(2, 3) is 5
  expect add(-1, 1) is 0
```

## Features

**Language**
- **English-like syntax** — `is`, `if/otherwise`, `for each`, `give back`
- **Type annotations with enforcement** — `to add a as number -> number:` checked by `quill check`
- **Pattern matching** — `match`/`when`/`otherwise` for clean branching
- **Algebraic data types** — `define Color: Red, Green, Blue` with variant constructors
- **Pipe operator** — `value | transform | format` for function chaining
- **Lambdas** — `with x: x * 2` arrow-style anonymous functions
- **Classes with inheritance** — `describe Dog extends Animal:` with `my` keyword
- **Try/catch** — `try:` / `if it fails err:` error handling
- **Generics** — `list of number` type annotations
- **Object literals** — `{name: "Alice", age: 30}`
- **Spread operator** — `...items` for expanding lists
- **String interpolation** — `"Hello {name}!"`
- **Break/continue** — loop control flow
- **60+ built-in functions** — `sort()`, `filter()`, `map_list()`, `hash()`, `uuid()`, and more
- **Async/await** — `data is await fetchJSON("url")`
- **Import system** — `use "express" as app` or `from "express" use Router, json`

**Tooling**
- **Type checker** — `quill check` catches type errors before you run
- **Static analyzer** — detects unused variables, infinite loops, bad patterns
- **Code formatter** — `quill fmt` for consistent style
- **Built-in testing** — `test` and `expect` keywords
- **Docs generator** — `quill docs` generates styled HTML documentation
- **Package manager** — `quill add express`, `quill remove express`
- **Interactive REPL** — `quill repl`
- **Friendly error messages** — with source context and hints
- **VS Code extension** — syntax highlighting, snippets, comment toggling

**Build Targets**
- **Node.js** — `quill build file.quill` (default)
- **Browser** — `quill build file.quill --browser` with DOM APIs
- **WASM** — `quill build file.quill --wasm` WASM-ready module
- **Standalone** — `quill build file.quill --standalone` self-executing binary
- **Single binary compiler** — no dependencies, runs on Node.js, Bun, or Deno

## Commands

| Command | Description |
|---------|-------------|
| `quill run file.quill` | Run a program |
| `quill build file.quill` | Compile to JavaScript (Node.js) |
| `quill build file.quill --browser` | Compile for the browser |
| `quill build file.quill --wasm` | Compile as WASM-ready module |
| `quill build file.quill --standalone` | Compile as standalone executable |
| `quill test file.quill` | Run tests |
| `quill fmt file.quill` | Format source code |
| `quill check file.quill` | Type check and lint |
| `quill docs file.quill` | Generate HTML documentation |
| `quill init` | Initialize a new project |
| `quill add package` | Install an npm package |
| `quill remove package` | Remove a package |
| `quill repl` | Start interactive mode |
| `quill help` | Show help |

## Language Reference

### Variables
```
name is "hello"
age is 25
active is yes
items are [1, 2, 3]
```

### Comparisons
```
if x is greater than 10:
if x is less than 5:
if x is equal to 0:
if x is not 0:
if list contains "hello":
```

### Loops
```
for each item in list:
  say item

while count is less than 10:
  count is count + 1
```

### Functions with Type Annotations
```
to greet name:
  say "Hello, {name}!"

to add a as number, b as number -> number:
  give back a + b

-- Lambdas
doubled is map_list(nums, with x: x * 2)
total is reduce(nums, with a, b: a + b)
```

### Pattern Matching
```
match status:
  when "active":
    say "Online"
  when "away":
    say "Be right back"
  otherwise:
    say "Unknown"
```

### Algebraic Data Types
```
define Color:
  Red
  Green
  Blue

define Shape:
  Circle of radius
  Rectangle of width, height
```

### Pipe Operator
```
result is "hello world" | upper | trim
processed is data | filter(with x: x > 0) | sort
```

### Classes with Inheritance
```
describe Animal:
  name is ""
  to speak:
    say my.name + " makes a sound"

describe Dog extends Animal:
  to speak:
    say my.name + " says woof!"

rex is new Dog()
rex.name is "Rex"
rex.speak()
```

### Error Handling
```
try:
  data is parseJSON(raw)
  say data.name
if it fails err:
  say "Error: " + err
```

### Imports
```
use "helpers.quill"                   -- import a Quill file
use "express" as app                  -- import npm package
from "express" use Router, json       -- destructured import
```

### Async/Await
```
data is await fetchJSON("https://api.example.com")
say data
```

### Web Server
```
server is createServer()
server.get("/", "Hello from Quill!")
server.listen(3000)
```

### Standard Library (60+ functions)
```
-- Lists
length(items)              sort(items)
reverse(items)             sum(numbers)
filter(list, fn)           map_list(list, fn)
find(list, fn)             every(list, fn)
some(list, fn)             reduce(list, fn, init)
unique(items)              concat(a, b)
slice(list, start, end)    push(list, item)
flat(items)                indexOf(list, item)

-- Strings
join(items, ", ")          split(text, " ")
upper("hello")             lower("HELLO")
trim("  hi  ")             replace_text(s, old, new)
startsWith(s, prefix)      endsWith(s, suffix)
capitalize("hello")        truncate(s, 20)
padStart(s, 6, "0")        padEnd(s, 10)
words("hello world")       lines(multiline)

-- Math
round(n)    floor(n)    ceil(n)    abs(n)
random()    randomInt(1, 10)    range(1, 5)

-- Objects
merge(a, b)                pick(obj, "name", "age")
omit(obj, "password")      keys(obj)
values(obj)                entries(obj)
hasKey(obj, "name")        deepCopy(obj)

-- Type checking
isText(x)    isNumber(x)    isList(x)
isObject(x)  isNothing(x)   isFunction(x)

-- Files
read("file.txt")           write("out.txt", data)
listFiles("./dir")         deleteFile("old.txt")
copyFile(src, dest)        moveFile(src, dest)
readJSON("data.json")      writeJSON("out.json", data)
fileInfo("file.txt")       fileExists("file.txt")

-- HTTP
fetchJSON("url")           postJSON("url", body)

-- Crypto & Encoding
hash("text")               uuid()
encodeBase64(s)            decodeBase64(s)

-- Date/Time
today()    now()    timestamp()    formatDate(d, "YYYY-MM-DD")

-- System
env("HOME")    platform()    run("ls")    args()
```

### Testing
```
test "my test":
  result is add(2, 3)
  expect result is 5
```

## VS Code Extension

Syntax highlighting, snippets, and language support for `.quill` files.

**Install manually:**
```bash
cp -r vscode-quill ~/.vscode/extensions/quill-lang
```

Then restart VS Code. Open any `.quill` file and you'll get highlighting, comment toggling (`Cmd+/`), and snippets.

## How it Works

Quill compiles to JavaScript through a standard compiler pipeline:

```
.quill source → Lexer → Parser → AST → JavaScript
```

The compiler is written in Go. Generated JS runs on any JavaScript runtime (Node.js, Bun, Deno).

## Project Structure

```
quill/
  main.go              CLI entry point
  lexer/               Tokenizer with indentation tracking
  parser/              Recursive descent parser
  ast/                 AST node types
  codegen/             JavaScript code generator
  typechecker/         Type inference and checking
  formatter/           Code formatter (quill fmt)
  analyzer/            Static analyzer (quill check)
  stdlib/              Standard library (60+ functions, Node + browser)
  repl/                Interactive REPL
  errors/              Friendly error messages
  examples/            Example programs
  site/                Website, docs, and playground
  vscode-quill/        VS Code extension
```

## Contributing

Pull requests welcome! Areas where help is needed:

- More standard library functions
- Better error messages
- More example programs
- Documentation improvements

## License

MIT
