// ============================================================
// v4：多个模块并排组装
// ============================================================
//
// 前面 v0～v3 只有「登录」一块。
// 模块一多就会出现这种写法：
//
//	先 New 一排 Service
//	再 New 一排 Handler
//	最后把 Handler 方法挂到路由上
//
// 本文件（main.go）几乎不写业务，只做「组装 + 启动」。
// 业务在 service.go，HTTP 在 handler.go，数据在 store.go。
//
// 演示账号：admin / admin123
// 端口：8084
// 额外接口：
//
//	GET /api/projects
//	GET /api/projects/1/tasks
//
// 学完本版后继续看 v5-repo-mw（Repository / Session / Middleware）。
package main

import "github.com/gin-gonic/gin"

func main() {
	// ---------- 0）准备数据源（对应真实项目里的 *gorm.DB）----------
	store := NewStore()

	// ---------- 1）依赖注入：先造 Service（「脑子」）----------
	// 每一行 = 创建一个业务对象，并把依赖塞进去。
	authSvc := NewAuthService(store)       // 登录业务只需要 Store
	projectSvc := NewProjectService(store) // 项目业务只需要 Store

	// 任务业务要复用「项目能不能看」→ 多注入一个 projectSvc
	taskSvc := NewTaskService(store, projectSvc)

	// ---------- 2）再造 Handler（「嘴巴」）----------
	// Handler 只依赖自己的 Service，不直接拿 Store。
	authH := NewAuthHandler(authSvc)
	projectH := NewProjectHandler(projectSvc)
	taskH := NewTaskHandler(taskSvc)

	// ---------- 3）挂路由（「门口菜单」）----------
	r := gin.Default()

	// Group("/api")：这一组路由都自动带前缀 /api
	// 这样不用每个路径都手写 "/api/...."
	api := r.Group("/api")
	{
		api.POST("/auth/login", authH.Login)                // POST /api/auth/login
		api.GET("/projects", projectH.List)                 // GET  /api/projects
		api.GET("/projects/:id/tasks", taskH.ListByProject) // GET  /api/projects/:id/tasks
	}

	// ---------- 4）启动服务 ----------
	_ = r.Run(":8084")
}
