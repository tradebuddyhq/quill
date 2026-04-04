# Quill

**A programming language for humans.** Write code that reads like English — no cryptic syntax, no confusing errors, just say what you want.

Built by [Trade Buddy Developers](https://tradebuddy.dev)

[Website](https://quill.tradebuddy.dev) · [Playground](https://quill.tradebuddy.dev/playground.html) · [Docs](https://quill.tradebuddy.dev/docs/)

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
- **40+ built-in functions** — `sort()`, `length()`, `sum()`, `upper()`, `trim()`, and more
- **Built-in testing** — `test` and `expect` keywords
- **Import system** — `use "math.quill"`
- **Interactive REPL** — `quill repl`
- **Friendly error messages** — with source context and hints
- **Compiles to JavaScript** — runs on Node.js, Bun, or Deno
- **Single binary** — no dependencies

## Commands

| Command | Description |
|---------|-------------|
| `quill run file.quill` | Run a program |
| `quill build file.quill` | Compile to JavaScript |
| `quill test file.quill` | Run tests |
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

### Standard Library
```
length(items)          -- length of list/string
sort(items)            -- sort a list
reverse(items)         -- reverse a list
sum(numbers)           -- sum a list of numbers
join(items, ", ")      -- join list into string
split(text, " ")       -- split string into list
upper("hello")         -- "HELLO"
lower("HELLO")         -- "hello"
trim("  hi  ")         -- "hi"
range(1, 5)            -- [1, 2, 3, 4]
random()               -- random number 0-1
today()                -- "2026-04-04"
read("file.txt")       -- read a file
write("out.txt", data) -- write a file
```

### Testing
```
test "my test":
  result is add(2, 3)
  expect result is 5
```

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
  stdlib/              Standard library (40+ built-in functions)
  repl/                Interactive REPL
  errors/              Friendly error messages
  examples/            Example programs
  site/                Website, docs, and playground
```

## Contributing

Pull requests welcome! Areas where help is needed:

- More standard library functions
- Better error messages
- VS Code extension
- More example programs
- Documentation improvements

## License

MIT
