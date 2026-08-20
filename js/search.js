'use strict';

// 全文搜索 —— 查询 + 渲染。搜索索引由 state.loadIndex 在启动时构建，
// 本模块保持纯"查询 + 渲染"，不负责 fetch。

import { $, esc, escapeRegExp, articleUrl } from './utils.js';
import { state } from './state.js';
import { renderSidebar } from './nav.js';

function searchArticles(q) {
  const query = q.trim().toLowerCase();
  if (!query) return [];
  const scored = [];
  state.searchIndex.forEach((a) => {
    const titleHit = a.title.toLowerCase().includes(query);
    const tagHit = (a.tags || []).some((t) => t.toLowerCase().includes(query));
    const textHit = a.plain.toLowerCase().includes(query);
    if (!titleHit && !tagHit && !textHit) return;
    const score = titleHit ? 0 : tagHit ? 1 : 2;
    scored.push({ article: a, score });
  });
  scored.sort((a, b) => a.score - b.score || a.article.title.length - b.article.title.length);
  return scored.slice(0, 12).map((s) => s.article);
}

function highlightText(text, q) {
  const idx = text.toLowerCase().indexOf(q.toLowerCase());
  if (idx < 0) return esc(text.slice(0, 160));
  const start = Math.max(0, idx - 42);
  const end = Math.min(text.length, idx + q.length + 110);
  const snippet = (start > 0 ? '…' : '') + text.slice(start, end) + (end < text.length ? '…' : '');
  return esc(snippet).replace(new RegExp(escapeRegExp(q), 'ig'), (m) => `<mark>${m}</mark>`);
}

function performSearch(q) {
  state.query = q;
  renderSidebar();
  const box = $('#search-results');
  if (!q.trim()) {
    box.hidden = true;
    box.innerHTML = '';
    return;
  }
  const hits = searchArticles(q);
  if (!hits.length) {
    box.innerHTML = `<div class="search-empty">没有找到与「${esc(q)}」相关的内容</div>`;
    box.hidden = false;
    return;
  }
  box.innerHTML = hits.map((a) => {
    const snippet = highlightText(a.plain, q);
    return `<a class="search-hit" href="${articleUrl(a.categoryId, a.id)}">
      <div class="search-hit-title"><span>${esc(a.title)}</span><span class="search-hit-cat">${esc(a.category.name)}</span></div>
      <div class="search-hit-snippet">${snippet}</div>
    </a>`;
  }).join('');
  box.hidden = false;
}

function closeSearch() {
  state.query = '';
  $('#search-input').value = '';
  $('#search-results').hidden = true;
  $('#search-results').innerHTML = '';
  renderSidebar();
}

export { searchArticles, highlightText, performSearch, closeSearch };
