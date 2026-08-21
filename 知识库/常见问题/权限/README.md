# Go 后端常见权限问题完整整理

独立专题。每个问题按同一结构展开：

1. 问题说明  
2. 典型后果  
3. 处理方式  
4. 示范代码（注释清晰）  
5. 错误写法对比  
6. 自测要点  

约定：

- **认证（Authentication）**：你是谁 → 失败 `401`
- **鉴权（Authorization）**：你能做什么 → 失败 `403`
- 前端藏按钮只改善体验；**真正权限以服务端为准**

---

## 0. 总览表

| # | 问题 | 一句话处理 |
|---|------|------------|
| 1 | 未做认证 / 认证不完整 | JWT/Session 中间件，失败返回 `401` |
| 2 | 只登录不鉴权、垂直越权 | `RequireRole` / 权限点，失败返回 `403` |
| 3 | 水平越权（IDOR） | 资源归属或成员关系 `canAccess` |
| 4 | 列表 / 搜索未收数据范围 | 查询强制带 `user_id` / 成员过滤 |
| 5 | 信任客户端传的身份 | 身份只从 Context；创建强制当前用户为 Owner |
| 6 | 只拦写、不拦读 | 读 / 写 / 导出 / 列表同一套鉴权 |
| 7 | JWT 配置不严 | 强 Secret、锁算法、设过期 |
| 8 | 角色变更后旧 token 仍有效 | 短过期，或关键操作回源查库角色 |
| 9 | 错误码泄露资源是否存在 | 无权限与不存在对外统一策略 |
| 10 | 权限只写在前端 | API 强制校验；用 curl 测接口 |

公共骨架（后面代码都基于此）见 [附录 A](#附录-a公共骨架完整注释版)。

---

## 1. 问题：未做认证 / 认证不完整

### 1.1 问题说明

接口完全公开，或只检查「有没有 Authorization 头」，不校验签名、不校验过期。任何人都能冒充用户。

### 1.2 典型后果

- 未登录访问个人数据接口成功
- 伪造 / 篡改 token 仍能通过
- 过期 token 继续使用

### 1.3 处理方式

| 步骤 | 做什么 |
|------|--------|
| 1 | 登录成功后签发 JWT（含 `userID`、`role`、`exp`） |
| 2 | 受保护路由挂认证中间件 |
| 3 | 校验 `Bearer` 前缀、签名、算法、过期 |
| 4 | 把 `userID` / `role` 写入 `gin.Context` |
| 5 | 失败统一 `401`，并 `Abort` |

### 1.4 示范代码

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuth 认证中间件：只回答「你是谁」，不回答「你能不能做某事」。
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) 取 Authorization 头，格式必须是：Bearer <token>
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "unauthorized"})
			return
		}

		// 2) 去掉前缀，得到纯 token 字符串
		tokenStr := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))

		// 3) 验签 + 校验过期 + 锁定算法（实现见附录 A）
		claims, err := auth.ParseToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "unauthorized"})
			return
		}

		// 4) 身份注入 Context，后续 Handler/Service 只从这里取
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)

		// 5) 放行到下一个处理器
		c.Next()
	}
}

// UserID 读取当前登录用户 ID；未登录或类型不对时返回 0。
func UserID(c *gin.Context) uint {
	v, _ := c.Get("userID")
	id, _ := v.(uint)
	return id
}

// Role 读取当前登录用户角色。
func Role(c *gin.Context) string {
	v, _ := c.Get("role")
	r, _ := v.(string)
	return r
}
```

路由挂载：

```go
r := gin.Default()

// 登录、注册：不挂 JWTAuth
r.POST("/api/login", loginHandler)
r.POST("/api/register", registerHandler)

// 业务接口：必须先认证
api := r.Group("/api", middleware.JWTAuth())
{
	api.GET("/me", meHandler)
	api.GET("/notes", listNotesHandler)
}
```

### 1.5 错误写法对比

```go
// ❌ 只看有没有头，不验签
if c.GetHeader("Authorization") == "" {
	c.AbortWithStatus(401)
	return
}
c.Next()

// ❌ 自己 decode payload 却不验签（可伪造）
parts := strings.Split(token, ".")
json.Unmarshal(base64Decode(parts[1]), &claims) // 危险

// ✅ 使用官方库 ParseWithClaims，并校验 Method + Valid
```

### 1.6 自测要点

-  不带 token → `401`
-  乱填 token → `401`
-  过期 token → `401`
-  合法 token → 能进入接口，且 `UserID` 正确

---

## 2. 问题：只登录不鉴权 / 垂直越权

### 2.1 问题说明

用户已经登录，就能调用管理员接口（如删除用户、改全局配置）。  
「垂直越权」= 低权限角色执行了高权限角色的操作。

核心对照：

| 概念 | 问的是什么 | 失败码 |
|------|------------|--------|
| 认证（Authentication） | 你是谁？有没有合法登录态 | `401` |
| 鉴权（Authorization） | 你能不能做这件事？ | `403` |

「只登录不鉴权」= 中间件只验了 JWT，管理接口没再问「你是不是 admin」。  
「垂直越权」= **角色层级**上越权：`user` 干了 `admin` 的活。  
对比「水平越权」（问题 3）：同级用户 A 操作 B 的资源 id（IDOR）。

典型攻击路径：

1. 用普通账号登录，拿到 token。
2. 直接调 `/api/admin/users`、改配置、批删。
3. 或改 body 里 `"role":"admin"` 骗后端（若角色从请求体读）。

前端藏按钮挡不住；`curl` / Postman 一样能打。真正权限以服务端为准。

### 2.2 典型后果

- 普通用户访问 `/api/admin/users` 成功
- 会员专属功能被免费用户调用
- 只要知道路径就能做管理操作
- Body / Header 伪造角色后被当成管理员

### 2.3 处理方式总览

最小落地步骤（任何规模都先做到）：

| 步骤 | 做什么 |
|------|--------|
| 1 | 管理类路由加 `RequireRole("admin")`（第一道门） |
| 2 | Service 内再判一次角色（防漏挂中间件） |
| 3 | 角色来自 Context / 数据库，绝不来自请求体 |
| 4 | 无权限返回 `403`（已登录但不够格） |

角色少用角色字符串；权限点很多时再用权限码（如 `user:write`）。

### 2.4 处理方法对照

从简单到复杂，按系统复杂度选，不要一上来上 Casbin。

| 方法 | 做法 | 适用 |
|------|------|------|
| A. 角色门闸 | 路由组挂 `RequireRole("admin")`，整组管理接口一刀切 | 角色 ≤ 3～5（如 `admin` / `user` / `vip`） |
| B. 权限码 | 用户 → 角色 → 权限码；接口挂 `RequirePerm("user:write")` | 菜单 / 按钮多、权限点常改 |
| C. 双层校验 | 中间件拦路由组 + Service 再判 | **几乎总该做**，防漏挂中间件、防内部调用绕过路由 |
| D. 高危回源 | 删用户、改角色、改全局配置时以 DB 当前角色为准 | 与问题 8 组合；不只信 JWT 里缓存的 `role` |
| E. 能力拆接口 | `/api/admin/*` 整组 admin；会员能力单独挂角色 / 权限码 | 避免「谁都能进」的接口里靠 body 的 `isAdmin` 分支 |

方法 A 请求链路：

```text
请求 → JWTAuth（是谁）→ RequireRole（够不够格）→ Handler → Service
```

方法 B 数据关系：

```text
用户 → 角色 → 权限码集合（如 user:write, config:read）
接口 → RequirePerm("user:write")
```

再大一点可用 Casbin：`Enforce(sub, obj, act)`（见第 13 节升级表）。

不要做的事：

- 只靠前端 `v-if="role==='admin'"`。
- 从 JSON / Query / Header 读客户端传的 `role` / `isAdmin`。
- 只拦写、不拦读（管理列表、导出同样要鉴权）。
- 认为「路径不公开就安全」。

### 2.5 数据库体量大时怎么办

先分清：**垂直鉴权查的是「人的能力」**，不是扫全库业务表。

| 数据 | 体量通常 | 鉴权时怎么用 |
|------|----------|--------------|
| 用户角色 / 权限码 | 相对很小 | 每次请求要判，但应 O(1) / 缓存 |
| 订单 / 笔记 / 日志等业务表 | 可能很大 | **垂直鉴权不必扫这些表** |

业务表再大，也不该为了「是不是 admin」去扫全表。数据库大 ≠ 垂直鉴权要变复杂；变复杂的是「权限结果怎么缓存、怎么失效」。

推荐策略：

| 策略 | 做法 | 说明 |
|------|------|------|
| 登录算好权限 | JWT / Session 带 `role`，或 Redis 存 `userID → []perm` | 中间件用 `map` O(1) 判断，避免每次 JOIN 三张权限表 |
| 权限表可索引 | `user_roles(user_id)`、`role_permissions(role_id)` 建索引 | 按 `user_id` 定点查，拼成 `map[string]bool`，禁止全表扫权限 |
| 高危才回源 | 普通接口信 JWT / Redis；删用户、改权限再查 DB | 大部分流量不打库，关键路径才打库（见问题 8） |
| 管理列表分页 | admin 的 `ListUsers` 等必须分页、限 `page_size` | 修好垂直越权后，仍可能因一次拉全库拖垮性能 |
| 权限点很多 | 请求入口加载一次权限集写入 Context；或多实例用 Redis | 避免每个接口各查一遍；变更时清缓存 / `token_version` |

规模对照（一句话选型）：

| 规模 | 建议 |
|------|------|
| 小系统 | JWT 带 `role` + `RequireRole` + Service 再判 |
| 中系统 | Redis 缓存权限码 + `RequirePerm` + 高危回源 |
| 大系统 / 权限常变 | RBAC 表或 Casbin + 缓存失效策略 + 管理接口分页（展开见 2.6） |

业务表用 `WHERE` / 分页解决数据范围（偏水平 / 列表问题）；角色 / 权限用索引 + 缓存解决垂直门闸。

### 2.6 大系统三件套展开

上表「大系统」那一行是三件事叠在一起，不是一个黑盒库名。拆开如下：

```text
RBAC 表 或 Casbin   →  权限规则怎么存、怎么判
+ 缓存失效策略       →  判得快，但改权限后不能继续用旧结果
+ 管理接口分页       →  有资格进管理端后，别一次把库拖垮
```

前两项二选一（或以后再迁），后两项几乎总要配。

#### 2.6.1 RBAC 表

RBAC = Role-Based Access Control（基于角色的访问控制）。  
不直接给每个用户贴一堆权限，而是：

```text
用户 ──(多对多)──► 角色 ──(多对多)──► 权限码
张三                 admin              user:write
李四                 editor             user:read
                                        config:read
```

典型表：

| 表 | 存什么 |
|----|--------|
| `users` | 用户 |
| `roles` | `admin` / `editor` / `viewer` |
| `permissions` | `user:write`、`config:read` 等 |
| `user_roles` | 谁有哪些角色 |
| `role_permissions` | 角色有哪些权限 |

接口挂权限码，而不是写死 `"admin"`：

```text
DELETE /api/users/:id  →  RequirePerm("user:write")
```

判权流程：用当前 `user_id` 查角色 → 再查这些角色的权限码 → 看集合里有没有目标权限。  
适用：权限点多、产品和运营经常加「能不能看某某菜单 / 能不能导出」，用代码里 `if role == "admin"` 维护不动时。

#### 2.6.2 Casbin

Casbin 是通用鉴权库：把规则写成策略，运行时 `Enforce(主体, 对象, 动作)`。

```text
Enforce("alice", "/api/users", "DELETE")  → true / false
Enforce("alice", "user", "write")         → true / false
```

策略可放文件、DB 或内存，常见一行示意：

```text
p, admin, /api/admin/*, *
p, editor, /api/notes, read
g, alice, admin
```

与自建 RBAC 表对照：

| 项 | 自建 RBAC 表 | Casbin |
|----|--------------|--------|
| 本质 | 自己写 SQL + 拼 `map` 判权 | 库做匹配 / 角色继承 / 多模型 |
| 模型 | 经典「用户-角色-权限」 | RBAC、ABAC、RESTful 路径等可配 |
| 成本 | 简单直观；规则复杂后代码膨胀 | 多学一层模型与策略语法 |
| 选型 | 权限结构稳定、偏菜单按钮 | 规则多、要路径级 / 属性级、想少写判权代码 |

「RBAC 表 **或** Casbin」= **二选一当权限引擎**，不是两个都要上。小团队往往先 RBAC 表；规则很杂再上 Casbin（也见第 13 节升级表）。

#### 2.6.3 缓存失效策略

权限若每次请求都 JOIN 多张表，QPS 一高库先炸，于是：

```text
登录 / 首次请求 → 算出「张三有哪些 perm」→ 放进 Redis / 本地内存 / JWT
之后请求       → 直接看缓存，O(1) 判断
```

若库里已把张三从 `admin` 降成 `user`，缓存仍是旧的，就会垂直越权一段时间。  
**有缓存就必须回答「降权后多久生效」**；没有失效策略的缓存等于定时炸弹。

| 策略 | 做法 | 延迟 |
|------|------|------|
| 主动删除 | 改角色时 `DEL perm:user:123` | 立刻（同集群内） |
| 版本号 | 用户表 `perm_version`，JWT / 缓存带版本，改权限就 +1 | 立刻可感知 |
| 短过期 | Access Token 15～30 分钟，或 Redis TTL | 最多等一个 TTL |
| 高危回源 | 删用户等操作不信缓存，直接查 DB | 该接口 0 延迟 |

常见组合：日常接口用 Redis 权限集 + 改权限时删 key；高危接口再查一次 DB 角色（问题 8）；需要「立刻踢下线」时用 `token_version` +1，旧 JWT 整段作废。

#### 2.6.4 管理接口分页

这和「有没有资格」是另一件事。  
修好垂直越权后，只有 admin 能调 `GET /api/admin/users`；若实现是一次 `SELECT * FROM users`，有权限的人一点开，DB / 内存 / 网关照样挂。

管理端必须：

- 使用 `page` / `page_size`，且服务端强制上限（例如最大 100）。
- 按索引字段筛选（创建时间、状态），禁止无条件全表扫。
- 列表只返回必要列，详情再按 id 查。

对照：

| 问题 | 问的是 |
|------|--------|
| 垂直鉴权 | 你能不能进这扇门？ |
| 管理接口分页 | 进门后一次只搬一箱货，别把仓库搬空 |

#### 2.6.5 怎么串起来

```text
1. 用 RBAC 表（或 Casbin）存「谁能干什么」
2. 把算好的权限放进 Redis，改权限就失效 / 升版本
3. 管理列表一律分页，防止有权限的人把库打穿
```

小系统不必一次上齐：角色就两三个时，JWT 里带 `role` + `RequireRole` 就够；等权限点变多、要频繁改、或用户量上去，再换成本节组合。

### 2.7 示范代码

```go
package middleware

import "net/http"

// RequireRole 角色门闸：当前用户角色必须在允许列表中。
// 用法：admin := api.Group("/admin", middleware.RequireRole("admin"))
func RequireRole(roles ...string) gin.HandlerFunc {
	// 允许的角色集合，O(1) 查找
	allow := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allow[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role := Role(c) // 来自 JWTAuth 注入，不是客户端传的
		if _, ok := allow[role]; !ok {
			// 已登录但角色不够 → 403
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"msg": "forbidden"})
			return
		}
		c.Next()
	}
}
```

```go
package service

// ListUsers 仅管理员可列出全部用户。
// actorRole 必须由 Handler 从 middleware.Role(c) 传入。
func (s *UserService) ListUsers(actorRole string) ([]User, error) {
	// Service 再判一次：即使有人忘记挂 RequireRole，这里仍拦住
	if actorRole != "admin" {
		return nil, ErrForbidden
	}
	// 体量大时改为分页接口，禁止一次 ListAll 拉全库（见 2.5、2.6.4）
	return s.repo.ListAll()
}
```

```go
// Handler：把 Context 身份传给 Service，不把 body 里的 role 传进去
func listUsersHandler(c *gin.Context) {
	users, err := userService.ListUsers(middleware.Role(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, users)
}
```

权限点写法（可选进阶）：

```go
// RequirePerm 校验是否拥有某个权限码，例如 "user:write"。
func RequirePerm(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// perms 可在登录时写入 Context，或这里按 userID 查库
		perms := loadPerms(c)
		if !perms[perm] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"msg": "forbidden"})
			return
		}
		c.Next()
	}
}
```

路由示例：

```go
api := r.Group("/api", middleware.JWTAuth())
{
	// 垂直权限：整组接口仅 admin
	admin := api.Group("/admin", middleware.RequireRole("admin"))
	{
		admin.GET("/users", listUsersHandler)
		admin.DELETE("/users/:id", deleteUserHandler)
	}
}
```

### 2.8 错误写法对比

```go
// ❌ 只要登录就能进管理接口
api.GET("/admin/users", listUsersHandler) // 只挂了 JWTAuth

// ❌ 从 JSON body 读角色
var req struct{ Role string }
c.BindJSON(&req)
if req.Role == "admin" { /* 客户端可伪造 */ }

// ✅ 路由门闸 + Service 再判 + 角色来自 Context
```

### 2.9 自测要点

-  `user` 角色调 `/api/admin/users` → `403`
-  `admin` 角色同接口 → `200`
-  Body 里写 `"role":"admin"` 不能提权
-  漏挂 `RequireRole` 时，Service 仍能拦住
-  权限变更后，高危接口不以过期 JWT 角色为准（见问题 8）
-  改角色 / 权限后，缓存或 `token_version` 按设计失效（见 2.6.3）
-  管理列表强制分页且 `page_size` 有服务端上限（见 2.6.4）

---

## 3. 问题：水平越权（IDOR）

### 3.1 问题说明

接口只验证「已登录」，不验证「这个资源是不是你的」。  
用户把 `/api/notes/12` 改成 `/api/notes/13`，就能看或改别人的数据。  
这叫 **IDOR**（Insecure Direct Object Reference）。

### 3.2 典型后果

- 越权读他人笔记 / 订单 / 文档
- 越权改、删他人资源
- 批量遍历 id 拖库

### 3.3 处理方式

| 步骤 | 做什么 |
|------|--------|
| 1 | 先按 id 取出资源 |
| 2 | 用 `canAccess` / `canEdit` 判断归属或成员关系 |
| 3 | 管理员可放行；普通用户必须是 Owner 或成员 |
| 4 | 读、改、删都要判，不能只判写 |
| 5 | Handler 只传当前用户身份，不传客户端伪造的 owner |

### 3.4 示范代码

```go
package service

// Note 业务资源：UserID 表示拥有者。
type Note struct {
	ID     uint
	UserID uint // 拥有者，创建时由服务端写入
	Title  string
	Body   string
}

// canAccess 判断 userID 是否有权访问该笔记。
// 规则：管理员可以；否则必须是拥有者。
func (s *NoteService) canAccess(note *Note, userID uint, role string) bool {
	if role == "admin" {
		return true
	}
	return note.UserID == userID
}

// Get 按 id 读取笔记；包含水平越权校验。
func (s *NoteService) Get(userID uint, role string, id uint) (*Note, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}

	// 关键：登录不够，还必须能访问这条资源
	if !s.canAccess(note, userID, role) {
		return nil, ErrForbidden
	}
	return note, nil
}

// Update 修改笔记；同样先做归属校验。
func (s *NoteService) Update(userID uint, role string, id uint, title, body string) error {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return ErrNotFound
	}
	if !s.canAccess(note, userID, role) {
		return ErrForbidden
	}

	note.Title = title
	note.Body = body
	return s.repo.Save(note)
}

// Delete 删除前复用 Get 的鉴权逻辑，避免漏判。
func (s *NoteService) Delete(userID uint, role string, id uint) error {
	note, err := s.Get(userID, role, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(note.ID)
}
```

Handler：

```go
// getNoteHandler 只从 Context 取当前用户，再交给 Service 鉴权。
func getNoteHandler(c *gin.Context) {
	id := parseUintParam(c, "id")

	note, err := noteService.Get(
		middleware.UserID(c), // 当前用户
		middleware.Role(c),   // 当前角色
		id,                   // 要访问的资源
	)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, note)
}
```

团队资源（成员 / Owner）扩展：

```go
// canView：管理员或项目成员可读。
func (s *ProjectService) canView(projectID, userID uint, role string) bool {
	if role == "admin" {
		return true
	}
	// IsMember 查 project_members 表
	return s.repo.IsMember(projectID, userID)
}

// canManage：管理员或项目负责人可改成员 / 归档。
func (s *ProjectService) canManage(projectID, userID uint, role string) bool {
	if role == "admin" {
		return true
	}
	p, err := s.repo.FindByID(projectID)
	if err != nil {
		return false
	}
	return p.OwnerID == userID
}
```

### 3.5 错误写法对比

```go
// ❌ 只查 id，不看归属
func (s *NoteService) Get(id uint) (*Note, error) {
	return s.repo.FindByID(id)
}

// ❌ 用客户端传的 userId 判断
func (s *NoteService) Get(reqUserID, id uint) (*Note, error) {
	note, _ := s.repo.FindByID(id)
	if note.UserID == reqUserID { // reqUserID 可伪造
		return note, nil
	}
	return nil, ErrForbidden
}

// ✅ 用中间件注入的 userID + canAccess
```

### 3.6 自测要点

-  用户 A 读自己的 note → `200`
-  用户 A 读 B 的 note id → `403` 或统一 `404`
-  用户 A 改 / 删 B 的 note → 失败
-  admin 可读他人 note（若业务允许）

---

## 4. 问题：列表 / 搜索未收数据范围

### 4.1 问题说明

详情接口做了归属校验，但列表、搜索、导出直接查全表。  
普通用户一次请求就拿到别人的数据。

### 4.2 典型后果

- `GET /api/notes` 返回全站笔记
- 搜索关键词扫出他人标题 / 内容
- 分页遍历即可拖库

### 4.3 处理方式

| 步骤 | 做什么 |
|------|--------|
| 1 | 列表在 SQL/ORM 层加范围条件 |
| 2 | 普通用户：`WHERE user_id = ?` 或 `JOIN` 成员表 |
| 3 | 管理员：可查全部（仍要分页） |
| 4 | 搜索条件与范围条件同时存在 |
| 5 | 禁止先 `Find` 全量再在内存里过滤（易漏、易炸内存） |

### 4.4 示范代码

```go
package service

// List 按角色返回可见列表。
func (s *NoteService) List(userID uint, role string, page, size int) ([]Note, int64, error) {
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20 // 限制页大小，降低拖库效率
	}

	// 管理员看全部；普通用户强制只看自己的
	if role == "admin" {
		return s.repo.ListAll(page, size)
	}
	return s.repo.ListByUserID(userID, page, size)
}

// Search 搜索时也必须带上数据范围。
func (s *NoteService) Search(userID uint, role, keyword string, page, size int) ([]Note, int64, error) {
	if role == "admin" {
		return s.repo.SearchAll(keyword, page, size)
	}
	return s.repo.SearchByUserID(userID, keyword, page, size)
}
```

```go
package repo

// ListByUserID 普通用户列表：范围写在 WHERE 里。
func (r *NoteRepo) ListByUserID(userID uint, page, size int) ([]Note, int64, error) {
	var (
		list  []Note
		total int64
	)
	q := r.db.Model(&Note{}).Where("user_id = ?", userID) // 强制范围
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// SearchByUserID 范围 + 关键词同时生效；参数用占位符防注入。
func (r *NoteRepo) SearchByUserID(userID uint, keyword string, page, size int) ([]Note, int64, error) {
	var (
		list  []Note
		total int64
	)
	// ✅ user_id 范围 与 title 条件同时存在
	q := r.db.Model(&Note{}).
		Where("user_id = ?", userID).
		Where("title LIKE ?", "%"+keyword+"%")

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}
```

项目成员场景的列表范围：

```go
// 只查「我是成员」的项目
func (r *ProjectRepo) ListForMember(userID uint, page, size int) ([]Project, error) {
	var list []Project
	err := r.db.Table("projects AS p").
		Joins("JOIN project_members m ON m.project_id = p.id").
		Where("m.user_id = ?", userID).
		Offset((page - 1) * size).Limit(size).
		Find(&list).Error
	return list, err
}
```

### 4.5 错误写法对比

```go
// ❌ 全表查出再在内存过滤（数据量大必出问题，也容易漏分支）
all, _ := r.db.Find(&notes).Error
for _, n := range all {
	if n.UserID == userID {
		out = append(out, n)
	}
}

// ❌ 搜索拼 SQL，且没有 user_id（注入 + 越权）
r.db.Raw("SELECT * FROM notes WHERE title LIKE '%" + keyword + "%'")

// ✅ WHERE user_id = ? 且 keyword 用占位符
```

### 4.6 自测要点

-  普通用户列表只有自己的数据
-  搜索他人专属关键词，搜不到他人私有数据
-  改 pageSize 很大也不能看到他人数据
-  admin 列表行为符合产品设计

---

## 5. 问题：信任客户端传的身份

### 5.1 问题说明

把请求体 / Query 里的 `userId`、`role`、`isAdmin` 当作授权依据，或直接写成资源的 Owner。

### 5.2 典型后果

```json
{ "userId": 1, "role": "admin", "title": "hack" }
```

- 伪造管理员角色
- 把数据写到别人名下
- 冒充他人操作

### 5.3 处理方式

| 规则 | 说明 |
|------|------|
| 操作者身份 | 只信中间件注入的 `userID` / `role` |
| 创建资源 | `OwnerID = 当前用户`，忽略 body 里的 userId |
| 更新资源 | 默认禁止改 `user_id`；若允许转移所有权，单独接口 + 严格鉴权 |
| DTO 设计 | 请求结构体不要暴露可伪造的身份字段 |

### 5.4 示范代码

```go
package service

// CreateNoteInput 客户端允许传的字段：不含 userId / role。
type CreateNoteInput struct {
	Title string
	Body  string
}

// Create 创建笔记：拥有者强制为当前登录用户。
func (s *NoteService) Create(actorID uint, in CreateNoteInput) (*Note, error) {
	if actorID == 0 {
		return nil, ErrUnauthorized
	}

	note := &Note{
		UserID: actorID, // ✅ 服务端强制写入，不读客户端 userId
		Title:  in.Title,
		Body:   in.Body,
	}
	if err := s.repo.Insert(note); err != nil {
		return nil, err
	}
	return note, nil
}
```

```go
package handler

// createNoteHandler 示范：绑定业务字段，身份从 Context 取。
func createNoteHandler(c *gin.Context) {
	var req struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		// 注意：即使客户端多传 "userId"/"role"，这里也不接收、不使用
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	note, err := noteService.Create(middleware.UserID(c), service.CreateNoteInput{
		Title: req.Title,
		Body:  req.Body,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, note)
}
```

### 5.5 错误写法对比

```go
// ❌ 危险：用 body 里的身份做授权和归属
type BadReq struct {
	UserID uint   `json:"userId"`
	Role   string `json:"role"`
	Title  string `json:"title"`
}

func badCreate(req BadReq) {
	if req.Role == "admin" {
		// 客户端写 role=admin 就提权了
	}
	note.UserID = req.UserID // 可以写到任意用户名下
}

// ✅ 身份只来自 Context；DTO 不含身份字段
```

### 5.6 自测要点

-  Body 带 `"role":"admin"` 不能提权
-  Body 带他人 `"userId"` 创建后，库中 `user_id` 仍是当前用户
-  更新接口无法把笔记改挂到别人名下（除非专门的转移接口且有鉴权）

---

## 6. 问题：只拦写、不拦读

### 6.1 问题说明

删除、修改做了权限，但详情、下载、导出、打印、预览、分享链接接口没做。  
攻击者「不能改，但能把数据全拿走」。

### 6.2 典型后果

- 不能 `DELETE`，但能 `GET` / `EXPORT` 他人数据
- 批量导出绕过页面上的按钮隐藏
- 附件下载链接无鉴权

### 6.3 处理方式

对同一资源，按动作分别鉴权，但**每个动作都要鉴权**：

| 动作 | 建议函数 |
|------|----------|
| 读详情 / 下载 / 预览 | `canView` |
| 修改 / 删除 | `canEdit`（通常更严） |
| 列表 / 搜索 / 导出 | 与 `canView` 同一数据范围 |
| 管成员 / 归档 | `canManage` |

能复用就复用：例如 `Delete` 先调 `Get`（内部已 `canView/canEdit`）。

### 6.4 示范代码

```go
package service

// canView 可读：管理员或拥有者。
func (s *NoteService) canView(note *Note, userID uint, role string) bool {
	return role == "admin" || note.UserID == userID
}

// canEdit 可写：这里与 canView 相同；有些业务会更严（例如只允许 Owner，成员只读）。
func (s *NoteService) canEdit(note *Note, userID uint, role string) bool {
	return s.canView(note, userID, role)
}

// Get 读详情：走 canView。
func (s *NoteService) Get(userID uint, role string, id uint) (*Note, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if !s.canView(note, userID, role) {
		return nil, ErrForbidden
	}
	return note, nil
}

// Export 导出多个 id：每个资源都要能 View，不能只检查登录。
func (s *NoteService) Export(userID uint, role string, ids []uint) ([]Note, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	list, err := s.repo.FindByIDs(ids)
	if err != nil {
		return nil, err
	}

	out := make([]Note, 0, len(list))
	for i := range list {
		// 逐条鉴权：没有权限的 id 直接跳过或整体失败（按产品选）
		if s.canView(&list[i], userID, role) {
			out = append(out, list[i])
		}
	}

	// 严格策略：只要有一个 id 无权限，就整体拒绝，防止用导出探活
	if len(out) != len(list) {
		return nil, ErrForbidden
	}
	return out, nil
}

// DownloadFile 附件下载同样要鉴权（示例）。
func (s *NoteService) DownloadFile(userID uint, role string, noteID uint) ([]byte, string, error) {
	note, err := s.Get(userID, role, noteID) // 复用读权限
	if err != nil {
		return nil, "", err
	}
	data, filename, err := s.repo.LoadAttachment(note.ID)
	return data, filename, err
}
```

### 6.5 错误写法对比

```go
// ❌ 删有鉴权，导出没有
func (s *NoteService) Delete(...) { /* canEdit */ }
func (s *NoteService) Export(ids []uint) ([]Note, error) {
	return s.repo.FindByIDs(ids) // 任意登录用户可导出任意 id
}

// ✅ 导出 / 下载 / 详情共用 canView
```

### 6.6 自测要点

-  无权限用户 `GET` 详情失败
-  无权限用户导出含他人 id 失败
-  无权限用户下载附件失败
-  有权限用户读 / 导出成功

---

## 7. 问题：JWT 配置不严

### 7.1 问题说明

Secret 太弱或写死在代码里、不校验签名算法、不过期或不检查过期。  
攻击者可伪造任意用户的 token。

### 7.2 典型后果

- 猜到 / 读到 Secret 后签发 `role=admin`
- `alg=none` 或算法混淆绕过验签
- 永久有效 token 泄露后无法失效

### 7.3 处理方式

| 项 | 要求 |
|----|------|
| Secret | 足够长的随机串；只放环境变量；禁止进 Git |
| 算法 | 只接受既定算法（如 HMAC HS256） |
| 过期 | 设置 `exp`；解析时校验 |
| 库 | 用 `golang-jwt/jwt` 等成熟库，不要手写验签 |

### 7.4 示范代码

```go
package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 放进 JWT 的业务字段。
type Claims struct {
	UserID uint   `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// secret 从环境变量读取；未配置直接失败，避免默默使用弱密钥。
func secret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		panic("JWT_SECRET is required")
	}
	return []byte(s)
}

// IssueToken 登录成功后签发 Access Token。
func IssueToken(userID uint, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			// 短过期：降低泄露窗口；需要长会话可用 Refresh Token
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	// 明确使用 HS256
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret())
}

// ParseToken 校验签名、算法、过期时间。
func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		// ✅ 锁定算法：拒绝 alg=none / 把 RS256 当 HS256 等混淆攻击
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret(), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
```

启动检查（可选）：

```go
func MustCheckJWTSecret() {
	s := os.Getenv("JWT_SECRET")
	if s == "" || s == "change-me" || len(s) < 32 {
		log.Fatal("insecure JWT_SECRET")
	}
}
```

### 7.5 错误写法对比

```go
// ❌ 写死弱密钥并提交仓库
var jwtSecret = []byte("123456")

// ❌ 不检查算法
jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
	return secret, nil
})

// ❌ 不设 ExpiresAt
claims := Claims{UserID: id, Role: role}

// ✅ 环境变量 + 锁算法 + exp
```

### 7.6 自测要点

-  未设置 `JWT_SECRET` 时进程拒绝启动（或明确失败）
-  篡改 token payload 后验签失败
-  过期 token 被拒绝
-  仓库检索无生产 Secret

---

## 8. 问题：角色变更后旧 token 仍有效

### 8.1 问题说明

JWT 是自包含凭证：签发时写了 `role=admin`，后来数据库已降为 `user`，在过期前旧 token 仍可能被当成管理员。

### 8.2 典型后果

- 已离职 / 已降权的人继续调管理接口
- 封禁用户在 token 有效期内仍能操作

### 8.3 处理方式

常用三选一或组合：

| 方案 | 做法 | 适用 |
|------|------|------|
| A. 短过期 | Access Token 15m～2h | 大多数系统 |
| B. 关键操作回源 | 删用户、改权限等操作查 DB 角色 | 高危接口必做 |
| C. 版本号 / 黑名单 |用户表 `token_version`，JWT 带版本；改角色就 +1 | 需要「立即失效」时 |

### 8.4 示范代码

方案 B：关键操作以数据库角色为准。

```go
package service

// DeleteUser 高危操作：不以 JWT 内 role 为准，回源查库。
func (s *UserService) DeleteUser(actorID, targetID uint) error {
	// 1) 查操作者当前真实角色
	actor, err := s.repo.FindUser(actorID)
	if err != nil {
		return ErrUnauthorized
	}

	// 2) 以数据库角色裁决
	if actor.Role != "admin" {
		return ErrForbidden
	}

	// 3) 业务保护：例如禁止删除自己
	if actorID == targetID {
		return ErrBadRequest
	}

	// 4) 可选：检查目标是否存在、是否允许删除
	return s.repo.DeleteUser(targetID)
}
```

方案 C：token 版本号（立即失效思路）。

```go
type Claims struct {
	UserID       uint   `json:"uid"`
	Role         string `json:"role"`
	TokenVersion int64  `json:"ver"` // 与用户表字段对齐
	jwt.RegisteredClaims
}

// ParseAndCheck 验签后，再比对 DB 中的 token_version。
func (s *AuthService) ParseAndCheck(tokenStr string) (*Claims, error) {
	claims, err := auth.ParseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	u, err := s.repo.FindUser(claims.UserID)
	if err != nil {
		return nil, ErrUnauthorized
	}
	// 角色变更 / 强制下线时把 DB 中 version +1，旧 token 全部失效
	if u.TokenVersion != claims.TokenVersion {
		return nil, ErrUnauthorized
	}
	// 可选：禁用账号
	if u.Disabled {
		return nil, ErrUnauthorized
	}
	return claims, nil
}

// Downgrade 降权时递增版本，使旧 JWT 立刻失效。
func (s *UserService) Downgrade(userID uint) error {
	return s.repo.BumpTokenVersionAndSetRole(userID, "user")
}
```

### 8.5 错误写法对比

```go
// ❌ 高危操作只信 JWT 角色，且 token 有效期 30 天
func (s *UserService) DeleteUser(actorRole string, targetID uint) error {
	if actorRole != "admin" { // 可能是一周前签发的旧角色
		return ErrForbidden
	}
	return s.repo.DeleteUser(targetID)
}

// ✅ 短过期 + 高危回源查库（必要时尚有 version）
```

### 8.6 自测要点

-  用户从 admin 降为 user 后，旧 token 调普通接口的行为符合预期
-  旧 token 调删除用户等关键接口 → 失败（若做了回源）
-  封禁用户无法继续调用（若检查 Disabled / version）

---

## 9. 问题：错误码泄露资源是否存在

### 9.1 问题说明

- 自己的资源 → `200`
- 别人的资源 → `403`
- 不存在的 id → `404`

攻击者通过 `403` 与 `404` 的差异，扫出哪些 id 真实存在。

### 9.2 典型后果

- 枚举有效订单号 / 文档 id
- 为下一步定向攻击提供目标列表

### 9.3 处理方式

团队选定一种对外策略并统一：

| 策略 | 行为 | 适用 |
|------|------|------|
| A. 统一当不存在（推荐） | 无权限也返回 `404` | 对外 C 端、敏感资源 |
| B. 统一无权限 | 不存在也返回 `403` | 较少用 |
| C. 对内区分、对外统一 | admin 可见真实原因；普通用户统一 `404` | 管理后台 + 用户端并存 |

Service 内部仍可区分 `ErrForbidden` / `ErrNotFound`，在写响应时合并。

### 9.4 示范代码

```go
package apierr

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrBadRequest   = errors.New("bad request")
)

// Write 将领域错误映射为 HTTP 响应。
// 策略 A：Forbidden 与 NotFound 对外都表现为 404，避免泄露存在性。
func Write(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "unauthorized"})
	case errors.Is(err, ErrForbidden), errors.Is(err, ErrNotFound):
		// ✅ 合并：不告诉调用方「到底是没有还是不让看」
		c.JSON(http.StatusNotFound, gin.H{"msg": "not found"})
	case errors.Is(err, ErrBadRequest):
		c.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
	default:
		// 生产环境不要把内部 err 原文返回给客户端
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
	}
}
```

按角色区分（策略 C）：

```go
func WriteByRole(c *gin.Context, role string, err error) {
	if role == "admin" {
		// 管理员需要明确原因，便于排障
		switch {
		case errors.Is(err, ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"msg": "forbidden"})
			return
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"msg": "not found"})
			return
		}
	}
	Write(c, err) // 普通用户走统一策略
}
```

### 9.5 错误写法对比

```go
// ❌ 对外清晰区分，方便扫库
if !canAccess {
	c.JSON(403, gin.H{"msg": "forbidden"}) // 说明这个 id 存在
	return
}

// ✅ 对外统一 404
```

### 9.6 自测要点

-  访问他人资源与访问不存在 id，对普通用户响应一致（若选策略 A）
-  自己的资源仍正常 `200`
-  管理端若需要区分，仅 admin 能看到差异

---

## 10. 问题：权限只写在前端

### 10.1 问题说明

前端用路由守卫、`v-if="role==='admin'"`、隐藏按钮控制权限，后端接口未做对等校验。  
攻击者可直接调 API，完全绕过页面。

### 10.2 典型后果

- 页面上看不到「删除」按钮，但 `DELETE` API 能打通
- 抓包改请求即可提权或越权
- 自动化脚本批量撞接口

### 10.3 处理方式

| 层 | 职责 |
|----|------|
| 前端 | 隐藏无权限入口，改善体验 |
| 后端 | **强制**认证 + 鉴权，这是安全底线 |
| 测试 | 用 curl / Postman / 自动化直接打 API，不要只点 UI |

原则：**前端可骗，后端不可骗。**

### 10.4 示范代码

后端：任何敏感接口都必须经过中间件 + Service 鉴权（见问题 1～3）。  
前端（示意，非安全边界）：

```ts
// 仅用于 UI 展示，不能当作安全措施
function canShowAdminMenu(role: string) {
  return role === "admin";
}
```

自测脚本示例：

```bash
# 1) 普通用户登录拿 token
USER_TOKEN='...'

# 2) 直接打管理接口：期望 403
curl -i -H "Authorization: Bearer $USER_TOKEN" \
  http://localhost:8080/api/admin/users

# 3) 直接打他人资源：期望 404/403
curl -i -H "Authorization: Bearer $USER_TOKEN" \
  http://localhost:8080/api/notes/99999

# 4) 无 token：期望 401
curl -i http://localhost:8080/api/notes
```

Go 侧接口测试示例：

```go
func TestUserCannotListAdminUsers(t *testing.T) {
	r := setupRouter() // 挂载真实中间件与路由
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+userToken) // 普通用户 token
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("got %d, want 403", w.Code)
	}
}
```

### 10.5 错误写法对比

```text
❌ 验收只点页面：没看到按钮就认为「没权限」
✅ 验收直接调 API：无权限必须 401/403/统一 404
```

### 10.6 自测要点

-  隐藏的管理菜单对应 API，普通用户直接调用失败
-  所有写操作、导出、下载均有后端校验
-  CI 或手工清单里包含「绕过 UI 的 API 用例」

---

## 11. 决策流程（串起来怎么用）

```text
来了一个请求
  │
  ├─ 需要登录吗？
  │    否 → 公开接口
  │    是 → JWTAuth（问题 1、7）
  │
  ├─ 整类接口是否仅某角色？
  │    是 → RequireRole / RequirePerm（问题 2）
  │
  ├─ 是否带资源 id？
  │    是 → canView / canEdit / 成员关系（问题 3、6）
  │    否 → 若是列表/搜索 → 强制 WHERE 范围（问题 4）
  │
  ├─ 身份是否只来自 Context？（问题 5）
  ├─ 高危操作是否回源查角色？（问题 8）
  └─ 对外错误码是否统一策略？（问题 9）
```

覆盖日常约 80% 场景的最小集合：

1. JWT 认证中间件  
2. 角色门闸（垂直）  
3. 资源归属 / 成员（水平）  
4. 列表数据范围  

---

## 12. 总自测清单

-  无 token → `401`
-  普通角色调管理接口 → `403`
-  用户 A 操作 B 的资源 id → 失败
-  列表 / 搜索看不到他人私有数据
-  Body 伪造 `role` / `userId` 无效
-  导出 / 下载 / 详情权限一致
-  JWT Secret 安全、算法锁定、有过期
-  降权 / 封禁后关键操作符合预期
-  普通用户对「无权限」与「不存在」响应策略符合设计
-  所有敏感能力都能在不点 UI 的情况下用 API 验证

---

## 13. 何时再升级（超出基础 80%）

| 现状 | 升级方向 |
|------|----------|
| 角色 ≤ 3，规则少 | 保持「角色 + canXxx」 |
| 权限点很多、常改 | 角色-权限表，或 Casbin `Enforce(sub, obj, act)` |
| 多租户 | 所有查询强制 `tenant_id`，与用户权限叠加 |
| 要合规审计 | 高危操作记审计日志（谁、何时、资源、结果） |
| 要立即踢人 | Access + Refresh，或 token 版本号 / 黑名单 |

---

## 附录 A：公共骨架（完整注释版）

```go
package apierr

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized") // → 401
	ErrForbidden    = errors.New("forbidden")    // → 403（对外可映射为 404）
	ErrNotFound     = errors.New("not found")    // → 404
	ErrBadRequest   = errors.New("bad request")  // → 400
)
```

```go
package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func secret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		panic("JWT_SECRET is required")
	}
	return []byte(s)
}

func IssueToken(userID uint, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret())
}

func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret(), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
```

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "unauthorized"})
			return
		}
		claims, err := auth.ParseToken(strings.TrimSpace(strings.TrimPrefix(h, "Bearer ")))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "unauthorized"})
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	allow := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allow[r] = struct{}{}
	}
	return func(c *gin.Context) {
		if _, ok := allow[Role(c)]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"msg": "forbidden"})
			return
		}
		c.Next()
	}
}

func UserID(c *gin.Context) uint {
	v, _ := c.Get("userID")
	id, _ := v.(uint)
	return id
}

func Role(c *gin.Context) string {
	v, _ := c.Get("role")
	r, _ := v.(string)
	return r
}
```

```go
package service

// NoteService 最小可运行业务模板：创建 / 列表 / 详情 / 删除都带权限。
type NoteService struct {
	repo *NoteRepo
}

type Note struct {
	ID     uint
	UserID uint
	Title  string
	Body   string
}

func (s *NoteService) canAccess(n *Note, userID uint, role string) bool {
	return role == "admin" || n.UserID == userID
}

func (s *NoteService) Create(userID uint, title, body string) (*Note, error) {
	if userID == 0 {
		return nil, ErrUnauthorized
	}
	n := &Note{UserID: userID, Title: title, Body: body} // Owner 强制当前用户
	return n, s.repo.Insert(n)
}

func (s *NoteService) List(userID uint, role string) ([]Note, error) {
	if userID == 0 {
		return nil, ErrUnauthorized
	}
	if role == "admin" {
		return s.repo.ListAll()
	}
	return s.repo.ListByUserID(userID) // 列表强制范围
}

func (s *NoteService) Get(userID uint, role string, id uint) (*Note, error) {
	n, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if !s.canAccess(n, userID, role) {
		return nil, ErrForbidden
	}
	return n, nil
}

func (s *NoteService) Delete(userID uint, role string, id uint) error {
	n, err := s.Get(userID, role, id) // 复用鉴权
	if err != nil {
		return err
	}
	return s.repo.Delete(n.ID)
}
```

记忆口诀：

1. **中间件认人，Service 判权**  
2. **角色管大门，归属管房间**  
3. **列表带范围，导出不例外**  
4. **身份不信客户端，高危回源查库**  
5. **前端可藏，后端必拦**  
