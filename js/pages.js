'use strict';

// 页面渲染：首页 / 分类页 / 文章页 / 404。
// 同时承载渲染时需要的"主流程辅助函数"（applySiteConfig / buildFlatArticles /
// setupReveal / renderMermaidBlocks / updateProgress / closeSidebar）——
// 这些只在页面渲染和 main 的启动流程中使用，放在这里可避免 main ↔ pages 循环导入。

import { $, $$, esc, icon, articleUrl } from './utils.js';
import { state, indexMap } from './state.js';
import { renderSidebar, catCardHtml, groupArticles } from './nav.js';
import { renderMarkdown } from './markdown.js';
import { getArticleMarkdown, getArticleSourceFiles, sourceSectionHtml, bindSourceControls } from './source.js';
import { closeSearch } from './search.js';
import { renderToc, initScrollSpy } from './toc.js';
import { attachCopyButtons, bindCodeFiles } from './code-block.js';

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

function closeSidebar() {
  document.body.classList.remove('sidebar-open');
}

function updateProgress() {
  const max = document.documentElement.scrollHeight - window.innerHeight;
  const p = max > 0 ? window.scrollY / max : 0;
  $('#progress').style.width = `${(p * 100).toFixed(2)}%`;
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

export {
  applySiteConfig, buildFlatArticles, setHomeActive, renderHome, renderCategory,
  renderArticle, renderNotFound, articleRowHtml, navLinkHtml, stripLeadingH1,
  renderMermaidBlocks, closeSidebar, updateProgress, setupReveal
};
