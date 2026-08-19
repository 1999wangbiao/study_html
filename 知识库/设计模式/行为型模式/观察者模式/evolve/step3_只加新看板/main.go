// 演变第 3 步：只加 ForecastDisplay，WeatherData / 旧看板不动
//
// 运行：
//
//	cd evolve/step3_只加新看板
//	go run .
//
// =============================================================================
// 相对 step2，改了什么 / 没改什么
// =============================================================================
//
// 没改（请对照 step2 文件，应几乎一字不差）：
//	- Observer / Subject 接口
//	- WeatherData 全部方法（Register / Remove / Notify / SetMeasurements）
//	- CurrentConditionsDisplay、StatisticsDisplay
//
// 只加了：
//	- 新类型 ForecastDisplay + NewForecastDisplay + Update
//	- main 里多一行 _ = NewForecastDisplay(wd)
//
// 这就是观察者「对扩展开放」的体感：
//	扩展点 = 新观察者类型；稳定点 = 主题。
//
// 若还用 step1 写法，加预报必须改 WeatherData.struct / New / measurementsChanged。
//
// 调用链（与 step2 相同，只是列表多一个观察者）：
//
//	main
//	  ├─ NewCurrentConditionsDisplay(wd) → Register
//	  ├─ NewStatisticsDisplay(wd)        → Register
//	  ├─ NewForecastDisplay(wd)          → Register   ← 本步唯一新增挂接
//	  └─ SetMeasurements(...)
//	       └─ NotifyObservers()
//	            ├─ current.Update(...)
//	            ├─ stats.Update(...)
//	            └─ forecast.Update(...)   ← Notify 循环不用改，自动喊到新人
package main

import "fmt"

// =============================================================================
// 以下 Subject / WeatherData / 旧看板：刻意与 step2 保持同构（稳定点）
// =============================================================================

// Observer 观察者：主题推送读数时的统一入口。
type Observer interface {
	Update(temp, humidity, pressure float64)
}

// Subject 主题：注册 / 注销 / 广播。
type Subject interface {
	RegisterObserver(o Observer)
	RemoveObserver(o Observer)
	NotifyObservers()
}

// WeatherData 具体主题（与 step2 相同，刻意不改）。
type WeatherData struct {
	observers   []Observer
	temperature float64
	humidity    float64
	pressure    float64
}

func NewWeatherData() *WeatherData {
	return &WeatherData{observers: make([]Observer, 0)}
}

func (w *WeatherData) RegisterObserver(o Observer) {
	w.observers = append(w.observers, o)
}

func (w *WeatherData) RemoveObserver(o Observer) {
	for i, obs := range w.observers {
		if obs == o {
			w.observers = append(w.observers[:i], w.observers[i+1:]...)
			return
		}
	}
}

// NotifyObservers 仍然是这段 for 循环——加新看板时这里一行都不用动。
func (w *WeatherData) NotifyObservers() {
	for _, o := range w.observers {
		o.Update(w.temperature, w.humidity, w.pressure)
	}
}

func (w *WeatherData) SetMeasurements(temp, humidity, pressure float64) {
	w.temperature = temp
	w.humidity = humidity
	w.pressure = pressure
	fmt.Println("--- 读数更新，NotifyObservers ---")
	w.NotifyObservers()
}

// CurrentConditionsDisplay 当前状况看板（与 step2 相同）。
type CurrentConditionsDisplay struct {
	temp     float64
	humidity float64
}

func NewCurrentConditionsDisplay(s Subject) *CurrentConditionsDisplay {
	d := &CurrentConditionsDisplay{}
	s.RegisterObserver(d)
	return d
}

func (d *CurrentConditionsDisplay) Update(temp, humidity, pressure float64) {
	d.temp = temp
	d.humidity = humidity
	fmt.Printf("  [当前状况] 温度=%.1f°C 湿度=%.1f%%\n", d.temp, d.humidity)
	_ = pressure
}

// StatisticsDisplay 统计看板（与 step2 相同）。
type StatisticsDisplay struct {
	maxTemp float64
	minTemp float64
	sumTemp float64
	count   int
}

func NewStatisticsDisplay(s Subject) *StatisticsDisplay {
	d := &StatisticsDisplay{}
	s.RegisterObserver(d)
	return d
}

func (d *StatisticsDisplay) Update(temp, humidity, pressure float64) {
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
	_, _ = humidity, pressure
}

// =============================================================================
// 本步唯一新增：ForecastDisplay（扩展点）
// =============================================================================

// ForecastDisplay 预报看板。
// 业务很简化：记住上次气压，和这次比较 → 升=晴好 / 降=阴雨 / 平=多云。
// 第一次还没有「上次」，固定报「多云」。
type ForecastDisplay struct {
	lastPressure float64 // 上一次收到的气压
	hasLast      bool    // 是否已经有过至少一次 Update
}

// NewForecastDisplay 与其它看板同一挂接方式：new → Register → 返回。
// WeatherData 无需增加字段，也无需改 NotifyObservers。
func NewForecastDisplay(s Subject) *ForecastDisplay {
	d := &ForecastDisplay{}
	s.RegisterObserver(d)
	return d
}

// Update 实现 Observer：根据气压变化打印预报，并记住本次气压供下次比较。
func (d *ForecastDisplay) Update(temp, humidity, pressure float64) {
	forecast := "多云"
	if d.hasLast {
		switch {
		case pressure > d.lastPressure:
			forecast = "晴好"
		case pressure < d.lastPressure:
			forecast = "阴雨"
			// pressure == last → 保持「多云」
		}
	}
	// 先算完再更新 last，避免和「本次」比。
	d.lastPressure = pressure
	d.hasLast = true
	fmt.Printf("  [预报] %s（气压=%.1f）\n", forecast, pressure)
	_, _ = temp, humidity
}

func main() {
	wd := NewWeatherData()

	// 旧看板：与 step2 一样挂上
	_ = NewCurrentConditionsDisplay(wd)
	_ = NewStatisticsDisplay(wd)

	// 新看板：只加这一行（外加上面 ForecastDisplay 类型定义）
	// 挂上后 observers = [current, stats, forecast]
	_ = NewForecastDisplay(wd)

	// 第一次：预报无历史 → 多云
	wd.SetMeasurements(26.5, 65, 1013.2)
	// 气压 1013.2 → 1012.8 下降 → 阴雨
	wd.SetMeasurements(28.0, 70, 1012.8)
	// 气压 1012.8 → 1014.0 上升 → 晴好
	wd.SetMeasurements(27.0, 68, 1014.0)
}
