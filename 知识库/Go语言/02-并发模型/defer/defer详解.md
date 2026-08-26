# Go 语言 defer 详解

> **defer = 延迟执行函数**（关闭资源、捕获异常、记录耗时）。**每个 Goroutine 维护一个 defer 链表，按先进后出（LIFO）顺序执行**。无论函数正常返回还是 panic，defer 都会执行。

---

## 一、defer 是什么、怎么用

`defer` 用于注册一个延迟执行的函数：当前函数**返回前**（无论正常返回还是 panic）执行。典型用途：

```go
// 1. 关闭资源
f, err := os.Open(path)
defer f.Close() // 函数结束自动关闭

// 2. 解锁
mu.Lock()
defer mu.Unlock() // 防止忘记解锁

// 3. 捕获 panic
defer func() {
    if r := recover(); r != nil {
        fmt.Println("recovered:", r)
    }
}()

// 4. 记录函数耗时
defer func(start time.Time) { fmt.Println("耗时:", time.Since(start)) }(time.Now())
```

---

## 二、底层实现

### defer 链表

多个 defer 通过 link 串联成**单向链表**，每个 Goroutine 维护一个 defer 链表。`_defer` 实例包含延迟函数指针（`fn`）、参数、是否已执行（`started`）等。

### 执行时机

当包含 defer 的函数即将退出时（无论正常返回还是 panic），运行时会触发 defer 执行：

1. **遍历 defer 链表**：从当前 Goroutine 的 defer 链表头部开始，依次取出 `_defer` 实例；
2. **执行延迟函数**：调用 `_defer.fn` 执行延迟函数，同时标记 `started=true`（避免重复执行）；
3. **清理链表**：执行完一个 `_defer` 后，将其从链表中移除，继续遍历下一个，直到链表为空。

### panic 时的行为

若函数因 panic 退出：**defer 执行完成后，panic 会继续向上传播**（除非 defer 中调用 `recover()` 捕获异常）。

### 执行顺序

- **链表结构**：多个 defer 以**头插法**形成单向链表，注册顺序与链表顺序相反；
- **执行顺序**：函数退出时从链表头部开始执行，即**后进先出（LIFO）**，与声明顺序相反。

```go
defer fmt.Println("1")
defer fmt.Println("2")
defer fmt.Println("3")
// 输出：3 2 1（后注册先执行）
```

---

## 三、易错点

### 1. 参数立即求值，闭包延迟求值

```go
i := 0
defer fmt.Println(i)            // 普通调用，参数立即求值 → 输出 0
defer func() { fmt.Println(i) }() // 闭包，延迟到 return 时执行 → 输出 100
i = 100
```

### 2. defer 与 return 的求值顺序

return 语句会**先给返回值赋值，再执行 defer**。经典陷阱：

```go
func f() (x int) {
    defer func() { x++ }() // defer 修改命名返回值
    return 1               // 先 x = 1，再 defer x++ → 返回 2
}
// f() == 2
```

### 3. 循环里 defer 资源

循环内 `defer f.Close()` 会在**整个函数结束**时才执行，导致资源堆积。应把循环体包进函数：

```go
for _, file := range files {
    func() {
        f, _ := os.Open(file)
        defer f.Close() // 每个匿名函数结束时关闭
    }()
}
```

---

## 四、高频自测

1. defer 的执行顺序？
   → **后进先出（LIFO）**。
2. defer 什么时候执行？
   → **函数返回前**，正常返回和 panic 都会执行。
3. `defer fmt.Println(i)` 和 `defer func(){fmt.Println(i)}()` 一样吗？
   → **不一样**，前者立即求值，后者延迟求值（见闭包篇）。
4. defer 能捕获 panic 吗？
   → **配合 recover 可以**。

---

## 五、一句话总结

> **defer = 延迟执行 + 资源清理 + 异常捕获，Goroutine 级 defer 链表按 LIFO 执行**。记住三件事：后注册先执行、参数立即求值（闭包才延迟）、return 先赋值再执行 defer。
