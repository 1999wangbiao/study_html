# Gin 分层学习笔记：Handler / Service / 依赖组装

> 配合本目录 `v0-inline` → `v4-multi` 示例阅读（只依赖本文件夹）。  
> 学完 v4 后继续：[`v5-repo-mw/`](./v5-repo-mw/) 与 [学习笔记-Repository与Middleware.md](./学习笔记-Repository与Middleware.md)。

---

## 1. 先记住一张图

```text
浏览器 / 前端
    │  HTTP（JSON、状态码）
    ▼
┌─────────────┐
│   Handler   │  控制层（嘴巴 / 柜台）
│  ≈ Java Controller
└──────┬──────┘
       │  普通参数 / 普通结果 / error
       ▼
┌─────────────┐
│   Service   │  业务层（脑子 / 后厨）
│  ≈ Java @Service
└──────┬──────┘
       │  查数据
       ▼
┌─────────────┐
│ Store / DB  │  数据层（仓库）
│  ≈ Repository / DAO
└─────────────┘
```

**口诀：**

- Handler：接请求 → 调业务 → 写响应  
- Service：具体业务规则（尽量不碰 gin）  
- Store：数据放哪

---

## 2. Handler 是什么？干什么？

**Handler = 控制层**（Go/Gin 里常叫 Handler，Java/Spring 里常叫 Controller）。

它做的事很固定：

1. **入**：从 `*gin.Context` 读参数（JSON、路径 `:id`、Query）  
2. **转**：调用对应的 Service  
3. **出**：按结果选 HTTP 状态码，写 JSON 返回  

它回答的问题是：**「这个 URL 怎么对前端说话？」**

它一般**不负责**：

- 密码怎么校验的细节循环  
- 项目权限规则怎么算  
（这些属于 Service）

---

## 3. Service 是什么？干什么？

**Service = 业务层**，写「这件事业务上对不对、结果是什么」。

例如：

- 用户名密码是否正确  
- 某个项目能不能看  
- 某项目下有哪些任务  

它回答的问题是：**「业务规则是什么？」**

好习惯：

- 方法参数用普通类型（`string`、`int`、自己的 struct）  
- **不要**传 `*gin.Context`，也尽量 **不要**在 Service 里 `c.JSON`  
- 失败就返回 `error`，由 Handler 决定是 401 还是 404

---

## 4. 用 Java 打个比方

| Go 演示 | Java / Spring | 生活比喻 |
|---------|---------------|----------|
| Handler | `@RestController` / Controller | 前台接待 |
| Service | `@Service` | 业务部门 |
| Store | `@Repository` / DAO | 档案室 |

Java 里常见写法：

```java
@RestController
class AuthController {          // ≈ AuthHandler
    AuthService authService;    // 注入业务

    @PostMapping("/api/auth/login")
    Object login(@RequestBody LoginReq req) {
        LoginResult res = authService.login(...);
        return Map.of("token", res.getToken());
    }
}

@Service
class AuthService {             // ≈ AuthService
    // 校验密码等业务，不写 HttpServletResponse
}
```

名字不同，分层思想一样：

> **Controller/Handler 是控制层；Service 是业务层。**

---

## 5. 一次请求怎么走？（以登录为例）

```text
POST /api/auth/login  {"username":"admin","password":"admin123"}
        │
        ▼
 AuthHandler.Login
   ├─ ShouldBindJSON        （控制层：解析参数）
   ├─ authSvc.Login(...)    （交给业务层）
   │        │
   │        ▼
   │  AuthService.Login
   │    ├─ 查 Store 里有没有这个用户
   │    ├─ 比密码
   │    └─ 返回 LoginResult 或 error
   │
   ├─ 若 error → 401 JSON
   └─ 若成功 → 200 JSON（token、username）
```

一句话版：

> **请求 → Handler（入）→ Service（业务）→ Handler（出）→ 响应**

---

## 6. 为啥要拆成两层？

| 好处 | 说明 |
|------|------|
| 职责清晰 | HTTP 归 Handler，规则归 Service |
| 好改 | 改「失败返回 401 还是 403」只动 Handler；改「怎么验密码」只动 Service |
| 好测 | Service 可以不启动 HTTP 就单测 |
| 可复用 | 同一套业务将来可给别的入口用（不只是这个 URL） |

---

## 7. 从 v0 到 v4：分层是怎么长出来的？

本目录每版端口不同，可对照跑（详见 [README.md](./README.md)）。

| 版本 | 目录 | 核心变化 | 还缺什么 |
|------|------|----------|----------|
| v0 | `v0-inline/` | 全写在路由回调里 | 乱、难测 |
| v1 | `v1-handler-fn/` | 抽出函数 | 仍依赖包级全局 |
| v2 | `v2-handler-db/` | Handler 持有 Store（注入） | 业务仍挤在 Handler |
| v3 | `v3-service/` | 拆出 Service | 只有登录一块 |
| v4 | `v4-multi/` | 多 Service/Handler 并排组装 | 登录后无门禁；Store 未拆 |
| v5 | `v5-repo-mw/` | Repository + Session + Middleware | 见下一篇笔记 |

### v3：一条链（最小完整分层）

```text
store → authSvc → authH → 路由
```

### v4：多条线 + Service 互相依赖

```text
store ─┬→ authSvc    → authH
       ├→ projectSvc → projectH
       │       ↑
       └→ taskSvc ──→ taskH
```

`NewTaskService(store, projectSvc)` 的含义：

- 任务业务要判断「项目能不能看」  
- 规则已经在 `ProjectService` 里了  
- **复用**，不要在任务里再抄一份  

---

## 8. main 里那几行 New 是在干嘛？

以 v4 `main.go` 为例：

```go
// 1）先造 Service（脑子）
authSvc    := NewAuthService(store)
projectSvc := NewProjectService(store)
taskSvc    := NewTaskService(store, projectSvc)

// 2）再造 Handler（嘴巴）
authH    := NewAuthHandler(authSvc)
projectH := NewProjectHandler(projectSvc)
taskH    := NewTaskHandler(taskSvc)

// 3）挂路由
api.POST("/auth/login", authH.Login)
```

这叫**手动依赖注入（DI）**：

- 需要什么，在构造时明确传进去  
- 依赖关系从 `NewXxx(...)` 参数就能看出来  
- 顺序重要：被依赖的先创建（先有 `projectSvc`，再有 `taskSvc`）

v5 把「造 Handler + 挂路由」挪到 `router.go`，`main` 只负责 Repo/Service；思想相同，见 [学习笔记-Repository与Middleware.md](./学习笔记-Repository与Middleware.md)。

---

## 9. 本目录各层叫什么？

| 本演示 | 常见别名（其它材料里也可能这么叫） |
|--------|------------------------------------|
| Handler | Controller / 控制层 |
| Service | 业务层 |
| Store（v2～v4） | 简易数据层 |
| Repository（v5） | 按领域拆开的数据访问层 |

读代码时抓住：

1. 路由挂的是哪个 Handler 方法  
2. Handler 调的是哪个 Service  
3. Service 是否还依赖别的 Service / Repository  

---

## 10. 复习自测（合上代码问自己）

1. Handler 和 Service，谁可以出现 `c.JSON`？谁尽量不要？  
2. 为什么 `TaskService` 要注入 `ProjectService`，而不是自己再写一遍「项目在不在」？  
3. v1 的全局 `users` 和 v2 的 `NewAuthHandler(store)`，哪个更好测？为什么？  
4. 用 Java 说：Handler / Service / Store 分别像什么注解/层？  
5. 画出登录请求从进到出的调用链。

参考答案要点：

1. Handler 可以；Service 尽量不要。  
2. 规则复用、改一处生效、避免复制粘贴。  
3. v2 更好测：依赖从构造函数传入，可换成假 Store。  
4. Controller / `@Service` / Repository。  
5. Handler → Service → Store → 回到 Handler → HTTP 响应。

---

## 11. 推荐复习路径

1. 按顺序读代码：`v0` → `v1` → `v2` → `v3` → `v4`（每版注释已按小白写细）  
2. 重点看 v4：`main.go`（组装）+ `handler.go` + `service.go`  
3. 再学 v5：[`v5-repo-mw/`](./v5-repo-mw/) + [学习笔记-Repository与Middleware.md](./学习笔记-Repository与Middleware.md)  
4. 隔几天做一遍第 10 节自测题，以及 v5 笔记里的自测题  

---

## 12. 一句话总纲

> **Handler（控制层）负责 HTTP；Service（业务层）负责规则；Store/Repository 负责数据。**  
> **main/router 负责把它们按依赖顺序组装起来。**  
> **下一步（v5）：Middleware 管门禁，Session 管登录态，AppError 统一错误。**
