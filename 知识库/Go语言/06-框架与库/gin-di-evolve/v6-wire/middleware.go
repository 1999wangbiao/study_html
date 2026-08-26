package main

import "github.com/gin-gonic/gin"

// ============================================================
// middleware.go：门禁 —— 在进 Handler 之前先拦一层
// ============================================================
//
// v0～v4 登录后返回假 token，但后续接口并不校验。
// v5 补上「登录态」和「角色」：
//   RequireAuth  → Cookie 里有没有有效 sid？
//   RequireRole  → 当前用户角色对不对？
//
// 挂路由时写在 Group 上，整组接口自动带门禁：
//   api.Group("/projects", RequireAuth(authSvc))
//   api.Group("/admin", RequireAuth(authSvc), RequireRole(RoleAdmin))

// RequireAuth 校验 Session Cookie，把用户放入 Context。
func RequireAuth(auth *AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid, err := c.Cookie(CookieName())
		if err != nil || sid == "" {
			JSONError(c, 401, "unauthorized", "authentication required")
			return
		}
		u, err := auth.CurrentUser(sid)
		if err != nil {
			HandleError(c, err)
			return
		}
		c.Set(contextUserKey, u)
		c.Next()
	}
}

// RequireRole 要求当前用户属于给定角色之一（须排在 RequireAuth 后面）。
func RequireRole(roles ...string) gin.HandlerFunc {
	set := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		set[r] = struct{}{}
	}
	return func(c *gin.Context) {
		u := CurrentUser(c)
		if u == nil {
			JSONError(c, 401, "unauthorized", "authentication required")
			return
		}
		if _, ok := set[u.Role]; !ok {
			JSONError(c, 403, "forbidden", "permission denied")
			return
		}
		c.Next()
	}
}
