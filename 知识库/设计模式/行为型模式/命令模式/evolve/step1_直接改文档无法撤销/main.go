// 演变第 1 步：直接改文档 —— 没有命令对象，也就没有撤销栈
//
// 对照 ET：command.h 里 KEtEditCommand / KEtEditCommandList 要解决的痛点，
// 先看「没有它们时」业务代码长什么样。
//
// 运行：
//
//	cd evolve/step1_直接改文档无法撤销
//	go run .
//
// =============================================================================
// 本步要体会的「错误直觉」
// =============================================================================
//
// 需求：改单元格、清单元格；用户还想 Ctrl+Z。
// 做法：Sheet.SetCell / ClearCell 直接改 map，改完就忘。
//
// 调用链：
//
//	main
//	  └─ sheet.SetCell("A1", "100")   ← 旧值丢掉了
//	  └─ sheet.SetCell("A1", "200")   ← 再也回不到 "100"
//
// 痛点：
//  1. 没有「一次编辑」的对象，Undo/Redo 无从挂载
//  2. 一次用户动作若含多个修改（插表+设样式），无法原子撤销
//  3. UI 无法问「还能撤几步 / 描述是什么」
//
// 下一步（step2）会引入：Operate + EditCommand + CommandList（对齐 ET）。
package main

import "fmt"

// Sheet 极简表格：单元格地址 → 文本（Receiver，本步被直接改）。
type Sheet struct {
	cells map[string]string
}

func NewSheet() *Sheet {
	return &Sheet{cells: make(map[string]string)}
}

// SetCell 直接写入；旧值不保存 → 无法 Undo。
func (s *Sheet) SetCell(addr, value string) {
	s.cells[addr] = value
	fmt.Printf("  [Sheet] %s = %q\n", addr, value)
}

// ClearCell 直接删除。
func (s *Sheet) ClearCell(addr string) {
	delete(s.cells, addr)
	fmt.Printf("  [Sheet] %s 已清空\n", addr)
}

func (s *Sheet) Dump() {
	fmt.Printf("  当前表: %v\n", s.cells)
}

func main() {
	sheet := NewSheet()

	// 业务代码「当场改文档」——和 ET 里绕开命令栈直接改内核类似
	sheet.SetCell("A1", "100")
	sheet.SetCell("B1", "hello")
	sheet.SetCell("A1", "200") // 旧值 "100" 已丢失
	sheet.ClearCell("B1")

	sheet.Dump()
	fmt.Println("→ 想 Ctrl+Z？做不到：没有命令，没有历史。")
}
