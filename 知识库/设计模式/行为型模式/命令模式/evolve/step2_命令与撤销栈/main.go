// 演变第 2 步：Operate + EditCommand + CommandList
//
// 运行：
//
//	cd evolve/step2_命令与撤销栈
//	go run .
//
// 一次用户编辑的调用链：
//
//	list.StartCommand("设置 A1")
//	list.AddOperate(SetCellOperate{A1, "100"})   // 先记旧值，再改表
//	list.CommitCommand()                         // 压栈，cur 指向它
//
//	list.Undo(1)  → commands[cur].Undo() → cur--
//	list.Redo(1)  → cur++ → commands[cur].Redo()
//
// 本步每个命令只放 1 个子操作；step3 演示「一个命令多个 Operate」。
package main

import "fmt"

// =============================================================================
// Receiver：Sheet（真正改数据的文档）
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

// =============================================================================
// Operate：子操作，可正向执行、可反向回退
// =============================================================================

// Operate 一次最小改动。EditCommand 里可挂多条。
type Operate interface {
	Do()   // 正向：改文档
	Undo() // 反向：恢复
}

// SetCellOperate 设置单元格：构造时记下 old，Do 写 new，Undo 写回 old。
type SetCellOperate struct {
	sheet    *Sheet
	addr     string
	newValue string
	oldValue string
}

// NewSetCellOperate 在「即将修改前」采样旧值（先快照再改）。
func NewSetCellOperate(sheet *Sheet, addr, newValue string) *SetCellOperate {
	return &SetCellOperate{
		sheet:    sheet,
		addr:     addr,
		newValue: newValue,
		oldValue: sheet.get(addr),
	}
}

func (o *SetCellOperate) Do() {
	o.sheet.set(o.addr, o.newValue)
	fmt.Printf("    [Operate Do]   %s: %q → %q\n", o.addr, o.oldValue, o.newValue)
}

func (o *SetCellOperate) Undo() {
	o.sheet.set(o.addr, o.oldValue)
	fmt.Printf("    [Operate Undo] %s: %q → %q\n", o.addr, o.newValue, o.oldValue)
}

// =============================================================================
// EditCommand：一次用户可见的「编辑命令」
// =============================================================================

// EditCommand 一个命令 = 描述 + 若干子操作。
// Undo：子操作逆序 Undo；Redo：子操作正序 Do。
type EditCommand struct {
	desc     string
	operates []Operate
}

func NewEditCommand(desc string) *EditCommand {
	return &EditCommand{desc: desc, operates: make([]Operate, 0)}
}

// AddOperate 添加子操作。
func (c *EditCommand) AddOperate(op Operate) {
	c.operates = append(c.operates, op)
}

// Do 提交时执行全部子操作（正序）。
func (c *EditCommand) Do() {
	fmt.Printf("  [Command Do] %s\n", c.desc)
	for _, op := range c.operates {
		op.Do()
	}
}

// Undo 撤销本命令（逆序回退子操作）。
func (c *EditCommand) Undo() {
	fmt.Printf("  [Command Undo] %s\n", c.desc)
	for i := len(c.operates) - 1; i >= 0; i-- {
		c.operates[i].Undo()
	}
}

// Redo 重做本命令（再正序执行一次）。
func (c *EditCommand) Redo() {
	fmt.Printf("  [Command Redo] %s\n", c.desc)
	for _, op := range c.operates {
		op.Do()
	}
}

func (c *EditCommand) Description() string { return c.desc }

// =============================================================================
// CommandList：命令栈 + 当前指针
// =============================================================================

// CommandList 管理已提交命令。
//
// 栈模型：
//
//	commands = [c0, c1, c2, ...]
//	cur 指向「当前已生效的最后一条」下标；-1 表示全部已撤回到初始。
//
//	Undo 1 步：commands[cur].Undo(); cur--
//	Redo 1 步：cur++; commands[cur].Redo()
//
//	新 Commit 时：丢掉 cur 之后的「未来」（重做分支作废），再追加新命令。
type CommandList struct {
	commands []*EditCommand
	cur      int          // 当前已生效命令下标；-1 表示初始状态
	building *EditCommand // StartCommand 后、Commit 前的「当前命令」
}

func NewCommandList() *CommandList {
	return &CommandList{commands: make([]*EditCommand, 0), cur: -1}
}

// StartCommand 开始一个新命令。
func (l *CommandList) StartCommand(desc string) {
	l.building = NewEditCommand(desc)
	fmt.Printf(">> StartCommand(%q)\n", desc)
}

// AddOperate 往「当前未提交命令」里挂子操作。
func (l *CommandList) AddOperate(op Operate) {
	if l.building == nil {
		panic("AddOperate: 请先 StartCommand")
	}
	l.building.AddOperate(op)
}

// CommitCommand 执行并压栈。
// 若此前 Undo 过，先截断重做分支。
func (l *CommandList) CommitCommand() {
	if l.building == nil {
		panic("CommitCommand: 没有进行中的命令")
	}
	// 丢弃 cur 之后的命令（用户 Undo 后又做了新编辑）
	if l.cur+1 < len(l.commands) {
		l.commands = l.commands[:l.cur+1]
	}

	cmd := l.building
	l.building = nil
	cmd.Do() // 真正改文档
	l.commands = append(l.commands, cmd)
	l.cur = len(l.commands) - 1
	fmt.Printf(">> CommitCommand → cur=%d undoSteps=%d\n", l.cur, l.UndoSteps())
}

// Undo 撤销 n 步。
func (l *CommandList) Undo(n int) {
	fmt.Printf(">> Undo(%d)\n", n)
	for i := 0; i < n && l.cur >= 0; i++ {
		l.commands[l.cur].Undo()
		l.cur--
	}
}

// Redo 重做 n 步。
func (l *CommandList) Redo(n int) {
	fmt.Printf(">> Redo(%d)\n", n)
	for i := 0; i < n && l.cur+1 < len(l.commands); i++ {
		l.cur++
		l.commands[l.cur].Redo()
	}
}

// UndoSteps 可撤销步数。
func (l *CommandList) UndoSteps() int {
	return l.cur + 1
}

// RedoSteps 可重做步数。
func (l *CommandList) RedoSteps() int {
	return len(l.commands) - l.cur - 1
}

func main() {
	sheet := NewSheet()
	list := NewCommandList()

	// —— 编辑 1：设置 A1 ——
	list.StartCommand("设置 A1=100")
	list.AddOperate(NewSetCellOperate(sheet, "A1", "100"))
	list.CommitCommand()
	sheet.Dump("编辑1后")

	// —— 编辑 2：设置 B1 ——
	list.StartCommand("设置 B1=hello")
	list.AddOperate(NewSetCellOperate(sheet, "B1", "hello"))
	list.CommitCommand()
	sheet.Dump("编辑2后")

	// —— 编辑 3：改 A1 ——
	list.StartCommand("设置 A1=200")
	list.AddOperate(NewSetCellOperate(sheet, "A1", "200"))
	list.CommitCommand()
	sheet.Dump("编辑3后")

	fmt.Printf("可撤销=%d 可重做=%d\n", list.UndoSteps(), list.RedoSteps())

	list.Undo(2) // 撤两步 → 回到只剩 A1=100
	sheet.Dump("Undo(2)后")
	fmt.Printf("可撤销=%d 可重做=%d\n", list.UndoSteps(), list.RedoSteps())

	list.Redo(1) // 重做一步 → B1 回来
	sheet.Dump("Redo(1)后")
}
