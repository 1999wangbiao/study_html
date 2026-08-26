package main

import "github.com/gin-gonic/gin"

// ============================================================
// router.go：造 Handler + 挂路由 + 挂中间件
// ============================================================
//
// v4 把「New Service / New Handler / 挂路由」全写在 main。
// v5 习惯拆开：
//   main   → 造 DB、Repository、Service（启动组装）
//   router → 造 Handler、挂 Middleware、挂 URL（路由表）
//
// 依赖用结构体传进来，避免 router 里再偷偷 New 全局变量。

// Dependencies 路由所需的业务依赖（由 main 注入）。
type Dependencies struct {
	Auth     *AuthService
	Projects *ProjectService
}

// NewRouter 注册全部 API。
func NewRouter(deps Dependencies) *gin.Engine {
	r := gin.Default()

	authH := NewAuthHandler(deps.Auth)
	projectH := NewProjectHandler(deps.Projects)
	adminH := NewAdminHandler(deps.Auth)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authH.Login) // 公开
			auth.POST("/logout", authH.Logout)
			auth.GET("/me", RequireAuth(deps.Auth), authH.Me)
		}

		// 登录后可看项目
		projects := api.Group("/projects", RequireAuth(deps.Auth))
		{
			projects.GET("", projectH.List)
		}

		// 仅 admin
		admin := api.Group("/admin", RequireAuth(deps.Auth), RequireRole(RoleAdmin))
		{
			admin.GET("/users", adminH.ListUsers)
		}
	}
	return r
}
