// ============================================================
// v1：比 v0 好一点 —— 把处理函数「挪」出路由，单独起个名字
// ============================================================
//
// 相对 v0 的进步：
//   - 路由那一行变成：r.POST(..., Login)，看起来清爽
//   - 登录逻辑集中在 Login 函数里，main 不再那么臃肿
//
// 还没解决的问题：
//   - users 是「包级全局变量」——谁都能改，测试时也很难换成假数据
//   - Login 函数「偷偷」依赖这个全局变量，从函数签名上看不出来
//
// 演示账号：admin / admin123
// 端口：8081
//
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ---------- 包级全局「假数据库」----------
//
// 写在 package 这一层（不在任何函数里面），整个包都能访问。
// 好处：Login 里直接用，写起来省事。
// 坏处：
//   1）依赖关系不直观——光看 func Login(...) 看不出它需要 users
//   2）单测时想换成另一份测试数据，只能改全局，容易互相干扰
//   3）以后多文件协作时，全局状态更容易踩坑
var users = map[string]string{
	"admin": "admin123",
}

// Login 处理「登录」这一次 HTTP 请求。
//
// 参数 c：Gin 的上下文，一次请求对应一个 c。
// 注意：函数参数列表里没有 users，但它内部用了包级变量 users —— 这叫「隐式依赖」。
func Login(c *gin.Context) {
	// 用来承接 JSON 请求体的临时结构
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// 把 body 解析到 req；失败就返回 400
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	// 查全局 users：用户不存在，或密码不对 → 401
	pwd, ok := users[req.Username]
	if !ok || pwd != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bad credentials"})
		return
	}

	// 成功：返回假 token
	c.JSON(http.StatusOK, gin.H{
		"token":    "fake-token-for-" + req.Username,
		"username": req.Username,
	})
}

func main() {
	// 创建引擎
	r := gin.Default()

	// 注册路由时，直接把「函数名 Login」传进去（不再写一大段匿名函数）
	// Gin 收到请求后，会调用 Login(c)
	r.POST("/api/auth/login", Login)

	// 监听 8081，避免和 v0 的 8080 冲突（方便多版同时对照）
	_ = r.Run(":8081")
}
