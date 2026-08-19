// 演变第 3 步：新买一台音箱 —— 只加设备，遥控代码不动
//
// 运行：
//
//	cd evolve/step3_只加新通道
//	go run .
//
// 对照 step1：那时要写 BasicSpeakerRemote、AdvancedSpeakerRemote。
// 现在：只加 Speaker，高级遥控直接 SetDevice(speaker) 就能用。
package main

import "fmt"

type Device interface {
	TurnOn()
	TurnOff()
	Mute()
	Name() string
}

type TV struct{}

func (d *TV) Name() string { return "电视" }
func (d *TV) TurnOn()     { fmt.Println("  [TV] 开机") }
func (d *TV) TurnOff()    { fmt.Println("  [TV] 关机") }
func (d *TV) Mute()       { fmt.Println("  [TV] 静音") }

type Radio struct{}

func (d *Radio) Name() string { return "收音机" }
func (d *Radio) TurnOn()      { fmt.Println("  [Radio] 开机") }
func (d *Radio) TurnOff()     { fmt.Println("  [Radio] 关机") }
func (d *Radio) Mute()        { fmt.Println("  [Radio] 静音") }

// ----- 新设备：音箱（本次新增）-----

type Speaker struct{}

func (d *Speaker) Name() string { return "音箱" }
func (d *Speaker) TurnOn()      { fmt.Println("  [Speaker] 开机，蓝牙已连接") }
func (d *Speaker) TurnOff()     { fmt.Println("  [Speaker] 关机") }
func (d *Speaker) Mute()        { fmt.Println("  [Speaker] 静音") }

// ----- 遥控：与 step2 相同，没有为音箱改一行按键逻辑 -----

type Remote struct{ device Device }

func (r *Remote) SetDevice(d Device) {
	r.device = d
	fmt.Printf("[遥控] 现在对准 → %s\n", d.Name())
}
func (r *Remote) PowerOn()  { r.device.TurnOn() }
func (r *Remote) PowerOff() { r.device.TurnOff() }

type AdvancedRemote struct{ Remote }

func (r *AdvancedRemote) Mute() { r.device.Mute() }

func main() {
	fmt.Println("========== 第 3 步：只加新设备（音箱） ==========")
	adv := &AdvancedRemote{}
	adv.SetDevice(&Speaker{})
	adv.PowerOn()
	adv.Mute()
	adv.PowerOff()

	fmt.Println()
	fmt.Println("看懂了吗？")
	fmt.Println("  加设备 = 新 Device；Remote / AdvancedRemote 不用改。")
	fmt.Println("  加「带语音的遥控」= 新遥控类型；TV/Radio/Speaker 不用改。")
	fmt.Println("  两维各自长，靠 device 字段接在一起 —— 这就是桥接。")
}
