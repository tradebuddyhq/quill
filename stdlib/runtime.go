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
const sortBy = (arr, key) => [...arr].sort((a, b) => a[key] > b[key] ? 1 : a[key] < b[key] ? -1 : 0);
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
const wait = (ms) => new Promise(resolve => setTimeout(resolve, ms));
const parseJSON = (s) => JSON.parse(s);
const toJSON = (x) => JSON.stringify(x, null, 2);
const prompt = (msg) => { const rl = require('readline').createInterface({input: process.stdin, output: process.stdout}); return new Promise(r => rl.question(msg || '> ', a => { rl.close(); r(a); })); };
const exit = (code) => process.exit(code || 0);
const filter = (arr, fn) => arr.filter(fn);
const map_list = (arr, fn) => arr.map(fn);
const find = (arr, fn) => arr.find(fn);
const every = (arr, fn) => arr.every(fn);
const some = (arr, fn) => arr.some(fn);
const reduce = (arr, fn, init) => init !== undefined ? arr.reduce(fn, init) : arr.reduce(fn);
const flat = (arr) => arr.flat();
const includes = (arr, item) => arr.includes(item);
const indexOf = (arr, item) => arr.indexOf(item);
const slice = (arr, start, end) => end !== undefined ? arr.slice(start, end) : arr.slice(start);
const push = (arr, ...items) => { arr.push(...items); return arr; };
const pop = (arr) => arr.pop();
const concat = (a, b) => [...a, ...b];
const zip = (a, b) => a.map((x, i) => [x, b[i]]);
const countWhere = (arr, fn) => arr.filter(fn).length;
const groupBy = (arr, fn) => arr.reduce((acc, item) => { const key = fn(item); (acc[key] = acc[key] || []).push(item); return acc; }, {});
const fetchURL = async (url, options) => { const ctrl = new AbortController(); const t = setTimeout(() => ctrl.abort(), 30000); try { const resp = await fetch(url, { ...options, signal: ctrl.signal }); clearTimeout(t); const text = await resp.text(); try { return JSON.parse(text); } catch { return text; } } catch(e) { clearTimeout(t); if (e.name === 'AbortError') throw new Error('fetchURL timed out after 30 seconds'); throw e; } };
const fetchJSON = async (url) => { const ctrl = new AbortController(); const t = setTimeout(() => ctrl.abort(), 30000); try { const resp = await fetch(url, { signal: ctrl.signal }); clearTimeout(t); if (!resp.ok) throw new Error('fetchJSON failed with status ' + resp.status); return resp.json(); } catch(e) { clearTimeout(t); if (e.name === 'AbortError') throw new Error('fetchJSON timed out after 30 seconds'); throw e; } };
const postJSON = async (url, body) => { const ctrl = new AbortController(); const t = setTimeout(() => ctrl.abort(), 30000); try { const resp = await fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body), signal: ctrl.signal }); clearTimeout(t); if (!resp.ok) throw new Error('postJSON failed with status ' + resp.status); return resp.json(); } catch(e) { clearTimeout(t); if (e.name === 'AbortError') throw new Error('postJSON timed out after 30 seconds'); throw e; } };
const putJSON = async (url, body) => { const ctrl = new AbortController(); const t = setTimeout(() => ctrl.abort(), 30000); try { const resp = await fetch(url, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body), signal: ctrl.signal }); clearTimeout(t); if (!resp.ok) throw new Error('putJSON failed with status ' + resp.status); return resp.json(); } catch(e) { clearTimeout(t); if (e.name === 'AbortError') throw new Error('putJSON timed out after 30 seconds'); throw e; } };
const patchJSON = async (url, body) => { const ctrl = new AbortController(); const t = setTimeout(() => ctrl.abort(), 30000); try { const resp = await fetch(url, { method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body), signal: ctrl.signal }); clearTimeout(t); if (!resp.ok) throw new Error('patchJSON failed with status ' + resp.status); return resp.json(); } catch(e) { clearTimeout(t); if (e.name === 'AbortError') throw new Error('patchJSON timed out after 30 seconds'); throw e; } };
const deleteJSON = async (url, body) => { const ctrl = new AbortController(); const t = setTimeout(() => ctrl.abort(), 30000); try { const resp = await fetch(url, { method: 'DELETE', headers: { 'Content-Type': 'application/json' }, body: body ? JSON.stringify(body) : undefined, signal: ctrl.signal }); clearTimeout(t); if (!resp.ok) throw new Error('deleteJSON failed with status ' + resp.status); const text = await resp.text(); try { return JSON.parse(text); } catch { return text; } } catch(e) { clearTimeout(t); if (e.name === 'AbortError') throw new Error('deleteJSON timed out after 30 seconds'); throw e; } };
const createServer = () => { const http = require('http'); const routes = {}; const srv = { get: (path, handler) => { routes['GET ' + path] = handler; return srv; }, post: (path, handler) => { routes['POST ' + path] = handler; return srv; }, put: (path, handler) => { routes['PUT ' + path] = handler; return srv; }, delete_route: (path, handler) => { routes['DELETE ' + path] = handler; return srv; }, listen: (port, cb) => { const server = http.createServer(async (req, res) => { const url = req.url.split('?')[0]; const key = req.method + ' ' + url; const handler = routes[key]; if (handler) { try { const result = await (typeof handler === 'function' ? handler(req, res) : handler); if (!res.headersSent) { const body = typeof result === 'string' ? result : JSON.stringify(result); res.writeHead(200, { 'Content-Type': typeof result === 'string' ? 'text/html' : 'application/json' }); res.end(body); } } catch(e) { if (!res.headersSent) { res.writeHead(500); res.end('Internal Server Error'); } } } else { res.writeHead(404); res.end('Not Found'); } }); server.listen(port, () => { if (cb) cb(); }); return server; } }; return srv; };
const listFiles = (dir) => require('fs').readdirSync(dir);
const listFilesDeep = (dir) => { const fs = require('fs'); const path = require('path'); const results = []; const walk = (d) => { for (const f of fs.readdirSync(d)) { const full = path.join(d, f); if (fs.statSync(full).isDirectory()) walk(full); else results.push(full); } }; walk(dir); return results; };
const fileInfo = (p) => { const s = require('fs').statSync(p); return { size: s.size, modified: s.mtime.toISOString(), created: s.birthtime.toISOString(), isFile: s.isFile(), isDir: s.isDirectory() }; };
const deleteFile = (p) => require('fs').unlinkSync(p);
const copyFile = (src, dest) => require('fs').copyFileSync(src, dest);
const moveFile = (src, dest) => require('fs').renameSync(src, dest);
const makeDir = (p) => require('fs').mkdirSync(p, { recursive: true });
const watchFiles = (p, callback) => { require('fs').watch(p, { recursive: true }, (event, filename) => { callback(filename || p, event); }); };
const fileExists = (p) => require('fs').existsSync(p);
const readLines = (p) => require('fs').readFileSync(p, 'utf8').split('\n');
const writeLines = (p, lines) => require('fs').writeFileSync(p, lines.join('\n'));
const readJSON = (p) => JSON.parse(require('fs').readFileSync(p, 'utf8'));
const writeJSON = (p, data) => require('fs').writeFileSync(p, JSON.stringify(data, null, 2));
const currentDir = () => process.cwd();
const homePath = () => require('os').homedir();
const joinPath = (...parts) => require('path').join(...parts);
const fileName = (p) => require('path').basename(p);
const fileExtension = (p) => require('path').extname(p);
const parentDir = (p) => require('path').dirname(p);

// Environment
const env = (key, fallback) => process.env[key] !== undefined ? process.env[key] : (fallback !== undefined ? fallback : '');
const setEnv = (key, val) => { process.env[key] = val; };

// Process & Shell
const args = () => process.argv.slice(2);
const run = (cmd) => { const { execSync } = require('child_process'); return execSync(cmd, { encoding: 'utf8' }).trim(); };
const runAsync = async (cmd) => { const { exec } = require('child_process'); return new Promise((resolve, reject) => { exec(cmd, (err, stdout, stderr) => { if (err) reject(stderr || err.message); else resolve(stdout.trim()); }); }); };
const platform = () => process.platform;
const cpuCount = () => require('os').cpus().length;
const memory = () => ({ total: require('os').totalmem(), free: require('os').freemem() });

// Database (SQLite via better-sqlite3)
const openDB = (path) => { try { const Database = require('better-sqlite3'); return new Database(path); } catch(e) { console.error('Install better-sqlite3: npm install better-sqlite3'); throw e; } };
const query = (db, sql, params) => params ? db.prepare(sql).all(...(Array.isArray(params) ? params : [params])) : db.prepare(sql).all();
const execute = (db, sql, params) => params ? db.prepare(sql).run(...(Array.isArray(params) ? params : [params])) : db.prepare(sql).run();
const closeDB = (db) => db.close();

// HTTP Server enhancements
const serveStatic = (dir) => { const fs = require('fs'); const path = require('path'); const mimeTypes = { '.html': 'text/html', '.css': 'text/css', '.js': 'application/javascript', '.json': 'application/json', '.png': 'image/png', '.jpg': 'image/jpeg', '.svg': 'image/svg+xml', '.ico': 'image/x-icon' }; return (req, res) => { const url = req.url === '/' ? '/index.html' : req.url.split('?')[0]; const filePath = path.join(dir, url); if (fs.existsSync(filePath) && fs.statSync(filePath).isFile()) { const ext = path.extname(filePath); res.writeHead(200, { 'Content-Type': mimeTypes[ext] || 'application/octet-stream' }); res.end(fs.readFileSync(filePath)); } else { res.writeHead(404); res.end('Not Found'); } }; };

// Template engine
const template = (str, data) => { return str.replace(/\{\{(\w+)\}\}/g, (match, key) => data[key] !== undefined ? data[key] : match); };

// Crypto/Hashing
const hash = (str, algo) => { const crypto = require('crypto'); return crypto.createHash(algo || 'sha256').update(str).digest('hex'); };
const uuid = () => { const crypto = require('crypto'); return crypto.randomUUID(); };

// Date/Time enhanced
const timestamp = () => Date.now();
const formatDate = (date, fmt) => { const d = new Date(date); const pad = (n) => String(n).padStart(2, '0'); return (fmt || 'YYYY-MM-DD').replace('YYYY', d.getFullYear()).replace('MM', pad(d.getMonth() + 1)).replace('DD', pad(d.getDate())).replace('HH', pad(d.getHours())).replace('mm', pad(d.getMinutes())).replace('ss', pad(d.getSeconds())); };
const formatDateUTC = (date, fmt) => { const d = new Date(date); const pad = (n) => String(n).padStart(2, '0'); return (fmt || 'YYYY-MM-DD').replace('YYYY', d.getUTCFullYear()).replace('MM', pad(d.getUTCMonth() + 1)).replace('DD', pad(d.getUTCDate())).replace('HH', pad(d.getUTCHours())).replace('mm', pad(d.getUTCMinutes())).replace('ss', pad(d.getUTCSeconds())); };
const addDays = (date, days) => { const d = new Date(date); d.setDate(d.getDate() + days); return d.toISOString(); };
const diffDays = (a, b) => Math.round((new Date(b) - new Date(a)) / 86400000);

// RegExp
const regex = (pattern, flags) => new RegExp(pattern, flags || '');
const matches = (str, pattern, flags) => { const re = new RegExp(pattern, flags || 'g'); const m = str.match(re); return m || []; };
const matchGroups = (str, pattern, flags) => { const re = new RegExp(pattern, flags || 'g'); const results = []; let m; while ((m = re.exec(str)) !== null) { results.push(m.slice(1)); if (!re.global) break; } return results; };
const matchFirst = (str, pattern, flags) => { const re = new RegExp(pattern, flags || ''); const m = re.exec(str); return m ? { match: m[0], groups: m.slice(1), index: m.index } : null; };
const matchesPattern = (str, pattern, flags) => new RegExp(pattern, flags || '').test(str);
const replacePattern = (str, pattern, replacement, flags) => str.replace(new RegExp(pattern, flags || 'g'), replacement);

// Encoding
const encodeBase64 = (str) => Buffer.from(str).toString('base64');
const decodeBase64 = (str) => Buffer.from(str, 'base64').toString('utf8');
const encodeURL = (str) => encodeURIComponent(str);
const decodeURL = (str) => decodeURIComponent(str);

// Concurrency
const parallel = async (...fns) => Promise.all(fns.map(fn => fn()));
const race = async (...fns) => Promise.race(fns.map(fn => fn()));
const delay = (ms) => new Promise(r => setTimeout(r, ms));

// Type checking
const isText = (x) => typeof x === 'string';
const isNumber = (x) => typeof x === 'number' && !isNaN(x);
const isList = (x) => Array.isArray(x);
const isObject = (x) => typeof x === 'object' && x !== null && !Array.isArray(x);
const isNothing = (x) => x === null || x === undefined;
const isFunction = (x) => typeof x === 'function';

// Object operations
const merge = (...objs) => Object.assign({}, ...objs);
const pick = (obj, ...keys) => keys.flat().reduce((o, k) => { if (k in obj) o[k] = obj[k]; return o; }, {});
const omit = (obj, ...keys) => { const ks = keys.flat(); return Object.fromEntries(Object.entries(obj).filter(([k]) => !ks.includes(k))); };
const entries = (obj) => Object.entries(obj);
const fromEntries = (arr) => Object.fromEntries(arr);
const hasKey = (obj, key) => key in obj;
const deepCopy = (x) => JSON.parse(JSON.stringify(x));

// String operations enhanced
const padStart = (s, len, ch) => s.padStart(len, ch || ' ');
const padEnd = (s, len, ch) => s.padEnd(len, ch || ' ');
const repeat = (s, n) => s.repeat(n);
const contains = (s, sub) => s.includes(sub);
const capitalize = (s) => s.charAt(0).toUpperCase() + s.slice(1);
const words = (s) => s.trim().split(/\s+/);
const lines = (s) => s.split('\n');
const truncate = (s, len) => s.length > len ? s.slice(0, len) + '...' : s;

// Number formatting
const formatNumber = (n, decimals) => decimals !== undefined ? n.toFixed(decimals) : n.toLocaleString();
const toFixed = (n, digits) => Number(n.toFixed(digits));
const percent = (n, decimals) => (n * 100).toFixed(decimals !== undefined ? decimals : 1) + '%';
const currency = (n, symbol) => (symbol || '$') + n.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
const ordinal = (n) => { const s = ['th','st','nd','rd']; const v = n % 100; return n + (s[(v - 20) % 10] || s[v] || s[0]); };
const clamp = (n, min, max) => Math.min(Math.max(n, min), max);
const lerp = (a, b, t) => a + (b - a) * t;
const mapRange = (n, inMin, inMax, outMin, outMax) => (n - inMin) * (outMax - outMin) / (inMax - inMin) + outMin;
`
