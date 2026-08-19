# Gin 进阶：Repository / Session / Middleware / 业务错误

> 配合本目录可运行示例 [`v5-repo-mw/`](./v5-repo-mw/) 阅读。  
> 前置：先读 [学习笔记-Handler与Service.md](./学习笔记-Handler与Service.md)（v0→v4）。  
> **本笔记与示例自成一体**，只依赖本文件夹即可学完、跑通。

---

## 1. 一句话结论

v0→v4 解决：**HTTP 与业务怎么拆、依赖怎么组装**。  
v5 再补四块工程常见能力：

1. **Repository**：数据访问按领域拆开  
2. **Session Cookie**：登录后服务端记住你是谁  
3. **Middleware**：进 Handler 前做登录 / 角色门禁  
4. **AppError**：业务错误与 HTTP 状态码解耦  

```text
请求
  → Middleware（鉴权 / 角色）
    → Handler（读参、写 Cookie / JSON）
      → Service（业务规则）
        → Repository（读写数据）
```

---

## 2. v4 还缺什么？v5 补什么？

| 能力 | v4 | v5（本目录） |
|------|----|--------------|
| 数据层 | 一个 `*Store` 装所有 map | `User` / `Session` / `Project` 多个 Repository |
| 登录产物 | JSON 里的假 `token`，后续接口不校验 | Set-Cookie：`sid`；后续接口查会话 |
| 门禁 | 无 | `RequireAuth`、`RequireRole` |
| 错误 | Handler 里手写 401/404 | `AppError` + `HandleError` 统一映射 |
| 组装文件 | 全在 `main.go` | `main` 组 Repo/Service；`router` 组 Handler 与路由 |

口诀：

> **v4 教会分层；v5 教会门禁与数据访问边界。**

---

## 3. Repository：为什么从 Store 再拆一层？

### 3.1 问题

v4 的 `Store` 像「一间大仓库」，用户、项目、任务全堆一起。模块一多会出现：

- Service 构造参数看不出它碰哪些数据  
- 改「用户怎么查」容易牵动无关代码  
- 单测想换成假数据源时，整库都要换  

### 3.2 做法

按领域拆：

```text
memDB（或真实 *gorm.DB）
  ├─ UserRepository
  ├─ SessionRepository
  └─ ProjectRepository
```

职责：

| 层 | 做 | 不做 |
|----|----|------|
| Repository | 增删改查 | 「角色能不能」「最多几条」 |
| Service | 业务规则 | SQL / 直接摸底层 map 细节（尽量） |
| Handler | HTTP | 业务规则 |

本目录演示仍用内存，是为了**零外部依赖**（不必装 MySQL）。换真实项目时，把 Repository 内部改成 GORM 即可，Service / Handler 形状不变。

---

## 4. Session Cookie：登录之后怎么证明身份？

### 4.1 v4 的缺口

登录返回 `"token": "fake-token-for-admin"`，但 `GET /api/projects` **并不检查**这个 token——谁都能直接调。

### 4.2 v5 的做法

```text
登录成功
  → 服务端 sessions 表（演示是内存 map）写入一条 Session
  → 响应 Set-Cookie: sid=....（HttpOnly）
后续请求
  → 浏览器自动带 Cookie
  → Middleware 用 sid 查 Session → 得到当前用户
```

要点：

- Cookie 里只放 **session id**，不放密码  
- 会话数据在服务端，可失效（TTL / 登出删除）  
- `HttpOnly`：前端 JS 读不到 Cookie，降低被 XSS 偷会话的风险  

演示账号（见 `v5-repo-mw/main.go` 注释）：

| 用户 | 密码 | 角色 |
|------|------|------|
| `admin` | `admin123` | `admin` |
| `student` | `student123` | `student` |

---

## 5. Middleware：门禁挂在路由组上

```go
projects := api.Group("/projects", RequireAuth(authSvc))
admin := api.Group("/admin", RequireAuth(authSvc), RequireRole(RoleAdmin))
```

| 中间件 | 问题 | 失败时 |
|--------|------|--------|
| `RequireAuth` | 有没有有效 `sid`？ | 401 |
| `RequireRole` | 角色对不对？（须排在 Auth 后） | 403 |

一次请求（学生访问管理接口）大致是：

```text
GET /api/admin/users
  → RequireAuth 通过（已登录）
  → RequireRole(admin) 失败 → 403
  → 到不了 AdminHandler
```

Handler 里就不用每个方法都写一遍「有没有登录」。

---

## 6. AppError：Service 不写 HTTP 状态码

Service：

```go
return "", nil, ErrBadCredentials // 带 HTTP=401、code、message
```

Handler：

```go
if err != nil {
    HandleError(c, err) // 统一写成 { "code", "message" }
    return
}
```

好处：

- 改文案 / 状态码只动一处  
- Service 可单测，不必起 HTTP  
- 多个接口共享同一错误语义  

---

## 7. main 与 router 怎么分工？

```text
main.go
  造 memDB → Repository → Service
  调用 NewRouter(Dependencies{...})

router.go
  造 Handler
  挂 Middleware
  挂 URL
```

依赖仍然是**手动注入**，只是文件职责更清晰：

- `main`：进程启动与组装  
- `router`：路由表一眼能看完  

### 7.1 除了手动注入，还有别的方式吗？

有。核心都是解决同一件事：**谁创建依赖、谁传给谁**。常见几类：

| 方式 | 怎么做 | 优点 | 代价 / 注意 |
|------|--------|------|-------------|
| **手动构造注入**（本目录） | `NewXxx(依赖)`，在 `main`/`router` 手写组装 | 零框架、依赖一眼可见、最好排查 | 模块很多时 `main` 变长 |
| **代码生成（如 Google Wire）** | 写 `Provider`，编译期生成组装代码 | 仍是编译期确定依赖，少手写样板 | 多一步 generate；要学 Provider 写法 |
| **运行时容器（如 uber/fx、uber/dig）** | 注册构造函数，容器按类型自动装配 | 模块多、生命周期（启动/停机）好管 | 魔法感更强；出错时要会读容器日志 |
| **全局 / 单例（不推荐当默认）** | 包级 `var db = ...` 或 Service Locator | 写起来短 | 难测、隐藏依赖，正是 v1 要摆脱的 |

和 Java 对照（帮助记忆，不是要求你用 Spring）：

| Go 常见做法 | 类似 Java |
|-------------|-----------|
| 手动 `New` / Wire | 自己 `new` 或编译期组装 |
| fx / dig | 有点像 Spring 容器（运行时按类型注入） |
| 包级全局 | 静态单例 / 到处 `getBean`（易失控） |

本目录为什么坚持手动注入？

1. 先看清依赖树，再谈框架  
2. 中小型 Gin 服务用手写组装通常足够  
3. 以后上 Wire / fx，只是换「谁来写那串 New」，**Handler → Service → Repository 分层不变**

口诀：

> **先会手动 DI；模块真的多到疼了，再考虑 Wire 或 fx。**

### 7.2 Google Wire 是什么？

[Wire](https://github.com/google/wire) 是 Google 开源的 **编译期依赖注入** 工具：你声明「怎么造每个零件」，它**生成**那串手写的 `New...` 组装代码。

和手动 DI、fx 的差别：

| | 手动 DI（本目录） | Wire | fx / dig |
|--|------------------|------|----------|
| 何时拼好依赖 | 你手写 | **编译 / generate 时** | **程序启动时** |
| 运行时有没有容器 | 无 | 无（生成的就是普通 Go 代码） | 有 |
| 依赖画错了何时发现 | 编译失败或你眼睛看出 | generate / 编译失败 | 启动时报错 |

一句话：

> **Wire = 用代码生成代替你手写 main 里那一排 New；跑起来仍然是普通函数调用，不是 Spring 那种反射容器。**

#### 核心概念（三个词）

1. **Provider**：告诉 Wire「这个类型怎么造」，通常就是已有的 `NewXxx`。  
2. **Injector**：你写一个函数签名（要什么结果），函数体里只写 `wire.Build(...)`；Wire 生成真正实现。  
3. **生成文件**：一般是 `wire_gen.go`，提交进仓库也可，CI 里再 `go generate` 也行。

#### 可运行示例：本目录 [`v6-wire/`](./v6-wire/)

业务、接口与 v5 完全相同（端口 `8086`），只把组装交给 Wire。

| 文件 | 作用 |
|------|------|
| `wire.go` | Injector：列出 Provider（带 `wireinject`，平时不编译） |
| `wire_gen.go` | 生成结果：一排 `New`（日常 `go run` 用这个） |
| `providers.go` | 胶水：`*AuthService` + `*ProjectService` → `Dependencies` |
| `main.go` | 只调 `InitializeApp()` 再 `Run` |

建议对照：

1. 打开 `v5-repo-mw/main.go` —— 手写组装  
2. 打开 `v6-wire/wire_gen.go` —— 生成组装（几乎一样）  
3. 打开 `v6-wire/wire.go` —— 你只声明「用哪些 Provider」  

```powershell
cd v6-wire
go run .
# 接口同 v5，端口改 8086
```

已提交 `wire_gen.go`，**不必先装 wire 也能跑**。若要自己重新生成：

```powershell
go install github.com/google/wire/cmd/wire@latest
cd v6-wire
wire
```

#### 什么时候值得用？

- `main` 里 New 已经多到抄错、漏传依赖  
- 同一套 Provider 要组装出「HTTP 服务 / 命令行工具 / 测试」多种入口  
- 团队希望「依赖图由工具检查」，而不是纯靠 Code Review  

什么时候不必上？

- 教学 / 小服务：手写（v5）往往更直观  
- 还没分清 Handler / Service / Repository：先别加工具  

#### 和本目录学习顺序的关系

```text
先会 v5：手写 NewUserRepository → NewAuthService → NewRouter
再跑 v6：Wire 生成同样的组装
不要：还不会分层就先上 Wire / fx
```

官方仓库：<https://github.com/google/wire>

---

## 8. 本版接口一览（可 curl）


端口：`8085`。

| 方法 | 路径 | 门禁 | 说明 |
|------|------|------|------|
| POST | `/api/auth/login` | 无 | 成功 Set-Cookie `sid` |
| POST | `/api/auth/logout` | 无 | 清会话与 Cookie |
| GET | `/api/auth/me` | 登录 | 当前用户 |
| GET | `/api/projects` | 登录 | 项目列表 |
| GET | `/api/admin/users` | 登录 + admin | 用户名列表 |

PowerShell 示例：

```powershell
cd v5-repo-mw
go mod tidy
go run .

# 另开终端：登录（保存 Cookie）
curl -c cookies.txt -X POST http://127.0.0.1:8085/api/auth/login `
  -H "Content-Type: application/json" `
  -d "{\"username\":\"student\",\"password\":\"student123\"}"

# 带 Cookie 访问
curl -b cookies.txt http://127.0.0.1:8085/api/projects
curl -b cookies.txt http://127.0.0.1:8085/api/auth/me

# student 访问 admin → 应 403
curl -b cookies.txt http://127.0.0.1:8085/api/admin/users
```

换 `admin` / `admin123` 登录后再调 `/api/admin/users` 应 200。

---

## 9. 文件地图（只看本目录）

```text
v5-repo-mw/
├── main.go         # 组装 Repo / Service，启动
├── router.go       # Handler + Middleware + 路由
├── middleware.go   # RequireAuth / RequireRole
├── handler.go      # HTTP + HandleError
├── service.go      # 登录会话、项目列表（无 gin）
├── repository.go   # 用户 / 会话 / 项目数据访问
└── apperr.go       # AppError 与预定义错误
```

建议阅读顺序：`main.go` → `router.go` → `middleware.go` → 顺着登录走 `handler` → `service` → `repository`。

---

## 10. 自测题

1. Repository 和 Service，谁不该出现「学生最多 2 条立项」这类规则？  
2. 为什么 Cookie 只放 session id，而不是把用户名密码放进 Cookie？  
3. `RequireRole` 为什么必须排在 `RequireAuth` 后面？  
4. Service 返回 `ErrForbidden` 时，Handler 还要不要自己写 `c.JSON(403, ...)`？  
5. 画出：`student` 登录后访问 `GET /api/admin/users` 的调用链（含中间件）。  

参考要点：

1. 规则在 Service；Repository 只读写。  
2. 会话可服务端失效；密码不应出现在客户端可持有的明文里。  
3. Role 依赖 Context 里的当前用户，用户由 Auth 放入。  
4. 一般不必；走 `HandleError` 即可。  
5. Router → RequireAuth（过）→ RequireRole（403）→ 不进 Handler。  

---

## 11. 一句话总纲

> **Repository 管数据，Service 管规则，Handler 管 HTTP，Middleware 管门禁。**  
> **v5 手动组装，v6 用 Wire 生成组装——换的是写法，不是分层。**
