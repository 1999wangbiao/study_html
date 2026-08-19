// 演变第 3 步：访问者模式 —— 把 step2 的 type switch「升级」成双重分派
//
// 运行：
//
//	cd evolve/step3_访问者
//	go run .
//
// =============================================================================
// 先建立心智模型（读代码前看完这段）
// =============================================================================
//
// 访问者把世界拆成两边：
//
//	左边 = 元素（Element / Shape）
//	  圆、矩形……「有哪些东西」，种类相对稳定。
//	  只负责保存数据 + 提供 Accept。
//
//	右边 = 访问者（Visitor）
//	  算面积、收集信息……「要对东西做什么」，操作经常增加。
//	  每一种操作做成一个实现 Visitor 的类型。
//
// 调用约定只有一句：
//
//	元素.Accept(某个访问者)
//	  → 元素在 Accept 里回调 访问者.VisitXxx(自己)
//	  → 真正干活的代码在 VisitXxx 里
//
// =============================================================================
// 和 step2 差在哪？（这是本步唯一要搞懂的点）
// =============================================================================
//
//	step2：调用方自己写 switch s.(type) { case *Circle: ... }
//	       「你是谁」由外面的函数判断。
//
//	step3：调用方只写 s.Accept(v)
//	       「我是圆」由 *Circle.Accept 自己声明，再调 v.VisitCircle(c)。
//
// 两次「按类型选实现」叠在一起，就叫双重分派（Double Dispatch）：
//
//	分派 1：s.Accept(v)
//	  s 的动态类型是 *Circle 还是 *Rectangle？
//	  → 进入对应的 Accept 方法。
//
//	分派 2：v.VisitCircle(c) 或 v.VisitRectangle(r)
//	  v 的动态类型是 *AreaVisitor 还是 *InfoVisitor？
//	  → 进入对应 Visitor 里、针对该图形的那份逻辑。
//
// Go 没有方法重载，所以第二次分派不能靠「同名 Visit 重载」，
// 只能靠 VisitCircle / VisitRectangle 两个不同方法名来区分。
//
// =============================================================================
// 好处 / 代价
// =============================================================================
//
// 好处：
//  - 新操作 = 新建一个 Visitor 实现，Circle / Rectangle 一行不用改
//  - 实现 Visitor 接口时，漏写某个 VisitXxx → 编译直接报错（比 step2 漏 case 安全）
//  - 同一种操作的所有分支聚在一个类型里，比散落多个 switch 函数好找
//
// 代价：
//  - 新加一种图形（Triangle）→ 改 Visitor 接口 + 改掉所有 Visitor 实现
//  - 因此适合：「元素种类稳、操作常增」；反过来就别用
package main

import (
	"fmt"
	"math"
)

// =============================================================================
// 一、Visitor 接口 —— 「一种操作要对每一种元素做什么」的清单
// =============================================================================
//
// 读法：凡是想当访问者的类型，必须同时会处理圆和矩形。
// 以后若增加 Triangle，这里就要多一行 VisitTriangle，
// 于是所有旧 Visitor 都会编译失败，逼你补全 —— 这是刻意设计。

type Visitor interface {
	// VisitCircle：遇到圆时怎么处理（面积？信息？周长？由具体 Visitor 决定）
	VisitCircle(c *Circle)

	// VisitRectangle：遇到矩形时怎么处理
	VisitRectangle(r *Rectangle)
}

// =============================================================================
// 二、Shape 接口 —— 「能被访问」的元素约定
// =============================================================================
//
// 元素侧只暴露一个口子：Accept。
// 调用方不直接调 VisitXxx，而是把 Visitor「递进去」，由元素回调回来。
// 这样元素能保证：回调时用的是「正确的 Visit 方法名」。

type Shape interface {
	// Accept：接纳一位访问者，并在内部调用对方对应的 VisitXxx。
	Accept(v Visitor)
}

// =============================================================================
// 三、具体元素：Circle / Rectangle —— 只存数据 + 很薄的 Accept
// =============================================================================
//
// 注意：这里没有 Area()、Info()。
// 业务逻辑全部搬到 Visitor 里，图形类型保持「瘦」。

// Circle 圆形元素：名字 + 半径。
type Circle struct {
	Name   string  // 仅用于打印演示，方便对照日志
	Radius float64 // 算面积时 Visitor 会读这个字段
}

// Accept 是双重分派发生的现场，建议对着运行日志看：
//
//	调用方：s.Accept(area)          // s 实际是 *Circle
//	  ↓ 分派 1：进入本方法 (*Circle).Accept
//	本方法：v.VisitCircle(c)        // v 实际是 *AreaVisitor
//	  ↓ 分派 2：进入 AreaVisitor.VisitCircle
//
// 参数 c（也就是接收者）必须传给 Visitor，
// 否则 Visitor 拿不到 Radius，算不了面积。
func (c *Circle) Accept(v Visitor) {
	fmt.Printf("  [Accept→VisitCircle] %s\n", c.Name)

	// 关键一行：不是 v.Visit(c)，而是明确叫 VisitCircle。
	// 因为 Go 无法靠参数类型重载选出不同 Visit，必须方法名不同。
	v.VisitCircle(c)
}

// Rectangle 矩形元素：名字 + 宽高。
type Rectangle struct {
	Name   string
	Width  float64
	Height float64
}

// Accept 与圆对称：声明「我是矩形」，并回调 VisitRectangle。
//
// 若这里误写成 v.VisitCircle(...)，编译期就能发现类型不对；
// 这比 step2 在巨大 switch 里抄错 case 更安全。
func (r *Rectangle) Accept(v Visitor) {
	fmt.Printf("  [Accept→VisitRectangle] %s\n", r.Name)
	v.VisitRectangle(r)
}

// =============================================================================
// 四、具体访问者 A：AreaVisitor —— 「算面积」这一种操作
// =============================================================================
//
// 对照 step2：以前是一个 totalArea(shapes) 函数，里面一大段 switch。
// 现在：switch 的每个 case 变成接口上的一个方法，聚在这个类型里。
//
// Total 是访问过程中的「累加器」：
// 每 Visit 一次就改一点，全部 Accept 完后从 Total 读结果。

type AreaVisitor struct {
	Total float64 // 遍历前为 0，遍历中累加，遍历后就是总面积
}

// VisitCircle：面积操作遇上圆 —— πr²
//
// 能进到这个方法，说明前面已经走过：
//
//	某 Shape.Accept(area) → (*Circle).Accept → area.VisitCircle
func (v *AreaVisitor) VisitCircle(c *Circle) {
	area := math.Pi * c.Radius * c.Radius
	v.Total += area
	fmt.Printf("  [面积] %s → %.2f，累计=%.2f\n", c.Name, area, v.Total)
}

// VisitRectangle：面积操作遇上矩形 —— 宽×高
func (v *AreaVisitor) VisitRectangle(r *Rectangle) {
	area := r.Width * r.Height
	v.Total += area
	fmt.Printf("  [面积] %s → %.2f，累计=%.2f\n", r.Name, area, v.Total)
}

// =============================================================================
// 五、具体访问者 B：InfoVisitor —— 「收集文字描述」另一种操作
// =============================================================================
//
// 和 AreaVisitor 走完全相同的 Accept 路径，但 VisitXxx 里干的事不同。
// 这就是访问者的卖点：换 Visitor = 换算法，元素列表和 Accept 代码不用动。

type InfoVisitor struct {
	Lines []string // 把每张图的描述存起来，最后一次性打印
}

// VisitCircle：信息操作遇上圆 —— 拼一句人类可读的话
func (v *InfoVisitor) VisitCircle(c *Circle) {
	line := fmt.Sprintf("圆(%s) 半径=%.1f", c.Name, c.Radius)
	v.Lines = append(v.Lines, line)
	fmt.Printf("  [信息] %s\n", line)
}

// VisitRectangle：信息操作遇上矩形
func (v *InfoVisitor) VisitRectangle(r *Rectangle) {
	line := fmt.Sprintf("矩形(%s) 宽=%.1f 高=%.1f", r.Name, r.Width, r.Height)
	v.Lines = append(v.Lines, line)
	fmt.Printf("  [信息] %s\n", line)
}

// =============================================================================
// 六、对象结构上的遍历 —— 对每个元素调用 Accept
// =============================================================================
//
// 访问者模式里常把「一堆元素」叫做 Object Structure。
// 这里用最简单的切片模拟：不关心树/图，只演示「挨个访问」。
//
// 注意签名：shapes 是 []Shape，v 是 Visitor。
// 调用方既不知道元素具体是圆还是矩形，也不在这里写 switch；
// 类型分辨全部交给 Accept + VisitXxx。

func acceptAll(shapes []Shape, v Visitor) {
	for _, s := range shapes {
		// s 的静态类型是 Shape，动态类型可能是 *Circle 或 *Rectangle。
		// 这一行触发分派 1；分派 2 发生在各自的 Accept 内部。
		s.Accept(v)
	}
}

// =============================================================================
// 七、main：同一批图形，请两位不同的 Visitor 各走一趟
// =============================================================================
//
// 建议运行后盯日志顺序，以「小圆 + AreaVisitor」为例应是：
//
//	[Accept→VisitCircle] 小圆          ← 分派 1 结束，进入圆的 Accept
//	[面积] 小圆 → 12.57，累计=...      ← 分派 2 结束，进入 AreaVisitor.VisitCircle
//
// 第二趟换成 InfoVisitor 时，Accept 日志一样，Visit 日志变成 [信息]。

func main() {
	// 对象结构：元素种类固定为圆 / 矩形（相对稳定的那一边）
	shapes := []Shape{
		&Circle{Name: "小圆", Radius: 2},
		&Rectangle{Name: "大门", Width: 3, Height: 4},
		&Circle{Name: "大圆", Radius: 5},
	}

	fmt.Println("========== 第 3 步：访问者 ==========")

	// ----- 第一趟：面积访问者 -----
	fmt.Println("--- 算面积 ---")
	area := &AreaVisitor{} // Total 零值为 0
	acceptAll(shapes, area)
	// 三张图的 Visit* 都跑完后，累加结果在 area.Total
	fmt.Printf("总面积: %.2f\n\n", area.Total)

	// ----- 第二趟：信息访问者（元素一个没改，只换了 v）-----
	fmt.Println("--- 收集信息 ---")
	info := &InfoVisitor{}
	acceptAll(shapes, info)
	for _, line := range info.Lines {
		fmt.Println(" -", line)
	}

	fmt.Println()
	fmt.Println("演变小结：")
	fmt.Println("  step1 方法堆类型上 → 加操作要改所有图形")
	fmt.Println("  step2 type switch  → 加操作方便，但 switch 重复，加图形易漏")
	fmt.Println("  step3 访问者       → 加操作=新 Visitor；加图形编译器逼你改齐 VisitXxx")
	fmt.Println()
	fmt.Println("下一步：evolve/step4_只加新操作（再挂一个周长 Visitor，图形仍不动）。")
}
