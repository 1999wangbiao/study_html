'use strict';

// 全局状态与搜索索引 —— 独立成叶模块，避免 main ↔ nav/pages 的循环导入。
// state.userOverrides 在模块顶层求值时调用 loadUserOverrides()（读 localStorage），
// 因此必须在此处先定义好函数，绝不能反向 import 本模块的调用方。

import { plainText } from './utils.js';

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

// 预取所有文章正文，构建全文搜索索引。
// 只在启动时调用一次；此后文章正文优先从 indexMap 取缓存。
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

export { state, loadUserOverrides, saveUserOverrides, defaultCollapsed, isNodeCollapsed, indexMap, loadIndex };
