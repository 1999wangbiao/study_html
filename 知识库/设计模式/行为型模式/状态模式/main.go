// 状态模式可运行示范（Go）—— 订单状态机
//
// 核心一句话：把「随状态变化的行为」拆到各个 State 类型里，
// Context（Order）只负责持有当前状态并转发请求。
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、角色 1：State（状态）——「当前状态下，每个动作怎么响应」
// =============================================================================

// OrderState 订单在某一状态下对 Pay / Ship / Cancel 的行为约定。
// 合法转换在具体状态里完成；非法则打印拒绝日志，不 panic。
type OrderState interface {
	Pay(o *Order)
	Ship(o *Order)
	Cancel(o *Order)
	Name() string
}

// =============================================================================
// 二、角色 2：Context（上下文）——「订单本体」
// =============================================================================
//
// Order 对外暴露业务动作；真正怎么响应，全部委托给当前 state。

// Order 订单上下文：持有当前状态，并提供切换入口。
type Order struct {
	id    string
	state OrderState
}

// NewOrder 创建待支付订单。
func NewOrder(id string) *Order {
	o := &Order{id: id}
	o.setState(&PendingState{})
	return o
}

// setState 切换当前状态（供 ConcreteState 在合法转换时调用）。
func (o *Order) setState(s OrderState) {
	o.state = s
	fmt.Printf("[Order %s] 当前状态 → %s\n", o.id, s.Name())
}

// Pay 支付：转发给当前状态。
func (o *Order) Pay() { o.state.Pay(o) }

// Ship 发货：转发给当前状态。
func (o *Order) Ship() { o.state.Ship(o) }

// Cancel 取消：转发给当前状态。
func (o *Order) Cancel() { o.state.Cancel(o) }

// =============================================================================
// 三、具体状态：Pending / Paid / Shipped / Cancelled
// =============================================================================

// PendingState 待支付。
type PendingState struct{}

func (s *PendingState) Name() string { return "Pending(待支付)" }

func (s *PendingState) Pay(o *Order) {
	fmt.Printf("  [Pay] 订单 %s 支付成功\n", o.id)
	o.setState(&PaidState{})
}

func (s *PendingState) Ship(o *Order) {
	fmt.Printf("  [Ship] 拒绝：订单 %s 仍待支付，不能发货\n", o.id)
}

func (s *PendingState) Cancel(o *Order) {
	fmt.Printf("  [Cancel] 订单 %s 已取消（未支付）\n", o.id)
	o.setState(&CancelledState{})
}

// PaidState 已支付。
type PaidState struct{}

func (s *PaidState) Name() string { return "Paid(已支付)" }

func (s *PaidState) Pay(o *Order) {
	fmt.Printf("  [Pay] 拒绝：订单 %s 已支付，勿重复付款\n", o.id)
}

func (s *PaidState) Ship(o *Order) {
	fmt.Printf("  [Ship] 订单 %s 已发货\n", o.id)
	o.setState(&ShippedState{})
}

func (s *PaidState) Cancel(o *Order) {
	fmt.Printf("  [Cancel] 订单 %s 已取消（退款流程略）\n", o.id)
	o.setState(&CancelledState{})
}

// ShippedState 已发货（终态）。
type ShippedState struct{}

func (s *ShippedState) Name() string { return "Shipped(已发货)" }

func (s *ShippedState) Pay(o *Order) {
	fmt.Printf("  [Pay] 拒绝：订单 %s 已发货\n", o.id)
}

func (s *ShippedState) Ship(o *Order) {
	fmt.Printf("  [Ship] 拒绝：订单 %s 已发货\n", o.id)
}

func (s *ShippedState) Cancel(o *Order) {
	fmt.Printf("  [Cancel] 拒绝：订单 %s 已发货，不能取消\n", o.id)
}

// CancelledState 已取消（终态）。
type CancelledState struct{}

func (s *CancelledState) Name() string { return "Cancelled(已取消)" }

func (s *CancelledState) Pay(o *Order) {
	fmt.Printf("  [Pay] 拒绝：订单 %s 已取消\n", o.id)
}

func (s *CancelledState) Ship(o *Order) {
	fmt.Printf("  [Ship] 拒绝：订单 %s 已取消\n", o.id)
}

func (s *CancelledState) Cancel(o *Order) {
	fmt.Printf("  [Cancel] 拒绝：订单 %s 已取消\n", o.id)
}

// =============================================================================
// 四、main：两条演示路径
// =============================================================================

func main() {
	fmt.Println("========== 路径 1：支付 → 发货 → 再试取消（应拒绝）==========")
	o1 := NewOrder("A100")
	o1.Pay()
	o1.Ship()
	o1.Cancel() // 已发货，拒绝

	fmt.Println()
	fmt.Println("========== 路径 2：待支付直接取消 → 再试支付（应拒绝）==========")
	o2 := NewOrder("B200")
	o2.Cancel()
	o2.Pay() // 已取消，拒绝

	fmt.Println()
	fmt.Println("========== 读懂输出后你会发现 ==========")
	fmt.Println("1. Order 不写 if 状态分支，只转发到当前 State")
	fmt.Println("2. 合法转换由具体 State 调用 setState 完成")
	fmt.Println("3. 非法动作留在对应 State 里拒绝，互不影响")
	fmt.Println("4. 新增状态 = 新类型实现 OrderState；改某状态行为只动那一个类型")
}
