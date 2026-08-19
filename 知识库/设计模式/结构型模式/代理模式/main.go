// 代理模式可运行示范（Go）—— 图片懒加载 + 权限校验
//
// 核心一句话：客户端调的是 Image；代理决定「能不能看、何时真正加载」。
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、抽象主题：客户端只认这个接口
// =============================================================================

// Image 图片显示接口（Subject）。
type Image interface {
	Display()
}

// =============================================================================
// 二、真实主题：真正加载并显示（贵）
// =============================================================================

// RealImage 真实图片；构造时模拟从磁盘加载。
type RealImage struct {
	filename string
}

// NewRealImage 创建并立即加载（昂贵操作）。
func NewRealImage(filename string) *RealImage {
	img := &RealImage{filename: filename}
	img.loadFromDisk()
	return img
}

// loadFromDisk 模拟耗时加载。
func (img *RealImage) loadFromDisk() {
	fmt.Printf("  [RealImage] 从磁盘加载 %s …（较慢）\n", img.filename)
}

// Display 显示已加载的图片。
func (img *RealImage) Display() {
	fmt.Printf("  [RealImage] 显示 %s\n", img.filename)
}

// =============================================================================
// 三、代理：懒加载 + 保护（鉴权）
// =============================================================================

// User 简易用户，带查看权限。
type User struct {
	Name    string
	CanView bool
}

// ImageProxy 图片代理：先占坑，首次 Display 且鉴权通过才创建 RealImage。
type ImageProxy struct {
	filename string
	user     User
	real     *RealImage // 按需持有
}

// NewImageProxy 创建代理（此时不加载真实图片）。
func NewImageProxy(filename string, user User) *ImageProxy {
	fmt.Printf("[Proxy] 创建占位代理 → %s（尚未加载）\n", filename)
	return &ImageProxy{filename: filename, user: user}
}

// Display 先鉴权；通过后懒加载真实图片并转发。
func (p *ImageProxy) Display() {
	if !p.user.CanView {
		fmt.Printf("[Proxy] 拒绝：用户 %s 无权查看 %s\n", p.user.Name, p.filename)
		return
	}
	if p.real == nil {
		fmt.Printf("[Proxy] 用户 %s 首次查看，开始加载…\n", p.user.Name)
		p.real = NewRealImage(p.filename)
	} else {
		fmt.Printf("[Proxy] 复用已加载的 RealImage\n")
	}
	p.real.Display()
}

// =============================================================================
// 四、客户端：只依赖 Image
// =============================================================================

// Show 通过 Image 接口展示（不关心是代理还是本体）。
func Show(img Image) {
	img.Display()
}

func main() {
	guest := User{Name: "访客", CanView: false}
	alice := User{Name: "Alice", CanView: true}

	fmt.Println("========== 无权限：保护代理拦截 ==========")
	denied := NewImageProxy("secret.png", guest)
	Show(denied)

	fmt.Println("========== 有权限：首次加载，再次复用 ==========")
	proxy := NewImageProxy("photo.png", alice)
	Show(proxy)
	Show(proxy)
}
