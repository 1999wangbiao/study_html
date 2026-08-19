# v6-wire：用 Google Wire 生成依赖组装

> 业务与 [`../v5-repo-mw/`](../v5-repo-mw/) **完全相同**（Session、Middleware、Repository、AppError）。  
> 本版只改一件事：把 v5 `main.go` 里那一排手写 `New...`，改成 **Wire 生成**。  
> 概念总述见上级笔记：[学习笔记-Repository与Middleware.md](../学习笔记-Repository与Middleware.md) §7.2。

---

## 1. 本版要说明白的一件事

| | v5 | v6（本目录） |
|--|----|--------------|
| 分层 | Repository → Service → Handler + Middleware | **不变** |
| 接口 / 账号 / Cookie | 相同 | **相同**（仅端口 `8086`） |
| 谁写组装代码 | 你在 `main.go` 手写 | Wire 生成到 `wire_gen.go` |
| `main` 里干什么 | 一排 `New` + `NewRouter` | 只调 `InitializeApp()` |

口诀：

> **Wire 不替代分层，只替代「手写那串 New」。**

---

## 2. 和 v5 的一一对应（先看这张表）

### 2.1 文件级：哪些一样、哪些只在 v6？

| v5 文件 | v6 文件 | 关系 |
|---------|---------|------|
| `repository.go` | `repository.go` | **相同**（内存 DB + 三个仓库） |
| `service.go` | `service.go` | **相同** |
| `handler.go` | `handler.go` | **相同** |
| `middleware.go` | `middleware.go` | **相同** |
| `apperr.go` | `apperr.go` | **相同** |
| `router.go` | `router.go` | **相同**（造 Handler、挂路由） |
| `main.go` 里第 27–43 行那串组装 | `wire_gen.go` 的 `InitializeApp` | **对应**：手写 ↔ 生成 |
| `main.go` 里 `Run(":8085")` | `main.go` 里 `Run(":8086")` | **对应**：只是端口不同 |
| （无） | `wire.go` | **v6 新增**：告诉 Wire「用哪些 Provider」 |
| （无） | `providers.go` | **v6 新增**：把两个 Service 打成 `Dependencies` |

结论：**业务文件不用对照差异；只要对照「组装」那一块。**

### 2.2 组装代码逐行对照（核心）

左边打开 [`../v5-repo-mw/main.go`](../v5-repo-mw/main.go)，右边打开本目录 [`wire_gen.go`](./wire_gen.go)：

| v5 `main.go` | v6 `wire_gen.go` |
|--------------|------------------|
| `db := newMemDB()` | `memDB := newMemDB()` |
| `userRepo := NewUserRepository(db)` | `userRepository := NewUserRepository(memDB)` |
| `sessionRepo := NewSessionRepository(db)` | `sessionRepository := NewSessionRepository(memDB)` |
| `projectRepo := NewProjectRepository(db)` | `projectRepository := NewProjectRepository(memDB)` |
| `authSvc := NewAuthService(userRepo, sessionRepo)` | `authService := NewAuthService(userRepository, sessionRepository)` |
| `projectSvc := NewProjectService(projectRepo)` | `projectService := NewProjectService(projectRepository)` |
| `NewRouter(Dependencies{Auth: authSvc, Projects: projectSvc})` | `provideDependencies(...)` 再 `NewRouter(dependencies)` |
| （写在 `main` 里） | （写在 `InitializeApp` 里，由 `main` 调用） |

v5 里「塞进 `Dependencies{...}`」这一步，在 v6 拆成了 `providers.go` 的 `provideDependencies`，方便 Wire 按类型接线。

### 2.3 `main` 入口对照

**v5：**

```go
func main() {
    db := newMemDB()
    // ... 一排 New ...
    r := NewRouter(Dependencies{...})
    r.Run(":8085")
}
```

**v6：**

```go
func main() {
    r := InitializeApp() // 上面那一排 New 都在 wire_gen.go 里
    r.Run(":8086")
}
```

### 2.4 `wire.go` 对应什么？

`wire.go` **没有** v5 对等文件。它对应的是你脑子里这句订单：

> 「请用这些 `NewXxx`，拼出一个 `*gin.Engine`。」

v5 是你自己按订单施工（写在 `main`）；v6 是把订单写在 `wire.go`，让工具生成施工单（`wire_gen.go`）。

### 2.5 通俗比喻：点奶茶

用「点奶茶」理解三个文件的分工：

| 文件 | 角色 | 干什么 |
|------|------|--------|
| `wire.go` | **点单小票** | 写清：要一杯「全套启动好的服务」，原料用这些配方（`NewXxx`） |
| `wire_gen.go` | **后厨按单做出来的步骤单** | 先打底、再加料、最后装杯——一步步写死了 |
| `main.go` | **顾客开口：我要喝** | 不管怎么调，端来就开始喝（`Listen`） |

对照本项目：

**点单小票（`wire.go`）**——你写要什么、用哪些配方：

```text
我要：一杯 *gin.Engine（能对外服务的 App）
配方：newMemDB、NewUserRepository、NewAuthService、NewRouter …
```

**后厨步骤单（`wire_gen.go`）**——工具按小票展开成具体动作：

```text
1. 先准备杯子底座（newMemDB）
2. 再加用户料、会话料（两个 Repository）
3. 调成 AuthService / ProjectService
4. 装进 Dependencies，交给 NewRouter
5. 端出 Engine
```

**顾客（`main.go`）**：

```text
给我那杯（InitializeApp）
开始喝（Run 监听端口）
```

为啥要拆成三份？

- 改菜单（换配方、多加一种 Repo）→ 改小票 `wire.go`，让后厨重打步骤单（重新跑 `wire`）。  
- 日常喝奶茶 → 只认步骤单 `wire_gen.go`，顾客 `main` 不用知道怎么调。  
- 小票本身不当饮品卖：`wire.go` 带了 `wireinject` 标签，正常 `go run` 不把它当成品编译进去。  

一句话：

> **`wire.go` 是点什么；`wire_gen.go` 是怎么做出来；`main.go` 是端上来开喝。**

---

## 3. 怎么运行

```powershell
cd v6-wire
go mod tidy
go run .
```

- 监听：`http://127.0.0.1:8086`
- 演示账号：
  - `admin` / `admin123`（角色 admin）
  - `student` / `student123`（角色 student）

已提交 `wire_gen.go`，**没有安装 wire 命令也能直接跑**。

登录示例（PowerShell）：

```powershell
curl -c cookies.txt -X POST http://127.0.0.1:8086/api/auth/login `
  -H "Content-Type: application/json" `
  -d "{\"username\":\"student\",\"password\":\"student123\"}"

curl -b cookies.txt http://127.0.0.1:8086/api/projects
curl -b cookies.txt http://127.0.0.1:8086/api/admin/users
# student 访问 admin → 403；换 admin 登录再试 → 200
```

---

## 4. 文件地图（按「实现职责」看）

| 文件 | 平时 `go run` 会编吗？ | 职责 |
|------|------------------------|------|
| `main.go` | 会 | 启动：调用 `InitializeApp()`，监听端口 |
| `wire_gen.go` | 会 | **Wire 生成的组装实现**（真正的一排 New） |
| `wire.go` | **不会**（`wireinject` 标签） | 声明 Injector + Provider 列表，给 `wire` 命令用 |
| `providers.go` | 会 | 胶水：两个 Service 打成 `Dependencies` |
| `repository.go` | 会 | 内存 DB + User/Session/Project 三个仓库 |
| `service.go` | 会 | 登录会话、项目列表（不碰 gin） |
| `handler.go` | 会 | HTTP、Cookie、`HandleError` |
| `middleware.go` | 会 | `RequireAuth` / `RequireRole` |
| `router.go` | 会 | 造 Handler、挂中间件与 URL |
| `apperr.go` | 会 | 业务错误类型 |

阅读顺序建议：

1. `main.go` —— 入口有多短  
2. `wire.go` —— 你声明了什么  
3. `wire_gen.go` —— 生成结果长什么样（对照 v5 `main.go`）  
4. `providers.go` —— 为什么需要这一步胶水  
5. 再按需看 `router` → `middleware` → `handler` → `service` → `repository`（与 v5 相同）

---

## 5. 实现原理：Wire 在本项目里怎么工作？

### 5.1 三个角色

```text
Provider（零件说明书）
  = 已有的构造函数，如 NewUserRepository、NewAuthService

Injector（总装订单）
  = wire.go 里的 InitializeApp
  = 「我最终要 *gin.Engine，请用下面这些 Provider 拼出来」

生成代码（总装流水线结果）
  = wire_gen.go 里的 InitializeApp
  = 普通 Go 函数，无反射、无运行时容器
```

### 5.2 你声明了什么？（`wire.go`）

```go
//go:build wireinject   ← 关键：平时编译跳过本文件

func InitializeApp() *gin.Engine {
    wire.Build(
        newMemDB,
        NewUserRepository,
        NewSessionRepository,
        NewProjectRepository,
        NewAuthService,
        NewProjectService,
        provideDependencies,
        NewRouter,
    )
    return nil // 占位；生成后不会用这个文件
}
```

Wire 根据：

- 每个 Provider 的**参数类型**（需要什么）  
- 每个 Provider 的**返回类型**（能提供什么）  
- Injector 的**返回类型**（最终要什么）  

在编译期画出依赖图，再写成 `wire_gen.go`。

### 5.3 生成结果长什么样？（`wire_gen.go`）

```go
func InitializeApp() *gin.Engine {
    memDB := newMemDB()
    userRepository := NewUserRepository(memDB)
    sessionRepository := NewSessionRepository(memDB)
    authService := NewAuthService(userRepository, sessionRepository)
    projectRepository := NewProjectRepository(memDB)
    projectService := NewProjectService(projectRepository)
    dependencies := provideDependencies(authService, projectService)
    engine := NewRouter(dependencies)
    return engine
}
```

和 v5 手写几乎一一对应：

| v5 `main.go`（手写） | v6 `wire_gen.go`（生成） |
|----------------------|--------------------------|
| `db := newMemDB()` | `memDB := newMemDB()` |
| `NewUserRepository(db)` | 同上 |
| `NewAuthService(userRepo, sessionRepo)` | 同上 |
| `NewRouter(Dependencies{...})` | `provideDependencies` + `NewRouter` |

### 5.4 为什么需要 `providers.go`？

`NewRouter` 的参数是结构体：

```go
func NewRouter(deps Dependencies) *gin.Engine
```

而 Wire 更擅长「按类型」接线：它能造出 `*AuthService`、`*ProjectService`，但**不会自动**把两个指针塞进你自定义的 `Dependencies` 字段里。

所以加一步胶水：

```go
func provideDependencies(auth *AuthService, projects *ProjectService) Dependencies {
    return Dependencies{Auth: auth, Projects: projects}
}
```

Wire 的推理是：

```text
要 Dependencies
  → 调用 provideDependencies
    → 需要 *AuthService、*ProjectService
      → 分别用 NewAuthService / NewProjectService 造
        → 再往下要 Repository、memDB ……
```

### 5.5 构建标签：为什么两个 `InitializeApp` 不冲突？

| 文件 | 标签 | 含义 |
|------|------|------|
| `wire.go` | `wireinject` | **只有**跑 `wire` 生成时才编译 |
| `wire_gen.go` | `!wireinject` | **平时** `go run` / `go build` 编译这个 |

因此：

- 日常运行：只有 `wire_gen.go` 里的 `InitializeApp`  
- 生成时：Wire 读 `wire.go` 的声明，写出/覆盖 `wire_gen.go`  

这是 Wire 的标准用法，不是本项目特例。

### 5.6 启动时实际发生了什么？

```text
go run .
  → 编译 main + wire_gen + 业务文件（不含 wire.go）
  → main 调用 InitializeApp()
       → newMemDB
       → 三个 Repository（共用同一 memDB）
       → AuthService / ProjectService
       → Dependencies
       → NewRouter（内部再 New Handler，挂 Middleware 与路由）
  → r.Run(":8086")
```

一次请求（例如已登录学生访问项目列表）仍是：

```text
HTTP
  → RequireAuth（查 Session）
  → ProjectHandler
  → ProjectService
  → ProjectRepository
  → JSON 响应
```

**Wire 只参与「进程启动时的组装」，不参与每个请求。**

---

## 6. 依赖树（本版拼出来的样子）

```text
                    newMemDB
                   /    |    \
                  /     |     \
     UserRepository  SessionRepository  ProjectRepository
            \           /               /
             \         /               /
              AuthService        ProjectService
                    \               /
                     \             /
                    provideDependencies
                            |
                         NewRouter
                            |
                       *gin.Engine
```

对照 v5：树一样，差别只是「树是人手写进 main，还是 Wire 写进 wire_gen」。

---

## 7. 和业务分层的关系（别搞混）

| 层 | 本目录文件 | Wire 管不管？ |
|----|------------|---------------|
| Repository | `repository.go` | 不管实现，只调用 `NewXxxRepository` |
| Service | `service.go` | 不管业务规则，只调用 `NewXxxService` |
| Handler / Middleware | `handler.go` / `middleware.go` | 不直接生成 Handler；由 `NewRouter` 内部创建 |
| Router | `router.go` | 把 `NewRouter` 当作最后一个 Provider |
| 组装 | `wire.go` + `wire_gen.go` | **Wire 只负责这一层** |

所以学 v6 时：

- 若关心「登录怎么校验、Cookie 怎么写」→ 仍看 v5/v6 的 service/handler（两边一样）  
- 若关心「依赖怎么拼起来」→ 只盯 `wire.go` / `wire_gen.go` / `main.go`

---

## 8. 如何重新生成 `wire_gen.go`（可选）

本目录已带生成结果；只有你改了 Provider 列表或构造函数签名时才需要重跑。

```powershell
go install github.com/google/wire/cmd/wire@latest
cd v6-wire
wire
# 或：go generate
```

常见失败原因：

- `wire.Build` 里漏了某个 Provider → 某类型造不出来  
- 两个 Provider 返回同一类型 → Wire 不知道选哪个（需 `wire.Bind` 等，本演示刻意避开）  
- 改了 `NewXxx` 参数却没重新 `wire` → `wire_gen.go` 过期，编译报错  

---

## 9. 接口一览（与 v5 相同）

| 方法 | 路径 | 门禁 | 说明 |
|------|------|------|------|
| POST | `/api/auth/login` | 无 | Set-Cookie：`sid` |
| POST | `/api/auth/logout` | 无 | 清会话 |
| GET | `/api/auth/me` | 登录 | 当前用户 |
| GET | `/api/projects` | 登录 | 项目列表 |
| GET | `/api/admin/users` | 登录 + admin | 用户名列表 |

---

## 10. 自学检查

合上代码问自己：

1. `wire.go` 为什么日常 `go run` 不会和 `wire_gen.go` 函数重名冲突？  
2. `provideDependencies` 解决的是什么问题？能不能删掉、让 Wire 直接调 `NewRouter`？  
3. 改 `AuthService.Login` 业务逻辑，要不要重新跑 `wire`？  
4. 画出从 `main` 到 `*gin.Engine` 的创建顺序（可对照 §6）。  

参考要点：

1. 构建标签：平时只编 `!wireinject` 的 `wire_gen.go`。  
2. `NewRouter` 要的是 `Dependencies` 结构体，Wire 不会自动填充字段，需要胶水。  
3. 不要；业务变了只影响 service，组装图没变就不用重新生成。  
4. memDB → Repos → Services → Dependencies → Router → Engine。  

---

## 11. 一句话

> **v6 = v5 的业务 + Wire 生成的组装。**  
> **看懂 `wire_gen.go` 等于看懂 v5 的 `main`；看懂 `wire.go` 等于看懂你如何「下订单」让工具代写。**
