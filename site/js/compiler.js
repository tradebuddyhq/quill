// ============================================
// Quill Compiler - JavaScript Port
// Lexer, Parser, Code Generator
// ============================================

(function () {
  'use strict';

  // ---- Token Types ----
  const T = {
    EOF: 'EOF', NEWLINE: 'NEWLINE', INDENT: 'INDENT', DEDENT: 'DEDENT',
    STRING: 'STRING', NUMBER: 'NUMBER', YES: 'YES', NO: 'NO', IDENT: 'IDENT',
    IS: 'IS', ARE: 'ARE', SAY: 'SAY', IF: 'IF', OTHERWISE: 'OTHERWISE',
    FOR: 'FOR', EACH: 'EACH', IN: 'IN', TO: 'TO', GIVE: 'GIVE', BACK: 'BACK',
    AND: 'AND', OR: 'OR', NOT: 'NOT', GREATER: 'GREATER', LESS: 'LESS',
    THAN: 'THAN', EQUAL: 'EQUAL', CONTAINS: 'CONTAINS', WHILE: 'WHILE',
    USE: 'USE', TEST: 'TEST', EXPECT: 'EXPECT', TRUE: 'TRUE', FALSE: 'FALSE',
    PLUS: 'PLUS', MINUS: 'MINUS', STAR: 'STAR', SLASH: 'SLASH', MODULO: 'MODULO',
    DOT: 'DOT', COLON: 'COLON', COMMA: 'COMMA',
    LPAREN: 'LPAREN', RPAREN: 'RPAREN', LBRACKET: 'LBRACKET', RBRACKET: 'RBRACKET',
  };

  const KEYWORDS = {
    'is': T.IS, 'are': T.ARE, 'say': T.SAY, 'if': T.IF, 'otherwise': T.OTHERWISE,
    'for': T.FOR, 'each': T.EACH, 'in': T.IN, 'to': T.TO, 'give': T.GIVE,
    'back': T.BACK, 'and': T.AND, 'or': T.OR, 'not': T.NOT,
    'greater': T.GREATER, 'less': T.LESS, 'than': T.THAN, 'equal': T.EQUAL,
    'contains': T.CONTAINS, 'while': T.WHILE, 'use': T.USE,
    'test': T.TEST, 'expect': T.EXPECT, 'yes': T.YES, 'no': T.NO,
    'true': T.TRUE, 'false': T.FALSE,
  };

  function token(type, value, line) {
    return { type, value, line };
  }

  // ---- Lexer ----
  function lex(source) {
    const tokens = [];
    const lines = source.split('\n');
    const indentStack = [0];

    for (let lineNum = 0; lineNum < lines.length; lineNum++) {
      const rawLine = lines[lineNum];

      // Skip blank lines and comment-only lines
      const trimmed = rawLine.trim();
      if (trimmed === '' || trimmed.startsWith('--')) {
        continue;
      }

      // Calculate indentation
      let spaces = 0;
      for (let i = 0; i < rawLine.length; i++) {
        if (rawLine[i] === ' ') spaces++;
        else if (rawLine[i] === '\t') spaces += 2;
        else break;
      }

      // Emit INDENT / DEDENT
      if (spaces > indentStack[indentStack.length - 1]) {
        indentStack.push(spaces);
        tokens.push(token(T.INDENT, null, lineNum + 1));
      } else {
        while (spaces < indentStack[indentStack.length - 1]) {
          indentStack.pop();
          tokens.push(token(T.DEDENT, null, lineNum + 1));
        }
      }

      // Tokenize the line content
      let pos = spaces;
      const line = rawLine;

      while (pos < line.length) {
        // Skip whitespace within line
        if (line[pos] === ' ' || line[pos] === '\t') {
          pos++;
          continue;
        }

        // Comment
        if (line[pos] === '-' && pos + 1 < line.length && line[pos + 1] === '-') {
          break; // rest of line is comment
        }

        // String
        if (line[pos] === '"') {
          let str = '';
          pos++; // skip opening quote
          while (pos < line.length && line[pos] !== '"') {
            if (line[pos] === '\\' && pos + 1 < line.length) {
              pos++;
              if (line[pos] === 'n') str += '\n';
              else if (line[pos] === 't') str += '\t';
              else if (line[pos] === '"') str += '"';
              else if (line[pos] === '\\') str += '\\';
              else str += line[pos];
            } else {
              str += line[pos];
            }
            pos++;
          }
          pos++; // skip closing quote
          tokens.push(token(T.STRING, str, lineNum + 1));
          continue;
        }

        // Number
        if (isDigit(line[pos]) || (line[pos] === '.' && pos + 1 < line.length && isDigit(line[pos + 1]))) {
          let num = '';
          while (pos < line.length && (isDigit(line[pos]) || line[pos] === '.')) {
            num += line[pos];
            pos++;
          }
          tokens.push(token(T.NUMBER, parseFloat(num), lineNum + 1));
          continue;
        }

        // Identifier / keyword
        if (isAlpha(line[pos])) {
          let ident = '';
          while (pos < line.length && isAlphaNum(line[pos])) {
            ident += line[pos];
            pos++;
          }
          const lower = ident.toLowerCase();
          if (KEYWORDS[lower] !== undefined) {
            tokens.push(token(KEYWORDS[lower], lower, lineNum + 1));
          } else {
            tokens.push(token(T.IDENT, ident, lineNum + 1));
          }
          continue;
        }

        // Single-char tokens
        const ch = line[pos];
        pos++;
        switch (ch) {
          case '+': tokens.push(token(T.PLUS, '+', lineNum + 1)); break;
          case '-': tokens.push(token(T.MINUS, '-', lineNum + 1)); break;
          case '*': tokens.push(token(T.STAR, '*', lineNum + 1)); break;
          case '/': tokens.push(token(T.SLASH, '/', lineNum + 1)); break;
          case '%': tokens.push(token(T.MODULO, '%', lineNum + 1)); break;
          case '.': tokens.push(token(T.DOT, '.', lineNum + 1)); break;
          case ':': tokens.push(token(T.COLON, ':', lineNum + 1)); break;
          case ',': tokens.push(token(T.COMMA, ',', lineNum + 1)); break;
          case '(': tokens.push(token(T.LPAREN, '(', lineNum + 1)); break;
          case ')': tokens.push(token(T.RPAREN, ')', lineNum + 1)); break;
          case '[': tokens.push(token(T.LBRACKET, '[', lineNum + 1)); break;
          case ']': tokens.push(token(T.RBRACKET, ']', lineNum + 1)); break;
          default:
            throw new Error(`Unexpected character '${ch}' on line ${lineNum + 1}`);
        }
      }

      tokens.push(token(T.NEWLINE, null, lineNum + 1));
    }

    // Close remaining indents
    while (indentStack.length > 1) {
      indentStack.pop();
      tokens.push(token(T.DEDENT, null, lines.length));
    }

    tokens.push(token(T.EOF, null, lines.length));
    return tokens;
  }

  function isDigit(c) { return c >= '0' && c <= '9'; }
  function isAlpha(c) { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c === '_'; }
  function isAlphaNum(c) { return isAlpha(c) || isDigit(c); }

  // ---- Parser ----
  function parse(tokens) {
    let pos = 0;

    function peek() { return tokens[pos] || token(T.EOF, null, 0); }
    function advance() { return tokens[pos++]; }
    function check(type) { return peek().type === type; }

    function expect(type) {
      if (!check(type)) {
        const t = peek();
        throw new Error(`Expected ${type} but got ${t.type} ('${t.value}') on line ${t.line}`);
      }
      return advance();
    }

    function match(...types) {
      if (types.includes(peek().type)) {
        return advance();
      }
      return null;
    }

    function skipNewlines() {
      while (check(T.NEWLINE)) advance();
    }

    // program → statement* EOF
    function program() {
      const stmts = [];
      skipNewlines();
      while (!check(T.EOF)) {
        stmts.push(statement());
        skipNewlines();
      }
      return { type: 'Program', body: stmts };
    }

    function statement() {
      skipNewlines();

      if (check(T.SAY)) return sayStmt();
      if (check(T.IF)) return ifStmt();
      if (check(T.FOR)) return forStmt();
      if (check(T.WHILE)) return whileStmt();
      if (check(T.TO)) return funcDef();
      if (check(T.GIVE)) return returnStmt();
      if (check(T.TEST)) return testBlock();
      if (check(T.EXPECT)) return expectStmt();
      if (check(T.USE)) return useStmt();

      // Could be assignment or expression statement
      // Check: IDENT (IS|ARE) ...
      if (check(T.IDENT) && pos + 1 < tokens.length &&
          (tokens[pos + 1].type === T.IS || tokens[pos + 1].type === T.ARE)) {
        // But we need to distinguish assignment from comparison:
        // "x is 5" at statement level = assignment
        // Only if the next token after IS/ARE is not GREATER/LESS/EQUAL/NOT (that would be comparison in expr)
        // Peek further ahead
        const afterIs = tokens[pos + 2];
        if (afterIs && (afterIs.type === T.GREATER || afterIs.type === T.LESS ||
            afterIs.type === T.EQUAL || afterIs.type === T.NOT)) {
          // This is a comparison expression statement
          return exprStmt();
        }
        return assignment();
      }

      return exprStmt();
    }

    function assignment() {
      const name = expect(T.IDENT);
      match(T.IS) || expect(T.ARE);
      const value = expression();
      match(T.NEWLINE);
      return { type: 'Assignment', name: name.value, value, line: name.line };
    }

    function sayStmt() {
      const tok = expect(T.SAY);
      const value = expression();
      match(T.NEWLINE);
      return { type: 'Say', value, line: tok.line };
    }

    function ifStmt() {
      const tok = expect(T.IF);
      const condition = expression();
      expect(T.COLON);
      match(T.NEWLINE);
      expect(T.INDENT);
      const body = block();

      const elseIfs = [];
      const elseBody = [];

      while (check(T.OTHERWISE)) {
        advance();
        if (check(T.IF)) {
          advance();
          const cond = expression();
          expect(T.COLON);
          match(T.NEWLINE);
          expect(T.INDENT);
          elseIfs.push({ condition: cond, body: block() });
        } else {
          expect(T.COLON);
          match(T.NEWLINE);
          expect(T.INDENT);
          elseBody.push(...block());
          break;
        }
      }

      return { type: 'If', condition, body, elseIfs, elseBody, line: tok.line };
    }

    function forStmt() {
      const tok = expect(T.FOR);
      expect(T.EACH);
      const varName = expect(T.IDENT);
      expect(T.IN);
      const iterable = expression();
      expect(T.COLON);
      match(T.NEWLINE);
      expect(T.INDENT);
      const body = block();
      return { type: 'ForEach', variable: varName.value, iterable, body, line: tok.line };
    }

    function whileStmt() {
      const tok = expect(T.WHILE);
      const condition = expression();
      expect(T.COLON);
      match(T.NEWLINE);
      expect(T.INDENT);
      const body = block();
      return { type: 'While', condition, body, line: tok.line };
    }

    function funcDef() {
      const tok = expect(T.TO);
      const name = expect(T.IDENT);
      const params = [];
      while (check(T.IDENT)) {
        params.push(advance().value);
      }
      expect(T.COLON);
      match(T.NEWLINE);
      expect(T.INDENT);
      const body = block();
      return { type: 'FuncDef', name: name.value, params, body, line: tok.line };
    }

    function returnStmt() {
      const tok = expect(T.GIVE);
      expect(T.BACK);
      const value = expression();
      match(T.NEWLINE);
      return { type: 'Return', value, line: tok.line };
    }

    function testBlock() {
      const tok = expect(T.TEST);
      const name = expect(T.STRING);
      expect(T.COLON);
      match(T.NEWLINE);
      expect(T.INDENT);
      const body = block();
      return { type: 'Test', name: name.value, body, line: tok.line };
    }

    function expectStmt() {
      const tok = expect(T.EXPECT);
      const expr = expression();
      match(T.NEWLINE);
      return { type: 'Expect', expression: expr, line: tok.line };
    }

    function useStmt() {
      const tok = expect(T.USE);
      const path = expect(T.STRING);
      match(T.NEWLINE);
      return { type: 'Use', path: path.value, line: tok.line };
    }

    function exprStmt() {
      const expr = expression();
      match(T.NEWLINE);
      return { type: 'ExprStmt', expression: expr };
    }

    function block() {
      const stmts = [];
      skipNewlines();
      while (!check(T.DEDENT) && !check(T.EOF)) {
        stmts.push(statement());
        skipNewlines();
      }
      match(T.DEDENT);
      return stmts;
    }

    // ---- Expressions ----
    function expression() {
      return orExpr();
    }

    function orExpr() {
      let left = andExpr();
      while (check(T.OR)) {
        advance();
        left = { type: 'BinaryOp', op: '||', left, right: andExpr() };
      }
      return left;
    }

    function andExpr() {
      let left = notExpr();
      while (check(T.AND)) {
        advance();
        left = { type: 'BinaryOp', op: '&&', left, right: notExpr() };
      }
      return left;
    }

    function notExpr() {
      if (check(T.NOT)) {
        advance();
        return { type: 'UnaryOp', op: '!', operand: notExpr() };
      }
      return comparison();
    }

    function comparison() {
      let left = addition();

      if (check(T.CONTAINS)) {
        advance();
        const right = addition();
        return { type: 'BinaryOp', op: '__contains', left, right };
      }

      if (check(T.IS) || check(T.ARE)) {
        advance();
        if (check(T.GREATER)) {
          advance();
          expect(T.THAN);
          return { type: 'BinaryOp', op: '>', left, right: addition() };
        }
        if (check(T.LESS)) {
          advance();
          expect(T.THAN);
          return { type: 'BinaryOp', op: '<', left, right: addition() };
        }
        if (check(T.EQUAL)) {
          advance();
          expect(T.TO);
          return { type: 'BinaryOp', op: '===', left, right: addition() };
        }
        if (check(T.NOT)) {
          advance();
          return { type: 'BinaryOp', op: '!==', left, right: addition() };
        }
        // Equality shorthand: x is y
        return { type: 'BinaryOp', op: '===', left, right: addition() };
      }

      return left;
    }

    function addition() {
      let left = multiplication();
      while (check(T.PLUS) || check(T.MINUS)) {
        const op = advance().type === T.PLUS ? '+' : '-';
        left = { type: 'BinaryOp', op, left, right: multiplication() };
      }
      return left;
    }

    function multiplication() {
      let left = unary();
      while (check(T.STAR) || check(T.SLASH) || check(T.MODULO)) {
        const t = advance();
        const op = t.type === T.STAR ? '*' : t.type === T.SLASH ? '/' : '%';
        left = { type: 'BinaryOp', op, left, right: unary() };
      }
      return left;
    }

    function unary() {
      if (check(T.MINUS)) {
        advance();
        return { type: 'UnaryOp', op: '-', operand: unary() };
      }
      return postfix();
    }

    function postfix() {
      let expr = primary();

      while (true) {
        if (check(T.DOT)) {
          advance();
          const prop = expect(T.IDENT);
          expr = { type: 'DotAccess', object: expr, property: prop.value };
        } else if (check(T.LBRACKET)) {
          advance();
          const index = expression();
          expect(T.RBRACKET);
          expr = { type: 'IndexAccess', object: expr, index };
        } else if (check(T.LPAREN)) {
          advance();
          const args = [];
          if (!check(T.RPAREN)) {
            args.push(expression());
            while (check(T.COMMA)) {
              advance();
              args.push(expression());
            }
          }
          expect(T.RPAREN);
          expr = { type: 'FuncCall', callee: expr, args };
        } else {
          break;
        }
      }

      return expr;
    }

    function primary() {
      if (check(T.STRING)) {
        const t = advance();
        return { type: 'String', value: t.value };
      }
      if (check(T.NUMBER)) {
        const t = advance();
        return { type: 'Number', value: t.value };
      }
      if (check(T.YES) || check(T.TRUE)) {
        advance();
        return { type: 'Boolean', value: true };
      }
      if (check(T.NO) || check(T.FALSE)) {
        advance();
        return { type: 'Boolean', value: false };
      }
      if (check(T.IDENT)) {
        const t = advance();
        return { type: 'Identifier', name: t.value };
      }
      if (check(T.LBRACKET)) {
        advance();
        const elements = [];
        if (!check(T.RBRACKET)) {
          elements.push(expression());
          while (check(T.COMMA)) {
            advance();
            elements.push(expression());
          }
        }
        expect(T.RBRACKET);
        return { type: 'List', elements };
      }
      if (check(T.LPAREN)) {
        advance();
        const expr = expression();
        expect(T.RPAREN);
        return expr;
      }

      const t = peek();
      throw new Error(`Unexpected token ${t.type} ('${t.value}') on line ${t.line}`);
    }

    return program();
  }

  // ---- Code Generator ----
  function generate(ast) {
    const declaredVars = new Set();
    let indent = 0;

    function ind() { return '  '.repeat(indent); }

    function genNode(node) {
      switch (node.type) {
        case 'Program':
          return node.body.map(genNode).join('\n');

        case 'Assignment': {
          const val = genExpr(node.value);
          if (declaredVars.has(node.name)) {
            return `${ind()}${node.name} = ${val};`;
          }
          declaredVars.add(node.name);
          return `${ind()}let ${node.name} = ${val};`;
        }

        case 'Say':
          return `${ind()}console.log(${genExpr(node.value)});`;

        case 'If': {
          let code = `${ind()}if (${genExpr(node.condition)}) {\n`;
          indent++;
          code += node.body.map(genNode).join('\n') + '\n';
          indent--;
          for (const ei of node.elseIfs) {
            code += `${ind()}} else if (${genExpr(ei.condition)}) {\n`;
            indent++;
            code += ei.body.map(genNode).join('\n') + '\n';
            indent--;
          }
          if (node.elseBody.length > 0) {
            code += `${ind()}} else {\n`;
            indent++;
            code += node.elseBody.map(genNode).join('\n') + '\n';
            indent--;
          }
          code += `${ind()}}`;
          return code;
        }

        case 'ForEach': {
          let code = `${ind()}for (const ${node.variable} of ${genExpr(node.iterable)}) {\n`;
          indent++;
          code += node.body.map(genNode).join('\n') + '\n';
          indent--;
          code += `${ind()}}`;
          return code;
        }

        case 'While': {
          let code = `${ind()}while (${genExpr(node.condition)}) {\n`;
          indent++;
          code += node.body.map(genNode).join('\n') + '\n';
          indent--;
          code += `${ind()}}`;
          return code;
        }

        case 'FuncDef': {
          declaredVars.add(node.name);
          const params = node.params.join(', ');
          let code = `${ind()}function ${node.name}(${params}) {\n`;
          indent++;
          const outerVars = new Set(declaredVars);
          node.params.forEach(p => declaredVars.add(p));
          code += node.body.map(genNode).join('\n') + '\n';
          // Restore outer scope minus what we want to keep
          node.params.forEach(p => { if (!outerVars.has(p)) declaredVars.delete(p); });
          indent--;
          code += `${ind()}}`;
          return code;
        }

        case 'Return':
          return `${ind()}return ${genExpr(node.value)};`;

        case 'Test': {
          const nameStr = JSON.stringify(node.name);
          let code = `${ind()}try {\n`;
          indent++;
          code += node.body.map(genNode).join('\n') + '\n';
          code += `${ind()}__test_passed++;\n`;
          code += `${ind()}console.log("  \\u2713 " + ${nameStr});\n`;
          indent--;
          code += `${ind()}} catch (__e) {\n`;
          indent++;
          code += `${ind()}__test_failed++;\n`;
          code += `${ind()}console.log("  \\u2717 " + ${nameStr} + ": " + __e.message);\n`;
          indent--;
          code += `${ind()}}`;
          return code;
        }

        case 'Expect': {
          const exprCode = genExpr(node.expression);
          const exprStr = JSON.stringify(exprCode);
          return `${ind()}if (!(${exprCode})) throw new Error("Expected " + ${exprStr} + " to be true");`;
        }

        case 'Use':
          return `${ind()}// use "${node.path}" (imports not supported in browser)`;

        case 'ExprStmt':
          return `${ind()}${genExpr(node.expression)};`;

        default:
          return `${ind()}/* unknown node: ${node.type} */`;
      }
    }

    function genExpr(node) {
      switch (node.type) {
        case 'Number':
          return String(node.value);

        case 'Boolean':
          return node.value ? 'true' : 'false';

        case 'String': {
          // Handle string interpolation: {var} becomes ${var}
          const val = node.value;
          if (val.includes('{')) {
            // Convert {expr} to ${expr}
            const escaped = val.replace(/\\/g, '\\\\').replace(/`/g, '\\`');
            const interpolated = escaped.replace(/\{([^}]+)\}/g, '${$1}');
            return '`' + interpolated + '`';
          }
          return JSON.stringify(val);
        }

        case 'Identifier':
          return node.name;

        case 'List':
          return '[' + node.elements.map(genExpr).join(', ') + ']';

        case 'BinaryOp':
          if (node.op === '__contains') {
            return `__contains(${genExpr(node.left)}, ${genExpr(node.right)})`;
          }
          return `(${genExpr(node.left)} ${node.op} ${genExpr(node.right)})`;

        case 'UnaryOp':
          return `(${node.op}${genExpr(node.operand)})`;

        case 'DotAccess':
          return `${genExpr(node.object)}.${node.property}`;

        case 'IndexAccess':
          return `${genExpr(node.object)}[${genExpr(node.index)}]`;

        case 'FuncCall': {
          const callee = genExpr(node.callee);
          const args = node.args.map(genExpr).join(', ');
          return `${callee}(${args})`;
        }

        default:
          return `/* unknown expr: ${node.type} */`;
      }
    }

    return genNode(ast);
  }

  // ---- Standard Library ----
  const STDLIB = `const __contains = (a, b) => { if (typeof a === 'string') return a.includes(b); if (Array.isArray(a)) return a.includes(b); return false; };
let __test_passed = 0;
let __test_failed = 0;
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
`;

  // ---- Public API ----
  function compile(source) {
    const tokens = lex(source);
    const ast = parse(tokens);
    const js = generate(ast);
    return STDLIB + '\n' + js + `
if (__test_passed + __test_failed > 0) {
  console.log("");
  console.log(__test_passed + " passed, " + __test_failed + " failed");
}`;
  }

  window.QuillCompiler = { compile };
})();
