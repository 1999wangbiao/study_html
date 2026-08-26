# Go 中的 select 语句

> **select 专门处理多路 channel 收发：监听多个通道，哪个可操作就执行哪个分支；多个同时可操作时随机选一个（防止饿死）**。是超时控制、退出通知、多事件监听的核心工具。

---

## 一、select 是什么

`select` 是用于处理多个通道（channel）操作的控制结构，类似于 switch，但**专门用于通道的读写操作**。它能同时等待多个通道的操作，当其中一个通道可操作时，就执行对应的分支；若多个通道可操作，则**随机选择一个执行**，实现 Goroutine 间的高效协作。

select 是 Go 并发模型中连接多个通道的关键工具，广泛用于**超时控制、退出通知、多事件监听**等场景。

```go
select {
case v := <-ch1:      // ch1 有数据
    handle(v)
case ch2 <- msg:      // ch2 可发送
    send(msg)
case <-time.After(time.Second): // 超时
    fmt.Println("timeout")
case <-ctx.Done():    // 退出通知
    return
default:              // 所有分支都不可用时立即执行
    fmt.Println("no case ready")
}
```

---

## 二、执行流程

1. **监听通道状态**：评估所有 case 的 channel 是否可操作（可接收 / 可发送）；
2. **若有多个 case 可执行**：通过**伪随机算法**选择一个分支执行（确保公平性，避免某个通道被饿死）；
3. **若只有一个 case 可执行**：直接执行该分支；
4. **若都不满足且没有 default**：阻塞等待，直到某个 channel 可操作；
5. **若有 default**：都不满足时立即执行 default（非阻塞 select）。

---

## 三、关键点

- **随机选择**：多个 case 同时就绪时随机选一个，**防止某个通道被饿死**；
- **空 select**：`select {}` 会永久阻塞（常用于让 main 常驻）；
- **nil channel 永远阻塞**：select 中的 `nil` channel 分支永远不会被选中，可用于动态禁用分支；
- **select 不阻塞**：配合 `default` 可实现非阻塞收发。

---

## 四、典型使用场景

### 1. 超时控制

```go
select {
case res := <-ch:
    fmt.Println(res)
case <-time.After(2 * time.Second):
    fmt.Println("操作超时")
}
```

### 2. 退出通知（配合 context）

```go
select {
case task := <-taskCh:
    handle(task)
case <-ctx.Done():
    return // 收到取消信号，退出
}
```

### 3. 多事件监听

```go
select {
case msg := <-msgCh:   // 业务消息
    fmt.Println("收到消息", msg)
case <-stopCh:          // 停止信号
    fmt.Println("收到停止")
}
```

---

## 五、高频自测

1. select 多个 case 同时就绪会执行哪个？
   → **伪随机选一个**，保证公平、防止饿死。
2. select 怎么实现超时？
   → **配合 `time.After` 分支**。
3. 所有 case 不可操作且没有 default 会怎样？
   → **阻塞等待**，直到某个 channel 可操作。
4. `select {}` 会怎样？
   → **永久阻塞**。

---

## 六、一句话总结

> **select = 多路 channel 收发 + 随机公平调度 + 可配 default 非阻塞**。监听是否有数据，多个就绪就随机进一个（防饿死），是做超时、退出、多事件监听的首选工具。
