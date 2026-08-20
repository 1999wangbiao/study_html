const fs = require('fs');
const path = require('path');
const { buildConfig } = require('./lib/scanner');

const ROOT = __dirname;

const config = buildConfig();
const outPath = path.join(ROOT, 'config.json');
fs.writeFileSync(outPath, JSON.stringify(config, null, 2), 'utf8');

const total = config.categories.reduce((n, c) => n + c.articles.length, 0);
console.log(`config.json 已生成: ${config.categories.length} 个分类, ${total} 篇文章`);
config.categories.forEach((c) => {
  console.log(`  - ${c.name}: ${c.articles.length} 篇`);
});
