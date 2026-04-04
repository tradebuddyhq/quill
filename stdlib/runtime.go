package stdlib

const Runtime = `
// Quill Standard Library
const __contains = (a, b) => { if (typeof a === 'string') return a.includes(b); if (Array.isArray(a)) return a.includes(b); return false; };
let __test_passed = 0;
let __test_failed = 0;
const read = (path) => require('fs').readFileSync(path, 'utf8');
const write = (path, content) => require('fs').writeFileSync(path, content);
const append_file = (path, content) => require('fs').appendFileSync(path, content);
const exists = (path) => require('fs').existsSync(path);
const length = (x) => x.length;
const join = (arr, sep) => arr.join(sep === undefined ? ', ' : sep);
const split = (str, sep) => str.split(sep);
const round = (n) => Math.round(n);
const floor = (n) => Math.floor(n);
const ceil = (n) => Math.ceil(n);
const abs = (n) => Math.abs(n);
const random = () => Math.random();
const randomInt = (a, b) => Math.floor(Math.random() * (b - a + 1)) + a;
const toNumber = (x) => Number(x);
const toText = (x) => String(x);
const keys = (obj) => Object.keys(obj);
const values = (obj) => Object.values(obj);
const typeOf = (x) => { if (Array.isArray(x)) return 'list'; if (x === null) return 'nothing'; return typeof x; };
const range = (start, end) => Array.from({length: end - start}, (_, i) => i + start);
const sort = (arr) => [...arr].sort((a, b) => a > b ? 1 : -1);
const reverse = (arr) => [...arr].reverse();
const unique = (arr) => [...new Set(arr)];
const sum = (arr) => arr.reduce((a, b) => a + b, 0);
const smallest = (...args) => Math.min(...args.flat());
const largest = (...args) => Math.max(...args.flat());
const trim = (s) => s.trim();
const upper = (s) => s.toUpperCase();
const lower = (s) => s.toLowerCase();
const replace_text = (s, old_text, new_text) => s.replaceAll(old_text, new_text);
const startsWith = (s, prefix) => s.startsWith(prefix);
const endsWith = (s, suffix) => s.endsWith(suffix);
const now = () => new Date().toISOString();
const today = () => new Date().toISOString().split('T')[0];
const wait = (ms) => { const end = Date.now() + ms; while (Date.now() < end) {} };
const parseJSON = (s) => JSON.parse(s);
const toJSON = (x) => JSON.stringify(x, null, 2);
const prompt = (msg) => { const rl = require('readline').createInterface({input: process.stdin, output: process.stdout}); return new Promise(r => rl.question(msg || '> ', a => { rl.close(); r(a); })); };
const exit = (code) => process.exit(code || 0);
`
