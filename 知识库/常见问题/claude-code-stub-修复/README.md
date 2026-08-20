# Claude Code `claude.exe` Stub 修复完整整理

> 报错：`程序"claude.exe"无法运行: 指定的可执行文件不是此操作系统平台的有效应用程序`

## 0. 总览

| # | 内容 | 一句话 |
|---|------|--------|
| 1 | 报错现象 | claude.exe 只有 500 字节，是占位 stub |
| 2 | 根因 | 子包没装上 + Anthropic 版本不同步 + auto-updater 自动升级 |
| 3 | 关键路径 | npm 全局目录 + TRAE 内置目录 |
| 4 | 修复方案 A | 本地已有真二进制 → 直接复制覆盖 |
| 5 | 修复方案 B | 本地无真二进制 → 官方源下载匹配版本 |
| 6 | 验证 | `claude --version` 返回 2.1.236 |
| 7 | 后续避坑 | 禁用自动更新 + 镜像源注意事项 |

---

## 1. 报错现象

```text
程序"claude.exe"无法运行: 指定的可执行文件不是此操作系统平台的有效应用程序
所在位置 C:\Users\Administrator\AppData\Roaming\npm\claude.ps1:14 字符: 3
+   & "$basedir/node_modules/@anthropic-ai/claude-code/bin/claude.exe"  ...
```

关键判别：执行目录下的 `claude.exe` **只有 500 字节**（占位 stub），不是真正的可执行文件。

---

## 2. 根因分析

三个问题叠加：

### 2.1 npm 全局模式 + 镜像源对 optionalDependencies 不可靠

`@anthropic-ai/claude-code` 主包通过 `optionalDependencies` 声明平台特定子包：

```json
"optionalDependencies": {
  "@anthropic-ai/claude-code-win32-x64": "2.1.237",
  "@anthropic-ai/claude-code-win32-arm64": "2.1.237",
  ...
}
```

在 npm 全局模式 + 淘宝/腾讯镜像源下，optional 子包常常装不上 → postinstall 脚本找不到真二进制 → 500 字节 stub 没被覆盖。

### 2.2 Anthropic 发布版本不同步

- 主包 `claude-code` 最新版：**2.1.237**
- 主包 2.1.237 要求子包 `claude-code-win32-x64@2.1.237`
- **子包最高版本只有 2.1.236**，2.1.237 根本没发布

npm 找不到精确匹配的子包版本，optional 失败时静默跳过。

### 2.3 claude 自带 auto-updater 自动升级

claude 运行时内置 auto-updater + 后台 daemon：

1. 启动时后台检查新版本
2. 发现 2.1.237 后**在后台悄悄执行 `npm install -g @anthropic-ai/claude-code`**
3. 拉到坏版本 2.1.237 → 写回 stub → 把自己搞坏

证据（`~/.claude/.last-update-result.json`）：

```json
{
  "version_from": "2.1.236",
  "version_to": "2.1.237",
  "outcome": "success",
  "path": "npm-global",
  "timestamp": "2026-08-20T02:22:21Z"
}
```

**这就是"明明啥都没干，它又坏了"的原因。**

---

## 3. 关键路径

系统里有两个 claude 安装，PATH 优先级决定实际执行哪个：

| 位置 | 路径 | 用途 |
|------|------|------|
| npm 全局 | `C:\Users\Administrator\AppData\Roaming\npm\node_modules\@anthropic-ai\claude-code\bin\claude.exe` | 默认 `claude` 命令走这里 |
| TRAE 内置 | `C:\Users\Administrator\AppData\Roaming\TRAE SOLO CN\ModularData\ai-agent\vm\tools\node\node_modules\@anthropic-ai\claude-code\bin\claude.exe` | TRAE 沙盒环境用 |

判断哪个在用：

```powershell
Get-Command claude | Select-Object Source
```

---

## 4. 修复方案 A：本地已存在真 exe（快速）

**适用**：TRAE 内置目录的 `claude.exe` 大于 100MB（说明已有真二进制）。

### 4.1 判断走哪个方案

```powershell
Get-Item "C:\Users\Administrator\AppData\Roaming\TRAE SOLO CN\ModularData\ai-agent\vm\tools\node\node_modules\@anthropic-ai\claude-code\bin\claude.exe" | Select-Object Length
```

- Length > 100MB → 走方案 A
- Length = 500 → 走方案 B

### 4.2 步骤

```powershell
# 0) 先禁用自动更新（编辑 ~/.claude/settings.json，加字段）
#    "autoUpdaterStatus": "disabled"

# 1) TRAE 真二进制覆盖 npm 全局 stub
Copy-Item "C:\Users\Administrator\AppData\Roaming\TRAE SOLO CN\ModularData\ai-agent\vm\tools\node\node_modules\@anthropic-ai\claude-code\bin\claude.exe" "C:\Users\Administrator\AppData\Roaming\npm\node_modules\@anthropic-ai\claude-code\bin\claude.exe" -Force

# 2) 验证
claude --version
# 应返回: 2.1.236 (Claude Code)
```

---

## 5. 修复方案 B：本地无真 exe（需下载）

**适用**：两个目录都是 500 字节 stub，或 TRAE 内置目录也没有真二进制。

### 5.1 步骤

```powershell
# 0) 先禁用自动更新（编辑 ~/.claude/settings.json，加字段）
#    "autoUpdaterStatus": "disabled"

# 1) 用官方源下载匹配版本子包（必须 2.1.236，2.1.237 子包不存在）
cd d:\workspace\class\class_study\study_html
npm install @anthropic-ai/claude-code-win32-x64@2.1.236 --registry=https://registry.npmjs.org --no-save

# 2) 复制真二进制覆盖 stub
Copy-Item ".\node_modules\@anthropic-ai\claude-code-win32-x64\claude.exe" "C:\Users\Administrator\AppData\Roaming\npm\node_modules\@anthropic-ai\claude-code\bin\claude.exe" -Force

# 3) 清理临时下载 + 验证
Remove-Item -Recurse -Force ".\node_modules\@anthropic-ai\claude-code-win32-x64"
claude --version
# 应返回: 2.1.236 (Claude Code)
```

### 5.2 关键点

- **必须用 2.1.236**：2.1.237 子包在官方源上不存在
- **必须用官方源**：`--registry=https://registry.npmjs.org`，淘宝/腾讯镜像对平台子包支持不完整
- **下载约 314MB**：耐心等几分钟

---

## 6. 验证清单

```powershell
# 1) claude 命令能跑
claude --version
# 期望: 2.1.236 (Claude Code)

# 2) 两个目录的 claude.exe 大小
Get-Item "C:\Users\Administrator\AppData\Roaming\npm\node_modules\@anthropic-ai\claude-code\bin\claude.exe" | Select-Object Length
Get-Item "C:\Users\Administrator\AppData\Roaming\TRAE SOLO CN\ModularData\ai-agent\vm\tools\node\node_modules\@anthropic-ai\claude-code\bin\claude.exe" | Select-Object Length
# 期望: 都约 330,097,824 字节（314.8 MB）

# 3) 自动更新已禁用
(Get-Content "$env:USERPROFILE\.claude\settings.json" -Raw | ConvertFrom-Json).autoUpdaterStatus
# 期望: disabled
```

---

## 7. 后续避坑

### 7.1 必须先禁用自动更新

**两个方案都必须先做第 0 步**。否则 claude 一运行，auto-updater 又会后台拉 2.1.237 把 stub 写回去，进入"修了又坏"死循环。

禁用方式（`~/.claude/settings.json` 加字段）：

```json
{
  "env": { ... },
  "permissions": { ... },
  "autoUpdaterStatus": "disabled"
}
```

PowerShell 一键添加：

```powershell
$cfg = "$env:USERPROFILE\.claude\settings.json"
$json = Get-Content $cfg -Raw | ConvertFrom-Json
$json | Add-Member -NotePropertyName "autoUpdaterStatus" -NotePropertyValue "disabled" -Force
$json | ConvertTo-Json -Depth 10 | Set-Content $cfg -Encoding UTF8
```

### 7.2 镜像源注意事项

日常用淘宝/腾讯镜像没问题，但**装带平台子包的工具时**（claude-code、esbuild、swc、@swc/core 等），临时加 `--registry=https://registry.npmjs.org` 用官方源更可靠：

```powershell
npm install -g @anthropic-ai/claude-code@2.1.236 --registry=https://registry.npmjs.org --include=optional
```

### 7.3 何时再升级

- 等 Anthropic 修复 2.1.237 子包发布问题（让主包和子包版本匹配）
- 升级前先把 `autoUpdaterStatus` 改回 `enabled` 或删掉
- 升级后若 `claude --version` 又报错，重新走方案 A 覆盖真二进制

### 7.4 `claude.exe.old` 文件

`claude.exe.old.时间戳` 是 npm 在 Windows 上更新被占用 exe 时的"改名绕过"产物：

1. 旧 `claude.exe` 正在运行被加锁
2. npm 重命名为 `claude.exe.old.时间戳` 腾位置
3. 写入新文件到 `claude.exe`
4. `.old` 残留（等下次清理或重启）

**不影响使用**。看着碍眼可手动删除：

```powershell
Get-ChildItem "C:\Users\Administrator\AppData\Roaming\npm\node_modules\@anthropic-ai" -Recurse -Filter "*.old.*" | Remove-Item -Force
```

---

## 8. 决策流程

```text
claude --version 报错？
  │
  ├─ 查 TRAE 内置目录 claude.exe 大小
  │    │
  │    ├─ > 100MB → 方案 A（复制覆盖，几秒搞定）
  │    │
  │    └─ = 500 字节 → 方案 B（官方源下载 314MB，几分钟）
  │
  ├─ 两个方案都先做第 0 步：禁用自动更新
  │
  ├─ 覆盖真二进制到 npm 全局目录
  │
  └─ claude --version 验证返回 2.1.236
```

---

## 附录 A：完整一键修复脚本（方案 A）

```powershell
# === Claude Code stub 修复脚本（方案 A：本地已有真二进制）===

# 0) 禁用自动更新
$cfg = "$env:USERPROFILE\.claude\settings.json"
$json = Get-Content $cfg -Raw | ConvertFrom-Json
$json | Add-Member -NotePropertyName "autoUpdaterStatus" -NotePropertyValue "disabled" -Force
$json | ConvertTo-Json -Depth 10 | Set-Content $cfg -Encoding UTF8
Write-Host "[OK] 自动更新已禁用"

# 1) 复制真二进制覆盖 stub
$src = "C:\Users\Administrator\AppData\Roaming\TRAE SOLO CN\ModularData\ai-agent\vm\tools\node\node_modules\@anthropic-ai\claude-code\bin\claude.exe"
$dst = "C:\Users\Administrator\AppData\Roaming\npm\node_modules\@anthropic-ai\claude-code\bin\claude.exe"
Copy-Item -Path $src -Destination $dst -Force
Write-Host "[OK] 真二进制已覆盖"

# 2) 验证
Write-Host "`n=== 验证 ==="
claude --version
```

## 附录 B：完整一键修复脚本（方案 B）

```powershell
# === Claude Code stub 修复脚本（方案 B：需下载真二进制）===

# 0) 禁用自动更新
$cfg = "$env:USERPROFILE\.claude\settings.json"
$json = Get-Content $cfg -Raw | ConvertFrom-Json
$json | Add-Member -NotePropertyName "autoUpdaterStatus" -NotePropertyValue "disabled" -Force
$json | ConvertTo-Json -Depth 10 | Set-Content $cfg -Encoding UTF8
Write-Host "[OK] 自动更新已禁用"

# 1) 用官方源下载匹配版本子包（必须 2.1.236）
$tmpDir = "$env:TEMP\claude-code-fix"
New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null
Push-Location $tmpDir
npm install @anthropic-ai/claude-code-win32-x64@2.1.236 --registry=https://registry.npmjs.org --no-save
Write-Host "[OK] 子包下载完成"

# 2) 复制真二进制覆盖 stub
$src = "$tmpDir\node_modules\@anthropic-ai\claude-code-win32-x64\claude.exe"
$dst = "C:\Users\Administrator\AppData\Roaming\npm\node_modules\@anthropic-ai\claude-code\bin\claude.exe"
Copy-Item -Path $src -Destination $dst -Force
Write-Host "[OK] 真二进制已覆盖"

# 3) 清理 + 验证
Pop-Location
Remove-Item -Recurse -Force $tmpDir
Write-Host "`n=== 验证 ==="
claude --version
```

---

## 记忆口诀

1. **报错先看 exe 大小**：500 字节就是 stub
2. **方案 A 走 TRAE，方案 B 走官方源**：子包必须 2.1.236
3. **修之前先禁用 auto-updater**：否则修了又被后台升回坏版本
4. **镜像源有坑**：带平台子包的工具用 `--registry=https://registry.npmjs.org`
5. **`.old` 不影响使用**：npm 在 Windows 上更新被占用 exe 的改名绕过产物
