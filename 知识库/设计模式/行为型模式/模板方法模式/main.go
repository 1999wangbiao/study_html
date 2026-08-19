// 模板方法模式可运行示范（Go）—— 咖啡与茶
//
// 核心一句话：把算法的「步骤骨架」固定在一个地方，
// 具体某一步怎么做（冲泡什么、加什么料）留给各子类覆写。
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、角色 1：AbstractClass（抽象类）——「步骤骨架」
// =============================================================================
//
// 固定不变的步骤（烧水、倒杯）写在这里；可变步骤通过 drink 接口回调到子类。

// Beverage 可变步骤的约定：冲泡 + 加料 + 是否加料。
type Beverage interface {
	Brew()
	AddCondiments()
	WantsCondiments() bool
}

// CaffeineBeverage 含咖啡因饮料模板：掌握「准备一杯饮品」的骨架。
type CaffeineBeverage struct {
	drink Beverage
}

// PrepareRecipe 模板方法：步骤顺序固定，第 2、4 步委托给具体饮品。
func (b *CaffeineBeverage) PrepareRecipe() {
	b.BoilWater()                  // 第 1 步：固定
	b.drink.Brew()                 // 第 2 步：交给具体饮品
	b.PourInCup()                  // 第 3 步：固定
	if b.drink.WantsCondiments() { // 钩子：子类可决定是否执行加料
		b.drink.AddCondiments() // 第 4 步：交给具体饮品
	}
}

// BoilWater 烧水：所有饮品的共同步骤。
func (b *CaffeineBeverage) BoilWater() { fmt.Println("  1. 烧开水") }

// PourInCup 倒入杯中：所有饮品的共同步骤。
func (b *CaffeineBeverage) PourInCup() { fmt.Println("  3. 倒入杯中") }

// =============================================================================
// 二、角色 2：ConcreteClass（具体子类）——「覆写可变步骤」
// =============================================================================
//
// Go 没有类继承：用「嵌入模板 + 构造时把自己的引用塞进 drink 字段」，
// 让模板方法里的 drink.Brew() 能派发到子类自己的实现。

// Coffee 咖啡。
type Coffee struct {
	*CaffeineBeverage
}

// NewCoffee 创建咖啡：把自身当作 drink，模板才能回调到子类覆写。
func NewCoffee() *Coffee {
	c := &Coffee{}
	c.CaffeineBeverage = &CaffeineBeverage{drink: c}
	return c
}

func (c *Coffee) Brew()                 { fmt.Println("  2. 冲泡咖啡粉，过滤") }
func (c *Coffee) AddCondiments()        { fmt.Println("  4. 加糖和牛奶") }
func (c *Coffee) WantsCondiments() bool { return true }

// Tea 茶。
type Tea struct {
	*CaffeineBeverage
}

// NewTea 创建茶。
func NewTea() *Tea {
	t := &Tea{}
	t.CaffeineBeverage = &CaffeineBeverage{drink: t}
	return t
}

func (t *Tea) Brew()                 { fmt.Println("  2. 浸泡茶叶 3 分钟") }
func (t *Tea) AddCondiments()        { fmt.Println("  4. 加柠檬切片") }
func (t *Tea) WantsCondiments() bool { return false } // 钩子：不加料

// =============================================================================
// 三、main：两条「同一骨架」的流程
// =============================================================================

func main() {
	fmt.Println("===== 冲一杯咖啡 =====")
	coffee := NewCoffee()
	coffee.PrepareRecipe()

	fmt.Println()
	fmt.Println("===== 泡一杯茶 =====")
	tea := NewTea()
	tea.PrepareRecipe()

	fmt.Println()
	fmt.Println("========== 读懂输出后你会发现 ==========")
	fmt.Println("1. 步骤骨架（烧水、倒杯）只写了一次")
	fmt.Println("2. 咖啡 / 茶 只写自己不同的那一两步")
	fmt.Println("3. 茶通过钩子 WantsCondiments()=false 跳过了加料")
	fmt.Println("4. 新增一种饮品 = 新写一个类型，不用动骨架")
}
