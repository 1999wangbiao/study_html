package main

import "fmt"

// fanOut 扇出：将同一路输入广播到两个 channel（每个值两路各一份）
func fanOut(input <-chan int) (<-chan int, <-chan int) {
	out1 := make(chan int)
	out2 := make(chan int)

	go func() {
		defer close(out1)
		defer close(out2)
		for val := range input {
			out1 <- val
			out2 <- val
		}
	}()

	return out1, out2
}

// fanIn 扇入：将多路输入合并到单路输出；某路关闭后不再 select 该路
func fanIn(input1, input2 <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case val, ok := <-input1:
				if ok {
					out <- val
				} else {
					input1 = nil
				}
			case val, ok := <-input2:
				if ok {
					out <- val
				} else {
					input2 = nil
				}
			}
			if input1 == nil && input2 == nil {
				break
			}
		}
	}()
	return out
}

// double 将上游每个整数乘以 2
func double(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * 2
		}
	}()
	return out
}

// fanOutIn 演示：generate → fanOut → 平方 / 加倍 → fanIn → 打印
func fanOutIn() {
	numbers := generate()
	ch1, ch2 := fanOut(numbers)
	squares := square(ch1)
	doubles := double(ch2)
	merged := fanIn(squares, doubles)

	fmt.Println("fan-out / fan-in (order may vary):")
	for v := range merged {
		fmt.Println(v)
	}
}
