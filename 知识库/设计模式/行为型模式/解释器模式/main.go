// 解释器模式可运行示范（Go）—— 布尔规则表达式
//
// 核心一句话：为一种简单「语言 / 规则」定义文法，并把每条文法规则
// 做成一个 Expression 类型；解释 = 在 Context 上求值整棵表达式树。
//
// 文法（本例）：
//
//	expr  := var | AND | OR
//	var   := 变量名（从 Context 取值）
//	AND   := left AND right
//	OR    := left OR right
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、Context（上下文）——「解释时用到的环境」
// =============================================================================
//
// 终端表达式（变量）会来这里查布尔值。
// 课堂里用 map 即可；生产里可以是请求、用户画像、配置等。

// Context 变量名 → 布尔值。
type Context map[string]bool

// Get 读取变量；未定义时视为 false（演示简化）。
func (c Context) Get(name string) bool {
	return c[name]
}

// =============================================================================
// 二、AbstractExpression（抽象表达式）
// =============================================================================

// Expression 所有表达式节点都能在 Context 上解释出 bool。
type Expression interface {
	Interpret(ctx Context) bool
	String() string // 仅用于打印树，方便对照日志
}

// =============================================================================
// 三、TerminalExpression：变量
// =============================================================================

// VarExpr 终端表达式：查 Context 里的变量。
type VarExpr struct {
	Name string
}

func (e *VarExpr) Interpret(ctx Context) bool {
	v := ctx.Get(e.Name)
	fmt.Printf("  [Var] %s = %v\n", e.Name, v)
	return v
}

func (e *VarExpr) String() string { return e.Name }

// =============================================================================
// 四、NonterminalExpression：AND / OR
// =============================================================================
//
// 非终端节点持有子表达式，解释时先解释左右，再组合结果。

// AndExpr：左 AND 右。
type AndExpr struct {
	Left, Right Expression
}

func (e *AndExpr) Interpret(ctx Context) bool {
	fmt.Printf("  [And] 解释 (%s) AND (%s)\n", e.Left, e.Right)
	l := e.Left.Interpret(ctx)
	// 短路：左边已 false 可不再算右边（教学上也演示「组合如何求值」）
	if !l {
		fmt.Printf("  [And] 左为 false，短路 → false\n")
		return false
	}
	r := e.Right.Interpret(ctx)
	fmt.Printf("  [And] %v AND %v → %v\n", l, r, l && r)
	return l && r
}

func (e *AndExpr) String() string {
	return fmt.Sprintf("(%s AND %s)", e.Left, e.Right)
}

// OrExpr：左 OR 右。
type OrExpr struct {
	Left, Right Expression
}

func (e *OrExpr) Interpret(ctx Context) bool {
	fmt.Printf("  [Or] 解释 (%s) OR (%s)\n", e.Left, e.Right)
	l := e.Left.Interpret(ctx)
	if l {
		fmt.Printf("  [Or] 左为 true，短路 → true\n")
		return true
	}
	r := e.Right.Interpret(ctx)
	fmt.Printf("  [Or] %v OR %v → %v\n", l, r, l || r)
	return l || r
}

func (e *OrExpr) String() string {
	return fmt.Sprintf("(%s OR %s)", e.Left, e.Right)
}

// =============================================================================
// 五、组装规则树 + main
// =============================================================================
//
// 业务规则（人工组装 AST，不写解析器）：
//
//	vip AND (score OR staff)
//
// 含义：会员，且（高分 或 内部员工）才放行。

func buildRule() Expression {
	return &AndExpr{
		Left: &VarExpr{Name: "vip"},
		Right: &OrExpr{
			Left:  &VarExpr{Name: "score"},
			Right: &VarExpr{Name: "staff"},
		},
	}
}

func main() {
	rule := buildRule()
	fmt.Println("规则树:", rule)
	fmt.Println()

	fmt.Println("========== 用例 1：vip=true, score=false, staff=true → 应 true ==========")
	ctx1 := Context{"vip": true, "score": false, "staff": true}
	fmt.Printf("结果: %v\n\n", rule.Interpret(ctx1))

	fmt.Println("========== 用例 2：vip=true, score=false, staff=false → 应 false ==========")
	ctx2 := Context{"vip": true, "score": false, "staff": false}
	fmt.Printf("结果: %v\n\n", rule.Interpret(ctx2))

	fmt.Println("========== 用例 3：vip=false, score=true, staff=true → 应 false（短路）==========")
	ctx3 := Context{"vip": false, "score": true, "staff": true}
	fmt.Printf("结果: %v\n\n", rule.Interpret(ctx3))

	fmt.Println("========== 读懂输出后你会发现 ==========")
	fmt.Println("1. 每条文法规则 ≈ 一个 Expression 类型")
	fmt.Println("2. Interpret 递归：非终端先解释子树，再组合")
	fmt.Println("3. Context 提供变量环境；换 Context = 同一规则套不同数据")
	fmt.Println("4. 本例 AST 手写组装；真语言还要词法/语法分析，那是编译器课的事")
}
