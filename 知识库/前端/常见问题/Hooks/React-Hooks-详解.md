# React Hooks 详解

> 本文整理自前端基础概念系列讲解，涵盖 Hooks 的来龙去脉、副作用的概念、常用 Hooks 的用法，以及使用时的两条铁律。

---

## 目录

- [一、什么是 Hooks](#一什么是-hooks)
- [二、Hooks 出现之前](#二hooks-出现之前)
- [三、命名规律](#三命名规律)
- [四、先理解「副作用」](#四先理解副作用)
- [五、常用 Hooks 详解](#五常用-hooks-详解)
- [六、两条铁律](#六两条铁律)
- [七、自定义 Hook](#七自定义-hook)
- [八、完整示例](#八完整示例)

---

## 一、什么是 Hooks

一句话概括：

> **Hooks 是 React 提供的一批「use 开头」的特殊函数，让函数组件也能拥有状态、执行副作用等能力。**

Hooks 的核心价值：让普通函数组件「充电」，从而替代笨重的 class 类组件。

---

## 二、Hooks 出现之前

早期的 React 里，组件分两种，能力不对等：

| 组件类型 | 能存状态吗 | 能执行副作用吗 |
|---------|-----------|--------------|
| **函数组件** | 不能，只接收 props 展示内容 | 不能 |
| **class 类组件** | 能（`this.state`） | 能（生命周期方法） |

想要「有状态」，就必须写成麻烦的 class 类组件：

```jsx
class TodoApp extends React.Component {
  constructor(props) {
    super(props);
    this.state = { todos: [] };       // 状态
  }
  componentDidMount() { ... }         // 生命周期
  render() { ... }
}
```

class 的问题：

- `this` 绑定绕，容易出错
- 逻辑分散在多个生命周期方法里，难以维护
- 逻辑难以复用

React 16.8 引入 Hooks，解决的核心问题就是：**让函数组件也能做之前只有 class 才能做的事**，且写法更简洁：

```jsx
function TodoApp() {
  const [todos, setTodos] = useState([]);   // 状态
  useEffect(() => { ... }, []);             // 副作用
  return (...);
}
```

---

## 三、命名规律

所有 Hook 都以 `use` 开头：`useState`、`useEffect`、`useRef`、`useContext`……

看到 `useXxx`，就知道「这是个 Hook」。

---

## 四、先理解「副作用」

### 1. 什么是纯函数

纯函数有两个特征：

1. **同样的输入，永远得到同样的输出**
2. **不对外界产生任何影响**

```js
function double(x) {
  return x * 2;
}
```

`double(3)` 永远是 `6`，且除了算出结果返回，什么都没改变。这就是「纯」的。

### 2. 什么是副作用

函数在执行时，**对外界产生的任何影响，或者依赖了外界**，都叫副作用。一句话：**除了「计算并返回值」之外，做的事都是副作用。**

```js
function add(x) {
  console.log(x);        // 副作用：往控制台打印了东西
  return x + 1;
}

function save(data) {
  localStorage.setItem("key", data);  // 副作用：改了浏览器存储
  return data;
}

function fetchData() {
  return fetch("/api/list");          // 副作用：发起了网络请求
}
```

### 3. React 里的副作用指什么

组件的**本职工作**是渲染界面：

```
输入 props 和 state → 计算 → 输出一段界面描述（JSX）
```

这个「根据数据算出界面」的过程，应该是纯的。**渲染之外的所有事情，都是副作用**：

| 副作用类型 | 例子 |
|-----------|------|
| 调接口拿数据 | `fetch` / `axios` 请求后端 |
| 操作浏览器存储 | `localStorage` / `sessionStorage` |
| 操作真实 DOM | 手动改某个标签（React 之外的 DOM 操作） |
| 定时器 | `setTimeout` / `setInterval` |
| 订阅外部数据 | 监听事件、WebSocket、订阅 store |
| 打印日志 | `console.log` |

### 4. 为什么要用 useEffect 专门放副作用

如果把副作用直接写在组件函数主体里，会出问题：

- 组件每次渲染都重新执行函数主体，副作用会被**重复执行**（重复发请求、重复设定时器）
- 渲染过程变得不纯粹，React 无法可靠判断界面应该是什么样

所以 React 规定：

> **副作用不能写在渲染过程中，必须放进 `useEffect`。** `useEffect` 会在组件渲染完成后才执行副作用，并通过依赖数组控制执行时机。

### 5. 副作用还需要「清理」

定时器、事件监听、订阅这类副作用，如果不清理，会造成内存泄漏。`useEffect` 支持返回一个清理函数：

```jsx
useEffect(() => {
  const timer = setInterval(() => { ... }, 1000);  // 副作用：开了个定时器
  return () => clearInterval(timer);               // 清理：组件卸载时关掉它
}, []);
```

---

## 五、常用 Hooks 详解

| Hook | 作用 | 一句话理解 |
|------|------|-----------|
| `useState` | 存状态 | 让组件「记住」会变化的数据 |
| `useEffect` | 执行副作用 | 渲染后去做事：调接口、订阅、定时器、改 DOM |
| `useRef` | 拿 DOM 引用 / 存不触发渲染的值 | 直接摸到某个真实标签 |
| `useContext` | 跨层级共享数据 | 不用一层层传 props，直接读全局数据 |
| `useMemo` / `useCallback` | 性能优化 | 缓存计算结果 / 缓存函数 |

### 1. useState

```jsx
const [count, setCount] = useState(0);
```

`useState` 返回一个包含两项的数组，用解构赋值拆开：

- `count` → 状态的值
- `setCount` → 修改状态的函数，要改状态只能通过它

**函数式更新**：要基于上一次的值计算时，用函数式写法更安全：

```jsx
setCount((c) => c + 1);   // 安全，拿到的一定是最新值
setCount(count + 1);      // 多次快速点击可能拿到旧值
```

### 2. useEffect

```jsx
useEffect(() => {
  document.title = `你好，${name}`;
}, [name]);
```

**依赖数组**控制执行时机：

| 写法 | 执行时机 |
|------|---------|
| `[]` | 只在组件第一次挂载时执行一次 |
| `[name]` | `name` 变化时才执行 |
| 不写 | 每次渲染都执行 |

### 3. useRef

```jsx
const inputRef = useRef(null);

// 点按钮时，通过 ref 直接让输入框获得焦点
const focusInput = () => {
  inputRef.current.focus();
};

<input ref={inputRef} type="text" />
```

两个用途：

1. **拿真实 DOM**：直接操作某个标签
2. **存不触发渲染的值**：改 `ref.current` 不会导致组件重新渲染

### 4. useMemo

缓存「昂贵的计算」结果，依赖不变时直接返回缓存值：

```jsx
const sum = useMemo(() => {
  let total = 0;
  for (let i = 1; i <= num; i++) total += i;
  return total;
}, [num]);   // 只有 num 变了才重新计算
```

### 5. useCallback

缓存函数本身，常配合 `React.memo` 避免子组件无谓重渲染：

```jsx
const handleClick = useCallback(() => {
  console.log("clicked");
}, []);   // 依赖不变，每次拿到的都是同一个函数

<Child onClick={handleClick} />
```

### 6. useContext

跨层级共享数据，不用一层层传 props：

```jsx
const ThemeContext = React.createContext("light");

// 子组件直接读
const theme = useContext(ThemeContext);
```

---

## 六、两条铁律

违反会出 bug，必须遵守：

**1. 只在组件的最顶层调用**

不能在 `if`、`for`、嵌套函数里调用 Hook：

```jsx
// 错误：Hooks 放在条件里
if (loading) {
  useState(...)   // ❌ 不允许
}

// 正确：无条件地写在顶层
const [todos, setTodos] = useState([]);
```

**2. 只在 React 函数里调用**

只在函数组件或自定义 Hook 里调用，不要在普通 JS 函数里用。

**为什么有这两条规则？**

React 靠「调用顺序」记住每个 Hook 对应哪个状态。如果某次渲染跳过了某个 Hook，顺序就乱了，React 会对不上号，导致状态错乱。

---

## 七、自定义 Hook

除了内置 Hook，还可以把复用的逻辑抽成**自己的 Hook**（名字必须以 `use` 开头）：

```jsx
// 监听窗口宽度的自定义 Hook
function useWindowWidth() {
  const [width, setWidth] = useState(window.innerWidth);

  useEffect(() => {
    const onResize = () => setWidth(window.innerWidth);
    window.addEventListener("resize", onResize);
    // 清理：组件卸载时移除监听，避免内存泄漏
    return () => window.removeEventListener("resize", onResize);
  }, []);

  return width;
}

// 任何组件都能一行复用
function MyComponent() {
  const width = useWindowWidth();
  return <p>窗口宽度：{width} px</p>;
}
```

再比如，把状态存进 localStorage 的自定义 Hook：

```jsx
function useLocalStorage(key, initialValue) {
  const [value, setValue] = useState(() => {
    return localStorage.getItem(key) ?? initialValue;
  });

  useEffect(() => {
    localStorage.setItem(key, value);
  }, [key, value]);

  return [value, setValue];
}

// 使用
const [todos, setTodos] = useLocalStorage("todos", []);
```

这是 Hooks 最大的价值之一：**逻辑可以像零件一样被抽出来、到处复用**，而 class 组件时代很难做到。

---

## 八、完整示例

一个待办事项应用中，`useState` 和 `useEffect` 的实际用法：

```jsx
import { useState, useEffect } from "react";

function TodoApp() {
  const [todos, setTodos] = useState([]);      // 状态：待办列表
  const [loading, setLoading] = useState(true); // 状态：是否加载中

  // 副作用：模拟"打开页面后去接口拿数据"
  useEffect(() => {
    setTimeout(() => {
      setTodos([
        { id: 1, text: "学习 React 组件", done: true },
        { id: 2, text: "理解 state 状态", done: false },
        { id: 3, text: "搞清楚 props 传值", done: false },
      ]);
      setLoading(false);
    }, 800);
  }, []);   // 空数组 = 只在组件第一次挂载时执行一次

  // 增：不可变更新，用展开运算符造新数组
  const addTodo = (text) => {
    setTodos([...todos, { id: Date.now(), text, done: false }]);
  };

  // 改：map 只替换匹配的那一项
  const toggleTodo = (id) => {
    setTodos(todos.map((t) => (t.id === id ? { ...t, done: !t.done } : t)));
  };

  // 删：filter 造一个"去掉那项"的新数组
  const deleteTodo = (id) => {
    setTodos(todos.filter((t) => t.id !== id));
  };

  return (
    <div>
      <h1>待办事项</h1>
      {loading ? <p>加载中...</p> : (
        <ul>
          {todos.map((todo) => (
            <li key={todo.id}>
              <input
                type="checkbox"
                checked={todo.done}
                onChange={() => toggleTodo(todo.id)}
              />
              {todo.text}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
```

---

## 总结

- **Hooks** 是一批 `use` 开头的函数，给函数组件「充电」，让它拥有状态和副作用。
- **副作用** 是组件在「渲染界面」之外对外界的任何影响（调接口、改存储、设定时器、操作 DOM……），统一交给 `useEffect` 管理。
- **两条铁律**：只在组件最顶层调用、只在 React 函数里调用。
- **自定义 Hook** 能把复用逻辑抽出来，是 Hooks 最强大的能力。
