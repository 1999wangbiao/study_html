# Gin 依赖组装演变：从「全写路由里」到 Repository + Middleware

> 本目录**独立可学、可跑**，不依赖其它工程代码。  
> 系统复习：  
> - **[学习笔记-Handler与Service.md](./学习笔记-Handler与Service.md)**（v0→v4：Handler / Service / 手动 DI）  
> - **[学习笔记-Repository与Middleware.md](./学习笔记-Repository与Middleware.md)**（v5 门禁分层 + v6 Google Wire）

用**同一套登录业务**演示分层怎么长出来；v5 补门禁与 Repository，v6 用 Wire 生成组装代码。

业务刻意保持可 curl：

- 各版均提供 `POST /api/auth/login`（账号见该版注释）
- 数据用内存结构，不强制装 MySQL
- 每版独立目录、独立端口，避免互相干扰

| 版本 | 目录 | 端口 | 一句话 |
|------|------|------|--------|
| v0 | [`v0-inline/`](./v0-inline/) | 8080 | 逻辑全写在路由回调里 |
| v1 | [`v1-handler-fn/`](./v1-handler-fn/) | 8081 | 抽出函数，依赖仍是包级全局 |
| v2 | [`v2-handler-db/`](./v2-handler-db/) | 8082 | Handler 结构体持有「db」 |
| v3 | [`v3-service/`](./v3-service/) | 8083 | `db → Service → Handler` |
| v4 | [`v4-multi/`](./v4-multi/) | 8084 | 多 Service/Handler 组装 |
| v5 | [`v5-repo-mw/`](./v5-repo-mw/) | 8085 | Repository + Session + Middleware + AppError |
| v6 | [`v6-wire/`](./v6-wire/)（[说明](./v6-wire/README.md)） | 8086 | 业务同 v5；用 Google Wire 生成组装 |

---

## 推荐阅读 / 运行顺序

按 v0 → v6 顺序看代码，每版只多一步抽象。

```powershell
cd v0-inline
go mod tidy
go run .
# 另开终端
curl -X POST http://127.0.0.1:8080/api/auth/login -H "Content-Type: application/json" -d "{\"username\":\"admin\",\"password\":\"admin123\"}"
```

其余版本把目录和端口换成上表即可。失败示例可试错误密码，应返回 401。

v4 额外提供：

- `GET /api/projects`
- `GET /api/projects/1/tasks`（内部会先问 ProjectService「能不能看」）

v5 额外提供（需先登录拿 Cookie）：

- `GET /api/auth/me`
- `GET /api/projects`（要登录）
- `GET /api/admin/users`（要 admin 角色）

详见 [学习笔记-Repository与Middleware.md](./学习笔记-Repository与Middleware.md) 第 8 节 curl 示例。

v6 接口与 v5 相同，只把端口换成 `8086`；重点对比 `v5/main.go` 与 `v6/wire_gen.go`。

---

## 演变故事

### v0：能跑就行

```go
r.POST("/api/auth/login", func(c *gin.Context) {
    // 读 JSON、查用户、比密码、写响应……全在这里
})
```

优点：最短路径。  
缺点：路由文件很快变成「业务垃圾桶」，难测、难复用。

### v1：抽出函数

```go
var users = map[string]string{...} // 包级「全局 db」

func Login(c *gin.Context) { ... }

r.POST("/api/auth/login", Login)
```

优点：路由表干净一点。  
缺点：依赖藏在全局变量里，单测要改全局，多人协作易踩脚。

### v2：Handler 持有 db

```go
type AuthHandler struct{ db *Store }
func NewAuthHandler(db *Store) *AuthHandler { ... }

h := NewAuthHandler(store)
r.POST("/api/auth/login", h.Login)
```

优点：构造时注入，不再靠全局。  
缺点：HTTP 解析和业务规则仍挤在 Handler 里。

### v3：再拆 Service

```go
authSvc := NewAuthService(store)
authH   := NewAuthHandler(authSvc)
r.POST("/api/auth/login", authH.Login)
```

方向固定：**资源 → Service（业务）→ Handler（HTTP）→ 路由**。  
Handler 只碰 `gin.Context`；Service 不 import gin。

### v4：多个模块并排

```go
authSvc    := NewAuthService(store)
projectSvc := NewProjectService(store)
taskSvc    := NewTaskService(store, projectSvc) // 多注入一个依赖

authH    := NewAuthHandler(authSvc)
projectH := NewProjectHandler(projectSvc)
taskH    := NewTaskHandler(taskSvc)
```

先造「脑子」，再造「嘴巴」，最后挂 URL。  
`TaskService` 需要项目可见性判断，所以复用已有的 `projectSvc`，而不是再抄一套。

### v5：Repository + 门禁 + Session

```go
userRepo := NewUserRepository(db)
sessionRepo := NewSessionRepository(db)
authSvc := NewAuthService(userRepo, sessionRepo)

r := NewRouter(Dependencies{Auth: authSvc, Projects: projectSvc})
// router 内：RequireAuth / RequireRole + Handler
```

相对 v4 多学：

- 数据访问按领域拆 Repository  
- 登录写 Session Cookie，后续请求可校验  
- Middleware 做登录态与角色门禁  
- `AppError` 统一错误映射  

细节见 [`v5-repo-mw/`](./v5-repo-mw/) 与 [学习笔记-Repository与Middleware.md](./学习笔记-Repository与Middleware.md)。

### v6：Google Wire 生成组装（业务同 v5）

```go
// main.go —— 不再手写一排 New
r := InitializeApp() // 实现见 wire_gen.go
r.Run(":8086")
```

```go
// wire.go（生成用）列出 Provider
wire.Build(newMemDB, NewUserRepository, ..., NewRouter)
```

对照：`v5-repo-mw/main.go`（手写）↔ `v6-wire/wire_gen.go`（生成）。  
说明见笔记 §7.2 / §7.3。

---

## 各版概念对照（本目录内）

| 概念 | 出现版本 | 本目录里是什么 |
|------|----------|----------------|
| 路由回调堆逻辑 | v0 | `main` 里匿名函数 |
| 包级全局依赖 | v1 | `var users = map...` |
| 构造注入 | v2 | `NewAuthHandler(store)` |
| Service 层 | v3+ | 业务不碰 gin |
| 多模块 DI | v4 | 多个 Service/Handler 并排 |
| Repository | v5 | `UserRepository` 等 |
| Session Cookie | v5 | `sid` + 内存 sessions |
| Middleware | v5 | `RequireAuth` / `RequireRole` |
| 统一业务错误 | v5 | `AppError` + `HandleError` |
| Wire 代码生成 | v6 | `wire.go` + `wire_gen.go` |

学会 v5 / v6 后，再看「handler → service → repository + middleware」或 Wire 组装的 Gin 项目，骨架都认得出来。

---

## 口诀

1. **v0→v1**：把代码挪出路由回调  
2. **v1→v2**：用构造函数注入依赖，干掉全局  
3. **v2→v3**：业务进 Service，HTTP 留 Handler  
4. **v3→v4**：模块变多时统一组装；服务之间也可以互相注入  
5. **v4→v5**：Store 拆成 Repository；补上 Session、Middleware、统一错误  
6. **v5→v6**：组装改由 Wire 生成；分层不变  
