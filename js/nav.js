'use strict';

// 侧边栏嵌套导航树 + 首页分类卡片。
// 依赖 state（折叠偏好持久化）与 utils（图标/转义/articleUrl）。

import { $, $$, esc, icon, cssesc, articleUrl } from './utils.js';
import { state, isNodeCollapsed } from './state.js';

const CATEGORY_PALETTE = [
  '#8b5cf6', // purple
  '#06b6d4', // cyan
  '#f97316', // orange
  '#22c55e', // green
  '#ec4899', // pink
  '#3b82f6'  // blue
];

function categoryColor(name, idx) {
  const map = {
    '设计模式': '#8b5cf6',
    'Go语言': '#06b6d4',
    '常见问题': '#f97316',
    '面试知识': '#ec4899',
    '计算机基础': '#22c55e',
    '前端': '#3b82f6',
    '后端': '#06b6d4'
  };
  if (map[name]) return map[name];
  return CATEGORY_PALETTE[Math.abs(idx) % CATEGORY_PALETTE.length];
}

function catCardHtml(cat, idx, tagCount) {
  const color = categoryColor(cat.name, idx);
  return `
    <a class="cat-card" href="#/c/${cat.id}" style="--cat-color: ${color};">
      <span class="cat-card-arrow" aria-hidden="true">${icon('chevron-right', 14)}</span>
      <div class="cat-card-top">
        <span class="cat-card-icon">${icon(cat.icon || 'file-text', 20)}</span>
      </div>
      <div class="cat-card-name">${esc(cat.name)}</div>
      <div class="cat-card-desc">${esc(cat.description || '')}</div>
      <div class="cat-card-meta">
        <span class="cat-card-meta-item"><strong>${cat.articles.length}</strong> 篇文章</span>
        <span class="cat-card-meta-item"><strong>${tagCount}</strong> 个标签</span>
      </div>
    </a>`;
}

function matchesQuery(article, q) {
  const hay = [
    article.title,
    article.group,
    (article.tags || []).join(' ')
  ].join(' ').toLowerCase();
  return hay.includes(q);
}

function groupArticles(articles) {
  const groups = new Map();
  articles.forEach((a) => {
    const key = a.group || '全部';
    if (!groups.has(key)) groups.set(key, []);
    groups.get(key).push(a);
  });
  return groups;
}

/* ============ 嵌套导航树 ============ */

function countTreeArticles(node) {
  let n = (node.articles || []).length;
  (node.children || []).forEach((c) => { n += countTreeArticles(c); });
  return n;
}

function firstArticleUrl(node, catId) {
  const first = node.articles && node.articles[0];
  if (first) return articleUrl(catId, first.id);
  for (const c of node.children || []) {
    const url = firstArticleUrl(c, catId);
    if (url) return url;
  }
  return null;
}

function buildArticleTree(articles) {
  const root = { name: '', key: '', articles: [], children: [] };
  articles.forEach((a) => {
    const parts = Array.isArray(a.path) ? a.path : [];
    let cursor = root;
    let acc = '';
    for (const seg of parts) {
      acc += '/' + seg;
      let next = cursor.children.find((c) => c.name === seg);
      if (!next) {
        next = { name: seg, key: acc, articles: [], children: [] };
        cursor.children.push(next);
      }
      cursor = next;
    }
    cursor.articles.push(a);
  });
  const sortByName = (nodes) => {
    nodes.forEach((n) => {
      n.articles.sort((a, b) => String(a.title).localeCompare(String(b.title), 'zh'));
      sortByName(n.children);
    });
    nodes.sort((a, b) => a.name.localeCompare(b.name, 'zh'));
  };
  sortByName(root.children);
  root.articles.sort((a, b) => String(a.title).localeCompare(String(b.title), 'zh'));
  return root;
}

function renderArticleTree(node, catId, parentKey) {
  let out = '';
  // 子目录
  node.children.forEach((child) => {
    const nodeKey = parentKey + '/' + child.name;
    const key = `${catId}::${nodeKey}`;
    const total = countTreeArticles(child);
    const collapsed = isNodeCollapsed(child, key) ? ' collapsed' : '';
    const ariaExpanded = isNodeCollapsed(child, key) ? 'false' : 'true';
    // 二级及更深：标题点击也能折叠/展开，所以不再做跳转链接
    out += `
      <div class="nav-folder${collapsed}" data-node-key="${esc(key)}">
        <div class="nav-folder-head">
          <button class="nav-folder-toggle" type="button" aria-label="折叠或展开 ${esc(child.name)}" aria-expanded="${ariaExpanded}">${icon('chevron-down', 13)}</button>
          <span class="nav-folder-name nav-folder-name--toggleable" data-folder-toggle role="button" tabindex="0" title="${esc(child.name)}">${esc(child.name)}</span>
          <span class="nav-folder-count">${total}</span>
        </div>
        <div class="nav-folder-children">
          ${renderArticleTree(child, catId, nodeKey)}
        </div>
      </div>`;
  });
  // 直接的文章
  node.articles.forEach((a) => {
    const active = state.current.type === 'article' && state.current.categoryId === catId && state.current.articleId === a.id;
    out += `<a class="nav-link ${active ? 'active' : ''}" href="${articleUrl(catId, a.id)}">
      <span class="nav-link-icon">${icon(a.icon || 'file-text', 14)}</span>
      <span>${esc(a.title)}</span>
    </a>`;
  });
  return out;
}

function renderNavArticles(cat, articles) {
  const tree = buildArticleTree(articles);
  // 顶层直接的文章（path 为空）
  let out = '';
  tree.articles.forEach((a) => {
    const active = state.current.type === 'article' && state.current.categoryId === cat.id && state.current.articleId === a.id;
    out += `<a class="nav-link ${active ? 'active' : ''}" href="${articleUrl(cat.id, a.id)}">
      <span class="nav-link-icon">${icon(a.icon || 'file-text', 14)}</span>
      <span>${esc(a.title)}</span>
    </a>`;
  });
  // 子目录递归
  tree.children.forEach((child) => {
    const nodeKey = '/' + child.name;
    const key = `${cat.id}::${nodeKey}`;
    const url = firstArticleUrl(child, cat.id);
    const total = countTreeArticles(child);
    const collapsed = isNodeCollapsed(child, key) ? ' collapsed' : '';
    const ariaExpanded = isNodeCollapsed(child, key) ? 'false' : 'true';
    out += `
      <div class="nav-folder${collapsed}" data-node-key="${esc(key)}">
        <div class="nav-folder-head">
          <button class="nav-folder-toggle" type="button" aria-label="折叠或展开 ${esc(child.name)}" aria-expanded="${ariaExpanded}">${icon('chevron-down', 13)}</button>
          <a class="nav-folder-name" href="${esc(url || '#')}" title="${esc(child.name)}">${esc(child.name)}</a>
          <span class="nav-folder-count">${total}</span>
        </div>
        <div class="nav-folder-children">
          ${renderArticleTree(child, cat.id, nodeKey)}
        </div>
      </div>`;
  });
  return out;
}

function renderSidebar() {
  if (!state.config) return;
  const q = state.query.trim().toLowerCase();
  const html = (state.config.categories || []).map((cat) => {
    let articles = cat.articles || [];
    let catMatch = !q || cat.name.toLowerCase().includes(q);
    if (q) articles = articles.filter((a) => matchesQuery(a, q));
    if (q && !catMatch && !articles.length) return '';
    const open = !q || catMatch || articles.length > 0;
    const body = articles.length
      ? renderNavArticles(cat, articles)
      : (q ? '<div class="nav-empty">无匹配</div>' : '');
    return `<div class="nav-cat ${state.current.categoryId === cat.id ? 'active' : ''} ${open ? 'open' : ''}" data-cat="${cat.id}">
      <div class="nav-cat-head">
        <button class="nav-cat-go" type="button" data-go="${cat.id}">
          <span class="nav-icon">${icon(cat.icon || 'file-text', 15)}</span>
          <span class="nav-name">${esc(cat.name)}</span>
          <span class="nav-count">${cat.articles.length}</span>
        </button>
        <button class="nav-toggle" type="button" data-toggle="${cat.id}" aria-label="展开或收起 ${esc(cat.name)}">${icon('chevron-down', 16)}</button>
      </div>
      <div class="nav-body">${body}</div>
    </div>`;
  }).join('');
  $('#sidebar').innerHTML = `<div class="sidebar-inner">${html || '<div class="empty-state">没有匹配的分类</div>'}</div>`;
  highlightActiveFolders();
}

/* 高亮：当前文章所在的所有父级目录 */
function highlightActiveFolders() {
  $$('#sidebar .nav-folder.active-path').forEach((n) => n.classList.remove('active-path'));
  if (state.current.type !== 'article') return;
  const catId = state.current.categoryId;
  const articleId = state.current.articleId;
  const cat = (state.config.categories || []).find((c) => c.id === catId);
  const article = cat && (cat.articles || []).find((a) => a.id === articleId);
  if (!article || !Array.isArray(article.path)) return;
  // 从浅到深逐层构造 key
  let acc = '';
  article.path.forEach((seg) => {
    acc += '/' + seg;
    const key = `${catId}::${acc}`;
    const node = $(`#sidebar .nav-folder[data-node-key="${cssesc(key)}"]`);
    if (node) {
      node.classList.add('active-path');
      // 自动展开当前文章所在路径
      if (node.classList.contains('collapsed')) {
        node.classList.remove('collapsed');
      }
    }
  });
}

export { renderSidebar, highlightActiveFolders, categoryColor, catCardHtml, matchesQuery, groupArticles };
