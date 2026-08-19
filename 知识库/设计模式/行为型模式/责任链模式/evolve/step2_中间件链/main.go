// 演变第 2 步：抽出 Middleware，用 Chain 组装管道
//
// 运行：
//
//	cd evolve/step2_中间件链
//	go run .
//
// =============================================================================
// 相对 step1，改了什么（对照看）
// =============================================================================
//
// step1：
//	Serve 里：if Token → if remain → Logging → handleBiz
//
// step2：
//	Handler    = func(*Request) *Response
//	Middleware = func(next Handler) Handler   ← 包一层，决定是否调用 next
//	Chain(ms...).Then(biz) 从内到外（或从右到左）套娃
//
// 调用链（洋葱 / 责任链）：
//
//	main
//	  └─ handler(req)          ← Chain(Auth, RateLimit, Logging).Then(Biz)
//	       └─ Auth
//	            └─ RateLimit     （Auth 通过才进）
//	                 └─ Logging
//	                      └─ Biz
//
// 本步额外演示：鉴权失败时不调用 next → 后面整段短路。
// 新中间件不必改 Auth / RateLimit / Biz —— 到 step3 用 Timing 验证。
package main

import "fmt"

// =============================================================================
// 一、抽象：Handler / Middleware / Chain（step1 没有这层）
// =============================================================================

// Request 模拟一次 HTTP 请求。
type Request struct {
	Path  string
	Token string
	User  string
}

// Response 模拟响应。
type Response struct {
	Status  int
	Message string
}

// Handler 责任链上的「节点约定」：吃请求，吐响应。
// 业务 handler 和「包完中间件后的入口」都是这个类型。
type Handler func(req *Request) *Response

// Middleware 中间件：拿到 next，返回一个新的 Handler。
// 新 Handler 里可以：前置逻辑 → 调 next → 后置逻辑；也可以直接 return 不调 next（中断链）。
type Middleware func(next Handler) Handler

// chain 保存待组装的中间件列表，提供 Then 接到业务 Handler。
type chain []Middleware

// Chain 把多个中间件按「从左到右」的执行顺序收进列表。
//
//	Chain(A, B, C).Then(biz) 实际调用顺序：A → B → C → biz
func Chain(ms ...Middleware) chain {
	return ms
}

// Then 从右往左套娃：最左边的中间件成为最外层，最先执行。
func (ms chain) Then(final Handler) Handler {
	h := final
	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i](h) // 从右往左包：先包 C，再包 B，最后包 A
	}
	return h
}

// =============================================================================
// 二、具体中间件（ConcreteHandler）+ 业务 Handler
// =============================================================================

// Auth 鉴权：无 Token 直接 401，不调用 next。
func Auth(next Handler) Handler {
	return func(req *Request) *Response {
		if req.Token == "" {
			fmt.Println("  [Auth] 拒绝：缺少 Token")
			return &Response{Status: 401, Message: "unauthorized"}
		}
		req.User = "user-" + req.Token
		fmt.Printf("  [Auth] 通过 user=%s\n", req.User)
		return next(req) // 交给链上下一个
	}
}

// RateLimit 限流：闭包抓住 remain 计数器。
func RateLimit(remain *int) Middleware {
	return func(next Handler) Handler {
		return func(req *Request) *Response {
			if *remain <= 0 {
				fmt.Println("  [RateLimit] 拒绝：超额")
				return &Response{Status: 429, Message: "too many requests"}
			}
			*remain--
			fmt.Printf("  [RateLimit] 放行，剩余=%d\n", *remain)
			return next(req)
		}
	}
}

// Logging 访问日志：通过后记一笔再往下。
func Logging(next Handler) Handler {
	return func(req *Request) *Response {
		fmt.Printf("  [Logging] %s user=%s\n", req.Path, req.User)
		return next(req)
	}
}

// Biz 真正的业务 Handler（链尾）。
func Biz(req *Request) *Response {
	msg := fmt.Sprintf("hello %s, path=%s", req.User, req.Path)
	fmt.Printf("  [Biz] %s\n", msg)
	return &Response{Status: 200, Message: msg}
}

func main() {
	remain := 2

	// 组装：Auth → RateLimit → Logging → Biz
	// 对比 step1：顺序只出现在这一行，Serve 方法体消失了
	handler := Chain(Auth, RateLimit(&remain), Logging).Then(Biz)

	fmt.Println("=== 请求 /api/ping token=\"\" ===")
	printResp(handler(&Request{Path: "/api/ping", Token: ""}))

	fmt.Println("=== 请求 /api/ping token=\"abc\" ===")
	printResp(handler(&Request{Path: "/api/ping", Token: "abc"}))

	fmt.Println("=== 请求 /api/me token=\"abc\" ===")
	printResp(handler(&Request{Path: "/api/me", Token: "abc"}))

	fmt.Println("=== 请求 /api/me token=\"abc\"（应被限流）===")
	printResp(handler(&Request{Path: "/api/me", Token: "abc"}))
}

// printResp 打印响应。
func printResp(r *Response) {
	fmt.Printf("→ %d %s\n\n", r.Status, r.Message)
}
