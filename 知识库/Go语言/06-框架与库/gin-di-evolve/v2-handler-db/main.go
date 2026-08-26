// ============================================================
// v2：去掉全局变量 —— 用「结构体 + 构造函数」把依赖传进去
// ============================================================
//
// 相对 v1 的进步：
//   - 不再用包级 var users
//   - AuthHandler 自己带着 db（Store），需要什么在 NewXxx 时明确传进来
//   - 这就是最朴素的「依赖注入」：外面创建好依赖，塞给使用者
//
// 还没拆开的点：
//   - 业务规则（比密码）和 HTTP（读 JSON / 写状态码）还混在 Handler 里
//   - 下一步 v3 会再拆一层 Service
//
// 演示账号：admin / admin123
// 端口：8082
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Store 模拟数据库。
// 现在只有一个 users 字段；字段小写开头 = 只在本包内访问（封装）。
type Store struct {
	users map[string]string
}

// NewStore 创建一个带演示账号的 Store。
// Go 里习惯用 NewXxx 当「构造函数」（其实就是普通函数，返回指针）。
func NewStore() *Store {
	return &Store{users: map[string]string{"admin": "admin123"}}
}

// ============================================================
// Handler 层：负责处理 HTTP；它「持有」一块 Store
// ============================================================

// AuthHandler 专门处理登录相关的 HTTP 请求。
//
// 字段 db：这个 Handler 干活时要用到的数据源。
// 谁创建 AuthHandler，谁就得把 Store 传进来 —— 依赖关系写在类型上，一目了然。
type AuthHandler struct {
	db *Store
}

// NewAuthHandler 构造 AuthHandler，并把 Store「注入」进去。
// 参数名叫 db，只是沿用常见叫法；这里其实是内存 Store，不是真数据库连接。
func NewAuthHandler(db *Store) *AuthHandler {
	return &AuthHandler{db: db}
}

// Login 是「方法」：挂在 *AuthHandler 上，所以能用到 h.db。
//
// 对比 v1：
//
//	v1 → 直接用全局 users
//	v2 → 用 h.db.users（依赖来自结构体字段）
//
// 这一版里：读请求、查库、比密码、写响应 —— 仍全部在 Handler 方法里。
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	// 通过「自己带着的 db」查用户，而不是全局变量
	pwd, ok := h.db.users[req.Username]
	if !ok || pwd != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bad credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    "fake-token-for-" + req.Username,
		"username": req.Username,
	})
}

func main() {
	// 第 1 步：造数据源
	store := NewStore()
	// 第 2 步：造 Handler，并把数据源塞进去（db → handler）
	authH := NewAuthHandler(store)

	r := gin.Default()
	// 第 3 步：把 Handler 的方法挂到路由上
	// 注意写的是 authH.Login（某个对象上的方法），不是全局函数
	r.POST("/api/auth/login", authH.Login)

	_ = r.Run(":8082")
}
