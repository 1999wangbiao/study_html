// =============================================
// 顶部导入：从 React 中引入 useState 这个 Hook
// =============================================
// Hook 是 React 16.8 之后提供的函数，让函数组件也能拥有"状态"。
// useState 就是用来在组件里创建和管理状态的。
import { useState } from "react";

// =============================================
// ① 子组件：ProductCard（商品卡片）
// =============================================
// 这是一个"函数组件"——用一个函数来定义一个组件。
// 组件的名字首字母必须大写，这样 React 才知道它是组件，
// 而不是普通的 HTML 标签。
function ProductCard({ name, price, onAdd }) {
  // 上方花括号里是"解构赋值"的写法，
  // 等于从 props 对象里取出 name、price、onAdd 三个属性。
  // 这三个值都是父组件传进来的（props），
  // 子组件只能"读"它们，不能直接修改——这是 React 的规矩。

  // 组件最终要 return 一段 JSX（长得像 HTML 的代码），
  // 这就是这个组件渲染到页面上的样子。
  return (
    // style={cardStyle} 表示给这个 div 应用底部定义的那组样式
    <div style={cardStyle}>
      {/* 花括号 {} 用来在 JSX 里插入 JS 表达式，这里插入变量 name */}
      <h3>{name}</h3>

      {/* 插入变量 price，前面拼一个人民币符号 */}
      <p>¥{price}</p>

      {/* onClick 是点击事件：按钮被点一下，就执行箭头函数 */}
      {/* 箭头函数里调用 onAdd(name)，
          把"当前这件商品的名字"回传给父组件。
          注意：这里不是立即执行，而是"点击时才执行"。 */}
      <button onClick={() => onAdd(name)}>加入购物车</button>
    </div>
  );
}

// =============================================
// ② 父组件：ProductList（商品列表 + 购物车）
// =============================================
function ProductList() {
  // ---------- 状态 state ----------
  // useState([]) 创建一个状态变量，初始值是空数组 []。
  // 它返回一个数组，我们用解构赋值拆成两项：
  //   cart      —— 状态本身，存购物车里的商品名列表
  //   setCart   —— 用来"修改"这个状态的函数（只能通过它改）
  const [cart, setCart] = useState([]);

  // ---------- 数据 ----------
  // 商品列表数据。现实中通常来自接口（API），这里先写死演示。
  // 每件商品有一个稳定的 id（后面要做 key）、名字和价格。
  const products = [
    { id: 1, name: "手机", price: 4999 },
    { id: 2, name: "耳机", price: 399 },
    { id: 3, name: "充电宝", price: 129 },
  ];

  // ---------- 修改状态的方法 ----------
  // addToCart 接收一个商品名，把它加进购物车。
  // 关键点：不能直接 push（那样不会触发界面更新），
  // 必须调用 setCart 传入"一个新的数组"。
  // [...cart, name] 是展开运算符：把旧数组展开，再拼上新产品，
  // 得到一个全新数组——React 发现引用变了，才会重新渲染。
  const addToCart = (name) => {
    setCart([...cart, name]);
  };

  // 组件 return 的 JSX 就是这个组件的界面
  return (
    <div>
      <h2>商品列表</h2>

      {/* ---------- 列表渲染 ---------- */}
      {/* products.map(p => ...) 遍历数组，每一项生成一个 ProductCard。
          key={p.id} 给每个节点一个唯一"身份证"，
          帮助 React 的 diff 算法识别哪个节点变了、哪个是新增的。
          key 不要用数组下标，要用稳定唯一的 id。 */}
      {products.map((p) => (
        <ProductCard
          key={p.id}          // 唯一标识，必需
          name={p.name}       // 把商品名作为 props 传下去
          price={p.price}     // 把价格作为 props 传下去
          onAdd={addToCart}   // 把"加购物车"的方法作为 props 传下去
        />
      ))}

      {/* ---------- 购物车展示 ---------- */}
      {/* cart.length 实时显示购物车里有几件商品。
          只要 cart 状态一变，这个数字就会自动更新。 */}
      <h3>购物车（{cart.length} 件）</h3>

      {/* 再用一次 map，把购物车里每件商品渲染成 li 列表项 */}
      <ul>
        {/* 这里用下标 i 当 key，仅因为演示购物车不会中途增删排序；
            真实项目里如果列表会增删/排序，key 应该用稳定唯一的值。 */}
        {cart.map((item, i) => (
          <li key={i}>{item}</li>
        ))}
      </ul>
    </div>
  );
}

// =============================================
// 样式对象
// =============================================
// 用 JS 对象写样式（React 里叫"内联样式"）。
// 注意 CSS 属性名要改成"小驼峰"写法：
//   border-radius  →  borderRadius
//   display: inline-block  →  display: "inline-block"（值用字符串）
const cardStyle = {
  border: "1px solid #ddd",   // 边框：1像素、实线、浅灰色
  borderRadius: 8,            // 圆角 8 像素（数字默认单位是 px）
  padding: 12,                // 内边距 12 像素
  margin: 8,                  // 外边距 8 像素
  display: "inline-block",    // 让卡片排成一行
  width: 150,                 // 宽度 150 像素
};

// =============================================
// 导出组件
// =============================================
// 把这个组件暴露出去，别的文件（比如 App.jsx）才能 import 使用。
export default ProductList;