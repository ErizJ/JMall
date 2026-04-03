# JMall 秒杀系统：原理、设计与实现

> 本文档基于 JMall 项目的真实代码，从原理到实现完整讲解秒杀系统。
> 技术栈：Go (go-zero) + MySQL + Redis + Kafka
> 目标：读完这篇文档，面试中能把秒杀从头到尾讲清楚。

---

## 一、秒杀的本质问题

秒杀 = 短时间内大量用户抢购少量商品。

核心矛盾：**读写比极端失衡**。假设 100 件商品，10 万人抢购：
- 10 万次请求涌入（读+写）
- 只有 100 次能成功下单（写）
- 99.9% 的请求是"无效"的

所以秒杀系统的设计哲学就一句话：**尽早拦截无效请求，让尽量少的请求到达数据库**。

### 1.1 秒杀面临的技术挑战

| 挑战 | 说明 | 后果 |
|------|------|------|
| 高并发 | 万级甚至十万级 QPS 瞬间涌入 | 服务器扛不住，直接宕机 |
| 超卖 | 库存 100 件，卖出 120 件 | 商家亏钱，用户投诉 |
| 重复下单 | 同一用户抢到多件 | 不公平，违反限购规则 |
| 主站被拖垮 | 秒杀流量冲击正常业务 | 整个商城不可用 |
| 恶意请求 | 机器人/脚本刷接口 | 正常用户抢不到 |

### 1.2 解决思路：分层拦截

```
请求量：100,000
    │
    ▼
[前端] 按钮防抖 + 本地校验 ──→ 拦截 ~30% 无效点击
    │
    ▼  ~70,000
[Nginx] 令牌桶限流 ──→ 超出阈值直接返回 "系统繁忙"
    │
    ▼  ~10,000
[Redis] Lua 原子扣库存 ──→ 库存不足直接返回 "已售罄"
    │
    ▼  ~100（只有库存数量的请求能通过）
[Kafka] 异步投递 ──→ 削峰填谷，不让 MySQL 被打崩
    │
    ▼  ~100
[MySQL] 事务下单 ──→ 最终落盘，WHERE product_num >= num 兜底
```

每一层都在过滤请求，最终到达 MySQL 的只有真正需要下单的那 100 个。

---

## 二、整体架构

### 2.1 架构分层图

```
┌─────────────────────────────────────────────────────────────┐
│                      用户层 (Client)                         │
│  浏览器/App → CDN 静态资源 → 前端倒计时 + 按钮防抖           │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTPS
┌────────────────────────▼────────────────────────────────────┐
│                    接入层 (Gateway)                           │
│  Nginx: 令牌桶限流 + IP 频率限制 + 负载均衡                  │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                    服务层 (Service)                           │
│                                                              │
│  ┌─────────────────┐      ┌─────────────────┐               │
│  │ seckill-service  │      │ order-service    │              │
│  │ · 活动校验       │      │ · 普通下单       │              │
│  │ · Redis 预扣库存 │      │ · 支付联动       │              │
│  │ · 投递 Kafka     │      │                  │              │
│  └────────┬────────┘      └────────▲────────┘               │
│           │                        │                         │
│  ┌────────▼────────────────────────┴───────┐                │
│  │            Kafka: seckill-order          │                │
│  │  · 削峰填谷（万级→千级）                  │                │
│  │  · 异步下单                              │                │
│  └────────────────┬────────────────────────┘                │
│                   │                                          │
│  ┌────────────────▼────────────────────────┐                │
│  │       Kafka Consumer (内嵌在 seckill)    │                │
│  │  · 幂等校验                              │                │
│  │  · MySQL 事务下单                        │                │
│  │  · 写结果到 Redis                        │                │
│  └─────────────────────────────────────────┘                │
│                                                              │
│  ┌─────────────────┐      ┌─────────────────┐               │
│  │ payment-service  │      │ management-svc  │               │
│  │ · 支付（复用）    │      │ · 后台管理      │               │
│  └─────────────────┘      └─────────────────┘               │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                    数据层 (Storage)                           │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────────────┐   │
│  │  Redis   │  │  MySQL   │  │  Kafka Broker             │   │
│  │ · 库存   │  │ · 订单   │  │ · 消息持久化              │   │
│  │ · 限购   │  │ · 库存   │  │ · 16 分区                 │   │
│  │ · 令牌   │  │ · 活动   │  │ · replication.factor=3    │   │
│  │ · 结果   │  │          │  │                           │   │
│  └──────────┘  └──────────┘  └──────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 为什么秒杀服务要独立部署？

秒杀流量是脉冲式的——活动开始的那几秒 QPS 暴涨，之后迅速归零。如果和主站混部署：
- 秒杀把 CPU/内存/连接池吃满 → 正常用户连商品列表都打不开
- 秒杀服务 OOM 崩了 → 连带主站一起挂

独立部署后：秒杀崩了，主站照常运行。这就是**故障隔离**。

---

## 三、核心原理详解

### 3.1 为什么用 Redis 而不是直接打 MySQL？

MySQL 单机写入 QPS 大约 2000-3000（取决于硬件和 SQL 复杂度）。
Redis 单机读写 QPS 可达 10 万+。

秒杀瞬间 1 万个请求同时来：
- 直接打 MySQL → 2000 QPS 上限，剩下 8000 个请求排队等待 → 超时 → 服务崩溃
- 先过 Redis → 10 万 QPS 轻松扛住，只有 100 个（库存数）通过 → MySQL 只处理 100 个

**Redis 在秒杀中就是一道高速过滤器。**

### 3.2 为什么库存扣减必须用 Lua 脚本？

假设不用 Lua，用两条 Redis 命令：
```
stock = GET seckill:stock:1    // 读库存
if stock > 0:
    DECRBY seckill:stock:1 1   // 扣库存
```

问题：GET 和 DECRBY 之间不是原子的。高并发下：
```
时刻1: 请求A GET → stock=1（够）
时刻2: 请求B GET → stock=1（也够，因为A还没扣）
时刻3: 请求A DECRBY → stock=0
时刻4: 请求B DECRBY → stock=-1  ← 超卖了！
```

Lua 脚本在 Redis 中是**原子执行**的（Redis 单线程），整个"检查+扣减"不会被其他命令打断：

```lua
-- 我们实际使用的 Lua 脚本（seckilllogic.go）
local bought = tonumber(redis.call('GET', KEYS[2]) or '0')
if bought + tonumber(ARGV[1]) > tonumber(ARGV[2]) then
    return -1  -- 超出限购
end
local stock = tonumber(redis.call('GET', KEYS[1]))
if stock == nil or stock < tonumber(ARGV[1]) then
    return 0   -- 库存不足
end
redis.call('DECRBY', KEYS[1], ARGV[1])  -- 扣库存
redis.call('INCRBY', KEYS[2], ARGV[1])  -- 记录已购数量
if bought == 0 then
    redis.call('EXPIRE', KEYS[2], 86400) -- 首次购买时设置 TTL
end
return 1  -- 成功
```

一个 Lua 脚本同时完成了：限购检查 + 库存检查 + 库存扣减 + 限购计数 + TTL 设置。
一次网络往返（RTT），五个操作，全部原子。

### 3.3 为什么需要 Kafka？不能 Redis 扣完直接写 MySQL 吗？

可以，但扛不住。

Redis 扣库存后，假设 100 个请求通过了，它们同时去写 MySQL：
- 100 个并发事务同时 INSERT + UPDATE → 行锁竞争 → 性能急剧下降
- 如果是 1000 件库存，1000 个并发事务 → MySQL 直接超载

引入 Kafka 后：
```
Redis 扣库存（微秒级）→ 投递 Kafka（毫秒级）→ 返回用户"排队中"
                                                    ↓
                                        Kafka Consumer 按 MySQL
                                        能承受的速率消费（如 500/s）
```

这就是**削峰填谷**：把瞬时的尖峰流量，通过消息队列拉平成 MySQL 能承受的平滑流量。

用户体验也更好：
- 没有 Kafka：用户等 Redis(1ms) + MySQL(50ms) = 51ms
- 有 Kafka：用户等 Redis(1ms) + Kafka投递(5ms) = 6ms，后台慢慢写库

### 3.4 本地售罄标记是什么？为什么需要？

当 Redis 库存扣到 0 后，后续所有请求都会打到 Redis 去执行 Lua 脚本，虽然 Redis 很快，但如果有 10 万个请求在售罄后继续打 Redis，也是浪费。

解决方案：在服务进程的内存里维护一个标记：

```go
var soldOutMap sync.Map // 进程内存，零网络开销

// 秒杀请求进来时，先检查本地标记
if v, ok := soldOutMap.Load(activityID); ok && v.(bool) {
    return "已售罄"  // 直接返回，不访问 Redis
}
```

售罄后：Redis QPS 从 10 万降到 0。这是一个非常有效的优化。

---

## 四、完整流程（从用户点击到下单成功）

### 4.1 时序图

```
用户          前端           Nginx        seckill-service      Redis         Kafka        Consumer       MySQL
 │             │              │               │                 │              │              │             │
 │──点击抢购──→│              │               │                 │              │              │             │
 │             │──防抖300ms──→│               │                 │              │              │             │
 │             │              │──限流检查────→│                 │              │              │             │
 │             │              │               │                 │              │              │             │
 │             │              │               │──1.JWT鉴权─────→│              │              │             │
 │             │              │               │                 │              │              │             │
 │             │              │               │──2.查活动缓存──→│              │              │             │
 │             │              │               │←─活动信息───────│              │              │             │
 │             │              │               │  (校验时间窗口)  │              │              │             │
 │             │              │               │                 │              │              │             │
 │             │              │               │──3.Lua原子扣库存→│              │              │             │
 │             │              │               │←─返回1(成功)────│              │              │             │
 │             │              │               │                 │              │              │             │
 │             │              │               │──4.写token──────→│              │              │             │
 │             │              │               │                 │              │              │             │
 │             │              │               │──5.投递消息─────→│──────────────→│             │             │
 │             │              │               │                 │              │              │             │
 │             │              │←──返回token───│                 │              │              │             │
 │             │←─"排队中"────│               │                 │              │              │             │
 │←─显示排队──│              │               │                 │              │              │             │
 │             │              │               │                 │              │              │             │
 │             │              │               │                 │              │──消费消息───→│             │
 │             │              │               │                 │              │              │──事务下单──→│
 │             │              │               │                 │              │              │←─成功───────│
 │             │              │               │                 │←─写结果──────│              │             │
 │             │              │               │                 │              │              │             │
 │──轮询结果──→│──────────────→│──────────────→│──读结果────────→│              │              │             │
 │             │              │               │←─{order_id}────│              │              │             │
 │←─下单成功──│              │               │                 │              │              │             │
 │             │              │               │                 │              │              │             │
 │──去支付────→│  (复用现有 payment-service 支付流程)           │              │              │             │
```

### 4.2 每一步详解

**步骤 1：本地售罄拦截**
- 检查进程内存中的 `soldOutMap`
- 如果已售罄，直接返回，零网络开销
- 这一步拦截了售罄后 99% 的请求

**步骤 2：活动校验**
- 从 Redis 缓存读取活动信息（命中率接近 100%，因为有预热）
- 校验：活动是否存在、是否在时间窗口内
- 缓存未命中时回源 DB，并写回缓存

**步骤 3：Redis Lua 原子扣库存**
- 一次 RTT 完成：限购检查 + 库存检查 + 库存扣减 + 限购计数
- 返回 1=成功，0=库存不足，-1=超出限购
- 库存不足时设置本地售罄标记

**步骤 4：生成排队令牌**
- UUID 生成唯一 token
- 写入 Redis，TTL=5 分钟
- token 是用户查询结果的凭证，也是 Consumer 的校验依据

**步骤 5：投递 Kafka**
- 按 activity_id 做 partition key（同一活动的消息有序）
- 投递失败时回滚 Redis 库存 + 删除 token，返回用户失败
- 投递成功后立即返回用户"排队中"

**步骤 6：Kafka Consumer 异步下单**
- 幂等校验（Redis SETNX）
- Token 有效性校验（只读不删，事务成功后才删）
- MySQL 事务：插入订单 + 扣减库存 + 插入秒杀关联记录
- 成功后写结果到 Redis，用户轮询即可拿到 order_id

---

## 五、防超卖：三层防线

这是面试最常问的问题。我们的方案是三层独立防线，任何一层都能独立防超卖。

### 第一层：Redis Lua 原子扣减

```lua
local stock = tonumber(redis.call('GET', KEYS[1]))
if stock == nil or stock < tonumber(ARGV[1]) then
    return 0  -- 库存不足，拒绝
end
redis.call('DECRBY', KEYS[1], ARGV[1])
return 1
```

- Redis 单线程，Lua 原子执行，不存在并发问题
- 挡住 99% 的请求，只有库存数量的请求能通过
- 这是性能最高的一层（微秒级）

### 第二层：Kafka Consumer 幂等校验

- Redis SETNX 幂等锁：防止同一消息被重复消费
- MySQL 唯一索引 `(activity_id, user_id)`：即使 Redis 失效，数据库也能拦截

### 第三层：MySQL 乐观锁（最终防线）

```sql
UPDATE product SET product_num = product_num - 1
WHERE product_id = ? AND product_num >= 1
```

- `WHERE product_num >= 1` 是关键：如果库存已经是 0，这条 SQL 影响行数为 0
- 影响行数 = 0 → 事务回滚 → 不会超卖
- 这是数据库级别的最终保障，即使 Redis 数据丢了也不会超卖

### 面试话术

> "我们的防超卖是三层防线：第一层 Redis Lua 原子扣减，挡住 99% 的请求；
> 第二层 Kafka Consumer 的幂等校验，防止消息重复消费；
> 第三层 MySQL 的乐观锁 WHERE product_num >= num，作为最终兜底。
> 三层是独立的，任何一层单独都能防超卖。"

---

## 六、防重复下单：三层幂等

### 为什么会重复？

1. 用户快速双击"抢购"按钮
2. Kafka Consumer rebalance 导致消息重新投递
3. 网络抖动导致 Producer 重试

### 三层幂等保障

| 层级 | 机制 | 位置 | 说明 |
|------|------|------|------|
| 第一层 | Redis 限购计数 | seckill-service | Lua 脚本中 `bought + 1 > limit` 直接拒绝 |
| 第二层 | Redis SETNX 幂等锁 | Kafka Consumer | `seckill:idempotent:{token}` 24h TTL |
| 第三层 | MySQL 唯一索引 | seckill_order 表 | `UNIQUE KEY (activity_id, user_id)` |

即使 Redis 全挂了，MySQL 唯一索引也能保证同一用户同一活动只能下一单。

---

## 七、Kafka 在秒杀中的角色

### 7.1 削峰填谷

```
没有 Kafka：
  10000 req/s ──→ MySQL ──→ 💥 崩溃（上限 ~2000 QPS）

有 Kafka：
  10000 req/s ──→ Redis 过滤 ──→ 100 msg ──→ Kafka ──→ Consumer 500 msg/s ──→ MySQL ✅
```

### 7.2 异步解耦

- seckill-service 只负责"快速响应用户"（Redis + Kafka 投递）
- Consumer 负责"慢慢写库"（MySQL 事务）
- 两者通过 Kafka 解耦，互不影响

### 7.3 消息不丢保障

| 环节 | 保障机制 |
|------|---------|
| Producer → Kafka | `RequiredAcks: kafka.RequireOne`（leader 写入即确认）+ 投递失败回滚 Redis |
| Kafka Broker | `replication.factor=3`（3 副本）|
| Kafka → Consumer | 手动提交 offset，处理成功才提交 |
| Consumer 失败 | 框架层原地重试 3 次，退避 1s/2s |
| 重试耗尽 | `OnExhausted` 回滚 Redis 库存 + 通知用户失败 |

### 7.4 Consumer 重试设计（这是我们踩过的坑）

最初的设计：事务失败 → 回滚 Redis 库存 → 返回 error 触发重试。

问题：第一次失败回滚了 Redis 库存，第二次重试成功了，但 Redis 库存已经多加了 1。
导致 Redis 库存比实际多，更多请求通过 Redis 进入 Kafka，虽然 MySQL 能兜底，但浪费资源。

修复后的设计：
```
事务失败（可重试）→ 只删幂等锁，不回滚 Redis → 返回 error → 框架重试
重试耗尽（不可恢复）→ OnExhausted 回滚 Redis + 通知用户
Token 过期（不可恢复）→ 直接回滚 Redis + 通知用户 → 返回 nil 不重试
```

这样 Redis 库存在整个重试周期内保持一致。

---

## 八、Redis Key 设计

| Key | 类型 | TTL | 用途 | 谁写 | 谁读 |
|-----|------|-----|------|------|------|
| `seckill:activity:{id}` | String(JSON) | 活动结束+1h | 活动信息缓存 | 预热/回源 | seckill-service |
| `seckill:stock:{activity_id}` | String(int) | 活动结束+1h | 秒杀库存 | 预热 | Lua 脚本 |
| `seckill:bought:{activity_id}:{user_id}` | String(int) | 24h | 用户已购数量 | Lua 脚本 | Lua 脚本 |
| `seckill:token:{token}` | String(JSON) | 5min | 排队令牌 | seckill-service | Consumer |
| `seckill:result:{token}` | String(JSON) | 30min | 下单结果 | Consumer | 前端轮询 |
| `seckill:idempotent:{token}` | String | 24h | 幂等去重 | Consumer | Consumer |

---

## 九、数据库设计

### 9.1 新增表

```sql
-- 秒杀活动表
CREATE TABLE `seckill_activity` (
  `id`              bigint       NOT NULL AUTO_INCREMENT,
  `title`           varchar(128) NOT NULL COMMENT '活动标题',
  `product_id`      int          NOT NULL COMMENT '关联商品ID',
  `seckill_price`   bigint       NOT NULL COMMENT '秒杀价（分）',
  `original_price`  bigint       NOT NULL COMMENT '原价（分）',
  `total_stock`     int          NOT NULL COMMENT '总库存',
  `available_stock` int          NOT NULL COMMENT '剩余库存',
  `limit_per_user`  int          NOT NULL DEFAULT 1 COMMENT '每人限购',
  `start_time`      bigint       NOT NULL COMMENT '开始时间（unix秒）',
  `end_time`        bigint       NOT NULL COMMENT '结束时间（unix秒）',
  `status`          tinyint      NOT NULL DEFAULT 0 COMMENT '0未开始 1进行中 2已结束 3已取消',
  `created_at`      bigint       NOT NULL,
  `updated_at`      bigint       NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_product_id` (`product_id`),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀活动表';

-- 秒杀订单关联表
CREATE TABLE `seckill_order` (
  `id`            bigint NOT NULL AUTO_INCREMENT,
  `activity_id`   bigint NOT NULL,
  `order_id`      bigint NOT NULL COMMENT '关联 orders.order_id',
  `user_id`       bigint NOT NULL,
  `product_id`    int    NOT NULL,
  `seckill_price` bigint NOT NULL COMMENT '成交价（分）',
  `num`           int    NOT NULL DEFAULT 1,
  `created_at`    bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_activity_user` (`activity_id`, `user_id`),  -- 防重复下单
  INDEX `idx_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀订单表';
```

### 9.2 为什么秒杀订单要单独建表？

秒杀订单最终还是写入现有的 `orders` 表（复用支付流程），`seckill_order` 只是关联表。

原因：
1. `UNIQUE KEY (activity_id, user_id)` — 数据库级别防重复下单
2. 方便查询"某个活动的所有秒杀订单"，不用在海量普通订单中过滤
3. 秒杀数据和普通订单数据隔离，互不影响

---

## 十、核心代码实现

以下是项目中的真实代码，不是伪代码。

### 10.1 秒杀入口（seckilllogic.go）

```go
func (l *SeckillLogic) Seckill(req *types.SeckillReq) (*types.SeckillResp, error) {
    userID, _ := ctxutil.UserIDFromCtx(l.ctx)

    // 1. 本地售罄快速拦截（零网络开销）
    if v, ok := soldOutMap.Load(req.ActivityID); ok && v.(bool) {
        return &types.SeckillResp{Code: "SOLD_OUT", Msg: "已售罄"}, nil
    }

    // 2. 活动校验（Redis 缓存，miss 时回源 DB）
    activity, err := l.getActivity(req.ActivityID)
    // ... 校验时间窗口 ...

    // 3. Redis Lua 原子操作：扣库存 + 限购校验
    result, _ := l.svcCtx.Cache.Eval(l.ctx, luaSeckillDecrStock,
        []string{stockKey, boughtKey}, 1, activity.LimitPerUser)

    ret, _ := result.(int64)
    switch ret {
    case -1: return "超出限购"
    case 0:  soldOutMap.Store(activityID, true); return "已售罄"
    }

    // 4. 生成 UUID 令牌，写入 Redis（TTL=5min）
    token := uuid.New().String()
    l.svcCtx.Cache.Set(l.ctx, tokenKey, tokenData, 5*time.Minute)

    // 5. 投递 Kafka（失败则回滚 Redis）
    l.svcCtx.KafkaProducer.Send(l.ctx, activityID, tokenData)

    // 6. 返回令牌
    return &types.SeckillResp{Code: "200", Token: token, Msg: "排队中"}
}
```

### 10.2 Kafka Consumer（seckillorderconsumer.go）

```go
func (c *SeckillOrderConsumer) Consume(ctx context.Context, key, value string) error {
    var msg types.SeckillMessage
    json.Unmarshal([]byte(value), &msg)

    // 1. 幂等校验（Redis SETNX，24h TTL）
    if err := c.cache.SetNX(ctx, idempotentKey, "1", 24*time.Hour); err != nil {
        return nil // 重复消息，跳过
    }

    // 2. Token 校验（只读不删！事务成功后才删）
    if err := c.cache.Get(ctx, tokenKey, &tokenData); err != nil {
        c.rollbackRedisStock(ctx, msg) // token 过期，回滚库存
        c.writeResult(ctx, msg.Token, 2, 0, "令牌已过期")
        return nil // 不可恢复，不重试
    }

    // 3. MySQL 事务
    txErr := c.seckillOrderModel.TransactCtx(ctx, func(ctx, session) error {
        txOrders.Insert(ctx, &model.Orders{...})           // 插入订单
        txProduct.DecrStock(ctx, productID, num)            // 扣库存（WHERE >= num）
        txSeckillOrder.Insert(ctx, &model.SeckillOrder{...}) // 秒杀关联
        return nil
    })

    if txErr != nil {
        c.cache.Del(ctx, idempotentKey) // 删幂等锁，允许重试
        // 不回滚 Redis！框架会重试，重试可能成功
        return txErr // 触发框架重试（最多 3 次）
    }

    // 4. 成功：删 token + 写结果
    c.cache.Del(ctx, tokenKey)
    c.writeResult(ctx, msg.Token, 1, orderID, "下单成功")
    return nil
}

// 框架重试 3 次都失败后调用
func (c *SeckillOrderConsumer) OnExhausted(ctx, key, value string) {
    c.rollbackRedisStock(ctx, msg)  // 这时才回滚 Redis
    c.writeResult(ctx, msg.Token, 2, 0, "下单失败")
}
```

### 10.3 Kafka Consumer 框架（kafka/consumer.go）

```go
func (c *Consumer) Start(ctx context.Context, handler ConsumeHandler, onExhausted ExhaustedHandler) {
    for {
        msg, _ := c.reader.FetchMessage(ctx)

        var handleErr error
        for attempt := 1; attempt <= 3; attempt++ {
            handleErr = handler(ctx, msg.Key, msg.Value)
            if handleErr == nil { break }
            time.Sleep(time.Duration(attempt) * time.Second) // 退避 1s, 2s
        }

        if handleErr != nil && onExhausted != nil {
            onExhausted(ctx, msg.Key, msg.Value) // 重试耗尽，清理
        }

        c.reader.CommitMessages(ctx, msg) // 无论成功失败都提交 offset
    }
}
```

### 10.4 预热逻辑（warmuplogic.go）

服务启动时自动执行，将活动数据加载到 Redis：

```go
func WarmUp(ctx context.Context, svcCtx *svc.ServiceContext) {
    // 查询即将开始（1h内）和进行中的活动
    upcoming := svcCtx.SeckillActivityModel.FindUpcoming(ctx, now)
    ongoing := svcCtx.SeckillActivityModel.FindOngoing(ctx, now)

    for _, activity := range activities {
        // 缓存活动信息
        svcCtx.Cache.Set(ctx, activityKey, activity, ttl)
        // 缓存库存（SetNX 避免覆盖已有的实时库存）
        svcCtx.Cache.SetNX(ctx, stockKey, activity.AvailableStock, ttl)
    }
}
```

为什么库存用 `SetNX` 而不是 `Set`？
- 如果服务重启，Redis 中可能已有被扣减过的实时库存（比如从 100 扣到 80）
- 用 `Set` 会把 80 覆盖回 100 → 超卖
- 用 `SetNX` 只在 key 不存在时才写入，不会覆盖已有数据

---

## 十一、高并发优化清单

### 11.1 各层优化

| 层级 | 优化点 | 做法 | 效果 |
|------|--------|------|------|
| 前端 | CDN 静态化 | 秒杀页面 HTML/JS/CSS/图片走 CDN | 源站零带宽压力 |
| 前端 | 倒计时 | 服务端下发开始时间，前端本地倒计时 | 避免用户疯狂刷新 |
| 前端 | 按钮防抖 | 点击后 300ms 不可再点 | 减少 30% 重复请求 |
| Nginx | 令牌桶限流 | `limit_req rate=10000r/s burst=20000` | 超出直接 503 |
| Nginx | IP 限频 | 单 IP 每秒最多 10 次 | 防脚本刷接口 |
| 服务 | 本地售罄标记 | `sync.Map` 进程内存 | 售罄后零网络开销 |
| 服务 | 超时控制 | 接口总超时 500ms | 超时快速失败 |
| 服务 | 预热 | 启动时加载活动到 Redis | 避免冷启动穿透 |
| Redis | Lua 合并操作 | 限购+库存一个脚本 | 1 次 RTT 替代 3 次 |
| Redis | 连接池 | 100+ 连接 | 避免连接瓶颈 |
| Kafka | Hash 分区 | 按 activity_id 分区 | 同活动消息有序 |
| Kafka | 批量+压缩 | batch=100, lz4 压缩 | 提高吞吐 |
| MySQL | 乐观锁 | `WHERE product_num >= num` | 无需悲观锁，性能好 |
| MySQL | 独立连接池 | 秒杀服务独立连接池 | 不占主站连接 |

### 11.2 Nginx 限流配置示例

```nginx
http {
    limit_req_zone $binary_remote_addr zone=seckill:10m rate=10000r/s;
    limit_req_zone $binary_remote_addr zone=seckill_ip:10m rate=10r/s;

    server {
        location /seckill/buy {
            limit_req zone=seckill burst=20000 nodelay;
            limit_req zone=seckill_ip burst=20 nodelay;
            proxy_pass http://seckill_upstream;
        }
    }
}
```

---

## 十二、异常场景分析

面试官最喜欢问"如果 XXX 挂了怎么办"。

### 12.1 Redis 挂了

- Lua 脚本执行失败 → 返回用户"系统繁忙，请重试"
- 不会超卖（MySQL 乐观锁兜底）
- 不会丢单（请求没进入 Kafka，用户可以重试）

### 12.2 Kafka 挂了

- Producer 投递失败 → 回滚 Redis 库存 → 返回用户"系统繁忙"
- 用户可以重试，Redis 库存已恢复
- 可以降级为同步下单（直接走 MySQL）

### 12.3 MySQL 挂了

- Consumer 事务失败 → 框架重试 3 次
- 3 次都失败 → OnExhausted 回滚 Redis 库存 → 通知用户失败
- 消息不丢（offset 已提交，但 Redis 库存已恢复，用户可重新抢购）

### 12.4 Consumer 进程崩溃

- Kafka 消息不丢（offset 未提交的消息会被重新投递）
- 幂等锁保证不会重复下单
- 进程恢复后继续消费

### 12.5 用户重复点击

- 前端：按钮防抖 300ms
- Redis：限购计数（Lua 原子检查）
- Consumer：幂等锁（SETNX）
- MySQL：唯一索引 `(activity_id, user_id)`

四层防护，任何一层都能拦截。

---

## 十三、面试高频问题

### Q1：秒杀系统怎么防超卖？

三层防线：
1. Redis Lua 原子扣减（挡 99% 请求）
2. Kafka Consumer 幂等校验（防重复消费）
3. MySQL `WHERE product_num >= num`（最终兜底）

### Q2：为什么用 Kafka 不直接写 MySQL？

削峰填谷。秒杀瞬间万级请求，MySQL 扛不住。Kafka 把尖峰流量拉平成 MySQL 能承受的平滑流量。同时实现异步解耦，用户不用等 MySQL 写完。

### Q3：Redis 和 MySQL 库存不一致怎么办？

Redis 库存是"乐观预扣"，可能比 MySQL 多扣（比如 Consumer 失败回滚了 Redis 但 MySQL 没扣）。
- 短期：以 MySQL 为准，Redis 只是快速过滤器
- 长期：定时对账任务，Redis 库存 vs MySQL 库存，不一致时以 MySQL 为准修正 Redis

### Q4：Kafka 消息丢了怎么办？

三层保障：
- Producer: `acks=1` + 投递失败回滚 Redis（用户可重试）
- Broker: 3 副本
- Consumer: 手动提交 offset，成功才提交

### Q5：如何保证不重复下单？

三层幂等：
- Redis 限购计数（Lua 原子）
- Redis SETNX 幂等锁（Consumer 层）
- MySQL 唯一索引 `(activity_id, user_id)`

### Q6：秒杀把主站打崩了怎么办？

独立部署 + 故障隔离：
- seckill-service 独立进程/容器
- 独立 Redis 实例（或独立 DB 编号）
- 独立 MySQL 连接池
- Nginx 按路径分流

### Q7：售罄后还有大量请求怎么办？

本地售罄标记（`sync.Map`）：
- Redis 返回库存不足时，设置进程内存标记
- 后续请求直接在内存中拦截，不访问 Redis
- Redis QPS 从 10 万降到 0

### Q8：订单超时未支付怎么处理？

- 下单时设置 Redis key `seckill:order:expire:{order_id}`，TTL=30min
- 定时任务扫描过期订单 → 关闭订单 → 回滚 MySQL 库存 + Redis 库存
- 或者用 Kafka 延迟消息实现

### Q9：你们的 Lua 脚本为什么要把限购和库存放在一起？

减少 RTT。如果分开：
- 第一次 RTT：检查限购
- 第二次 RTT：扣减库存
- 两次之间不是原子的，可能出现"限购检查通过但库存已被别人扣完"的情况

合并后一次 RTT 完成所有操作，既快又安全。

### Q10：Consumer 事务失败时为什么不立即回滚 Redis？

因为框架会立即重试。如果第一次失败就回滚 Redis，第二次重试成功了，Redis 库存就比实际多 1。
我们的做法是：重试期间不回滚，只在确定放弃时（3 次都失败）才回滚。

---

## 十四、项目结构

```
backend/
├── kafka/                              # Kafka 封装
│   ├── producer.go                     # Producer（JSON 序列化 + Hash 分区 + Lz4 压缩）
│   └── consumer.go                     # Consumer（重试 3 次 + 退避 + OnExhausted 回调）
├── model/
│   ├── seckillactivitymodel.go         # 活动 Model（DecrStock/IncrStock/状态更新）
│   ├── seckillactivitymodel_gen.go     # 活动 Model（goctl 生成）
│   ├── seckillordermodel.go            # 秒杀订单 Model（事务 + 按活动+用户查询）
│   ├── seckillordermodel_gen.go        # 秒杀订单 Model（goctl 生成）
│   └── sql/seckill.sql                 # 建表 SQL
├── api/seckill.api                     # API 定义
└── service/seckill/                    # 秒杀服务
    ├── seckill.go                      # 入口（HTTP + 预热 + Consumer 启动 + 优雅关闭）
    ├── etc/seckill-api.yaml            # 配置（DB + Redis + Kafka）
    └── internal/
        ├── config/config.go            # 配置结构体
        ├── handler/                    # HTTP handlers（4 个接口）
        ├── logic/
        │   ├── seckilllogic.go         # 核心：本地拦截→Redis Lua→Kafka 投递
        │   ├── seckillresultlogic.go   # 前端轮询结果
        │   ├── getseckillactivitylogic.go  # 活动详情
        │   ├── listseckillactivitieslogic.go # 活动列表
        │   └── warmuplogic.go          # 预热（启动时加载活动到 Redis）
        ├── consumer/
        │   └── seckillorderconsumer.go # Kafka 消费者（幂等+事务+重试+回滚）
        ├── middleware/authmiddleware.go # JWT 鉴权
        ├── svc/servicecontext.go       # 依赖注入
        └── types/types.go             # 请求/响应/消息结构体
```

---

## 十五、一句话总结

秒杀系统的核心就是四个字：**分层拦截**。

前端拦无效点击，Nginx 拦超量请求，Redis 拦库存不足，Kafka 削峰填谷，MySQL 兜底防超卖。每一层都在减少下一层的压力，最终只有真正需要下单的请求到达数据库。
