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

// ========== 桌面端侧栏：折叠 / 拖拽调宽（宽度与状态记忆到 localStorage） ==========

const SIDEBAR_W_KEY = 'kb-sidebar-w';
const SIDEBAR_COLLAPSED_KEY = 'kb-sidebar-collapsed';
const SIDEBAR_MIN = 200;
const SIDEBAR_MAX = 440;

function getStoredNumber(key, fallback) {
  try {
    const n = Number(localStorage.getItem(key));
    return Number.isFinite(n) && n > 0 ? n : fallback;
  } catch (e) { return fallback; }
}
function storeNumber(key, v) {
  try { localStorage.setItem(key, String(v)); } catch (e) { /* ignore */ }
}
function isMobile() {
  return window.matchMedia('(max-width: 900px)').matches;
}

// 应用桌面端侧栏折叠态 + 用户记忆的宽度（移动端为抽屉，不参与）
function initSidebarState() {
  const toggleBtn = $('#sidebar-toggle-btn');
  let collapsed = false;
  try { collapsed = localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === '1'; } catch (e) { /* ignore */ }

  const apply = () => {
    const root = document.documentElement;
    root.style.setProperty('--sidebar-w', getStoredNumber(SIDEBAR_W_KEY, 264) + 'px');
    root.classList.toggle('sidebar-collapsed', collapsed);
    if (toggleBtn) {
      const show = !isMobile();
      toggleBtn.hidden = !show;
      toggleBtn.setAttribute('aria-expanded', collapsed ? 'false' : 'true');
      toggleBtn.title = collapsed ? '展开侧栏' : '折叠侧栏';
    }
  };
  apply();

  if (toggleBtn) {
    toggleBtn.addEventListener('click', () => {
      collapsed = !collapsed;
      try { localStorage.setItem(SIDEBAR_COLLAPSED_KEY, collapsed ? '1' : '0'); } catch (e) { /* ignore */ }
      apply();
    });
  }

  // 宽度变化（含首次加载）时同步折叠按钮显隐：折叠后应只剩展开入口
  const ro = new ResizeObserver(() => {
    if (toggleBtn) {
      const show = !isMobile();
      toggleBtn.hidden = show ? collapsed : true;
      toggleBtn.setAttribute('aria-expanded', collapsed ? 'false' : 'true');
    }
  });
  if (typeof ResizeObserver !== 'undefined') ro.observe(document.body);
}

// 拖拽侧栏右缘调宽：mousedown 记录起点 → mousemove 更新 --sidebar-w → mouseup 持久化
function initSidebarResizer() {
  const resizer = $('#sidebar-resizer');
  if (!resizer) return;

  resizer.addEventListener('mousedown', (e) => {
    if (e.button !== 0 || isMobile() || document.body.classList.contains('sidebar-collapsed')) return;
    e.preventDefault();
    const startX = e.clientX;
    const startW = getStoredNumber(SIDEBAR_W_KEY, 264);
    let delta = 0;

    const onMove = (ev) => {
      delta = ev.clientX - startX;
      const w = Math.min(SIDEBAR_MAX, Math.max(SIDEBAR_MIN, startW + delta));
      document.documentElement.style.setProperty('--sidebar-w', w + 'px');
    };
    const onUp = () => {
      storeNumber(SIDEBAR_W_KEY, Math.min(SIDEBAR_MAX, Math.max(SIDEBAR_MIN, startW + delta)));
      document.body.classList.remove('sidebar-resizing');
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
    };
    document.body.classList.add('sidebar-resizing');
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  });
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
  initSidebarState();
  initSidebarResizer();
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
