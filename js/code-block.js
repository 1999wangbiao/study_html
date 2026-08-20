'use strict';

// 代码文件展示 —— 同时被 markdown.js（渲染代码围栏）和 source.js（源码展示）使用，
// 因此独立成叶模块。

import { $, $$, icon, esc, copyTextToClipboard } from './utils.js';

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

export { codeFileHtml, bindCodeFiles, attachCopyButtons };
