// 演变第 3 步：一个命令挂多个 Operate
//
// 运行：
//
//	cd evolve/step3_命令含多子操作
//	go run .
//
// 相对 step2，改了什么 / 没改什么：
//
// 没改（稳定点）：
//   - Operate 接口、EditCommand 的 Undo/Redo 逆序/正序逻辑
//   - CommandList：Start / AddOperate / Commit / Undo(n) / Redo(n) / cur 指针
//
// 只加了：
//   - ClearCellOperate（新子操作类型）
//   - main：一次 StartCommand 里 AddOperate 多次，再 Commit
//     → 用户看到的是一步「填充表头」；Ctrl+Z 一次全部回退
//
// 复合命令调用链：
//
//	StartCommand("填充表头")
//	  AddOperate(Set A1="姓名")
//	  AddOperate(Set B1="年龄")
//	  AddOperate(Set C1="部门")
//	CommitCommand()  → 三个 Do 正序执行，算 1 个 undo 步
//	Undo(1)          → 三个 Undo 逆序执行（先 C1，再 B1，再 A1）
package main

import "fmt"

// =============================================================================
// 以下 Sheet / Operate / EditCommand / CommandList：与 step2 同构（稳定点）
// =============================================================================

// Sheet 表格文档（Receiver）：命令与 Operate 最终改动的对象。
type Sheet struct {
	// cells 单元格存储：键为地址（如 "A1"），值为单元格文本；未出现的键表示空单元格。
	cells map[string]string
}

func NewSheet() *Sheet {
	return &Sheet{cells: make(map[string]string)}
}

func (s *Sheet) get(addr string) string { return s.cells[addr] }

func (s *Sheet) set(addr, value string) {
	if value == "" {
		delete(s.cells, addr)
		return
	}
	s.cells[addr] = value
}

func (s *Sheet) Dump(tag string) {
	fmt.Printf("  [%s] 表=%v\n", tag, s.cells)
}

// Operate 子操作接口。
type Operate interface {
	Do()
	Undo()
}

// SetCellOperate 设置单元格。
type SetCellOperate struct {
	sheet                    *Sheet
	addr, newValue, oldValue string
}

func NewSetCellOperate(sheet *Sheet, addr, newValue string) *SetCellOperate {
	return &SetCellOperate{sheet: sheet, addr: addr, newValue: newValue, oldValue: sheet.get(addr)}
}

func (o *SetCellOperate) Do() {
	o.sheet.set(o.addr, o.newValue)
	fmt.Printf("    [Operate Do]   Set %s: %q → %q\n", o.addr, o.oldValue, o.newValue)
}

func (o *SetCellOperate) Undo() {
	o.sheet.set(o.addr, o.oldValue)
	fmt.Printf("    [Operate Undo] Set %s: %q → %q\n", o.addr, o.newValue, o.oldValue)
}

// =============================================================================
// 本步新增子操作类型：ClearCellOperate（扩展点，CommandList 不用改）
// =============================================================================

// ClearCellOperate 清空单元格：Do 删除，Undo 写回旧值。
type ClearCellOperate struct {
	sheet    *Sheet
	addr     string
	oldValue string
}

func NewClearCellOperate(sheet *Sheet, addr string) *ClearCellOperate {
	return &ClearCellOperate{sheet: sheet, addr: addr, oldValue: sheet.get(addr)}
}

func (o *ClearCellOperate) Do() {
	o.sheet.set(o.addr, "")
	fmt.Printf("    [Operate Do]   Clear %s (旧=%q)\n", o.addr, o.oldValue)
}

func (o *ClearCellOperate) Undo() {
	o.sheet.set(o.addr, o.oldValue)
	fmt.Printf("    [Operate Undo] Clear %s → 恢复 %q\n", o.addr, o.oldValue)
}

// =============================================================================
// EditCommand / CommandList（与 step2 相同）
// =============================================================================

// EditCommand 一次用户命令；Undo 时对 operates 逆序回退。
type EditCommand struct {
	desc     string
	operates []Operate
}

func NewEditCommand(desc string) *EditCommand {
	return &EditCommand{desc: desc, operates: make([]Operate, 0)}
}

func (c *EditCommand) AddOperate(op Operate) { c.operates = append(c.operates, op) }

func (c *EditCommand) Do() {
	fmt.Printf("  [Command Do] %s（%d 个子操作）\n", c.desc, len(c.operates))
	for _, op := range c.operates {
		op.Do()
	}
}

func (c *EditCommand) Undo() {
	fmt.Printf("  [Command Undo] %s（逆序 %d 个子操作）\n", c.desc, len(c.operates))
	for i := len(c.operates) - 1; i >= 0; i-- {
		c.operates[i].Undo()
	}
}

func (c *EditCommand) Redo() {
	fmt.Printf("  [Command Redo] %s\n", c.desc)
	for _, op := range c.operates {
		op.Do()
	}
}

func (c *EditCommand) Description() string { return c.desc }

// CommandList 命令栈 + 当前指针。
type CommandList struct {
	commands []*EditCommand
	cur      int          // 当前已生效命令下标；-1 表示初始状态
	building *EditCommand // StartCommand 后、Commit 前的「当前命令」
}

func NewCommandList() *CommandList {
	return &CommandList{commands: make([]*EditCommand, 0), cur: -1}
}

func (l *CommandList) StartCommand(desc string) {
	l.building = NewEditCommand(desc)
	fmt.Printf(">> StartCommand(%q)\n", desc)
}

func (l *CommandList) AddOperate(op Operate) {
	if l.building == nil {
		panic("AddOperate: 请先 StartCommand")
	}
	l.building.AddOperate(op)
}

func (l *CommandList) CommitCommand() {
	if l.building == nil {
		panic("CommitCommand: 没有进行中的命令")
	}
	if l.cur+1 < len(l.commands) {
		l.commands = l.commands[:l.cur+1]
	}
	cmd := l.building
	l.building = nil
	cmd.Do()
	l.commands = append(l.commands, cmd)
	l.cur = len(l.commands) - 1
	fmt.Printf(">> CommitCommand → cur=%d undoSteps=%d\n", l.cur, l.UndoSteps())
}

func (l *CommandList) Undo(n int) {
	fmt.Printf(">> Undo(%d)\n", n)
	for i := 0; i < n && l.cur >= 0; i++ {
		l.commands[l.cur].Undo()
		l.cur--
	}
}

func (l *CommandList) Redo(n int) {
	fmt.Printf(">> Redo(%d)\n", n)
	for i := 0; i < n && l.cur+1 < len(l.commands); i++ {
		l.cur++
		l.commands[l.cur].Redo()
	}
}

func (l *CommandList) UndoSteps() int { return l.cur + 1 }
func (l *CommandList) RedoSteps() int { return len(l.commands) - l.cur - 1 }

func main() {
	sheet := NewSheet()
	list := NewCommandList()

	// —— 复合命令：一次 Commit = 多个单元格修改，UI 上只算 1 步撤销 ——
	list.StartCommand("填充表头")
	list.AddOperate(NewSetCellOperate(sheet, "A1", "姓名"))
	list.AddOperate(NewSetCellOperate(sheet, "B1", "年龄"))
	list.AddOperate(NewSetCellOperate(sheet, "C1", "部门"))
	list.CommitCommand()
	sheet.Dump("填充表头后")
	fmt.Printf("可撤销=%d（注意：3 次改单元格只占 1 步）\n", list.UndoSteps())

	// —— 另一条命令：清空一列（新 Operate 类型）——
	list.StartCommand("清空表头 C1")
	list.AddOperate(NewClearCellOperate(sheet, "C1"))
	list.CommitCommand()
	sheet.Dump("清空 C1 后")

	// Undo 1 步：只撤「清空」，表头三格仍在
	list.Undo(1)
	sheet.Dump("Undo(1)后")

	// 再 Undo 1 步：整步「填充表头」逆序撤掉 A1/B1/C1
	list.Undo(1)
	sheet.Dump("再 Undo(1)后")

	list.Redo(1) // 重做填充表头
	sheet.Dump("Redo(1)后")
}
