package main

import (
	"fmt"
	"time"
)

const numWorkers = 3

// Job 表示待处理任务
type Job struct {
	ID    int
	Value int
}

// Result 表示任务处理结果
type Result struct {
	JobID int
	Value int
}

// worker 从 jobs 取任务，计算后写入 results，直到 jobs 被关闭且耗尽
func worker(jobs <-chan Job, results chan<- Result) {
	for job := range jobs {
		time.Sleep(50 * time.Millisecond) // 模拟耗时操作
		results <- Result{
			JobID: job.ID,
			Value: job.Value * job.Value,
		}
	}
}

// processResult 消费单条结果（此处打印）
func processResult(result Result) {
	fmt.Printf("job %d => %d\n", result.JobID, result.Value)
}

// demoJobs 构造示例任务列表
func demoJobs() []Job {
	jobs := make([]Job, 0, 10)
	for i := 0; i < 10; i++ {
		jobs = append(jobs, Job{ID: i, Value: i})
	}
	return jobs
}

// workerPool 任务分发和结果收集
func workerPool() {
	allJobs := demoJobs()

	jobs := make(chan Job, 100)
	results := make(chan Result, 100)

	// 启动固定数量的 worker
	for i := 0; i < numWorkers; i++ {
		go worker(jobs, results)
	}

	// 发送任务
	for _, job := range allJobs {
		jobs <- job
	}
	close(jobs)

	// 收集结果（任务数与结果数一致）
	for i := 0; i < len(allJobs); i++ {
		result := <-results
		processResult(result)
	}
}
