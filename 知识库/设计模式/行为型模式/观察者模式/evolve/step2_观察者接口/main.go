// 演变第 2 步：抽出 Observer，Subject 只维护列表并广播
//
// 运行：
//
//	cd evolve/step2_观察者接口
//	go run .
//
// =============================================================================
// 相对 step1，改了什么（对照看）
// =============================================================================
//
// step1：
//	WeatherData 字段：current *CurrentConditionsDisplay, stats *StatisticsDisplay
//	通知：w.current.Update(...); w.stats.Update(...)
//
// step2：
//	WeatherData 字段：observers []Observer   ← 只认接口，不认具体看板
//	通知：for _, o := range w.observers { o.Update(...) }
//	看板：自己实现 Update，构造时调用 RegisterObserver 挂上去
//
// 调用链（松耦合）：
//
//	main
//	  ├─ NewCurrentConditionsDisplay(wd)  → RegisterObserver(看板)
//	  ├─ NewStatisticsDisplay(wd)         → RegisterObserver(看板)
//	  └─ SetMeasurements(...)
//	       ├─ 存三个数到 WeatherData
//	       └─ NotifyObservers()
//	            └─ 遍历 observers，每个 o.Update(...)   ← 循环里没有具体类型名
//
// 本步额外演示：RemoveObserver 后，再更新就只剩仍挂着的看板。
// 新看板不必改 WeatherData —— 到 step3 用 ForecastDisplay 验证。
package main

import "fmt"

// =============================================================================
// 一、抽象：Observer / Subject（step1 没有这两层）
// =============================================================================

// Observer 观察者约定：主题只通过这个接口推数据，不关心你是哪块看板。
// Update 的三个参数是 Push 模式：主题把最新读数直接塞过来。
type Observer interface {
	Update(temp, humidity, pressure float64)
}

// Subject 主题约定：谁想听，就来登记；不想听，就注销；数据变了，主题负责喊一声。
type Subject interface {
	RegisterObserver(o Observer)
	RemoveObserver(o Observer)
	NotifyObservers()
}

// =============================================================================
// 二、具体主题 WeatherData
// =============================================================================

// WeatherData 只存「读数 + 观察者列表」。
// 对比 step1：再也没有 *CurrentConditionsDisplay / *StatisticsDisplay 字段。
type WeatherData struct {
	observers   []Observer // 关键：依赖抽象，不依赖具体看板
	temperature float64
	humidity    float64
	pressure    float64
}

// NewWeatherData 只建空列表；看板由外部创建后再 Register，主题不再「内建」看板。
func NewWeatherData() *WeatherData {
	return &WeatherData{observers: make([]Observer, 0)}
}

// RegisterObserver 把观察者挂到列表末尾。之后每次 Notify 都会喊到它。
func (w *WeatherData) RegisterObserver(o Observer) {
	w.observers = append(w.observers, o)
}

// RemoveObserver 按指针从列表摘掉（接口值比较的是动态类型+指针）。
// 摘掉后，后续 Notify 不再调用它 —— step1 的硬字段做不到这点。
func (w *WeatherData) RemoveObserver(o Observer) {
	for i, obs := range w.observers {
		if obs == o {
			// 删掉下标 i：拼前后两段
			w.observers = append(w.observers[:i], w.observers[i+1:]...)
			return
		}
	}
}

// NotifyObservers 广播：主题唯一的「通知逻辑」。
// 循环变量类型是 Observer，编译器在这里看不到任何具体看板名。
func (w *WeatherData) NotifyObservers() {
	for _, o := range w.observers {
		// Push：把当前读数推给每一个观察者；观察者内部自己决定怎么展示
		o.Update(w.temperature, w.humidity, w.pressure)
	}
}

// SetMeasurements 逻辑仍是「先存再通知」，但通知改走 NotifyObservers。
// 对比 step1 的 measurementsChanged：那里是点名，这里是遍历列表。
func (w *WeatherData) SetMeasurements(temp, humidity, pressure float64) {
	w.temperature = temp
	w.humidity = humidity
	w.pressure = pressure
	fmt.Println("--- 读数更新，NotifyObservers ---")
	w.NotifyObservers()
}

// =============================================================================
// 三、具体观察者：实现 Observer，并在构造时自己注册
// =============================================================================

// CurrentConditionsDisplay 当前状况看板。
type CurrentConditionsDisplay struct {
	temp     float64
	humidity float64
}

// NewCurrentConditionsDisplay 创建看板并立刻挂到主题上。
// 逻辑：看板主动订阅，而不是像 step1 那样由主题 new 出来再握着。
func NewCurrentConditionsDisplay(s Subject) *CurrentConditionsDisplay {
	d := &CurrentConditionsDisplay{}
	s.RegisterObserver(d) // d 满足 Observer（有 Update 方法）
	return d
}

// Update 实现 Observer：缓存并打印当前温湿度。
func (d *CurrentConditionsDisplay) Update(temp, humidity, pressure float64) {
	d.temp = temp
	d.humidity = humidity
	fmt.Printf("  [当前状况] 温度=%.1f°C 湿度=%.1f%%\n", d.temp, d.humidity)
	_ = pressure
}

// StatisticsDisplay 统计看板（业务逻辑与 step1 相同，变的是挂接方式）。
type StatisticsDisplay struct {
	maxTemp float64
	minTemp float64
	sumTemp float64
	count   int
}

// NewStatisticsDisplay 同样：创建 → Register → 返回，供 main 需要时 Remove。
func NewStatisticsDisplay(s Subject) *StatisticsDisplay {
	d := &StatisticsDisplay{}
	s.RegisterObserver(d)
	return d
}

// Update 实现 Observer：累加温度统计并打印。
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

func main() {
	wd := NewWeatherData()

	// 挂接：看板构造时 Register，主题列表变成 [current, stats]
	_ = NewCurrentConditionsDisplay(wd)
	stats := NewStatisticsDisplay(wd)

	// 两次更新：两块看板都会收到 Update
	wd.SetMeasurements(26.5, 65, 1013.2)
	wd.SetMeasurements(28.0, 70, 1012.8)

	// 注销演示：从列表摘掉 stats 后，再更新只剩「当前状况」
	fmt.Println("--- 注销统计看板后再更新 ---")
	wd.RemoveObserver(stats)
	wd.SetMeasurements(24.0, 60, 1014.0)
}
