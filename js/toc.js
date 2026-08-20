'use strict';

// 文章页右侧目录 + 滚动高亮。

import { $, $$, esc } from './utils.js';

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

export { renderToc, initScrollSpy };
