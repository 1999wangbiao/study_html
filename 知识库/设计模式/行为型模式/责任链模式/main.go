// 责任链模式可运行示范（Go）—— HTTP 中间件管道（最终形态 ≈ step3）
//
// 建议先按顺序跑演变，对照注释里的调用链：
//
//	evolve/step1_写死管道
//	  → Serve 里写死 Auth / RateLimit / Logging / Biz
//	evolve/step2_中间件链
//	  → Middleware + Chain 组装；不调 next 即短路
//	evolve/step3_只加新中间件
//	  → 旧中间件不动，只加 Timing（本文件同形态）
//
// 最终形态调用链：
//
//	main
//	  └─ Chain(Timing, Auth, RateLimit, Logging).Then(Biz)
//	       Timing → Auth → RateLimit → Logging → Biz
//
// 一句话：请求沿中间件链传递；能拦就拦，能过就 next；发送方只认链头。
//
// 本目录运行：
//
//	go run .
package main

import (
	"fmt"
	"time"
)

// =============================================================================
// 一、角色：Handler / Middleware / Chain
// =============================================================================

// Request 模拟一次 HTTP 请求（不上真 net/http，专心看链）。
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

// Handler 链上统一入口：业务与「包完中间件后的总入口」都是它。
type Handler func(req *Request) *Response

// Middleware 中间件：拿到 next，返回新 Handler（可调用 next，也可短路）。
type Middleware func(next Handler) Handler

// chain 保存待组装的中间件列表。
type chain []Middleware

// Chain 按声明顺序收集中间件。
func Chain(ms ...Middleware) chain {
	return ms
}

// Then 从右往左套娃，使执行顺序为从左到右。
func (ms chain) Then(final Handler) Handler {
	h := final
	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i](h)
	}
	return h
}

// =============================================================================
// 二、具体中间件 + 业务
// =============================================================================

// Timing 最外层计时：能量到整条链（含短路）。
func Timing(next Handler) Handler {
	return func(req *Request) *Response {
		start := time.Now()
		resp := next(req)
		fmt.Printf("  [Timing] %s 耗时=%s status=%d\n",
			req.Path, time.Since(start), resp.Status)
		return resp
	}
}

// Auth 鉴权：无 Token 直接 401，不调用 next。
func Auth(next Handler) Handler {
	return func(req *Request) *Response {
		if req.Token == "" {
			fmt.Println("  [Auth] 拒绝：缺少 Token")
			return &Response{Status: 401, Message: "unauthorized"}
		}
		req.User = "user-" + req.Token
		fmt.Printf("  [Auth] 通过 user=%s\n", req.User)
		return next(req)
	}
}

// RateLimit 限流：闭包持有剩余次数。
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

// Logging 访问日志。
func Logging(next Handler) Handler {
	return func(req *Request) *Response {
		fmt.Printf("  [Logging] %s user=%s\n", req.Path, req.User)
		return next(req)
	}
}

// Biz 链尾业务。
func Biz(req *Request) *Response {
	msg := fmt.Sprintf("hello %s, path=%s", req.User, req.Path)
	fmt.Printf("  [Biz] %s\n", msg)
	return &Response{Status: 200, Message: msg}
}

func main() {
	remain := 2
	handler := Chain(Timing, Auth, RateLimit(&remain), Logging).Then(Biz)

	// 无 Token → Auth 短路；Timing 仍会打印（包在最外）
	fmt.Println("=== 请求 /api/ping token=\"\" ===")
	printResp(handler(&Request{Path: "/api/ping", Token: ""}))

	// 全通过
	fmt.Println("=== 请求 /api/ping token=\"abc\" ===")
	printResp(handler(&Request{Path: "/api/ping", Token: "abc"}))

	// 再过一次
	fmt.Println("=== 请求 /api/me token=\"abc\" ===")
	printResp(handler(&Request{Path: "/api/me", Token: "abc"}))

	// 限流拦下
	fmt.Println("=== 请求 /api/me token=\"abc\"（应被限流）===")
	printResp(handler(&Request{Path: "/api/me", Token: "abc"}))
}

// printResp 打印响应。
func printResp(r *Response) {
	fmt.Printf("→ %d %s\n\n", r.Status, r.Message)
}
