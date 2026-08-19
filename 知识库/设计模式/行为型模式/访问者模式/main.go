// 访问者模式可运行示范（Go）—— 最终形态
//
// 若还没跟上「为什么要访问者」，请先按顺序跑演变目录：
//
//	evolve/step1_方法堆在类型上   → 操作写在图形上，加操作就改所有类型
//	evolve/step2_type_switch抽出去 → 操作抽出去，但 switch 重复且易漏
//	evolve/step3_访问者           → Accept + VisitXxx（本文件同思路，日志更细）
//	evolve/step4_只加新操作       → 只加 PerimeterVisitor，图形不动
//
// 本目录运行：
//
//	go run .
package main

import (
	"fmt"
	"math"
)

// =============================================================================
// 一、角色 1：Visitor（访问者）——「要对图形做什么」
// =============================================================================
//
// 注意：每增加一种图形类型，这里就要多一个 VisitXxx。
// 这是访问者模式的代价：元素集合最好相对稳定。

// Visitor 为每一种具体图形声明一个访问方法。
type Visitor interface {
	VisitCircle(c *Circle)
	VisitRectangle(r *Rectangle)
}

// =============================================================================
// 二、角色 2：Element（元素）——「可以被访问的东西」
// =============================================================================

// Shape 所有图形都要实现 Accept：把自己交给 Visitor。
type Shape interface {
	Accept(v Visitor)
}

// =============================================================================
// 三、具体元素：圆、矩形
// =============================================================================
//
// 它们只保存自己的数据，以及一个很薄的 Accept。
// 业务（算面积、打印）不写在这里，写在 Visitor 里。

// Circle 圆形。
type Circle struct {
	Name   string
	Radius float64
}

// Accept 是双重分派的关键步骤：
//
//  分派 1：调用方写的是 shape.Accept(v)，运行时根据 shape 真实类型
//         进入 *Circle.Accept 或 *Rectangle.Accept。
//  分派 2：在本方法里显式调用 v.VisitCircle(c)，根据 v 的真实类型
//         进入 AreaVisitor.VisitCircle 或 InfoVisitor.VisitCircle。
//
// Go 没有方法重载，所以「第二次分派」靠我们手写 VisitCircle / VisitRectangle 完成。
func (c *Circle) Accept(v Visitor) {
	fmt.Printf("  [Accept] 我是圆 %q，现在把我交给 Visitor → 调用 VisitCircle\n", c.Name)
	v.VisitCircle(c) // 把自己传进去，Visitor 才能读 Radius
}

// Rectangle 矩形。
type Rectangle struct {
	Name   string
	Width  float64
	Height float64
}

// Accept 同理：矩形只负责声明「我是矩形」，并调用 VisitRectangle。
func (r *Rectangle) Accept(v Visitor) {
	fmt.Printf("  [Accept] 我是矩形 %q，现在把我交给 Visitor → 调用 VisitRectangle\n", r.Name)
	v.VisitRectangle(r)
}

// =============================================================================
// 四、具体访问者 A：算面积
// =============================================================================
//
// 新增一种「操作」= 新增一个实现 Visitor 的类型。
// Circle / Rectangle 的代码一行都不用改。

// AreaVisitor 遍历图形时累加面积。
type AreaVisitor struct {
	Total float64 // 访问过程中不断累加
}

func (v *AreaVisitor) VisitCircle(c *Circle) {
	area := math.Pi * c.Radius * c.Radius
	v.Total += area
	fmt.Printf("  [VisitCircle/面积] %s: π×%.1f² = %.2f，累计=%.2f\n",
		c.Name, c.Radius, area, v.Total)
}

func (v *AreaVisitor) VisitRectangle(r *Rectangle) {
	area := r.Width * r.Height
	v.Total += area
	fmt.Printf("  [VisitRectangle/面积] %s: %.1f×%.1f = %.2f，累计=%.2f\n",
		r.Name, r.Width, r.Height, area, v.Total)
}

// =============================================================================
// 五、具体访问者 B：打印描述（另一种完全不同的操作）
// =============================================================================
//
// 重点对比：同样走过 Circle/Rectangle，干的事完全不同，
// 但图形类型本身没有增加任何方法。

// InfoVisitor 收集每张图形的文字描述。
type InfoVisitor struct {
	Lines []string
}

func (v *InfoVisitor) VisitCircle(c *Circle) {
	line := fmt.Sprintf("圆(%s) 半径=%.1f", c.Name, c.Radius)
	v.Lines = append(v.Lines, line)
	fmt.Printf("  [VisitCircle/信息] 记下: %s\n", line)
}

func (v *InfoVisitor) VisitRectangle(r *Rectangle) {
	line := fmt.Sprintf("矩形(%s) 宽=%.1f 高=%.1f", r.Name, r.Width, r.Height)
	v.Lines = append(v.Lines, line)
	fmt.Printf("  [VisitRectangle/信息] 记下: %s\n", line)
}

// =============================================================================
// 六、对象结构：一堆图形（稳定结构）
// =============================================================================

// 对列表中每个元素调用 Accept，相当于「请这位 Visitor 挨个访问」。
func acceptAll(shapes []Shape, v Visitor) {
	for i, s := range shapes {
		fmt.Printf("→ 处理第 %d 个 Shape（接口变量，运行时类型见 Accept 日志）\n", i+1)
		s.Accept(v)
	}
}

// =============================================================================
// 七、main：组装数据，换不同 Visitor 跑两遍
// =============================================================================

func main() {
	// 对象结构：种类固定（目前只有圆、矩形），以后也可能只偶尔加新图形。
	shapes := []Shape{
		&Circle{Name: "小圆", Radius: 2},
		&Rectangle{Name: "大门", Width: 3, Height: 4},
		&Circle{Name: "大圆", Radius: 5},
	}

	fmt.Println("========== 第一趟：请「面积访问者」进来 ==========")
	area := &AreaVisitor{}
	acceptAll(shapes, area)
	fmt.Printf("结果：总面积 = %.2f\n\n", area.Total)

	fmt.Println("========== 第二趟：请「信息访问者」进来 ==========")
	info := &InfoVisitor{}
	acceptAll(shapes, info)
	fmt.Println("结果：描述列表")
	for _, line := range info.Lines {
		fmt.Println("  -", line)
	}

	fmt.Println()
	fmt.Println("========== 读懂输出后你会发现 ==========")
	fmt.Println("1. 每张图都先走自己的 Accept（第一次按「元素类型」分派）")
	fmt.Println("2. Accept 里再调 VisitXxx（第二次按「访问者类型+元素类型」分派）")
	fmt.Println("3. 换 Visitor = 换操作；Circle/Rectangle 源码不用动")
	fmt.Println("4. 若将来加 Triangle，要改 Visitor 接口，并给 Area/Info 都补 VisitTriangle")
}
