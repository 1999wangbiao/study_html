// 演变第 1 步：遥控器功能和设备品牌「粘」在同一个类里
//
// 运行：
//
//	cd evolve/step1_继承乘积爆炸
//	go run .
//
// 生活场景：
//	你手里有「基础遥控 / 高级遥控」，家里有「电视 / 收音机」。
//
// 错误直觉：每种组合做一个类
//	BasicTVRemote、BasicRadioRemote、AdvancedTVRemote、AdvancedRadioRemote
//
// 问题：
//	- 再买一台「音箱」→ 立刻再写两个遥控类
//	- 「高级遥控的静音」在 TV 版和 Radio 版里各抄一遍
package main

import "fmt"

// BasicTVRemote 基础遥控 × 电视（功能和设备粘死）。
type BasicTVRemote struct{ on bool }

func (r *BasicTVRemote) Power() {
	r.on = !r.on
	fmt.Printf("[BasicTVRemote] 电视电源 → %v\n", r.on)
}

// AdvancedTVRemote 高级遥控 × 电视。
type AdvancedTVRemote struct{ on bool }

func (r *AdvancedTVRemote) Power() {
	r.on = !r.on
	fmt.Printf("[AdvancedTVRemote] 电视电源 → %v\n", r.on)
}

func (r *AdvancedTVRemote) Mute() {
	fmt.Println("[AdvancedTVRemote] 电视静音")
}

// BasicRadioRemote 基础遥控 × 收音机。
type BasicRadioRemote struct{ on bool }

func (r *BasicRadioRemote) Power() {
	r.on = !r.on
	fmt.Printf("[BasicRadioRemote] 收音机电源 → %v\n", r.on)
}

// AdvancedRadioRemote 高级遥控 × 收音机。
type AdvancedRadioRemote struct{ on bool }

func (r *AdvancedRadioRemote) Power() {
	r.on = !r.on
	fmt.Printf("[AdvancedRadioRemote] 收音机电源 → %v\n", r.on)
}

func (r *AdvancedRadioRemote) Mute() {
	fmt.Println("[AdvancedRadioRemote] 收音机静音") // 和电视版几乎一样，又抄一遍
}

func main() {
	fmt.Println("========== 第 1 步：功能和设备粘在一起 ==========")
	(&BasicTVRemote{}).Power()
	advTV := &AdvancedTVRemote{}
	advTV.Power()
	advTV.Mute()
	(&BasicRadioRemote{}).Power()
	advRadio := &AdvancedRadioRemote{}
	advRadio.Power()
	advRadio.Mute()

	fmt.Println()
	fmt.Println("痛点：遥控功能 × 设备种类 = 类的数量。")
	fmt.Println("下一站 step2：遥控只负责按键；设备自己负责怎么开关。")
}
