package main

import (
	"sync"
	"time"
)

// ============================================================
// repository.go：数据访问层 —— 替代 v4 的「一个大 Store」
// ============================================================
//
// v4：所有表塞进一个 *Store，Service 直接摸 map。
// v5：按领域拆 Repository（用户 / 会话 / 项目）。
//
// 真实项目里这里往往是 GORM；本目录为保持独立可跑，仍用内存。
// 重点学「边界」：Repository 只读写，不做「最多几条」「角色能不能」这类规则。

// ---------- 内存库（演示底座，真实项目 ≈ *gorm.DB）----------

// memDB 假装一整库；各 Repository 共用同一份数据。
type memDB struct {
	mu       sync.Mutex
	users    map[string]User // username → user
	sessions map[string]Session
	projects []Project
}

func newMemDB() *memDB {
	return &memDB{
		users: map[string]User{
			"admin":   {Username: "admin", Password: "admin123", Role: RoleAdmin},
			"student": {Username: "student", Password: "student123", Role: RoleStudent},
		},
		sessions: map[string]Session{},
		projects: []Project{
			{ID: 1, Name: "Demo Project"},
			{ID: 2, Name: "Homework Tracker"},
		},
	}
}

// ---------- 领域模型（演示用）----------

const (
	RoleAdmin   = "admin"
	RoleStudent = "student"
)

// User 用户。
type User struct {
	Username string `json:"username"`
	Password string `json:"-"` // 不返回给前端
	Role     string `json:"role"`
}

// Session 服务端会话（Cookie 只带 session id）。
type Session struct {
	ID       string
	Username string
	ExpireAt time.Time
}

// Project 项目。
type Project struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ---------- UserRepository ----------

// UserRepository 用户读写。
type UserRepository struct {
	db *memDB
}

func NewUserRepository(db *memDB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByUsername 按用户名查找；找不到返回 nil。
func (r *UserRepository) FindByUsername(username string) *User {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	u, ok := r.db.users[username]
	if !ok {
		return nil
	}
	cp := u
	return &cp
}

// ListUsernames 列出全部用户名（管理端演示用）。
func (r *UserRepository) ListUsernames() []string {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	out := make([]string, 0, len(r.db.users))
	for name := range r.db.users {
		out = append(out, name)
	}
	return out
}

// ---------- SessionRepository ----------

// SessionRepository 会话读写。
type SessionRepository struct {
	db *memDB
}

func NewSessionRepository(db *memDB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create 写入会话。
func (r *SessionRepository) Create(s Session) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	r.db.sessions[s.ID] = s
}

// Find 查找未过期会话；无效则返回 nil。
func (r *SessionRepository) Find(id string) *Session {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	s, ok := r.db.sessions[id]
	if !ok || time.Now().After(s.ExpireAt) {
		return nil
	}
	cp := s
	return &cp
}

// Delete 删除会话（登出）。
func (r *SessionRepository) Delete(id string) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	delete(r.db.sessions, id)
}

// ---------- ProjectRepository ----------

// ProjectRepository 项目读写。
type ProjectRepository struct {
	db *memDB
}

func NewProjectRepository(db *memDB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// List 全部项目。
func (r *ProjectRepository) List() []Project {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	out := make([]Project, len(r.db.projects))
	copy(out, r.db.projects)
	return out
}
