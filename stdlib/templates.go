package stdlib

// GetTemplateRuntime returns the JavaScript runtime for Quill's built-in HTML template engine.
// Provides: html(), tag(), escape(), page(), layout(), partial(), slot()
func GetTemplateRuntime() string {
	return `
// Quill Template Engine Runtime

// Escape HTML entities to prevent XSS
function escapeHTML(str) {
  if (typeof str !== 'string') return String(str != null ? str : '');
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;');
}

// Build an HTML tag with attributes and children
function tag(name, attrsOrChildren, children) {
  var attrs = {};
  var kids = [];
  if (Array.isArray(attrsOrChildren)) {
    kids = attrsOrChildren;
  } else if (typeof attrsOrChildren === 'string') {
    kids = [attrsOrChildren];
  } else if (attrsOrChildren && typeof attrsOrChildren === 'object') {
    attrs = attrsOrChildren;
    if (Array.isArray(children)) kids = children;
    else if (typeof children === 'string') kids = [children];
    else if (children != null) kids = [String(children)];
  }
  var attrStr = '';
  for (var key in attrs) {
    if (attrs[key] === true) { attrStr += ' ' + key; }
    else if (attrs[key] !== false && attrs[key] != null) { attrStr += ' ' + key + '="' + escapeHTML(attrs[key]) + '"'; }
  }
  var voidTags = {area:1,base:1,br:1,col:1,embed:1,hr:1,img:1,input:1,link:1,meta:1,source:1,track:1,wbr:1};
  if (voidTags[name]) return '<' + name + attrStr + ' />';
  return '<' + name + attrStr + '>' + kids.join('') + '</' + name + '>';
}

// Build a full HTML page with head and body
// Supports: page({title: "My Page", body: "..."}) or page("My Page", "...")
function page(optionsOrTitle, bodyArg) {
  var options = optionsOrTitle;
  if (typeof optionsOrTitle === 'string') {
    options = { title: optionsOrTitle, body: bodyArg || '' };
  }
  var title = options.title || 'Quill App';
  var lang = options.lang || 'en';
  var head = options.head || '';
  var body = options.body || '';
  var styles = options.styles || '';
  var scripts = options.scripts || '';
  var charset = options.charset || 'utf-8';
  var viewport = options.viewport !== false ? '<meta name="viewport" content="width=device-width, initial-scale=1.0" />' : '';

  var headContent = '<meta charset="' + charset + '" />' + viewport + '<title>' + escapeHTML(title) + '</title>';
  if (styles) headContent += '<style>' + styles + '</style>';
  if (head) headContent += head;

  var bodyContent = body;
  if (scripts) bodyContent += '<script>' + scripts + '</script>';

  return '<!DOCTYPE html><html lang="' + lang + '"><head>' + headContent + '</head><body>' + bodyContent + '</body></html>';
}

// Layout system for template inheritance
var __layouts = {};

function layout(name, templateFn) {
  __layouts[name] = templateFn;
}

function renderLayout(name, data) {
  if (!__layouts[name]) throw new Error('Layout "' + name + '" not found');
  return __layouts[name](data || {});
}

// Partial system for reusable template fragments
var __partials = {};

function partial(name, templateFn) {
  if (typeof templateFn === 'function') {
    __partials[name] = templateFn;
  } else {
    // Render a previously registered partial
    if (!__partials[name]) throw new Error('Partial "' + name + '" not found');
    return __partials[name](templateFn || {});
  }
}

// Conditional rendering helper
function showIf(condition, content) {
  return condition ? content : '';
}

// List rendering helper
function eachItem(items, fn) {
  if (!Array.isArray(items)) return '';
  return items.map(fn).join('');
}
function each(items, fn) {
  if (!Array.isArray(items)) return '';
  return items.map(fn).join('');
}

// Raw HTML (bypass escaping — use with caution)
function raw(str) { return str; }

// Common shorthand tag builders
function div(a, b) { return tag('div', a, b); }
function span(a, b) { return tag('span', a, b); }
function p(a, b) { return tag('p', a, b); }
function a(a, b) { return tag('a', a, b); }
function img(attrs) { return tag('img', attrs); }
function ul(a, b) { return tag('ul', a, b); }
function ol(a, b) { return tag('ol', a, b); }
function li(a, b) { return tag('li', a, b); }
function h1(a, b) { return tag('h1', a, b); }
function h2(a, b) { return tag('h2', a, b); }
function h3(a, b) { return tag('h3', a, b); }
function form(a, b) { return tag('form', a, b); }
function input(attrs) { return tag('input', attrs); }
function button(a, b) { return tag('button', a, b); }
function label(a, b) { return tag('label', a, b); }
function textarea(a, b) { return tag('textarea', a, b); }
function table(a, b) { return tag('table', a, b); }
function tr(a, b) { return tag('tr', a, b); }
function th(a, b) { return tag('th', a, b); }
function td(a, b) { return tag('td', a, b); }
function nav(a, b) { return tag('nav', a, b); }
function header(a, b) { return tag('header', a, b); }
function footer(a, b) { return tag('footer', a, b); }
function section(a, b) { return tag('section', a, b); }
function article(a, b) { return tag('article', a, b); }
function main(a, b) { return tag('main', a, b); }
`
}
