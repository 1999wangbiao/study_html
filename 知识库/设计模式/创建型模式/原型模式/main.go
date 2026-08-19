// 原型模式可运行示范（Go）—— 图形克隆 + 浅拷 / 深拷对照
//
// 核心一句话：用已有对象当模板 Clone 一份再改；深拷还是浅拷由具体类型自己负责。
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、原型接口
// =============================================================================

// Shape 可克隆的图形。
type Shape interface {
	Clone() Shape
	Move(dx, dy int)
	String() string
}

// =============================================================================
// 二、只有值字段：结构体拷贝就够（没有「共享底层」问题）
// =============================================================================

// Circle 圆形（字段全是值类型）。
type Circle struct {
	Color  string
	Radius int
	X, Y   int
}

// Clone 拷贝圆形；无切片 / 指针，*c 赋值即独立副本。
func (c *Circle) Clone() Shape {
	copy := *c
	return &copy
}

// Move 移动圆心。
func (c *Circle) Move(dx, dy int) {
	c.X += dx
	c.Y += dy
}

// String 打印圆形摘要。
func (c *Circle) String() string {
	return fmt.Sprintf("Circle{color=%s, r=%d, pos=(%d,%d)}", c.Color, c.Radius, c.X, c.Y)
}

// =============================================================================
// 三、含切片字段：浅拷 vs 深拷（对照重点）
// =============================================================================

// Rectangle 矩形；Tags 是切片，决定浅拷还是深拷会不一样。
type Rectangle struct {
	Color  string
	Width  int
	Height int
	X, Y   int
	Tags   []string
}

// CloneShallow 浅拷贝：结构体字段复制，但 Tags 仍指向同一底层数组。
func (r *Rectangle) CloneShallow() *Rectangle {
	copy := *r // Tags 的指针头 / len / cap 被复制，底层数组共享
	return &copy
}

// Clone 深拷贝：Tags 另起一份底层数组（实现 Shape 接口时用这个）。
func (r *Rectangle) Clone() Shape {
	copy := *r
	if r.Tags != nil {
		copy.Tags = append([]string(nil), r.Tags...)
	}
	return &copy
}

// Move 移动矩形左上角。
func (r *Rectangle) Move(dx, dy int) {
	r.X += dx
	r.Y += dy
}

// String 打印矩形摘要。
func (r *Rectangle) String() string {
	return fmt.Sprintf("Rectangle{color=%s, %dx%d, pos=(%d,%d), tags=%v}",
		r.Color, r.Width, r.Height, r.X, r.Y, r.Tags)
}

// =============================================================================
// 四、客户端
// =============================================================================

// Duplicate 多态克隆（走深拷 Clone）。
func Duplicate(s Shape) Shape {
	return s.Clone()
}

func main() {
	fmt.Println("========== 1. 纯值字段：Clone 后互不影响 ==========")
	c1 := &Circle{Color: "红", Radius: 10, X: 0, Y: 0}
	c2 := Duplicate(c1).(*Circle)
	c2.Move(5, 8)
	c2.Color = "蓝"
	fmt.Printf("  原件: %s\n", c1)
	fmt.Printf("  副本: %s\n", c2)

	fmt.Println("========== 2. 浅拷贝：改副本 Tags，原件也被改 ==========")
	shallowSrc := &Rectangle{
		Color: "绿", Width: 20, Height: 15, X: 1, Y: 2,
		Tags: []string{"选中", "图层1"},
	}
	shallowCopy := shallowSrc.CloneShallow()
	fmt.Printf("  改前 原件.Tags 底层: %p\n", shallowSrc.Tags)
	fmt.Printf("  改前 副本.Tags 底层: %p  （同一块）\n", shallowCopy.Tags)

	shallowCopy.Tags[0] = "未选中" // 改的是共享数组
	shallowCopy.X = 99            // 值字段各自一份，只动副本
	fmt.Printf("  改后 原件: %s  ← Tags 被连坐！\n", shallowSrc)
	fmt.Printf("  改后 副本: %s\n", shallowCopy)

	fmt.Println("========== 3. 深拷贝：改副本 Tags，原件不动 ==========")
	deepSrc := &Rectangle{
		Color: "绿", Width: 20, Height: 15, X: 1, Y: 2,
		Tags: []string{"选中", "图层1"},
	}
	deepCopy := Duplicate(deepSrc).(*Rectangle)
	fmt.Printf("  改前 原件.Tags 底层: %p\n", deepSrc.Tags)
	fmt.Printf("  改前 副本.Tags 底层: %p  （另一块）\n", deepCopy.Tags)

	deepCopy.Tags[0] = "未选中"
	deepCopy.Move(3, 4)
	fmt.Printf("  改后 原件: %s  ← Tags 仍是「选中」\n", deepSrc)
	fmt.Printf("  改后 副本: %s\n", deepCopy)
}
