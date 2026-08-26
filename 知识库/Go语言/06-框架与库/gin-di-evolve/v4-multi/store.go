package main

// ============================================================
// store.go：数据层 —— 「假数据库」放这里
// ============================================================
//
// 真实项目里，这一层往往是 GORM / SQL，对应 *gorm.DB。
// 教学演示用内存结构体就够：启动时塞几条数据，程序退出就没了。

// Store 把用户、项目、任务都放在一起，假装是一整库数据。
type Store struct {
	users    map[string]string // 用户名 → 密码
	projects []Project         // 项目列表
	tasks    []Task            // 任务列表
}

// Project 演示用的「项目」结构。
// 后面的 `json:"id"` 是标签：转成 JSON 时字段名用 id（小写），给前端好看。
type Project struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Task 演示用的「任务」结构。
// ProjectID 表示这条任务属于哪个项目（外键的简化版）。
type Task struct {
	ID        int    `json:"id"`
	ProjectID int    `json:"projectId"`
	Title     string `json:"title"`
}

// NewStore 创建 Store，并写入演示数据。
// 记住演示账号：admin / admin123
func NewStore() *Store {
	return &Store{
		users: map[string]string{"admin": "admin123"},
		projects: []Project{
			{ID: 1, Name: "Demo Project"},
		},
		tasks: []Task{
			{ID: 1, ProjectID: 1, Title: "Write README"},
			{ID: 2, ProjectID: 1, Title: "Wire routes"},
		},
	}
}
