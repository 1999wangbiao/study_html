# gRPC 知识点（Go 语言）

> **gRPC = Google 开源的高性能 RPC 框架**，默认用 HTTP/2 传输 + Protocol Buffers 序列化，支持四类 API（Unary / Server Streaming / Client Streaming / Bidirectional Streaming）。**服务端是接口定义（IDL），proto 一份、多语言生成代码**，是微服务内部通信的主流方案。

---

## 一、gRPC 是什么

### 1.1 一句话

**gRPC** 是一个跨语言、高性能的远程过程调用（RPC）框架：客户端像调用本地函数一样调用远程服务，底层由框架完成**网络传输 + 序列化 + 服务发现（可选）**。

### 1.2 核心三件套

| 组成 | 作用 | 对应物 |
|------|------|--------|
| **Protocol Buffers（protobuf）** | 定义接口与数据的 IDL + 序列化格式 | 接口契约 |
| **HTTP/2** | 传输层：多路复用、双向流、头部压缩 | 传输协议 |
| **代码生成器（protoc）** | 由 `.proto` 生成各语言的客户端/服务端骨架 | 脚手架 |

### 1.3 与传统 REST 对比

| 维度 | REST（HTTP/1.1 + JSON） | gRPC（HTTP/2 + protobuf） |
|------|--------------------------|---------------------------|
| 接口定义 | 无强契约，靠文档 | **`.proto` 强契约，生成代码** |
| 序列化 | JSON 文本，体积大、慢 | **protobuf 二进制，体积小、快** |
| 传输 | HTTP/1.1 半双工 | **HTTP/2 多路复用 + 双向流** |
| 浏览器支持 | 好 | 差（需 gRPC-Web 网关） |
| 调试/工具 | curl / Postman 方便 | 需 grpcurl / 反射 |
| 适用 | 对外 API、浏览器 | **内部服务间通信、高吞吐低延迟** |

> 一句话选型：**对外的开放 API 用 REST，内部微服务之间用 gRPC。**

---

## 二、Protocol Buffers 基础

### 2.1 一个最小示例

```proto
// proto3 语法
syntax = "proto3";

package hello;

option go_package = "hello/hellopb;hellopb"; // Go 模块路径;包名

// 服务定义：一个接口、一个方法
service Greeter {
  rpc SayHello(HelloRequest) returns (HelloReply);
}

// 请求消息
message HelloRequest {
  string name = 1; // 字段编号：二进制中的标识，一旦发布不可改
}

// 响应消息
message HelloReply {
  string message = 1;
}
```

### 2.2 字段编号是契约的一部分（重点！）

protobuf 序列化时**只传字段编号 + 值**，不传字段名：

```text
field 1 (string name) -> 0x0A 0x05 "world"
```

- 字段编号一旦发布**不能改**（改了就破坏兼容性）；
- 可以增删字段（前提是不要复用已删除的编号）；
- 1~15 编号用 1 字节 tag，优先留给高频字段。

### 2.3 常用标量类型

| proto 类型 | Go 类型 | 说明 |
|-----------|---------|------|
| `string` | `string` | UTF-8 字符串 |
| `int32` / `int64` | `int32` / `int64` | 负数建议用 `sint*` |
| `uint32` / `uint64` | `uint32` / `uint64` | 无符号 |
| `double` / `float` | `float64` / `float32` | 浮点 |
| `bool` | `bool` | 布尔 |
| `bytes` | `[]byte` | 原始字节 |
| `repeated T` | `[]T` | 切片 |
| `map<K,V>` | `map[K]V` | 映射 |

> 注意：protobuf 里没有指针，所有字段都是**可选 + 零值默认**（proto3）。"是否有值"要通过 `optional` + `*T` 或 `oneof` 来表达。

### 2.4 `oneof` / `enum` / 嵌套消息

```proto
message SearchRequest {
  oneof sort_by {      // 多选一，只有一个字段会被设置
    string sort_field = 1;
    int32  sort_index = 2;
  }
}

enum Status {          // 枚举
  STATUS_UNKNOWN = 0;  // 0 是默认值，必须存在
  STATUS_OK = 1;
  STATUS_ERR = 2;
}
```

### 2.5 从 `.proto` 生成 Go 代码

```bash
protoc --go_out=. --go-grpc_out=. hello.proto
```

生成两个文件：

| 文件 | 内容 |
|------|------|
| `hello.pb.go` | 消息结构体 + 序列化/反序列化 |
| `hello_grpc.pb.go` | `GreeterClient` 接口、`GreeterServer` 接口、注册函数 |

> 从 Go 1.20 起推荐用 `google.golang.org/protobuf` 的新 API，旧库 `github.com/golang/protobuf` 已冻结。

---

## 三、四类 API（重点考点）

gRPC 基于 HTTP/2 流，把"请求/响应"扩展为四种调用模型：

| 类型 | 客户端 | 服务端 | 适用场景 |
|------|--------|--------|----------|
| **Unary** | 1 个请求 | 1 个响应 | 普通 CRUD、查询 |
| **Server Streaming** | 1 个请求 | 多个响应（流式） | 拉取列表、日志推送、进度 |
| **Client Streaming** | 多个请求（流式） | 1 个响应 | 上传、批量提交、聚合 |
| **Bidirectional Streaming** | 多个请求 | 多个响应 | 聊天、实时协作 |

```proto
service ChatService {
  rpc GetUser(UserID) returns (User);              // Unary
  rpc ListLogs(LogQuery) returns (stream LogLine); // Server Streaming
  rpc Upload(stream Chunk) returns (UploadStatus); // Client Streaming
  rpc Chat(stream Msg) returns (stream Msg);       // Bidi Streaming
}
```

---

## 四、Go 服务端实现（完整 Demo）

### 4.1 目录结构

```text
hello/
├── go.mod
├── hello.proto              # 接口定义
├── hellopb/                 # 生成代码
│   ├── hello.pb.go
│   └── hello_grpc.pb.go
└── server/main.go           # 服务端
```

### 4.2 服务端代码

```go
package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	"example/hello/hellopb"
)

// 1. 实现生成的接口（结构体嵌入 Unimplemented 保证向前兼容）
type greeterServer struct {
	hellopb.UnimplementedGreeterServer
}

// 2. 实现业务方法
func (s *greeterServer) SayHello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloReply, error) {
	return &hellopb.HelloReply{
		Message: "Hello, " + req.GetName(),
	}, nil
}

func main() {
	// 3. 监听端口
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 4. 创建 gRPC server（默认带日志拦截器、流控等）
	s := grpc.NewServer()

	// 5. 注册服务
	hellopb.RegisterGreeterServer(s, &greeterServer{})

	log.Println("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
```

> 生产环境一定要 `s.GracefulStop()`（先停止接收新请求，等存量请求处理完再退出）而不是 `s.Stop()` 硬杀。

### 4.3 客户端代码

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"example/hello/hellopb"
)

func main() {
	// 1. 建立连接（示例用不安全连接；生产务必用 TLS）
	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// 2. 创建客户端桩
	c := hellopb.NewGreeterClient(conn)

	// 3. 带超时地调用（一定要设 context 超时，防止永久阻塞）
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reply, err := c.SayHello(ctx, &hellopb.HelloRequest{Name: "world"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Println(reply.GetMessage())
}
```

> 新 API 用 `grpc.NewClient`；旧代码里常见的 `grpc.Dial` 已废弃（Connectivity State 语义不同，`Dial` 是阻塞建立，`NewClient` 是懒连接）。

---

## 五、流式调用示例（Server Streaming）

### 5.1 服务端

```go
func (s *greeterServer) ListNames(req *hellopb.ListReq, stream hellopb.Greeter_ListNamesServer) error {
	for _, name := range []string{"A", "B", "C"} {
		if err := stream.Send(&hellopb.HelloReply{Message: name}); err != nil {
			return err // 客户端断开或网络错误
		}
	}
	return nil // 返回 nil 表示流正常结束
}
```

### 5.2 客户端

```go
stream, err := c.ListNames(ctx, &hellopb.ListReq{})
if err != nil {
	log.Fatal(err)
}
for {
	reply, err := stream.Recv()
	if err == io.EOF {
		break // 服务端正常结束
	}
	if err != nil {
		log.Fatal(err) // 出错
	}
	fmt.Println(reply.GetMessage())
}
```

> **注意**：客户端循环里判断 `io.EOF` 是流式调用最常见的模式；`grpc_util.Status` 可拿到更细的错误码。

---

## 六、拦截器（Interceptor / Middleware，高频考点）

拦截器 = gRPC 版的中间件，可统一做日志、鉴权、限流、链路追踪、恢复 panic。

| 位置 | 类型 | 签名 |
|------|------|------|
| 服务端 | `UnaryServerInterceptor` | `func(ctx, req, info, handler) (resp, err)` |
| 服务端 | `StreamServerInterceptor` | 流式版 |
| 客户端 | `UnaryClientInterceptor` | 客户端版 |

### 6.1 服务端 Unary 拦截器

```go
func loggingInterceptor(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req) // 调用真正的方法
	log.Printf("%s took %v", info.FullMethod, time.Since(start))
	return resp, err
}

s := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))
```

### 6.2 链式组合（多个拦截器叠加）

```go
s := grpc.NewServer(
	grpc.ChainUnaryInterceptor(
		loggingInterceptor,  // 先执行
		authInterceptor,     // 再执行
		rateLimitInterceptor,
	),
)
```

---

## 七、认证与加密（TLS / Token）

### 7.1 为什么必须加密

gRPC 默认传输是明文的（对内部也不该裸奔），生产环境必须用 TLS：

```go
// 服务端
creds, _ := credentials.NewServerTLSFromFile("server.crt", "server.key")
s := grpc.NewServer(grpc.Creds(creds))

// 客户端
creds, _ := credentials.NewClientTLSFromFile("ca.crt", "server.example.com")
conn, _ := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(creds))
```

### 7.2 无 TLS 时禁用安全连接（仅本地开发）

```go
grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
```

### 7.3 带 Token 的拦截器鉴权

```go
func authInterceptor(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
	if len(md["authorization"]) == 0 || md["authorization"][0] != "Bearer secret" {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return handler(ctx, req)
}
```

客户端发送元数据：

```go
md := metadata.Pairs("authorization", "Bearer secret")
ctx := metadata.NewOutgoingContext(context.Background(), md)
reply, err := c.SayHello(ctx, &hellopb.HelloRequest{Name: "world"})
```

---

## 八、错误处理与状态码（gRPC Status）

gRPC 用 **Status Code + Message + Details** 表示错误，跨语言一致。

### 8.1 常用状态码

| 常量 | 含义 |
|------|------|
| `codes.OK` | 成功 |
| `codes.InvalidArgument` | 参数错误 |
| `codes.NotFound` | 资源不存在 |
| `codes.AlreadyExists` | 资源已存在 |
| `codes.PermissionDenied` | 权限不足 |
| `codes.Unauthenticated` | 未认证 |
| `codes.ResourceExhausted` | 配额/限流 |
| `codes.Internal` | 服务端内部错误 |
| `codes.Unavailable` | 服务不可用（负载均衡/连接断开时常见）|
| `codes.DeadlineExceeded` | 超时 |

### 8.2 服务端返回错误

```go
return nil, status.Error(codes.NotFound, "user not found")
```

### 8.3 客户端处理错误

```go
reply, err := c.SayHello(ctx, req)
if err != nil {
	st, ok := status.FromError(err)
	if ok && st.Code() == codes.NotFound {
		// 按业务处理
	}
	return
}
```

> **常见坑**：客户端收到的 `err` 是 gRPC 包装过的错误，直接 `==` 比较无效，必须用 `status.FromError(err)` 解包。

---

## 九、超时控制（Deadline 传播）

### 9.1 客户端设置超时

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
reply, err := c.SayHello(ctx, req)
// ctx 到期后，客户端连接自动关闭，服务端会收到 ctx.Done()
```

### 9.2 服务端感知超时

```go
func (s *greeterServer) SayHello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloReply, error) {
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
	case res := <-doHeavyWork():
		return res, nil
	}
}
```

### 9.3 Deadline 会随调用链传播

客户端设置的 deadline 会通过 metadata 传给服务端，服务端再调用下游 gRPC 时同样带上——**整条调用链共用一个截止时间**，防止某个环节无限等待。

---

## 十、负载均衡与连接管理

### 10.1 客户端连接是"长连接"

- `grpc.NewClient` 建立的是 **HTTP/2 长连接**，可复用；
- 不要每次请求都 `NewClient`/`Close`，应复用连接；
- 一个连接内部由 HTTP/2 多路复用承载多个并发请求（并发安全）。

### 10.2 负载均衡

客户端侧负载均衡（client-side load balancing），配合服务发现：

```go
conn, _ := grpc.NewClient(
	"dns:///myservice.internal:50051", // 用 DNS 解析出多个后端
	grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
)
```

生产环境更常见的组合是 **etcd / consul / k8s headless service + resolver 插件**，服务端不感知负载均衡。

---

## 十一、生态与工具

| 工具/库 | 用途 |
|---------|------|
| `protoc` + `protoc-gen-go` | 生成消息代码 |
| `protoc-gen-go-grpc` | 生成服务代码 |
| `grpcurl` | 命令行调 gRPC（类比 curl） |
| **gRPC Reflection** | 运行时暴露 proto 描述，配合 grpcurl 免写客户端 |
| `buf` | 更现代的 proto 管理/构建工具 |
| **grpc-gateway** | 把 gRPC 暴露为 RESTful HTTP（对外用） |
| **gRPC-Web** | 浏览器通过 HTTP/1.1 访问 gRPC 服务 |
| `connect-go` | 更轻量的现代替代（同时支持 HTTP/1.1 + HTTP/2） |

---

## 十二、高频自测

1. gRPC 用哪两个技术做传输和序列化？
   → **HTTP/2 传输 + Protocol Buffers 序列化**。
2. protobuf 里字段编号能否修改？
   → **发布后不能改**，编号是二进制契约的一部分；1~15 留高频字段。
3. 四类 API 是什么？
   → **Unary / Server Streaming / Client Streaming / Bidi Streaming**。
4. 客户端如何处理 gRPC 错误？
   → 用 `status.FromError(err)` 解包，比较 `st.Code()`，不能直接 `==`。
5. 客户端不设超时会怎样？
   → 可能**永久阻塞**，且占用连接；务必用 `context.WithTimeout`。
6. 服务端怎么优雅停机？
   → `s.GracefulStop()`（先拒绝新请求、处理完存量再退出）。
7. 为什么浏览器不方便直接用 gRPC？
   → 浏览器只能用 HTTP/1.1，需要 **gRPC-Web** 或 **grpc-gateway** 中转。

---

## 十三、一句话总结

> **gRPC = HTTP/2 + protobuf 的高性能 RPC 框架**，`.proto` 一份契约多语言生成代码；四类流式 API 覆盖普通调用到实时双向通信；拦截器做中间件、TLS + Token 保安全、Status + Deadline 管错误与超时。**内部微服务通信首选，对外 API 用 REST + grpc-gateway 桥接。**
