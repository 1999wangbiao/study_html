// 单例模式可运行示范（Go）—— 应用配置中心
//
// 核心一句话：整个进程里只允许一个实例；所有人通过 GetInstance 拿到同一份。
//
// 本目录运行：
//
//	go run .
package main

import (
	"fmt"
	"sync"
)

// =============================================================================
// 一、单例：AppConfig（包级 instance + sync.Once）
// =============================================================================

// AppConfig 应用配置：进程内唯一。
type AppConfig struct {
	appName string
	port    int
}

var (
	configInstance *AppConfig
	configOnce     sync.Once
)

// GetInstance 返回全局唯一的 AppConfig；首次调用时初始化。
func GetInstance() *AppConfig {
	configOnce.Do(func() {
		fmt.Println("[Singleton] 初始化 AppConfig（整个进程只会走这里一次）")
		configInstance = &AppConfig{
			appName: "demo-app",
			port:    8080,
		}
	})
	return configInstance
}

// AppName 返回应用名。
func (c *AppConfig) AppName() string { return c.appName }

// Port 返回监听端口。
func (c *AppConfig) Port() int { return c.port }

// SetPort 修改端口（演示：改一处，所有持有者都看见）。
func (c *AppConfig) SetPort(p int) {
	c.port = p
	fmt.Printf("[AppConfig] 端口改为 %d\n", p)
}

// =============================================================================
// 二、客户端：多处获取，应是同一指针
// =============================================================================

func main() {
	fmt.Println("========== 两次获取 ==========")
	a := GetInstance()
	b := GetInstance()
	fmt.Printf("a 地址: %p\n", a)
	fmt.Printf("b 地址: %p\n", b)
	fmt.Printf("a == b: %v\n", a == b)
	fmt.Printf("当前: name=%s port=%d\n", a.AppName(), a.Port())

	fmt.Println("========== 改一处，另一处看见 ==========")
	a.SetPort(9090)
	fmt.Printf("b.Port() = %d（与 a 共享同一实例）\n", b.Port())

	fmt.Println("========== 并发下仍只初始化一次 ==========")
	// 重新演示 Once 语义：上面已经初始化过，这里再并发 Get 不会再次打印初始化日志。
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cfg := GetInstance()
			fmt.Printf("  goroutine-%d 拿到 %p port=%d\n", id, cfg, cfg.Port())
		}(i)
	}
	wg.Wait()
}
