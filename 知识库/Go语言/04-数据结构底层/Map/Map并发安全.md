# 如何避免 Map 的并发问题

> **Map 并发写会 panic**。三种解决思路：**互斥锁（Mutex/RWMutex）、sync.Map（读多写少优化）、复制私有副本（无共享）**。

---

## 一、使用互斥锁（sync.Mutex / sync.RWMutex）

最通用、可控的方式：

```go
type SafeMap struct {
    mu sync.RWMutex       // 读写锁
    m  map[string]int
}

func (s *SafeMap) Get(key string) int {
    s.mu.RLock()          // 读锁：可并发读
    defer s.mu.RUnlock()
    return s.m[key]
}

func (s *SafeMap) Set(key string, v int) {
    s.mu.Lock()           // 写锁：互斥
    defer s.mu.Unlock()
    s.m[key] = v
}
```

> 读多写少用 **RWMutex**（读读不互斥），写多场景用普通 Mutex 即可。

---

## 二、使用 sync.Map

`sync.Map` 是 Go 标准库提供的**并发安全 map**，专为 **"读多写少"** 场景优化，内部通过**分离读写 map 和原子操作**实现高效并发：

```go
var m sync.Map

// 写
m.Store("key", 42)
// 读
if v, ok := m.Load("key"); ok {
    fmt.Println(v.(int))
}
// 删除
m.Delete("key")
// 遍历
m.Range(func(k, v interface{}) bool {
    fmt.Println(k, v)
    return true // 继续遍历
})
```

- **读多写少**（缓存、配置、静态数据）场景性能远超 map + Mutex（大部分读无锁）；
- **写操作密集**场景反而更慢（dirty map 复制开销大）。

---

## 三、复制私有副本（无共享访问）

让**每个 Goroutine 持有 map 的私有副本**，通过避免共享来实现并发安全：

```go
// 每个 worker 拿到自己的副本，互不干扰
func worker(base map[string]int) {
    my := make(map[string]int, len(base))
    for k, v := range base {
        my[k] = v // 复制私有副本
    }
    // 只操作 my，无需加锁
}
```

适用于**读多写少且数据变更频率低**的场景（如只读配置的初始化快照）。

---

## 四、三种方式对比

| 方式 | 原理 | 适用场景 | 性能 |
|------|------|----------|------|
| Mutex / RWMutex | 锁保护临界区 | 通用、写多 | 中等（有锁开销） |
| **sync.Map** | 读写 map 分离 + 原子操作 | **读多写少** | 读场景接近无锁 |
| 私有副本 | 不共享就无竞争 | 只读/低频变更 | 最高（无锁） |

---

## 五、高频自测

1. 直接并发写 map 会怎样？
   → **panic**（concurrent map writes）。
2. sync.Map 适合什么场景？
   → **读多写少**（缓存、配置、静态数据）。
3. 三个方案怎么选？
   → **通用选锁，读多写少选 sync.Map，只读低频选副本**。

---

## 六、一句话总结

> **Map 并发安全三件套：锁（通用）、sync.Map（读多写少最优）、私有副本（只读场景无锁）**。直接并发写 map 会 panic，务必选择其一。
