'use strict';

class QuillCompiler {
  compile(source) {
    const errors = [];
    try {
      const js = this.transpile(source, errors);
      return { js, errors };
    } catch (e) {
      errors.push(e.message);
      return { js: '', errors };
    }
  }

  transpile(source, errors) {
    // Normalize line endings
    let lines = source.replace(/\r\n/g, '\n').replace(/\r/g, '\n').split('\n');

    // Process into a structured form with indent levels
    const processed = [];
    for (let i = 0; i < lines.length; i++) {
      const raw = lines[i];
      const stripped = raw.replace(/\t/g, '  ');
      const trimmed = stripped.trimStart();
      const indent = stripped.length - trimmed.length;
      const level = Math.floor(indent / 2);
      processed.push({ raw, trimmed, level, lineNum: i + 1 });
    }

    // Generate JS with indentation-based blocks
    const jsLines = [];
    const indentStack = [0]; // stack of indent levels for closing braces
    const blockTypeStack = ['root']; // track what type of block we're in

    for (let i = 0; i < processed.length; i++) {
      const { trimmed, level, lineNum } = processed[i];

      // Skip empty lines
      if (trimmed === '') {
        jsLines.push('');
        continue;
      }

      // Check if this is a continuation keyword (otherwise, otherwise if, when)
      // These should not close the parent block, they continue it
      const isContinuation = /^(otherwise|when\s)/.test(trimmed.replace(/:$/, '').trim());

      // Close blocks: if current level < previous block levels, close them
      // But for continuation keywords, only close down to one level above
      while (indentStack.length > 1 && level < indentStack[indentStack.length - 1]) {
        // For continuation keywords at the same level as the block opener,
        // we need to close the inner block but NOT the outer block
        if (isContinuation && indentStack.length === 2 && level === indentStack[0]) {
          // Close inner block only
          indentStack.pop();
          blockTypeStack.pop();
          const closeIndent = '  '.repeat(indentStack.length - 1);
          jsLines.push(closeIndent + '}');
          break;
        }
        indentStack.pop();
        blockTypeStack.pop();
        const closeIndent = '  '.repeat(indentStack.length - 1);
        jsLines.push(closeIndent + '}');
      }

      const jsIndent = '  '.repeat(indentStack.length - 1);

      // Check if next non-empty line has greater indentation (block opener)
      const opensBlock = this.nextNonEmptyLevel(processed, i) > level;

      // Determine if we're inside a match/switch block
      const inMatch = blockTypeStack.includes('match');

      const js = this.compileLine(trimmed, opensBlock, errors, lineNum, inMatch);
      if (js !== null) {
        jsLines.push(jsIndent + js);
        if (opensBlock && (js.endsWith('{') || js.endsWith('{ // match'))) {
          indentStack.push(level + 1);
          // Track block type
          if (js.includes('// match')) {
            blockTypeStack.push('match');
          } else if (js.includes('case ') || js.includes('default:')) {
            blockTypeStack.push('case');
          } else {
            blockTypeStack.push('block');
          }
        }
      }
    }

    // Close remaining open blocks
    while (indentStack.length > 1) {
      indentStack.pop();
      const closeIndent = '  '.repeat(indentStack.length - 1);
      jsLines.push(closeIndent + '}');
    }

    return jsLines.join('\n');
  }

  nextNonEmptyLevel(processed, currentIdx) {
    for (let j = currentIdx + 1; j < processed.length; j++) {
      if (processed[j].trimmed !== '') return processed[j].level;
    }
    return 0;
  }

  compileLine(line, opensBlock, errors, lineNum, inMatch) {
    // Comments
    if (line.startsWith('--')) {
      return '//' + line.slice(2);
    }

    // Remove trailing colon for block openers
    let stmt = line;
    if (opensBlock && stmt.endsWith(':')) {
      stmt = stmt.slice(0, -1).trimEnd();
    }

    const compileExpr = (expr) => this.compileExpression(expr);

    // ---- use "module" (import) ----
    const useMatch = stmt.match(/^use\s+"([^"]+)"$/);
    if (useMatch) {
      return `const ${useMatch[1].replace(/[^a-zA-Z0-9_]/g, '_')} = require("${useMatch[1]}");`;
    }

    // ---- describe ClassName (class) ----
    const describeMatch = stmt.match(/^describe\s+(\w+)$/);
    if (describeMatch) {
      return `class ${describeMatch[1]} {`;
    }

    // ---- to create (constructor) ----
    const createMatch = stmt.match(/^to\s+create(?:\s+(.*))?$/);
    if (createMatch) {
      const params = createMatch[1] ? createMatch[1].split(/\s*,\s*/).join(', ') : '';
      return `constructor(${params}) {`;
    }

    // ---- to functionName params (method/function) ----
    const fnMatch = stmt.match(/^to\s+(\w+)(?:\s+(.*))?$/);
    if (fnMatch) {
      const name = fnMatch[1];
      const params = fnMatch[2] ? fnMatch[2].split(/\s*,\s*/).join(', ') : '';
      return `function ${name}(${params}) {`;
    }

    // ---- give back (return) ----
    const returnMatch = stmt.match(/^give\s+back\s+(.+)$/);
    if (returnMatch) {
      return `return ${compileExpr(returnMatch[1])};`;
    }

    // ---- say (print) ----
    const sayMatch = stmt.match(/^say\s+(.+)$/);
    if (sayMatch) {
      return `say(${compileExpr(sayMatch[1])});`;
    }

    // ---- for each x in y ----
    const forMatch = stmt.match(/^for\s+each\s+(\w+)\s+in\s+(.+)$/);
    if (forMatch) {
      return `for (const ${forMatch[1]} of ${compileExpr(forMatch[2])}) {`;
    }

    // ---- repeat while condition ----
    const whileMatch = stmt.match(/^repeat\s+while\s+(.+)$/);
    if (whileMatch) {
      return `while (${compileExpr(whileMatch[1])}) {`;
    }

    // ---- while condition ----
    const whileMatch2 = stmt.match(/^while\s+(.+)$/);
    if (whileMatch2) {
      return `while (${compileExpr(whileMatch2[1])}) {`;
    }

    // ---- match value ----
    const matchMatch = stmt.match(/^match\s+(.+)$/);
    if (matchMatch) {
      return `switch (${compileExpr(matchMatch[1])}) { // match`;
    }

    // ---- when value ----
    const whenMatch = stmt.match(/^when\s+(.+)$/);
    if (whenMatch) {
      const val = whenMatch[1].trim();
      return `case ${compileExpr(val)}: {`;
    }

    // ---- otherwise (in match = default, in if = else) ----
    if (stmt === 'otherwise') {
      if (inMatch) {
        return `default: {`;
      }
      return `} else {`;
    }

    // ---- otherwise if ----
    const elseIfMatch = stmt.match(/^otherwise\s+if\s+(.+)$/);
    if (elseIfMatch) {
      return `} else if (${compileExpr(elseIfMatch[1])}) {`;
    }

    // ---- if condition ----
    const ifMatch = stmt.match(/^if\s+(.+)$/);
    if (ifMatch) {
      return `if (${compileExpr(ifMatch[1])}) {`;
    }

    // ---- test "name" ----
    const testMatch = stmt.match(/^test\s+"([^"]+)"$/);
    if (testMatch) {
      return `/* test: ${testMatch[1]} */ (function() {`;
    }

    // ---- expect expression ----
    const expectMatch = stmt.match(/^expect\s+(.+)$/);
    if (expectMatch) {
      return `if (!(${compileExpr(expectMatch[1])})) throw new Error("Expectation failed: ${expectMatch[1].replace(/"/g, '\\"')}");`;
    }

    // ---- spawn (async task) ----
    const spawnMatch = stmt.match(/^spawn\s+(.+)$/);
    if (spawnMatch) {
      return `(async () => { await ${compileExpr(spawnMatch[1])}; })();`;
    }

    // ---- await ----
    const awaitMatch = stmt.match(/^await\s+(.+)$/);
    if (awaitMatch) {
      return `await ${compileExpr(awaitMatch[1])};`;
    }

    // ---- set x is value / x is value (variable assignment) ----
    const setMatch = stmt.match(/^set\s+(\w+)\s+is\s+(.+)$/);
    if (setMatch) {
      return `let ${setMatch[1]} = ${compileExpr(setMatch[2])};`;
    }

    // ---- my x is value (this.x = value) ----
    const myMatch = stmt.match(/^my\s+(\w+)\s+is\s+(.+)$/);
    if (myMatch) {
      return `this.${myMatch[1]} = ${compileExpr(myMatch[2])};`;
    }

    // ---- x is value (variable) ----
    const varMatch = stmt.match(/^(\w+)\s+is\s+(.+)$/);
    if (varMatch && !varMatch[2].match(/^(greater|less|not)\b/) && varMatch[1] !== 'if' && varMatch[1] !== 'otherwise') {
      return `let ${varMatch[1]} = ${compileExpr(varMatch[2])};`;
    }

    // ---- x are value (alias for is, often for lists) ----
    const areMatch = stmt.match(/^(\w+)\s+are\s+(.+)$/);
    if (areMatch) {
      return `let ${areMatch[1]} = ${compileExpr(areMatch[2])};`;
    }

    // ---- fallback: treat as expression ----
    return compileExpr(stmt) + ';';
  }

  compileExpression(expr) {
    if (!expr) return expr;

    let result = expr;

    // String interpolation: "Hello {name}" -> `Hello ${name}`
    result = result.replace(/"([^"]*)"/g, (match, inner) => {
      if (inner.includes('{')) {
        const tmpl = inner.replace(/\{([^}]+)\}/g, (_, e) => '${' + this.compileExpression(e) + '}');
        return '`' + tmpl + '`';
      }
      return match;
    });

    // Pipe operator: val | fn -> fn(val)
    if (result.includes(' | ')) {
      const parts = result.split(/\s*\|\s*/);
      let compiled = this.compileExpression(parts[0]);
      for (let i = 1; i < parts.length; i++) {
        compiled = this.compileExpression(parts[i]) + '(' + compiled + ')';
      }
      return compiled;
    }

    // Boolean literals
    result = result.replace(/\byes\b/g, 'true');
    result = result.replace(/\bno\b/g, 'false');
    result = result.replace(/\bnothing\b/g, 'null');

    // "new ClassName(...)" stays as-is
    // result = result; // no-op, new is valid JS

    // Logical operators
    result = result.replace(/\band\b/g, '&&');
    result = result.replace(/\bor\b/g, '||');
    result = result.replace(/\bnot\s+/g, '!');

    // Comparison operators (multi-word, must come before single-word "is")
    result = result.replace(/\bis greater than or equal to\b/g, '>=');
    result = result.replace(/\bis less than or equal to\b/g, '<=');
    result = result.replace(/\bis greater than\b/g, '>');
    result = result.replace(/\bis less than\b/g, '<');
    result = result.replace(/\bis not\b/g, '!==');
    // "is" as equality (only in expression context)
    result = result.replace(/\bis\b/g, '===');

    // "my " -> "this."
    result = result.replace(/\bmy\s+/g, 'this.');

    // "contains" -> .includes()
    result = result.replace(/(\w+)\s+contains\s+(.+)/g, '$1.includes($2)');

    return result;
  }
}

// Post-processing passes

function fixClassMethods(js) {
  const lines = js.split('\n');
  let inClass = false;
  let classDepth = 0;
  let braceCount = 0;

  for (let i = 0; i < lines.length; i++) {
    const trimmed = lines[i].trim();

    if (trimmed.match(/^class\s+\w+\s*\{/)) {
      inClass = true;
      classDepth = braceCount;
    }

    if (inClass) {
      lines[i] = lines[i].replace(/^(\s*)function\s+(\w+)\s*\(/, '$1$2(');
    }

    for (const ch of lines[i]) {
      if (ch === '{') braceCount++;
      if (ch === '}') {
        braceCount--;
        if (inClass && braceCount === classDepth) {
          inClass = false;
        }
      }
    }
  }

  return lines.join('\n');
}

function fixDuplicateLets(js) {
  const declared = new Set();
  return js.split('\n').map(line => {
    const m = line.match(/^(\s*)let\s+(\w+)\s*=\s*(.+)$/);
    if (m) {
      const [, indent, name, value] = m;
      if (declared.has(name)) {
        return `${indent}${name} = ${value}`;
      }
      declared.add(name);
    }
    return line;
  }).join('\n');
}

function fixSwitchCase(js) {
  const lines = js.split('\n');
  const result = [];
  let inSwitch = false;
  let switchBraceDepth = 0;
  let braceDepth = 0;

  for (let i = 0; i < lines.length; i++) {
    const trimmed = lines[i].trim();

    // Track switch blocks
    if (trimmed.endsWith('{ // match')) {
      inSwitch = true;
      switchBraceDepth = braceDepth;
    }

    // Only add break/close inside switch blocks
    if (inSwitch && (trimmed.startsWith('case ') || trimmed.startsWith('default:')) && result.length > 0) {
      const prev = result[result.length - 1].trim();
      if (prev !== '{' && !prev.startsWith('case ') && !prev.startsWith('default:') && prev !== '// match' && !prev.endsWith('{ // match')) {
        const indent = lines[i].match(/^(\s*)/)[1];
        result.push(indent + '  break;');
        result.push(indent + '}');
      }
    }

    result.push(lines[i]);

    // Track braces
    for (const ch of lines[i]) {
      if (ch === '{') braceDepth++;
      if (ch === '}') {
        braceDepth--;
        if (inSwitch && braceDepth === switchBraceDepth) {
          inSwitch = false;
        }
      }
    }
  }
  return result.join('\n');
}

/**
 * Compile Quill source code to JavaScript.
 * @param {string} source - Quill source code
 * @returns {{ js: string, errors: string[] }}
 */
function compile(source) {
  const compiler = new QuillCompiler();
  const { js, errors } = compiler.compile(source);

  if (errors.length > 0) {
    return { js, errors };
  }

  // Post-process
  let finalJs = fixClassMethods(js);
  finalJs = fixDuplicateLets(finalJs);
  finalJs = fixSwitchCase(finalJs);

  return { js: finalJs, errors };
}

module.exports = { compile, QuillCompiler };
