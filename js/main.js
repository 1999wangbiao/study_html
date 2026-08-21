'use strict';

// 应用入口：主题、路由、启动引导与全局事件绑定。
// main.js 不导出任何东西——它是 <script type="module"> 的加载入口；
// 其余模块（pages/nav/search/state/utils 等）都从这里或从页面渲染链路被 import。

import { $, $$, icon } from './utils.js';
import { state, saveUserOverrides, loadIndex } from './state.js';
import { renderSidebar } from './nav.js';
import { performSearch, closeSearch } from './search.js';
import { renderHome, renderCategory, renderArticle, closeSidebar, updateProgress, setupReveal, applySiteConfig, buildFlatArticles } from './pages.js';

function decodeFragmentPart(s) {
  try {
    return decodeURIComponent(s);
  } catch (e) {
    return s;
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

  // 切页后：当前分类保持张开，方便看到当前位置。
  // 若用户用右侧按钮收起了它，则尊重偏好、不强制展开。
  const curCat = document.querySelector('.nav-cat[data-cat="' + (state.current.categoryId || '') + '"]');
  if (curCat && !curCat.classList.contains('open')) {
    const catKey = 'cat::' + state.current.categoryId;
    const userCollapsed = state.userOverrides.get(catKey) === true;
    if (!userCollapsed) {
      curCat.classList.add('open');
      state.userOverrides.set(catKey, false);
      saveUserOverrides();
    }
  }
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

  // 一级分类的展开/收起 —— 与文件夹对称，写入持久化偏好（key 为 "cat::<catId>"）
  function toggleCategoryByEl(cat) {
    if (!cat) return;
    cat.classList.toggle('open');
    const nowOpen = cat.classList.contains('open');
    state.userOverrides.set('cat::' + cat.dataset.cat, !nowOpen);
    saveUserOverrides();
  }

  $('#sidebar').addEventListener('click', (e) => {
    // 一级分类：点名称/图标/数字 → 进入该分类主页（不做收缩）；点右侧 chevron → 收缩/展开（持久化）
    const go = e.target.closest('[data-go]');
    if (go) {
      location.hash = `#/c/${go.dataset.go}`;
      closeSidebar();
      return;
    }
    const toggle = e.target.closest('[data-toggle]');
    if (toggle) {
      toggleCategoryByEl(toggle.closest('.nav-cat'));
      return;
    }
    // 嵌套目录折叠 — chevron 按钮（任意层级）
    const folderToggle = e.target.closest('.nav-folder-toggle');
    if (folderToggle) {
      e.preventDefault();
      e.stopPropagation();
      toggleFolderByEl(folderToggle.closest('.nav-folder'));
      return;
    }
    // 所有层级目录标题点击均可折叠/展开
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

init();
