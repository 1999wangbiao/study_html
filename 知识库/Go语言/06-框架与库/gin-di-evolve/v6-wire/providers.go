package main

// ============================================================
// providers.go：给 Wire 用的「装配胶水」
// ============================================================
//
// 已有的 NewUserRepository / NewAuthService 等本身就是 Provider。
// 这里只补 Wire 推不出来、或需要打包的那一步：
//   两个 Service → Dependencies 结构体

// provideDependencies 把分散的 Service 收成 router 需要的依赖包。
func provideDependencies(auth *AuthService, projects *ProjectService) Dependencies {
	return Dependencies{Auth: auth, Projects: projects}
}
