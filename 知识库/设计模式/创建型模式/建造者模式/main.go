// 建造者模式可运行示范（Go）—— 组装电脑
//
// 核心一句话：同样的组装步骤，换 Builder 产出不同配置；也可用链式 WithXxx。
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、产品：组装完成后的电脑
// =============================================================================

// Computer 成品电脑。
type Computer struct {
	CPU    string
	Memory string
	Disk   string
}

// String 打印配置摘要。
func (c Computer) String() string {
	return fmt.Sprintf("CPU=%s, Memory=%s, Disk=%s", c.CPU, c.Memory, c.Disk)
}

// =============================================================================
// 二、建造者接口 + 具体建造者
// =============================================================================

// ComputerBuilder 声明组装步骤。
type ComputerBuilder interface {
	SetCPU()
	SetMemory()
	SetDisk()
	Build() Computer
}

// GamingBuilder 游戏主机配置。
type GamingBuilder struct {
	c Computer
}

// SetCPU 安装高性能 CPU。
func (b *GamingBuilder) SetCPU() { b.c.CPU = "8核高性能" }

// SetMemory 安装大内存。
func (b *GamingBuilder) SetMemory() { b.c.Memory = "32GB" }

// SetDisk 安装大容量 SSD。
func (b *GamingBuilder) SetDisk() { b.c.Disk = "1TB SSD" }

// Build 返回游戏主机成品。
func (b *GamingBuilder) Build() Computer { return b.c }

// OfficeBuilder 办公主机配置。
type OfficeBuilder struct {
	c Computer
}

// SetCPU 安装节能 CPU。
func (b *OfficeBuilder) SetCPU() { b.c.CPU = "4核节能" }

// SetMemory 安装常规内存。
func (b *OfficeBuilder) SetMemory() { b.c.Memory = "16GB" }

// SetDisk 安装中等容量 SSD。
func (b *OfficeBuilder) SetDisk() { b.c.Disk = "512GB SSD" }

// Build 返回办公主机成品。
func (b *OfficeBuilder) Build() Computer { return b.c }

// =============================================================================
// 三、导演：固定组装顺序
// =============================================================================

// Director 按统一流程调用建造步骤。
type Director struct{}

// Construct 执行 CPU → 内存 → 硬盘，返回成品。
func (d *Director) Construct(b ComputerBuilder) Computer {
	b.SetCPU()
	b.SetMemory()
	b.SetDisk()
	return b.Build()
}

// =============================================================================
// 四、链式建造者（Go 里更常见的写法，无 Director）
// =============================================================================

// FluentComputerBuilder 支持可选参数链式设置。
type FluentComputerBuilder struct {
	c Computer
}

// NewFluentComputerBuilder 创建链式建造者，并给默认值。
func NewFluentComputerBuilder() *FluentComputerBuilder {
	return &FluentComputerBuilder{
		c: Computer{CPU: "4核", Memory: "8GB", Disk: "256GB SSD"},
	}
}

// WithCPU 设置 CPU。
func (b *FluentComputerBuilder) WithCPU(cpu string) *FluentComputerBuilder {
	b.c.CPU = cpu
	return b
}

// WithMemory 设置内存。
func (b *FluentComputerBuilder) WithMemory(mem string) *FluentComputerBuilder {
	b.c.Memory = mem
	return b
}

// WithDisk 设置硬盘。
func (b *FluentComputerBuilder) WithDisk(disk string) *FluentComputerBuilder {
	b.c.Disk = disk
	return b
}

// Build 返回成品。
func (b *FluentComputerBuilder) Build() Computer { return b.c }

// =============================================================================
// 五、客户端
// =============================================================================

func main() {
	director := &Director{}

	fmt.Println("========== Director × 游戏主机 ==========")
	gaming := director.Construct(&GamingBuilder{})
	fmt.Printf("  %s\n", gaming)

	fmt.Println("========== Director × 办公主机 ==========")
	office := director.Construct(&OfficeBuilder{})
	fmt.Printf("  %s\n", office)

	fmt.Println("========== 链式 Builder（无 Director） ==========")
	custom := NewFluentComputerBuilder().
		WithCPU("6核").
		WithMemory("24GB").
		WithDisk("2TB SSD").
		Build()
	fmt.Printf("  %s\n", custom)
}
