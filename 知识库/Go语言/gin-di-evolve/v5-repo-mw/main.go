// ============================================================
// v5：Repository + Middleware + Session + 统一业务错误
// ============================================================
//
// 相对 v4 多学四件事（本目录内可独立跑通，不依赖其它工程）：
//
//   1）Store 拆成多个 Repository（按领域访问数据）
//   2）登录用 Session Cookie，而不是只返回假 token
//   3）Middleware：RequireAuth / RequireRole 做门禁
//   4）AppError + HandleError：业务错误与 HTTP 解耦
//
// 组装顺序：
//
//	memDB → Repository → Service →（router 内）Handler + Middleware → 路由
//
// 演示账号：
//
//	admin / admin123     （角色 admin，可访问 /api/admin/users）
//	student / student123 （角色 student，登录后可看项目，不可进 admin）
//
// 端口：8085
package main

import "log"

func main() {
	// 0）数据底座（真实项目这里往往是 gorm.Open(...)）
	db := newMemDB()

	// 1）Repository（仓库）
	userRepo := NewUserRepository(db)
	sessionRepo := NewSessionRepository(db)
	projectRepo := NewProjectRepository(db)

	// 2）Service（脑子）
	authSvc := NewAuthService(userRepo, sessionRepo)
	projectSvc := NewProjectService(projectRepo)

	// 3）交给 router：内部再造 Handler、挂中间件与 URL
	r := NewRouter(Dependencies{
		Auth:     authSvc,
		Projects: projectSvc,
	})

	log.Println("v5 listening on :8085")
	if err := r.Run(":8085"); err != nil {
		log.Fatal(err)
	}
}
