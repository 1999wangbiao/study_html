// 迭代器模式可运行示范（Go）—— 目录树（组合 + DFS 迭代器）
//
// 核心一句话：用组合搭树，用 All() iter.Seq 做深度优先遍历；
// 客户端 for range 列出所有节点，自己不写递归。
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
// 一、Component：统一接口（组合）+ 遍历入口（迭代器）
// =============================================================================

// Entry 文件系统条目：文件和文件夹都实现它。
type Entry interface {
	Name() string
	Size() int
	// All 深度优先：先交出自己，再依次遍历孩子（叶子无孩子）。
	All() iter.Seq[Entry]
}

// =============================================================================
// 二、Leaf：文件
// =============================================================================

// File 叶子节点。
type File struct {
	name string
	size int
}

// NewFile 创建文件。
func NewFile(name string, size int) *File {
	return &File{name: name, size: size}
}

func (f *File) Name() string { return f.name }
func (f *File) Size() int    { return f.size }

// All 文件只有自身一个节点。
func (f *File) All() iter.Seq[Entry] {
	return func(yield func(Entry) bool) {
		yield(f)
	}
}

// =============================================================================
// 三、Composite：文件夹
// =============================================================================

// Folder 容器：孩子可以是 File 或 Folder。
type Folder struct {
	name     string
	children []Entry
}

// NewFolder 创建空文件夹。
func NewFolder(name string) *Folder {
	return &Folder{name: name}
}

// Add 加入子条目。
func (d *Folder) Add(child Entry) {
	d.children = append(d.children, child)
	fmt.Printf("[Add] %s ← %s\n", d.name, child.Name())
}

func (d *Folder) Name() string { return d.name }

// Size 文件夹大小 = 所有孩子 Size 之和。
func (d *Folder) Size() int {
	total := 0
	for _, c := range d.children {
		total += c.Size()
	}
	return total
}

// All 深度优先：先自己，再对每个孩子 range child.All()。
//
// 提前 break 时通过 yield==false 向上传递，整棵子树停止。
func (d *Folder) All() iter.Seq[Entry] {
	return func(yield func(Entry) bool) {
		if !yield(d) {
			return
		}
		for _, c := range d.children {
			for e := range c.All() {
				if !yield(e) {
					return
				}
			}
		}
	}
}

// =============================================================================
// 四、main：搭树后只用 All() 遍历
// =============================================================================

func main() {
	// docs/
	//   readme.txt
	//   images/
	//     logo.png
	//     banner.png
	root := NewFolder("docs")
	root.Add(NewFile("readme.txt", 10))

	images := NewFolder("images")
	images.Add(NewFile("logo.png", 40))
	images.Add(NewFile("banner.png", 80))
	root.Add(images)

	fmt.Println()
	fmt.Println("========== DFS：for e := range root.All() ==========")
	for e := range root.All() {
		kind := "file"
		if _, ok := e.(*Folder); ok {
			kind = "dir "
		}
		fmt.Printf("  [%s] %s  (%dKB)\n", kind, e.Name(), e.Size())
	}

	fmt.Println()
	fmt.Println("========== 提前 break（只看前两个节点）==========")
	n := 0
	for e := range root.All() {
		fmt.Printf("  → %s\n", e.Name())
		n++
		if n == 2 {
			break
		}
	}

	fmt.Println()
	fmt.Println("========== 读懂输出后你会发现 ==========")
	fmt.Println("1. 组合：File / Folder 都是 Entry（搭树、Size）")
	fmt.Println("2. 迭代器：All() 把 DFS 收进 Seq，main 无手写递归")
	fmt.Println("3. 访问顺序：docs → readme.txt → images → logo → banner")
	fmt.Println("4. 与组合模式目录对照：那边 Print 内递归；这里遍历外置为 Seq")
}
