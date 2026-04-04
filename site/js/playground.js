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

  // Tab key support
  editor.addEventListener('keydown', function (e) {
    if (e.key === 'Tab') {
      e.preventDefault();
      const start = this.selectionStart;
      const end = this.selectionEnd;
      this.value = this.value.substring(0, start) + '  ' + this.value.substring(end);
      this.selectionStart = this.selectionEnd = start + 2;
      updateLineNumbers();
    }
  });

  // Ctrl/Cmd + Enter to run
  document.addEventListener('keydown', function (e) {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      e.preventDefault();
      run();
    }
  });

  // Initialize
  loadExample('Hello World');
  exampleSelect.value = 'Hello World';
})();
