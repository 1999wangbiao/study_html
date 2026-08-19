package main

import "fmt"

// ============================================================
// apperr.go：业务错误 —— Service 只返回「错了什么」，不碰 HTTP
// ============================================================
//
// 为什么单独做一层错误类型？
//   - Service 说「凭证不对」，不必知道要回 401 还是别的
//   - Handler 用 HandleError 统一映射，避免到处手写状态码字符串
//   - 同一业务错误可被多个 Handler 复用

// AppError 带 HTTP 状态与机器可读 code。
type AppError struct {
	HTTP    int
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAppError 构造业务错误。
func NewAppError(httpStatus int, code, message string) *AppError {
	return &AppError{HTTP: httpStatus, Code: code, Message: message}
}

var (
	ErrBadCredentials = NewAppError(401, "invalid_credentials", "invalid username or password")
	ErrUnauthorized   = NewAppError(401, "unauthorized", "authentication required")
	ErrForbidden      = NewAppError(403, "forbidden", "permission denied")
	ErrNotFound       = NewAppError(404, "not_found", "resource not found")
)
