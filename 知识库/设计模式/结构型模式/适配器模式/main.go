// 适配器模式可运行示范（Go）—— 媒体播放器转接头
//
// 核心一句话：客户端只认 MediaPlayer.Play；mp4/vlc 经适配器翻译成第三方调用。
//
// 本目录运行：
//
//	go run .
package main

import (
	"fmt"
	"strings"
)

// =============================================================================
// 一、目标接口：客户端期望的样子
// =============================================================================

// MediaPlayer 统一播放接口（Target）。
type MediaPlayer interface {
	Play(filename string)
}

// =============================================================================
// 二、被适配者：第三方已有接口（不能改）
// =============================================================================

// AdvancedMediaPlayer 高级格式播放器（Adaptee 侧）。
type AdvancedMediaPlayer interface {
	PlayMP4(filename string)
	PlayVLC(filename string)
}

// MP4Player 只能播 MP4。
type MP4Player struct{}

// PlayMP4 播放 MP4。
func (p *MP4Player) PlayMP4(filename string) {
	fmt.Printf("  [MP4Player] 播放 MP4: %s\n", filename)
}

// PlayVLC MP4 播放器不支持 VLC（空实现占位）。
func (p *MP4Player) PlayVLC(filename string) {}

// VLCPlayer 只能播 VLC。
type VLCPlayer struct{}

// PlayVLC 播放 VLC。
func (p *VLCPlayer) PlayVLC(filename string) {
	fmt.Printf("  [VLCPlayer] 播放 VLC: %s\n", filename)
}

// PlayMP4 VLC 播放器不支持 MP4（空实现占位）。
func (p *VLCPlayer) PlayMP4(filename string) {}

// =============================================================================
// 三、适配器：把第三方接口翻译成 MediaPlayer
// =============================================================================

// MediaAdapter 媒体适配器（Adapter）。
type MediaAdapter struct {
	advanced AdvancedMediaPlayer
}

// NewMediaAdapter 按文件扩展名选择被适配的第三方播放器。
func NewMediaAdapter(audioType string) *MediaAdapter {
	switch strings.ToLower(audioType) {
	case "mp4":
		return &MediaAdapter{advanced: &MP4Player{}}
	case "vlc":
		return &MediaAdapter{advanced: &VLCPlayer{}}
	default:
		return nil
	}
}

// Play 实现 Target：内部转发到 Adaptee 的专用方法。
func (a *MediaAdapter) Play(filename string) {
	ext := extOf(filename)
	switch ext {
	case "mp4":
		a.advanced.PlayMP4(filename)
	case "vlc":
		a.advanced.PlayVLC(filename)
	}
}

// =============================================================================
// 四、客户端封装：对外仍是 MediaPlayer
// =============================================================================

// AudioPlayer 音频播放器：mp3 自己播，其它格式交给适配器。
type AudioPlayer struct{}

// Play 按扩展名播放；非 mp3 走适配器。
func (p *AudioPlayer) Play(filename string) {
	ext := extOf(filename)
	switch ext {
	case "mp3":
		fmt.Printf("  [AudioPlayer] 直接播放 MP3: %s\n", filename)
	case "mp4", "vlc":
		adapter := NewMediaAdapter(ext)
		fmt.Printf("[Adapter] 使用适配器播放 %s\n", ext)
		adapter.Play(filename)
	default:
		fmt.Printf("[AudioPlayer] 不支持的格式: %s\n", ext)
	}
}

// extOf 返回小写扩展名（无点）。
func extOf(filename string) string {
	i := strings.LastIndex(filename, ".")
	if i < 0 || i == len(filename)-1 {
		return ""
	}
	return strings.ToLower(filename[i+1:])
}

func main() {
	player := &AudioPlayer{}

	fmt.Println("========== MP3：无需适配 ==========")
	player.Play("alone.mp3")

	fmt.Println("========== MP4：经适配器 → MP4Player ==========")
	player.Play("movie.mp4")

	fmt.Println("========== VLC：经适配器 → VLCPlayer ==========")
	player.Play("clip.vlc")

	fmt.Println("========== 未知格式 ==========")
	player.Play("notes.avi")
}
