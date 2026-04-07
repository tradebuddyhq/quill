#!/usr/bin/env node
'use strict';

// Try native Go binary first, fall back to JS compiler
const _nativePath = require('path').join(__dirname, 'quill-native');
if (require('fs').existsSync(_nativePath)) {
  try {
    const { execFileSync } = require('child_process');
    execFileSync(_nativePath, process.argv.slice(2), { stdio: 'inherit' });
    process.exit(0);
  } catch (e) {
    if (e.status != null) process.exit(e.status);
    // If exec failed entirely (e.g. corrupt binary), fall through to JS compiler
  }
}

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

// ---- Scaffold: Discord Bot ----
function cmdDiscord() {
  const { execSync } = require('child_process');
  const projectName = process.argv[3] || 'my-discord-bot';

  if (fs.existsSync(projectName)) {
    error(`Directory "${projectName}" already exists.`);
    process.exit(1);
  }

  fs.mkdirSync(projectName, { recursive: true });

  // package.json
  fs.writeFileSync(path.join(projectName, 'package.json'), JSON.stringify({
    name: projectName,
    version: '1.0.0',
    description: 'A Discord bot built with Quill',
    main: 'bot.js',
    scripts: { start: 'node bot.js', dev: 'quill run bot.quill' },
    dependencies: { 'discord.js': '^14.14.1' }
  }, null, 2) + '\n');

  // bot.quill
  fs.writeFileSync(path.join(projectName, 'bot.quill'), `-- Discord Bot built with Quill

use "discord.js" as Discord

bot is Discord.bot(env("DISCORD_TOKEN"))

command "ping" described "Check if bot is alive":
  reply "Pong!"

command "help" described "Learn about this bot":
  reply embed "My Bot":
    color green
    description "A Discord bot built with Quill"
    field "Ping" "Check if the bot is alive"
    field "Hello" "Get a greeting"

command "hello" with user described "Greet someone":
  reply "Hello, {user}!"
`);

  // .env.example
  fs.writeFileSync(path.join(projectName, '.env.example'), `DISCORD_TOKEN=your-bot-token-here
`);

  // .gitignore
  fs.writeFileSync(path.join(projectName, '.gitignore'), `node_modules/
.env
*.js
`);

  console.log(`\n${GREEN}${BOLD}Created Discord bot project: ${projectName}${RESET}\n`);
  console.log(`  ${DIM}cd ${projectName}${RESET}`);
  console.log(`  ${DIM}cp .env.example .env${RESET}       ${DIM}# Add your bot token${RESET}`);
  console.log(`  ${DIM}npm install${RESET}`);
  console.log(`  ${DIM}quill run bot.quill${RESET}\n`);
  console.log(`${DIM}Get a bot token: https://discord.com/developers/applications${RESET}`);
}

// ---- Scaffold: Web Server ----
function cmdWeb() {
  const projectName = process.argv[3] || 'my-web-app';

  if (fs.existsSync(projectName)) {
    error(`Directory "${projectName}" already exists.`);
    process.exit(1);
  }

  fs.mkdirSync(projectName, { recursive: true });

  // package.json
  fs.writeFileSync(path.join(projectName, 'package.json'), JSON.stringify({
    name: projectName,
    version: '1.0.0',
    description: 'A web server built with Quill',
    main: 'server.js',
    scripts: { start: 'node server.js', dev: 'quill run server.quill' },
    dependencies: { express: '^4.18.2' }
  }, null, 2) + '\n');

  // server.quill
  fs.writeFileSync(path.join(projectName, 'server.quill'), `-- Web Server
-- Built with Quill

use "express" as express

app is express()
app.use(express.json())

-- Routes
app on get "/" with req res:
  res.json({ message: "Hello from Quill!" })

app on get "/api/health" with req res:
  res.json({ status: "ok", uptime: process.uptime() })

-- Start server
portNum is process.env.PORT
if portNum is nothing:
  portNum is 3000

app.listen(portNum, with:
  say "Server running on http://localhost:{portNum}"
)
`);

  // .env.example
  fs.writeFileSync(path.join(projectName, '.env.example'), `PORT=3000
`);

  // .gitignore
  fs.writeFileSync(path.join(projectName, '.gitignore'), `node_modules/
.env
*.js
`);

  console.log(`\n${GREEN}${BOLD}Created web server project: ${projectName}${RESET}\n`);
  console.log(`  ${DIM}cd ${projectName}${RESET}`);
  console.log(`  ${DIM}npm install${RESET}`);
  console.log(`  ${DIM}quill run server.quill${RESET}\n`);
}

// ---- Scaffold: Cloudflare Worker ----
function cmdWorker() {
  const projectName = process.argv[3] || 'my-worker';

  if (fs.existsSync(projectName)) {
    error(`Directory "${projectName}" already exists.`);
    process.exit(1);
  }

  fs.mkdirSync(projectName, { recursive: true });

  // package.json
  fs.writeFileSync(path.join(projectName, 'package.json'), JSON.stringify({
    name: projectName,
    version: '1.0.0',
    description: 'A Cloudflare Worker built with Quill',
    main: 'worker.js',
    scripts: {
      build: 'quill build worker.quill',
      dev: 'quill build worker.quill && npx wrangler dev worker.js',
      deploy: 'quill build worker.quill && npx wrangler deploy'
    },
    devDependencies: { wrangler: '^3.0.0' }
  }, null, 2) + '\n');

  // wrangler.toml
  fs.writeFileSync(path.join(projectName, 'wrangler.toml'), `name = "${projectName}"
main = "worker.js"
compatibility_date = "2024-01-01"
`);

  // worker.quill
  fs.writeFileSync(path.join(projectName, 'worker.quill'), `-- Cloudflare Worker
-- Built with Quill

worker on fetch with request:
  url is new URL(request.url)
  path is url.pathname

  if path is "/":
    respond html "<h1>Hello from Quill!</h1>"

  if path is "/api/hello":
    name is url.searchParams.get("name")
    if name is nothing:
      name is "World"
    respond json { message: "Hello, {name}!" }

  respond "Not found" status 404
`);

  // .gitignore
  fs.writeFileSync(path.join(projectName, '.gitignore'), `node_modules/
worker.js
.wrangler/
`);

  console.log(`\n${GREEN}${BOLD}Created Cloudflare Worker project: ${projectName}${RESET}\n`);
  console.log(`  ${DIM}cd ${projectName}${RESET}`);
  console.log(`  ${DIM}npm install${RESET}`);
  console.log(`  ${DIM}npm run dev${RESET}              ${DIM}# Start local dev server${RESET}`);
  console.log(`  ${DIM}npm run deploy${RESET}           ${DIM}# Deploy to Cloudflare${RESET}\n`);
  console.log(`${DIM}Docs: https://quill.tradebuddy.dev/docs/workers${RESET}`);
}

// ---- Scaffold: AI App ----
function cmdAI() {
  const projectName = process.argv[3] || 'my-ai-app';

  if (fs.existsSync(projectName)) {
    error(`Directory "${projectName}" already exists.`);
    process.exit(1);
  }

  fs.mkdirSync(projectName, { recursive: true });

  // package.json
  fs.writeFileSync(path.join(projectName, 'package.json'), JSON.stringify({
    name: projectName,
    version: '1.0.0',
    description: 'An AI app built with Quill',
    main: 'app.js',
    scripts: { start: 'node app.js', dev: 'quill run app.quill' },
    dependencies: { '@anthropic-ai/sdk': '^0.39.0' }
  }, null, 2) + '\n');

  // app.quill
  fs.writeFileSync(path.join(projectName, 'app.quill'), `-- AI App built with Quill
-- Powered by Claude

answer is ask claude "What are 3 fun facts about programming?"
say answer
`);

  // .env.example
  fs.writeFileSync(path.join(projectName, '.env.example'), `# Get your API key from https://console.anthropic.com/
ANTHROPIC_API_KEY=your-api-key-here
`);

  // .gitignore
  fs.writeFileSync(path.join(projectName, '.gitignore'), `node_modules/
.env
*.js
`);

  console.log(`\n${GREEN}${BOLD}Created AI app project: ${projectName}${RESET}\n`);
  console.log(`  ${DIM}cd ${projectName}${RESET}`);
  console.log(`  ${DIM}cp .env.example .env${RESET}       ${DIM}# Add your API key${RESET}`);
  console.log(`  ${DIM}npm install${RESET}`);
  console.log(`  ${DIM}quill run app.quill${RESET}\n`);
  console.log(`${DIM}Get an API key: https://console.anthropic.com/${RESET}`);
}

// ---- Scaffold: Expo / React Native ----
function cmdExpo() {
  const projectName = process.argv[3] || 'my-expo-app';

  if (fs.existsSync(projectName)) {
    error(`Directory "${projectName}" already exists.`);
    process.exit(1);
  }

  fs.mkdirSync(path.join(projectName, 'screens'), { recursive: true });

  // package.json
  fs.writeFileSync(path.join(projectName, 'package.json'), JSON.stringify({
    name: projectName,
    version: '1.0.0',
    main: 'App.js',
    scripts: {
      start: 'expo start',
      build: 'quill build --expo App.quill && for f in screens/*.quill; do quill build --expo "$f"; done',
      android: 'expo start --android',
      ios: 'expo start --ios'
    },
    dependencies: {
      'expo': '~50.0.0',
      'expo-status-bar': '~1.11.1',
      'react': '18.2.0',
      'react-native': '0.73.4',
      '@react-navigation/native': '^6.1.9',
      '@react-navigation/native-stack': '^6.9.17',
      'react-native-screens': '~3.29.0',
      'react-native-safe-area-context': '4.8.2'
    },
    devDependencies: { '@babel/core': '^7.20.0' }
  }, null, 2) + '\n');

  // App.quill
  fs.writeFileSync(path.join(projectName, 'App.quill'), `-- Expo App built with Quill
-- Run: quill build --expo App.quill

use navigation

app navigation:
  stack:
    screen "Home" component HomeScreen
    screen "Details" component DetailsScreen
`);

  // screens/Home.quill
  fs.writeFileSync(path.join(projectName, 'screens', 'Home.quill'), `-- Home Screen

component HomeScreen with navigation:
  state count is 0

  to increment:
    count is count + 1

  to goToDetails:
    navigate to "Details" with { count: count }

  to render:
    view style container:
      text style title: "Welcome to Quill!"
      text style subtitle: "You tapped {count} times"
      button onPress increment style button:
        text style buttonText: "Tap me"
      button onPress goToDetails style link:
        text style linkText: "See Details"

  style native:
    container:
      flex is 1
      align items is "center"
      justify content is "center"
      background color is "#f5f5f5"
    title:
      font size is 28
      font weight is "bold"
      margin bottom is 8
    subtitle:
      font size is 16
      color is "#666"
      margin bottom is 24
    button:
      background color is "#6C5CE7"
      padding horizontal is 32
      padding vertical is 14
      border radius is 12
      margin bottom is 12
    buttonText:
      color is "#fff"
      font size is 16
      font weight is "600"
    link:
      padding is 12
    linkText:
      color is "#6C5CE7"
      font size is 16
`);

  // screens/Details.quill
  fs.writeFileSync(path.join(projectName, 'screens', 'Details.quill'), `-- Details Screen

component DetailsScreen with route navigation:
  state liked is no

  to toggleLike:
    liked is not liked

  to render:
    view style container:
      text style title: "Details"
      text: "Count from Home: {route.params.count}"
      button onPress toggleLike style button:
        if liked:
          text style buttonText: "Liked!"
        otherwise:
          text style buttonText: "Like"

  style native:
    container:
      flex is 1
      align items is "center"
      justify content is "center"
      background color is "#fff"
    title:
      font size is 24
      font weight is "bold"
      margin bottom is 16
    button:
      background color is "#6C5CE7"
      padding horizontal is 32
      padding vertical is 14
      border radius is 12
      margin bottom is 12
    buttonText:
      color is "#fff"
      font size is 16
`);

  // .gitignore
  fs.writeFileSync(path.join(projectName, '.gitignore'), `node_modules/
.expo/
*.jsx
`);

  console.log(`\n${GREEN}${BOLD}Created Expo app project: ${projectName}${RESET}\n`);
  console.log(`  ${DIM}cd ${projectName}${RESET}`);
  console.log(`  ${DIM}npm install${RESET}`);
  console.log(`  ${DIM}quill build --expo App.quill${RESET}`);
  console.log(`  ${DIM}quill build --expo screens/Home.quill${RESET}`);
  console.log(`  ${DIM}quill build --expo screens/Details.quill${RESET}`);
  console.log(`  ${DIM}npx expo start${RESET}\n`);
  console.log(`${DIM}Scan the QR code with Expo Go on your phone!${RESET}`);
}

// ---- AI Generate ----
function cmdGenerate(prompt) {
  const { execSync } = require('child_process');

  const aiPrompt = `You are a Quill programming language code generator. Quill is a beginner-friendly language that compiles to JavaScript and reads like English.

Key Quill syntax:
- "say" instead of console.log
- "is" for assignment: name is "hello"
- "are" for arrays: colors are ["red", "blue"]
- "to" for functions: to greet name: say "Hello, {name}!"
- "give back" instead of return
- "if/otherwise" for conditionals
- "for each x in list:" for loops
- "use" for imports: use "express" as express
- "on" for event handlers: app on get "/" with req res:
- "test/expect" for testing
- String interpolation: "Hello, {name}!"
- Comments start with --

Generate a complete Quill application for the following request. Output ONLY the Quill code, no explanations, no markdown fences.

Request: ${prompt}`;

  // Try Claude CLI
  try {
    execSync('which claude', { stdio: 'ignore' });
    console.log(`${CYAN}🤖 Generating with Claude AI...${RESET}`);
    const output = execSync(`claude -p ${JSON.stringify(aiPrompt)}`, {
      encoding: 'utf-8',
      timeout: 60000,
      stdio: ['pipe', 'pipe', 'ignore']
    });
    if (output && output.trim()) {
      return writeAIOutput(output.trim(), prompt);
    }
  } catch (e) {}

  // Try Gemini CLI
  try {
    execSync('which gemini', { stdio: 'ignore' });
    console.log(`${CYAN}🤖 Generating with Gemini AI...${RESET}`);
    const output = execSync(`gemini -p ${JSON.stringify(aiPrompt)}`, {
      encoding: 'utf-8',
      timeout: 60000,
      stdio: ['pipe', 'pipe', 'ignore']
    });
    if (output && output.trim()) {
      return writeAIOutput(output.trim(), prompt);
    }
  } catch (e) {}

  // Fall back to templates
  console.log(`${YELLOW}No AI CLI found, using built-in templates.${RESET}`);
  console.log(`${DIM}Install Claude CLI or Gemini CLI for AI-powered generation.${RESET}\n`);
  cmdGenerateTemplate(prompt);
}

function writeAIOutput(code, prompt) {
  // Strip markdown fences if present
  if (code.startsWith('```')) {
    const lines = code.split('\n');
    lines.shift(); // remove opening fence
    if (lines[lines.length - 1].trim() === '```') lines.pop();
    code = lines.join('\n');
  }

  const filename = 'app.quill';
  fs.writeFileSync(filename, code + '\n');
  success(`Created ${filename}`);
  console.log(`\n${DIM}Run it: quill run app.quill${RESET}`);
}

function cmdGenerateTemplate(prompt) {
  const lower = prompt.toLowerCase();

  const templates = {
    blog: {
      name: 'blog',
      code: `-- Blog Application
-- Generated by Quill

use "express" as express

app is express()
app.use(express.json())

posts are []

app on get "/api/posts" with req res:
  res.json(posts)

app on post "/api/posts" with req res:
  post is { id: posts.length + 1, title: req.body.title, content: req.body.content, date: new Date() }
  posts.push(post)
  res.status(201).json(post)

app on get "/api/posts/:id" with req res:
  post is posts.find(with p: p.id is Number(req.params.id))
  if post:
    res.json(post)
  otherwise:
    res.status(404).json({ error: "Post not found" })

app.listen(3000, with:
  say "Blog API running on http://localhost:3000"
)
`
    },
    api: {
      name: 'api',
      code: `-- REST API
-- Generated by Quill

use "express" as express

app is express()
app.use(express.json())

items are []

app on get "/api/items" with req res:
  res.json(items)

app on post "/api/items" with req res:
  item is { id: items.length + 1, name: req.body.name, description: req.body.description }
  items.push(item)
  res.status(201).json(item)

app on get "/api/health" with req res:
  res.json({ status: "ok", uptime: process.uptime() })

app.listen(3000, with:
  say "API running on http://localhost:3000"
)
`
    },
    chat: {
      name: 'chat',
      code: `-- Chat Server
-- Generated by Quill

use "express" as express

app is express()
app.use(express.json())

messages are []

app on get "/api/messages" with req res:
  res.json(messages)

app on post "/api/messages" with req res:
  message is { id: messages.length + 1, sender: req.body.sender, content: req.body.content, time: new Date() }
  messages.push(message)
  res.status(201).json(message)

app.listen(3000, with:
  say "Chat API running on http://localhost:3000"
)
`
    },
    discord: {
      name: 'discord',
      code: `-- Discord Bot
-- Generated by Quill

use "discord.js" as Discord

client is new Discord.Client({
  intents: [
    Discord.GatewayIntentBits.Guilds,
    Discord.GatewayIntentBits.GuildMessages,
    Discord.GatewayIntentBits.MessageContent
  ]
})

client on "ready" with:
  say "Bot is online as {client.user.tag}!"

client on "messageCreate" with msg:
  if msg.author.bot:
    give back nothing

  if msg.content is "!ping":
    msg.reply("Pong!")

  if msg.content is "!hello":
    msg.reply("Hello, {msg.author.username}!")

client.login(process.env.DISCORD_TOKEN)
`
    }
  };

  // Match template
  let matched = 'api';
  const keywords = { blog: ['blog', 'post', 'article'], chat: ['chat', 'messaging', 'websocket'], discord: ['discord', 'bot'], api: ['api', 'rest', 'endpoint'] };
  for (const [name, words] of Object.entries(keywords)) {
    if (words.some(w => lower.includes(w))) { matched = name; break; }
  }

  const tmpl = templates[matched];
  const filename = 'app.quill';
  fs.writeFileSync(filename, tmpl.code);
  console.log(`${BOLD}Generating ${tmpl.name} app${RESET}\n`);
  success(`Created ${filename}`);
  console.log(`\n${DIM}Run it: quill run app.quill${RESET}`);
}

// ---- Deploy ----
function cmdDeploy() {
  const dockerfile = `FROM node:20-slim
WORKDIR /app
COPY package.json .
RUN npm install --production
COPY *.js .
CMD ["node", "app.js"]
`;
  fs.writeFileSync('Dockerfile', dockerfile);
  success('Created Dockerfile');

  if (!fs.existsSync('.dockerignore')) {
    fs.writeFileSync('.dockerignore', `node_modules
.env
*.quill
.git
`);
    success('Created .dockerignore');
  }

  console.log(`\n${DIM}Build: docker build -t my-app .${RESET}`);
  console.log(`${DIM}Run:   docker run -d --env-file .env my-app${RESET}`);
}

// ---- Test Runner ----
function cmdTest(filePath) {
  const source = readQuillFile(filePath);
  const { js, errors } = compile(source);

  if (errors.length > 0) {
    for (const err of errors) { error(err); }
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
      JSON, Math, Array, Object, String, Number, Boolean, Date, RegExp, Error, Map, Set, Promise,
      require, process,
    });
    script.runInContext(context);
    success(`\nAll tests passed in ${filePath}`);
  } catch (e) {
    error(e.message);
    process.exit(1);
  }
}

function showHelp() {
  console.log(`
${BOLD}Quill${RESET} v${VERSION} ${DIM}\u2014 a language that reads like English${RESET}

${BOLD}Usage:${RESET}
  quill run <file.quill>       Compile and run a Quill file
  quill build <file.quill>     Compile to JavaScript
  quill check <file.quill>     Check for errors without running
  quill test <file.quill>      Run tests in a Quill file
  quill repl                   Start interactive REPL
  quill init                   Create starter files

${BOLD}Scaffolding:${RESET}
  quill discord [name]         Scaffold a Discord bot project
  quill web [name]             Scaffold an Express web server project
  quill worker [name]          Scaffold a Cloudflare Worker project
  quill ai [name]              Scaffold an AI app project (Claude)
  quill expo [name]            Scaffold an Expo / React Native app
  quill generate "<prompt>"    AI-powered app generation (Claude/Gemini)
  quill deploy                 Generate Dockerfile for deployment

${BOLD}Options:${RESET}
  --version, -v                Show version
  --help, -h                   Show this help

${BOLD}Examples:${RESET}
  ${DIM}$ quill init${RESET}
  ${DIM}$ quill run hello.quill${RESET}
  ${DIM}$ quill build app.quill${RESET}
  ${DIM}$ quill discord my-bot${RESET}
  ${DIM}$ quill web my-api${RESET}
  ${DIM}$ quill generate "a todo API with auth"${RESET}

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
  case 'test':
    if (!args[1]) { error('Missing file argument. Usage: quill test <file.quill>'); process.exit(1); }
    cmdTest(args[1]);
    break;
  case 'repl':
    cmdRepl();
    break;
  case 'init':
    cmdInit();
    break;
  case 'discord':
    cmdDiscord();
    break;
  case 'web':
    cmdWeb();
    break;
  case 'worker':
    cmdWorker();
    break;
  case 'ai':
    cmdAI();
    break;
  case 'expo':
    cmdExpo();
    break;
  case 'generate':
    if (!args[1]) { error('Missing prompt. Usage: quill generate "<prompt>"'); process.exit(1); }
    cmdGenerate(args.slice(1).join(' '));
    break;
  case 'deploy':
    cmdDeploy();
    break;
  default:
    error(`Unknown command: ${cmd}`);
    showHelp();
    process.exit(1);
}
