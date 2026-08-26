// ============================================================
// v6：用 Google Wire 生成组装代码（业务同 v5）
// ============================================================
//
// 和 v5 比，只改「谁写那串 New」：
//
//	v5：main 里手写 memDB → Repo → Service → Router
//	v6：main 只调 InitializeApp()；组装在 wire_gen.go
//
// 分层不变：Repository / Service / Handler / Middleware 与 v5 相同。
// 请对照阅读：
//
//	wire.go      → 你声明 Provider 列表（生成用）
//	wire_gen.go  → Wire 生成的组装（日常运行用这个）
//	main.go      → 启动
//
// 演示账号同 v5：
//
//	admin / admin123
//	student / student123
//
// 端口：8086
package main

import "log"

func main() {
	// 一行搞定：依赖树在 wire_gen.go 里
	r := InitializeApp()

	log.Println("v6 (wire) listening on :8086")
	if err := r.Run(":8086"); err != nil {
		log.Fatal(err)
	}
}
