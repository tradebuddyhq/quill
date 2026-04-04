# Quill

Code that reads like English. A beginner-friendly language that compiles to JavaScript. No semicolons, no curly braces, no confusion.

[Website](https://quill.tradebuddy.dev) · [Playground](https://quill.tradebuddy.dev/playground.html) · [Docs](https://quill.tradebuddy.dev/docs/) · [VS Code Extension](https://marketplace.visualstudio.com/items?itemName=tradebuddyhq.quill-lang)

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
-- Variables
name is "Sarah"
age is 25

-- Print with interpolation
say "Hello, {name}!"

-- Conditionals
if age is greater than 18:
  say "You are an adult"
otherwise:
  say "You are young"

-- Functions
to add a b:
  give back a + b

result is add(10, 20)
say "10 + 20 = {result}"

-- Loops
colors are ["red", "green", "blue"]
for each color in colors:
  say "I like {color}"

-- Testing
test "math works":
  expect add(2, 3) is 5
  expect add(-1, 1) is 0
```

## Features

- **English-like syntax** — `is`, `if/otherwise`, `for each`, `give back`
- **String interpolation** — `"Hello {name}!"`
- **Classes** — `describe Dog:` with properties, methods, and `my` keyword
- **60+ built-in functions** — `sort()`, `filter()`, `map_list()`, `length()`, `sum()`, and more
- **Built-in testing** — `test` and `expect` keywords
- **Async/await** — `data is await fetchJSON("url")`
- **Import system** — `use "math.quill"` or `use "express" as app`
- **Web server** — `createServer()` with routing built in
- **Interactive REPL** — `quill repl`
- **Code formatter** — `quill fmt` for consistent style
- **Static analyzer** — `quill check` catches bugs before you run
- **Friendly error messages** — with source context and hints
- **VS Code extension** — syntax highlighting, snippets, comment toggling
- **Compiles to JavaScript** — runs on Node.js, Bun, or Deno
- **Single binary** — no dependencies

## Commands

| Command | Description |
|---------|-------------|
| `quill run file.quill` | Run a program |
| `quill build file.quill` | Compile to JavaScript |
| `quill test file.quill` | Run tests |
| `quill fmt file.quill` | Format source code |
| `quill check file.quill` | Check for common issues |
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

### Functions
```
to greet name:
  say "Hello, {name}!"

to add a b:
  give back a + b
```

### Classes
```
describe Dog:
  name is ""
  sound is ""

  to bark:
    say "{my.name} says {my.sound}!"

rex is new Dog()
rex.name is "Rex"
rex.sound is "Woof"
rex.bark()
```

### Functional Programming
```
to isEven n:
  give back n % 2 is 0

evens is filter(numbers, isEven)
doubled is map_list(numbers, double)
found is find(numbers, bigEnough)
```

### Async/Await
```
data is await fetchJSON("https://api.example.com")
say data
```

### Imports
```
use "helpers.quill"            -- import another Quill file
use "express" as app           -- import npm package
```

### Web Server
```
server is createServer()
server.get("/", "Hello from Quill!")
server.listen(3000)
```

### Standard Library
```
-- Lists
length(items)              sort(items)
reverse(items)             sum(numbers)
filter(list, fn)           map_list(list, fn)
find(list, fn)             every(list, fn)
some(list, fn)             reduce(list, fn, init)
unique(items)              concat(a, b)
slice(list, start, end)    push(list, item)

-- Strings
join(items, ", ")          split(text, " ")
upper("hello")             lower("HELLO")
trim("  hi  ")             replace_text(s, old, new)
startsWith(s, prefix)      endsWith(s, suffix)

-- Math
round(n)    floor(n)    ceil(n)    abs(n)
random()    randomInt(1, 10)    range(1, 5)

-- Files
read("file.txt")           write("out.txt", data)
listFiles("./dir")         deleteFile("old.txt")
copyFile(src, dest)        moveFile(src, dest)
readJSON("data.json")      writeJSON("out.json", data)
fileInfo("file.txt")       fileExists("file.txt")

-- HTTP
fetchJSON("url")           postJSON("url", body)

-- Other
today()    now()    toNumber(x)    toText(x)    typeOf(x)
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
  formatter/           Code formatter (quill fmt)
  analyzer/            Static analyzer (quill check)
  stdlib/              Standard library (60+ built-in functions)
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
- Publish VS Code extension to marketplace
- More example programs
- Documentation improvements

## License

MIT
