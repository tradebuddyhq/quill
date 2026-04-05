#!/usr/bin/env node
'use strict';

const fs = require('fs');
const path = require('path');
const vm = require('vm');
const { compile } = require('../lib/compiler');
const runtime = require('../lib/runtime');

// ---- ANSI Colors ----
const RED = '\x1b[31m';
const GREEN = '\x1b[32m';
const YELLOW = '\x1b[33m';
const CYAN = '\x1b[36m';
const DIM = '\x1b[2m';
const BOLD = '\x1b[1m';
const RESET = '\x1b[0m';

const VERSION = require('../package.json').version;

// ---- Helpers ----
function error(msg) {
  console.error(`${RED}Error:${RESET} ${msg}`);
}

function success(msg) {
  console.log(`${GREEN}${msg}${RESET}`);
}

function readQuillFile(filePath) {
  const resolved = path.resolve(filePath);
  if (!fs.existsSync(resolved)) {
    error(`File not found: ${filePath}`);
    process.exit(1);
  }
  if (!resolved.endsWith('.quill')) {
    error(`Expected a .quill file, got: ${filePath}`);
    process.exit(1);
  }
  return fs.readFileSync(resolved, 'utf-8');
}

function buildRuntime() {
  // Build a string that declares all runtime functions in scope
  const fns = Object.keys(runtime);
  const lines = fns.map(name => `const ${name} = __runtime.${name};`);
  return lines.join('\n') + '\n';
}

// ---- Commands ----

function cmdRun(filePath) {
  const source = readQuillFile(filePath);
  const { js, errors } = compile(source);

  if (errors.length > 0) {
    for (const err of errors) {
      error(err);
    }
    process.exit(1);
  }

  const runtimePreamble = buildRuntime();
  const fullCode = runtimePreamble + js;

  try {
    const script = new vm.Script(fullCode, { filename: filePath });
    const context = vm.createContext({
      __runtime: runtime,
      console,
      setTimeout,
      setInterval,
      clearTimeout,
      clearInterval,
      fetch: typeof fetch !== 'undefined' ? fetch : undefined,
      JSON,
      Math,
      Array,
      Object,
      String,
      Number,
      Boolean,
      Date,
      RegExp,
      Error,
      Map,
      Set,
      Promise,
      require,
      process,
    });
    script.runInContext(context);
  } catch (e) {
    error(e.message);
    process.exit(1);
  }
}

function cmdBuild(filePath) {
  const source = readQuillFile(filePath);
  const { js, errors } = compile(source);

  if (errors.length > 0) {
    for (const err of errors) {
      error(err);
    }
    process.exit(1);
  }

  const runtimePreamble = buildRuntime();
  const fullCode = `// Compiled from ${path.basename(filePath)} by Quill ${VERSION}\n` +
    `const __runtime = require('@tradebuddyhq/quill/lib/runtime');\n` +
    runtimePreamble + js + '\n';

  const outPath = filePath.replace(/\.quill$/, '.js');
  fs.writeFileSync(outPath, fullCode, 'utf-8');
  success(`Built: ${outPath}`);
}

function cmdCheck(filePath) {
  const source = readQuillFile(filePath);
  const { js, errors } = compile(source);

  if (errors.length > 0) {
    console.log(`${RED}${BOLD}Found ${errors.length} error(s):${RESET}`);
    for (const err of errors) {
      error(err);
    }
    process.exit(1);
  }

  success(`No errors in ${filePath}`);
}

function cmdRepl() {
  const readline = require('readline');
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    prompt: `${CYAN}quill>${RESET} `,
  });

  console.log(`${BOLD}Quill REPL v${VERSION}${RESET} ${DIM}(type "exit" to quit)${RESET}`);

  // Persistent context for the REPL session
  const context = vm.createContext({
    __runtime: runtime,
    console,
    setTimeout,
    setInterval,
    clearTimeout,
    clearInterval,
    JSON,
    Math,
    Array,
    Object,
    String,
    Number,
    Boolean,
    Date,
    RegExp,
    Error,
    Map,
    Set,
    Promise,
  });

  // Inject runtime into context
  for (const [name, fn] of Object.entries(runtime)) {
    context[name] = fn;
  }

  rl.prompt();

  rl.on('line', (line) => {
    const trimmed = line.trim();
    if (trimmed === 'exit' || trimmed === 'quit') {
      rl.close();
      return;
    }
    if (trimmed === '') {
      rl.prompt();
      return;
    }

    const { js, errors } = compile(trimmed);

    if (errors.length > 0) {
      for (const err of errors) {
        error(err);
      }
    } else {
      try {
        const result = vm.runInContext(js, context, { filename: '<repl>' });
        if (result !== undefined) {
          console.log(result);
        }
      } catch (e) {
        error(e.message);
      }
    }

    rl.prompt();
  });

  rl.on('close', () => {
    console.log(`\n${DIM}Bye!${RESET}`);
    process.exit(0);
  });
}

function cmdInit() {
  const helloPath = path.resolve('hello.quill');
  const configPath = path.resolve('quill.toml');

  if (fs.existsSync(helloPath)) {
    console.log(`${YELLOW}hello.quill already exists, skipping.${RESET}`);
  } else {
    fs.writeFileSync(helloPath, `-- My first Quill program\nsay "Hello, World!"\n`, 'utf-8');
  }

  if (fs.existsSync(configPath)) {
    console.log(`${YELLOW}quill.toml already exists, skipping.${RESET}`);
  } else {
    fs.writeFileSync(configPath, `[project]\nname = "my-quill-app"\nversion = "0.1.0"\n\n[build]\noutput = "dist"\n`, 'utf-8');
  }

  success('Created hello.quill \u2014 run it with: quill run hello.quill');
}

function showHelp() {
  console.log(`
${BOLD}Quill${RESET} v${VERSION} ${DIM}\u2014 a language that reads like English${RESET}

${BOLD}Usage:${RESET}
  quill run <file.quill>     Compile and run a Quill file
  quill build <file.quill>   Compile to JavaScript
  quill check <file.quill>   Check for errors without running
  quill repl                 Start interactive REPL
  quill init                 Create starter files

${BOLD}Options:${RESET}
  --version, -v              Show version
  --help, -h                 Show this help

${BOLD}Examples:${RESET}
  ${DIM}$ quill init${RESET}
  ${DIM}$ quill run hello.quill${RESET}
  ${DIM}$ quill build app.quill${RESET}

${DIM}Docs: https://quill.tradebuddy.dev${RESET}
`);
}

// ---- Main ----
const args = process.argv.slice(2);
const cmd = args[0];

if (!cmd || cmd === '--help' || cmd === '-h') {
  showHelp();
  process.exit(0);
}

if (cmd === '--version' || cmd === '-v') {
  console.log(`quill ${VERSION}`);
  process.exit(0);
}

switch (cmd) {
  case 'run':
    if (!args[1]) { error('Missing file argument. Usage: quill run <file.quill>'); process.exit(1); }
    cmdRun(args[1]);
    break;
  case 'build':
    if (!args[1]) { error('Missing file argument. Usage: quill build <file.quill>'); process.exit(1); }
    cmdBuild(args[1]);
    break;
  case 'check':
    if (!args[1]) { error('Missing file argument. Usage: quill check <file.quill>'); process.exit(1); }
    cmdCheck(args[1]);
    break;
  case 'repl':
    cmdRepl();
    break;
  case 'init':
    cmdInit();
    break;
  default:
    error(`Unknown command: ${cmd}`);
    showHelp();
    process.exit(1);
}
