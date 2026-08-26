const fs = require('fs');
const path = require('path');
const { buildConfig } = require('./lib/scanner');

const ROOT = __dirname;

const config = buildConfig();
const outPath = path.join(ROOT, 'config.json');
fs.writeFileSync(outPath, JSON.stringify(config, null, 2), 'utf8');

// 构建版本号：注入 index.html 的资源 URL（?v=xxx），发版时递增以强制浏览器拉新资源。
// 用"构建时刻"作为版本号，任何一次 build 都会产生新版本，天然防缓存。
const version = new Date().toISOString().slice(0, 16).replace(/[-:T]/g, ''); // 形如 202608261530
const indexPath = path.join(ROOT, 'index.html');
const html = fs.readFileSync(indexPath, 'utf8');
const injected = html.replace(/__BUILD_VERSION__/g, version);
fs.writeFileSync(indexPath, injected, 'utf8');

const total = config.categories.reduce((n, c) => n + c.articles.length, 0);
console.log(`config.json 已生成: ${config.categories.length} 个分类, ${total} 篇文章`);
console.log(`index.html 资源版本: v=${version}`);
config.categories.forEach((c) => {
  console.log(`  - ${c.name}: ${c.articles.length} 篇`);
});
