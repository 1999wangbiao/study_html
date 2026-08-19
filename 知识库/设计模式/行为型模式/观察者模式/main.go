// 观察者模式可运行示范（Go）—— 气象站 × 多看板（最终形态 ≈ step3）
//
// 建议先按顺序跑演变，对照注释里的调用链：
//
//	evolve/step1_写死通知看板
//	  → SetMeasurements → measurementsChanged → 点名 current/stats.Update
//	evolve/step2_观察者接口
//	  → Register 挂列表 → SetMeasurements → NotifyObservers 循环 Update
//	evolve/step3_只加新看板
//	  → 主题代码不动，只加 ForecastDisplay 并 Register（本文件同形态）
//
// 最终形态调用链：
//
//	main
//	  ├─ NewXxxDisplay(wd) × 3  → 各自 RegisterObserver
//	  ├─ SetMeasurements(...)   → 存数 → NotifyObservers → 各看板 Update
//	  └─ RemoveObserver(stats)  → 再 SetMeasurements 时统计不再收到
//
// 一句话：主题变了，观察者各自更新；主题不认识具体看板类型。
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、角色：Observer / Subject
// =============================================================================

// Observer 观察者：被推送温湿度气压后刷新自己的展示（Push）。
type Observer interface {
	Update(temp, humidity, pressure float64)
}

// Subject 主题：管理观察者的注册、注销与通知。
type Subject interface {
	RegisterObserver(o Observer)
	RemoveObserver(o Observer)
	NotifyObservers()
}

// =============================================================================
// 二、具体主题：WeatherData
// =============================================================================

// WeatherData 气象数据源：存读数 + []Observer；变化时广播。
// 对比 step1：这里没有 *CurrentConditionsDisplay 之类的具体字段。
type WeatherData struct {
	observers   []Observer
	temperature float64
	humidity    float64
	pressure    float64
}

// NewWeatherData 创建空观察者列表的气象站；看板由外部 Register 挂上。
func NewWeatherData() *WeatherData {
	return &WeatherData{observers: make([]Observer, 0)}
}

// RegisterObserver 将观察者加入列表，之后每次 Notify 都会喊到。
func (w *WeatherData) RegisterObserver(o Observer) {
	w.observers = append(w.observers, o)
}

// RemoveObserver 按指针从列表摘掉；摘掉后不再收到 Update。
func (w *WeatherData) RemoveObserver(o Observer) {
	for i, obs := range w.observers {
		if obs == o {
			w.observers = append(w.observers[:i], w.observers[i+1:]...)
			return
		}
	}
}

// NotifyObservers 遍历列表 Push 当前读数；循环里只有 Observer，无具体看板名。
func (w *WeatherData) NotifyObservers() {
	for _, o := range w.observers {
		o.Update(w.temperature, w.humidity, w.pressure)
	}
}

// SetMeasurements 先更新内部读数，再 NotifyObservers。
func (w *WeatherData) SetMeasurements(temp, humidity, pressure float64) {
	w.temperature = temp
	w.humidity = humidity
	w.pressure = pressure
	fmt.Println("--- 读数更新，NotifyObservers ---")
	w.NotifyObservers()
}

// =============================================================================
// 三、具体观察者：当前状况 / 统计 / 预报
// =============================================================================

// CurrentConditionsDisplay 当前温湿度看板。
type CurrentConditionsDisplay struct {
	temp     float64
	humidity float64
}

// NewCurrentConditionsDisplay 创建并 Register 到主题。
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

// StatisticsDisplay 温度统计看板。
type StatisticsDisplay struct {
	maxTemp float64
	minTemp float64
	sumTemp float64
	count   int
}

// NewStatisticsDisplay 创建并 Register 到主题。
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

// ForecastDisplay 简易预报：气压相对上次升→晴好、降→阴雨、平/首次→多云。
type ForecastDisplay struct {
	lastPressure float64
	hasLast      bool
}

// NewForecastDisplay 创建并 Register；加此类型无需改 WeatherData（见 step3）。
func NewForecastDisplay(s Subject) *ForecastDisplay {
	d := &ForecastDisplay{}
	s.RegisterObserver(d)
	return d
}

func (d *ForecastDisplay) Update(temp, humidity, pressure float64) {
	forecast := "多云"
	if d.hasLast {
		switch {
		case pressure > d.lastPressure:
			forecast = "晴好"
		case pressure < d.lastPressure:
			forecast = "阴雨"
		}
	}
	d.lastPressure = pressure
	d.hasLast = true
	fmt.Printf("  [预报] %s（气压=%.1f）\n", forecast, pressure)
	_, _ = temp, humidity
}

func main() {
	wd := NewWeatherData()
	// 挂接三块看板 → observers 列表变长；主题代码零修改
	_ = NewCurrentConditionsDisplay(wd)
	stats := NewStatisticsDisplay(wd)
	_ = NewForecastDisplay(wd)

	wd.SetMeasurements(26.5, 65, 1013.2)
	wd.SetMeasurements(28.0, 70, 1012.8)
	wd.SetMeasurements(27.0, 68, 1014.0)

	// 注销统计后，再更新只剩当前状况 + 预报
	fmt.Println("--- 注销统计看板后再更新 ---")
	wd.RemoveObserver(stats)
	wd.SetMeasurements(25.0, 62, 1013.5)
}
