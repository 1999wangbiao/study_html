package main

import (
	"net/http"
	"strconv" // 字符串转整数，用来解析 URL 里的 :id

	"github.com/gin-gonic/gin"
)

// ============================================================
// handler.go：HTTP 层 —— 只做「翻译」
// ============================================================
//
// 每个 Handler 对应一块接口：
//   AuthHandler    → 登录
//   ProjectHandler → 项目列表
//   TaskHandler    → 某项目下的任务
//
// Handler 的标准三步：
//   1）从 c 里取出参数（JSON / 路径参数 / Query）
//   2）调用自己的 Service
//   3）把结果或错误写成 HTTP 响应
//
// 不要在 Handler 里写「查库循环」「权限规则」——那些属于 Service。

// ---------- 认证 Handler ----------

// AuthHandler 持有 AuthService（不再直接持有 Store）。
type AuthHandler struct {
	svc *AuthService
}

// NewAuthHandler 注入 AuthService。
func NewAuthHandler(svc *AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Login 对应路由：POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	// 1）解析 JSON body
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	// 2）调用业务
	res, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		// 3a）业务失败 → 401
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	// 3b）成功 → 200
	c.JSON(http.StatusOK, gin.H{"token": res.Token, "username": res.Username})
}

// ---------- 项目 Handler ----------

// ProjectHandler 持有 ProjectService。
type ProjectHandler struct {
	svc *ProjectService
}

// NewProjectHandler 注入 ProjectService。
func NewProjectHandler(svc *ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// List 对应路由：GET /api/projects
// 这个接口没有请求体，直接问 Service 要列表再包成 JSON。
func (h *ProjectHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": h.svc.List()})
}

// ---------- 任务 Handler ----------

// TaskHandler 持有 TaskService。
type TaskHandler struct {
	svc *TaskService
}

// NewTaskHandler 注入 TaskService。
func NewTaskHandler(svc *TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

// ListByProject 对应路由：GET /api/projects/:id/tasks
//
// URL 里的 :id 是路径参数，例如 /api/projects/1/tasks → id = "1"（字符串）
// 要用 strconv.Atoi 转成 int，才能传给 Service。
func (h *TaskHandler) ListByProject(c *gin.Context) {
	// c.Param("id") 取出路由里名为 id 的那一段
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	items, err := h.svc.ListByProject(id)
	if err != nil {
		// 演示里「项目不可见」统一当成 404；真实项目也可能是 403
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
