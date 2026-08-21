# Docker 使用技巧与 Dockerfile / compose 对比完整整理

独立专题。内容分为两部分：

- **第一部分**：Docker 使用技巧（Dockerfile 优化、容器运行、调试、数据持久化、compose、构建提速、清理、安全）。
- **第二部分**：Dockerfile 与 docker compose 的核心区别、分工与配合关系。

约定：

- **Image 镜像**：静态模板，由 Dockerfile 构建产出。
- **Container 容器**：镜像的运行实例，由 `docker run` 或 `docker compose up` 启动。
- **Dockerfile 造镜像，compose 编容器；build 出 image，up 起一堆。**

---

## 第一部分：Docker 使用技巧

### 0. 总览表

| # | 主题 | 一句话要点 |
|---|------|-----------|
| 1 | Dockerfile 优化 | 多阶段构建 + 合并 RUN + .dockerignore |
| 2 | 容器运行参数 | `-d` 后台、`-p` 端口、`-v` 卷、`-e` 环境变量 |
| 3 | 调试三板斧 | `exec` 进容器、`logs` 看日志、`inspect` 看状态 |
| 4 | 数据持久化与网络 | Volume 优于 bind；自定义网络用容器名互访 |
| 5 | docker compose 技巧 | `depends_on` + `healthcheck` + `profiles` |
| 6 | 构建提速 | BuildKit + 缓存挂载 + buildx 多平台 |
| 7 | 清理瘦身 | `system df` 体检，`prune` 清理 |
| 8 | 安全最佳实践 | 非 root 运行 + 资源限制 + 密钥不入镜像 |

### 1. Dockerfile 优化（构建更小更快）

#### 1.1 要点

| 技巧 | 作用 |
|------|------|
| 多阶段构建 | 用 `FROM ... AS builder` 编译，再 `COPY --from=builder` 到运行镜像，体积可从 GB 降到几十 MB |
| 合并 RUN 层 | 每条 `RUN` 产生一层，用 `&&` 串联减少层数 |
| 善用 .dockerignore | 排除 `node_modules`、`.git`、`dist`，避免无用文件进入上下文 |
| 依赖先 COPY | 先 `COPY package*.json` 再 `COPY .`，利用层缓存，改代码不重装依赖 |
| 固定基础镜像版本 | 用 `node:20-alpine` 而非 `node:latest`，避免隐式升级踩坑 |

#### 1.2 示范 Dockerfile

```dockerfile
# 多阶段构建：builder 阶段负责编译
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# 运行阶段：只携带产物，体积小
FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
CMD ["node", "dist/index.js"]
```

#### 1.3 错误写法对比

```dockerfile
# 错误：每条 RUN 一层，且把源码和依赖一起拷，改代码就重装依赖
FROM node:20-alpine
WORKDIR /app
COPY . .
RUN npm install
RUN npm run build
RUN npm prune --production
# 镜像里残留构建工具、源码，体积大、缓存易失效
```

#### 1.4 自测要点

-  用 `docker images` 对比优化前后镜像大小。
-  改一行业务代码重新 `build`，是否触发 `npm ci`（不应该触发）。
-  `docker history <image>` 查看层数是否合理。

---

### 2. 容器运行常用参数

#### 2.1 标准模板

```bash
docker run -d \                  # 后台运行
  --name myapp \                 # 起个好记的名字
  -p 8080:80 \                   # 端口映射 主机:容器
  -v $(pwd)/data:/app/data \     # 数据卷挂载
  -e NODE_ENV=production \       # 环境变量
  --restart unless-stopped \     # 重启策略
  --memory=512m --cpus=1 \       # 资源限制
  myimage:latest
```

#### 2.2 重启策略对比

| 策略 | 行为 | 适用场景 |
|------|------|---------|
| `no` | 不重启（默认） | 临时任务 |
| `on-failure:5` | 非零退出码才重启，最多 5 次 | 可能瞬时失败的服务 |
| `always` | 总是重启，包括手动 stop 后开机自启 | 常驻服务 |
| `unless-stopped` | 总是重启，但手动 stop 后不再自启（推荐） | 大多数生产服务 |

#### 2.3 常用标志速查

| 标志 | 作用 |
|------|------|
| `--rm` | 容器退出即删除，适合跑测试 |
| `-it` | 交互式终端，进容器用 |
| `--network mynet` | 加入自定义网络 |
| `--read-only` | 文件系统只读，防写入 |
| `--tmpfs /tmp` | 临时内存目录 |

#### 2.4 自测要点

-  `docker ps` 能看到容器状态为 `Up`。
-  `curl localhost:8080` 能通。
-  宿主机重启后，`unless-stopped` 的容器是否自动起来。

---

### 3. 调试三板斧

#### 3.1 常用调试命令

```bash
docker exec -it myapp sh              # 进容器看现场（alpine 用 sh，ubuntu 用 bash）
docker logs -f --tail 100 myapp       # 实时跟踪最后 100 行
docker inspect myapp | grep -A5 State # 看退出码/状态
docker stats                          # 实时 CPU/内存占用
docker top myapp                      # 看容器内进程
```

#### 3.2 容器内无 curl/ping 的排查技巧

容器里通常没装 `curl`/`ping`，用共享网络栈的方式从外部容器排查：

```bash
docker run --rm --network container:myapp alpine sh -c "apk add curl && curl localhost:80"
```

#### 3.3 退出码速查

| 退出码 | 含义 |
|--------|------|
| `0` | 正常退出 |
| `1` | 应用错误 |
| `125` | docker 自身错误 |
| `126` | 命令不可执行 |
| `127` | 命令未找到 |
| `137` | 被 SIGKILL（常因 OOM 或 `docker kill`） |
| `139` | 段错误（SIGSEGV） |
| `143` | 收到 SIGTERM 正常终止 |

#### 3.4 自测要点

-  能用 `exec` 进容器查看 `/app` 目录。
-  `logs -f` 能实时滚动看到新日志。
-  容器异常退出时，能通过 `inspect` 的 `State.Error` 定位原因。

---

### 4. 数据持久化与网络

#### 4.1 三种挂载方式对比

| 方式 | 命令 | 管理方 | 适用场景 |
|------|------|--------|---------|
| Volume | `-v mydata:/data` | Docker | 生产数据，跨平台安全（推荐） |
| Bind mount | `-v /host/path:/container/path` | 用户 | 开发热重载，权限/路径易踩坑 |
| tmpfs | `--tmpfs /cache` | 内存 | 临时敏感数据，不落盘 |

#### 4.2 Volume 操作

```bash
docker volume create mydata
docker run -d -v mydata:/data myimage
# 备份卷
docker run --rm -v mydata:/data -v $(pwd):/backup alpine \
  tar czf /backup/data.tgz /data
# 恢复卷
docker run --rm -v mydata:/data -v $(pwd):/backup alpine \
  tar xzf /backup/data.tgz -C /
```

#### 4.3 自定义网络（容器名互访）

```bash
docker network create mynet
docker run -d --name redis --network mynet redis:alpine
docker run -d --name app --network mynet \
  -e REDIS_URL=redis://redis:6379 myapp
# app 容器内可直接用 redis://redis:6379 访问，Docker 内置 DNS
```

#### 4.4 自测要点

-  删除容器后重新 `run`，Volume 数据仍在。
-  两个容器在同一自定义网络内，能用容器名 ping 通。
-  `docker volume ls` 能看到创建的卷。

---

### 5. docker compose 技巧

#### 5.1 完整示例

```yaml
services:
  web:
    build: .
    ports: ["8080:80"]
    volumes: ["./src:/app/src"]      # 开发热重载
    env_file: .env                  # 统一管理环境变量
    depends_on:
      db:
        condition: service_healthy  # 等依赖健康再启动
    profiles: ["dev"]               # 按环境分组，--profile dev 才起
  db:
    image: postgres:16
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U app"]
      interval: 5s
```

#### 5.2 常用命令

| 命令 | 作用 |
|------|------|
| `docker compose up -d` | 后台启动全部服务 |
| `docker compose logs -f web` | 跟踪某服务日志 |
| `docker compose exec web sh` | 进容器 |
| `docker compose up -d --build` | 改代码后重新构建 |
| `docker compose down` | 停并删容器（保留数据） |
| `docker compose down -v` | 连卷一起删（慎用，数据没了） |
| `docker compose --profile dev up` | 按环境启动 |

#### 5.3 自测要点

-  `docker compose config` 能正确解析 YAML。
-  `db` 健康检查通过后 `web` 才启动。
-  `down` 后卷数据仍在；`down -v` 后卷被删除。

---

### 6. 构建提速

#### 6.1 开启 BuildKit

```bash
DOCKER_BUILDKIT=1 docker build -t app .
# 或在 ~/.docker/config.json 中设 "features": {"buildkit": true}
```

#### 6.2 缓存挂载（包管理器缓存跨构建复用）

```dockerfile
RUN --mount=type=cache,target=/root/.npm \
    npm ci
```

#### 6.3 多平台构建

```bash
docker buildx build --platform linux/amd64,linux/arm64 -t img . --push
```

#### 6.4 自测要点

-  对比开启 BuildKit 前后的构建时间。
-  改一行代码重建，`npm ci` 是否命中缓存。
-  `buildx` 能同时产出 amd64 和 arm64 镜像。

---

### 7. 清理瘦身

#### 7.1 清理命令

```bash
docker system df                          # 看磁盘占用全景
docker system prune -a --volumes          # 一键清理（谨慎）
docker image prune -a                     # 删除未被容器使用的镜像
docker container prune                    # 删除停止的容器
docker rm $(docker ps -aqf status=exited) # 批量删已退出容器
docker rmi $(docker images -f "dangling=true" -q)  # 删悬空镜像
```

#### 7.2 清理动作影响范围

| 命令 | 删除对象 | 是否影响运行容器 |
|------|---------|----------------|
| `container prune` | 已停止的容器 | 否 |
| `image prune` | 悬空镜像（无标签） | 否 |
| `image prune -a` | 未被任何容器使用的镜像 | 否 |
| `volume prune` | 未被容器引用的卷 | **是（数据丢失）** |
| `system prune -a --volumes` | 以上全部 | 否（但卷数据没了） |

#### 7.3 自测要点

-  定期 `docker system df` 检查占用。
-  `prune --volumes` 前确认没有要保留的孤立卷。
-  生产环境避免用 `-a --volumes`，改用精确的 `image prune`。

---

### 8. 安全最佳实践

#### 8.1 非 root 运行

```dockerfile
RUN adduser -D appuser
USER appuser
```

#### 8.2 只读文件系统 + 临时目录

```bash
docker run --read-only --tmpfs /tmp myimage
```

#### 8.3 资源限制

```bash
docker run --memory=512m --cpus=1 --pids-limit=100 myimage
```

#### 8.4 镜像扫描

```bash
docker scout quickview       # Docker 官方
trivy image myimage          # 第三方，查 CVE
```

#### 8.5 密钥处理对比

| 错误做法 | 正确做法 |
|---------|---------|
| `ENV API_KEY=xxx` 写进 Dockerfile | 运行时 `-e` 或 `--env-file` 注入 |
| 把 `.env` 文件 `COPY` 进镜像 | `.env` 加入 `.dockerignore` |
| 在镜像里写死数据库密码 | 用 `docker secret`（Swarm）或外部密钥管理 |

#### 8.6 自测要点

-  容器内 `whoami` 不是 root。
-  `--read-only` 下应用仍能正常写 `/tmp`。
-  `trivy image` 扫描无高危 CVE。

---

## 第二部分：Dockerfile vs docker compose

### 9. 核心区别一句话

- **Dockerfile**：菜谱 —— 描述「一个镜像怎么从零搭起来」。
- **docker compose**：餐桌排布 —— 描述「几个容器怎么组在一起、怎么互通、怎么共享数据」。

### 10. 定位与层次对比

| 维度 | Dockerfile | docker compose |
|---|---|---|
| 解决问题 | 单个镜像怎么构建 | 多个容器怎么编排运行 |
| 文件 | `Dockerfile` | `docker-compose.yml` |
| 产出物 | Image（静态模板） | 一组运行中的容器 |
| 关注点 | 基础镜像、依赖安装、代码打包、启动命令 | 端口映射、卷挂载、网络、依赖顺序、环境变量 |
| 是否感知其他容器 | 否 | 是（`depends_on`、共享网络） |
| 典型命令 | `docker build -t app .` | `docker compose up -d` |

### 11. 配合关系（不是二选一）

Compose 文件里每个 service 有两种取镜像方式：

```yaml
services:
  web:
    build: .                 # 方式A：指向一个 Dockerfile，自动构建
    # image: myapp:latest    # 方式B：直接用现成镜像
    ports: ["8080:80"]

  db:
    image: postgres:16       # 现成镜像，不需要 Dockerfile
```

- `build: .` 会调用同目录的 `Dockerfile` 构建出镜像，再以该镜像启动容器。
- 所以 **Compose 包含 Dockerfile 的使用**，是上层编排，Dockerfile 是底层构建。

### 12. 典型分工

**Dockerfile 负责**（容器内部的事）：

- 选基础镜像 `FROM node:20-alpine`
- 装依赖 `RUN npm ci`
- 拷代码 `COPY . .`
- 暴露端口 `EXPOSE 8080`
- 启动命令 `CMD ["node", "dist/index.js"]`

**docker compose 负责**（容器之间的事）：

- `ports` 主机端口 ↔ 容器端口映射
- `volumes` 数据持久化、热重载挂载
- `environment` / `env_file` 注入配置
- `depends_on` 启动顺序、健康检查
- `networks` 自定义网络让容器用名字互访
- `profiles` 区分 dev/prod 环境

### 13. 何时只用其中一个

| 场景 | 用什么 |
|------|--------|
| 单容器、无外部依赖（如打包 CLI 工具镜像） | 只用 Dockerfile |
| 所有服务都用现成镜像（`redis` + `postgres` + 官方镜像） | 只用 compose |
| 自定义应用 + 数据库/缓存一起跑（最常见） | 两者配合 |

### 14. 完整示例对照

#### 14.1 Dockerfile（构建 Node 应用镜像）

```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev
COPY . .
CMD ["node", "index.js"]
```

#### 14.2 docker-compose.yml（应用 + Redis 一起跑）

```yaml
services:
  app:
    build: .                  # 调用上面的 Dockerfile
    ports: ["8080:3000"]
    environment:
      REDIS_URL: redis://redis:6379   # 用服务名访问
    depends_on: [redis]
  redis:
    image: redis:7-alpine
    volumes: ["redis-data:/data"]
volumes:
  redis-data:
```

一条命令 `docker compose up -d` 就同时构建镜像、起容器、建网络、挂卷，比手敲一串 `docker run` 可靠得多。

---

## 附录 A：常用命令速查表

### A.1 镜像命令

| 命令 | 作用 |
|------|------|
| `docker pull <image>` | 拉取镜像 |
| `docker images` | 列出本地镜像 |
| `docker rmi <image>` | 删除镜像 |
| `docker tag <src> <dst>` | 打标签 |
| `docker save -o file.tar <image>` | 导出镜像 |
| `docker load -i file.tar` | 导入镜像 |
| `docker build -t <name> .` | 构建 |

### A.2 容器命令

| 命令 | 作用 |
|------|------|
| `docker run ...` | 创建并启动 |
| `docker ps -a` | 列出所有容器（含停止） |
| `docker stop <name>` | 停止 |
| `docker start <name>` | 启动已停止的容器 |
| `docker rm <name>` | 删除容器 |
| `docker rename <old> <new>` | 重命名 |

### A.3 调试命令

| 命令 | 作用 |
|------|------|
| `docker exec -it <name> sh` | 进容器 |
| `docker logs -f <name>` | 跟踪日志 |
| `docker inspect <name>` | 查看详情 |
| `docker top <name>` | 看容器内进程 |
| `docker stats` | 实时资源占用 |

### A.4 清理命令

| 命令 | 作用 |
|------|------|
| `docker system df` | 查看磁盘占用 |
| `docker system prune` | 清理停止的容器+悬空镜像 |
| `docker volume prune` | 清理未使用的卷 |
| `docker builder prune` | 清理构建缓存 |

### A.5 compose 一条龙

```bash
# 改代码 → 重建 → 重启
docker compose up -d --build

# 看日志排错
docker compose logs -f --tail=200

# 干净退出（保留数据）
docker compose down

# 连卷一起删（慎用，数据没了）
docker compose down -v
```

---

## 附录 B：决策流程

遇到任务时按以下顺序判断：

1. **是否需要自定义镜像？**
   - 是 → 写 Dockerfile
   - 否 → 直接用现成镜像

2. **是否涉及多个容器？**
   - 是 → 写 docker-compose.yml
   - 否 → 单个 `docker run`

3. **是否需要数据持久化？**
   - 是 → 用 Volume（生产）或 bind mount（开发）
   - 否 → 不挂卷，容器删了数据就没了

4. **是否需要区分环境？**
   - 是 → compose 用 `profiles`
   - 否 → 单一 compose 文件即可

5. **是否进入生产？**
   - 是 → 非 root + 资源限制 + 只读文件系统 + 镜像扫描
   - 否 → 开发环境可简化

---

## 附录 C：自测清单汇总

### Dockerfile
-  多阶段构建，运行镜像不含构建工具
-  有 `.dockerignore`，排除 `node_modules`/`.git`/`dist`
-  依赖文件先于源码 COPY
-  基础镜像固定版本（非 `latest`）
-  以非 root 用户运行

### 容器运行
-  设置 `--restart unless-stopped`
-  配置资源限制 `--memory`/`--cpus`
-  端口/卷/环境变量配置正确

### compose
-  `depends_on` 带 `condition: service_healthy`
-  敏感配置走 `env_file` 而非写死
-  用 `profiles` 区分环境
-  `down` 与 `down -v` 分清

### 安全
-  容器内 `whoami` 非 root
-  无密钥写进镜像
-  `trivy image` 无高危 CVE
-  生产用 `--read-only` + `--tmpfs`

### 运维
-  定期 `docker system df` 体检
-  清理前确认卷数据是否要保留
-  备份重要 Volume

---

## 附录 D：升级时机参考

| 信号 | 建议动作 |
|------|---------|
| 镜像超过 1GB | 引入多阶段构建 |
| 构建超过 5 分钟 | 开启 BuildKit + 缓存挂载 |
| 手敲一串 `docker run` 维护多容器 | 改用 docker compose |
| 容器频繁 OOM | 加 `--memory` 限制 + 查内存泄漏 |
| 磁盘占用持续增长 | 定期 `system prune` + 清理旧镜像 |
| 需要部署到 ARM（如 M1/树莓派） | 用 `buildx` 多平台构建 |
| 生产环境上线 | 补齐非 root + 资源限制 + 镜像扫描 |
