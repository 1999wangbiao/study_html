// 备忘录模式可运行示范（Go）—— 文本编辑器撤销
//
// 核心一句话：在不暴露对象内部结构的前提下，把「状态快照」取出来
// 存到别处，需要时再整体恢复回去——Originator（编辑器）造快照，
// Caretaker（历史栈）只负责存与取，双方都不解读对方太多内部细节。
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、角色 1：Originator（发起人）——「被备份的对象：文本编辑器」
// =============================================================================

// Editor 文本编辑器：持有正文与光标，可以生成快照、从快照恢复。
type Editor struct {
	text   string
	cursor int
}

// Type 输入一个字符（追加到末尾，光标后移）。
func (e *Editor) Type(ch rune) {
	e.text += string(ch)
	e.cursor++
}

// Show 展示当前状态。
func (e *Editor) Show() {
	fmt.Printf("  [Editor] 内容=%q 光标=%d\n", e.text, e.cursor)
}

// CreateSnapshot 生成当前状态快照（备忘录）。
// 只把必要字段拷进快照，后续继续编辑不会污染已有历史。
func (e *Editor) CreateSnapshot() *Snapshot {
	return &Snapshot{text: e.text, cursor: e.cursor}
}

// Restore 用一份快照整体恢复状态。
func (e *Editor) Restore(s *Snapshot) {
	e.text = s.text
	e.cursor = s.cursor
}

// =============================================================================
// 二、角色 2：Memento（备忘录）——「状态快照本体」
// =============================================================================
//
// Snapshot 把 Editor 的内部状态收成一份拷贝；对 Caretaker 而言它就是一份
// 「不透明」的数据：只负责存与取，不解读内部字段。

// Snapshot 快照：字段对外隐藏，只留只读查看入口。
type Snapshot struct {
	text   string
	cursor int
}

// Describe 只读查看快照内容（便于演示；Caretaker 一般不需要解读）。
func (s *Snapshot) Describe() string {
	return fmt.Sprintf("内容=%q 光标=%d", s.text, s.cursor)
}

// =============================================================================
// 三、角色 3：Caretaker（负责人）——「历史记录栈」
// =============================================================================
//
// History 只管存快照 / 取快照，完全不理解 Editor 的内部逻辑。

// History 撤销历史栈。
type History struct {
	items []*Snapshot
}

// Push 压入一份快照。
func (h *History) Push(s *Snapshot) {
	h.items = append(h.items, s)
}

// Pop 弹出最近一份快照；栈空返回 nil。
func (h *History) Pop() *Snapshot {
	if len(h.items) == 0 {
		return nil
	}
	last := h.items[len(h.items)-1]
	h.items = h.items[:len(h.items)-1]
	return last
}

// =============================================================================
// 四、main：编辑 → 存档 → 编辑 → 撤销
// =============================================================================
//
// 约定：每次「编辑之前」，先把当前状态存一份，作为撤销的锚点。
// 这样弹出栈顶快照并恢复，就能真正回到上一步，而不是原地不动。

func main() {
	e := &Editor{}
	hist := &History{}

	// 空文档先存一份：撤销第一段输入时的锚点
	hist.Push(e.CreateSnapshot())
	fmt.Println("  存档 ①：", e.CreateSnapshot().Describe())

	// 第一稿：输入「你好，世界」
	for _, r := range []rune("你好，世界") {
		e.Type(r)
	}
	e.Show()

	// 编辑前存档：撤销第二段输入时的锚点
	hist.Push(e.CreateSnapshot())
	fmt.Println("  存档 ②：", e.CreateSnapshot().Describe())

	// 继续编辑，补一句完整的话
	for _, r := range []rune("！开始学设计模式") {
		e.Type(r)
	}
	e.Show()

	// 撤销一次：回到「你好，世界」
	fmt.Println("-- 撤销一次 --")
	if s := hist.Pop(); s != nil {
		e.Restore(s)
	}
	e.Show()

	// 再撤销：回到空文档
	fmt.Println("-- 再撤销 --")
	if s := hist.Pop(); s != nil {
		e.Restore(s)
	}
	e.Show()

	// 再撤销：历史已空，忽略
	fmt.Println("-- 第三次撤销 --")
	if s := hist.Pop(); s != nil {
		e.Restore(s)
		e.Show()
	} else {
		fmt.Println("  没有更早的历史，忽略")
	}

	fmt.Println()
	fmt.Println("========== 读懂输出后你会发现 ==========")
	fmt.Println("1. 快照把 Editor 的内部状态（正文+光标）整体拷走，互不干扰")
	fmt.Println("2. History 只存/取快照，看不懂内容也能完成撤销")
	fmt.Println("3. 恢复 = 用一份旧快照整体替换当前状态，而不是逐字段打补丁")
	fmt.Println("4. 新增历史能力（如重做/差异压缩）只动 History，不改 Editor")
}
