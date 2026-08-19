package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ============================================================
// handler.go：HTTP 层 + 统一响应
// ============================================================
//
// v5 相对 v4 多了两件事：
//   1）登录成功用 SetCookie 写 sid（不再只返回假 token 字符串）
//   2）业务错误走 HandleError，统一成 {code, message}

const contextUserKey = "currentUser"

// JSONError 统一错误体。
func JSONError(c *gin.Context, status int, code, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"code": code, "message": msg})
}

// JSONData 统一成功体。
func JSONData(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

// HandleError 把 *AppError 映射成 HTTP；未知错误回 500。
func HandleError(c *gin.Context, err error) {
	var ae *AppError
	if errors.As(err, &ae) {
		JSONError(c, ae.HTTP, ae.Code, ae.Message)
		return
	}
	JSONError(c, http.StatusInternalServerError, "internal_error", "internal server error")
}

// CurrentUser 从 Context 取 Middleware 放入的用户。
func CurrentUser(c *gin.Context) *User {
	v, ok := c.Get(contextUserKey)
	if !ok {
		return nil
	}
	u, _ := v.(*User)
	return u
}

// ---------- AuthHandler ----------

// AuthHandler 认证接口。
type AuthHandler struct {
	svc *AuthService
}

func NewAuthHandler(svc *AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Login POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, "bad_request", "invalid body")
		return
	}
	sid, user, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		HandleError(c, err)
		return
	}
	// HttpOnly：JS 读不到 Cookie，降低 XSS 偷会话风险（演示要点）
	c.SetCookie(CookieName(), sid, int(sessionTTL.Seconds()), "/", "", false, true)
	JSONData(c, http.StatusOK, user)
}

// Logout POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	sid, _ := c.Cookie(CookieName())
	h.svc.Logout(sid)
	c.SetCookie(CookieName(), "", -1, "/", "", false, true)
	JSONData(c, http.StatusOK, gin.H{"ok": true})
}

// Me GET /api/auth/me —— 需要先过 RequireAuth
func (h *AuthHandler) Me(c *gin.Context) {
	JSONData(c, http.StatusOK, CurrentUser(c))
}

// ---------- ProjectHandler ----------

// ProjectHandler 项目接口。
type ProjectHandler struct {
	svc *ProjectService
}

func NewProjectHandler(svc *ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// List GET /api/projects —— 需要登录
func (h *ProjectHandler) List(c *gin.Context) {
	JSONData(c, http.StatusOK, gin.H{"items": h.svc.List()})
}

// ---------- AdminHandler ----------

// AdminHandler 管理端演示接口。
type AdminHandler struct {
	auth *AuthService
}

func NewAdminHandler(auth *AuthService) *AdminHandler {
	return &AdminHandler{auth: auth}
}

// ListUsers GET /api/admin/users —— 需要 admin 角色
func (h *AdminHandler) ListUsers(c *gin.Context) {
	JSONData(c, http.StatusOK, gin.H{"usernames": h.auth.ListUsernames()})
}
