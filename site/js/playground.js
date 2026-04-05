// ============================================
// Quill Playground - UI Logic
// ============================================

(function () {
  'use strict';

  const examples = {
    "Hello World": `-- Welcome to Quill!\nname is "World"\nsay "Hello, {name}!"\n\nage is 25\nif age is greater than 18:\n  say "You are an adult"\notherwise:\n  say "You are young"`,

    "FizzBuzz": `to fizzbuzz n:\n  i is 1\n  while i is less than n + 1:\n    if i % 15 is 0:\n      say "FizzBuzz"\n    otherwise if i % 3 is 0:\n      say "Fizz"\n    otherwise if i % 5 is 0:\n      say "Buzz"\n    otherwise:\n      say i\n    i is i + 1\n\nfizzbuzz(20)`,

    "Calculator": `to add a b:\n  give back a + b\n\nto subtract a b:\n  give back a - b\n\nto multiply a b:\n  give back a * b\n\nto divide a b:\n  if b is 0:\n    say "Cannot divide by zero!"\n    give back 0\n  give back a / b\n\nx is 100\ny is 25\nsay "{x} + {y} = {add(x, y)}"\nsay "{x} - {y} = {subtract(x, y)}"\nsay "{x} * {y} = {multiply(x, y)}"\nsay "{x} / {y} = {divide(x, y)}"`,

    "Lists & Loops": `colors are ["red", "green", "blue", "yellow"]\nsay "Colors: {join(colors, ', ')}"\n\nfor each color in colors:\n  say "I like {color}!"\n\nnumbers are [5, 3, 8, 1, 9, 2]\nsay "Sorted: {join(sort(numbers), ', ')}"\nsay "Sum: {sum(numbers)}"\nsay "Length: {length(numbers)}"`,

    "Testing": `to add a b:\n  give back a + b\n\nto multiply a b:\n  give back a * b\n\ntest "addition":\n  expect add(2, 3) is 5\n  expect add(-1, 1) is 0\n\ntest "multiplication":\n  expect multiply(3, 4) is 12\n  expect multiply(0, 100) is 0\n\ntest "strings":\n  expect upper("hello") is "HELLO"\n  expect length("quill") is 5`
  };

  const editor = document.getElementById('editor');
  const output = document.getElementById('output');
  const lineNumbers = document.getElementById('line-numbers');
  const runBtn = document.getElementById('run-btn');
  const clearBtn = document.getElementById('clear-btn');
  const exampleSelect = document.getElementById('example-select');

  function updateLineNumbers() {
    const lines = editor.value.split('\n').length;
    const nums = [];
    for (let i = 1; i <= lines; i++) {
      nums.push(i);
    }
    lineNumbers.textContent = nums.join('\n');
  }

  function loadExample(name) {
    if (examples[name]) {
      editor.value = examples[name];
      updateLineNumbers();
      clearOutput();
    }
  }

  function clearOutput() {
    output.innerHTML = '<span class="output-line" style="color: var(--muted-text);">Click "Run" or press Ctrl+Enter to execute your code.</span>';
  }

  function appendOutput(text, className) {
    if (output.querySelector('.output-line[style]') && output.children.length === 1) {
      output.innerHTML = '';
    }
    const line = document.createElement('div');
    line.className = 'output-line' + (className ? ' ' + className : '');
    line.textContent = text;
    output.appendChild(line);
  }

  function run() {
    output.innerHTML = '';
    const source = editor.value;

    try {
      const js = window.QuillCompiler.compile(source);

      // Override console.log to capture output
      const originalLog = console.log;
      const logs = [];
      console.log = function (...args) {
        logs.push(args.map(a => {
          if (typeof a === 'object') return JSON.stringify(a, null, 2);
          return String(a);
        }).join(' '));
      };

      try {
        // Use Function constructor to run in a somewhat isolated scope
        const fn = new Function(js);
        fn();

        // Display captured output
        if (logs.length === 0) {
          appendOutput('(no output)', '');
        } else {
          for (const log of logs) {
            // Color test results
            if (log.startsWith('  \u2713')) {
              appendOutput(log, 'output-success');
            } else if (log.startsWith('  \u2717')) {
              appendOutput(log, 'output-error');
            } else {
              appendOutput(log, '');
            }
          }
        }
      } catch (runtimeErr) {
        appendOutput('Runtime Error: ' + runtimeErr.message, 'output-error');
      } finally {
        console.log = originalLog;
      }
    } catch (compileErr) {
      appendOutput('Compile Error: ' + compileErr.message, 'output-error');
    }
  }

  // Event listeners
  runBtn.addEventListener('click', run);

  clearBtn.addEventListener('click', function () {
    editor.value = '';
    updateLineNumbers();
    clearOutput();
  });

  exampleSelect.addEventListener('change', function () {
    if (this.value) {
      loadExample(this.value);
    }
  });

  editor.addEventListener('input', updateLineNumbers);
  editor.addEventListener('scroll', function () {
    lineNumbers.style.transform = 'translateY(' + (-editor.scrollTop) + 'px)';
  });

  // ---- IDE Features ----
  const INDENT = '  ';
  const AUTO_CLOSE = { '(': ')', '[': ']', '{': '}', '"': '"', "'": "'" };
  const BLOCK_STARTERS = /^(to |if |otherwise|for each |while |describe |it |test |match |class |trait |spawn |parallel|race|select|server|database|auth|websocket )/;

  function getLineAt(text, pos) {
    const start = text.lastIndexOf('\n', pos - 1) + 1;
    const end = text.indexOf('\n', pos);
    return text.substring(start, end === -1 ? text.length : end);
  }

  function getIndent(line) {
    const match = line.match(/^(\s*)/);
    return match ? match[1] : '';
  }

  editor.addEventListener('keydown', function (e) {
    const val = this.value;
    const start = this.selectionStart;
    const end = this.selectionEnd;
    const hasSelection = start !== end;

    // Ctrl/Cmd + Enter to run
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      e.preventDefault();
      run();
      return;
    }

    // Ctrl/Cmd + / to toggle comment
    if ((e.ctrlKey || e.metaKey) && e.key === '/') {
      e.preventDefault();
      const lineStart = val.lastIndexOf('\n', start - 1) + 1;
      const lineEnd = val.indexOf('\n', end);
      const selEnd = lineEnd === -1 ? val.length : lineEnd;
      const lines = val.substring(lineStart, selEnd).split('\n');
      const allCommented = lines.every(function(l) { return l.trimStart().startsWith('-- ') || l.trim() === ''; });
      const newLines = lines.map(function(l) {
        if (l.trim() === '') return l;
        if (allCommented) return l.replace(/^(\s*)-- /, '$1');
        return l.match(/^(\s*)/)[1] + '-- ' + l.trimStart();
      });
      const newText = newLines.join('\n');
      editor.value = val.substring(0, lineStart) + newText + val.substring(selEnd);
      editor.selectionStart = lineStart;
      editor.selectionEnd = lineStart + newText.length;
      updateLineNumbers();
      return;
    }

    // Ctrl/Cmd + D to duplicate line
    if ((e.ctrlKey || e.metaKey) && e.key === 'd') {
      e.preventDefault();
      const lineStart = val.lastIndexOf('\n', start - 1) + 1;
      const lineEnd = val.indexOf('\n', start);
      const lineEndPos = lineEnd === -1 ? val.length : lineEnd;
      const line = val.substring(lineStart, lineEndPos);
      editor.value = val.substring(0, lineEndPos) + '\n' + line + val.substring(lineEndPos);
      editor.selectionStart = editor.selectionEnd = lineEndPos + 1 + line.length;
      updateLineNumbers();
      return;
    }

    // Enter with auto-indent
    if (e.key === 'Enter' && !e.ctrlKey && !e.metaKey) {
      e.preventDefault();
      const currentLine = getLineAt(val, start);
      const currentIndent = getIndent(currentLine);
      const trimmedLine = currentLine.trim();
      var newIndent = currentIndent;
      if (trimmedLine.endsWith(':') || BLOCK_STARTERS.test(trimmedLine)) {
        newIndent = currentIndent + INDENT;
      }
      const insertion = '\n' + newIndent;
      editor.value = val.substring(0, start) + insertion + val.substring(end);
      editor.selectionStart = editor.selectionEnd = start + insertion.length;
      updateLineNumbers();
      return;
    }

    // Tab / Shift+Tab
    if (e.key === 'Tab') {
      e.preventDefault();
      if (hasSelection) {
        const lineStart = val.lastIndexOf('\n', start - 1) + 1;
        const lineEnd = val.indexOf('\n', end - 1);
        const selEnd = lineEnd === -1 ? val.length : lineEnd;
        const lines = val.substring(lineStart, selEnd).split('\n');
        var newLines;
        if (e.shiftKey) {
          newLines = lines.map(function(l) { return l.startsWith(INDENT) ? l.substring(INDENT.length) : l.replace(/^\s/, ''); });
        } else {
          newLines = lines.map(function(l) { return INDENT + l; });
        }
        const newText = newLines.join('\n');
        editor.value = val.substring(0, lineStart) + newText + val.substring(selEnd);
        editor.selectionStart = lineStart;
        editor.selectionEnd = lineStart + newText.length;
      } else if (e.shiftKey) {
        const lineStart = val.lastIndexOf('\n', start - 1) + 1;
        const line = getLineAt(val, start);
        if (line.startsWith(INDENT)) {
          const cursorOffset = start - lineStart;
          const lineEnd = val.indexOf('\n', start);
          const lineEndPos = lineEnd === -1 ? val.length : lineEnd;
          editor.value = val.substring(0, lineStart) + line.substring(INDENT.length) + val.substring(lineEndPos);
          editor.selectionStart = editor.selectionEnd = lineStart + Math.max(0, cursorOffset - INDENT.length);
        }
      } else {
        editor.value = val.substring(0, start) + INDENT + val.substring(end);
        editor.selectionStart = editor.selectionEnd = start + INDENT.length;
      }
      updateLineNumbers();
      return;
    }

    // Auto-close brackets and quotes
    if (AUTO_CLOSE[e.key]) {
      const closing = AUTO_CLOSE[e.key];
      if (e.key === '"' || e.key === "'") {
        if (start > 0 && /\w/.test(val[start - 1])) return;
        if (val[start] === e.key) { e.preventDefault(); editor.selectionStart = editor.selectionEnd = start + 1; return; }
      }
      if (!hasSelection) {
        e.preventDefault();
        editor.value = val.substring(0, start) + e.key + closing + val.substring(end);
        editor.selectionStart = editor.selectionEnd = start + 1;
      }
      return;
    }

    // Skip over closing brackets/quotes
    if ((e.key === ')' || e.key === ']' || e.key === '}' || e.key === '"' || e.key === "'") && val[start] === e.key) {
      e.preventDefault();
      editor.selectionStart = editor.selectionEnd = start + 1;
      return;
    }

    // Backspace: delete matching pair
    if (e.key === 'Backspace' && !hasSelection && start > 0) {
      const before = val[start - 1];
      const after = val[start];
      if (AUTO_CLOSE[before] && AUTO_CLOSE[before] === after) {
        e.preventDefault();
        editor.value = val.substring(0, start - 1) + val.substring(start + 1);
        editor.selectionStart = editor.selectionEnd = start - 1;
        updateLineNumbers();
        return;
      }
    }
  });

  // Initialize
  loadExample('Hello World');
  exampleSelect.value = 'Hello World';
})();
