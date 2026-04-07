package stdlib

// CLIRuntime provides built-in argument and flag parsing for CLI tools.
const CLIRuntime = `
// Quill CLI Runtime

function arg(index) {
  return process.argv[index + 2] || null;
}

function args() {
  return process.argv.slice(2);
}

function flag(name) {
  const prefix = "--" + name;
  const idx = process.argv.indexOf(prefix);
  if (idx === -1) {
    // Check short form
    const short = "-" + name[0];
    const sIdx = process.argv.indexOf(short);
    if (sIdx === -1) return null;
    if (sIdx + 1 < process.argv.length && !process.argv[sIdx + 1].startsWith("-")) {
      return process.argv[sIdx + 1];
    }
    return true;
  }
  if (idx + 1 < process.argv.length && !process.argv[idx + 1].startsWith("-")) {
    return process.argv[idx + 1];
  }
  return true;
}

function hasFlag(name) {
  return process.argv.includes("--" + name) || process.argv.includes("-" + name[0]);
}

const colors = {
  red: (t) => "\x1b[31m" + t + "\x1b[0m",
  green: (t) => "\x1b[32m" + t + "\x1b[0m",
  yellow: (t) => "\x1b[33m" + t + "\x1b[0m",
  blue: (t) => "\x1b[34m" + t + "\x1b[0m",
  cyan: (t) => "\x1b[36m" + t + "\x1b[0m",
  bold: (t) => "\x1b[1m" + t + "\x1b[0m",
  dim: (t) => "\x1b[2m" + t + "\x1b[0m"
};

function exitWith(code, message) {
  if (message) console.error(message);
  process.exit(code || 0);
}
`

// GetCLIRuntime returns the CLI runtime string.
func GetCLIRuntime() string {
	return CLIRuntime
}
