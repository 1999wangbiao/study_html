// 演变第 1 步：最直觉的写法 —— 把操作直接写成图形的方法
//
// 运行：
//
//	cd evolve/step1_方法堆在类型上
//	go run .
//
// 场景：有圆、矩形，要「算面积」和「打印信息」。
// 写法：每个图形自己实现 Area()、Info()。
//
// 问题马上出现：
//  - 每加一种操作（比如「画到屏幕」「导出 SVG」），Circle 和 Rectangle 都要改
//  - 图形类型被业务方法越堆越胖，和「只表示形状数据」混在一起
package main

import (
	"fmt"
	"math"
)

type Circle struct {
	Name   string
	Radius float64
}

func (c *Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c *Circle) Info() string {
	return fmt.Sprintf("圆(%s) 半径=%.1f", c.Name, c.Radius)
}

type Rectangle struct {
	Name   string
	Width  float64
	Height float64
}

func (r *Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r *Rectangle) Info() string {
	return fmt.Sprintf("矩形(%s) 宽=%.1f 高=%.1f", r.Name, r.Width, r.Height)
}

// 想统一遍历时会发现：Circle 和 Rectangle 没有共同接口装「所有操作」。
// 这里先分别处理，体会「结构不统一」的别扭。
func main() {
	c1 := &Circle{Name: "小圆", Radius: 2}
	r1 := &Rectangle{Name: "大门", Width: 3, Height: 4}
	c2 := &Circle{Name: "大圆", Radius: 5}

	fmt.Println("========== 第 1 步：方法堆在类型上 ==========")
	fmt.Println("总面积:", c1.Area()+r1.Area()+c2.Area())
	fmt.Println("描述:")
	fmt.Println(" -", c1.Info())
	fmt.Println(" -", r1.Info())
	fmt.Println(" -", c2.Info())

	fmt.Println()
	fmt.Println("痛点：再加「导出 SVG」→ Circle、Rectangle 都要再加方法。")
	fmt.Println("下一站 step2：把操作抽到外面，用 type switch 统一处理。")
}
