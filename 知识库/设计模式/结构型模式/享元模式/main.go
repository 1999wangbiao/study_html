// 享元模式可运行示范（Go）—— 棋盘棋子
//
// 核心一句话：黑白外观（内在状态）只各存一份；坐标（外在状态）放在棋盘侧。
// 棋子再多，PieceType 缓存里仍然只有 2 个对象。
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、Flyweight：棋子类型（内在状态）
// =============================================================================

// Color 棋子颜色。
type Color string

const (
	Black Color = "black"
	White Color = "white"
)

// PieceType 享元：可共享的外观信息（与落子位置无关）。
type PieceType struct {
	color  Color
	symbol string // 打印用符号
}

// Color 返回颜色。
func (t *PieceType) Color() Color { return t.color }

// Symbol 返回显示符号。
func (t *PieceType) Symbol() string { return t.symbol }

// Draw 绘制：外在状态（坐标）由调用方传入，不存在享元内部。
func (t *PieceType) Draw(row, col int) {
	fmt.Printf("  在 (%d,%d) 落子 %s(%s)\n", row, col, t.symbol, t.color)
}

// =============================================================================
// 二、FlyweightFactory：按颜色缓存享元
// =============================================================================

// PieceFactory 棋子类型工厂：同色只创建一次。
type PieceFactory struct {
	cache map[Color]*PieceType
}

// NewPieceFactory 创建工厂。
func NewPieceFactory() *PieceFactory {
	return &PieceFactory{cache: make(map[Color]*PieceType)}
}

// Get 获取（或创建）指定颜色的共享 PieceType。
func (f *PieceFactory) Get(c Color) *PieceType {
	if t, ok := f.cache[c]; ok {
		fmt.Printf("[Factory] 复用享元: %s（缓存已有）\n", c)
		return t
	}
	symbol := "●"
	if c == White {
		symbol = "○"
	}
	t := &PieceType{color: c, symbol: symbol}
	f.cache[c] = t
	fmt.Printf("[Factory] 新建享元: %s（缓存现有 %d 种）\n", c, len(f.cache))
	return t
}

// CachedCount 当前缓存的享元个数。
func (f *PieceFactory) CachedCount() int { return len(f.cache) }

// =============================================================================
// 三、外在状态 + 棋盘：落子记录持有「享元引用 + 坐标」
// =============================================================================

// Placement 一枚落子：共享类型 + 外在坐标。
type Placement struct {
	typ *PieceType
	row int
	col int
}

// Board 棋盘：只存落子列表；外观全部来自工厂共享对象。
type Board struct {
	factory *PieceFactory
	stones  []Placement
}

// NewBoard 创建空棋盘。
func NewBoard(f *PieceFactory) *Board {
	return &Board{factory: f}
}

// Place 在 (row,col) 落一子；类型从工厂取共享实例。
func (b *Board) Place(c Color, row, col int) {
	t := b.factory.Get(c)
	b.stones = append(b.stones, Placement{typ: t, row: row, col: col})
	t.Draw(row, col)
}

// StoneCount 已落子数量。
func (b *Board) StoneCount() int { return len(b.stones) }

// PrintSummary 打印享元复用结论。
func (b *Board) PrintSummary() {
	fmt.Println()
	fmt.Println("========== 享元复用核对 ==========")
	fmt.Printf("落子总数:     %d\n", b.StoneCount())
	fmt.Printf("享元缓存数:   %d  （期望 2：黑 + 白）\n", b.factory.CachedCount())
	fmt.Println()
	fmt.Println("各落子指向的 PieceType 指针（同色应相同）:")
	for i, s := range b.stones {
		fmt.Printf("  #%d (%d,%d) %s  typ=%p\n",
			i+1, s.row, s.col, s.typ.Symbol(), s.typ)
	}
}

// =============================================================================
// 四、main
// =============================================================================

func main() {
	factory := NewPieceFactory()
	board := NewBoard(factory)

	fmt.Println("========== 开始落子 ==========")
	board.Place(Black, 3, 3)
	board.Place(White, 3, 4)
	board.Place(Black, 4, 4)
	board.Place(White, 4, 3)
	board.Place(Black, 5, 5)
	board.Place(White, 2, 2)

	board.PrintSummary()

	fmt.Println()
	fmt.Println("========== 读懂输出后你会发现 ==========")
	fmt.Println("1. 内在状态（颜色/符号）在 PieceType 里，由工厂缓存")
	fmt.Println("2. 外在状态（行列）在 Placement / Place 参数里")
	fmt.Println("3. 6 次落子，工厂仍只有 2 个享元；同色指针相同")
	fmt.Println("4. 这就是享元：共享不变部分，变化部分外置")
}
