package stdlib

const BrowserRuntime = `
// Quill Browser Runtime
const __contains = (a, b) => { if (typeof a === 'string') return a.includes(b); if (Array.isArray(a)) return a.includes(b); return false; };
let __test_passed = 0;
let __test_failed = 0;

// DOM Selection
const select = (selector) => document.querySelector(selector);
const selectAll = (selector) => [...document.querySelectorAll(selector)];

// DOM Manipulation
const setText = (el, text) => { el.textContent = text; };
const getText = (el) => el.textContent;
const setHTML = (el, html) => { el.innerHTML = html; };
const getHTML = (el) => el.innerHTML;
const sanitizeHTML = (text) => { const d = document.createElement('div'); d.textContent = text; return d.innerHTML; };
const setValue = (el, val) => { el.value = val; };
const getValue = (el) => el.value;
const setAttribute = (el, attr, val) => { el.setAttribute(attr, val); };
const getAttribute = (el, attr) => el.getAttribute(attr);

// CSS Classes
const addClass = (el, ...classes) => { el.classList.add(...classes); };
const removeClass = (el, ...classes) => { el.classList.remove(...classes); };
const toggleClass = (el, cls) => { el.classList.toggle(cls); };
const hasClass = (el, cls) => el.classList.contains(cls);

// Styling
const setStyle = (el, prop, val) => { el.style[prop] = val; };
const getStyle = (el, prop) => getComputedStyle(el)[prop];
const hide = (el) => { el.style.display = 'none'; };
const show = (el) => { el.style.display = ''; };

// Events
const onClick = (el, fn) => { el.addEventListener('click', fn); };
const onInput = (el, fn) => { el.addEventListener('input', (e) => fn(e.target.value)); };
const onChange = (el, fn) => { el.addEventListener('change', (e) => fn(e.target.value)); };
const onSubmit = (el, fn) => { el.addEventListener('submit', (e) => { e.preventDefault(); fn(e); }); };
const onKeyPress = (el, fn) => { el.addEventListener('keydown', (e) => fn(e.key, e)); };
const onLoad = (fn) => { window.addEventListener('DOMContentLoaded', fn); };
const onScroll = (fn) => { window.addEventListener('scroll', fn); };

// Element Creation
const createElement = (tag, text) => { const el = document.createElement(tag); if (text) el.textContent = text; return el; };
const append = (parent, child) => { parent.appendChild(child); return child; };
const prepend = (parent, child) => { parent.prepend(child); return child; };
const removeElement = (el) => { el.remove(); };
const cloneElement = (el) => el.cloneNode(true);

// Navigation & URL
const goTo = (url) => { window.location.href = url; };
const reload = () => { window.location.reload(); };
const currentURL = () => window.location.href;
const getParam = (name) => new URLSearchParams(window.location.search).get(name);

// Storage
const save = (key, value) => { localStorage.setItem(key, JSON.stringify(value)); };
const load = (key) => { const v = localStorage.getItem(key); try { return JSON.parse(v); } catch { return v; } };
const removeData = (key) => { localStorage.removeItem(key); };
const clearData = () => { localStorage.clear(); };

// Timers
const wait = (ms) => new Promise(r => setTimeout(r, ms));
const every = (ms, fn) => setInterval(fn, ms);
const after = (ms, fn) => setTimeout(fn, ms);
const stopTimer = (id) => { clearInterval(id); clearTimeout(id); };

// Fetch (browser-native)
const fetchURL = async (url, options) => { const ctrl = new AbortController(); const t = setTimeout(() => ctrl.abort(), 30000); try { const resp = await fetch(url, { ...options, signal: ctrl.signal }); clearTimeout(t); const text = await resp.text(); try { return JSON.parse(text); } catch { return text; } } catch(e) { clearTimeout(t); if (e.name === 'AbortError') throw new Error('fetchURL timed out after 30 seconds'); throw e; } };
const fetchJSON = async (url) => { const ctrl = new AbortController(); const t = setTimeout(() => ctrl.abort(), 30000); try { const resp = await fetch(url, { signal: ctrl.signal }); clearTimeout(t); if (!resp.ok) throw new Error('fetchJSON failed with status ' + resp.status); return resp.json(); } catch(e) { clearTimeout(t); if (e.name === 'AbortError') throw new Error('fetchJSON timed out after 30 seconds'); throw e; } };
const postJSON = async (url, body) => { const ctrl = new AbortController(); const t = setTimeout(() => ctrl.abort(), 30000); try { const resp = await fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body), signal: ctrl.signal }); clearTimeout(t); if (!resp.ok) throw new Error('postJSON failed with status ' + resp.status); return resp.json(); } catch(e) { clearTimeout(t); if (e.name === 'AbortError') throw new Error('postJSON timed out after 30 seconds'); throw e; } };

// Console
const say = (msg) => console.log(msg);

// Utility (browser-safe versions)
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
const parseJSON = (s) => JSON.parse(s);
const toJSON = (x) => JSON.stringify(x, null, 2);
const keys = (obj) => Object.keys(obj);
const values = (obj) => Object.values(obj);
const filter = (arr, fn) => arr.filter(fn);
const map_list = (arr, fn) => arr.map(fn);
const find = (arr, fn) => arr.find(fn);
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

// Date/Time enhanced
const timestamp = () => Date.now();
const formatDate = (date, fmt) => { const d = new Date(date); const pad = (n) => String(n).padStart(2, '0'); return (fmt || 'YYYY-MM-DD').replace('YYYY', d.getFullYear()).replace('MM', pad(d.getMonth() + 1)).replace('DD', pad(d.getDate())).replace('HH', pad(d.getHours())).replace('mm', pad(d.getMinutes())).replace('ss', pad(d.getSeconds())); };
const formatDateUTC = (date, fmt) => { const d = new Date(date); const pad = (n) => String(n).padStart(2, '0'); return (fmt || 'YYYY-MM-DD').replace('YYYY', d.getUTCFullYear()).replace('MM', pad(d.getUTCMonth() + 1)).replace('DD', pad(d.getUTCDate())).replace('HH', pad(d.getUTCHours())).replace('mm', pad(d.getUTCMinutes())).replace('ss', pad(d.getUTCSeconds())); };
const addDays = (date, days) => { const d = new Date(date); d.setDate(d.getDate() + days); return d.toISOString(); };
const diffDays = (a, b) => Math.round((new Date(b) - new Date(a)) / 86400000);

// RegExp
const matches = (str, pattern) => { const re = new RegExp(pattern, 'g'); const m = str.match(re); return m || []; };
const replacePattern = (str, pattern, replacement) => str.replace(new RegExp(pattern, 'g'), replacement);

// Encoding
const encodeBase64 = (str) => btoa(str);
const decodeBase64 = (str) => atob(str);
const encodeURL = (str) => encodeURIComponent(str);
const decodeURL = (str) => decodeURIComponent(str);

// Concurrency
const parallel = async (...fns) => Promise.all(fns.map(fn => fn()));
const race = async (...fns) => Promise.race(fns.map(fn => fn()));
const delay = (ms) => new Promise(r => setTimeout(r, ms));

// Template engine
const template = (str, data) => { return str.replace(/\{\{(\w+)\}\}/g, (match, key) => data[key] !== undefined ? data[key] : match); };

// UUID (browser-safe)
const uuid = () => { return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => { const r = Math.random() * 16 | 0; return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16); }); };

// Hash (browser-safe, simple)
const hash = async (str) => { const encoder = new TextEncoder(); const data = encoder.encode(str); const hashBuffer = await crypto.subtle.digest('SHA-256', data); return Array.from(new Uint8Array(hashBuffer)).map(b => b.toString(16).padStart(2, '0')).join(''); };
`
