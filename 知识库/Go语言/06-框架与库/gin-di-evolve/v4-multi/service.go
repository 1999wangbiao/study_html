package main

import "errors"

// ============================================================
// service.go：业务层 —— 这里仍然不要出现 gin
// ============================================================
//
// 本文件里有 3 个 Service，对应 3 块业务：
//   AuthService    → 登录
//   ProjectService → 项目列表 / 能不能看某个项目
//   TaskService    → 某项目下的任务（会先问 ProjectService）
//
// 重点看 TaskService：它除了拿 Store，还拿了一个 *ProjectService。
// 含义：复用已有「项目能不能看」的逻辑，别在任务里再抄一份。

// ---------- 1）认证业务 ----------

// AuthService 登录业务，持有数据源 Store。
type AuthService struct {
	db *Store
}

// NewAuthService 构造认证服务：外面把 Store 传进来。
func NewAuthService(db *Store) *AuthService {
	return &AuthService{db: db}
}

// LoginResult 登录成功后，业务层返回给上层的数据（还不是 HTTP 响应）。
type LoginResult struct {
	Token    string
	Username string
}

// Login 校验账号密码。
// 成功返回结果；失败返回 error，由 Handler 决定返回 401 还是别的。
func (s *AuthService) Login(username, password string) (*LoginResult, error) {
	pwd, ok := s.db.users[username]
	if !ok || pwd != password {
		return nil, errors.New("bad credentials")
	}
	return &LoginResult{Token: "fake-token-for-" + username, Username: username}, nil
}

// ---------- 2）项目业务 ----------

// ProjectService 管「项目」相关规则。
type ProjectService struct {
	db *Store
}

// NewProjectService 构造项目服务。
func NewProjectService(db *Store) *ProjectService {
	return &ProjectService{db: db}
}

// List 返回全部项目。
// 真实项目里这里可能还要按「当前用户能看见哪些」过滤；演示先全返回。
func (s *ProjectService) List() []Project {
	return s.db.projects
}

// CanView 判断某个 projectID 是否「可见」。
// 演示版规则很简单：库里有这个 id 就算可见。
// 真实项目可能是：是成员 / 是负责人 / 是 admin 等。
//
// 这个方法会被 TaskService 调用 —— 所以「能不能看项目」的规则只写一份。
func (s *ProjectService) CanView(projectID int) bool {
	for _, p := range s.db.projects {
		if p.ID == projectID {
			return true
		}
	}
	return false
}

// ---------- 3）任务业务（依赖 ProjectService）----------

// TaskService 管任务。
// 字段说明：
//   db → 自己查任务列表要用
//   ps → 复用 ProjectService.CanView，避免在任务里再写一遍「项目是否存在」
type TaskService struct {
	db *Store
	ps *ProjectService
}

// NewTaskService 注入两个依赖：Store + ProjectService。
func NewTaskService(db *Store, ps *ProjectService) *TaskService {
	return &TaskService{db: db, ps: ps}
}

// ListByProject 列出某项目下的任务。
// 步骤：
//   1）先问项目服务：这个项目能不能看？不能 → 直接报错
//   2）能看 → 从 Store 里筛出 ProjectID 相同的任务
func (s *TaskService) ListByProject(projectID int) ([]Task, error) {
	// 关键：业务复用发生在这里，不是在 Handler 里复制 if
	if !s.ps.CanView(projectID) {
		return nil, errors.New("project not found or not visible")
	}
	var out []Task
	for _, t := range s.db.tasks {
		if t.ProjectID == projectID {
			out = append(out, t)
		}
	}
	return out, nil
}
