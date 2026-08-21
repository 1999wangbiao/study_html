# Docker 开发环境 vs 生产环境完整对比

独立专题。围绕「不同 / 处理 / 为什么」三段式展开，覆盖 Dockerfile、依赖、源码、运行用户、资源限制、重启策略、密钥、端口、日志、镜像标签、调试工具、compose 配置共 12 个维度。

约定：

- **开发环境 Development**：服务开发者，目标「方便、快、可调试」。
- **生产环境 Production**：服务稳定与安全，目标「最小镜像、最小权限、最小攻击面」。
- **核心原则**：开发优先体验，生产优先稳定性与安全；两套配置通过 `compose profiles` 或独立文件分离，互不污染。

---

## 0. 总览表

| # | 维度 | 开发环境 | 生产环境 |
|---|------|---------|---------|
| 1 | Dockerfile | 单阶段，含全工具链 | 多阶段，仅运行时 |
| 2 | 依赖 | 含 devDependencies | `--omit=dev` 只留 prod |
| 3 | 源码 | bind mount 热重载 | COPY 进镜像 |
| 4 | 运行用户 | root（方便） | 非 root |
| 5 | 资源限制 | 不限 | `--memory`/`--cpus` |
| 6 | 重启策略 | `no` / `on-failure` | `unless-stopped` |
| 7 | 密钥 | `.env` 明文 | secret / 外部管理 |
| 8 | 端口 | 直接 `-p` 暴露 | 内网 + 反向代理 |
| 9 | 日志 | console 直看 | 集中收集（ELK） |
| 10 | 镜像标签 | `latest` / `dev` | 语义化版本 `v1.2.3` |
| 11 | 调试工具 | 全装 | 最小化 |
| 12 | compose | `profiles: [dev]` | `profiles: [prod]` |

---

## 1. Dockerfile 构建策略

### 1.1 不同

开发用单阶段、镜像大；生产用多阶段、镜像小。

### 1.2 简单处理（Node 单服务）

```dockerfile
# 开发：单阶段，包含全部构建工具
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install              # 装全部依赖含 dev
COPY . .
CMD ["npm", "run", "dev"]    # nodemon 热重载

# 生产：多阶段，运行镜像不含构建工具
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
CMD ["node", "dist/index.js"]
```

### 1.3 为什么生产用多阶段（详解）

多阶段构建解决的是一个核心矛盾：**「构建时需要的」和「运行时需要的」是完全不同的两套东西。**

用一个类比理解：

> 你要生产一辆汽车。
> - **构建阶段**需要：钢板、焊接机器人、冲压机、喷漆车间……这些是「造车工具」。
> - **运行阶段**只需要：车本身。
> - 如果把焊接机器人也装在车上一起卖，车会重得开不动，而且机器人本身是危险品（可能伤人）。

Docker 镜像同理。下面用一个**全栈项目**（Go 后端 API + React 前端 SPA）说明，这种场景多阶段的优势最明显，因为涉及两种完全不同的构建工具链。

### 1.4 复杂案例：Go 后端 + React 前端

项目结构：

```
myapp/
├── backend/              # Go API
│   ├── main.go
│   ├── go.mod
│   └── handlers/
├── frontend/             # React SPA
│   ├── package.json
│   ├── src/
│   └── vite.config.ts
└── Dockerfile
```

后端用 `go build` 编译成静态二进制，前端用 `npm run build` 产出静态 `dist/`。Go 二进制运行时不需要 Go 运行时，静态文件用 Go 内嵌的 `embed` 打进二进制。

#### 1.4.1 单阶段（错误写法）

```dockerfile
# ❌ 单阶段：所有工具链都进镜像
FROM golang:1.22-bookworm

# 装 Node 构建前端
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs

WORKDIR /app
# 拷源码
COPY backend/ ./backend/
COPY frontend/ ./frontend/

# 构建前端
WORKDIR /app/frontend
RUN npm ci && npm run build

# 构建后端
WORKDIR /app/backend
RUN go mod download && go build -o /app/server .

WORKDIR /app
EXPOSE 8080
CMD ["/app/server"]
```

构建结果：

```bash
$ docker images myapp:single
REPOSITORY   TAG     SIZE
myapp        single  1.25 GB    # ← 巨大
```

镜像里**白白带着**这些运行时根本用不到的东西：

| 冗余内容 | 大小（约） | 为什么运行时不需要 |
|---------|-----------|-------------------|
| Go 工具链（`go`、编译器、标准库源码） | ~600 MB | 已经 `go build` 成二进制了 |
| Node.js + npm | ~180 MB | 已经 `npm run build` 出 `dist/` 了 |
| `frontend/node_modules` | ~250 MB | 构建产物已生成 |
| `frontend/src`（TS 源码） | ~5 MB | 只要 `dist/` |
| `backend/*.go`（Go 源码） | ~1 MB | 只要编译后的二进制 |
| `curl`、`git`、`make` 等构建辅助工具 | ~80 MB | 运行时无人调用 |

**带来的实际问题：**

1. **拉取慢**：1.25 GB 镜像，部署到 10 台机器 = 传输 12.5 GB。
2. **攻击面大**：容器内被入侵后，攻击者能直接用 `curl` 发起 SSRF、用 `go`/`npm` 下载并运行恶意代码。
3. **漏洞多**：基础镜像 `golang:1.22-bookworm` 带 400+ 个 Debian 包，每个都是潜在 CVE。
4. **缓存易失效**：改一行业务代码，整个 1.25 GB 都要重新推。

#### 1.4.2 多阶段（正确写法）

```dockerfile
# ✅ 多阶段：构建工具留在 builder，运行镜像极简

# 阶段1：构建前端
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build              # 产出 dist/

# 阶段2：构建后端
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app/backend
COPY backend/go.* ./
RUN go mod download
COPY backend/ ./
# 把前端产物拷进来，用 embed 打进二进制
COPY --from=frontend-builder /app/frontend/dist ./internal/static/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server .

# 阶段3：运行镜像（最小化）
FROM alpine:3.19
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=backend-builder /app/server .
USER app
EXPOSE 8080
ENTRYPOINT ["/app/server"]
```

构建结果：

```bash
$ docker images myapp:multi
REPOSITORY   TAG     SIZE
myapp        multi   18 MB    # ← 缩小 70 倍
```

体积对比一览：

| 项目 | 单阶段 | 多阶段 | 差距 |
|------|--------|--------|------|
| 基础镜像 | `golang:1.22-bookworm` ~810 MB | `alpine:3.19` ~7 MB | -800 MB |
| Go 工具链 | 600 MB | 0（已剥离） | -600 MB |
| Node + npm | 180 MB | 0 | -180 MB |
| `node_modules` | 250 MB | 0（用 `embed` 打进二进制） | -250 MB |
| 应用本体 | 二进制 ~15 MB + dist | 二进制 ~18 MB（含前端） | 持平 |
| **总计** | **~1.25 GB** | **~18 MB** | **约 70×** |

#### 1.4.3 关键技巧解读

| 写法 | 作用 |
|------|------|
| `FROM ... AS frontend-builder` | 给阶段起名，后面 `COPY --from=` 引用 |
| `COPY --from=frontend-builder /app/frontend/dist ./...` | 只拷产物，不拷源码和工具 |
| `CGO_ENABLED=0` | 纯静态编译，运行镜像不需要 libc，可用 `scratch` |
| `-ldflags="-s -w"` | 去掉调试符号和 DWARF 表，二进制再小 30% |
| `go:embed`（在 main.go 里 `//go:embed static/dist`） | 把前端静态文件嵌进二进制，运行镜像连 dist 目录都不用单独放 |
| 最终 `FROM alpine:3.19` 而非 `golang` | 运行镜像只含 OS + 二进制，无任何编译器 |
| 极致可用 `FROM scratch` | 连 OS 都没有，只有二进制本身（~15 MB），但调试困难需谨慎 |

#### 1.4.4 多阶段的额外收益

除了体积，多阶段还带来这些隐性好处：

1. **层缓存利用率更高**：改前端代码只重建 stage1 + stage3；改后端只重建 stage2 + stage3。Docker 会复用未变阶段的缓存。
2. **构建可并行**：BuildKit 下两个不互相依赖的 builder 阶段可并行跑（本例 stage2 依赖 stage1 的产物，但更复杂的项目里多个 builder 并行能省一半时间）。
3. **强制分离关注点**：开发者天然被引导思考「这个文件运行时用不用得到」，否则会被 `COPY --from` 这道关卡拦住。
4. **供应链安全**：运行镜像不含 `git`/`curl`/`sh`（用 `scratch` 时甚至没有 shell），被入侵后攻击者无法下载后续 payload。
5. **CVE 面收窄**：运行镜像只依赖 `alpine` 基础包（几十个），扫描器报的漏洞数远少于含 400+ Debian 包的构建镜像。

### 1.5 单阶段 vs 多阶段 完整对比

| 维度 | 单阶段 | 多阶段 |
|------|--------|--------|
| 镜像体积 | 1.25 GB（本例） | 18 MB（本例） |
| 拉取/推送时间 | 慢 | 快 70 倍 |
| 攻击面 | 大（含编译器/包管理器/shell） | 极小 |
| CVE 数量 | 多 | 少 |
| 构建缓存命中 | 差（一改全重建） | 好（分层缓存） |
| 可并行构建 | 否 | 是（BuildKit） |
| 运行时启动 | 慢（镜像大） | 快 |
| 适合场景 | 仅开发 | 生产必用 |

### 1.6 什么时候可以不用多阶段

- **开发环境**：要频繁 `exec` 进容器调试，工具齐全反而方便，单阶段可接受。
- **纯脚本类镜像**（如一个只有 10 行 Python 的容器）：直接 `FROM python:3-alpine` + `COPY` 即可，多阶段收益不明显。
- **镜像只在本机用、从不推送**：体积无所谓，多阶段是过度设计。

其余情况，特别是**涉及编译型语言、多语言混合、需要构建产物分离的场景，多阶段是必选项。**

### 1.7 自测要点

- [ ] `docker images` 对比两套镜像大小，生产镜像应小一个数量级（本例 70 倍）。
- [ ] 进生产镜像 `exec sh`，确认没有 `npm`/`go`/`tsc`/`curl` 等构建工具。
- [ ] `docker history <image>` 运行镜像层数合理（应只有 `COPY` 二进制 + `USER` + `CMD` 几层）。
- [ ] `docker history <image>` 中无 `RUN go build` / `RUN npm install` 等构建层（它们应留在 builder 阶段，不会进最终镜像）。
- [ ] 改一行业务代码重建，`go mod download` 层命中缓存（不重新下载依赖）。
- [ ] 用 `trivy image <image>` 扫描，生产镜像 CVE 数显著少于单阶段镜像。
- [ ] 生产镜像以非 root 用户启动。

---

## 2. 依赖安装

### 2.1 不同

开发装全部依赖；生产只装运行时依赖。

### 2.2 处理

```bash
# 开发
npm install                    # 含 devDependencies

# 生产
npm ci --omit=dev              # 只装 production 依赖
```

### 2.3 为什么

- devDependencies 含 `typescript`、`eslint`、`nodemon` 等开发和构建工具，运行时根本用不到。
- 生产镜像里留着它们：① 体积膨胀几十到几百 MB；② 多了不该有的可执行文件，扩大攻击面。

### 2.4 自测要点

- [ ] 生产镜像内 `ls node_modules` 无 `typescript`/`eslint` 等 dev 包。
- [ ] `npm ls --omit=dev` 在生产镜像里能正常列出运行时依赖。

---

## 3. 源码处理

### 3.1 不同

开发用 bind mount 挂载源码实现热重载；生产把源码 `COPY` 进镜像。

### 3.2 处理

```yaml
# 开发 compose
services:
  web:
    build:
      dockerfile: Dockerfile.dev
    volumes:
      - ./src:/app/src          # 改代码容器内即时生效
      - /app/node_modules        # 防主机覆盖容器依赖（匿名卷）
    command: npm run dev

# 生产 compose
services:
  web:
    build: .
    # 不挂载源码，镜像即唯一真相
```

### 3.3 为什么

- 开发改一行代码就重建镜像太慢，bind mount 让本地编辑器改完容器内立即生效。
- 生产中，**镜像是不可变制品（immutable artifact）**，必须自包含全部代码与依赖，保证「构建一次，到处运行」。挂载宿主机目录会破坏可重复性，且生产机不一定有源码。

### 3.4 自测要点

- [ ] 开发环境改一行代码，容器内日志立即显示热重载。
- [ ] 生产镜像在另一台无源码的机器上 `docker run` 能正常启动。
- [ ] 开发环境的 `/app/node_modules` 没被宿主机目录覆盖（匿名卷生效）。

---

## 4. 运行用户

### 4.1 不同

开发常以 root 运行（方便）；生产必须非 root。

### 4.2 处理

```dockerfile
# 开发：默认 root，方便装包、写文件

# 生产：创建专用用户
RUN addgroup -S app && adduser -S app -G app
USER app
```

### 4.3 为什么

- 开发时 root 方便进容器装 `vim`/`curl`、写任意目录，不必纠结权限。
- 生产一旦容器被攻破（如 RCE 漏洞），root 容器意味着攻击者可借此提权到宿主机。非 root 把损害范围限制在容器内，是纵深防御的基本一环。

### 4.4 自测要点

- [ ] 生产容器内 `whoami` 返回 `app` 而非 `root`。
- [ ] 生产容器内尝试 `apk add` 或写 `/etc` 失败（权限拒绝）。

---

## 5. 资源限制

### 5.1 不同

开发不限制；生产严格限制。

### 5.2 处理

```bash
# 开发：不限制，随便吃
docker run -d myapp

# 生产：限制内存和 CPU
docker run -d \
  --memory=512m \
  --memory-swap=512m \
  --cpus=1 \
  --pids-limit=200 \
  myapp
```

### 5.3 为什么

- 开发机就一两个容器，不限也无所谓。
- 生产一台宿主机跑几十个容器，**一个容器内存泄漏会拖垮整机（OOM 影响邻居）**。资源限制把爆炸半径控制在一个容器内，符合「故障隔离」原则。

### 5.4 自测要点

- [ ] 生产容器 `docker stats` 显示的内存上限符合配置。
- [ ] 模拟内存泄漏，容器被 OOM kill 而非拖垮宿主机。

---

## 6. 重启策略

### 6.1 不同

开发可用 `no` 或 `on-failure`；生产用 `unless-stopped` 或配合编排器。

### 6.2 处理

| 环境 | 策略 | 原因 |
|------|------|------|
| 开发 | `no` 或 `on-failure:3` | 崩了想立刻看到，自动重启反而掩盖问题 |
| 生产 | `unless-stopped` | 宿主机重启后服务自动恢复 |
| 大规模生产 | K8s/Swarm 编排器接管 | 单机 `restart` 不够，需滚动更新、健康检查、自愈 |

### 6.3 为什么

开发要第一时间暴露 bug；生产要最大化可用性，临时故障应自动恢复，避免人为介入。

### 6.4 自测要点

- [ ] 开发容器崩溃后不自动重启，能看到退出码。
- [ ] 生产容器崩溃后自动恢复，`docker ps` 显示 `Up`。
- [ ] 宿主机重启后，`unless-stopped` 的容器自动起来。

---

## 7. 密钥与环境变量

### 7.1 不同

开发用 `.env` 明文；生产用外部密钥管理。

### 7.2 处理

```bash
# 开发：.env 文件，明文存 token（已被 .gitignore 排除）
docker run --env-file .env myapp

# 生产：密钥不入镜像，运行时注入
docker run -e DATABASE_URL=$DATABASE_URL myapp
# 或用 Docker Secret（Swarm）/ Vault / 云厂商 KMS
```

### 7.3 为什么

- 开发环境方便优先，明文 `.env` 已被 `.gitignore` 排除，可接受。
- 生产镜像会被推到仓库、可能被多人拉取，**密钥一旦烤进镜像等于公开**。即使后续删除层，旧层仍可被 `docker history` 还原。所以密钥必须运行时注入。

### 7.4 错误写法对比

```dockerfile
# 错误：把密钥写进 Dockerfile
ENV DATABASE_URL=postgres://user:password@db:5432/prod
# 镜像任何拉取者都能 docker history 看到

# 正确：运行时注入
# CMD ["node", "dist/index.js"]
# 运行时：docker run -e DATABASE_URL=xxx myapp
```

### 7.5 自测要点

- [ ] `docker history <生产镜像>` 中无任何密钥痕迹。
- [ ] `.env` 文件在 `.gitignore` 中。
- [ ] 生产容器 `env` 能看到注入的变量，但镜像层里搜不到。

---

## 8. 端口与网络

### 8.1 不同

开发直接 `-p` 暴露；生产走反向代理 + 内网。

### 8.2 处理

```bash
# 开发：端口直接映射到主机
docker run -p 8080:80 myapp      # 访问 localhost:8080

# 生产：容器只在内部网络，由 Nginx/Traefik 反代
docker network create prod-net
docker run -d --network prod-net --name app myapp   # 不 -p
docker run -d --network prod-net -p 80:80 nginx     # 只有反代暴露
```

### 8.3 为什么

- 开发要本机直接访问调试。
- 生产中，直接暴露容器端口意味着每个服务都是攻击入口。反向代理统一 TLS、限流、WAF，且容器间用名字互访不经过公网，缩小暴露面。

### 8.4 自测要点

- [ ] 生产环境 `docker ps` 中应用容器无端口映射，只有反代容器暴露端口。
- [ ] 容器间能用服务名互访（如 `redis://redis:6379`）。
- [ ] 外部无法直接访问应用容器端口。

---

## 9. 日志

### 9.1 不同

开发直接看 console；生产集中收集。

### 9.2 处理

```bash
# 开发
docker logs -f myapp

# 生产
# 容器 stdout/stderr → json-file 或 logging driver
# → ELK / Loki / CloudWatch 集中查询、告警、保留
docker run --log-driver=json-file \
  --log-opt max-size=10m --log-opt max-file=3 \
  myapp
```

### 9.3 为什么

- 开发容器少，`docker logs -f` 够用。
- 生产几十上百容器分布在多台机器，**单机日志无意义**，必须集中才能关联查询、设置告警、满足审计合规。同时要限制单文件大小防止撑爆磁盘。

### 9.4 自测要点

- [ ] 生产容器日志能被 ELK/Loki 采集到。
- [ ] 单个日志文件不超过 `max-size`，旧文件按 `max-file` 轮转。
- [ ] 日志中无敏感信息（密码/token）。

---

## 10. 镜像标签与版本

### 10.1 不同

开发用 `latest`/`dev`；生产用语义化版本。

### 10.2 处理

```bash
# 开发
docker build -t myapp:dev .           # 反复覆盖

# 生产
docker build -t myapp:1.2.3 .
docker build -t myapp:1.2.3-abc1234 .  # 加 git short sha 可追溯
```

### 10.3 为什么

- 开发不在乎标签，覆盖即可。
- 生产**绝不能用 `latest`**：`latest` 会漂移，回滚时不知道具体跑的是哪个版本。语义化版本 + git sha 让每次发布可追溯、可回滚、可审计。

### 10.4 自测要点

- [ ] 生产镜像标签为 `1.2.3` 或 `1.2.3-<sha>`，无 `latest`。
- [ ] 通过 `docker image inspect` 能看到镜像构建时间和层 sha。
- [ ] 回滚时能精确指定上一个版本标签。

---

## 11. 调试工具

### 11.1 不同

开发镜像装 `curl`/`vim`/`ps`/`ping`；生产最小化。

### 11.2 为什么

- 开发要频繁进容器排查。
- 生产里这些工具本身是攻击面（`curl` 可被用于 SSRF、`bash` 被入侵后利用）。生产镜像越精简越安全，排查应通过日志/监控完成，确实要调试用临时 `ephemeral container` 共享命名空间排查。

### 11.3 自测要点

- [ ] 生产镜像内 `which curl vim ping` 全部不存在。
- [ ] 开发镜像内能直接 `curl localhost:80` 排查。

---

## 12. compose 文件分离实践

### 12.1 处理

用 `profiles` 或独立文件把两套配置分开：

```yaml
# docker-compose.yml（公共基础）
services:
  app:
    build: .
    profiles: ["prod"]
    user: "app"
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 512m

  app-dev:
    build:
      dockerfile: Dockerfile.dev
    profiles: ["dev"]
    volumes:
      - ./src:/app/src
    command: npm run dev

  db:
    image: postgres:16      # 公共依赖，两个环境都用
```

```bash
docker compose --profile dev up     # 开发
docker compose --profile prod up -d # 生产
```

### 12.2 为什么

同一份 compose 描述基础设施（如 db、redis），用 profiles 切换 app 服务配置，避免维护两套重复文件。

### 12.3 自测要点

- [ ] `docker compose --profile dev config` 和 `--profile prod config` 输出不同。
- [ ] 公共服务（db）在两个 profile 下都启动。
- [ ] 开发 profile 不带资源限制和 user 字段。

---

## 附录 A：环境差异决策清单

部署前逐项确认：

1. **镜像构建** → 是否多阶段？运行镜像是否最小化？
2. **依赖** → 是否 `--omit=dev`？
3. **源码** → 是否 COPY 进镜像（非挂载）？
4. **运行用户** → 是否非 root？
5. **资源限制** → 是否设了 `--memory`/`--cpus`？
6. **重启策略** → 是否 `unless-stopped`？
7. **密钥** → 是否运行时注入，未烤进镜像？
8. **端口** → 是否走反代，未直接暴露？
9. **日志** → 是否集中收集 + 轮转限制？
10. **镜像标签** → 是否语义化版本，无 `latest`？
11. **调试工具** → 生产镜像是否无多余工具？
12. **compose** → 是否用 `profiles` 分离？

---

## 附录 B：一键自测脚本

```bash
#!/bin/sh
# 生产镜像上线前自检脚本
IMAGE=$1

echo "=== 检查镜像标签（不能是 latest）==="
echo "$IMAGE" | grep -q ":latest" && echo "FAIL: 使用了 latest 标签" || echo "PASS"

echo "=== 检查镜像内是否有 curl（不应有）==="
docker run --rm "$IMAGE" which curl 2>/dev/null && echo "FAIL: 含 curl" || echo "PASS"

echo "=== 检查运行用户（应非 root）==="
docker run --rm "$IMAGE" whoami 2>/dev/null | grep -q "^root$" && echo "FAIL: root 运行" || echo "PASS"

echo "=== 检查 Dockerfile 历史中是否有密钥 ==="
docker history "$IMAGE" | grep -iE "(password|secret|token|api_key)=" && echo "FAIL: 历史层含密钥" || echo "PASS"

echo "=== 检查镜像大小 ==="
docker images "$IMAGE" --format "{{.Size}}"
```

---

## 附录 C：环境配置速查

| 配置项 | 开发 | 生产 |
|--------|------|------|
| Dockerfile | `Dockerfile.dev` | `Dockerfile`（多阶段） |
| 依赖命令 | `npm install` | `npm ci --omit=dev` |
| 启动命令 | `npm run dev` | `node dist/index.js` |
| 源码 | bind mount | COPY 进镜像 |
| 用户 | root | 非 root |
| 内存限制 | 无 | `--memory=512m` |
| CPU 限制 | 无 | `--cpus=1` |
| 重启策略 | `no` | `unless-stopped` |
| 密钥来源 | `.env` | 外部注入 |
| 端口暴露 | `-p 8080:80` | 反代内网 |
| 日志 | console | json-file + ELK |
| 镜像标签 | `:dev` | `:1.2.3` |
| 调试工具 | 全装 | 无 |
| compose profile | `dev` | `prod` |

---

## 附录 D：升级时机参考

| 信号 | 建议动作 |
|------|---------|
| 开发和用同一份 Dockerfile | 拆分为 `Dockerfile.dev` + 多阶段 `Dockerfile` |
| 生产镜像超过 1GB | 引入多阶段 + `--omit=dev` |
| 生产容器被 OOM 拖垮 | 加 `--memory` 限制 |
| 生产用 `latest` 标签 | 改语义化版本 + git sha |
| 生产镜像内能 `curl` | 切最小基础镜像（如 `alpine`/`distroless`） |
| 密钥写进 Dockerfile | 改运行时注入 + `docker history` 验证 |
| 容器直接暴露公网 | 加反向代理 + 内网网络 |
| 单机 `restart` 不够用 | 引入 K8s/Swarm 编排 |
