# Go 的 sync.Map 底层实现

> **sync.Map = 两个普通 map（read + dirty）＋ 原子操作，专为"读多写少"优化**：读尽量走无锁的 read map，写走加锁的 dirty map，`misses` 累积到阈值时把 dirty 提升为 read。**读多写少性能远超 map+Mutex，写密集反而更慢。**

---

## 一、核心设计思路

`sync.Map` 是 Go 标准库中为并发场景设计的**线程安全映射表**，专为**读多写少**场景优化，底层通过**两个普通 map + 原子操作**实现，避免了传统 `map + Mutex` 方案的全局锁开销。

内部维护两个 map：

- **read map**：一个只读的 map（通过**原子指针**访问），存储大部分稳定数据，支持**无锁读取**；
- **dirty map**：一个可写的 map，存储新写入或修改的数据，需要**加锁访问**。

通过分离读写路径，sync.Map 在读取频繁的场景下，**大部分操作无需加锁**，显著提升性能。

---

## 二、底层结构（简化版）

```go
type Map struct {
    mu    Mutex                // 保护 dirty map 的互斥锁
    read  atomic.Value         // 存储 readOnly 结构体（包含只读 map）
    dirty map[interface{}]*entry // 可写的 dirty map
    misses int                 // 记录 read map 未命中后查询 dirty map 的次数
}

// readOnly 是 read 字段的实际类型，包含一个只读 map 和一个修改标记
type readOnly struct {
    m       map[interface{}]*entry // 只读 map
    amended bool                   // 标记 dirty map 有 read map 没有的键（true 表示有差异）
}

// entry 存储值的指针，支持原子更新
type entry struct {
    p unsafe.Pointer // 指向实际值（nil 表示已删除，expunged 表示从 dirty 中删除）
}
```

---

## 三、核心操作原理

### 1. 读取操作（Load）——最优化

1. 先从 **read map** 中读取（**无锁**，通过原子操作获取 readOnly 结构体）；
2. 若找到键且值未被删除，**直接返回结果（无需加锁，性能极高）**；
3. 若 read map 中未找到，且 `amended` 为 true（表示 dirty map 有新数据），则**加锁查询 dirty map**：
   - 若在 dirty map 中找到，`misses` 计数器 +1；
   - 当 `misses` 等于 dirty map 长度时，**将 dirty map 提升为 read map**（减少后续查询开销），并重置 dirty map 和 misses。

### 2. 写入操作（Store）

1. 若 **read map 中已存在该键**，且值未被删除，**直接通过原子操作更新值（无需加锁）**；
2. 若 read map 中不存在或值已被删除：
   - 加锁检查 dirty map：
     - 若 dirty map 存在该键，直接更新；
     - 若 dirty map 不存在，**先将 read map 中所有未删除的键复制到 dirty map**（仅第一次需要），再写入新键值对，并标记 `amended = true`。

### 3. 删除操作（Delete）

1. 先尝试从 **read map** 中找到该键，若存在且未被删除，通过**原子操作将其标记为删除**（设置 p = nil）；
2. 若 read map 中无该键，且 dirty map 存在，则**加锁从 dirty map 中删除**，并标记为 **expunged**（表示彻底删除，区别于 nil）。

---

## 四、核心优势与适用场景

- **优势**：读多写少场景下性能远超 `map + Mutex`，因为**大部分读操作无需加锁**；
- **劣势**：写操作（尤其是**频繁写入新键**）会触发 **dirty map 复制**，开销较大；
- **适用场景**：缓存（如配置项、静态数据）、日志收集等**读频繁、写较少**的并发场景。

---

## 五、高频自测

1. sync.Map 底层用几个 map？
   → **两个：read（无锁读）+ dirty（加锁写）**。
2. 什么情况下把 dirty 提升为 read？
   → **read 未命中次数（misses）达到 dirty 长度时**。
3. sync.Map 适合什么场景？
   → **读多写少**（缓存、配置）；写密集反而更慢。
4. 写入已存在的键需要加锁吗？
   → **read map 里已存在时不需要**（原子更新即可）。

---

## 六、一句话总结

> **sync.Map = read（原子读，无锁）＋ dirty（加锁写）＋ misses 升级机制**：读先查 read，miss 多了就把 dirty 抬上来。用"读写分离 + 原子操作"把读多写少场景的锁竞争降到最低——但写密集不划算。
