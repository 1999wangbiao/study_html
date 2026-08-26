const http = require('http');
const fs = require('fs');
const path = require('path');
const { MIME, buildConfig } = require('./lib/scanner');

const ROOT = __dirname;
const PORT = Number(process.env.PORT) || 4173;
// 资源版本号：注入 index.html 的 ?v=xxx，防止浏览器缓存旧 JS/CSS。
// 直接 node server.js 时用启动时间戳；跑过 build.js 后 index.html 里是构建版本，
// 这里兜底替换，保证两种启动方式都带上版本号。
const SERVER_VERSION = new Date().toISOString().slice(0, 16).replace(/[-:T]/g, '');

function injectVersion(html) {
  return html.replace(/__BUILD_VERSION__/g, SERVER_VERSION);
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
      let body = data;
      if (ext === '.html') body = Buffer.from(injectVersion(data.toString('utf8')), 'utf8');
      res.writeHead(200, {
        'Content-Type': MIME[ext] || 'application/octet-stream',
        'Cache-Control': 'no-store'
      });
      res.end(body);
    });
  });
});

server.listen(PORT, () => {
  console.log(`Knowledge base running at http://localhost:${PORT}`);
});
