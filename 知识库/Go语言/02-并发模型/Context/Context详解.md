# Go 语言 Context

> **Context = 树形结构 + 继承派生**：可以从一个 context 派生出多个子 context，每个子 context 有自己的取消函数和超时时间。**父取消 → 递归取消所有子**（"一父死，全子亡"）。用途：控制协程退出、超时控制、传递元数据。

---

## 一、Context 的核心原理

Context 的设计基于**树形结构和接口继承**，核心思想：

- 每个 Context 实例可以派生出**子 Context**（通过 `WithCancel`、`WithTimeout` 等方法），形成**父子关系**；
- 当父 Context 被取消时，所有子 Context 会被**递归取消**，实现"一父死，全子亡"的生命周期管理；
- 可以从一个 context 生成多个 context，每个子 context **都有自己的取消函数和超时时间**（可组合性）。

```go
parent := context.Background()                       // 根
ctx1, cancel1 := context.WithCancel(parent)          // 子1：可手动取消
ctx2, cancel2 := context.WithTimeout(parent, 2*time.Second) // 子2：2 秒超时
```

---

## 二、取消传播机制

当父 Context 被取消时（调用取消函数或超时），触发以下流程：

1. **关闭自身的 `Done()` 通道**（通知依赖它的 Goroutine 退出）；
2. **递归取消所有子 Context**（调用子 Context 的取消函数）；
3. **清除父与子的关联**（避免内存泄漏）。

这种机制确保了**所有关联的 Goroutine 能被统一、及时地终止**。

---

## 三、使用场景

### 1. 控制 Goroutine 退出（避免泄漏）

当一个请求需要启动多个 Goroutine 协作（如并行查询多个接口），可用 Context 确保所有 Goroutine 在请求结束（或超时）时及时退出，避免资源泄漏：

```go
ctx, cancel := context.WithCancel(context.Background())

go func() {
    for {
        select {
        case task := <-taskCh:
            handle(task)
        case <-ctx.Done(): // 监听取消
            return
        }
    }
}()

// 请求结束
cancel() // 通知所有子 Goroutine 退出
```

### 2. 超时控制

对外部服务调用（HTTP 请求、数据库查询）设置超时时间，防止长时间阻塞：

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

reply, err := client.SayHello(ctx, req) // 超时自动中断
```

### 3. 传递请求元数据

在分布式系统中，一个请求可能经过多个函数或服务，Context 可传递**请求 ID、用户认证信息**等，无需修改函数参数列表：

```go
ctx = context.WithValue(ctx, "request_id", reqID)
ctx = context.WithValue(ctx, "user", userInfo)

// 后续函数通过 ctx.Value("request_id") 读取
```

---

## 四、四种派生方式

| 方法 | 作用 |
|------|------|
| `WithCancel` | 手动取消 |
| `WithTimeout` | 超时自动取消（毫秒/秒） |
| `WithDeadline` | 指定截止时间点取消 |
| `WithValue` | 携带请求元数据 |

---

## 五、高频自测

1. 父 context 取消时子 context 会怎样？
   → **递归全部取消**（"一父死，全子亡"）。
2. Context 的三个主要用途？
   → **控制协程退出、超时控制、传递元数据**。
3. 怎么监听取消信号？
   → **`<-ctx.Done()`**。
4. 不监听 ctx.Done() 会怎样？
   → **Goroutine 在取消后仍运行，可能泄漏**。

---

## 六、一句话总结

> **Context = 树形派生 + 取消传播**：一个父可派生多个子（各带 cancel/超时），父取消子全亡；用于**退出协程、超时、传元数据**。记住监听 `ctx.Done()` 才能被优雅终止。
