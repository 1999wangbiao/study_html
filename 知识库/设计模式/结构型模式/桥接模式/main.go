// 桥接模式可运行示范（Go）—— 遥控器 × 设备（最终形态）
//
// 若还是觉得抽象，请按顺序跑演变（换了一种讲法：遥控插着设备）：
//
//	evolve/step1_继承乘积爆炸  → 功能×设备粘成 4 个类
//	evolve/step2_桥接拆开      → 遥控持有 device（本文件同思路）
//	evolve/step3_只加新设备    → 只加 Speaker，遥控不动
//
// 一句话：
//
//	桥 = Remote.device
//	按键在遥控上，真正开关在 TV / Radio 上。
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 设备侧（Implementor）：会被遥控「对准」的东西
// =============================================================================

type Device interface {
	TurnOn()
	TurnOff()
	Mute()
	Name() string
}

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

// =============================================================================
// 遥控侧（Abstraction）：按键；内部插着一台 Device
// =============================================================================

type Remote struct {
	device Device // ← 桥在这里
}

func (r *Remote) SetDevice(d Device) {
	r.device = d
	fmt.Printf("[遥控] 现在对准 → %s\n", d.Name())
}

func (r *Remote) PowerOn() {
	fmt.Printf("[基础遥控] 按电源开 → 交给 %s\n", r.device.Name())
	r.device.TurnOn()
}

func (r *Remote) PowerOff() {
	fmt.Printf("[基础遥控] 按电源关 → 交给 %s\n", r.device.Name())
	r.device.TurnOff()
}

// AdvancedRemote 高级遥控：多静音键，仍走同一座桥。
type AdvancedRemote struct {
	Remote
}

func (r *AdvancedRemote) Mute() {
	fmt.Printf("[高级遥控] 按静音 → 交给 %s\n", r.device.Name())
	r.device.Mute()
}

func main() {
	tv := &TV{}
	radio := &Radio{}

	fmt.Println("========== 高级遥控 × 电视 ==========")
	adv := &AdvancedRemote{}
	adv.SetDevice(tv)
	adv.PowerOn()
	adv.Mute()
	adv.PowerOff()

	fmt.Println()
	fmt.Println("========== 同一只遥控，换对准收音机 ==========")
	adv.SetDevice(radio)
	adv.PowerOn()
	adv.Mute()

	fmt.Println()
	fmt.Println("========== 读懂后只记三句 ==========")
	fmt.Println("1. 桥 = device 字段（遥控握着设备）")
	fmt.Println("2. 换设备 = SetDevice，不用换遥控类型")
	fmt.Println("3. 两维独立：遥控功能变 / 设备种类变，互不拖累")
}
