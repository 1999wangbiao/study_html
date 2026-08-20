'use strict';

// Markdown 渲染器 —— 纯函数，无状态、不碰 DOM 之外的东西。
// 内部 KW_* / tokenize / highlightCode / renderBlock 等均为局部，不导出。

import { esc } from './utils.js';
import { codeFileHtml } from './code-block.js';

function renderMarkdown(md) {
  const fullLines = String(md || '').replace(/\r\n/g, '\n').split('\n');
  const headings = [];
  let headingId = 0;

  // 各语言共享的关键字集合 —— C 家族（Go/Java/Cpp/JS/TS）的关键字比较通用，
  // 各自专属的几个再补一下。PowerShell / Bash 用常见的命令名。
  const KW_GO = new Set([
    'break','default','func','interface','select','case','defer','go','map','struct',
    'chan','else','goto','package','switch','const','fallthrough','if','range','type',
    'continue','for','import','return','var','nil','true','false','iota'
  ]);
  const KW_TS = new Set([
    'abstract','any','as','async','await','boolean','break','case','catch','class',
    'const','continue','debugger','default','delete','do','else','enum','export',
    'extends','false','finally','for','from','function','get','if','implements',
    'import','in','instanceof','interface','keyof','let','new','null','number',
    'of','private','protected','public','readonly','return','set','static','string',
    'super','switch','this','throw','true','try','type','typeof','undefined','var',
    'void','while','with','yield'
  ]);
  const KW_C = new Set([
    'alignas','alignof','auto','break','case','catch','class','const','constexpr',
    'continue','default','delete','do','else','enum','explicit','extern','false',
    'finally','for','friend','goto','if','inline','mutable','namespace','new','noexcept',
    'nullptr','operator','private','protected','public','register','return','sizeof',
    'static','struct','switch','template','this','throw','true','try','typedef',
    'typeid','typename','union','using','virtual','void','volatile','while'
  ]);
  const KW_JAVA = new Set([
    'abstract','assert','boolean','break','byte','case','catch','char','class','const',
    'continue','default','do','double','else','enum','extends','false','final','finally',
    'float','for','goto','if','implements','import','instanceof','int','interface',
    'long','native','new','null','package','private','protected','public','return',
    'short','static','strictfp','super','switch','synchronized','this','throw','throws',
    'transient','true','try','void','volatile','while','yield','var','record','sealed',
    'var'
  ]);
  const KW_BASH = new Set([
    'if','then','else','elif','fi','case','esac','for','select','while','until','do',
    'done','in','function','return','break','continue','time','coproc','export',
    'local','readonly','declare','set','unset','echo','printf','test','source','alias',
    'true','false'
  ]);
  const KW_POWERSHELL = new Set([
    'function','param','begin','process','end','if','else','elseif','switch','for',
    'foreach','while','do','until','return','break','continue','throw','try','catch',
    'finally','trap','in','not','and','or','band','bor','bnot','eq','ne','gt','lt',
    'ge','le','class','true','false','null'
  ]);
  // 类型与字面量：跟普通关键字合并到一个集合，渲染时同一类配色
  const TYPES_COMMON = new Set([
    'int','uint','uint8','uint16','uint32','uint64','int8','int16','int32','int64',
    'float','float32','float64','bool','byte','rune','string','error','byte','void',
    'size_t','ssize_t','char','short','long','double','Object','String','Array','List',
    'Map','Set','Vec','Option','Result','Box','Cell','RefCell','Rc','Arc','Mutex',
    'HashMap','String','Promise','number','boolean','object','undefined','any','unknown',
    'never'
  ]);
  function kwSetFor(lang) {
    switch (lang) {
      case 'go': return KW_GO;
      case 'ts': case 'js': case 'tsx': case 'jsx': return KW_TS;
      case 'c': case 'cpp': case 'cc': case 'cxx': case 'h': case 'hpp': return KW_C;
      case 'java': return KW_JAVA;
      case 'bash': case 'sh': case 'zsh': return KW_BASH;
      case 'powershell': case 'ps1': case 'pwsh': return KW_POWERSHELL;
      default: return null;
    }
  }
  function supportsStyle(lang) {
    return kwSetFor(lang) !== null || lang === 'json';
  }

  // 通用 token 扫描 —— 先收集位置再渲染，避免一边匹配一边替换导致冲突。
  // 返回有序的 [start, end, type]，type ∈ comment / string / number / keyword / type / function / bool / null
  function tokenize(code, lang) {
    const tokens = [];
    const kw = kwSetFor(lang);
    const len = code.length;
    let i2 = 0;
    const push = (a, b, t) => tokens.push([a, b, t]);

    while (i2 < len) {
      const ch = code[i2];

      // 行注释
      if ((lang === 'bash' || lang === 'sh' || lang === 'zsh' || lang === 'powershell' || lang === 'ps1' || lang === 'pwsh') && ch === '#') {
        const start = i2;
        while (i2 < len && code[i2] !== '\n') i2++;
        push(start, i2, 'comment');
        continue;
      }
      if (ch === '/' && code[i2 + 1] === '/') {
        const start = i2;
        while (i2 < len && code[i2] !== '\n') i2++;
        push(start, i2, 'comment');
        continue;
      }
      // 块注释 /* ... */
      if (ch === '/' && code[i2 + 1] === '*') {
        const start = i2;
        i2 += 2;
        while (i2 < len && !(code[i2] === '*' && code[i2 + 1] === '/')) i2++;
        i2 = Math.min(len, i2 + 2);
        push(start, i2, 'comment');
        continue;
      }
      // PowerShell 块注释 <# ... #>
      if (lang === 'powershell' && code.substr(i2, 2) === '<#') {
        const start = i2;
        i2 += 2;
        while (i2 < len && code.substr(i2, 2) !== '#>') i2++;
        i2 = Math.min(len, i2 + 2);
        push(start, i2, 'comment');
        continue;
      }

      // 字符串
      if (ch === '"' || ch === "'" || ch === '`') {
        const quote = ch;
        const start = i2;
        i2++;
        if (lang === 'json' && quote !== '"') {
          // JSON 严格要求双引号，但这里宽松一点不当作错
        }
        while (i2 < len) {
          if (code[i2] === '\\' && i2 + 1 < len) { i2 += 2; continue; }
          if (code[i2] === quote) { i2++; break; }
          if (code[i2] === '\n' && quote !== '`') { break; }
          i2++;
        }
        push(start, i2, 'string');
        continue;
      }

      // 数字
      if (/[0-9]/.test(ch) && (i2 === 0 || /[\s;,()\[\]{}<>+\-*/=%&|^!?:.]/.test(code[i2 - 1]))) {
        const start = i2;
        while (i2 < len && /[0-9a-fA-FxXoObB._]/.test(code[i2])) i2++;
        push(start, i2, 'number');
        continue;
      }

      // JSON 关键字 true/false/null
      if (lang === 'json') {
        const m = code.substr(i2).match(/^(true|false|null)\b/);
        if (m) {
          push(i2, i2 + m[1].length, 'bool');
          i2 += m[1].length;
          continue;
        }
      }

      // 标识符和关键字
      if (/[A-Za-z_]/.test(ch)) {
        const start = i2;
        while (i2 < len && /[A-Za-z0-9_]/.test(code[i2])) i2++;
        const word = code.slice(start, i2);
        let type = null;
        if (kw && kw.has(word)) type = 'keyword';
        else if (TYPES_COMMON.has(word)) type = 'type';
        else if (lang !== 'json' && code[i2] === '(') type = 'function';
        else if (lang !== 'json' && /^(true|false|nil|null|undefined|None|True|False)$/.test(word)) type = 'bool';
        if (type) push(start, i2, type);
        continue;
      }

      i2++;
    }
    return tokens;
  }

  function highlightCode(code, langRaw) {
    const lang = (langRaw || '').toLowerCase().trim();
    if (!lang || lang === 'text' || lang === 'plain' || lang === 'txt' || !supportsStyle(lang)) {
      // 无高亮也要保证转义
      return esc(code);
    }
    const tokens = tokenize(code, lang);
    let out = '';
    let last = 0;
    tokens.forEach(([a, b, type]) => {
      if (a < last) return; // 防重叠（不应该发生）
      out += esc(code.slice(last, a));
      out += `<span class="hl-${type}">${esc(code.slice(a, b))}</span>`;
      last = b;
    });
    out += esc(code.slice(last));
    return out;
  }

  const renderInline = (text) => {
    const tokens = [];
    let t = esc(text);
    t = t.replace(/`([^`]+)`/g, (m, code) => {
      tokens.push({ type: 'code', value: code });
      return ` ${tokens.length - 1} `;
    });
    t = t.replace(/!\[([^\]]*)\]\(([^)\s]+)(?:\s+"[^"]*")?\)/g, (m, alt, src) => {
      tokens.push({ type: 'img', alt, src });
      return ` ${tokens.length - 1} `;
    });
    t = t.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (m, label, href) => {
      tokens.push({ type: 'link', label, href });
      return ` ${tokens.length - 1} `;
    });
    t = t.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    t = t.replace(/(^|[^*])\*([^*\n]+)\*(?!\*)/g, '$1<em>$2</em>');
    return t.replace(/ (\d+) /g, (m, n) => {
      const token = tokens[Number(n)];
      if (!token) return m;
      if (token.type === 'code') return `<code>${token.value}</code>`;
      if (token.type === 'img') return `<img src="${esc(token.src)}" alt="${esc(token.alt)}" loading="lazy">`;
      return `<a href="${esc(token.href)}" target="_blank" rel="noopener">${token.label}</a>`;
    });
  };

  const isBlank = (s) => !s.trim();
  const isCodeFence = (s) => /^```/.test(s.trim());
  const isHeading = (s) => /^#{1,6}\s+/.test(s);
  const isHr = (s) => /^([-*_])\1{2,}$/.test(s.trim());
  const isUl = (s) => /^[-*+]\s+/.test(s);
  const isOl = (s) => /^\d+\.\s+/.test(s);
  const isBlockquote = (s) => /^>\s?/.test(s);
  const isTableRow = (s) => s.trim().startsWith('|');

  const parseTableRow = (line) => {
    let s = line.trim();
    if (s.startsWith('|')) s = s.slice(1);
    if (s.endsWith('|')) s = s.slice(0, -1);
    return s.split('|').map((c) => renderInline(c.trim()));
  };

  const isTableSeparator = (line) => {
    if (!line || !isTableRow(line)) return false;
    return parseTableRow(line).every((cell) => /^:?-+:?$/.test(cell.replace(/<[^>]+>/g, '').trim()));
  };

  // 核心行级渲染器，可被引用块递归调用
  const renderBlock = (inputLines, isInner) => {
    let lHtml = '';
    let li = 0;

    while (li < inputLines.length) {
      const line = inputLines[li];

      if (isBlank(line)) { li++; continue; }

      if (isCodeFence(line)) {
        const lang = line.trim().slice(3).trim();
        const buf = [];
        li++;
        while (li < inputLines.length && !isCodeFence(inputLines[li])) {
          buf.push(inputLines[li]);
          li++;
        }
        li++;
        const code = buf.join('\n');
        if (lang.toLowerCase() === 'mermaid') {
          lHtml += `<div class="mermaid">${esc(code)}</div>\n`;
        } else {
          lHtml += codeFileHtml(highlightCode(code, lang), lang) + '\n';
        }
        continue;
      }

      if (isHeading(line)) {
        const m = line.match(/^(#{1,6})\s+(.*)$/);
        const level = m[1].length;
        const rawText = m[2].trim();
        const id = `h-${++headingId}`;
        if (!isInner) headings.push({ level, id, text: rawText });
        lHtml += `<h${level} id="${id}">${renderInline(rawText)}</h${level}>\n`;
        li++;
        continue;
      }

      if (isHr(line)) {
        lHtml += '<hr>\n';
        li++;
        continue;
      }

      if (isBlockquote(line)) {
        const buf = [];
        while (li < inputLines.length && isBlockquote(inputLines[li])) {
          buf.push(inputLines[li].replace(/^>\s?/, ''));
          li++;
        }
        lHtml += `<blockquote>${renderBlock(buf, true)}</blockquote>\n`;
        continue;
      }

      if (isTableRow(line) && isTableSeparator(inputLines[li + 1])) {
        const header = parseTableRow(line);
        li += 2;
        const rows = [];
        while (li < inputLines.length && isTableRow(inputLines[li])) {
          rows.push(parseTableRow(inputLines[li]));
          li++;
        }
        lHtml += `<div class="table-wrap"><table><thead><tr>${header.map((h) => `<th>${h}</th>`).join('')}</tr></thead><tbody>${rows.map((r) => `<tr>${r.map((c) => `<td>${c}</td>`).join('')}</tr>`).join('')}</tbody></table></div>\n`;
        continue;
      }

      if (isUl(line)) {
        const items = [];
        while (li < inputLines.length) {
          const l = inputLines[li];
          if (isBlank(l)) { li++; continue; }
          const m = l.match(/^[-*+]\s+(.*)$/);
          if (m) { items.push(renderInline(m[1])); li++; }
          else if (/^\s+/.test(l) && items.length) { items[items.length - 1] += ' ' + renderInline(l.trim()); li++; }
          else break;
        }
        lHtml += `<ul>${items.map((it) => `<li>${it}</li>`).join('')}</ul>\n`;
        continue;
      }

      if (isOl(line)) {
        const items = [];
        while (li < inputLines.length) {
          const l = inputLines[li];
          if (isBlank(l)) { li++; continue; }
          const m = l.match(/^\d+\.\s+(.*)$/);
          if (m) { items.push(renderInline(m[1])); li++; }
          else if (/^\s+/.test(l) && items.length) { items[items.length - 1] += ' ' + renderInline(l.trim()); li++; }
          else break;
        }
        lHtml += `<ol>${items.map((it) => `<li>${it}</li>`).join('')}</ol>\n`;
        continue;
      }

      const buf = [];
      while (
        li < inputLines.length &&
        !isBlank(inputLines[li]) &&
        !isCodeFence(inputLines[li]) &&
        !isHeading(inputLines[li]) &&
        !isHr(inputLines[li]) &&
        !isBlockquote(inputLines[li]) &&
        !isUl(inputLines[li]) &&
        !isOl(inputLines[li]) &&
        !(isTableRow(inputLines[li]) && isTableSeparator(inputLines[li + 1]))
      ) {
        buf.push(inputLines[li]);
        li++;
      }
      lHtml += `<p>${renderInline(buf.join(' '))}</p>\n`;
    }
    return lHtml;
  };

  const html = renderBlock(fullLines, false);

  return { html, headings };
}

export { renderMarkdown };
