# Quill

Quill is a programming language that reads like English. Write clean, readable code without the ceremony of traditional syntax.

## Install

```bash
npm install -g @tradebuddyhq/quill
```

## Quick Start

```bash
quill init
quill run hello.quill
```

Or create a file manually:

```
-- greeting.quill
name is "World"
say "Hello, {name}!"
```

```bash
quill run greeting.quill
```

## Commands

| Command | Description |
|---------|-------------|
| `quill run <file>` | Compile and run a Quill file |
| `quill build <file>` | Compile to JavaScript |
| `quill check <file>` | Check for errors |
| `quill repl` | Interactive REPL |
| `quill init` | Create starter files |

## Links

- [Documentation](https://quill.tradebuddy.dev)
- [Playground](https://quill.tradebuddy.dev/playground)
- [GitHub](https://github.com/tradebuddyhq/quill)
- [Discord](https://discord.gg/quill)
