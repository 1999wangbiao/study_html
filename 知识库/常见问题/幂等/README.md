# Go 后端常见幂等问题完整整理

独立专题。每个问题按同一结构展开：

1. 问题说明  
2. 典型后果  
3. 处理方式  
4. 示范代码（注释清晰）  
5. 错误写法对比  
6. 自测要点  

约定：

- **幂等（Idempotency）**：一个操作执行 **一次** 和执行 **多次**，对系统状态的影响完全一致 → `f(f(x)) = f(x)`
- 幂等 ≠ 消灭重复请求；而是「重复请求不会造成重复副作用」
- 重复请求天然存在（双击、重试、MQ 重投、回调多次到达）；**幂等由服务端保证**，前端只能改善体验
- 选型原则：**先靠 DB 约束兜底，再用业务状态/锁做精细控制**

---

## 0. 总览表

| # | 问题 | 一句话处理 |
|---|------|------------|
| 1 | 重复提交 / 双击 | 唯一 Token + 服务端校验 |
| 2 | 网络重试导致重复扣款 | 唯一业务键 + 防重表 / 唯一索引 |
| 3 | MQ 消息重复消费 | 业务键去重，DB 唯一索引或 Redis SETNX |
| 4 | 更新类操作被并发覆盖（ABA） | 乐观锁 `version` |
| 5 | 状态机未收口导致重复流转 | `UPDATE ... WHERE status = 预期` |
| 6 | 接口设计未考虑幂等 | 幂等方法优先；POST 加业务键 |
| 7 | 并发下单超卖 | 库存乐观锁或唯一索引 |
| 8 | 分布式锁未覆盖异常路径 | 锁 + 业务回滚 + 防重表 |
| 9 | 防重表 / Token 记录膨胀 | TTL 或定期归档 |
| 10 | 幂等键设计不当 | 业务唯一键优先；客户端 token 作辅助 |

公共骨架（后面代码都基于此）见 [附录 A](#附录-a公共骨架完整注释版)。

---

## 1. 问题：重复提交 / 双击

### 1.1 问题说明

用户网络慢时狂点「提交」「支付」，或前端未做防抖。  
每个请求都带业务意图，服务端没有任何去重，于是同一次提交被处理多次。

### 1.2 典型后果

- 同一笔订单被创建多次
- 同一笔支付被扣款多次
- 同一条评论被插入多份

### 1.3 处理方式

| 步骤 | 做什么 |
|------|--------|
| 1 | 关键写操作前，前端先向后端申请一个一次性 Token |
| 2 | 提交时把 Token 带上；服务端校验存在并原子删除 |
| 3 | 校验失败即拒绝；删除成功才执行业务 |
| 4 | 前端按钮提交后立即禁用（辅助，不是安全边界） |

Token 请求链路：

```text
1. GET  /api/idempotent/token     → 返回 token（写入 Redis，TTL 10~30 分钟）
2. POST /api/order  Header: X-Idempotent-Token: <token>
   → 服务端 DEL token；返回 1 才放行
```

### 1.4 示范代码

```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// IdempotentToken 要求请求头携带有效的幂等 Token。
// 用法：api.POST("/order", middleware.IdempotentToken(rdb), createOrderHandler)
func IdempotentToken(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Idempotent-Token")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "missing idempotent token"})
			return
		}

		// 原子删除：只有第一个请求能删掉，其余全部拒绝
		// 用 Lua 脚本保证 GET + DEL 的原子性
		script := redis.NewScript(`
			if redis.call("GET", KEYS[1]) == ARGV[1] then
				return redis.call("DEL", KEYS[1])
			else
				return 0
			end
		`)
		key := "idem:token:" + token
		n, err := script.Run(c, rdb, []string{key}, token).Int64()
		if err != nil || n == 0 {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"msg": "duplicate request"})
			return
		}
		c.Next()
	}
}
```

签发 Token：

```go
// IssueToken 生成一次性幂等 Token，写入 Redis 并返回给前端。
func IssueToken(c *gin.Context, rdb *redis.Client) {
	t := uuid.NewString()
	key := "idem:token:" + t
	// TTL 不宜过长，30 分钟覆盖大多数提交场景
	if err := rdb.Set(c, key, t, 30*time.Minute).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": t})
}
```

路由示例：

```go
r := gin.Default()

// 申请 Token：可要求登录态
r.GET("/api/idempotent/token", middleware.JWTAuth(), issueTokenHandler)

// 提交订单：JWT + 幂等 Token 双保险
api := r.Group("/api", middleware.JWTAuth())
{
	api.POST("/order", middleware.IdempotentToken(rdb), createOrderHandler)
	api.POST("/payment", middleware.IdempotentToken(rdb), payHandler)
}
```

### 1.5 错误写法对比

```go
// ❌ 用 GET 后再 DEL，非原子：高并发下多个请求都能 GET 到
v, _ := rdb.Get(ctx, key).Result()
if v == token {
	rdb.Del(ctx, key)
	c.Next()
	return
}
c.AbortWithStatus(http.StatusConflict)

// ❌ 只看前端按钮 disabled，没有服务端校验
// （绕过 UI 直接 curl 即可重复提交）

// ✅ 用 Lua 脚本原子 GET + DEL
```

### 1.6 自测要点

-  同一 Token 提交两次：第一次 `200`，第二次 `409`
-  不带 Token → `400`
-  Token 过期（30 分钟后）→ `409`
-  直接 curl 绕过前端 disabled 按钮，重复提交仍被拦截

---

## 2. 问题：网络重试导致重复扣款

### 2.1 问题说明

支付 / 下单链路中，客户端或网关因超时自动重试。每次请求都带业务字段（订单号、外部流水号），但服务端不识别「这笔业务我已经处理过」。

### 2.2 典型后果

- 同一笔订单扣款 2 次
- 调用方多次重试后余额被扣穿
- 对账时才发现，资金事故

### 2.3 处理方式

| 步骤 | 做什么 |
|------|--------|
| 1 | 调用方生成全局唯一业务键（订单号 + 业务类型，或外部流水号） |
| 2 | 服务端在执行业务前，先写防重表或唯一索引 |
| 3 | 写入成功 → 首次请求，执行业务 |
| 4 | 写入冲突 → 重复请求，直接返回首次结果 |
| 5 | 业务执行失败时回滚防重记录，允许下次重试 |

### 2.4 示范代码

方案 A：DB 唯一索引（推荐，最可靠）。

```go
package service

import "errors"

var ErrDuplicateRequest = errors.New("duplicate request")

// PayRequest 支付请求：业务键由调用方保证全局唯一。
type PayRequest struct {
	OutTradeNo string // 调用方业务流水号，全局唯一
	UserID     uint
	OrderID    uint
	Amount     int64 // 分
}

// Pay 幂等支付：以 OutTradeNo 为唯一键。
func (s *PayService) Pay(req PayRequest) (*PayResult, error) {
	// 1) 先写防重表（含请求摘要），靠唯一索引拦重复
	rec := &IdempotentRecord{
		Key:       "pay:" + req.OutTradeNo,
		Status:    "processing",
		CreatedAt: time.Now(),
	}
	if err := s.repo.InsertIdempotent(rec); err != nil {
		if isDuplicateKeyErr(err) {
			// 2) 重复请求：返回首次的处理结果
			return s.fetchExistingResult(req.OutTradeNo)
		}
		return nil, err
	}

	// 3) 首次：执行支付
	res, err := s.gateway.Charge(req)
	if err != nil {
		// 业务失败：回滚防重记录，允许下次重试
		_ = s.repo.DeleteIdempotent(rec.Key)
		return nil, err
	}

	// 4) 成功：更新防重记录为 done，并落结果
	_ = s.repo.MarkIdempotentDone(rec.Key, res)
	return res, nil
}
```

方案 B：Redis SETNX（适合短时去重，性能高）。

```go
// PayWithRedis 用 SETNX 做前置去重。
func (s *PayService) PayWithRedis(ctx context.Context, req PayRequest) (*PayResult, error) {
	key := "idem:pay:" + req.OutTradeNo

	// SET NX EX：只有第一个请求能写入
	ok, err := s.rdb.SetNX(ctx, key, "processing", 10*time.Minute).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		// 重复请求：返回已有结果
		return s.fetchExistingResult(ctx, req.OutTradeNo)
	}

	res, err := s.gateway.Charge(req)
	if err != nil {
		// 失败：删 key 允许重试；不能让一次失败永久占用名额
		s.rdb.Del(ctx, key)
		return nil, err
	}
	// 成功：把结果写回 Redis，供重复请求读取
	s.rdb.Set(ctx, "idem:pay:result:"+req.OutTradeNo, encode(res), 24*time.Hour)
	return res, nil
}
```

防重表结构：

```sql
CREATE TABLE idempotent_records (
  key         VARCHAR(128) PRIMARY KEY,     -- 业务键，唯一
  status      VARCHAR(16)  NOT NULL,        -- processing / done / failed
  result      TEXT,                          -- 成功后落 JSON 结果
  created_at  DATETIME     NOT NULL,
  updated_at  DATETIME     NOT NULL,
  UNIQUE KEY uk_key (key)                    -- 兜底唯一约束
);
```

### 2.5 错误写法对比

```go
// ❌ 完全无去重：网络重试一次扣款一次
func (s *PayService) Pay(req PayRequest) (*PayResult, error) {
	return s.gateway.Charge(req)
}

// ❌ 用 select 后 insert，存在并发窗口
exist, _ := s.repo.Exist(outTradeNo)
if !exist {
	s.repo.Insert(...) // 两个并发请求都可能走到这里
}

// ❌ 失败不回滚防重记录：一次网络抖动让这笔业务永远无法重试

// ✅ DB 唯一索引 + 失败回滚 + 成功落结果
```

### 2.6 自测要点

-  同一 `OutTradeNo` 请求两次：只扣款一次，第二次返回首次结果
-  支付网关失败后，防重记录被清理，可重试
-  并发请求同一业务键：只有一个成功，其余拒绝或返回结果
-  模拟网关重试 3 次，最终只扣款 1 次

---

## 3. 问题：MQ 消息重复消费

### 3.1 问题说明

主流 MQ（Kafka、RabbitMQ、RocketMQ）默认是 **at-least-once** 投递：消费者处理完但 ack 丢失时，broker 会重投。  
若消费逻辑没有幂等，同一条消息会被处理多次。

### 3.2 典型后果

- 同一订单状态被改多次
- 同一条积分发放记录被加多次
- 同一通知被发多次

### 3.3 处理方式

| 步骤 | 做什么 |
|------|--------|
| 1 | 提取消息中的业务唯一键（订单号、事件 ID、`messageId`） |
| 2 | 消费前先判重：DB 唯一索引、防重表、Redis SETNX 任选 |
| 3 | 处理成功后再 ack；处理失败回滚 + 不 ack 或 nack 重投 |
| 4 | 注意：DB 唯一索引 + 业务事务一起提交，避免「记了已处理但业务没成」 |

### 3.4 示范代码

```go
package consumer

import (
	"context"
	"encoding/json"
)

// OrderPaidEvent 订单已支付事件。
type OrderPaidEvent struct {
	EventID   string `json:"event_id"`   // 事件唯一 ID，生产端用 UUID
	OrderID   uint   `json:"order_id"`
	UserID    uint   `json:"user_id"`
	Amount    int64  `json:"amount"`
}

// Handle 消费订单支付事件：以 EventID 做幂等。
func (h *OrderPaidHandler) Handle(ctx context.Context, body []byte) error {
	var ev OrderPaidEvent
	if err := json.Unmarshal(body, &ev); err != nil {
		return err // 解析失败：直接 ack，避免毒消息
	}

	// 1) 在事务内插防重记录 + 执行业务，一起提交
	//    唯一索引冲突 = 重复消息
	err := h.tx.Run(ctx, func(tx *gorm.DB) error {
		rec := &IdempotentRecord{
			Key:      "mq:order_paid:" + ev.EventID,
			Status:   "done",
		}
		if err := tx.Create(rec).Error; err != nil {
			if isDuplicateKeyErr(err) {
				return ErrDuplicateRequest // 重复：上层 ack 即可
			}
			return err
		}
		// 业务：加积分、发券、通知等
		return h.points.Grant(ctx, tx, ev.UserID, ev.Amount/100)
	})

	if errors.Is(err, ErrDuplicateRequest) {
		return nil // 已处理过，直接当成功 ack
	}
	return err
}
```

消费循环：

```go
// ConsumeLoop 消费循环：处理失败 nack 重投，处理成功 ack。
func (w *Worker) ConsumeLoop(ctx context.Context, deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		err := w.handler.Handle(ctx, d.Body)
		if err == nil {
			d.Ack(false) // 处理成功：确认
			continue
		}
		// 处理失败：nack 且 requeue
		// 注意：要设置最大重试次数，避免毒消息无限循环
		d.Nack(false, true)
	}
}
```

Redis 版本（高频小消息去重）：

```go
func (h *OrderPaidHandler) HandleRedis(ctx context.Context, body []byte) error {
	var ev OrderPaidEvent
	json.Unmarshal(body, &ev)

	key := "idem:mq:order_paid:" + ev.EventID
	// SETNX + 较长 TTL（如 7 天），覆盖 MQ 的重投窗口
	ok, _ := h.rdb.SetNX(ctx, key, "1", 7*24*time.Hour).Result()
	if !ok {
		return nil // 重复消息，直接 ack
	}
	return h.points.Grant(ctx, nil, ev.UserID, ev.Amount/100)
}
```

### 3.5 错误写法对比

```go
// ❌ 完全无幂等：ack 丢了 → 重投 → 又加一次积分
func (h *OrderPaidHandler) Handle(body []byte) error {
	json.Unmarshal(body, &ev)
	return h.points.Grant(ev.UserID, ev.Amount/100)
}

// ❌ 先 select 后 insert，跨事务有窗口
if exist { return nil }
h.points.Grant(...)
h.repo.InsertIdempotent(...)

// ❌ 防重记录与业务分两个事务：可能记了已处理但业务没成
tx.Begin(); tx.Create(rec); tx.Commit() // 事务 1
h.points.Grant(...)                      // 事务 2，失败则丢消息

// ✅ 防重记录 + 业务在同一事务内一起提交
```

### 3.6 自测要点

-  同一 `event_id` 投递两次：只加一次积分
-  业务处理失败：消息被 nack 重投，下次能继续处理
-  防重记录与业务在同一事务，不会出现「记了已处理但业务没成」
-  Redis 版本 TTL 大于 MQ 最大重投窗口

---

## 4. 问题：更新类操作被并发覆盖（ABA）

### 4.1 问题说明

并发更新同一资源，后写覆盖先写。经典场景：用户 A、B 同时改同一笔记，A 先提交，B 后提交，B 把 A 的修改覆盖掉。

### 4.2 典型后果

- 数据丢失（A 的修改没了）
- 计数错乱：库存 -1 两次实际只 -1
- 评论数、点赞数统计偏差

### 4.3 处理方式

| 方案 | 做法 | 适用 |
|------|------|------|
| A. 乐观锁 | 表加 `version`，更新时 `WHERE version = ?` | 大部分更新场景，冲突不频繁 |
| B. 唯一键 + 状态 | 在状态字段上加条件 | 状态机类业务 |
| C. CAS 更新 | 直接 `UPDATE SET count = count - 1 WHERE count > 0` | 计数、库存类 |
| D. 悲观锁 | `SELECT ... FOR UPDATE` | 冲突非常频繁、复杂事务 |

### 4.4 示范代码

方案 A：乐观锁 `version`。

```go
package repo

// Note 笔记表：version 用于乐观锁。
type Note struct {
	ID      uint
	UserID  uint
	Title   string
	Body    string
	Version int    // 乐观锁字段
}

// Update 更新笔记：必须带上 version 条件。
// n == 0 说明并发冲突或记录不存在。
func (r *NoteRepo) Update(note *Note) (int64, error) {
	result := r.db.Model(&Note{}).
		Where("id = ? AND version = ?", note.ID, note.Version).
		Updates(map[string]interface{}{
			"title":   note.Title,
			"body":     note.Body,
			"version": gorm.Expr("version + 1"),
		})
	return result.RowsAffected, result.Error
}
```

```go
// Service：读取 → 修改 → 写回，冲突时返回明确错误。
func (s *NoteService) Update(userID uint, id uint, title, body string) error {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return ErrNotFound
	}
	if !s.canEdit(note, userID) {
		return ErrForbidden
	}

	note.Title = title
	note.Body = body

	n, err := s.repo.Update(note)
	if err != nil {
		return err
	}
	if n == 0 {
		// 并发冲突：让上层决定重试或报错给用户
		return ErrConcurrentUpdate
	}
	return nil
}
```

方案 C：CAS 直接更新（适合计数、库存）。

```go
// DeductStock 扣库存：用 CAS 保证不超卖。
// 行级原子操作，无需显式锁。
func (r *ProductRepo) DeductStock(productID uint, qty int) error {
	result := r.db.Model(&Product{}).
		Where("id = ? AND stock >= ?", productID, qty).
		Update("stock", gorm.Expr("stock - ?", qty))
	if result.RowsAffected == 0 {
		return ErrOutOfStock // 库存不足或不存在
	}
	return result.Error
}
```

### 4.5 错误写法对比

```go
// ❌ 读 → 改 → 全字段保存，无 version：B 后提交覆盖 A
note, _ := repo.FindByID(id)
note.Title = newTitle
repo.Save(note) // 全字段 UPDATE，会覆盖期间的并发修改

// ❌ 库存先读再算，再写
p, _ := repo.Find(productID)
p.Stock = p.Stock - 1  // 内存计算
repo.Save(p)            // 高并发下多个请求都读到 stock=10，最后都写 9 → 超卖

// ✅ 乐观锁 version 或 CAS 原子更新
```

### 4.6 自测要点

-  两个并发更新同一笔记：失败的那个返回 `ErrConcurrentUpdate`
-  库存为 0 时 `DeductStock` 失败
-  100 并发扣 1 件库存：最终库存精确减少，不超卖不少卖
-  version 字段在每次更新后递增

---

## 5. 问题：状态机未收口导致重复流转

### 5.1 问题说明

订单、审批、工单等有状态流转的业务，更新时只 `UPDATE SET status = 新状态 WHERE id = ?`，不校验当前状态。  
并发请求或重试请求可能让状态回退或被多次触发。

### 5.2 典型后果

- 已支付的订单被重复「取消」
- 已发货订单被重新「发货」触发多次物流
- 审批已通过又被「驳回」

### 5.3 处理方式

| 规则 | 做法 |
|------|------|
| 状态收口 | `UPDATE ... WHERE id = ? AND status = 期望前置状态` |
| 不允许跳变 | 在 Service 层显式定义合法流转图 |
| 返回值判定 | `RowsAffected == 0` 即非法流转，返回错误 |
| 配合幂等 | 状态字段本身就承担了「是否已处理」的判定 |

### 5.4 示范代码

```go
package service

// 订单状态枚举
const (
	StatusCreated   = "created"
	StatusPaid      = "paid"
	StatusShipped   = "shipped"
	StatusCanceled  = "canceled"
	StatusCompleted = "completed"
)

// 状态流转图：定义合法的前后关系
var orderTransitions = map[string]map[string]bool{
	StatusCreated:   {StatusPaid: true, StatusCanceled: true},
	StatusPaid:      {StatusShipped: true, StatusCanceled: true},
	StatusShipped:   {StatusCompleted: true},
	StatusCanceled:  {},
	StatusCompleted: {},
}

// canTransition 判断从 from 到 to 是否合法。
func canTransition(from, to string) bool {
	return orderTransitions[from][to]
}

// Pay 订单支付：只在「created」时允许变「paid」。
func (s *OrderService) Pay(orderID uint) error {
	// 关键：WHERE status = 'created'，原子判定 + 状态推进
	result := s.db.Model(&Order{}).
		Where("id = ? AND status = ?", orderID, StatusCreated).
		Updates(map[string]interface{}{
			"status":     StatusPaid,
			"paid_at":    time.Now(),
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// 可能是已支付、已取消、不存在；查一次区分
		o, err := s.repo.FindByID(orderID)
		if err != nil {
			return ErrNotFound
		}
		if o.Status == StatusPaid {
			return nil // 重复支付请求：返回成功即可（幂等）
		}
		return ErrIllegalTransition
	}
	return nil
}

// Cancel 取消订单：只在「created」「paid」时可取消。
func (s *OrderService) Cancel(orderID uint) error {
	result := s.db.Model(&Order{}).
		Where("id = ? AND status IN ?", orderID, []string{StatusCreated, StatusPaid}).
		Update("status", StatusCanceled)
	if result.RowsAffected == 0 {
		return ErrIllegalTransition
	}
	return result.Error
}
```

### 5.5 错误写法对比

```go
// ❌ 无条件更新状态：已支付订单也能被取消
db.Model(&Order{}).Where("id = ?", orderID).Update("status", "canceled")

// ❌ 先读后写，跨事务有窗口
o, _ := repo.FindByID(orderID)
if o.Status == "created" {
	o.Status = "paid"
	repo.Save(o) // 并发请求都读到 created，最后都改成 paid
}

// ✅ WHERE status = 前置状态，原子推进
```

### 5.6 自测要点

-  已支付订单再次调 `Pay` → 不报错、不重复扣款（幂等）
-  已取消订单调 `Pay` → `ErrIllegalTransition`
-  已发货订单调 `Cancel` → 失败
-  并发两个 `Pay` 请求：只有一个真正推进状态

---

## 6. 问题：接口设计未考虑幂等

### 6.1 问题说明

接口方法选错或参数设计不当，导致本身就不幂等。例如用 `POST /api/like` 做点赞，每次调用都插入一条记录。

### 6.2 典型后果

- 重复点赞被当成多次点赞
- 重复关注被插入多份
- 接口必须靠业务方做去重，否则天然脆弱

### 6.3 处理方式

| 类型 | 是否天然幂等 | 处理建议 |
|------|--------------|----------|
| GET / HEAD / OPTIONS | 幂等 | 不修改状态 |
| PUT | 幂等 | 用客户端指定的完整资源覆盖 |
| DELETE | 幂等 | 删除第二次结果与第一次一致 |
| POST | 不幂等 | 加业务键 / 唯一索引 / Token |
| PATCH | 不一定 | 用 PATCH 改部分字段时，配合 version |

设计原则：

- **创建资源用 PUT + 客户端生成 ID**：`PUT /api/orders/{order_id}`，重复提交第二次覆盖第一次，最终一致
- **必须用 POST 时**：业务键唯一索引或前置 Token
- **状态变更用 PATCH + 状态条件**：天然幂等

### 6.4 示范代码

```go
// PUT /api/orders/:id 用客户端生成 ID，重复提交天然幂等。
// 第二次提交覆盖第一次（或返回已存在）。
func createOrReplaceOrderHandler(c *gin.Context) {
	orderID := c.Param("id")
	var req struct {
		UserID uint
		Items  []OrderItem
	}
	c.ShouldBindJSON(&req)

	// Upsert：不存在则插入，存在则返回（避免覆盖业务字段）
	result := db.Where("id = ?", orderID).
		Assign(map[string]interface{}{
			"user_id": req.UserID,
			"items":   req.Items,
			"status":   StatusCreated,
		}).
		FirstOrCreate(&Order{ID: orderID})

	c.JSON(http.StatusOK, result.Value)
}

// POST /api/likes 点赞：用复合唯一索引天然幂等。
// 用户对同一目标只能点赞一次，重复点赞返回成功（已点赞）。
func likeHandler(c *gin.Context) {
	var req struct {
		TargetType string
		TargetID   uint
	}
	c.ShouldBindJSON(&req)

	like := &Like{
		UserID:     middleware.UserID(c),
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
	}
	err := db.Create(like).Error
	if isDuplicateKeyErr(err) {
		// 已点赞：幂等返回成功
		c.JSON(http.StatusOK, gin.H{"liked": true})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"liked": true})
}
```

唯一索引：

```sql
-- 点赞表：用户 + 目标组合唯一
CREATE TABLE likes (
  id           BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id      BIGINT NOT NULL,
  target_type  VARCHAR(32) NOT NULL,
  target_id    BIGINT NOT NULL,
  created_at   DATETIME NOT NULL,
  UNIQUE KEY uk_user_target (user_id, target_type, target_id)
);
```

### 6.5 错误写法对比

```go
// ❌ 点赞用纯 INSERT：重复点击就点赞多次
db.Create(&Like{UserID: uid, TargetID: targetID})

// ❌ POST /api/orders 不带业务键：每次提交都生成新订单
db.Create(&Order{ID: uuid.New(), ...})

// ✅ 用复合唯一索引或 PUT + 客户端 ID，天然幂等
```

### 6.6 自测要点

-  同一用户对同一目标连续点赞两次：库中只有一条记录
- 用同一 `order_id` PUT 两次：最终只有一条订单
- DELETE 同一资源两次：第二次不报错（或返回 404 视设计而定，但状态一致）

---

## 7. 问题：并发下单超卖

### 7.1 问题说明

库存扣减用「读 → 算 → 写」三步：先 `SELECT stock`，内存里减一，再 `UPDATE stock = ?`。  
高并发下多个请求同时读到 stock=1，最后都写 0，结果超卖。

### 7.2 典型后果

- 库存为负，发了用户买不到的货
- 库存对不上账
- 资金损失

### 7.3 处理方式

| 方案 | 做法 | 适用 |
|------|------|------|
| A. CAS 原子更新 | `UPDATE ... SET stock = stock - ? WHERE stock >= ?` | 简单计数，首选 |
| B. 乐观锁 | `WHERE id = ? AND version = ?` | 复杂资源更新 |
| C. 唯一索引 | 用户对 SKU 唯一约束 | 防同用户重复下单 |
| D. 悲行锁 / 排队 | `FOR UPDATE` 或 Redis 分布式锁 | 库存极敏感、冲突频繁 |

### 7.4 示范代码

```go
package repo

// DeductStock 扣库存：CAS 原子更新，无需锁。
// RowsAffected == 0 即库存不足。
func (r *ProductRepo) DeductStock(ctx context.Context, productID uint, qty int) error {
	result := r.db.WithContext(ctx).
		Model(&Product{}).
		Where("id = ? AND stock >= ?", productID, qty).
		Update("stock", gorm.Expr("stock - ?", qty))
	if result.RowsAffected == 0 {
		return ErrOutOfStock
	}
	return result.Error
}

// RestoreStock 回库存：同样用 CAS。
func (r *ProductRepo) RestoreStock(ctx context.Context, productID uint, qty int) error {
	return r.db.WithContext(ctx).
		Model(&Product{}).
		Where("id = ?", productID).
		Update("stock", gorm.Expr("stock + ?", qty)).Error
}
```

```go
package service

// PlaceOrder 下单：库存扣减 + 订单创建在同一事务内。
func (s *OrderService) PlaceOrder(ctx context.Context, userID uint, items []OrderItem) (*Order, error) {
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		// 1) 扣库存（CAS）
		for _, it := range items {
			if err := s.productRepo.WithTx(tx).DeductStock(it.ProductID, it.Qty); err != nil {
				return err // 库存不足：整个事务回滚
			}
		}
		// 2) 创建订单
		order := &Order{
			ID:     uuid.NewString(),
			UserID: userID,
			Items:  items,
			Status: StatusCreated,
		}
		return s.orderRepo.WithTx(tx).Create(order)
	})
	if err != nil {
		return nil, err
	}
	return &Order{}, nil
}
```

防同用户重复下单（同 SKU 唯一）：

```sql
-- 用户 + SKU 唯一约束：同用户对同 SKU 同时只能下一单（业务允许时）
CREATE UNIQUE INDEX uk_user_sku ON orders(user_id, product_id, created_at_date);
```

### 7.5 错误写法对比

```go
// ❌ 读 → 算 → 写：超卖
p, _ := repo.Find(productID)
if p.Stock >= qty {
	p.Stock -= qty
	repo.Save(p) // 多个请求都通过 stock >= qty 检查
}

// ❌ 用 SELECT FOR UPDATE 后做内存计算（虽不超卖，但并发退化）
// 简单库存场景不必上锁

// ✅ CAS：UPDATE SET stock = stock - ? WHERE stock >= ?
```

### 7.6 自测要点

-  库存 1，100 并发下单：最终成功 1 单，其余返回 `ErrOutOfStock`
-  库存不会被扣成负数
-  扣库存与创建订单在同一事务，订单失败库存自动回滚

---

## 8. 问题：分布式锁未覆盖异常路径

### 8.1 问题说明

跨服务长流程用 Redis / Zookeeper 分布式锁串行化执行。  
锁获取成功，但业务执行中抛异常 / 网络超时，锁被自动释放，下一请求进来时业务可能尚未完成，造成重复执行。

### 8.2 典型后果

- 锁超时自动释放，业务还在跑，第二个请求进来重复执行
- 异常时锁未释放，后续请求全部阻塞（死锁）
- 与幂等记录脱节，重试仍然能打穿

### 8.3 处理方式

| 步骤 | 做什么 |
|------|--------|
| 1 | 锁必须设合理 TTL，覆盖业务最大执行时间 |
| 2 | 业务异常时显式释放锁（defer） |
| 3 | **锁 + 幂等记录双保险**：锁防并发，幂等防重投 |
| 4 | 锁续期（watchdog）应对长流程 |

### 8.4 示范代码

```go
package service

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// PayWithLock 分布式锁 + 幂等记录双保险。
func (s *PayService) PayWithLock(ctx context.Context, req PayRequest) (*PayResult, error) {
	lockKey := "lock:pay:" + req.OutTradeNo

	// 1) 加锁：TTL 30 秒，覆盖单次支付正常耗时
	token := uuid.NewString()
	ok, err := s.rdb.SetNX(ctx, lockKey, token, 30*time.Second).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrConcurrentRequest // 并发请求：直接拒绝
	}

	// 2) defer 释放：异常时也能释放
	defer func() {
		// 用 Lua 脚本避免误删别人的锁
		script := redis.NewScript(`
			if redis.call("GET", KEYS[1]) == ARGV[1] then
				return redis.call("DEL", KEYS[1])
			else
				return 0
			end
		`)
		script.Run(ctx, s.rdb, []string{lockKey}, token)
	}()

	// 3) 幂等记录：即使锁失效，重复请求也会被防重表拦住
	rec := &IdempotentRecord{Key: "pay:" + req.OutTradeNo, Status: "processing"}
	if err := s.repo.InsertIdempotent(rec); err != nil {
		if isDuplicateKeyErr(err) {
			return s.fetchExistingResult(req.OutTradeNo)
		}
		return nil, err
	}

	// 4) 业务执行
	res, err := s.gateway.Charge(req)
	if err != nil {
		s.repo.DeleteIdempotent(rec.Key) // 失败回滚，允许重试
		return nil, err
	}
	s.repo.MarkIdempotentDone(rec.Key, res)
	return res, nil
}
```

长流程的锁续期（简化版）：

```go
// refreshLock 后台续期，直到 ctx 取消或业务结束。
// 真实场景建议用 redisson-go 等成熟库，避免自己造轮子。
func (s *PayService) refreshLock(ctx context.Context, lockKey, token string, interval, ttl time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 用 Lua 脚本续期：只有持有者能续
			script := redis.NewScript(`
				if redis.call("GET", KEYS[1]) == ARGV[1] then
					return redis.call("EXPIRE", KEYS[1], ARGV[2])
				else
					return 0
				end
			`)
			script.Run(ctx, s.rdb, []string{lockKey}, token, int(ttl.Seconds()))
		}
	}
}
```

### 8.5 错误写法对比

```go
// ❌ 锁不设 TTL：业务崩溃后死锁
rdb.SetNX(ctx, key, token, 0)

// ❌ 直接 DEL 释放：可能误删别人的锁
defer rdb.Del(ctx, lockKey)

// ❌ 只加锁不幂等：锁超时后重投仍会重复执行
// ✅ 锁 + 幂等双保险 + Lua 释放
```

### 8.6 自测要点

-  业务执行中崩溃：锁 TTL 后自动释放
-  业务异常：锁被 defer 正确释放
-  锁失效后重投：幂等记录仍能拦截
-  不持有锁的请求不会误删别人的锁

---

## 9. 问题：防重表 / Token 记录膨胀

### 9.1 问题说明

每次请求都写一条防重记录，长期不清理导致表无限增长。  
Redis 同样会面临 key 过多、内存吃满。

### 9.2 典型后果

- 防重表查询变慢
- Redis 内存吃满触发淘汰
- 业务表（orders、payments）正常，但防重表成了性能瓶颈

### 9.3 处理方式

| 策略 | 做法 |
|------|------|
| A. TTL | Redis key 必须设过期，覆盖业务重投窗口 |
| B. 防重记录带状态 | `done` 状态短保留，`processing` 长保留 |
| C. 定期归档 | 历史 `done` 记录搬到归档表 |
| D. 用业务表兜底 | 订单表本身就有唯一约束，不必额外防重表 |

### 9.4 示范代码

```go
// 清理脚本：定期删 30 天前已完成的防重记录。
// 建议用定时任务（cron）在低峰期执行。
func (r *IdempotentRepo) CleanupOldRecords(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", "done", before).
		Delete(&IdempotentRecord{}).Error
}
```

```go
// 清理脚本主流程：分批删除，避免大事务锁表。
func (w *CleanupWorker) Run(ctx context.Context) error {
	cutoff := time.Now().Add(-30 * 24 * time.Hour)
	for {
		// 每次删 1000 条，避免长事务
		n, err := w.repo.DeleteBatch(ctx, cutoff, 1000)
		if err != nil {
			return err
		}
		if n == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond) // 控制速率
	}
	return nil
}
```

Redis TTL 设计：

```go
// Token / 防重 key 必须设 TTL，且覆盖业务最长重投窗口
// - 短流程（表单提交）：30 分钟
// - 支付场景：24 小时（覆盖对账前重投）
// - MQ 消费：7 天（覆盖 MQ 重投窗口）
rdb.Set(ctx, "idem:token:"+t, t, 30*time.Minute)
rdb.Set(ctx, "idem:pay:"+outTradeNo, "1", 24*time.Hour)
rdb.Set(ctx, "idem:mq:"+eventID, "1", 7*24*time.Hour)
```

### 9.5 错误写法对比

```go
// ❌ 防重记录永久保留：表无限增长
rdb.Set(ctx, key, "1", 0) // 0 表示永不过期

// ❌ 一次性大批量删除：长事务锁表
db.Where("created_at < ?", cutoff).Delete(&IdempotentRecord{})

// ✅ TTL + 分批清理 + 归档
```

### 9.6 自测要点

-  30 天前的 `done` 记录被定时清理
-  Redis 内存占用稳定，不会无限增长
-  清理脚本执行不影响线上业务（分批 + 控速）
-  清理后重复请求重新进入，业务能正常处理

---

## 10. 问题：幂等键设计不当

### 10.1 问题说明

幂等键选错或没有，导致：

- 错把客户端传的随机 ID 当业务键
- 用 `requestId` 当幂等键，但不同业务用同一个，互相覆盖
- 多个不同业务键混用，去重失效

### 10.2 典型后果

- 该拦的拦不住（业务键不唯一）
- 不该拦的被拦（不同业务共用一个键）
- 同一业务被多次处理

### 10.3 处理方式

| 原则 | 说明 |
|------|------|
| 业务唯一键优先 | 用订单号、流水号、事件 ID 等"业务自然唯一"字段 |
| 加业务前缀 | Redis key / 防重表 key 用 `业务类型:业务键` 形式 |
| 客户端 Token 辅助 | 没有自然业务键时才用前端 Token |
| 不要混用 | 支付、下单、消息消费各用各的 key 空间 |

### 10.4 示范代码

```go
package service

// 幂等键构造器：业务前缀 + 业务唯一键，避免不同业务冲突。
const (
	BizPay       = "pay"
	BizOrder     = "order"
	BizMQOrder   = "mq:order_paid"
	BizMQPoints  = "mq:points_grant"
	BizLike      = "like"
)

// idemKey 构造幂等键：业务前缀 : 业务键。
func idemKey(biz, key string) string {
	return "idem:" + biz + ":" + key
}

// 例 1：支付用 OutTradeNo
key := idemKey(BizPay, req.OutTradeNo)
// → idem:pay:PAY202608190001

// 例 2：MQ 消费用 EventID
key := idemKey(BizMQOrder, ev.EventID)
// → idem:mq:order_paid:550e8400-e29b-41d4-a716-446655440000

// 例 3：点赞用 userID + targetType + targetID 复合键
key := idemKey(BizLike, fmt.Sprintf("%d:%s:%d", uid, targetType, targetID))
```

错误案例对照：

```go
// ❌ 用客户端随机 requestId 当业务键：每次都不同，等于没去重
key := "idem:" + c.GetHeader("X-Request-ID")

// ❌ 不同业务共用一个 key：支付和订单互相覆盖
key := "idem:" + req.RequestID // 既有支付又有下单

// ❌ 只用 userID：用户只能下一单
key := "idem:" + strconv.Itoa(int(req.UserID))

// ✅ 业务前缀 + 业务自然唯一键
```

### 10.5 错误写法对比

```go
// ❌ 用时间戳当键
key := "idem:" + strconv.FormatInt(time.Now().Unix(), 10)

// ❌ 用自增 ID
key := "idem:" + strconv.Itoa(int(nextID()))

// ❌ 多业务共用同一前缀
//   支付、退款、订单都用 "idem:business:" + req.BusinessID
//   支付成功后，退款时同 ID 会被误判为已处理

// ✅ 业务前缀隔离 + 业务自然唯一键
```

### 10.6 自测要点

-  同一业务同一键：被去重
-  不同业务即使键相同：互不影响
-  业务键在调用方视角是稳定且唯一的（重试不会变）
-  业务前缀隔离清晰，未来新增业务不会与现有冲突

---

## 11. 决策流程（串起来怎么用）

```text
来了一个写请求
  │
  ├─ 是创建 / 插入吗？
  │    是 → 用业务唯一键 + DB 唯一索引（问题 2、6）
  │
  ├─ 是更新 / 修改吗？
  │    是 → 用乐观锁 version 或 CAS（问题 4）
  │
  ├─ 有状态流转吗？
  │    是 → WHERE status = 前置状态（问题 5）
  │
  ├─ 是 MQ 消费吗？
  │    是 → 业务键去重 + 同事务（问题 3）
  │
  ├─ 涉及并发 / 库存吗？
  │    是 → CAS 原子更新（问题 7）
  │
  ├─ 跨服务长流程？
  │    是 → 分布式锁 + 幂等记录双保险（问题 8）
  │
  └─ 选好幂等键（问题 10），定期清理防重记录（问题 9）
```

覆盖日常约 80% 场景的最小集合：

1. 创建类用 DB 唯一索引
2. 更新类用乐观锁 version
3. 状态机用 `WHERE status = ?`
4. MQ 消费用业务键去重

---

## 12. 总自测清单

-  同一 Token / 业务键提交两次：只处理一次
-  网络重试不会造成重复扣款 / 重复下单
-  MQ 重投：只消费一次
-  并发更新：失败的那个收到明确错误
-  状态流转：非法跳变被拒绝
-  库存不会超卖
-  分布式锁异常路径仍能保证幂等
-  防重记录有 TTL 或定期清理
-  幂等键设计：业务前缀 + 业务唯一键
-  直接 curl 绕过前端，重复请求仍被拦截

---

## 13. 何时再升级（超出基础 80%）

| 现状 | 升级方向 |
|------|----------|
| 单库够用 | DB 唯一索引 + 状态机 |
| 高并发 / 短时去重 | Redis SETNX + TTL |
| 跨服务长流程 | 分布式锁 + 幂等记录双保险 |
| 资金级强一致 | 防重表 + 业务事务 + 对账兜底 |
| 海量历史防重记录 | 分表 / 归档 / 时序库 |
| 多端协同 | 全局业务事件 ID（雪花 / UUID） |

---

## 附录 A：公共骨架（完整注释版）

```go
package apierr

import "errors"

var (
	ErrDuplicateRequest    = errors.New("duplicate request")     // → 409
	ErrConcurrentRequest   = errors.New("concurrent request")    // → 409
	ErrConcurrentUpdate    = errors.New("concurrent update")     // → 409
	ErrOutOfStock          = errors.New("out of stock")           // → 409
	ErrIllegalTransition   = errors.New("illegal state transition") // → 409
	ErrBadRequest          = errors.New("bad request")            // → 400
)
```

```go
package repo

import (
	"errors"

	"gorm.io/gorm"
)

// isDuplicateKeyErr 判断是否唯一索引冲突。
func isDuplicateKeyErr(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
```

```go
package repo

import (
	"time"

	"gorm.io/gorm"
)

// IdempotentRecord 防重记录：通用模板。
type IdempotentRecord struct {
	Key       string    `gorm:"primaryKey;size:128"`
	Status    string    `gorm:"size:16"`  // processing / done / failed
	Result    string    `gorm:"type:text"` // 成功后落 JSON 结果
	CreatedAt time.Time
	UpdatedAt time.Time
}

// InsertIdempotent 插入防重记录：唯一冲突即重复请求。
func (r *IdempotentRepo) InsertIdempotent(rec *IdempotentRecord) error {
	rec.CreatedAt = time.Now()
	rec.UpdatedAt = time.Now()
	return r.db.Create(rec).Error
}

// MarkIdempotentDone 标记防重记录为 done 并落结果。
func (r *IdempotentRepo) MarkIdempotentDone(key, result string) error {
	return r.db.Model(&IdempotentRecord{}).
		Where("key = ?", key).
		Updates(map[string]interface{}{
			"status": "done",
			"result": result,
		}).Error
}

// DeleteIdempotent 删除防重记录：业务失败时回滚，允许重试。
func (r *IdempotentRepo) DeleteIdempotent(key string) error {
	return r.db.Where("key = ?", key).Delete(&IdempotentRecord{}).Error
}

// FindByIdempotent 查询防重记录。
func (r *IdempotentRepo) FindByIdempotent(key string) (*IdempotentRecord, error) {
	var rec IdempotentRecord
	err := r.db.Where("key = ?", key).First(&rec).Error
	return &rec, err
}
```

```go
package service

import "encoding/json"

// fetchExistingResult 重复请求时返回首次的处理结果。
func (s *PayService) fetchExistingResult(outTradeNo string) (*PayResult, error) {
	rec, err := s.repo.FindByIdempotent("pay:" + outTradeNo)
	if err != nil {
		return nil, err
	}
	if rec.Status != "done" {
		// 首次还在处理中：返回特殊码让调用方稍后重试
		return nil, ErrProcessing
	}
	var res PayResult
	if err := json.Unmarshal([]byte(rec.Result), &res); err != nil {
		return nil, err
	}
	return &res, nil
}
```

记忆口诀：

1. **创建靠唯一索引，更新靠乐观锁**  
2. **状态机用 WHERE，库存用 CAS**  
3. **MQ 用业务键，事务一起提交**  
4. **锁防并发，幂等防重投，双保险才稳**  
5. **键要带业务前缀，记录要有 TTL**  
