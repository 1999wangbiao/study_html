# sync.WaitGroup 底层实现

> **WaitGroup = 计数器 + 信号量**：`Add` 加计数、`Done` 减计数、`Wait` 阻塞到计数归零。底层用原子操作维护计数器和等待者数量，最后一个 `Done` 归零时唤醒所有等待者。

---

## 一、核心原理

WaitGroup 内部维护一个**计数器（count）**，通过三个核心方法协同工作：

- **`Add(n int)`**：增加计数器的值（n 通常为正，代表要等待的 goroutine 数量）；
- **`Done()`**：计数器减 1（等价于 `Add(-1)`，每个子 goroutine 结束时调用）；
- **`Wait()`**：阻塞当前 goroutine，直到计数器变为 0。

```go
func main() {
    var wg sync.WaitGroup
    wg.Add(2) // 注册 2 个待等待的 goroutine

    go func() {
        defer wg.Done() // 完成后计数器减 1
        fmt.Println("子任务 1 完成")
    }()

    go func() {
        defer wg.Done()
        fmt.Println("子任务 2 完成")
    }()

    wg.Wait() // 阻塞，直到两个子任务都调用 Done()
    fmt.Println("所有任务完成")
}
```

---

## 二、底层实现细节

WaitGroup 的结构体（简化版）包含三个关键字段：

```go
type WaitGroup struct {
    noCopy noCopy    // 避免值拷贝（通过编译期检查）
    state1 [3]uint32 // 存储计数器和信号量状态
}
```

其中 `state1` 通过位运算划分出两个核心信息：

- **计数器（count）**：表示未完成的 goroutine 数量（state1[0]）；
- **等待者数量（waiters）**：表示调用 `Wait()` 阻塞的 goroutine 数量（state1[1]）；
- **信号量（semaphore）**：基于操作系统信号量实现，用于阻塞/唤醒等待的 goroutine（state1[2] 相关）。

---

## 三、工作流程

### 1. 初始化与添加任务

主 goroutine 调用 `wg.Add(n)`，将计数器增加 n（如启动 3 个子 goroutine 则 Add(3)）。

### 2. 子 goroutine 执行与结束

每个子 goroutine 执行完毕前调用 `wg.Done()`，将计数器减 1。

### 3. 等待所有任务完成

主 goroutine 调用 `wg.Wait()` 后：

- 若计数器已为 0，**直接返回**（无需等待）；
- 若计数器 > 0，自己加入**等待队列**（等待者数量 +1），并通过**信号量阻塞**。

### 4. 唤醒等待者

当**最后一个子 goroutine** 调用 `Done()` 使计数器变为 0 时：

- 若有等待者（waiters > 0），则**释放信号量，唤醒所有阻塞的等待者**（通常是主 goroutine）。

---

## 四、关键特性

1. **不可复制**：`noCopy` 字段通过编译期检查阻止 WaitGroup 被拷贝（避免计数器状态混乱）；
2. **原子操作**：计数器的增减和状态检查均通过**原子操作（atomic 包）**实现，保证并发安全；
3. **一次性使用**：WaitGroup 计数器归 0 后，若再次调用 `Add()` 可能导致不可预期的行为（建议用完即丢弃，不重复使用）。

---

## 五、高频自测

1. WaitGroup 的三个方法分别做什么？
   → **Add 加计数、Done 减计数、Wait 阻塞到归零**。
2. Wait 阻塞时是怎么被唤醒的？
   → **最后一个 Done 把计数器归零时，释放信号量唤醒所有等待者**。
3. 为什么 WaitGroup 不能被复制？
   → **noCopy 编译期检查**，避免计数器状态混乱。

---

## 六、一句话总结

> **WaitGroup = 原子计数器（count）＋ 等待者计数（waiters）＋ 信号量**：Add/Done 维护计数，Wait 挂起等待者，最后一个 Done 归零时发信号唤醒所有人。核心一句话：**计数器归 0 时唤醒所有等待者**。
