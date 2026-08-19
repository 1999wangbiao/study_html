package main

import "fmt"

// pipeline 串联生成、平方、输出三阶段
func pipeline() {
    // 阶段1：数据生成
    numbers := generate()

    // 阶段2：数据处理
    squares := square(numbers)

    // 阶段3：数据输出
    output(squares)
}

// generate 在后台 goroutine 中向 channel 写入 0..9
func generate() <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for i := 0; i < 10; i++ {
            out <- i
        }
    }()
    return out
}

// square 从上游读整数，将平方写入下游 channel
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * n
        }
    }()
    return out
}

// output 消费最终结果并打印
func output(in <-chan int) {
    for n := range in {
        fmt.Println(n)
	}
}