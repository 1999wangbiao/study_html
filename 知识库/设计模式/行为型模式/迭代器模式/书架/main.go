// 迭代器模式可运行示范（Go）—— 书架（线性聚合）
//
// 核心一句话：集合只暴露 Books() iter.Seq，客户端 for range 消费，
// 不必（也不该）直接摸内部切片。
//
// 本目录运行：
//
//	go run .
package main

import (
	"fmt"
	"iter"
)

// =============================================================================
// 一、元素：Book（被遍历的单项）
// =============================================================================

// Book 书架上的一本书，是迭代器每次 yield 出去的「元素」。
type Book struct {
	name string // 书名；小写 = 包外不可直接改字段，只能通过 Name() 读
}

// NewBook 构造一本书。
//
// 含义：把书名封进 Book，返回指针供 Add 放入书架。
func NewBook(name string) *Book {
	return &Book{name: name}
}

// Name 读取书名。
//
// 含义：对外只暴露「读」，不暴露可写字段，避免客户端改坏数据。
func (b *Book) Name() string { return b.name }

// =============================================================================
// 二、聚合：BookShelf（持有元素，并提供迭代器）
// =============================================================================

// BookShelf 书架（GoF 里的 Aggregate）。
//
// 含义：负责「存书」；遍历方式只通过 Books() 提供，不导出 books 切片。
type BookShelf struct {
	books []*Book // 内部存储；故意小写，强迫走迭代器而不是 shelf.books[i]
}

// NewBookShelf 构造空书架。
//
// 含义：得到一个可 Add / Books / Len 的聚合对象。
func NewBookShelf() *BookShelf {
	return &BookShelf{}
}

// Add 往书架末尾放一本书。
//
// 含义：修改聚合内容（写入侧）；和 Books()（只读遍历侧）分开。
func (s *BookShelf) Add(b *Book) {
	s.books = append(s.books, b)
	fmt.Printf("[Add] %s\n", b.Name())
}

// Len 返回当前册数。
//
// 含义：只回答「有多少本」，仍然不把内部切片交出去。
func (s *BookShelf) Len() int { return len(s.books) }

// Books 返回「按放入顺序逐本取出」的迭代器（GoF 里的 Iterator）。
//
// 含义（最重要）：
//   - 返回类型 iter.Seq[*Book]：可以被 for b := range shelf.Books() 使用。
//   - 客户端拿不到 []*Book，只能一本一本接收，从而与内部存储解耦。
//   - 函数体里的 yield：把下一本书「推」给 range；若对方 break，yield 返回 false，
//     此时必须立刻 return，否则提前退出无效。
func (s *BookShelf) Books() iter.Seq[*Book] {
	// 返回的是「推送函数」：Go 的 range 会调用它，并传入 yield
	return func(yield func(*Book) bool) {
		for _, b := range s.books {
			// yield(b)：把当前书交给 for-range 循环体
			// 返回 false：调用方写了 break / return，停止继续推送
			if !yield(b) {
				return
			}
		}
	}
}

// =============================================================================
// 三、main：客户端 —— 只认 Books()，不碰内部切片
// =============================================================================

func main() {
	// 准备数据：创建书架并放入三本书
	shelf := NewBookShelf()
	shelf.Add(NewBook("设计模式"))
	shelf.Add(NewBook("重构"))
	shelf.Add(NewBook("代码整洁之道"))

	// 正常遍历：range Books() ≈ 「请书架按顺序一本本给我」
	fmt.Println()
	fmt.Println("========== 用迭代器遍历（不碰内部切片）==========")
	i := 1
	for b := range shelf.Books() {
		fmt.Printf("  %d. %s\n", i, b.Name())
		i++
	}

	// 提前结束：break 会让 Books() 里的 yield 返回 false 并停止
	fmt.Println()
	fmt.Println("========== 提前 break（验证 yield 协议）==========")
	for b := range shelf.Books() {
		fmt.Printf("  读到: %s（然后 break）\n", b.Name())
		break
	}

	fmt.Println()
	fmt.Printf("书架共 %d 册\n", shelf.Len())
	fmt.Println()
	fmt.Println("========== 读懂输出后你会发现 ==========")
	fmt.Println("1. main 从未写 shelf.books，只调用 Books()")
	fmt.Println("2. Books() 返回 iter.Seq，可用 for range")
	fmt.Println("3. break 时 yield 返回 false，迭代器停止推送")
	fmt.Println("4. 以后若改成链表存储，只需改 Books() 内部")
}
