// 命令模式可运行示范（Go）—— 对齐 ET 表格编辑器撤销/重做引擎
//
// 对应 C++（command.h 思路）：
//
//	KEtEditCommand      → EditCommand（Undo/Redo + AddOperate + 描述）
//	KEtEditCommandList  → CommandList（Start/Commit/Abort + Undo(n)/Redo(n) + cur）
//	IOperate            → Operate（Do/Undo 子操作）
//
// 本文件相对 evolve 更「丰满」一点：
//	- Sheet 带值 + 样式（粗体 / 填充色）
//	- 多种 Operate：设值、清空、设样式、复制单元格、填充区域
//	- 复合命令（插表头=多格赋值+样式）、Abort、Undo 后新编辑截断重做分支、栈深度上限
//
// 建议先看 evolve/step1→3，再回来看本文件。
//
// 本目录运行：
//
//	go run .
package main

import (
	"fmt"
	"sort"
	"strings"
)

// =============================================================================
// 一、Receiver：Sheet（值 + 简易样式）
// =============================================================================

// CellStyle 单元格样式快照（够演示 Undo 即可）。
type CellStyle struct {
	Bold bool
	Fill string // ""=无填充，如 "yellow"
}

func (s CellStyle) String() string {
	parts := make([]string, 0, 2)
	if s.Bold {
		parts = append(parts, "粗体")
	}
	if s.Fill != "" {
		parts = append(parts, "底色="+s.Fill)
	}
	if len(parts) == 0 {
		return "默认"
	}
	return strings.Join(parts, ",")
}

// Cell 一个格子：文本 + 样式。
type Cell struct {
	Value string
	Style CellStyle
}

// Sheet 极简表格文档。
type Sheet struct {
	cells map[string]Cell
}

func NewSheet() *Sheet {
	return &Sheet{cells: make(map[string]Cell)}
}

func (s *Sheet) getCell(addr string) Cell {
	return s.cells[addr]
}

func (s *Sheet) putCell(addr string, c Cell) {
	// 空值且默认样式 → 从 map 删掉，模拟「格子不存在」
	if c.Value == "" && !c.Style.Bold && c.Style.Fill == "" {
		delete(s.cells, addr)
		return
	}
	s.cells[addr] = c
}

// Dump 按地址排序打印，方便对照 Undo 前后。
func (s *Sheet) Dump(tag string) {
	addrs := make([]string, 0, len(s.cells))
	for a := range s.cells {
		addrs = append(addrs, a)
	}
	sort.Strings(addrs)
	fmt.Printf("  [%s]\n", tag)
	if len(addrs) == 0 {
		fmt.Println("    (空表)")
		return
	}
	for _, a := range addrs {
		c := s.cells[a]
		fmt.Printf("    %s = %q  {%s}\n", a, c.Value, c.Style)
	}
}

// =============================================================================
// 二、Operate（≈ IOperate）：多种可逆改动
// =============================================================================

// Operate 子操作：Do 正向，Undo 反向。
type Operate interface {
	Do()
	Undo()
	Name() string // 调试用，看命令里挂了哪些子操作
}

// ----- 2.1 设值 -----

// SetCellOperate 设置单元格文本（保留原样式，只改 Value）。
type SetCellOperate struct {
	sheet              *Sheet
	addr               string
	newValue, oldValue string
	oldStyle           CellStyle // 格子原先可能不存在，Undo 时整格恢复
	hadCell            bool
}

func NewSetCellOperate(sheet *Sheet, addr, newValue string) *SetCellOperate {
	old := sheet.getCell(addr)
	_, ok := sheet.cells[addr]
	return &SetCellOperate{
		sheet: sheet, addr: addr, newValue: newValue,
		oldValue: old.Value, oldStyle: old.Style, hadCell: ok,
	}
}

func (o *SetCellOperate) Name() string { return "SetCell(" + o.addr + ")" }

func (o *SetCellOperate) Do() {
	c := o.sheet.getCell(o.addr)
	c.Value = o.newValue
	o.sheet.putCell(o.addr, c)
	fmt.Printf("    [Do]   SetCell %s: %q → %q\n", o.addr, o.oldValue, o.newValue)
}

func (o *SetCellOperate) Undo() {
	if !o.hadCell && o.oldValue == "" {
		delete(o.sheet.cells, o.addr)
	} else {
		o.sheet.putCell(o.addr, Cell{Value: o.oldValue, Style: o.oldStyle})
	}
	fmt.Printf("    [Undo] SetCell %s → %q\n", o.addr, o.oldValue)
}

// ----- 2.2 清空（值+样式一起清） -----

// ClearCellOperate 清空格子（整格删除）。
type ClearCellOperate struct {
	sheet   *Sheet
	addr    string
	oldCell Cell
	hadCell bool
}

func NewClearCellOperate(sheet *Sheet, addr string) *ClearCellOperate {
	c, ok := sheet.cells[addr]
	return &ClearCellOperate{sheet: sheet, addr: addr, oldCell: c, hadCell: ok}
}

func (o *ClearCellOperate) Name() string { return "ClearCell(" + o.addr + ")" }

func (o *ClearCellOperate) Do() {
	delete(o.sheet.cells, o.addr)
	fmt.Printf("    [Do]   ClearCell %s (旧=%q {%s})\n", o.addr, o.oldCell.Value, o.oldCell.Style)
}

func (o *ClearCellOperate) Undo() {
	if o.hadCell {
		o.sheet.putCell(o.addr, o.oldCell)
	}
	fmt.Printf("    [Undo] ClearCell %s → 恢复\n", o.addr)
}

// ----- 2.3 设样式 -----

// SetStyleOperate 只改样式，文本不动。
type SetStyleOperate struct {
	sheet              *Sheet
	addr               string
	newStyle, oldStyle CellStyle
	oldValue           string
	hadCell            bool
}

func NewSetStyleOperate(sheet *Sheet, addr string, style CellStyle) *SetStyleOperate {
	old := sheet.getCell(addr)
	_, ok := sheet.cells[addr]
	return &SetStyleOperate{
		sheet: sheet, addr: addr, newStyle: style,
		oldStyle: old.Style, oldValue: old.Value, hadCell: ok,
	}
}

func (o *SetStyleOperate) Name() string { return "SetStyle(" + o.addr + ")" }

func (o *SetStyleOperate) Do() {
	c := o.sheet.getCell(o.addr)
	c.Style = o.newStyle
	o.sheet.putCell(o.addr, c)
	fmt.Printf("    [Do]   SetStyle %s: {%s} → {%s}\n", o.addr, o.oldStyle, o.newStyle)
}

func (o *SetStyleOperate) Undo() {
	if !o.hadCell && o.oldValue == "" && !o.oldStyle.Bold && o.oldStyle.Fill == "" {
		delete(o.sheet.cells, o.addr)
	} else {
		o.sheet.putCell(o.addr, Cell{Value: o.oldValue, Style: o.oldStyle})
	}
	fmt.Printf("    [Undo] SetStyle %s → {%s}\n", o.addr, o.oldStyle)
}

// ----- 2.4 复制单元格 -----

// CopyCellOperate 把 src 整格复制到 dst（覆盖 dst，快照 dst 旧内容以便 Undo）。
type CopyCellOperate struct {
	sheet   *Sheet
	src     string
	dst     string
	srcCell Cell // Do 时读取；Redo 时仍用这份（对齐「命令已封存」）
	oldDst  Cell
	hadDst  bool
}

func NewCopyCellOperate(sheet *Sheet, src, dst string) *CopyCellOperate {
	old, ok := sheet.cells[dst]
	return &CopyCellOperate{
		sheet: sheet, src: src, dst: dst,
		srcCell: sheet.getCell(src),
		oldDst:  old, hadDst: ok,
	}
}

func (o *CopyCellOperate) Name() string { return "CopyCell(" + o.src + "→" + o.dst + ")" }

func (o *CopyCellOperate) Do() {
	o.sheet.putCell(o.dst, o.srcCell)
	fmt.Printf("    [Do]   CopyCell %s → %s  (%q)\n", o.src, o.dst, o.srcCell.Value)
}

func (o *CopyCellOperate) Undo() {
	if o.hadDst {
		o.sheet.putCell(o.dst, o.oldDst)
	} else {
		delete(o.sheet.cells, o.dst)
	}
	fmt.Printf("    [Undo] CopyCell %s → 恢复目标\n", o.dst)
}

// ----- 2.5 区域填充（一个 Operate 改多格，内部仍可逆） -----

// FillRangeOperate 把同一值写入多个地址（演示「子操作本身也可复合」）。
type FillRangeOperate struct {
	sheet    *Sheet
	addrs    []string
	value    string
	oldCells map[string]Cell
	had      map[string]bool
}

func NewFillRangeOperate(sheet *Sheet, addrs []string, value string) *FillRangeOperate {
	oldCells := make(map[string]Cell, len(addrs))
	had := make(map[string]bool, len(addrs))
	for _, a := range addrs {
		c, ok := sheet.cells[a]
		oldCells[a] = c
		had[a] = ok
	}
	return &FillRangeOperate{sheet: sheet, addrs: append([]string{}, addrs...), value: value, oldCells: oldCells, had: had}
}

func (o *FillRangeOperate) Name() string {
	return fmt.Sprintf("FillRange(%v=%q)", o.addrs, o.value)
}

func (o *FillRangeOperate) Do() {
	for _, a := range o.addrs {
		c := o.sheet.getCell(a)
		c.Value = o.value
		o.sheet.putCell(a, c)
	}
	fmt.Printf("    [Do]   FillRange %v → %q\n", o.addrs, o.value)
}

func (o *FillRangeOperate) Undo() {
	for _, a := range o.addrs {
		if o.had[a] {
			o.sheet.putCell(a, o.oldCells[a])
		} else {
			delete(o.sheet.cells, a)
		}
	}
	fmt.Printf("    [Undo] FillRange %v 已恢复\n", o.addrs)
}

// =============================================================================
// 三、EditCommand（≈ KEtEditCommand）
// =============================================================================

// EditCommand 一次用户可见编辑：描述 + 子操作列表。
type EditCommand struct {
	desc     string
	operates []Operate
}

func NewEditCommand(desc string) *EditCommand {
	return &EditCommand{desc: desc, operates: make([]Operate, 0)}
}

// AddOperate 添加子操作（对齐 AddOperate）。
func (c *EditCommand) AddOperate(op Operate) {
	c.operates = append(c.operates, op)
}

func (c *EditCommand) Do() {
	fmt.Printf("  [Command Do] %s（%d 个子操作: %s）\n", c.desc, len(c.operates), joinOpNames(c.operates))
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

func joinOpNames(ops []Operate) string {
	names := make([]string, len(ops))
	for i, op := range ops {
		names[i] = op.Name()
	}
	return strings.Join(names, ", ")
}

// =============================================================================
// 四、CommandList（≈ KEtEditCommandList）
// =============================================================================

const defaultMaxUndo = 20 // 对齐 GetMaxUndoStep 的简化版

// CommandList 命令栈。
// cur 对齐 m_nCurCommandID：指向当前已生效的最后一条；-1 = 初始状态。
type CommandList struct {
	commands  []*EditCommand
	cur       int
	building  *EditCommand
	maxUndo   int // 超过则丢掉最老的已提交命令
}

func NewCommandList() *CommandList {
	return &CommandList{commands: make([]*EditCommand, 0), cur: -1, maxUndo: defaultMaxUndo}
}

// StartCommand 开始收集子操作（对齐 StartCommand）。
func (l *CommandList) StartCommand(desc string) {
	if l.building != nil {
		panic("StartCommand: 上一条尚未 Commit/Abort")
	}
	l.building = NewEditCommand(desc)
	fmt.Printf(">> StartCommand(%q)\n", desc)
}

// AddOperate 挂到当前未提交命令。
func (l *CommandList) AddOperate(op Operate) {
	if l.building == nil {
		panic("AddOperate: 请先 StartCommand")
	}
	l.building.AddOperate(op)
}

// AbortCommand 放弃进行中的命令（未 Commit，不改文档、不压栈）。
func (l *CommandList) AbortCommand() {
	if l.building == nil {
		return
	}
	fmt.Printf(">> AbortCommand(%q) 已丢弃 %d 个子操作\n", l.building.desc, len(l.building.operates))
	l.building = nil
}

// CommitCommand 执行并压栈；若曾 Undo，先丢掉重做分支。
func (l *CommandList) CommitCommand() {
	if l.building == nil {
		panic("CommitCommand: 没有进行中的命令")
	}
	if len(l.building.operates) == 0 {
		fmt.Printf(">> CommitCommand(%q) 无子操作，视为 Abort\n", l.building.desc)
		l.building = nil
		return
	}
	if l.cur+1 < len(l.commands) {
		dropped := len(l.commands) - l.cur - 1
		l.commands = l.commands[:l.cur+1]
		fmt.Printf("  (截断重做分支 %d 条)\n", dropped)
	}

	cmd := l.building
	l.building = nil
	cmd.Do()
	l.commands = append(l.commands, cmd)
	l.cur = len(l.commands) - 1

	// 超出上限：从头部丢掉最老命令（已无法再 Undo 到更早）
	for l.UndoSteps() > l.maxUndo {
		l.commands = l.commands[1:]
		l.cur--
		fmt.Println("  (超出 maxUndo，丢弃最老一条命令)")
	}
	fmt.Printf(">> CommitCommand → cur=%d undo=%d redo=%d\n", l.cur, l.UndoSteps(), l.RedoSteps())
}

// Undo 撤销 n 步（对齐 Undo(UINT)）。
func (l *CommandList) Undo(n int) {
	fmt.Printf(">> Undo(%d)\n", n)
	for i := 0; i < n && l.cur >= 0; i++ {
		l.commands[l.cur].Undo()
		l.cur--
	}
}

// Redo 重做 n 步（对齐 Redo(UINT)）。
func (l *CommandList) Redo(n int) {
	fmt.Printf(">> Redo(%d)\n", n)
	for i := 0; i < n && l.cur+1 < len(l.commands); i++ {
		l.cur++
		l.commands[l.cur].Redo()
	}
}

func (l *CommandList) UndoSteps() int { return l.cur + 1 }
func (l *CommandList) RedoSteps() int { return len(l.commands) - l.cur - 1 }

// UndoInfo 可撤销命令描述（近 → 远，对齐 GetUndoInfo 简化）。
func (l *CommandList) UndoInfo() []string {
	out := make([]string, 0, l.UndoSteps())
	for i := l.cur; i >= 0; i-- {
		out = append(out, l.commands[i].Description())
	}
	return out
}

// RedoInfo 可重做命令描述（近 → 远）。
func (l *CommandList) RedoInfo() []string {
	out := make([]string, 0, l.RedoSteps())
	for i := l.cur + 1; i < len(l.commands); i++ {
		out = append(out, l.commands[i].Description())
	}
	return out
}

// ClearUndoRedoList 清空历史（对齐 ClearUndoRedoList 简化：全清）。
func (l *CommandList) ClearUndoRedoList() {
	l.commands = nil
	l.cur = -1
	fmt.Println(">> ClearUndoRedoList")
}

// =============================================================================
// 五、演示脚本
// =============================================================================

func main() {
	sheet := NewSheet()
	list := NewCommandList()
	list.maxUndo = 10

	headerStyle := CellStyle{Bold: true, Fill: "yellow"}

	// ---------- 命令1：插入表头 = 三格赋值 + 三格样式（对齐「插表=创建+设样式」）----------
	list.StartCommand("插入表头")
	list.AddOperate(NewSetCellOperate(sheet, "A1", "姓名"))
	list.AddOperate(NewSetCellOperate(sheet, "B1", "年龄"))
	list.AddOperate(NewSetCellOperate(sheet, "C1", "部门"))
	list.AddOperate(NewSetStyleOperate(sheet, "A1", headerStyle))
	list.AddOperate(NewSetStyleOperate(sheet, "B1", headerStyle))
	list.AddOperate(NewSetStyleOperate(sheet, "C1", headerStyle))
	list.CommitCommand()
	sheet.Dump("命令1后")

	// ---------- 命令2：录入一行数据 ----------
	list.StartCommand("录入第2行")
	list.AddOperate(NewSetCellOperate(sheet, "A2", "张三"))
	list.AddOperate(NewSetCellOperate(sheet, "B2", "28"))
	list.AddOperate(NewSetCellOperate(sheet, "C2", "研发"))
	list.CommitCommand()

	// ---------- 命令3：区域填充占位符 ----------
	list.StartCommand("填充占位 (A3:C3)")
	list.AddOperate(NewFillRangeOperate(sheet, []string{"A3", "B3", "C3"}, "-"))
	list.CommitCommand()
	sheet.Dump("命令3后")
	fmt.Printf("Undo菜单: %v\n", list.UndoInfo())

	// ---------- 命令4：复制 A2 → A4（整格含样式）----------
	list.StartCommand("复制 A2 → A4")
	list.AddOperate(NewCopyCellOperate(sheet, "A2", "A4"))
	list.CommitCommand()

	// ---------- Abort：开始了又取消（不进栈）----------
	list.StartCommand("误操作草稿")
	list.AddOperate(NewClearCellOperate(sheet, "A1"))
	list.AbortCommand()
	sheet.Dump("Abort后（表不应变）")

	// ---------- Undo 两步，再看重做菜单 ----------
	list.Undo(2)
	sheet.Dump("Undo(2)后")
	fmt.Printf("Undo菜单: %v\n", list.UndoInfo())
	fmt.Printf("Redo菜单: %v\n", list.RedoInfo())

	// ---------- 关键点：Undo 之后做新编辑 → 截断重做分支 ----------
	list.StartCommand("改部门为平台")
	list.AddOperate(NewSetCellOperate(sheet, "C2", "平台"))
	list.CommitCommand()
	sheet.Dump("新编辑后（原 Redo 分支应已丢）")
	fmt.Printf("Redo菜单: %v  (应为空)\n", list.RedoInfo())

	// ---------- 再 Undo/Redo 验证复合命令 ----------
	list.Undo(1)
	sheet.Dump("撤「改部门」")
	list.Redo(1)
	sheet.Dump("重做「改部门」")

	list.Undo(list.UndoSteps()) // 一路撤到空表
	sheet.Dump("全部 Undo 后")
	list.Redo(1) // 只重做「插入表头」
	sheet.Dump("只 Redo 表头")
}
