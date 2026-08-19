# C++ 性能认知：从内存布局到工程实践

> 来源：[金山文档](https://365.kdocs.cn/l/ctypxU6VzuOR) · 部门周会内部分享  
> 讲师：佘庆斌（WPS 表格内核研发 / 协同与计算组）

## 1. 背景与课程目标

### 1.1 性能问题并不遥远

以表格组件为例：表面上是二维数组 `Cells[M][N]`，实际规模可达 **1048576 × 16384**。即便每格仅 4 字节，单 sheet 理论内存约 **64 GB**，完全不可接受。

因此需要：

- 稀疏数据：多级寻址、动态扩缩容、资源共享
- 数据处理：百万级公式引用网络、图表/透视表统计、宏与加载项

任意一点处理不当，都可能成为性能热点。

### 1.2 课程目标

本课程关注 C++ 语言特性背后的**性能代价**：

- 每种抽象「花了多少钱」
- 看似等价的写法为何性能不同
- 如何在正确性与效率之间合理选型
- 永远追问：**根本问题是什么？愿意付出什么代价？**

---

## 2. 模块一：内存与存储代价

核心问题：**数据住在哪？住的位置决定了什么代价？**

### 2.1 存储开销阶梯

优化优先级：**编译期 > 栈 > 静态区 > 堆 > 外存**；能不分配就不分配。

| 存储位置 | 分配成本 | 典型场景 | 注意点 |
|---------|---------|---------|--------|
| `constexpr` | 编译期，运行时零成本 | 常量表、编译期计算 | 比 `const` 更强，强制编译期求值 |
| 栈 | ~1 条指令（移动栈指针） | 局部变量、小数组 | 热路径首选 |
| 静态区 | 程序加载时一次性完成 | 全局变量、局部 static | 有 SIOF、优化阻碍、线程安全问题 |
| 堆 | 数百条指令 + 可能锁竞争 | 动态大小、生命周期不确定 | 热路径慎用 `new`/`malloc` |

### 2.2 constexpr vs const

```cpp
const int a = strlen("hello");  // 运行时求值
constexpr int b = 5;            // 编译期确定
```

### 2.3 栈 vs 堆

```cpp
void stack_alloc() {
    int arr[256];   // ~1 条指令
    arr[0] = 42;
}

void heap_alloc() {
    int* arr = new int[256];  // 数百条指令
    arr[0] = 42;
    delete[] arr;               // 又是数百条指令
}
```

堆分配开销链：`operator new` → 堆管理器 → 搜索空闲块 → 可能系统调用 → 可能多线程锁 → 返回指针。

### 2.4 全局变量与局部 static

全局变量三种隐性代价：

1. **SIOF**（Static Initialization Order Fiasco）：跨编译单元初始化顺序未定义
2. **阻碍优化**：编译器无法假设不被外部修改
3. **线程安全**：非 `constexpr` 全局变量需额外同步

Meyers 单例（局部 static）有 guard 开销：首次构造加锁，后续每次原子读（通常 < 1ns，但高频累积可观测）。

| 单例实现 | 优点 | 缺点 |
|---------|------|------|
| Meyers 单例（局部 static） | 简洁、线程安全 | 每次调用有原子读 |
| 饿汉式（全局 static） | 零运行时开销 | 拖慢启动（构造重时） |

---

## 3. 模块二：对象模型与调用代价

核心问题：**数据长什么样？代码怎么执行？**

### 3.1 对齐与填充

成员排列影响 `sizeof`，进而影响容器内存占用与缓存效率。

```cpp
struct Bad  { char a; double b; char c; };  // sizeof = 24
struct Good { double b; char a; char c; };  // sizeof = 16，节省 33%
```

- `alignas(64)`：缓存行对齐，适合热点数据结构
- `#pragma pack`：仅 I/O 序列化等特定场景考虑，一般不值得

### 3.2 Trivially Copyable

| 类型 | 特征 | 容器性能 |
|------|------|---------|
| `Point { int x, y; }` | Trivially Copyable | `vector` 可 `memcpy` |
| `Widget { std::string name; }` | 非 Trivial | 必须逐成员构造/析构 |

虚函数会让类型失去 Trivially Copyable 属性。

### 3.3 虚函数代价

- **空间**：每个对象多一个 vptr（64 位下 8 字节）
- **时间**：读 vptr → 读 vtable → 间接跳转（难以内联）
- **继承越深代价越大**：多继承多个 vptr，虚继承需偏移查表

### 3.4 调用代价光谱

```
内联展开  <  直接调用  <  虚调用  <  std::function
≈0 开销      ~1ns        ~2-5ns      ~5-15ns
```

| 机制 | 要点 |
|------|------|
| 内联 | `inline` 关键字主要是 ODR 辅助；真正决定权在编译器；过度 `__forceinline` 增加 I-cache 压力 |
| LTO | 跨编译单元内联，但显著增加链接时间 |
| `std::function` | 类型擦除 + 可能堆分配 + 无法内联；热路径用模板参数替代 |
| 模板 / CRTP | 编译期绑定，零运行时分发；WPS 实例：`KAtomVariety2` |
| `final` | 帮助编译器 devirtualize |

```cpp
// 慢
void forEach(const std::vector<int>& v, std::function<void(int)> f);

// 快：编译器可内联
template<typename F>
void forEach(const std::vector<int>& v, F f);
```

### 3.5 类型转换代价阶梯

```
static_cast / reinterpret_cast  <  dynamic_cast  <  QueryInterface
（编译期，近零开销）              （RTTI 遍历）      （虚调用 + GUID + AddRef）
```

- `dynamic_cast`：增加二进制体积，继承越深越慢；类型可枚举时用 `enum + switch` 或 visitor
- COM `QueryInterface`：循环中应缓存结果

### 3.6 异常与 noexcept

| 维度 | 代价 |
|------|------|
| 空间 | 启用 `/EH` 后 unwind tables 使二进制增大 10%~30%（WPS 部分模块禁用异常） |
| 时间 | 正常路径零开销；抛出时慢 10~100 倍 |
| `noexcept` | 影响 `vector` 扩容：无 `noexcept` 移动构造则退回 copy |

选型：罕见严重错误 → 异常；频繁可恢复 → 错误码（如 `HRESULT`）；极端敏感 → 错误码 + `[[nodiscard]]`。

---

## 4. 模块三：常用抽象的性能画像

### 4.1 智能指针

| 类型 | 开销 | 适用场景 |
|------|------|---------|
| `unique_ptr` | 零开销，与裸指针同大小 | 独占所有权，默认首选 |
| `shared_ptr` | 控制块 16~40B + 原子操作 + 可能两次堆分配 | 跨模块共享；用 `make_shared` |
| `ks_stdptr`（WPS） | 侵入式引用计数，无独立控制块 | COM 对象、跨 DLL |

**`unique_ptr` deleter 注意**：无状态仿函数零开销；函数指针 deleter 大小翻倍。

**`shared_ptr` 传参**：

```cpp
void process(const std::shared_ptr<Widget>& w);  // 好：不改引用计数
void process(Widget* w);                          // 好
void process(std::shared_ptr<Widget> w);          // 差：拷贝触发原子操作
```

**WPS `KS_NEW` / attach / detach 陷阱**：

```cpp
// 错：双重 AddRef → 泄漏
ks_stdptr<MyObj> p = KS_NEW(MyObj);

// 对
ks_stdptr<MyObj> p;
p.attach(KS_NEW(MyObj));
```

#### 智能指针选型决策树

```
需要共享所有权？
├─ 否 → unique_ptr
└─ 是 → 继承 IUnknown / COM 基类？
         ├─ 是 → ks_stdptr / ks_comptr
         └─ 否 → shared_ptr（make_shared）
                  └─ 需要弱引用？→ weak_ptr
```

### 4.2 字符串

#### SSO（Small String Optimization）

- MSVC/libstdc++ x64：`sizeof(string) = 32`，SSO 阈值 ≤15 字节
- libc++ x64：`sizeof = 24`，阈值 ≤22 字节
- SSO 范围内 move 与 copy 开销相同（都是 memcpy 32 字节）

#### 编码策略

- 编码转换是 CPU 密集型操作
- **WPS 方案**：内核统一 UTF-16（`ks_wstring` + `WCHAR`），外部接口按需转换

#### 常见陷阱

| 陷阱 | 说明 | 建议 |
|------|------|------|
| `operator+` 连锁 | 每个 `+` 产生临时对象 | `Format` 或 `reserve + append` |
| `std::move(const_obj)` | 静默退化为拷贝，无警告 | 确保源对象非 const |
| `string_view` 生命周期 | 不拥有数据、不保证 `\0` 结尾 | 传 C API 前先转 `string` |
| 参数类型过大 | `const ks_wstring&` 强制构造 | 用 `PCWSTR` 等最小语义类型 |

#### QString COW 的隐性代价

拷贝/析构需原子操作；可写访问可能意外深拷贝；缓存行失效。C++11 后 `std::string` 已禁止 COW。

### 4.3 容器选型

#### 顺序容器：默认 vector

- 连续内存 → 缓存预取友好
- `list` 节点分散堆上 → 缓存不友好
- 热路径小 vector 应 `reserve`，避免多线程下分配器竞争

#### 关联容器：复杂度 ≠ 实际性能

| 场景 | 结论 |
|------|------|
| N < 几百 | `vector` 线性查找可能快于 `unordered_map` |
| 哈希设计不当 | O(1) 退化为 O(n)，如 `hash = row << 20 \| col` 导致同列碰撞 |

**通用原则：默认选 `vector`，除非有明确理由选其他容器。**

---

## 5. 核心认知总结

```
1. 存储层次决定代价
   寄存器 → 栈 → 静态区 → 堆 → 外存

2. 抽象不是免费的
   虚函数有 vptr，shared_ptr 有原子操作
   但 unique_ptr 和 CRTP 证明：正确抽象可以零开销

3. 选型比优化更重要
   容器、智能指针、编码方式选对了，收益往往大于事后调优

4. 复杂度 ≠ 实际性能
   缓存友好性、常数因子、分支预测可颠覆理论分析
```

### 5.1 工程态度

- **不要怕性能**：逐步压榨性能本身有成就感
- **不要过早优化**：先做对，再在正确前提下优化
- **保持独立判断**：理解「为什么」比死记「怎么做」更重要
- **AI 时代**：AI 能给出方案，但人的洞察用于纠偏与选型仍不可替代

---

## 6. 附录：真实案例速查

| # | 问题 | 根因 | 对策 |
|---|------|------|------|
| 1 | 哈希表退化为链表 | 行坐标编码到高位，同列低位相同 | 哈希设计需考虑真实数据分布 |
| 2 | 资源二次释放 | 资源转移未 `detach` | 内外语义一致：接管方必须释放 |
| 3 | 内存泄漏 | 内部 `detach` 交由调用方，调用方未释放 | 明确所有权转移契约 |
| 4 | 分支泄漏 | `pToken` 按 managed 传入，某分支未管理 | RAII 保证所有分支释放 |
| 5 | begin/end 不匹配 | 中间步骤失败提前退出 | RAII 保护配对操作 |
| 6 | 宏副作用 | 表达式作为宏参数被多次展开 | 避免把表达式传入宏 |
| 7 | 迭代器失效 | 链式 `++`/`--` | 用 `+=` 或 `std::advance` |
| 8 | 野指针 | 链式调用返回临时对象 | 避免持有临时对象引用/指针 |
| 9 | 30% 性能差 | 返回 `IUnknown` 包装 vs 原始对象 | 高频路径避免不必要包装与计数 |

---

## 7. 性能速查表

| 抽象 / 机制 | 主要代价 | 零开销替代 |
|------------|---------|-----------|
| `constexpr` | 无（编译期） | — |
| 栈分配 | 极低 | — |
| 堆分配 | 高（含锁） | 栈、SSO、预分配 |
| 全局变量 | SIOF、优化阻碍 | Meyers 单例 |
| 虚函数 | vptr + 间接调用 | CRTP、模板、`final` |
| `std::function` | 类型擦除 + 堆分配 | 模板回调 |
| `dynamic_cast` | RTTI 遍历 | `static_cast`（类型确定时） |
| 异常 `/EH` | 二进制 +10~30% | 错误码 |
| `unique_ptr` | 零 | — |
| `shared_ptr` | 控制块 + 原子操作 | `unique_ptr`、传引用 |
| `ks_stdptr` | 侵入式计数 | COM 场景首选 |
| `string` SSO 内 | 零堆分配 | 控制字符串长度 |
| `string_view` | 零拷贝（不拥有） | 只读参数 |
| `vector` | realloc（可 reserve） | 默认容器 |
| `unordered_map` | 哈希 + 堆节点 | 小数据用 `vector` 线性查找 |
