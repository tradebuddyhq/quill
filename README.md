# Quill

A beginner-friendly language that compiles to JavaScript. English-like syntax with real type checking, pattern matching, and a growing standard library.

[Website](https://quill.tradebuddy.dev) · [Playground](https://quill.tradebuddy.dev/playground) · [Docs](https://quill.tradebuddy.dev/docs/) · [VS Code Extension](https://marketplace.visualstudio.com/items?itemName=tradebuddyhq.quill-lang)

## Community

Join our [Discord](https://discord.gg/9rRyGRrh8E) for help, questions, and discussion. Visit [tradebuddy.dev](https://tradebuddy.dev) for more about the project.

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

```
-- Reactive web components
component Counter:
  state count is 0
  to increment:
    count is count + 1
  to render:
    div:
      h1: "Count: {count}"
      button onClick increment: "+1"

mount Counter to "#app"
```

## Features

**Language**
- **English-like syntax** — `is`, `if/otherwise`, `for each`, `give back`
- **Type annotations with enforcement** — `to add a as number -> number:` checked by `quill check`
- **Pattern matching** — `match`/`when`/`otherwise` for clean branching, type-based matching, guard clauses, exhaustive checking
- **Algebraic data types** — `define Color: Red, Green, Blue` with variant constructors
- **Traits** — `describe trait Printable:` with method signatures for interface contracts
- **Generics with constraints** — `to sort items as list of T where T is Comparable`
- **Destructuring** — `{name, age} is person`, `[first, ...rest] are items`, nested patterns
- **Type narrowing** — `if x is text:` and the compiler knows `x` is text inside the block
- **Pipe operator** — `value | transform | format` for function chaining
- **Lambdas** — `with x: x * 2` arrow-style anonymous functions
- **Classes with inheritance** — `describe Dog extends Animal:` with `my` keyword
- **Try/catch** — `try:` / `if it fails err:` error handling
- **Iterators & generators** — `yield`, `loop:` infinite loops, lazy evaluation chains
- **Lazy evaluation** — `range(1, 1000000) | filter | map_list | take 10 | collect`
- **Object literals** — `{name: "Alice", age: 30}`
- **Spread operator** — `...items` for expanding lists
- **String interpolation** — `"Hello {name}!"`
- **Break/continue** — loop control flow
- **60+ built-in functions** — `sort()`, `filter()`, `map_list()`, `hash()`, `uuid()`, and more
- **Async/await** — `data is await fetchJSON("url")`
- **Import system** — `use "express" as app` or `from "express" use Router, json`
- **Union types** — `number | text` for values that can be multiple types
- **Nullable types** — `?number` shorthand for `number | nothing`
- **Reactive web framework** — Svelte-like `component`/`state`/`mount` with virtual DOM
- **Full-stack in one file** — `server:`, `database:`, `component:` blocks in a single `.quill` file

**Tooling**
- **Type checker** — `quill check` catches type errors before you run
- **Static analyzer** — detects unused variables, infinite loops, bad patterns
- **Code formatter** — `quill fmt` for consistent style
- **Built-in testing** — `test` and `expect` keywords, `quill test --coverage` for coverage reports
- **Docs generator** — `quill docs` generates styled HTML documentation
- **Package manager** — `quill add express`, `quill remove express`
- **Interactive REPL** — `quill repl`
- **Dev server** — `quill serve` with hot reload and file-based routing
- **Profiler** — `quill profile app.quill` for function timing reports
- **Workspaces** — monorepo support via `[workspace]` in quill.toml
- **Migration tool** — `quill fix --from v0.1 --to v0.3` for automated code migration
- **Friendly error messages** — with source context and hints
- **VS Code extension** — syntax highlighting, snippets, comment toggling
- **LSP server** — `quill lsp` for editor integration (diagnostics, hover, autocomplete)
- **Source maps** — `.map` files generated alongside JS for debugging
- **Package registry** — `quill publish`, `quill search`, `quill install` with version resolution

**Web Framework (SvelteKit-level)**
- **File-based routing** — `pages/about.quill` maps to `/about`, `pages/blog/[id].quill` maps to `/blog/:id`
- **SSR with hydration** — `to load request:` runs on server, `to render data:` hydrates on client
- **Scoped styles** — `style:` block in components with auto-hashed CSS scoping
- **Client-side routing** — `link to="/about" "About Us"` with pushState navigation
- **Head management** — `head:` block for title/meta tags
- **Form actions** — `form action=handler:` for server-side form handling

**Build Targets**
- **Node.js** — `quill build file.quill` (default)
- **Browser** — `quill build file.quill --browser` with DOM APIs
- **WASM** — `quill build file.quill --wasm` WASM-ready module
- **Standalone** — `quill build file.quill --standalone` self-executing binary
- **LLVM/Native** — `quill build file.quill --llvm` generates LLVM IR, compiles to native binary
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
| `quill build file.quill --llvm` | Compile to LLVM IR / native binary |
| `quill serve` | Start dev server with hot reload |
| `quill profile file.quill` | Run profiler with function timing report |
| `quill fix --from v0.1 --to v0.3` | Automated code migration between versions |
| `quill test file.quill --coverage` | Run tests with coverage report |
| `quill test --coverage-html` | Generate HTML coverage report |
| `quill test --coverage-min 80` | Fail if coverage is below threshold |
| `quill debug file.quill` | Launch step-through debugger |
| `quill lsp` | Start LSP server for editors |
| `quill publish` | Publish package to registry |
| `quill search query` | Search the package registry |
| `quill install` | Install all dependencies |
| `quill bump patch` | Bump version in quill.json |
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

### Reactive Web Components
```
component Counter:
  state count is 0
  to increment:
    count is count + 1
  to render:
    div:
      h1: "Count: {count}"
      button onClick increment: "+1"

mount Counter to "#app"
```

### Union & Nullable Types
```
to process value as number | text -> text:
  give back toText(value)

to findUser id as number -> ?User:
  give back nothing
```

### Testing
```
test "my test":
  result is add(2, 3)
  expect result is 5
```

### Traits
```
describe trait Printable:
  to display -> text

describe trait Serializable:
  to toJSON -> text
  to fromJSON data as text

describe User implements Printable, Serializable:
  name is ""
  to display -> text:
    give back my.name
  to toJSON -> text:
    give back "{\"name\": \"{my.name}\"}"
  to fromJSON data as text:
    say "parsing..."
```

### Generics with Constraints
```
to sort items as list of T where T is Comparable:
  give back items | sort

to first items as list of T -> T:
  give back items[0]

to merge a as list of T, b as list of T -> list of T:
  give back concat(a, b)
```

### Destructuring
```
-- Object destructuring
{name, age} is person
{name, ...rest} is config

-- List destructuring
[first, second] are items
[head, ...tail] are numbers

-- Nested destructuring
{address: {city, zip}} is user

-- In function parameters
to greet {name, age}:
  say "{name} is {age} years old"
```

### Type Narrowing
```
to process value as number | text:
  if value is text:
    -- compiler knows value is text here
    say upper(value)
  otherwise:
    -- compiler knows value is number here
    say value + 1
```

### Iterators & Generators
```
-- Generator function (compiles to function*)
to fibonacci:
  a is 0
  b is 1
  loop:
    yield a
    temp is a
    a is b
    b is temp + b

-- Lazy evaluation chains
result is range(1, 1000000) | filter with n: n % 2 is 0 | map_list with n: n * n | take 10 | collect

-- Full lazy API
range(1, 100)
  | filter with n: n % 3 is 0
  | skip 5
  | takeWhile with n: n is less than 50
  | enumerate
  | collect
```

### Pattern Matching (Complete)
```
-- Type-based matching
match value:
  when text t:
    say "Got text: {t}"
  when number n:
    say "Got number: {n}"
  when list l:
    say "Got a list with {length(l)} items"
  when nothing:
    say "Got nothing"

-- Guard clauses
match age:
  when n if n is less than 18:
    say "Minor"
  when n if n is less than 65:
    say "Adult"
  otherwise:
    say "Senior"

-- Exhaustive checking on algebraic types
define Shape:
  Circle of radius
  Rectangle of width, height
  Triangle of base, height

match shape:
  when Circle r:
    say "Circle with radius {r}"
  when Rectangle w, h:
    say "Rectangle {w}x{h}"
  -- compiler error: missing Triangle variant!
```

### Web Framework
```
-- File-based routing: pages/about.quill -> /about
-- pages/blog/[id].quill -> /blog/:id

-- Server-side rendering with hydration
to load request:
  users is await DB.find({})
  give back {users: users}

to render data:
  h1 "Users"
  for each user in data.users:
    p "{user.name}"

-- Scoped styles
style:
  h1:
    color is "blue"
    font-size is "2rem"

-- Head management
head:
  title "My App"
  meta name="description" content="A Quill app"

-- Client-side routing
link to="/about" "About Us"
link to="/blog/{post.id}" "{post.title}"

-- Form actions
form action=handleSubmit:
  input bind:value=email placeholder="Email"
  button "Submit"
```

### Full-Stack in One File

Write your entire app -- server, database, auth, and UI -- in a single `.quill` file. No config files, no `package.json`, no `node_modules`, no webpack. Just `quill run app.quill`.

```
server:
    port is 3000
    route get "/api/users":
        users is await DB.find({})
        respond with users

database:
    connect "sqlite://app.db"
    model User:
        name as text
        email as text

component App:
    state:
        users is []
    to render:
        h1 "My App"
        for each user in users:
            p "{user.name}"

mount App to "#app"
```

Run it with one command:
```bash
quill run app.quill
```

That is it. Zero config. One file. Full-stack application.

## Ecosystem (Standard Library)

Quill ships with high-level libraries for common backend tasks, all with English-like syntax.

### Auth

Built-in authentication with hashing, JWT tokens, and sessions.

```
use "auth"

-- Hash and verify passwords
hashed is Auth.hash("my-password")
valid is Auth.verify("my-password", hashed)

-- JWT tokens
token is Auth.createToken({userId: 42}, "secret", {expiresIn: "1h"})
payload is Auth.verifyToken(token, "secret")

-- Session middleware
app.use(Auth.session({secret: "keyboard-cat"}))
```

### ORM

Database access with models, query builder, migrations, and transactions.

```
use "db"

-- Connect and define a model
DB.connect("postgres://localhost/myapp")

User is DB.model("users", {
  name: "text",
  email: "text",
  age: "number"
})

-- CRUD operations
User.create({name: "Alice", email: "alice@example.com", age: 30})
users is User.find({age: {gte: 18}})
User.update({email: "alice@example.com"}, {age: 31})
User.delete({name: "Alice"})

-- Transactions
DB.transaction(with tx:
  tx.run("INSERT INTO orders ...")
  tx.run("UPDATE inventory ...")
)
```

### Validation

Schema-based validation with built-in rules and custom validators.

```
use "validate"

userSchema is Validate.schema({
  name: {required: yes, min: 2},
  email: {required: yes, email: yes},
  age: {min: 0, max: 150},
  role: {pattern: "^(admin|user)$"},
  tags: {arrayOf: {min: 1}},
  address: {
    street: {required: yes},
    zip: {pattern: "^[0-9]{5}$"}
  }
})

result is userSchema.validate(input)
if result.valid:
  say "All good"
otherwise:
  say result.errors
```

### Logging

Structured logging with levels, JSON output, colored console, and child loggers.

```
use "log"

logger is Log.create({level: "info", json: yes})

logger.debug("Startup details")
logger.info("Server started", {port: 3000})
logger.warn("Disk space low")
logger.error("Connection failed", {host: "db.local"})
logger.fatal("Out of memory")

-- Child loggers inherit config
reqLogger is logger.child({requestId: "abc-123"})
reqLogger.info("Handling request")

-- Use as middleware
app.use(logger.middleware())
```

## Concurrency

Quill has built-in concurrency primitives: tasks, channels, parallel blocks, race blocks, and select.

```
-- Spawn a background task
spawn task:
  data is await fetchJSON("https://api.example.com")
  say data

-- Run tasks in parallel (all must complete)
parallel:
  users is await fetchJSON("/users")
  posts is await fetchJSON("/posts")
  stats is await fetchJSON("/stats")

-- Race: first to finish wins
race:
  result is await fetchJSON("/fast-api")
  result is await fetchJSON("/slow-api")

-- Channels with buffered capacity
ch is channel(5)
send ch, "hello"
msg is receive ch

-- Select with timeout
select:
  when receive ch1:
    say "Got from ch1: {it}"
  when receive ch2:
    say "Got from ch2: {it}"
  after 5000:
    say "Timed out"
```

## Debugger

Launch a step-through debugger for any Quill program:

```bash
quill debug examples/hello.quill
```

Features:
- **Breakpoints** -- set on any Quill source line
- **Step controls** -- step over, step into, step out
- **Variable inspection** -- view locals and globals at any point
- **Call stacks** -- see the full call chain
- **Source maps** -- you debug Quill source, not compiled JS

REPL commands inside the debugger:

| Command | Description |
|---------|-------------|
| `break <line>` | Set a breakpoint |
| `continue` / `c` | Continue execution |
| `step` / `s` | Step to next line |
| `into` / `i` | Step into function call |
| `out` / `o` | Step out of current function |
| `print <expr>` | Evaluate and print expression |
| `locals` | Show all local variables |
| `stack` | Show call stack |
| `list` | Show source around current line |
| `quit` / `q` | Exit debugger |

## Production Tooling

### Test Coverage
```bash
quill test --coverage                # Print coverage summary
quill test --coverage-html           # Generate HTML coverage report
quill test --coverage-min 80         # Fail if coverage drops below 80%
```

### Profiler
```bash
quill profile app.quill
```
Outputs a function-level timing report showing where your program spends its time.

### Workspaces
Monorepo support via `[workspace]` in quill.toml:
```toml
[workspace]
members = ["packages/core", "packages/web", "packages/cli"]
```
Run commands across all workspace members: `quill test --workspace`, `quill build --workspace`.

### Migration Tool
Automated code migration between Quill versions:
```bash
quill fix --from v0.1 --to v0.3
```
Rewrites deprecated syntax, renames changed functions, and updates import paths automatically.

### Dev Server
```bash
quill serve
```
Starts a development server with hot reload, file-based routing, and SSR support. Perfect for web projects.

## Stability

### Rust-Style Error Messages

The parser recovers from errors and keeps going, collecting all problems in a single pass instead of stopping at the first one. Error messages include source context, underlines, and helpful hints.

```
error[E001]: type mismatch
  --> app.quill:12:5
   |
12 |   count is "hello"
   |            ^^^^^^^ expected number, found text
   |
   = hint: did you mean to use toNumber("hello")?
```

### Project Configuration

`quill.toml` defines project settings, dependencies, and build options:

```toml
[project]
name = "my-app"
version = "0.1.0"
entry = "src/main.quill"

[dependencies]
express = "^4.18.0"

[build]
target = "node"
minify = true
```

### Project Scaffolding

```bash
quill init my-app
```

Creates a new project with `quill.toml`, a `src/` directory, example files, and a `.gitignore`.

### GitHub Actions CI

Quill projects include a CI template that runs `quill check`, `quill test`, and `quill build` on every push and pull request.

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
.quill source → Lexer → Parser → AST → JavaScript (default)
                                      → LLVM IR → Native binary
                                      → Browser JS + Virtual DOM
```

The compiler is written in Go. Generated JS runs on any JavaScript runtime (Node.js, Bun, Deno). The LLVM backend produces native binaries, and the browser target includes a virtual DOM for reactive components.

## Project Structure

```
quill/
  main.go              CLI entry point
  lexer/               Tokenizer with indentation tracking
  parser/              Recursive descent parser (with error recovery)
  ast/                 AST node types
  codegen/             JavaScript code generator
  typechecker/         Type inference and checking
  formatter/           Code formatter (quill fmt)
  analyzer/            Static analyzer (quill check)
  stdlib/              Standard library (60+ functions, Node + browser)
  lsp/                 Language Server Protocol server
  debugger/            Step-through debugger with source map support
  registry/            Package registry client & resolver
  repl/                Interactive REPL
  config/              Project configuration (quill.toml) handling
  errors/              Rust-style error messages with hints
  server/              Dev server, file-based routing, SSR
  tools/               Profiler, coverage, migration tool, workspaces
  tests/               Test suite and test runner
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
