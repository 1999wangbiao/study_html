// 工厂方法模式可运行示范（Go）—— 物流发货
//
// 核心一句话：「创建哪个具体产品」交给具体工厂；客户端只认抽象。
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、抽象产品：运输工具
// =============================================================================

// Transport 运输工具抽象。
type Transport interface {
	Deliver()
}

// =============================================================================
// 二、具体产品
// =============================================================================

// Truck 陆运卡车。
type Truck struct{}

// Deliver 陆路配送。
func (t *Truck) Deliver() {
	fmt.Println("  [Truck] 走高速公路送货")
}

// Ship 海运货轮。
type Ship struct{}

// Deliver 海路配送。
func (s *Ship) Deliver() {
	fmt.Println("  [Ship] 走航线海运送货")
}

// =============================================================================
// 三、抽象工厂 + 具体工厂（工厂方法）
// =============================================================================

// LogisticsFactory 物流工厂：工厂方法 Create 返回抽象产品。
type LogisticsFactory interface {
	Create() Transport
}

// TruckFactory 陆运工厂。
type TruckFactory struct{}

// Create 创建卡车。
func (f *TruckFactory) Create() Transport { return &Truck{} }

// ShipFactory 海运工厂。
type ShipFactory struct{}

// Create 创建货轮。
func (f *ShipFactory) Create() Transport { return &Ship{} }

// =============================================================================
// 四、客户端：只依赖抽象，换工厂即换运输方式
// =============================================================================

// PlanDelivery 用给定工厂规划并执行一次配送。
func PlanDelivery(factory LogisticsFactory) {
	t := factory.Create()
	t.Deliver()
}

func main() {
	fmt.Println("========== 陆运 ==========")
	PlanDelivery(&TruckFactory{})

	fmt.Println("========== 海运 ==========")
	PlanDelivery(&ShipFactory{})
}
