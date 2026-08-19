// 演变第 1 步：HTTP 管道写死在 Serve 里
//
// 运行：
//
//	cd evolve/step1_写死管道
//	go run .
//
// =============================================================================
// 本步要体会的「错误直觉」
// =============================================================================
//
// 需求：每个请求先鉴权，再限流，再记日志，最后进业务。
// 做法：Server.Serve 里按顺序写死 if / 调用，像一条硬编码流水线。
//
// 调用链（死板、写死）：
//
//	main
//	  └─ Serve(req)
//	       ├─ 检查 Token（写在 Serve 里）
//	       ├─ 检查限流（写在 Serve 里）
//	       ├─ 打日志（写在 Serve 里）
//	       └─ handleBiz(req)
//
// 痛点：
//  1. 加「计时 / CORS」→ 只能改 Serve，核心函数越堆越长
//  2. 想换顺序（先日志再鉴权）→ 改 Serve 里语句顺序，容易漏
//  3. 鉴权失败时「后面不跑」靠 return 散落各处，没有统一「next」语义
//
// 下一步（step2）会改成：Middleware = func(next Handler) Handler，用 Chain 组装。
package main

import "fmt"

// =============================================================================
// 轻量 Request / Response：不上真 net/http，专心看「管道」怎么长出来
// =============================================================================

// Request 模拟一次 HTTP 请求。
type Request struct {
	Path  string
	Token string
	User  string // 鉴权通过后填上
}

// Response 模拟响应。
type Response struct {
	Status  int
	Message string
}

// =============================================================================
// Server：业务入口 + 横切逻辑全揉在 Serve（本步的核心问题）
// =============================================================================

// Server 同时干两件事：跑业务，以及「记得先过哪几道关」。
// 后一件事正是耦合来源——横切关注点写死在 Serve 方法体里。
type Server struct {
	// 极简限流：全局还能放行几次（演示用，不是真令牌桶）
	remain int
}

// NewServer 创建服务；remain 表示还能接受几次「已鉴权」请求。
func NewServer(remain int) *Server {
	return &Server{remain: remain}
}

// Serve 通知逻辑的「痛点现场」：鉴权 / 限流 / 日志 / 业务全挤在一起。
func (s *Server) Serve(req *Request) *Response {
	fmt.Printf("=== 请求 %s token=%q ===\n", req.Path, req.Token)

	// ① 鉴权：写死在 Serve 开头
	if req.Token == "" {
		fmt.Println("  [Auth] 拒绝：缺少 Token")
		return &Response{Status: 401, Message: "unauthorized"}
	}
	req.User = "user-" + req.Token
	fmt.Printf("  [Auth] 通过 user=%s\n", req.User)

	// ② 限流：又写死一段
	if s.remain <= 0 {
		fmt.Println("  [RateLimit] 拒绝：超额")
		return &Response{Status: 429, Message: "too many requests"}
	}
	s.remain--
	fmt.Printf("  [RateLimit] 放行，剩余=%d\n", s.remain)

	// ③ 日志：再写死一段
	fmt.Printf("  [Logging] %s user=%s\n", req.Path, req.User)

	// ④ 业务：终于轮到真正干活的
	// 想加 Timing？只能在上面再插一段，或包一层——都要动 Serve。
	return s.handleBiz(req)
}

// handleBiz 唯一的「正经业务」。
func (s *Server) handleBiz(req *Request) *Response {
	msg := fmt.Sprintf("hello %s, path=%s", req.User, req.Path)
	fmt.Printf("  [Biz] %s\n", msg)
	return &Response{Status: 200, Message: msg}
}

func main() {
	srv := NewServer(2)

	// 无 Token → 鉴权拦下，限流 / 日志 / 业务都不跑
	printResp(srv.Serve(&Request{Path: "/api/ping", Token: ""}))

	// 有 Token → 全管道走过
	printResp(srv.Serve(&Request{Path: "/api/ping", Token: "abc"}))

	// 再一次有 Token → 仍可放行（remain 从 2 减到 0 前还能过）
	printResp(srv.Serve(&Request{Path: "/api/me", Token: "abc"}))

	// 第三次有 Token → 限流拦下（业务不跑）
	printResp(srv.Serve(&Request{Path: "/api/me", Token: "abc"}))
}

// printResp 打印响应，方便对照。
func printResp(r *Response) {
	fmt.Printf("→ %d %s\n\n", r.Status, r.Message)
}
