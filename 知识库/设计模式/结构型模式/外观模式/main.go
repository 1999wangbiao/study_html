// 外观模式可运行示范（Go）—— 家庭影院一键观影
//
// 核心一句话：客户端只调 WatchMovie / EndMovie；里面编排功放、播放器、投影、灯光。
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、子系统：各自只做一件事
// =============================================================================

// Amplifier 功放。
type Amplifier struct{}

// On 打开功放。
func (a *Amplifier) On() { fmt.Println("  [Amplifier] 开机") }

// Off 关闭功放。
func (a *Amplifier) Off() { fmt.Println("  [Amplifier] 关机") }

// SetVolume 设置音量。
func (a *Amplifier) SetVolume(level int) {
	fmt.Printf("  [Amplifier] 音量 = %d\n", level)
}

// StreamingPlayer 流媒体播放器。
type StreamingPlayer struct{}

// On 打开播放器。
func (p *StreamingPlayer) On() { fmt.Println("  [Player] 开机") }

// Off 关闭播放器。
func (p *StreamingPlayer) Off() { fmt.Println("  [Player] 关机") }

// Play 播放影片。
func (p *StreamingPlayer) Play(title string) {
	fmt.Printf("  [Player] 播放《%s》\n", title)
}

// Stop 停止播放。
func (p *StreamingPlayer) Stop() { fmt.Println("  [Player] 停止") }

// Projector 投影仪。
type Projector struct{}

// On 打开投影。
func (p *Projector) On() { fmt.Println("  [Projector] 开机") }

// Off 关闭投影。
func (p *Projector) Off() { fmt.Println("  [Projector] 关机") }

// WideScreen 切换宽屏模式。
func (p *Projector) WideScreen() { fmt.Println("  [Projector] 宽屏模式") }

// TheaterLights 影院灯光。
type TheaterLights struct{}

// Dim 调暗到指定亮度。
func (l *TheaterLights) Dim(percent int) {
	fmt.Printf("  [Lights] 调暗到 %d%%\n", percent)
}

// On 灯光全开。
func (l *TheaterLights) On() { fmt.Println("  [Lights] 全开") }

// =============================================================================
// 二、外观：对外只开简门
// =============================================================================

// HomeTheaterFacade 家庭影院外观。
type HomeTheaterFacade struct {
	amp       *Amplifier
	player    *StreamingPlayer
	projector *Projector
	lights    *TheaterLights
}

// NewHomeTheaterFacade 注入子系统并创建外观。
func NewHomeTheaterFacade(
	amp *Amplifier,
	player *StreamingPlayer,
	projector *Projector,
	lights *TheaterLights,
) *HomeTheaterFacade {
	return &HomeTheaterFacade{
		amp: amp, player: player, projector: projector, lights: lights,
	}
}

// WatchMovie 一键观影：按固定顺序编排子系统。
func (f *HomeTheaterFacade) WatchMovie(title string) {
	fmt.Println("[Facade] 准备观影…")
	f.lights.Dim(10)
	f.projector.On()
	f.projector.WideScreen()
	f.amp.On()
	f.amp.SetVolume(5)
	f.player.On()
	f.player.Play(title)
	fmt.Println("[Facade] 可以开始看了")
}

// EndMovie 一键收场：逆序关闭。
func (f *HomeTheaterFacade) EndMovie() {
	fmt.Println("[Facade] 结束观影…")
	f.player.Stop()
	f.player.Off()
	f.amp.Off()
	f.projector.Off()
	f.lights.On()
	fmt.Println("[Facade] 设备已收好")
}

// =============================================================================
// 三、客户端：只依赖外观
// =============================================================================

func main() {
	facade := NewHomeTheaterFacade(
		&Amplifier{},
		&StreamingPlayer{},
		&Projector{},
		&TheaterLights{},
	)

	fmt.Println("========== 一键观影 ==========")
	facade.WatchMovie("Inception")

	fmt.Println("========== 一键收场 ==========")
	facade.EndMovie()
}
