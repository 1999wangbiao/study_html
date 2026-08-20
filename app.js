'use strict';

const $ = (sel, root = document) => root.querySelector(sel);
const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));

const ICONS = {
  'chevron-down': '<polyline points="6 9 12 15 18 9"></polyline>',
  'chevron-up': '<polyline points="18 15 12 9 6 15"></polyline>',
  'chevron-right': '<polyline points="9 18 15 12 9 6"></polyline>',
  'arrow-left': '<path d="M19 12H5"></path><path d="m12 19-7-7 7-7"></path>',
  'arrow-right': '<path d="M5 12h14"></path><path d="m12 5 7 7-7 7"></path>',
  'copy': '<rect width="14" height="14" x="8" y="8" rx="2" ry="2"></rect><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"></path>',
  'check': '<path d="M20 6 9 17l-5-5"></path>',
  'calendar': '<rect width="18" height="18" x="3" y="4" rx="2"></rect><path d="M16 2v4M8 2v4M3 10h18"></path>',
  'clock': '<circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline>',
  'file-text': '<path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"></path><path d="M14 2v4a2 2 0 0 0 2 2h4"></path><path d="M10 9H8M16 13H8M16 17H8"></path>',
  'home': '<path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline>',
  'sun': '<circle cx="12" cy="12" r="4"></circle><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"></path>',
  'moon': '<path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"></path>',
  'code': '<polyline points="16 18 22 12 16 6"></polyline><polyline points="8 6 2 12 8 18"></polyline>',
  'server': '<rect width="20" height="8" x="2" y="2" rx="2" ry="2"></rect><rect width="20" height="8" x="2" y="14" rx="2" ry="2"></rect><line x1="6" x2="6.01" y1="6" y2="6"></line><line x1="6" x2="6.01" y1="18" y2="18"></line>',
  'cpu': '<rect width="16" height="16" x="4" y="4" rx="2"></rect><rect width="6" height="6" x="9" y="9"></rect><path d="M15 2v2M15 20v2M2 15h2M2 9h2M20 15h2M20 9h2M9 2v2M9 20v2"></path>',
  'layers': '<path d="m12 2 10 6.5-10 6.5L2 8.5 12 2z"></path><path d="m2 15.5 10 6.5 10-6.5"></path>',
  'network': '<rect x="9" y="2" width="6" height="6" rx="1"></rect><rect x="2" y="16" width="6" height="6" rx="1"></rect><rect x="16" y="16" width="6" height="6" rx="1"></rect><path d="M12 8v4M7 19h2M15 19h2M12 12l-5 4M12 12l5 4"></path>',
  'database': '<ellipse cx="12" cy="5" rx="9" ry="3"></ellipse><path d="M3 5v14a9 3 0 0 0 18 0V5"></path><path d="M3 12a9 3 0 0 0 18 0"></path>',
  'tag': '<path d="M12 2H2v10l9.29 9.29a1 1 0 0 0 1.42 0l8.58-8.58a1 1 0 0 0 0-1.42L12 2Z"></path><circle cx="7" cy="7" r="1.5"></circle>',
  'folder': '<path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"></path>'
};

function icon(name, size = 16) {
  const body = ICONS[name] || ICONS['file-text'];
  return `<svg class="icon" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${body}</svg>`;
}

function esc(s) {
  return String(s == null ? '' : s).replace(/[&<>"']/g, (c) => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;'
  }[c]));
}

function escapeRegExp(s) {
  return String(s).replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

async function copyTextToClipboard(text, btn, label = '') {
  const show = (name) => {
    btn.classList.add('copied');
    btn.innerHTML = `${icon(name, 14)}${label}`;
    setTimeout(() => {
      btn.classList.remove('copied');
      btn.innerHTML = `${icon('copy', 14)}${label}`;
    }, 1200);
  };
  try {
    await navigator.clipboard.writeText(text);
    show('check');
  } catch (e) {
    const ta = document.createElement('textarea');
    ta.value = text;
    document.body.appendChild(ta);
    ta.select();
    document.execCommand('copy');
    ta.remove();
    show('check');
  }
}

const state = {
  config: null,
  flatArticles: [],
  searchIndex: [],
  current: { type: 'home', categoryId: null, articleId: null },
  query: '',
  // 用户对每个文件夹的「显式切换偏好」：true = 主动收过，false = 主动展过；未记录则用默认规则。
  userOverrides: loadUserOverrides()
};

// 上一次用户主动切换过的记录会写回 localStorage，
// 下次加载时覆盖默认规则 —— 但默认规则（叶子文件夹默认折叠）依旧生效于从未被用户碰过的节点。
function loadUserOverrides() {
  try {
    const obj = JSON.parse(localStorage.getItem('kb-nav-overrides') || '{}');
    const m = new Map();
    Object.entries(obj).forEach(([k, v]) => {
      if (typeof v === 'boolean') m.set(k, v);
    });
    return m;
  } catch (e) {
    return new Map();
  }
}

function saveUserOverrides() {
  try {
    const obj = {};
    state.userOverrides.forEach((v, k) => { obj[k] = v; });
    localStorage.setItem('kb-nav-overrides', JSON.stringify(obj));
  } catch (e) { /* ignore */ }
}

// 默认规则：只要这个文件夹直接挂着文章（叶子文章），就视为"末端文件夹"并默认收起，
// 它的上级因为没有挂文章（只挂子目录）所以保持展开 —— 这样用户先看见骨架，再点开看具体文章。
function defaultCollapsed(node) {
  return !!(node && node.articles && node.articles.length > 0);
}

function isNodeCollapsed(node, key) {
  if (state.userOverrides.has(key)) return state.userOverrides.get(key) === true;
  return defaultCollapsed(node);
}

const indexMap = new Map();

function applySiteConfig() {
  const site = state.config.site || {};
  document.title = site.title || '计算机知识库';
  $('#site-title').textContent = site.title || '';
  $('#site-subtitle').textContent = site.subtitle || '';
  $('#site-footer').textContent = site.footer || '';
  if (site.accent) {
    document.documentElement.style.setProperty('--accent', site.accent);
  }
}

function buildFlatArticles() {
  state.flatArticles = [];
  (state.config.categories || []).forEach((cat) => {
    (cat.articles || []).forEach((article) => {
      state.flatArticles.push({ ...article, categoryId: cat.id, category: cat });
    });
  });
}

function articleUrl(categoryId, articleId) {
  return `#/a/${categoryId}/${articleId}`;
}

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

function cssesc(s) {
  return String(s).replace(/["\\]/g, '\\$&');
}

function setHomeActive(active) {
  document.body.classList.toggle('home-active', !!active);
}

function renderHome() {
  state.current = { type: 'home', categoryId: null, articleId: null };
  document.title = state.config.site.title || '计算机知识库';
  setHomeActive(true);
  $('#toc').innerHTML = '';
  renderSidebar();

  const cats = state.config.categories || [];
  const tagSet = new Set();
  cats.forEach((c) => (c.articles || []).forEach((a) => (a.tags || []).forEach((t) => tagSet.add(t))));
  const recent = [...state.flatArticles]
    .sort((a, b) => String(b.updated || '').localeCompare(String(a.updated || '')))
    .slice(0, 6);

  const catCards = cats.map((cat, i) => {
    const catTags = new Set();
    (cat.articles || []).forEach((a) => (a.tags || []).forEach((t) => catTags.add(t)));
    return catCardHtml(cat, i, catTags.size);
  }).join('');

  const stats = [
    { n: cats.length, label: '个分类' },
    { n: state.flatArticles.length, label: '篇文章' },
    { n: tagSet.size, label: '个标签' }
  ];

  $('#content').innerHTML = `
    <div class="page home-page">
      <header class="page-head">
        <div class="eyebrow">${icon('home', 16)} 知识总览</div>
        <h1>${esc(state.config.site.title)}</h1>
        <p>${esc(state.config.site.description || '')}</p>
        <div class="stat-strip">${stats.map((s) => `<span><strong>${s.n}</strong> ${s.label}</span>`).join('')}</div>
      </header>

      <div class="section-title">知识分类<span class="chev">${icon('chevron-down', 16)}</span></div>
      <div class="cat-grid">
        ${catCards}
      </div>

      <div class="section-title">最近更新<span class="chev">${icon('chevron-down', 16)}</span></div>
      <div class="recent-list">
        ${recent.map((a) => articleRowHtml(a.categoryId, a)).join('') || '<div class="empty-state">暂无文章</div>'}
      </div>
    </div>`;
  window.scrollTo(0, 0);
  setupReveal();
}

function articleRowHtml(categoryId, a) {
  return `<a class="recent-row" href="${articleUrl(categoryId, a.id)}">
    <div class="recent-row-main">
      <div class="recent-row-title">${esc(a.title)}</div>
      <div class="recent-row-meta">${a.updated ? esc(a.updated) : ''}</div>
      <div class="recent-row-tags">${(a.tags || []).slice(0, 3).map((t) => `<span class="tag">${esc(t)}</span>`).join('')}</div>
    </div>
    <span class="recent-row-arrow">${icon('chevron-right', 16)}</span>
  </a>`;
}

function renderCategory(categoryId) {
  const cat = (state.config.categories || []).find((c) => c.id === categoryId);
  if (!cat) {
    renderNotFound(`分类「${esc(categoryId)}」不存在`);
    return;
  }
  state.current = { type: 'category', categoryId, articleId: null };
  setHomeActive(false);
  document.title = `${cat.name} · ${state.config.site.title}`;
  $('#toc').innerHTML = '';
  renderSidebar();
  closeSearch();

  const tagSet = new Set((cat.articles || []).flatMap((a) => a.tags || []));
  let groupsHtml = '';
  groupArticles(cat.articles || []).forEach((list, groupName) => {
    groupsHtml += `
      <div class="section-title">${esc(groupName)}</div>
      <div class="recent-list">${list.map((a) => articleRowHtml(cat.id, a)).join('')}</div>`;
  });

  $('#content').innerHTML = `
    <div class="page category-page">
      <header class="page-head">
        <div class="eyebrow">${icon(cat.icon || 'file-text', 16)} 分类</div>
        <h1>${esc(cat.name)}</h1>
        <p>${esc(cat.description || '')}</p>
        <div class="stat-strip">
          <span class="stat-chip"><strong>${cat.articles.length}</strong> 篇文章</span>
          <span class="stat-chip"><strong>${tagSet.size}</strong> 个标签</span>
        </div>
      </header>
      ${groupsHtml}
    </div>`;
  window.scrollTo(0, 0);
  setupReveal();
}

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
    rawPaths.push(base + 'main.go');
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

function stripLeadingH1(md) {
  const trimmed = md.replace(/^\uFEFF?/, '').trimStart();
  const lines = trimmed.split('\n');
  if (/^#\s+/.test(lines[0])) return lines.slice(1).join('\n').trimStart();
  return trimmed;
}

function navLinkHtml(article, label, cls) {
  const arrow = icon(cls === 'next' ? 'arrow-right' : 'arrow-left', 14);
  return `<a class="article-nav-link ${cls}" href="${articleUrl(article.categoryId, article.id)}">
    <span class="article-nav-label">${cls === 'next' ? `${label} ${arrow}` : `${arrow} ${label}`}</span>
    <span class="article-nav-title">${esc(article.title)}</span>
  </a>`;
}

async function renderArticle(categoryId, articleId, anchor) {
  const cat = (state.config.categories || []).find((c) => c.id === categoryId);
  const article = cat && (cat.articles || []).find((a) => a.id === articleId);
  if (!cat || !article) {
    renderNotFound(`文章「${esc(articleId)}」不存在`);
    return;
  }

  state.current = { type: 'article', categoryId, articleId };
  setHomeActive(false);
  document.title = `${article.title} · ${state.config.site.title}`;
  $('#toc').innerHTML = '';
  renderSidebar();
  closeSearch();
  closeSidebar();

  const md = await getArticleMarkdown(article);
  const sourceFiles = await getArticleSourceFiles(article);
  const cleanMd = stripLeadingH1(md);
  const { html, headings } = renderMarkdown(cleanMd);
  const idx = state.flatArticles.findIndex((a) => a.id === articleId && a.categoryId === categoryId);
  const prev = idx > 0 ? state.flatArticles[idx - 1] : null;
  const next = idx >= 0 && idx < state.flatArticles.length - 1 ? state.flatArticles[idx + 1] : null;
  const wordCount = cleanMd.replace(/\s+/g, '').length;
  const minutes = Math.max(1, Math.round(wordCount / 420));

  $('#content').innerHTML = `
    <div class="page article-page">
      <nav class="breadcrumb">
        <a href="#/">${icon('home', 14)} 首页</a>
        <span>/</span>
        <a href="#/c/${cat.id}">${esc(cat.name)}</a>
        <span>/</span>
        <span>${esc(article.title)}</span>
      </nav>
      <header class="article-header">
        <div class="article-meta">
          <span class="tag tag-accent">${esc(cat.name)}</span>
          ${article.updated ? `<span class="article-meta-item">${icon('calendar', 14)} ${esc(article.updated)}</span>` : ''}
          <span class="article-meta-item">${icon('clock', 14)} 约 ${minutes} 分钟</span>
          <span class="article-meta-item">${icon('file-text', 14)} ${wordCount} 字</span>
          ${(article.tags || []).map((t) => `<span class="tag">${esc(t)}</span>`).join('')}
        </div>
        <h1>${esc(article.title)}</h1>
      </header>
      <div class="md-body">${html}</div>
      ${sourceSectionHtml(sourceFiles)}
      <nav class="article-nav">
        ${prev ? navLinkHtml(prev, '上一篇', 'prev') : '<span></span>'}
        ${next ? navLinkHtml(next, '下一篇', 'next') : '<span></span>'}
      </nav>
    </div>`;

  renderToc(headings);
  attachCopyButtons();
  bindCodeFiles();
  initScrollSpy();
  renderMermaidBlocks();
  bindSourceControls(sourceFiles);
  updateProgress();

  requestAnimationFrame(() => {
    if (anchor) {
      const el = document.getElementById(anchor);
      if (el) el.scrollIntoView({ block: 'start' });
      return;
    }
    window.scrollTo(0, 0);
  });
  setupReveal();
}

function renderNotFound(message) {
  state.current = { type: '404' };
  setHomeActive(false);
  $('#toc').innerHTML = '';
  $('#content').innerHTML = `<div class="page"><div class="empty-state">${message || '页面不存在'}</div></div>`;
}

function renderToc(headings) {
  const el = $('#toc');
  if (!headings.length) {
    el.innerHTML = '';
    return;
  }
  el.innerHTML = `
    <div class="toc-inner">
      <div class="toc-title">本页目录</div>
      <nav>
        ${headings.map((h) => `<a class="toc-link lv${Math.min(h.level, 3)}" href="#${h.id}">${esc(h.text)}</a>`).join('')}
      </nav>
    </div>`;
}

function initScrollSpy() {
  const links = $$('.toc-link');
  if (!links.length || !('IntersectionObserver' in window)) return;
  const targets = links
    .map((l) => document.getElementById(l.getAttribute('href').slice(1)))
    .filter(Boolean);
  const observer = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        links.forEach((l) => l.classList.toggle('active', l.getAttribute('href') === `#${entry.target.id}`));
      }
    });
  }, { rootMargin: '-80px 0px -70% 0px', threshold: 0 });
  targets.forEach((el) => observer.observe(el));
}

function bindCodeFiles() {
  $$('.code-file').forEach((file) => {
    if (file.dataset.bound === '1') return;
    file.dataset.bound = '1';
    const toggle = file.querySelector('.code-file-toggle');
    const copy = file.querySelector('.code-file-copy');
    const onToggle = () => {
      const isOpen = !file.classList.contains('collapsed');
      if (isOpen) {
        file.classList.add('collapsed');
        file.dataset.open = 'false';
        if (toggle) toggle.setAttribute('aria-expanded', 'false');
      } else {
        file.classList.remove('collapsed');
        file.dataset.open = 'true';
        if (toggle) toggle.setAttribute('aria-expanded', 'true');
      }
    };
    if (toggle) toggle.addEventListener('click', (ev) => { ev.stopPropagation(); onToggle(); });
    // 整条 head 也能点击切换，但点 buttons/copy 时不重复触发
    const head = file.querySelector('.code-file-head');
    if (head) {
      head.addEventListener('click', (ev) => {
        if (ev.target.closest('button')) return;
        onToggle();
      });
    }
    if (copy) {
      const code = file.querySelector('code');
      copy.addEventListener('click', (ev) => {
        ev.stopPropagation();
        if (!code) return;
        copyTextToClipboard(code.innerText, copy, ' 复制');
      });
    }
  });
}

function attachCopyButtons() {
  $$('.md-body pre').forEach((pre) => {
    if (pre.querySelector('.copy-btn')) return;
    const code = pre.querySelector('code');
    if (code) {
      const m = (code.className || '').match(/language-([\w+-]+)/i);
      if (m) pre.setAttribute('data-lang', m[1]);
      else pre.setAttribute('data-lang', 'code');
    }
    const btn = document.createElement('button');
    btn.className = 'copy-btn';
    btn.type = 'button';
    btn.title = '复制代码';
    btn.setAttribute('aria-label', '复制代码');
    btn.innerHTML = icon('copy', 14);
    btn.addEventListener('click', () => copyTextToClipboard(code ? code.innerText : '', btn));
    pre.appendChild(btn);
  });
}

function renderMermaidBlocks() {
  const blocks = $$('.mermaid');
  if (!blocks.length) return;
  const site = state.config.site || {};
  if (site.mermaid === false) return;

  const init = (mermaid) => {
    if (!mermaid) return;
    try {
      mermaid.initialize({
        startOnLoad: false,
        theme: document.documentElement.dataset.theme === 'dark' ? 'dark' : 'base',
        securityLevel: 'loose'
      });
      blocks.forEach((el) => {
        try {
          mermaid.run({ nodes: [el] });
        } catch (e) { /* keep source visible */ }
      });
    } catch (e) { /* keep source visible */ }
  };

  if (window.mermaid) {
    init(window.mermaid);
    return;
  }
  const src = site.mermaidCdn || 'https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js';
  const script = document.createElement('script');
  script.src = src;
  script.onload = () => init(window.mermaid);
  script.onerror = () => init(null);
  document.head.appendChild(script);
}

function codeFileHtml(innerHtml, lang, displayName) {
  const langName = String(lang || 'text').toLowerCase().trim() || 'text';
  return `
<div class="code-file" data-lang="${esc(langName)}" data-open="true">
  <div class="code-file-head">
    <span class="code-file-dots" aria-hidden="true"><i></i><i></i><i></i></span>
    <span class="code-file-name" title="${esc(displayName || langName)}">${esc(displayName || langName)}</span>
    <div class="code-file-actions">
      <button type="button" class="code-file-copy" title="复制代码" aria-label="复制代码">${icon('copy', 13)}<span class="code-file-copy-text">复制</span></button>
      <button type="button" class="code-file-toggle" aria-expanded="true" title="收起或展开代码" aria-label="收起或展开代码">${icon('chevron-up', 14)}</button>
    </div>
  </div>
  <div class="code-file-body"><pre><code class="language-${esc(langName)}">${innerHtml}</code></pre></div>
</div>`;
}

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
      return `\u0000${tokens.length - 1}\u0000`;
    });
    t = t.replace(/!\[([^\]]*)\]\(([^)\s]+)(?:\s+"[^"]*")?\)/g, (m, alt, src) => {
      tokens.push({ type: 'img', alt, src });
      return `\u0000${tokens.length - 1}\u0000`;
    });
    t = t.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (m, label, href) => {
      tokens.push({ type: 'link', label, href });
      return `\u0000${tokens.length - 1}\u0000`;
    });
    t = t.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    t = t.replace(/(^|[^*])\*([^*\n]+)\*(?!\*)/g, '$1<em>$2</em>');
    return t.replace(/\u0000(\d+)\u0000/g, (m, n) => {
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

function plainText(md) {
  return String(md || '')
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/!\[[^\]]*\]\([^)]*\)/g, ' ')
    .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1')
    .replace(/[#>*`|~_\-]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
}

async function loadIndex() {
  await Promise.all(state.flatArticles.map(async (a) => {
    let text = a.content || '';
    if (a.file) {
      try {
        const res = await fetch(a.file);
        if (res.ok) text = await res.text();
      } catch (e) { /* keep inline content */ }
    }
    indexMap.set(a.id, text);
    state.searchIndex.push({ ...a, plain: plainText(text) });
  }));
}

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

function closeSidebar() {
  document.body.classList.remove('sidebar-open');
}

function decodeFragmentPart(s) {
  try {
    return decodeURIComponent(s);
  } catch (e) {
    return s;
  }
}

function updateProgress() {
  const max = document.documentElement.scrollHeight - window.innerHeight;
  const p = max > 0 ? window.scrollY / max : 0;
  $('#progress').style.width = `${(p * 100).toFixed(2)}%`;
}

function route() {
  if (!state.config) return;
  const hash = location.hash.replace(/^#\/?/, '');
  const [pathPart, queryPart] = hash.split('?');
  const parts = pathPart.split('/').filter(Boolean).map(decodeFragmentPart);
  const params = new URLSearchParams(queryPart || '');
  const anchor = params.get('anchor') || '';

  if (parts[0] === 'a' && parts.length >= 3) {
    renderArticle(parts[1], parts[2], anchor);
  } else if (parts[0] === 'c' && parts[1]) {
    renderCategory(parts[1]);
  } else {
    renderHome();
  }
}

function initTheme() {
  const btn = $('#theme-btn');
  const update = () => {
    const dark = document.documentElement.dataset.theme === 'dark';
    btn.innerHTML = icon(dark ? 'sun' : 'moon', 18);
    btn.title = dark ? '切换到亮色模式' : '切换到暗色模式';
  };
  update();
  btn.addEventListener('click', () => {
    const next = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark';
    document.documentElement.dataset.theme = next;
    try {
      localStorage.setItem('kb-theme', next);
    } catch (e) { /* ignore */ }
    update();
  });
}

function initBackToTop() {
  const btn = $('#back-to-top');
  if (!btn) return;
  btn.addEventListener('click', () => window.scrollTo({ top: 0, behavior: 'smooth' }));
  toggleBackToTop();
}

function toggleBackToTop() {
  const btn = $('#back-to-top');
  if (!btn) return;
  btn.classList.toggle('show', window.scrollY > 480);
}

function setupReveal() {
  if (!('IntersectionObserver' in window)) return;
  const root = $('#content');
  if (!root) return;
  const targets = $$('.cat-card, .recent-row, .stat-chip, .md-body h2, .md-body pre, .md-body blockquote, .source-section, .article-nav-link, .empty-state');
  targets.forEach((el, i) => {
    if (!el.classList.contains('reveal')) {
      el.classList.add('reveal');
      el.style.transitionDelay = `${Math.min(i, 8) * 40}ms`;
    }
  });
  const io = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        entry.target.classList.add('in');
        io.unobserve(entry.target);
      }
    });
  }, { rootMargin: '0px 0px -8% 0px', threshold: 0.05 });
  targets.forEach((el) => io.observe(el));
}

async function init() {
  initTheme();

  $('#menu-btn').addEventListener('click', () => document.body.classList.add('sidebar-open'));
  $('#overlay').addEventListener('click', closeSidebar);

  function toggleFolderByEl(folder) {
    if (!folder) return;
    const key = folder.dataset.nodeKey;
    folder.classList.toggle('collapsed');
    const nowCollapsed = folder.classList.contains('collapsed');
    const btn = folder.querySelector('.nav-folder-toggle');
    if (btn) btn.setAttribute('aria-expanded', nowCollapsed ? 'false' : 'true');
    state.userOverrides.set(key, nowCollapsed);
    saveUserOverrides();
  }

  $('#sidebar').addEventListener('click', (e) => {
    const link = e.target.closest('a.nav-folder-name');
    if (link) closeSidebar();
    const go = e.target.closest('[data-go]');
    if (go) {
      location.hash = `#/c/${go.dataset.go}`;
      closeSidebar();
      return;
    }
    const toggle = e.target.closest('[data-toggle]');
    if (toggle) {
      const cat = toggle.closest('.nav-cat');
      if (cat) cat.classList.toggle('open');
    }
    // 嵌套目录折叠 — chevron 按钮（任意层级）
    const folderToggle = e.target.closest('.nav-folder-toggle');
    if (folderToggle) {
      e.preventDefault();
      e.stopPropagation();
      toggleFolderByEl(folderToggle.closest('.nav-folder'));
      return;
    }
    // 二级及更深：标题点击也会折叠/展开
    const titleToggle = e.target.closest('[data-folder-toggle]');
    if (titleToggle) {
      e.preventDefault();
      e.stopPropagation();
      toggleFolderByEl(titleToggle.closest('.nav-folder'));
    }
  });

  // 二级及以上标题，键盘 Enter / Space 也能切换
  $('#sidebar').addEventListener('keydown', (e) => {
    if (e.key !== 'Enter' && e.key !== ' ') return;
    const titleToggle = e.target.closest('[data-folder-toggle]');
    if (!titleToggle) return;
    e.preventDefault();
    toggleFolderByEl(titleToggle.closest('.nav-folder'));
  });

  $('#toc').addEventListener('click', (e) => {
    const link = e.target.closest('.toc-link');
    if (link) {
      $$('.toc-link').forEach((l) => l.classList.toggle('active', l === link));
    }
  });

  const searchInput = $('#search-input');
  searchInput.addEventListener('input', (e) => performSearch(e.target.value));
  searchInput.addEventListener('focus', (e) => {
    if (e.target.value.trim()) performSearch(e.target.value);
  });
  searchInput.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      closeSearch();
      e.target.blur();
    }
    if (e.key === 'Enter') {
      const first = $('#search-results a.search-hit');
      if (first) location.hash = first.getAttribute('href');
    }
  });
  document.addEventListener('click', (e) => {
    if (!e.target.closest('.search')) closeSearch();
  });
  document.addEventListener('keydown', (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
      e.preventDefault();
      searchInput.focus();
    } else if (e.key === '/' && !/input|textarea|select/i.test(document.activeElement.tagName)) {
      e.preventDefault();
      searchInput.focus();
    }
  });

  window.addEventListener('hashchange', () => {
    if (/^#h-\d+/.test(location.hash)) {
      const el = document.getElementById(location.hash.slice(1));
      if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
      return;
    }
    route();
  });

  window.addEventListener('scroll', () => {
    updateProgress();
    toggleBackToTop();
  }, { passive: true });
  window.addEventListener('resize', updateProgress);

  initBackToTop();
  setupReveal();

  try {
    const res = await fetch('config.json?t=' + Date.now());
    if (!res.ok) throw new Error('config not found');
    state.config = await res.json();
    applySiteConfig();
    buildFlatArticles();
    renderSidebar();
    await loadIndex();
    route();
  } catch (err) {
    $('#content').innerHTML = `<div class="page"><div class="empty-state">无法加载 config.json，请通过本地静态服务器打开本站。</div></div>`;
  }
}

document.addEventListener('DOMContentLoaded', init);
