// 抽象工厂模式可运行示范（Go）—— UI 皮肤成套创建
//
// 核心一句话：一次拿一整套互相配套的产品；换工厂 = 换整套风格。
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、抽象产品：客户端只认这些接口
// =============================================================================

// Button 按钮抽象。
type Button interface {
	Paint()
}

// Dialog 对话框抽象。
type Dialog interface {
	Show()
}

// =============================================================================
// 二、具体产品：Windows 风
// =============================================================================

// WinButton Windows 风格按钮。
type WinButton struct{}

// Paint 绘制 Windows 风格按钮。
func (b *WinButton) Paint() {
	fmt.Println("  [WinButton] 直角灰底按钮")
}

// WinDialog Windows 风格对话框。
type WinDialog struct{}

// Show 显示 Windows 风格对话框。
func (d *WinDialog) Show() {
	fmt.Println("  [WinDialog] 标题栏带最小化 / 最大化 / 关闭")
}

// =============================================================================
// 三、具体产品：Mac 风
// =============================================================================

// MacButton Mac 风格按钮。
type MacButton struct{}

// Paint 绘制 Mac 风格按钮。
func (b *MacButton) Paint() {
	fmt.Println("  [MacButton] 圆角胶囊按钮")
}

// MacDialog Mac 风格对话框。
type MacDialog struct{}

// Show 显示 Mac 风格对话框。
func (d *MacDialog) Show() {
	fmt.Println("  [MacDialog] 左上角红黄绿三色灯")
}

// =============================================================================
// 四、抽象工厂 + 具体工厂
// =============================================================================

// UIFactory 抽象工厂：声明一整套 UI 产品的创建方法。
type UIFactory interface {
	CreateButton() Button
	CreateDialog() Dialog
}

// WinFactory Windows 产品族工厂。
type WinFactory struct{}

// CreateButton 创建 Windows 按钮。
func (f *WinFactory) CreateButton() Button { return &WinButton{} }

// CreateDialog 创建 Windows 对话框。
func (f *WinFactory) CreateDialog() Dialog { return &WinDialog{} }

// MacFactory Mac 产品族工厂。
type MacFactory struct{}

// CreateButton 创建 Mac 按钮。
func (f *MacFactory) CreateButton() Button { return &MacButton{} }

// CreateDialog 创建 Mac 对话框。
func (f *MacFactory) CreateDialog() Dialog { return &MacDialog{} }

// =============================================================================
// 五、客户端：只依赖抽象，换工厂即换皮肤
// =============================================================================

// RenderUI 用给定工厂渲染一套 UI（不关心具体平台类名）。
func RenderUI(factory UIFactory) {
	btn := factory.CreateButton()
	dlg := factory.CreateDialog()
	btn.Paint()
	dlg.Show()
}

func main() {
	fmt.Println("========== Windows 皮肤 ==========")
	RenderUI(&WinFactory{})

	fmt.Println("========== Mac 皮肤 ==========")
	RenderUI(&MacFactory{})
}
