// 策略模式可运行示范（Go）—— 电商促销折扣
//
// 核心一句话：把「可替换的算法」各自封装成独立的策略类型，
// Context（Order）只负责持有当前策略并转发结算，运行时想换就换。
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、角色 1：Strategy（策略）——「计价规则」的统一形状
// =============================================================================

// Discount 折扣策略：对原价应用一种计价规则，返回最终应付。
type Discount interface {
	Name() string
	Apply(price float64) float64
}

// =============================================================================
// 二、角色 2：Context（上下文）——「订单本体」
// =============================================================================
//
// Order 对外只暴露结算动作；具体怎么打折，全部委托给当前策略。
// 换策略不用改 Order 一行代码。

// Order 订单上下文：持有当前折扣策略。
type Order struct {
	id       string
	discount Discount
}

// NewOrder 创建一个订单，并注入初始折扣策略（nil 时回退为无折扣）。
func NewOrder(id string, d Discount) *Order {
	if d == nil {
		d = &NoDiscount{}
	}
	return &Order{id: id, discount: d}
}

// SetDiscount 运行时切换折扣策略。
func (o *Order) SetDiscount(d Discount) {
	o.discount = d
}

// Settle 结算：原价交给当前策略计算，并打印明细。
func (o *Order) Settle(price float64) {
	final := o.discount.Apply(price)
	fmt.Printf("[订单 %s] 原价 %.0f → %s 后：%.0f\n", o.id, price, o.discount.Name(), final)
}

// =============================================================================
// 三、具体策略：无折扣 / 满减 / 会员折
// =============================================================================

// NoDiscount 无折扣：原价直付。
type NoDiscount struct{}

func (d *NoDiscount) Name() string            { return "无折扣" }
func (d *NoDiscount) Apply(p float64) float64 { return p }

// FullReduction 满减：满 threshold 元减 cut 元。
type FullReduction struct {
	threshold float64
	cut       float64
}

func (d *FullReduction) Name() string { return fmt.Sprintf("满%.0f减%.0f", d.threshold, d.cut) }

func (d *FullReduction) Apply(p float64) float64 {
	if p >= d.threshold {
		return p - d.cut
	}
	return p
}

// MemberDiscount 会员折：按比例打折（rate 0.9 表示九折）。
type MemberDiscount struct {
	rate float64
}

func (d *MemberDiscount) Name() string { return fmt.Sprintf("会员%.0f折", d.rate*10) }

func (d *MemberDiscount) Apply(p float64) float64 { return p * d.rate }

// =============================================================================
// 四、main：同一订单，运行时切换三种策略
// =============================================================================

func main() {
	// 同一笔订单：先按无折扣结算，再运行时切成满减、会员折各算一次。
	o := NewOrder("A100", &NoDiscount{})
	o.Settle(320) // 原价 320

	o.SetDiscount(&FullReduction{threshold: 300, cut: 50})
	o.Settle(320) // 满 300 减 50 → 270

	o.SetDiscount(&MemberDiscount{rate: 0.9})
	o.Settle(320) // 九折 → 288

	fmt.Println()
	fmt.Println("========== 读懂输出后你会发现 ==========")
	fmt.Println("1. Order 只调用 Discount.Apply，不关心是哪一种计价规则")
	fmt.Println("2. 换优惠 = 换一个策略对象，Order 的结算逻辑一行没改")
	fmt.Println("3. 新增一种优惠 = 新写一个实现 Discount 的类型")
	fmt.Println("4. 策略可以带参数（满减门槛、折扣率），构造时注入")
}
