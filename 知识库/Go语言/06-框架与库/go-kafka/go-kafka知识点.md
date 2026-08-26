# Go 语言 Kafka 知识点

> **Go 操作 Kafka 的主流客户端是 `segmentio/kafka-go`（纯 Go）与 `confluent-kafka-go`（C 绑定，性能更高）**。本篇文章从选型、生产者、消费者、Consumer Group、消息可靠性与顺序性、分区分配策略到调优与面试题,系统梳理 Go 生态下的 Kafka 开发。

---

## 一、Go 客户端选型

| 客户端 | 特性 | 适用场景 |
|--------|------|----------|
| **segmentio/kafka-go** | 纯 Go 实现、无 CGO 依赖、API 简洁、支持 Reader/Writer 高级抽象 | **大多数业务（推荐）** |
| **confluent-kafka-go** | 包装 C 库 librdkafka、性能高、配置丰富 | 高吞吐、已有 librdkafka 经验 |
| **Shopify/sarama** | 老牌纯 Go 客户端、API 偏底层、功能全但较繁琐 | 老项目、需要细粒度控制 |

> 选择建议：新项目无 CGO 限制、追求简单 → **kafka-go**；追求极致吞吐、能接受 CGO → **confluent-kafka-go**。

```bash
# 安装 kafka-go
go get github.com/segmentio/kafka-go
```

---

## 二、环境准备

### 2.1 本地快速起一个 Kafka

Kafka 3.x 起支持 **KRaft 模式（无需 ZooKeeper）**，本地用 Docker 一条命令即可：

```bash
docker run -d --name kafka -p 9092:9092 \
  -e KAFKA_CFG_NODE_ID=1 \
  -e KAFKA_CFG_PROCESS_ROLES=broker,controller \
  -e KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=1@localhost:9093 \
  -e KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093 \
  -e KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER \
  bitnami/kafka:latest
```

### 2.2 核心概念速览（Go 视角）

| 概念 | 说明 |
|------|------|
| **Topic** | 消息的逻辑分类 |
| **Partition** | Topic 的物理分片，一个分区内消息**有序**，跨分区无序 |
| **Offset** | 分区内消息的单调递增序号 |
| **Consumer Group** | 一组消费者共享一个 Topic 的分区，实现**负载均衡与水平扩展** |
| **Broker** | Kafka 服务器节点，分区 Leader 所在节点负责读写 |

---

## 三、生产者（Producer / Writer）

### 3.1 kafka-go 的 Writer

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	writer := &kafka.Writer{
		Addr:         kafka.TCP("localhost:9092"),
		Topic:        "order-events",
		Balancer:     &kafka.Hash{},          // 分区分配策略（见下文）
		RequiredAcks: kafka.RequireAll,       // 等待 ISR 全部确认（可靠性）
		Async:        false,                  // 同步发送（便于看结果）
		BatchTimeout: 10 * time.Millisecond,  // 批量攒消息的时间
	}
	defer writer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := writer.WriteMessages(ctx,
		kafka.Message{Key: []byte("user-1"), Value: []byte("order created: 1001")},
		kafka.Message{Key: []byte("user-2"), Value: []byte("order created: 1002")},
	)
	if err != nil {
		log.Fatal("写入失败:", err)
	}
	fmt.Println("发送成功")
}
```

### 3.2 关键配置

| 配置 | 作用 |
|------|------|
| `Balancer` | 分区选择策略：`RoundRobin`（轮询）、`Hash`（同 Key 同分区）、`LeastBytes`（最小字节） |
| `RequiredAcks` | `RequireNone`（不确认）/ `RequireOne`（Leader 确认）/ `RequireAll`（ISR 全确认） |
| `BatchTimeout` / `BatchSize` | 批量攒批参数，影响吞吐与延迟 |
| `Async` | 异步发送提升吞吐，但错误需通过 `writer.Errors()` 通道读取 |

---

## 四、消费者（Consumer / Reader）

### 4.1 单消费者 Reader

```go
reader := kafka.NewReader(kafka.ReaderConfig{
	Brokers:     []string{"localhost:9092"},
	Topic:       "order-events",
	Partition:   0,          // 单分区消费可指定
	MinBytes:    1,          // 一次拉取的最小字节
	MaxBytes:    10e6,       // 一次拉取的最大字节（默认 10MB）
	GroupID:     "order-svc",// 设置 GroupID 即加入消费者组（见下文）
	StartOffset: kafka.FirstOffset, // 从头消费；LastOffset 从最新开始
})

for {
	msg, err := reader.ReadMessage(context.Background())
	if err != nil {
		log.Fatal("读取失败:", err)
	}
	fmt.Printf("收到: partition=%d offset=%d key=%s value=%s\n",
		msg.Partition, msg.Offset, msg.Key, msg.Value)
}
```

### 4.2 消费者组（Consumer Group）

设置 `GroupID` 后，kafka-go 自动处理组管理与分区再分配（Rebalance）。**组内一个分区同一时刻只被一个消费者消费**，组间互相独立（可广播）。

```go
reader := kafka.NewReader(kafka.ReaderConfig{
	Brokers:  []string{"localhost:9092"},
	GroupID:  "order-svc", // 组名
	Topic:    "order-events",
	MinBytes: 1,
	MaxBytes: 10e6,
})

// 提交 offset 的两种策略：
// 1. 自动提交（默认，提交周期 FetchMinBytes? 实际是 CommitInterval）
//    reader.Config().CommitInterval
// 2. 手动提交：ReadMessage 后调用 reader.CommitMessages(ctx, msg)
//    保证"处理完再提交"，避免消息丢失/重复
```

---

## 五、消息可靠性：ACK 与 Offset

### 5.1 生产者端三种 ACK

| `RequiredAcks` | 含义 | 可靠性 | 性能 |
|----------------|------|--------|------|
| `RequireNone` | 发完即走，不确认 | 最低（可能丢） | 最高 |
| `RequireOne` | Leader 写入成功即返回 | 中（Leader 挂可能丢） | 中 |
| `RequireAll` | **ISR 所有副本确认** | 最高（不丢已确认消息） | 较低 |

> 搭配 `kafka.Writer{RequiredAcks: kafka.RequireAll}` 可做到"**已确认消息不丢失**"（除非所有副本同时宕机）。

### 5.2 消费者端：At Least Once 与幂等

Kafka 默认是 **At Least Once（至少一次）**：消费者可能重复收到消息（处理完没来得及提交 offset 就崩溃）。解决方案：

1. **业务幂等**：消费侧用唯一键（订单号、事件 ID）去重；
2. 手动提交：处理成功后再 `CommitMessages`，缩短重复窗口。

```go
// 手动提交示例：先处理后提交
msg, err := reader.FetchMessage(ctx) // 只拉取，不自动提交
if err != nil {
	log.Fatal(err)
}
if err := process(msg); err != nil { // 处理失败则不提交 → 下次重新拉取
	return err
}
if err := reader.CommitMessages(ctx, msg); err != nil { // 处理成功才提交
	log.Fatal(err)
}
```

> **At Most Once（至多一次）**：先提交 offset 再处理，消息可能丢。**Exactly Once（精确一次）**：需 Kafka 事务 + 幂等生产者，成本高，多数业务用"至少一次 + 幂等"即可。

---

## 六、消息顺序性

### 6.1 Kafka 的顺序保证范围

- **分区内有序**：同一 Partition 内消息按 offset 顺序追加，消费者按顺序读取；
- **跨分区无序**：不同 Partition 之间没有全局顺序。

### 6.2 如何保证同业务的有序？

核心思路：**让同一业务键（如同一订单、同一用户）进入同一分区**。

```go
writer := &kafka.Writer{
	Addr:     kafka.TCP("localhost:9092"),
	Topic:    "order-events",
	Balancer: &kafka.Hash{}, // 按 Key 哈希选分区 → 同 Key 同分区 → 同分区有序
}
```

- 生产者：`Balancer: &kafka.Hash{}`，并给消息设置 `Key`；
- 消费者：**单分区单线程消费**（一个分区只被组内一个消费者处理，且该消费者内避免并发处理同一 Key）；
- 若要"按 Key 串行处理"，可在消费者内用 `Key % workerNum` 做本地分片 + 每分片一个 goroutine。

---

## 七、分区分配策略（Balancer）

kafka-go 提供多种分区选择策略：

| 策略 | 行为 | 适用 |
|------|------|------|
| `kafka.RoundRobin` | 轮询，均匀分散 | 消息无 Key、希望均衡分布 |
| `kafka.Hash` | 按 Key 哈希选分区 | **需要同 Key 有序** |
| `kafka.LeastBytes` | 选当前字节最少的分区 | 负载均衡更精细 |
| `kafka.Random` | 随机 | 极少数场景 |

```go
// 实现自定义分配：只要实现 kafka.Balancer 接口
type MyBalancer struct{}
func (b *MyBalancer) Balance(msg kafka.Message, partitions ...int) int {
	// 自定义选分区逻辑
	return partitions[0]
}
```

---

## 八、常见坑与最佳实践

### 8.1 常见坑

1. **创建 Reader 后必须先 close**：`defer reader.Close()`，否则资源泄漏；
2. **手动提交 vs 自动提交混淆**：需要"至少一次 + 幂等"时用手动提交，先处理再提交；
3. **StartOffset 只在新建/无提交记录时生效**：已有 offset 后从记录处继续；
4. **消费者拉取延迟**：`MinBytes`/`MaxWait` 影响延迟，实时场景调小；
5. **单分区吞吐瓶颈**：业务量上来了要增加分区数（分区数决定并行消费上限）。

### 8.2 最佳实践清单

- 生产：`RequiredAcks: RequireAll` + 合理 `BatchTimeout`（吞吐/延迟平衡）；
- 消费：**GroupID 实现水平扩展**，处理完再提交 offset；
- 顺序：关键业务用 `Key` + `Hash` 分区 + 单分区单线程消费；
- 幂等：消费侧按事件 ID 去重，容忍"至少一次"；
- 监控：跟踪 `producer_error`、`consumer_lag`（消费滞后）等指标。

---

## 九、高频自测

1. Go 主流的两个 Kafka 客户端？
   → **kafka-go（纯 Go）** 与 **confluent-kafka-go（C 绑定）**。
2. 如何保证同一订单的消息有序？
   → **用 Key + Hash 分区让同 Key 进同一分区**，分区内有序，再单分区单线程消费。
3. 生产端三种 ACK 哪个最可靠？
   → **RequireAll（ISR 全确认）**。
4. 消费者怎么做到不丢消息？
   → **处理成功后再提交 offset（手动提交）**，并配合幂等。
5. Kafka 默认的投递语义？
   → **At Least Once（至少一次）**，需业务幂等兜底。
6. 怎么水平扩展消费者？
   → **同一 GroupID 多实例**，分区自动分配给组内消费者。

---

## 十、一句话总结

> **Go + Kafka：kafka-go 写生产者（Writer + Balancer + RequiredAcks）、消费者（Reader + GroupID）**；可靠性靠"RequireAll 确认 + 处理完再提交 offset"（至少一次 + 幂等），顺序性靠"Key + Hash 分区 + 分区内单线程消费"。选型、配置、可靠性、顺序、幂等是五大考点。
