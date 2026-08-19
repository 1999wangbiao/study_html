// 演变第 2 步：操作抽到外面 —— 用 type switch 按类型办事
//
// 运行：
//
//	cd evolve/step2_type_switch抽出去
//	go run .
//
// 进步：
//  - Circle / Rectangle 只剩数据，干净多了
//  - 可以 []any 或空接口列表统一遍历
//  - 新操作 = 新写一个 totalArea / collectInfo 函数，不必改图形类型
//
// 新痛点：
//  - 每种操作里都要写一遍完整的 type switch
//  - 一加 Triangle，所有 switch 都要改（编译器还不一定提醒你漏了哪个函数）
//  - 操作一多，switch 复制粘贴，容易漏分支
package main

import (
	"fmt"
	"math"
)

// 图形只保留数据，不再背业务方法。
type Circle struct {
	Name   string
	Radius float64
}

type Rectangle struct {
	Name   string
	Width  float64
	Height float64
}

// totalArea：操作 A —— 算总面积（逻辑在图形外面）
func totalArea(shapes []any) float64 {
	var sum float64
	for _, s := range shapes {
		// 第一次「按类型分派」靠手写 switch
		switch x := s.(type) {
		case *Circle:
			sum += math.Pi * x.Radius * x.Radius
			fmt.Printf("  [面积] 圆 %s → %.2f\n", x.Name, math.Pi*x.Radius*x.Radius)
		case *Rectangle:
			sum += x.Width * x.Height
			fmt.Printf("  [面积] 矩形 %s → %.2f\n", x.Name, x.Width*x.Height)
			// 将来加 *Triangle 时：这里必须记得补，否则静默跳过！
		default:
			fmt.Printf("  [面积] 未知类型 %T，跳过\n", s)
		}
	}
	return sum
}

// collectInfo：操作 B —— 又是一整份几乎同样结构的 switch
func collectInfo(shapes []any) []string {
	var lines []string
	for _, s := range shapes {
		switch x := s.(type) {
		case *Circle:
			line := fmt.Sprintf("圆(%s) 半径=%.1f", x.Name, x.Radius)
			lines = append(lines, line)
			fmt.Printf("  [信息] %s\n", line)
		case *Rectangle:
			line := fmt.Sprintf("矩形(%s) 宽=%.1f 高=%.1f", x.Name, x.Width, x.Height)
			lines = append(lines, line)
			fmt.Printf("  [信息] %s\n", line)
		default:
			fmt.Printf("  [信息] 未知类型 %T，跳过\n", s)
		}
	}
	return lines
}

func main() {
	shapes := []any{
		&Circle{Name: "小圆", Radius: 2},
		&Rectangle{Name: "大门", Width: 3, Height: 4},
		&Circle{Name: "大圆", Radius: 5},
	}

	fmt.Println("========== 第 2 步：type switch 把操作抽出去 ==========")
	fmt.Println("--- 算面积 ---")
	fmt.Printf("总面积: %.2f\n\n", totalArea(shapes))

	fmt.Println("--- 收集信息 ---")
	for _, line := range collectInfo(shapes) {
		fmt.Println(" -", line)
	}

	fmt.Println()
	fmt.Println("进步：图形类型不用为每个新操作改代码。")
	fmt.Println("痛点：每种操作复制一份 switch；加新图形要改遍所有函数，还容易漏。")
	fmt.Println("下一站 step3：用 Visitor 把「按类型分派」固化成接口，漏实现会编译报错。")
}
