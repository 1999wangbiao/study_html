// ============================================================
// v0：最简单的写法 —— 所有逻辑都写在「路由回调」里面
// ============================================================
//
// 这一版适合：刚学 Gin、只想先让接口跑起来。
// 问题：登录相关的「读参数 / 查用户 / 比密码 / 写响应」全挤在一起，
//       以后接口一多，main 会又长又乱，也不好单独测试业务逻辑。
//
// 演示账号：admin / admin123
// 启动后访问：POST http://127.0.0.1:8080/api/auth/login
//
package main

import (
	"net/http" // 标准库：里面有 200、400、401 这些 HTTP 状态码常量

	"github.com/gin-gonic/gin" // Gin 框架：帮你做路由、读 JSON、写 JSON
)

func main() {
	// ---------- 1）假数据库 ----------
	// 真实项目里这里一般是连 MySQL / SQLite。
	// 教学演示用 map 就够了：key = 用户名，value = 密码。
	//
	// 注意：users 定义在 main 里面，下面的匿名函数（闭包）可以直接用它。
	// 这叫「闭包捕获」——回调函数能看见外面的变量。
	users := map[string]string{
		"admin": "admin123",
	}

	// ---------- 2）创建 Gin 引擎（可以理解成「网站总开关」）----------
	// gin.Default() = 新建一个引擎，并自带日志、panic 恢复等中间件。
	r := gin.Default()

	// ---------- 3）注册路由：POST /api/auth/login ----------
	// 意思：有人用 POST 访问这个路径时，就执行后面这个函数。
	// func(c *gin.Context) { ... } 是「匿名函数」，没有名字，直接写在这里。
	// c 是本次请求的上下文：能读请求体、也能写响应。
	r.POST("/api/auth/login", func(c *gin.Context) {
		// 定义一个临时结构体，用来承接前端发来的 JSON。
		// `json:"username"` 表示：JSON 里字段名是 username，填到 Username 里。
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		// ShouldBindJSON：把请求体里的 JSON 解析进 req。
		// 如果 body 不是合法 JSON，或格式对不上，会返回 error。
		if err := c.ShouldBindJSON(&req); err != nil {
			// 400 = 客户端请求有问题（参数坏了）
			// gin.H 就是 map[string]any 的快捷写法，用来拼 JSON 对象
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return // 出错就结束，别继续往下跑
		}

		// 从假数据库里查这个用户名对应的密码
		// ok == false 表示 map 里没有这个用户名
		pwd, ok := users[req.Username]
		if !ok || pwd != req.Password {
			// 401 = 未认证 / 账号密码不对
			c.JSON(http.StatusUnauthorized, gin.H{"error": "bad credentials"})
			return
		}

		// 登录成功：返回一个「假 token」。
		// 真实项目这里会签发 JWT；本演示只拼一个字符串，方便看懂流程。
		// 注意：查库、比密码、造 token、写 HTTP 响应 —— 全部挤在这个回调里。
		c.JSON(http.StatusOK, gin.H{
			"token":    "fake-token-for-" + req.Username,
			"username": req.Username,
		})
	})

	// ---------- 4）启动 HTTP 服务，监听 8080 端口 ----------
	// r.Run 会阻塞在这里，直到进程退出。
	// 前面的 `_ =` 表示：暂时忽略 Run 返回的 error（教学代码简化写法）。
	_ = r.Run(":8080")
}
