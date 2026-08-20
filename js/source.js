'use strict';

// 文章正文加载 + 源码文件展示。
// getArticleMarkdown 优先读 indexMap 缓存（启动时 loadIndex 已预取），否则按需 fetch。

import { $, $$, esc, icon, copyTextToClipboard } from './utils.js';
import { state, indexMap } from './state.js';
import { codeFileHtml } from './code-block.js';

async function getArticleMarkdown(article) {
  if (indexMap.has(article.id)) return indexMap.get(article.id);
  if (article.content) return article.content;
  const url = article.file.split('/').map(encodeURIComponent).join('/');
  const rawUrl = article.file;
  const urlsToTry = [url];
  if (rawUrl !== url) urlsToTry.push(rawUrl);
  for (const tryUrl of urlsToTry) {
    try {
      const res = await fetch(tryUrl, { cache: 'no-store' });
      if (!res.ok) {
        if (res.status === 404) {
          return `# 加载失败\n\n**文件不存在**：\`${article.file}\`\n\n服务器返回 404。请确认：\n\n1. 服务器上是否存在 \`知识库/\` 目录\n2. 文件路径是否正确\n3. 运行 \`node build.js\` 重新生成配置`;
        }
        continue;
      }
      const text = await res.text();
      if (text && text.trim().length > 0) {
        if (/^\s*<!doctype\s|<html[\s>]/i.test(text.slice(0, 200))) {
          return `# 加载失败\n\n**服务器返回了 HTML 页面而非 Markdown 文件**。\n\n这通常是因为 Nginx 配置了 SPA 回退（\`try_files $uri /index.html\`），导致所有请求都被重定向到 \`index.html\`。\n\n**解决方案**：\n\n在 Nginx 配置中，为 \`.md\` 文件和 \`知识库/\` 目录添加正确的 location 规则，或者使用 \`server.js\`（Node.js）直接服务文件。`;
        }
        indexMap.set(article.id, text);
        return text;
      }
    } catch (e) {
      continue;
    }
  }
  return `# 加载失败\n\n无法读取 \`${article.file}\`\n\n已尝试路径：\n- 编码URL：\`${url}\`\n- 原始URL：\`${rawUrl}\`\n\n请检查：\n1. 服务器上 \`知识库/\` 目录是否存在\n2. 文件权限是否正确\n3. 在浏览器中直接访问 [${encodeURI(rawUrl)}](${rawUrl}) 查看返回内容`;
}

async function getArticleSourceFiles(article) {
  const slash = article.file ? article.file.lastIndexOf('/') : -1;
  const base = slash >= 0 ? article.file.slice(0, slash + 1) : '';
  const rawPaths = [];
  if (Array.isArray(article.code) || Array.isArray(article.codes)) {
    rawPaths.push(...(article.code || article.codes).slice());
  } else if (typeof article.code === 'string') {
    rawPaths.push(article.code);
  } else if (article.file) {
    rawPaths.push('main.go');
  }
  const paths = rawPaths.map((p) => (p.startsWith('/') ? p.slice(1) : base + p));
  const results = await Promise.all(paths.map(async (path, idx) => {
    try {
      const url = path.split('/').map(encodeURIComponent).join('/');
      const res = await fetch(url);
      if (!res.ok) return null;
      const text = await res.text();
      const name = path.split('/').pop() || path;
      const dot = name.lastIndexOf('.');
      const lang = dot > 0 ? name.slice(dot + 1) : 'text';
      const rel = rawPaths[idx];
      return { name, path: url, lang, text, rel };
    } catch (e) {
      return null;
    }
  }));
  return results.filter(Boolean);
}

function groupSourceFiles(files) {
  const groups = [];
  const byDir = new Map();
  files.forEach((f, i) => {
    const rel = f.rel || f.name;
    const parts = rel.split('/');
    const dir = parts.length > 1 ? parts.slice(0, -1).join('/') : '';
    if (!byDir.has(dir)) {
      byDir.set(dir, { dir, items: [] });
      groups.push(byDir.get(dir));
    }
    byDir.get(dir).items.push({ file: f, index: i });
  });
  return groups;
}

function sourceSectionHtml(files) {
  if (!files.length) return '';
  const current = files[0];
  // 多文件 source 用 tabs 区分，单文件就只展示一个 code-file
  // tabs 按文件夹分组：根目录文件归入“当前目录”，子目录各成一组
  const groups = groupSourceFiles(files);
  const tabs = files.length > 1 ? `
      <div class="source-tabs" role="tablist" aria-label="源码文件">
        ${groups.map((g) => `
          <div class="source-tab-group">
            <span class="source-tab-dir">${icon('folder', 12)} ${esc(g.dir || '当前目录')}</span>
            <div class="source-tab-items">
              ${g.items.map(({ file, index }) => `<button class="source-tab${index === 0 ? ' active' : ''}" type="button" role="tab" aria-selected="${index === 0}" data-source-index="${index}" title="${esc(file.rel || file.name)}">${esc(file.name)}</button>`).join('')}
            </div>
          </div>`).join('')}
      </div>` : '';
  // 多文件：每个文件一个 code-file 包，初始只显示第一个；标题显示相对路径便于区分同名文件
  const filesHtml = files.map((f, i) => `
    <div class="source-file-slot" data-source-slot="${i}" ${i === 0 ? '' : 'hidden'}>
      ${codeFileHtml(esc(f.text), f.lang, f.rel || f.name)}
    </div>`).join('');
  return `
    <section class="source-section">
      <div class="source-head">
        <h2>${icon('code', 18)} 源码</h2>
        <div class="source-actions">
          <a class="source-raw" href="${esc(current.path)}" target="_blank" rel="noopener">${icon('file-text', 14)} 原始文件</a>
          <button class="source-copy-all" type="button" title="复制源码" aria-label="复制源码">${icon('copy', 14)} 全部复制</button>
        </div>
      </div>${tabs}
      <div class="source-body">
        ${filesHtml}
      </div>
    </section>`;
}

function bindSourceControls(files) {
  if (!files.length) return;
  const root = $('.source-section');
  if (!root) return;
  const tabs = $$('.source-tab', root);
  const slots = $$('.source-file-slot', root);
  const raw = $('.source-raw', root);
  const copyAll = $('.source-copy-all', root);

  function activate(idx) {
    if (!files[idx]) return;
    const file = files[idx];
    slots.forEach((s, j) => {
      if (j === idx) s.removeAttribute('hidden');
      else s.setAttribute('hidden', '');
    });
    tabs.forEach((t, j) => {
      const active = j === idx;
      t.classList.toggle('active', active);
      t.setAttribute('aria-selected', String(active));
    });
    if (raw) raw.href = file.path;
  }

  tabs.forEach((tab) => {
    tab.addEventListener('click', () => activate(Number(tab.dataset.sourceIndex)));
  });

  if (copyAll) {
    copyAll.addEventListener('click', () => {
      const text = files.map((f) => `// ===== ${f.rel || f.name} =====\n\n${f.text}`).join('\n\n');
      copyTextToClipboard(text, copyAll, ' 全部复制');
    });
  }
}

export { getArticleMarkdown, getArticleSourceFiles, groupSourceFiles, sourceSectionHtml, bindSourceControls };
