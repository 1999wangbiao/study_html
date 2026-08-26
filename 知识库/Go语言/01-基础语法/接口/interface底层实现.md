# interface 的底层实现

> **interface 底层是"两个指针"：类型信息（`_type`/`itab`）＋ 数据指针（data）**。空接口 `interface{}` 用 eface（类型 + 数据）；非空接口用 iface（itab 含方法集 + 数据）。类型断言和方法调用都基于这两部分实现。

---

## 一、底层结构

interface 在 runtime 中表现为 **iface** 或 **eface** 两种结构体，取决于是否包含方法：

### 1. 非空接口（iface）——包含方法

```go
type iface struct {
    tab  *itab          // 类型信息表（包含类型和方法集）
    data unsafe.Pointer // 指向实际数据的指针
}

// itab 存储接口与具体类型的匹配信息
type itab struct {
    inter *interfacetype // 接口自身的类型信息（如方法集）
    _type *_type         // 具体类型的元信息（如 int、struct 等）
    hash  uint32         // 类型哈希（用于快速比较）
    _     [4]byte
    fun   [1]uintptr     // 方法集（动态分派的函数指针列表）
}
```

### 2. 空接口（eface）——无方法

```go
type eface struct {
    _type *_type         // 具体类型的元信息
    data  unsafe.Pointer // 指向实际数据的指针
}
```

其中 `_type` 是所有类型的公共元信息结构体，包含**类型名称、大小、对齐方式、哈希函数**等基础信息。

---

## 二、核心原理

### 1. 接口赋值过程

当一个具体类型的值赋值给接口时，Go 会：

```go
var x interface{} = 42      // eface{_type: int类型, data: 指向42的指针}
var r io.Reader = os.Stdin  // iface{tab: 匹配信息, data: 指向os.Stdin的指针}
```

1. 提取该值的**类型信息（_type）**；
2. 若为非空接口，还会生成 **itab**（验证类型是否实现接口方法集，并缓存方法指针）；
3. 复制值的地址到 **data 指针**（对于值类型，会先分配内存并拷贝值；对于引用类型，直接存储原指针）。

### 2. 类型断言（type assertion）

判断接口中存储的具体类型时，底层通过**比较 `_type` 或 `itab._type`** 实现：

- 若类型匹配，返回 data 指向的值；
- 若不匹配，返回 false（或触发 panic，当使用 `x.(T)` 形式时）。

```go
v, ok := x.(int) // 底层比较 _type 是否 int
```

### 3. 方法调用

非空接口调用方法时，通过 **itab.fun 中的函数指针直接调用**（动态分派），无需再进行类型检查，效率接近直接调用。

---

## 三、关键特性

1. **值拷贝语义**：接口存储的是值的副本（值类型）或引用（引用类型），修改接口内部数据不会影响原变量；
2. **延迟绑定**：接口与具体类型的匹配在**运行时**完成（编译期仅做静态检查）；
3. **内存优化**：对于小值类型（如 int、bool），data 指针可能**直接存储值**（通过指针对齐优化，避免额外内存分配）。

---

## 四、iface vs eface

| 维度 | 空接口 eface | 非空接口 iface |
|------|--------------|----------------|
| 是否含方法 | 否（interface{}） | 是（如 io.Reader） |
| 结构 | `_type` + data | **itab** + data |
| 方法分派 | 无 | **itab.fun 函数指针动态分派** |

---

## 五、高频自测

1. 空接口底层结构？
   → **eface：`_type`（类型）+ data（数据指针）**。
2. 非空接口底层结构？
   → **iface：itab（类型 + 方法集）+ data**。
3. 类型断言靠什么判断？
   → **比较 `_type` 是否匹配**。
4. 接口方法调用为什么快？
   → **itab.fun 里缓存了函数指针，直接动态分派**。

---

## 六、一句话总结

> **interface = 类型信息（`_type`/`itab`）＋ 数据指针（data）**：空接口只管"类型 + 数据"，非空接口多了 itab（方法集 + 函数指针）支持快速动态分派。赋值生成 itab、断言比对 _type、调用走 fun 指针——多态的地基。
