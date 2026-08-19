// 组合模式可运行示范（Go）—— 文件系统（文件 + 文件夹）
//
// 核心一句话：让「单个对象」和「对象树」用同一套接口，
// 客户端不用区分眼前是文件还是文件夹，照样 Add / Size / Print。
//
//	Entry（统一接口）
//	  ├── File（叶子：不能再装孩子）
//	  └── Folder（容器：可以装 Entry，包括别的 Folder）
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、Component：统一接口
// =============================================================================

// Entry 文件系统条目：文件和文件夹都实现它。
type Entry interface {
	Name() string
	Size() int      // 文件=自身大小；文件夹=所有孩子之和
	Print(prefix string)
}

// =============================================================================
// 二、Leaf：文件（叶子，没有孩子）
// =============================================================================

// File 叶子节点。
type File struct {
	name string
	size int
}

func NewFile(name string, size int) *File {
	return &File{name: name, size: size}
}

func (f *File) Name() string { return f.name }
func (f *File) Size() int    { return f.size }

func (f *File) Print(prefix string) {
	fmt.Printf("%s- %s (%dKB)\n", prefix, f.name, f.size)
}

// =============================================================================
// 三、Composite：文件夹（可以装 Entry）
// =============================================================================

// Folder 容器节点：孩子可以是 File，也可以是 Folder。
type Folder struct {
	name     string
	children []Entry
}

func NewFolder(name string) *Folder {
	return &Folder{name: name}
}

// Add 加入子条目（文件或文件夹都行——关键：参数类型是 Entry）。
func (d *Folder) Add(child Entry) {
	d.children = append(d.children, child)
	fmt.Printf("[Add] %s ← %s\n", d.name, child.Name())
}

func (d *Folder) Name() string { return d.name }

// Size 递归：文件夹大小 = 所有孩子 Size 之和。
func (d *Folder) Size() int {
	total := 0
	for _, c := range d.children {
		total += c.Size() // 孩子若是 Folder，会继续递归
	}
	return total
}

func (d *Folder) Print(prefix string) {
	fmt.Printf("%s+ %s/ (%dKB)\n", prefix, d.name, d.Size())
	for _, c := range d.children {
		c.Print(prefix + "  ")
	}
}

// =============================================================================
// 四、main：搭一棵树，用同一接口遍历
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
	root.Add(images) // 文件夹里再挂文件夹

	fmt.Println()
	fmt.Println("========== 树形打印（客户端只认 Entry）==========")
	root.Print("")

	fmt.Println()
	fmt.Printf("根目录总大小: %dKB\n", root.Size())

	fmt.Println()
	fmt.Println("========== 读懂输出后你会发现 ==========")
	fmt.Println("1. Add / Size / Print 的参数和接收者都围绕 Entry")
	fmt.Println("2. Folder.Size 递归调用孩子的 Size，不用 type switch")
	fmt.Println("3. 客户端搭树时：文件和文件夹一视同仁地 Add 进去")
	fmt.Println("4. 这就是组合：部分与整体同一接口")
}
