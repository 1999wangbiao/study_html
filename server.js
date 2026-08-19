const http = require('http');
const fs = require('fs');
const path = require('path');

const ROOT = __dirname;
const KB_ROOT = path.join(ROOT, '知识库');
const PORT = Number(process.env.PORT) || 4173;

const MIME = {
  '.html': 'text/html; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.md': 'text/markdown; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.webp': 'image/webp',
  '.ico': 'image/x-icon',
  '.woff2': 'font/woff2'
};

const DEFAULT_SITE = {
  title: '计算机知识库',
  subtitle: '',
  description: '',
  footer: '',
  accent: '#0f766e',
  mermaid: true
};

function loadSite() {
  try {
    const cfg = JSON.parse(fs.readFileSync(path.join(ROOT, 'config.json'), 'utf8'));
    return Object.assign({}, DEFAULT_SITE, cfg.site || {});
  } catch (e) {
    return Object.assign({}, DEFAULT_SITE);
  }
}

const site = loadSite();

function slugId(s) {
  return String(s).replace(/[\\/:*?"<>|]/g, '-').replace(/\s+/g, '-').trim();
}

function categoryIcon(name) {
  const map = {
    '设计模式': 'layers',
    '常见问题': 'tag',
    'Go语言': 'code',
    '前端': 'code',
    '后端': 'server',
    '计算机基础': 'cpu',
    '面试知识': 'tag'
  };
  return map[name] || 'file-text';
}

function listEntries(dir) {
  const dirs = [];
  const files = [];
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    if (e.name.startsWith('.')) continue;
    const full = path.join(dir, e.name);
    if (e.isDirectory()) dirs.push({ name: e.name, full });
    else files.push({ name: e.name, full });
  }
  dirs.sort((a, b) => a.name.localeCompare(b.name, 'zh'));
  files.sort((a, b) => a.name.localeCompare(b.name, 'zh'));
  return { dirs, files };
}

function findMarkdown(root) {
  const out = [];
  const walk = (dir) => {
    const { dirs, files } = listEntries(dir);
    for (const d of dirs) walk(d.full);
    for (const f of files) {
      if (/\.md$/i.test(f.name)) out.push(f);
    }
  };
  walk(root);
  return out;
}

function headingAndSummary(text) {
  const clean = String(text || '').replace(/^\uFEFF/, '').trim();
  const lines = clean.split(/\r?\n/);
  let title = '';
  for (const line of lines) {
    if (/^#\s+/.test(line)) {
      title = line.replace(/^#+\s*/, '').trim();
      break;
    }
  }
  return { title, summary: '' };
}

function sourceDepth(p) {
  return p.split('/').length;
}

function collectSources(articleDir) {
  const go = [];
  const walk = (dir, prefix) => {
    const { dirs, files } = listEntries(dir);
    for (const f of files) {
      if (/\.go$/i.test(f.name)) go.push((prefix ? prefix + '/' : '') + f.name);
    }
    for (const d of dirs) walk(d.full, (prefix ? prefix + '/' : '') + d.name);
  };
  walk(articleDir, '');
  go.sort((a, b) => {
    const da = sourceDepth(a);
    const db = sourceDepth(b);
    if (da !== db) return da - db;
    if (a === 'main.go') return -1;
    if (b === 'main.go') return 1;
    return a.localeCompare(b, 'zh');
  });
  return go;
}

function buildArticle(catName, md, catDir, usedIds) {
  const dirParts = path.relative(catDir, path.dirname(md.full)).split(path.sep).filter(Boolean);
  const group = dirParts[0] || catName;
  const file = path.relative(ROOT, md.full).split(path.sep).join('/');
  const stat = fs.statSync(md.full);
  const updated = stat.mtime.toISOString().slice(0, 10);

  const folderName = path.basename(path.dirname(md.full));
  const baseName = path.basename(md.full, path.extname(md.full));
  const rawId = baseName.toLowerCase() === 'readme' ? folderName : baseName;
  const orig = slugId(rawId);
  let id = orig;
  let n = 2;
  while (usedIds.has(id)) {
    id = orig + '-' + n++;
  }
  usedIds.add(id);

  let title = baseName.toLowerCase() === 'readme' ? folderName : baseName;
  let summary = '';
  try {
    const hs = headingAndSummary(fs.readFileSync(md.full, 'utf8'));
    if (hs.title) title = hs.title;
    summary = hs.summary;
  } catch (e) {
    /* keep filename fallback */
  }

  const article = {
    id,
    title,
    file,
    group,
    path: dirParts,
    summary,
    updated
  };
  const code = collectSources(path.dirname(md.full));
  if (code.length) article.code = code;
  return article;
}

function finalizeSite(categories) {
  const names = categories.map((c) => c.name).join(' · ');
  const total = categories.reduce((n, c) => n + c.articles.length, 0);
  return Object.assign({}, site, {
    subtitle: (site.subtitle || names).trim(),
    description: (site.description || `自动扫描知识库目录，当前 ${categories.length} 个分类 · ${total} 篇文章`).trim(),
    footer: (site.footer || '计算机知识库 · 自动识别“知识库”目录').trim()
  });
}

function buildConfig() {
  const categories = [];
  if (fs.existsSync(KB_ROOT)) {
    const topDirs = fs.readdirSync(KB_ROOT, { withFileTypes: true })
      .filter((d) => d.isDirectory() && !d.name.startsWith('.'))
      .sort((a, b) => a.name.localeCompare(b.name, 'zh'));

    topDirs.forEach((d) => {
      const catDir = path.join(KB_ROOT, d.name);
      const mds = findMarkdown(catDir);
      if (!mds.length) return;
      const usedIds = new Set();
      const articles = mds.map((md) => buildArticle(d.name, md, catDir, usedIds));
      const secs = new Set();
      articles.forEach((a) => {
        if (a.group && a.group !== d.name) secs.add(a.group);
      });
      categories.push({
        id: slugId(d.name),
        name: d.name,
        icon: categoryIcon(d.name),
        description: Array.from(secs).join(' · ') || d.name,
        articles
      });
    });
  }
  return { site: finalizeSite(categories), categories };
}

const server = http.createServer((req, res) => {
  let pathname;
  try {
    pathname = decodeURIComponent(new URL(req.url, 'http://localhost').pathname);
  } catch (e) {
    res.writeHead(400);
    res.end('Bad Request');
    return;
  }

  if (pathname === '/config.json') {
    res.writeHead(200, { 'Content-Type': MIME['.json'], 'Cache-Control': 'no-store' });
    res.end(JSON.stringify(buildConfig()));
    return;
  }

  let filePath = path.normalize(path.join(ROOT, pathname === '/' ? 'index.html' : pathname));
  if (!filePath.startsWith(ROOT)) {
    res.writeHead(403, { 'Content-Type': 'text/plain; charset=utf-8' });
    res.end('Forbidden');
    return;
  }

  fs.stat(filePath, (statErr, stat) => {
    if (!statErr && stat.isDirectory()) {
      filePath = path.join(filePath, 'index.html');
    }
    fs.readFile(filePath, (readErr, data) => {
      if (readErr) {
        res.writeHead(404, { 'Content-Type': 'text/plain; charset=utf-8' });
        res.end('404 Not Found');
        return;
      }
      const ext = path.extname(filePath).toLowerCase();
      res.writeHead(200, {
        'Content-Type': MIME[ext] || 'application/octet-stream',
        'Cache-Control': 'no-store'
      });
      res.end(data);
    });
  });
});

server.listen(PORT, () => {
  console.log(`Knowledge base running at http://localhost:${PORT}`);
});
