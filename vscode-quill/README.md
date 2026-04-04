# Quill for VS Code

Syntax highlighting, snippets, and language support for [Quill](https://quill.tradebuddy.dev) — code that reads like English.

## Features

- Syntax highlighting for `.quill` files
- Code snippets (type `if`, `for`, `to`, `describe`, etc.)
- Comment toggling (`Cmd+/`)
- Auto-closing brackets and quotes
- Indentation-based folding

## Installation

### From VSIX (recommended)
1. Download the latest `.vsix` from [releases](https://github.com/tradebuddyhq/quill/releases)
2. In VS Code: `Ctrl+Shift+P` → "Install from VSIX" → select the file

### Manual
1. Copy the `vscode-quill` folder to `~/.vscode/extensions/`
2. Restart VS Code

## Screenshot

```quill
-- Quill code in VS Code
name is "World"
say "Hello, {name}!"

to greet person:
  say "Hi {person}!"

greet("Alice")
```

## Learn More

- [Quill Website](https://quill.tradebuddy.dev)
- [Playground](https://quill.tradebuddy.dev/playground.html)
- [Documentation](https://quill.tradebuddy.dev/docs/)
- [GitHub](https://github.com/tradebuddyhq/quill)
