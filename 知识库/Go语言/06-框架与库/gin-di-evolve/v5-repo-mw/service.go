package main

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// newSessionID 用标准库生成会话 id（避免额外依赖，本目录保持可独立运行）。
func newSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ============================================================
// service.go：业务层 —— 仍然不 import gin
// ============================================================
//
// 和 v4 的差别：
//   - 依赖的是 Repository，不再是一个大 Store
//   - 登录成功创建 Session，返回 sessionID（给 Handler 写 Cookie）
//   - 失败返回 *AppError，由 Handler 统一映射 HTTP

const sessionTTL = 2 * time.Hour
const cookieName = "sid"

// CookieName 给 Middleware / Handler 共用 Cookie 名。
func CookieName() string { return cookieName }

// ---------- AuthService ----------

// AuthService 登录 / 会话 / 当前用户。
type AuthService struct {
	users    *UserRepository
	sessions *SessionRepository
}

func NewAuthService(users *UserRepository, sessions *SessionRepository) *AuthService {
	return &AuthService{users: users, sessions: sessions}
}

// Login 校验密码并创建会话，返回 sessionID 与用户公开信息。
func (s *AuthService) Login(username, password string) (sessionID string, user *User, err error) {
	u := s.users.FindByUsername(username)
	if u == nil || u.Password != password {
		return "", nil, ErrBadCredentials
	}
	sid := newSessionID()
	s.sessions.Create(Session{
		ID:       sid,
		Username: u.Username,
		ExpireAt: time.Now().Add(sessionTTL),
	})
	// 返回时去掉密码字段的拷贝（Password 已有 json:"-"，这里再保险清空）
	safe := *u
	safe.Password = ""
	return sid, &safe, nil
}

// Logout 删除会话。
func (s *AuthService) Logout(sessionID string) {
	if sessionID != "" {
		s.sessions.Delete(sessionID)
	}
}

// CurrentUser 根据 sessionID 取当前用户；无效则 unauthorized。
func (s *AuthService) CurrentUser(sessionID string) (*User, error) {
	sess := s.sessions.Find(sessionID)
	if sess == nil {
		return nil, ErrUnauthorized
	}
	u := s.users.FindByUsername(sess.Username)
	if u == nil {
		return nil, ErrUnauthorized
	}
	safe := *u
	safe.Password = ""
	return &safe, nil
}

// ListUsernames 仅给管理端业务用（演示「规则可放 Service」）。
func (s *AuthService) ListUsernames() []string {
	return s.users.ListUsernames()
}

// ---------- ProjectService ----------

// ProjectService 项目业务。
type ProjectService struct {
	projects *ProjectRepository
}

func NewProjectService(projects *ProjectRepository) *ProjectService {
	return &ProjectService{projects: projects}
}

// List 返回项目列表。
func (s *ProjectService) List() []Project {
	return s.projects.List()
}
