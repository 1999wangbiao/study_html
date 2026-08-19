// 演变第 2 步：换种理解 —— 遥控器「插着」一台设备
//
// 运行：
//
//	cd evolve/step2_桥接拆开
//	go run .
//
// =============================================================================
// 先忘掉类名，只记一件事
// =============================================================================
//
//	遥控器里有一个字段：device（这就是「桥」）
//	按电源键时：遥控不自己变魔术，而是喊 device.TurnOn/TurnOff
//
//	同一只高级遥控：
//	  - 先对准电视   → SetDevice(TV)
//	  - 再对准收音机 → SetDevice(Radio)
//	按键逻辑不用改，换的是「手里握着谁」。
//
//	这不是玄学，就是组合：Remote 持有 Device。
package main

import "fmt"

// ----- 设备（实现侧）：电视 / 收音机自己知道怎么开关 -----

type Device interface {
	TurnOn()
	TurnOff()
	Mute()
	Name() string
}

// TV 电视
type TV struct{ on bool }

func (d *TV) Name() string { return "电视" }
func (d *TV) TurnOn() {
	d.on = true
	fmt.Println("  [TV] 开机，画面亮了")
}
func (d *TV) TurnOff() {
	d.on = false
	fmt.Println("  [TV] 关机")
}
func (d *TV) Mute() { fmt.Println("  [TV] 静音") }

// Radio 收音机
type Radio struct{ on bool }

func (d *Radio) Name() string { return "收音机" }
func (d *Radio) TurnOn() {
	d.on = true
	fmt.Println("  [Radio] 开机，开始出声")
}
func (d *Radio) TurnOff() {
	d.on = false
	fmt.Println("  [Radio] 关机")
}
func (d *Radio) Mute() { fmt.Println("  [Radio] 静音") }

// ----- 遥控器（抽象侧）：只定义「有哪些键」，键落到 device 上 -----

// Remote 基础遥控：里面插着一台 Device（桥）。
type Remote struct {
	device Device
}

func (r *Remote) SetDevice(d Device) {
	r.device = d
	fmt.Printf("[遥控] 现在对准 → %s\n", d.Name())
}

func (r *Remote) PowerOn() {
	fmt.Printf("[基础遥控] 按电源开（交给 %s）\n", r.device.Name())
	r.device.TurnOn()
}

func (r *Remote) PowerOff() {
	fmt.Printf("[基础遥控] 按电源关（交给 %s）\n", r.device.Name())
	r.device.TurnOff()
}

// AdvancedRemote 高级遥控：多一个静音键，仍然走同一座桥。
type AdvancedRemote struct {
	Remote
}

func (r *AdvancedRemote) Mute() {
	fmt.Printf("[高级遥控] 按静音（交给 %s）\n", r.device.Name())
	r.device.Mute()
}

func main() {
	tv := &TV{}
	radio := &Radio{}

	fmt.Println("========== 第 2 步：遥控插着设备（桥） ==========")

	// 一只高级遥控，先对准电视
	adv := &AdvancedRemote{}
	adv.SetDevice(tv)
	adv.PowerOn()
	adv.Mute()
	adv.PowerOff()

	fmt.Println()
	// 同一只遥控，换对准收音机 —— 没换遥控类型，只换了 device
	adv.SetDevice(radio)
	adv.PowerOn()
	adv.Mute()

	fmt.Println()
	fmt.Println("记住这一句：")
	fmt.Println("  桥 = Remote 里的 device 字段。")
	fmt.Println("  按键在遥控上，干活在设备上。")
	fmt.Println()
	fmt.Println("下一步 step3_只加新设备：家里新买「音箱」，遥控不用改，只加 Speaker。")
}
