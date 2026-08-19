// 演变第 1 步：气象站写死通知每一个看板
//
// 运行：
//
//	cd evolve/step1_写死通知看板
//	go run .
//
// =============================================================================
// 本步要体会的「错误直觉」
// =============================================================================
//
// 需求：读数一变，两块看板都要刷新。
// 做法：WeatherData 里直接握着具体看板指针，变化时逐个点名调用。
//
// 调用链（死板、写死）：
//
//	main
//	  └─ SetMeasurements(temp, humidity, pressure)
//	       ├─ 把三个数存进 WeatherData 字段
//	       └─ measurementsChanged()
//	            ├─ current.Update(...)   ← 必须知道 CurrentConditionsDisplay
//	            └─ stats.Update(...)     ← 必须知道 StatisticsDisplay
//
// 痛点：
//  1. 加第三块看板 → 改 struct 字段 + New + measurementsChanged，三处都动
//  2. 想运行时摘掉某块看板 → 做不到，字段写死了
//  3. WeatherData 依赖具体类型，和看板互相粘住
//
// 下一步（step2）会改成：只握 []Observer，循环 Notify，不再出现具体看板名。
package main

import "fmt"

// =============================================================================
// 看板：只管「收到三个数之后怎么显示」
// 本步没有 Observer 接口；看板只是普通结构体，被 WeatherData 字段硬引用。
// =============================================================================

// CurrentConditionsDisplay 当前温湿度看板。
type CurrentConditionsDisplay struct{}

// Update 刷新展示。名字碰巧叫 Update，但本步它不是接口方法，只是普通函数。
func (d *CurrentConditionsDisplay) Update(temp, humidity, pressure float64) {
	// 当前状况只关心温度、湿度；气压本看板不用。
	fmt.Printf("  [当前状况] 温度=%.1f°C 湿度=%.1f%%\n", temp, humidity)
	_ = pressure
}

// StatisticsDisplay 温度统计看板：自己累加历史，算均温 / 最高 / 最低。
type StatisticsDisplay struct {
	maxTemp float64
	minTemp float64
	sumTemp float64
	count   int
}

// Update 用新温度更新统计，再打印。
func (d *StatisticsDisplay) Update(temp, humidity, pressure float64) {
	// 第一次收到数据时，最高 = 最低 = 当前温度。
	if d.count == 0 {
		d.maxTemp, d.minTemp = temp, temp
	}
	if temp > d.maxTemp {
		d.maxTemp = temp
	}
	if temp < d.minTemp {
		d.minTemp = temp
	}
	d.sumTemp += temp
	d.count++
	avg := d.sumTemp / float64(d.count)
	fmt.Printf("  [统计] 均温=%.1f 最高=%.1f 最低=%.1f\n", avg, d.maxTemp, d.minTemp)
	// 统计看板本步也不用湿度、气压。
	_, _ = humidity, pressure
}

// =============================================================================
// WeatherData：数据源 +「通知谁」全写在自己身上（本步的核心问题）
// =============================================================================

// WeatherData 同时干两件事：存读数，以及「知道要通知哪几个具体看板」。
// 后一件事正是耦合来源——字段类型写成了具体看板，而不是抽象接口。
type WeatherData struct {
	temperature float64
	humidity    float64
	pressure    float64

	// ↓ 写死：主题必须 import / 认识这两种具体类型
	current *CurrentConditionsDisplay
	stats   *StatisticsDisplay
	// 若再加 forecast *ForecastDisplay，struct 又要改一行
}

// NewWeatherData 创建气象站时就把两块看板 new 出来并挂上。
// 看板的生命周期被主题绑死：主题创建它们，主题决定何时调用它们。
func NewWeatherData() *WeatherData {
	return &WeatherData{
		current: &CurrentConditionsDisplay{},
		stats:   &StatisticsDisplay{},
	}
}

// SetMeasurements 外部写入新读数的唯一入口。
// 逻辑两步：① 存数据 ② 立刻触发通知（本步用 measurementsChanged）。
func (w *WeatherData) SetMeasurements(temp, humidity, pressure float64) {
	// ① 先更新内部状态
	w.temperature = temp
	w.humidity = humidity
	w.pressure = pressure
	// ② 状态变了 → 通知依赖方（写死版）
	w.measurementsChanged()
}

// measurementsChanged 通知逻辑的「痛点现场」。
// 这里出现了具体类型的方法调用；每多一个依赖方，这里就多一行。
func (w *WeatherData) measurementsChanged() {
	fmt.Println("--- 读数更新，写死通知看板 ---")

	// 点名调用：编译期就绑死了「有且只有这两块看板」
	w.current.Update(w.temperature, w.humidity, w.pressure)
	w.stats.Update(w.temperature, w.humidity, w.pressure)

	// 想加预报看板时，只能再写：
	//   w.forecast.Update(w.temperature, w.humidity, w.pressure)
	// 同时还要改 struct、NewWeatherData —— 主题对扩展不开放。
}

func main() {
	// 使用方只碰 WeatherData；看板藏在主题内部，外面插拔不了。
	wd := NewWeatherData()

	// 每次 SetMeasurements → 存数 → 写死通知两块看板
	wd.SetMeasurements(26.5, 65, 1013.2)
	wd.SetMeasurements(28.0, 70, 1012.8)
}
