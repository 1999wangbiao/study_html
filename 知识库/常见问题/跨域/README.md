# 前后端跨域（CORS）

用一句话抓住：

> **跨域是浏览器同源策略拦的，不是服务端收不到请求。**  
> 要么服务端发 CORS 头让浏览器放行，要么用代理 / 同域部署让浏览器「看起来不跨域」。

---

## 1. 出现原理

### 1.1 同源策略

浏览器有同源策略（Same-Origin Policy）：页面默认只能读写「同源」资源。

**同源** = 协议 + 域名 + 端口 三者都相同。任一不同就是跨域。

| 页面 | 接口 | 是否跨域 | 原因 |
|------|------|----------|------|
| `http://localhost:5173` | `http://localhost:8080` | 是 | 端口不同 |
| `https://a.com` | `http://a.com` | 是 | 协议不同 |
| `https://web.a.com` | `https://api.a.com` | 是 | 子域名不同 |
| `https://a.com` | `https://a.com/api` | 否 | 三者相同 |

### 1.2 为何前后端联调常踩坑

本地常见拆分：

- 前端：Vite / Webpack 在 `http://localhost:5173`
- 后端：Gin / Spring 在 `http://localhost:8080`

浏览器判定为跨域，于是拦截前端 JS 读取响应。

### 1.3 核心句细讲：拦的是「读响应」，不是「发请求」

这句话要拆成两层：

1. **网络层**：浏览器往往已经把 HTTP 请求发出去了，服务端也可能已经查库、写库、返回了 200。
2. **安全层**：响应回到浏览器后，**浏览器先按 CORS 规则检查；不通过**，就**不把响应体交给页面里的 JS**。

所以：服务端日志里可能看到这次请求；前端 `await fetch(...)` 却进了 `catch`，控制台报 CORS。两边并不矛盾。

#### 1.3.1 参与者各管什么

| 角色 | 做什么 | 管不管同源 |
|------|--------|------------|
| 页面 JS | 调用 `fetch` / `axios`，期望拿到 JSON | 不管；它只发起调用 |
| 浏览器 | 真正发包、收包，并执行同源 / CORS 检查 | **管**；不通过就不给 JS 数据 |
| 服务端 | 收请求、跑业务、写响应 | **不管**同源；HTTP 请求照常处理 |
| Postman / curl | 自己发 HTTP、自己读响应 | **不管**；所以看起来「接口是好的」 |

同源策略是浏览器对「网页里的脚本」加的门锁，不是对整个互联网加的门锁。

#### 1.3.2 简单请求的完整时序（服务端通常已收到）

以页面 `http://localhost:5173` 去请求 `http://localhost:8080/api/user` 为例（端口不同 → 跨域）：

```text
① 页面 JS
   fetch('http://localhost:8080/api/user')

② 浏览器（发出前）
   · 发现页面 Origin 与目标 URL 不同源 → 这是跨域 XHR/fetch
   · 仍然发出请求（简单请求不会先拦发送）
   · 自动带上：Origin: http://localhost:5173

③ 服务端（8080）
   · 正常收到 GET /api/user
   · 正常查库、组装 JSON
   · 返回 HTTP 200 + body（若未配 CORS，响应里没有
     Access-Control-Allow-Origin）

④ 浏览器（收到后，关键一步）
   · 先看响应头有没有允许当前 Origin 的 CORS 头
   · 没有 / 不匹配 → 判定「页面脚本不能读这份响应」
   · 把错误抛给 JS（CORS policy ...）
   · JS 拿不到 body，哪怕服务端其实返回了 200

⑤ 页面 JS
   · Promise reject / 进 catch
   · 开发者容易误判成「接口挂了」
```

画成对照：

```text
                    发出请求                收到响应
页面 JS  ──────►  浏览器  ──────►  服务端  ──────►  浏览器  ──╳──►  页面 JS
                  （放行发送）     （照常处理）      （CORS 检查失败，
                                                    不把 body 交给 JS）
```

**被拦的是箭头最后一截「浏览器 → 页面 JS」**，不是中间「浏览器 → 服务端」。

#### 1.3.3 用 Network 面板怎么验证这句话

打开 DevTools → Network，点开那条失败的请求，常能看到：

| 你看到的 | 含义 |
|----------|------|
| 请求确实出现在列表里 | 浏览器发过包 |
| Status 可能是 200 | 服务端处理成功并回了包 |
| Response / Preview 可能空白或提示被 CORS 挡住 | 浏览器不让你（以及页面 JS）读 body |
| Console 有 CORS 红字 | 安全策略在「读」这一步否决了 |

再对比：同一 URL 用 Postman 调用——没有「页面 Origin」，也没有「把结果交给敌对页面脚本」的风险模型，所以直接看到 200 和 JSON。这正好证明：**问题不在服务端收不到，而在浏览器不给网页脚本看。**

#### 1.3.4 例外：预检失败时，正式请求可能根本没发出

复杂请求（如 `Content-Type: application/json`、自定义 `Authorization`、`PUT` / `DELETE` 等）流程不同：

```text
① JS 要发「复杂」跨域请求
② 浏览器先发 OPTIONS 预检（仍会到服务端）
③ 浏览器检查预检响应的 CORS 头
   · 通过 → 再发真正的 GET/POST/...（服务端会再收到一次）
   · 失败 → 正式请求不再发出；JS 同样报 CORS
```

因此更精确的说法是：

- **简单请求 / 预检已通过**：服务端通常已处理正式请求；拦的是「JS 读响应」。
- **预检未通过**：正式业务请求可能从未到达服务端；拦在「允不允许发这类跨域请求」。

日常说「不是服务端收不到」，主要针对大家最容易误会的第一种情况——日志里明明有 200，前端却喊跨域。

### 1.4 简单请求与预检（简表）

| 类型 | 条件（示意） | 浏览器行为 | 服务端是否收到正式请求 |
|------|--------------|------------|------------------------|
| 简单请求 | 常见 `GET` / `POST` + 有限 Header | 直接发正式请求，再检查响应头 | 通常已收到 |
| 预检请求 | 自定义 Header、JSON 体、部分方法等 | 先 `OPTIONS`，通过后再发正式请求 | 预检失败则正式请求未发 |

### 1.5 为何要这么设计（直觉）

若没有这层检查：恶意站点 `evil.com` 也能让你的浏览器带着你的登录态去读 `bank.com` 的接口，并把 JSON 交给 `evil.com` 的脚本。

CORS 的默认态度是：**跨域响应默认对网页脚本不可读**；只有目标站通过响应头明确说「我允许某某 Origin 读」，浏览器才松手。

服务端照常处理请求，是因为 HTTP 本身没有「同源」概念；安全边界加在浏览器与网页脚本之间。

---

## 2. 不解决会出现什么错误

### 2.1 控制台典型报错

```text
Access to fetch at 'http://localhost:8080/api/xxx' from origin 'http://localhost:5173'
has been blocked by CORS policy: No 'Access-Control-Allow-Origin' header is present
on the requested resource.
```

预检失败时常见：

```text
Response to preflight request doesn't pass access control check
```

带 Cookie 时还可能：

```text
The value of the 'Access-Control-Allow-Origin' header in the response must not be
the wildcard '*' when the request's credentials mode is 'include'.
```

### 2.2 现象对照

| 现象 | 说明 |
|------|------|
| 控制台 CORS 红字 | 浏览器按策略拦截了 JS 读响应 |
| Network 里请求红 / failed | 或 OPTIONS 失败导致正式请求未发 |
| `fetch` / `axios` 进 `catch` | 前端拿不到响应体（即便服务端曾返回 200） |
| Postman 正常、页面异常 | 典型「只有浏览器拦」的特征 |

---

## 3. 解决方案有多少种

工程上常用 **5 类**（另有 1 类仅调试、不算产品方案）。

| 方案 | 一句话 | 谁改 | 典型场景 |
|------|--------|------|----------|
| A. 服务端 CORS 头 | 告诉浏览器允许该 Origin | 后端 | 前后端不同域且必须直连 |
| B. 开发环境代理 | 浏览器只打同源，工具转发 | 前端构建工具 | 本地联调 |
| C. 反向代理同域 | 对外同一 Origin | 运维 / 网关 | 生产标准做法 |
| D. BFF / 后端中转 | 服务端代调真正 API | 后端 | 第三方、聚合、藏真实地址 |
| E. JSONP | 用 `<script>` 绕过 | 前后端 | 遗留，不推荐 |
| （非方案）关浏览器安全 | 仅个人调试 | 本机浏览器 | 不能上线 |

---

## 4. 每一种方案通过什么手段解决

### 4.1 方案 A：服务端开 CORS

**手段**：在响应里加 CORS 响应头，让浏览器放行「读跨域响应」。

关键响应头：

| 头 | 作用 |
|----|------|
| `Access-Control-Allow-Origin` | 允许的前端源（具体 Origin，或开发时临时 `*`） |
| `Access-Control-Allow-Methods` | 允许的方法，如 `GET, POST, PUT, DELETE, OPTIONS` |
| `Access-Control-Allow-Headers` | 允许的请求头，如 `Content-Type, Authorization` |
| `Access-Control-Allow-Credentials` | 允许带 Cookie 时设为 `true`（此时 Origin 不能是 `*`） |
| `Access-Control-Max-Age` | 预检结果缓存秒数 |

Gin / Express / Spring 等通常用 CORS 中间件统一处理，并正确响应 `OPTIONS`。

**本质**：不改变「跨域」事实，只让浏览器同意 JS 读响应。

### 4.2 方案 B：开发环境代理

**手段**：前端请求同源路径（如 `/api/...`），由 Vite / Webpack Dev Server 转发到后端。

浏览器看到的是 `5173 → 5173`（同源）；跨域发生在「开发服务器 ↔ 后端」，浏览器不管。

Vite 示例：

```ts
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
    },
  },
}
```

**适用**：本地联调；生产需另配 Nginx / 网关（见方案 C）。

### 4.3 方案 C：反向代理 / 同域部署

**手段**：Nginx / Caddy / API Gateway 把前端与 API 挂在同一域名下：

- `https://app.com/` → 静态前端
- `https://app.com/api/` → 后端服务

浏览器始终同源，自然无 CORS。

**本质**：从架构上消掉跨域，而不是「允许跨域」。生产最推荐。

### 4.4 方案 D：BFF / 后端中转

**手段**：前端只调自己的同源后端（BFF）；由 BFF 再请求真正的第三方或内网 API。

跨域发生在「服务端 ↔ 服务端」，不受浏览器同源策略限制。

**适用**：对接第三方、隐藏真实 API、统一鉴权与聚合。

### 4.5 方案 E：JSONP（历史方案）

**手段**：用 `<script src="...?callback=fn">` 加载；服务端返回 `fn(data)`。

限制与风险：

- 只适合 **GET**
- 存在脚本注入等安全风险
- 现代项目用 CORS / 代理即可，不必再用

### 4.6 不算正经方案

- 浏览器关安全策略（如 `--disable-web-security`）：仅个人调试，不能当产品方案。
- 生产环境对所有源返回 `Access-Control-Allow-Origin: *`：开发图省事可以，上线应收紧 Origin；带 Cookie 时更不能用 `*`。

---

## 5. 怎么选

| 场景 | 优先选 |
|------|--------|
| 本地前后端不同端口 | B 开发代理，或 A 后端 CORS 中间件 |
| 生产前后端分离域名 | C 同域反代优先；必须跨域直连再用 A |
| 调第三方且不想暴露密钥 / 地址 | D BFF |
| 老系统只支持 JSONP | E 仅维持遗留，新功能勿扩 |

---

## 6. 自检清单

- 确认是否真跨域：协议 / 域名 / 端口是否有一处不同。
- 看 Network：是正式请求被拦，还是 `OPTIONS` 预检失败。
- 需要 Cookie 时：前端 `credentials`、后端 `Allow-Credentials`、具体 Origin（非 `*`）是否齐全。
- 用 Postman 能通、浏览器不通：优先查 CORS，而不是先怀疑业务逻辑写错。
`)