// 演变第 4 步：在访问者之上「只加新操作」—— 证明图形类型可以不动
//
// 运行：
//
//	cd evolve/step4_只加新操作
//	go run .
//
// 相对 step3，本文件多了一个 PerimeterVisitor（算周长）。
// 请注意：Circle / Rectangle / Shape / Visitor 接口的「元素侧」用法不变；
// 我们只是多实现了一个 Visitor。
//
// （若要加 Triangle，才需要改 Visitor 接口——那是另一条演变线，刻意没做。）
package main

import (
	"fmt"
	"math"
)

type Visitor interface {
	VisitCircle(c *Circle)
	VisitRectangle(r *Rectangle)
}

type Shape interface {
	Accept(v Visitor)
}

type Circle struct {
	Name   string
	Radius float64
}

func (c *Circle) Accept(v Visitor) { v.VisitCircle(c) }

type Rectangle struct {
	Name   string
	Width  float64
	Height float64
}

func (r *Rectangle) Accept(v Visitor) { v.VisitRectangle(r) }

// ----- 旧操作：面积（step3 已有）-----

type AreaVisitor struct{ Total float64 }

func (v *AreaVisitor) VisitCircle(c *Circle) {
	v.Total += math.Pi * c.Radius * c.Radius
}

func (v *AreaVisitor) VisitRectangle(r *Rectangle) {
	v.Total += r.Width * r.Height
}

// ----- 新操作：周长（本次演变新增；未改 Circle/Rectangle 源码结构）-----

type PerimeterVisitor struct{ Total float64 }

func (v *PerimeterVisitor) VisitCircle(c *Circle) {
	p := 2 * math.Pi * c.Radius
	v.Total += p
	fmt.Printf("  [周长] 圆 %s → %.2f\n", c.Name, p)
}

func (v *PerimeterVisitor) VisitRectangle(r *Rectangle) {
	p := 2 * (r.Width + r.Height)
	v.Total += p
	fmt.Printf("  [周长] 矩形 %s → %.2f\n", r.Name, p)
}

func acceptAll(shapes []Shape, v Visitor) {
	for _, s := range shapes {
		s.Accept(v)
	}
}

func main() {
	shapes := []Shape{
		&Circle{Name: "小圆", Radius: 2},
		&Rectangle{Name: "大门", Width: 3, Height: 4},
		&Circle{Name: "大圆", Radius: 5},
	}

	fmt.Println("========== 第 4 步：只加新操作（周长） ==========")

	area := &AreaVisitor{}
	acceptAll(shapes, area)
	fmt.Printf("总面积: %.2f\n\n", area.Total)

	fmt.Println("--- 新来的周长访问者 ---")
	peri := &PerimeterVisitor{}
	acceptAll(shapes, peri)
	fmt.Printf("总周长: %.2f\n\n", peri.Total)

	fmt.Println("看懂了吗？")
	fmt.Println("  数据（圆/矩形）稳定 → 用访问者往外挂操作最划算。")
	fmt.Println("  若天天加新图形种类 → 别用访问者，回到接口方法或策略更合适。")
}
