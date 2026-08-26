// ============================================================
// v3：再拆一层 Service —— 业务规则和 HTTP 分开
// ============================================================
//
// 分层口诀（从下往上）：
//   Store     = 数据放哪（本演示是内存 map）
//   Service   = 业务怎么做（校验密码、造 token 结果）—— 不碰 gin
//   Handler   = HTTP 怎么接（读 JSON、选状态码、写响应）—— 不写业务细节
//   main/路由 = 把上面几层「拧」在一起，挂到 URL 上
//
// 为什么要拆：
//   - 同一套登录业务，将来可能被 HTTP、命令行、测试一起调用
//   - Service 返回 error，由 Handler 决定是 401 还是 500，职责清晰
//
// 组装顺序永远记住：
//   store → service → handler → 路由
//
// 演示账号：admin / admin123
// 端口：8083
//
package main

import (
	"errors"   // 用来 new 一个错误对象返回给上层
	"net/http" // HTTP 状态码

	"github.com/gin-gonic/gin"
)

// ============================================================
// 一、数据层（演示用内存库）
// ============================================================

// Store 模拟数据库：用户名 → 密码。
type Store struct {
	users map[string]string
}

// NewStore 准备一份演示数据。
func NewStore() *Store {
	return &Store{users: map[string]string{"admin": "admin123"}}
}

// ============================================================
// 二、Service 层：只关心「业务」，不要 import / 使用 gin
// ============================================================
//
// 好习惯：Service 的方法参数尽量是普通类型（string、int、struct），
// 不要传 *gin.Context。这样它就和 Web 框架解耦了。

// AuthService 负责「登录」这件事的业务逻辑。
type AuthService struct {
	db *Store // 它需要能查用户，所以持有 Store
}

// NewAuthService 把 Store 注入进 AuthService。
func NewAuthService(db *Store) *AuthService {
	return &AuthService{db: db}
}

// LoginResult 是业务层的「成功结果」。
// 注意：这里没有 http 状态码，也没有 gin.H —— 业务层不该关心 HTTP。
type LoginResult struct {
	Token    string
	Username string
}

// Login 校验用户名密码。
//
// 成功：返回 LoginResult
// 失败：返回 error（比如密码错）—— 具体映射成 401，交给 Handler 决定
func (s *AuthService) Login(username, password string) (*LoginResult, error) {
	pwd, ok := s.db.users[username]
	if !ok || pwd != password {
		return nil, errors.New("bad credentials")
	}
	return &LoginResult{
		Token:    "fake-token-for-" + username, // 演示用假 token
		Username: username,
	}, nil
}

// ============================================================
// 三、Handler 层：只关心 HTTP 入参 / 出参
// ============================================================

// AuthHandler 是 HTTP 适配器：把「一次 Web 请求」翻译成「一次 Service 调用」。
type AuthHandler struct {
	svc *AuthService // 注意：这里持有的是 Service，不再直接持有 Store
}

// NewAuthHandler 注入 AuthService。
func NewAuthHandler(svc *AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Login：读 JSON → 调 service → 按结果写 JSON / 状态码。
func (h *AuthHandler) Login(c *gin.Context) {
	// 1）解析请求体
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	// 2）把「纯业务」交给 Service（Handler 自己不查 map、不比密码）
	res, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		// 3a）业务失败 → 选一个合适的 HTTP 状态码返回
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	// 3b）业务成功 → 200 + JSON
	c.JSON(http.StatusOK, gin.H{"token": res.Token, "username": res.Username})
}

// ============================================================
// 四、组装：store → service → handler → 路由
// ============================================================

func main() {
	// 像搭积木一样，从里到外创建：
	store := NewStore()                 // 原料：数据
	authSvc := NewAuthService(store)    // 车间：业务（db → service）
	authH := NewAuthHandler(authSvc)    // 柜台：HTTP（service → handler）

	r := gin.Default()
	// 最后把柜台方法挂到网址上
	r.POST("/api/auth/login", authH.Login)

	_ = r.Run(":8083")
}
