// 演变第 3 步：只加 Timing 中间件，Chain / Auth / RateLimit / Logging / Biz 不动
//
// 运行：
//
//	cd evolve/step3_只加新中间件
//	go run .
//
// =============================================================================
// 相对 step2，改了什么 / 没改什么
// =============================================================================
//
// 没改（请对照 step2 文件，应几乎一字不差）：
//	- Handler / Middleware / Chain
//	- Auth / RateLimit / Logging / Biz
//
// 只加了：
//	- 新中间件 Timing（记录耗时）
//	- main 里 Chain 多传一个 Timing
//
// 这就是责任链 / 中间件「对扩展开放」的体感：
//	扩展点 = 新 Middleware；稳定点 = 旧中间件 + 业务 + Chain 本身。
//
// 若还用 step1 写法，加计时必须改 Serve 方法体。
//
// 调用链（与 step2 相同，只是多一层）：
//
//	handler = Chain(Timing, Auth, RateLimit, Logging).Then(Biz)
//	  Timing → Auth → RateLimit → Logging → Biz
package main

import (
	"fmt"
	"time"
)

// =============================================================================
// 以下 Handler / Middleware / Chain / 旧中间件：刻意与 step2 保持同构（稳定点）
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

// Handler 处理函数。
type Handler func(req *Request) *Response

// Middleware 中间件：包装 next。
type Middleware func(next Handler) Handler

// chain 保存待组装的中间件列表。
type chain []Middleware

// Chain 收集中间件，供 Then 接到业务 Handler。
func Chain(ms ...Middleware) chain {
	return ms
}

// Then 从右往左套娃，保证从左到右执行。
func (ms chain) Then(final Handler) Handler {
	h := final
	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i](h)
	}
	return h
}

// Auth 鉴权中间件。
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

// RateLimit 限流中间件。
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

// Logging 日志中间件。
func Logging(next Handler) Handler {
	return func(req *Request) *Response {
		fmt.Printf("  [Logging] %s user=%s\n", req.Path, req.User)
		return next(req)
	}
}

// Biz 业务 Handler。
func Biz(req *Request) *Response {
	msg := fmt.Sprintf("hello %s, path=%s", req.User, req.Path)
	fmt.Printf("  [Biz] %s\n", msg)
	return &Response{Status: 200, Message: msg}
}

// =============================================================================
// 本步唯一新增：Timing（扩展点）
// =============================================================================

// Timing 计时中间件：进链前记下时刻，next 返回后再打印耗时。
// 注意它包在最外层时，能量到「整条链」的时间（含鉴权失败的短路径）。
func Timing(next Handler) Handler {
	return func(req *Request) *Response {
		start := time.Now()
		resp := next(req) // 无论后面成功还是短路，都会回到这里
		fmt.Printf("  [Timing] %s 耗时=%s status=%d\n",
			req.Path, time.Since(start), resp.Status)
		return resp
	}
}

func main() {
	remain := 2

	// 旧中间件不动；只在 Chain 最前面加 Timing
	// 执行顺序：Timing → Auth → RateLimit → Logging → Biz
	handler := Chain(Timing, Auth, RateLimit(&remain), Logging).Then(Biz)

	fmt.Println("=== 请求 /api/ping token=\"\" ===")
	printResp(handler(&Request{Path: "/api/ping", Token: ""}))

	fmt.Println("=== 请求 /api/ping token=\"abc\" ===")
	printResp(handler(&Request{Path: "/api/ping", Token: "abc"}))

	fmt.Println("=== 请求 /api/me token=\"abc\" ===")
	printResp(handler(&Request{Path: "/api/me", Token: "abc"}))
}

// printResp 打印响应。
func printResp(r *Response) {
	fmt.Printf("→ %d %s\n\n", r.Status, r.Message)
}
