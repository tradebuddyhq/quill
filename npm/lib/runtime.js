'use strict';

// ---- Quill Runtime Library ----
// These functions are available to all compiled Quill programs.

function say() {
  const args = Array.from(arguments).map(a => {
    if (a === null || a === undefined) return 'nothing';
    if (typeof a === 'boolean') return a ? 'yes' : 'no';
    if (Array.isArray(a)) return '[' + a.map(x => typeof x === 'string' ? '"' + x + '"' : x).join(', ') + ']';
    if (typeof a === 'object') return JSON.stringify(a);
    return String(a);
  });
  console.log(args.join(' '));
}

function length(x) {
  if (x == null) return 0;
  return x.length;
}

function push(arr, item) {
  arr.push(item);
  return arr;
}

function includes(arr, item) {
  if (arr == null) return false;
  return arr.includes(item);
}

function toText(x) {
  return String(x);
}

function toNumber(x) {
  return Number(x);
}

function trim(s) {
  if (s == null) return '';
  return String(s).trim();
}

function upper(s) {
  if (s == null) return '';
  return String(s).toUpperCase();
}

function lower(s) {
  if (s == null) return '';
  return String(s).toLowerCase();
}

function sort(arr) {
  if (!Array.isArray(arr)) return arr;
  return [...arr].sort((a, b) => {
    if (typeof a === 'number' && typeof b === 'number') return a - b;
    return String(a).localeCompare(String(b));
  });
}

function filter(arr, fn) {
  if (!Array.isArray(arr)) return [];
  return arr.filter(fn);
}

function map_list(arr, fn) {
  if (!Array.isArray(arr)) return [];
  return arr.map(fn);
}

function reduce_list(arr, fn, initial) {
  if (!Array.isArray(arr)) return initial;
  return arr.reduce(fn, initial);
}

function range(start, end) {
  const result = [];
  if (end === undefined) {
    end = start;
    start = 0;
  }
  for (let i = start; i < end; i++) {
    result.push(i);
  }
  return result;
}

function unique(arr) {
  if (!Array.isArray(arr)) return [];
  return [...new Set(arr)];
}

function countWhere(arr, fn) {
  if (!Array.isArray(arr)) return 0;
  return arr.filter(fn).length;
}

async function fetchJSON(url) {
  const res = await fetch(url);
  return res.json();
}

async function postJSON(url, data) {
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  return res.json();
}

function hash(s) {
  let h = 0;
  const str = String(s);
  for (let i = 0; i < str.length; i++) {
    const ch = str.charCodeAt(i);
    h = ((h << 5) - h) + ch;
    h = h & h; // Convert to 32-bit integer
  }
  return Math.abs(h).toString(16);
}

function uuid() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, c => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

// Export all runtime functions
module.exports = {
  say, length, push, includes, toText, toNumber,
  trim, upper, lower,
  sort, filter, map_list, reduce_list,
  range, unique, countWhere,
  fetchJSON, postJSON,
  hash, uuid,
};
