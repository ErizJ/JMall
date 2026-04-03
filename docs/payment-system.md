# JMall 支付系统：原理与实现

> 本文档基于 JMall 项目真实代码，从支付架构到安全防护完整讲解。
> 技术栈：Go (go-zero) + MySQL + Redis + Strategy 模式
> 目标：读完这篇文档，面试中能把支付系统从流程到安全讲清楚。

---

## 一、支付系统的本质问题

支付系统要解决的核心问题：**把用户的钱安全、准确地从 A 转到 B，并且不能多转、不能少转、不能转丢**。

听起来简单，但工程上极其复杂：
- 网络不可靠：请求可能超时、重复、乱序
- 第三方不可控：微信/支付宝的回调可能延迟、重复、伪造
- 并发冲突：同一订单可能被同时发起两次支付
- 金额敏感：差一分钱都是事故

所以支付系统的设计哲学：**宁可多校验一次，不可少校验一次。宁可慢一点，不可错一点。**

### 1.1 我们的支付系统支持什么

| 功能 | 接口 | 说明 |
|------|------|------|
| 创建支付单 | POST /payment/create | 发起支付，返回支付链接 |
| 查询支付状态 | POST /payment/status | 前端轮询支付结果 |
| 支付回调 | POST /payment/notify | 第三方渠道通知支付结果 |
| 模拟支付 | POST /payment/mock/pay | 开发环境模拟支付成功 |
| 申请退款 | POST /payment/refund | 退款 + 库存回滚 |
| 查询支付记录 | POST /payment/list | 用户的支付历史 |

### 1.2 支付渠道

通过 Strategy 模式支持多渠道，目前实现了 Mock 渠道，预留了微信/支付宝接口：

| 渠道 | 状态 | 说明 |
|------|------|------|
| mock | 已实现 | 开发测试用，模拟支付行为 |
| wechat | 预留 | 微信支付（扫码/JSAPI/APP） |
| alipay | 预留 | 支付宝（网页/手机/APP） |

---

## 二、整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        前端                                  │
│  下单 → 创建支付单 → 跳转支付页 → 轮询结果 → 支付成功        │
└──────────────┬──────────────────────────┬───────────────────┘
               │                          │
    POST /payment/create          POST /payment/status
               │                          │
┌──────────────▼──────────────────────────▼───────────────────┐
│                   payment-service                            │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              CreatePaymentLogic                       │   │
│  │  1. JWT 鉴权                                         │   │
│  │  2. 订单校验（存在性 + 归属 + 状态）                  │   │
│  │  3. Redis SETNX 防重复支付                            │   │
│  │  4. 计算金额（元→分）                                 │   │
│  │  5. 生成支付流水号                                    │   │
│  │  6. 调用支付渠道（Strategy 模式）                     │   │
│  │  7. 写入 payment_order 表                             │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              PaymentNotifyLogic                       │   │
│  │  1. Redis 幂等锁                                     │   │
│  │  2. 查支付单                                         │   │
│  │  3. 终态检查                                         │   │
│  │  4. 渠道验签                                         │   │
│  │  5. 金额校验                                         │   │
│  │  6. 过期检查                                         │   │
│  │  7. 事务：更新支付单 + 订单状态                       │   │
│  │  8. 清理锁和缓存                                     │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌────────────────┐                                         │
│  │ Channel Layer  │  PayChannel 接口                        │
│  │ ├── mock       │  ├── CreatePayment()                    │
│  │ ├── wechat     │  ├── QueryPayment()                     │
│  │ └── alipay     │  ├── Refund()                           │
│  └────────────────┘  └── VerifyNotify()                     │
│                                                              │
└──────────────────────────────────────────────────────────────┘
               │                          ▲
               ▼                          │
┌──────────────────────────┐   ┌──────────────────────────┐
│  第三方支付渠道            │   │  第三方回调               │
│  微信/支付宝/Mock          │   │  POST /payment/notify    │
└──────────────────────────┘   └──────────────────────────┘
```

---

## 三、支付状态机

支付单有 6 个状态，只能单向流转：

```
                    ┌──────────┐
                    │  0 待支付 │
                    └────┬─────┘
                         │
              ┌──────────┼──────────┐
              ▼          ▼          ▼
        ┌──────────┐ ┌──────────┐ ┌──────────┐
        │ 1 支付中  │ │ 4 已关闭 │ │ 3 支付失败│
        └────┬─────┘ └──────────┘ └──────────┘
             │           (超时)       (渠道失败)
             ▼
        ┌──────────┐
        │ 2 支付成功│
        └────┬─────┘
             │
             ▼
        ┌──────────┐
        │ 5 已退款  │
        └──────────┘
```

```go
const (
    PaymentStatusPending = 0 // 待支付
    PaymentStatusPaying  = 1 // 支付中
    PaymentStatusSuccess = 2 // 支付成功
    PaymentStatusFailed  = 3 // 支付失败
    PaymentStatusClosed  = 4 // 已关闭（超时）
    PaymentStatusRefund  = 5 // 已退款
)
```

**状态机的关键约束：**
- 只有待支付(0)/支付中(1)的单才能变成支付成功(2)
- 只有支付成功(2)的单才能退款(5)
- 已成功/已退款的单不会被重复更新（终态不可变）

这个约束在 SQL 层面强制执行：
```sql
UPDATE payment_order SET status = 2, channel_trade_no = ?, paid_time = ?
WHERE payment_no = ? AND status IN (0, 1)  -- 只有待支付/支付中才能更新
```

---

## 四、完整支付流程

### 4.1 创建支付单

```
用户点击"去支付"
    │
    ▼
POST /payment/create {order_id: 123, channel: "mock"}
    │
    ▼
CreatePaymentLogic:
    │
    ├── 1. JWT 鉴权 → 提取 userID
    │
    ├── 2. 查订单 → FindByOrderId(123)
    │      → 订单不存在？返回 003
    │      → 不是本人订单？返回 004
    │      → 订单状态不是待支付(0)？返回 011
    │
    ├── 3. 防重复支付 → Redis SETNX("jmall:payment:lock:123", TTL=30min)
    │      → key 已存在？说明已有进行中的支付单 → 返回 005
    │      → SETNX 成功 → 继续
    │
    ├── 4. 计算金额 → 遍历订单商品，sum(price × num) → 转为分
    │      → 1599.00 元 × 1 + 799.00 元 × 1 = 2398.00 元 = 239800 分
    │
    ├── 5. 生成支付流水号 → "PAY" + 时间戳毫秒 + 6位随机数
    │      → "PAY1712345678901123456"
    │
    ├── 6. 调用支付渠道 → channel.Get("mock").CreatePayment(...)
    │      → 返回 {payUrl: "/payment/mock/pay?payment_no=PAY...", channelTradeNo: "MOCK_..."}
    │      → 失败？释放 Redis 锁，返回 006
    │
    ├── 7. 写入 payment_order 表
    │      → {payment_no, order_id, user_id, amount=239800, channel="mock", status=0, expire_time=now+30min}
    │
    └── 返回 {code: "200", payment_no: "PAY...", pay_url: "/payment/mock/pay?..."}
```

### 4.2 支付回调（核心中的核心）

```
第三方渠道通知支付结果
    │
    ▼
POST /payment/notify {payment_no, channel_trade_no, status, amount, paid_time, sign}
    │
    ▼
PaymentNotifyLogic（8 步安全校验）:
    │
    ├── 1. Redis 幂等锁 → SETNX("jmall:payment:notify:PAY...", TTL=24h)
    │      → key 已存在？重复回调，直接返回 200（告诉渠道"我收到了"）
    │
    ├── 2. 查支付单 → FindByPaymentNo("PAY...")
    │      → 不存在？删幂等锁，返回 404
    │
    ├── 3. 终态检查 → 支付单已成功/已退款？
    │      → 是 → 直接返回 200（幂等，不重复处理）
    │
    ├── 4. 渠道验签 → channel.VerifyNotify(params)
    │      → 签名无效？删幂等锁，返回 401
    │      → 防止伪造回调攻击
    │
    ├── 5. 金额校验 → 回调金额 == 支付单金额？
    │      → 不一致？删幂等锁，返回 015
    │      → 防止金额篡改攻击
    │
    ├── 6. 过期检查 → now > expire_time？
    │      → 已过期 → 事务：关闭支付单 + 取消订单 + 回滚库存
    │
    ├── 7. 事务更新（支付成功时）:
    │      ├── UPDATE payment_order SET status=2 WHERE payment_no=? AND status IN (0,1)
    │      └── UPDATE orders SET status=1 WHERE order_id=?
    │      → 事务失败？删幂等锁，允许渠道重试
    │
    │   事务更新（支付失败时）:
    │      ├── UPDATE payment_order SET status=3
    │      ├── UPDATE orders SET status=2（取消）
    │      └── 回滚库存：product_num += num
    │
    └── 8. 清理 → 删除支付锁 + 清理用户缓存
```

### 4.3 为什么回调处理这么复杂？

因为第三方回调是**不可信的外部输入**：
- 可能被伪造（黑客构造假回调）→ 需要验签
- 可能金额被篡改（支付 1 分钱声称支付了 1000 元）→ 需要金额校验
- 可能重复发送（网络重试）→ 需要幂等
- 可能延迟到达（支付单已过期）→ 需要过期检查

每一步校验都是一道防线，缺一不可。

---

## 五、Strategy 模式：支付渠道抽象

### 5.1 为什么用 Strategy 模式？

不同支付渠道（微信/支付宝/Mock）的 API 完全不同，但业务逻辑是一样的：创建支付 → 等回调 → 更新状态。

Strategy 模式把"变化的部分"（渠道差异）封装到接口后面，业务逻辑层不感知具体渠道：

```go
// channel.go — 统一接口
type PayChannel interface {
    Name() string
    CreatePayment(ctx context.Context, req *PayRequest) (*PayResponse, error)
    QueryPayment(ctx context.Context, paymentNo string) (bool, string, error)
    Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error)
    VerifyNotify(ctx context.Context, params map[string]string) (bool, error)
}
```

### 5.2 渠道注册中心

```go
// 全局注册表
var registry = make(map[string]PayChannel)

func Register(ch PayChannel) {
    registry[ch.Name()] = ch
}

func Get(name string) (PayChannel, error) {
    ch, ok := registry[name]
    if !ok {
        return nil, fmt.Errorf("unsupported channel: %s", name)
    }
    return ch, nil
}
```

Mock 渠道在 `init()` 中自动注册：
```go
// mock.go
func init() {
    Register(NewMockChannel())
}
```

### 5.3 新增渠道只需 3 步

1. 实现 `PayChannel` 接口
2. 在 `init()` 中调用 `Register()`
3. 前端传 `channel: "wechat"` 即可

业务逻辑层（CreatePaymentLogic、PaymentNotifyLogic）完全不用改。这就是 Strategy 模式的价值。

### 5.4 面试话术

> "我们的支付系统用 Strategy 模式抽象了支付渠道。定义了统一的 PayChannel 接口，
> 包含创建支付、查询、退款、验签四个方法。每个渠道（Mock/微信/支付宝）实现这个接口，
> 通过注册中心按名称路由。业务逻辑层完全不感知渠道差异，新增渠道只需实现接口并注册。"

---

## 六、防重复支付

### 6.1 为什么会重复支付？

- 用户快速双击"支付"按钮
- 网络超时后用户重试
- 前端 bug 发了两次请求

如果不防重，同一订单会创建两个支付单，用户可能被扣两次钱。

### 6.2 两层防护

**第一层：Redis SETNX（快速拦截）**

```go
lockKey := fmt.Sprintf("jmall:payment:lock:%d", req.OrderID)
lockErr := l.svcCtx.Cache.SetNX(l.ctx, lockKey, "1", 30*time.Minute)
if lockErr != nil {
    return "重复支付"  // key 已存在，说明已有进行中的支付单
}
```

- TTL = 支付过期时间（30 分钟）
- 支付成功/失败/过期后删除 key，允许重新发起

**第二层：订单状态校验（业务兜底）**

```go
if orderItems[0].Status != 0 {
    return "订单状态不允许支付"  // 只有待支付(0)的订单才能发起支付
}
```

Redis 锁在支付完成后会被清理，如果清理后用户又点了支付，Redis 锁拦不住。
但订单状态已经变成"已支付(1)"，业务层校验能拦住。

两层独立防护，任何一层都能防重复。

---

## 七、回调幂等：三层保障

### 7.1 为什么回调会重复？

第三方渠道的回调机制是"至少一次"（at-least-once）：
- 你返回非 200 → 渠道会重试（微信最多 15 次，间隔递增）
- 网络抖动导致你返回了 200 但渠道没收到 → 渠道重试
- 渠道系统 bug → 发了两次

所以回调处理必须是幂等的——处理 1 次和处理 100 次，结果一样。

### 7.2 三层幂等

| 层级 | 机制 | 说明 |
|------|------|------|
| 第一层 | Redis SETNX | `seckill:payment:notify:{payment_no}`，24h TTL。O(1) 快速拦截 |
| 第二层 | DB 状态机 | `UPDATE WHERE status IN (0,1)`，已成功的单不会被重复更新 |
| 第三层 | 终态检查 | 代码层面检查 `status == Success || status == Refund`，直接返回 |

```go
// 第一层：Redis
idempotentKey := fmt.Sprintf("jmall:payment:notify:%s", req.PaymentNo)
if err := cache.SetNX(idempotentKey, "1", 24*time.Hour); err != nil {
    return "200"  // 重复回调，直接返回成功
}

// 第三层：终态检查
if payment.Status == PaymentStatusSuccess || payment.Status == PaymentStatusRefund {
    return "200"  // 已处理，直接返回
}

// 第二层：DB 状态机
UPDATE payment_order SET status = 2
WHERE payment_no = ? AND status IN (0, 1)  -- 已成功的不会被更新
```

### 7.3 失败可重试

如果事务失败，删除 Redis 幂等锁，允许渠道下次回调时重试：

```go
if txErr != nil {
    _ = cache.Del(idempotentKey)  // 删锁，允许重试
    return "500"  // 告诉渠道"我处理失败了，请重试"
}
```

---

## 八、金额处理：为什么用分？

### 8.1 浮点数的坑

```go
// ❌ 错误：浮点数精度丢失
0.1 + 0.2 = 0.30000000000000004

// 用户支付 19.99 元，系统算出 19.990000000000002 元
// 和回调金额 1999 分对不上 → 金额校验失败 → 支付单卡住
```

### 8.2 正确做法：全程用分（整数）

```go
// 订单金额（元）→ 支付金额（分）
totalAmountFen := int64(math.Round(totalAmountYuan * 100))
// 1599.00 元 → 159900 分

// 数据库存储：bigint，单位分
// payment_order.amount = 159900

// 回调校验：整数比较，精确无误
if req.Amount != payment.Amount {
    return "金额不匹配"
}
```

**规则：数据库存分，传输用分，只在展示给用户时才转元。**

---

## 九、退款流程

```
用户申请退款
    │
    ▼
POST /payment/refund {payment_no, refund_amount, reason}
    │
    ▼
RefundLogic:
    │
    ├── 1. 查支付单 → 校验归属 + 状态必须是"支付成功(2)"
    │
    ├── 2. 退款幂等 → Redis SETNX("jmall:refund:lock:{payment_no}", 30s)
    │      → 防止用户快速双击退款按钮
    │
    ├── 3. 校验退款金额 → 0 < refund_amount ≤ payment.amount
    │
    ├── 4. 生成退款流水号 → "REF" + 时间戳 + 随机数
    │
    ├── 5. 调用渠道退款 → channel.Refund(...)
    │      → Mock 渠道：同步成功（Sync=true）
    │      → 真实渠道：异步处理（Sync=false，需等回调）
    │
    ├── 6. 事务：
    │      ├── 创建退款单（payment_refund 表）
    │      ├── 更新支付单状态 → 已退款(5)
    │      ├── 更新订单状态 → 已退款(3)
    │      └── 回滚库存 → product_num += num（逐个商品）
    │
    └── 7. 清理缓存（订单缓存 + 库存缓存）
```

退款的关键点：**库存回滚**。退款不只是退钱，还要把库存加回来，否则商品永远"卖出去"了。

---

## 十、Mock 支付：开发环境怎么测？

真实支付需要商户号、证书、公网回调地址，开发环境不具备。Mock 渠道解决这个问题。

### 10.1 Mock 支付流程

```
1. 创建支付单 → channel="mock"
   → MockChannel.CreatePayment() 返回 payUrl="/payment/mock/pay?payment_no=PAY..."

2. 前端跳转到 mock 支付页（或直接调接口）
   → POST /payment/mock/pay {payment_no: "PAY..."}

3. MockPayLogic 内部模拟支付成功
   → 幂等检查 → 过期检查 → 事务更新支付单+订单状态
   → 和真实回调处理逻辑完全一致
```

### 10.2 从 Mock 升级到真实支付

只需要：
1. 实现 `WechatChannel` / `AlipayChannel`（实现 `PayChannel` 接口）
2. 在 `init()` 中注册
3. 前端传 `channel: "wechat"` 或 `channel: "alipay"`

业务逻辑层零改动。代码中已经预留了详细的接入指南（`wechat.go` 和 `alipay.go`）。

---

## 十一、数据库设计

### 11.1 支付单表

```sql
CREATE TABLE `payment_order` (
  `id`               bigint       NOT NULL AUTO_INCREMENT,
  `payment_no`       varchar(64)  NOT NULL COMMENT '支付流水号（全局唯一）',
  `order_id`         bigint       NOT NULL COMMENT '关联业务订单ID',
  `user_id`          bigint       NOT NULL,
  `amount`           bigint       NOT NULL COMMENT '支付金额（分）',
  `channel`          varchar(32)  NOT NULL COMMENT 'mock/wechat/alipay',
  `channel_trade_no` varchar(128) NOT NULL DEFAULT '',
  `status`           tinyint      NOT NULL DEFAULT 0,
  `expire_time`      bigint       NOT NULL DEFAULT 0 COMMENT '过期时间（unix秒）',
  `paid_time`        bigint       NOT NULL DEFAULT 0,
  `notify_url`       varchar(256) NOT NULL DEFAULT '',
  `extra`            text         COMMENT '渠道扩展字段（JSON）',
  `created_at`       bigint       NOT NULL,
  `updated_at`       bigint       NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payment_no` (`payment_no`),
  INDEX `idx_order_id` (`order_id`),
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB COMMENT='支付单表';
```

设计要点：
- `payment_no` 全局唯一，对外交互用（给渠道、给前端）
- `amount` 用 bigint 存分，避免浮点精度问题
- `order_id` 和 `payment_no` 是 1:N 关系（一个订单可能多次发起支付）
- `expire_time` 支付过期时间，超时自动关闭

### 11.2 退款单表

```sql
CREATE TABLE `payment_refund` (
  `id`                bigint       NOT NULL AUTO_INCREMENT,
  `refund_no`         varchar(64)  NOT NULL COMMENT '退款流水号',
  `payment_no`        varchar(64)  NOT NULL COMMENT '关联支付流水号',
  `order_id`          bigint       NOT NULL,
  `user_id`           bigint       NOT NULL,
  `refund_amount`     bigint       NOT NULL COMMENT '退款金额（分）',
  `reason`            varchar(256) NOT NULL DEFAULT '',
  `channel`           varchar(32)  NOT NULL,
  `channel_refund_no` varchar(128) NOT NULL DEFAULT '',
  `status`            tinyint      NOT NULL DEFAULT 0 COMMENT '0退款中 1成功 2失败',
  `created_at`        bigint       NOT NULL,
  `updated_at`        bigint       NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_refund_no` (`refund_no`),
  INDEX `idx_payment_no` (`payment_no`)
) ENGINE=InnoDB COMMENT='退款单表';
```

---

## 十二、Redis Key 设计

| Key | TTL | 用途 | 谁写 | 谁读 |
|-----|-----|------|------|------|
| `jmall:payment:lock:{order_id}` | 30min | 防重复支付 | CreatePayment | CreatePayment |
| `jmall:payment:notify:{payment_no}` | 24h | 回调幂等 | PaymentNotify | PaymentNotify |
| `jmall:refund:lock:{payment_no}` | 30s | 退款防重 | Refund | Refund |

---

## 十三、安全防护总结

| 攻击类型 | 防护机制 | 代码位置 |
|---------|---------|---------|
| 伪造回调 | 渠道验签 `VerifyNotify()` | PaymentNotifyLogic 步骤 4 |
| 金额篡改 | 回调金额 vs 支付单金额比对 | PaymentNotifyLogic 步骤 5 |
| 重复支付 | Redis SETNX + 订单状态校验 | CreatePaymentLogic 步骤 3/3.5 |
| 重复回调 | Redis 幂等锁 + DB 状态机 + 终态检查 | PaymentNotifyLogic 步骤 1/2/3 |
| 重复退款 | Redis SETNX 30s 锁 | RefundLogic 步骤 3.5 |
| 越权操作 | JWT 鉴权 + 订单归属校验 | 所有接口步骤 1/2 |
| 过期支付 | expire_time 检查 + 事务关闭 | PaymentNotifyLogic 步骤 6 |
| 浮点精度 | 全程用分（整数） | 金额计算 + DB 存储 |

---

## 十四、面试高频问题

### Q1：支付回调怎么保证幂等？

三层保障：第一层 Redis SETNX 快速拦截重复回调；第二层 DB 状态机 `UPDATE WHERE status IN (0,1)` 保证已成功的单不会被重复更新；第三层代码中终态检查直接返回。三层独立，任何一层都能保证幂等。

### Q2：如果回调处理失败了怎么办？

删除 Redis 幂等锁，返回非 200 给渠道。渠道会按策略重试（微信最多 15 次，间隔递增）。下次回调时 Redis 锁已删除，可以重新处理。

### Q3：怎么防止伪造回调？

渠道验签。每个支付渠道的回调都带签名（sign），我们用渠道公钥验证签名。签名无效直接拒绝。Mock 渠道跳过验签，生产环境必须开启。

### Q4：为什么金额要用分而不是元？

浮点数有精度问题。`0.1 + 0.2 != 0.3`。如果用元存储，回调金额校验可能因为精度差异失败。用分（整数）完全避免这个问题。数据库存分，传输用分，只在展示时转元。

### Q5：一个订单可以多次发起支付吗？

可以。`order_id` 和 `payment_no` 是 1:N 关系。第一次支付超时/失败后，Redis 锁过期，用户可以重新发起支付，生成新的 payment_no。但同一时间只能有一个进行中的支付单（Redis SETNX 保证）。

### Q6：支付渠道怎么扩展？

Strategy 模式。定义了 `PayChannel` 接口（CreatePayment/QueryPayment/Refund/VerifyNotify），每个渠道实现这个接口并注册到全局 Registry。新增渠道只需实现接口 + 注册，业务逻辑零改动。

### Q7：支付超时怎么处理？

两个时机：1）回调到达时检查 `expire_time`，已过期则事务关闭支付单 + 取消订单 + 回滚库存；2）定时任务扫描过期未支付的单，主动关闭。

### Q8：退款时为什么要回滚库存？

退款意味着商品"退回来了"。如果不回滚库存，这些商品永远显示"已售出"，其他用户买不到。退款事务中逐个商品 `product_num += num`。

### Q9：Mock 支付和真实支付的区别？

Mock 支付是同步的：前端调 `/payment/mock/pay` → 后端直接更新状态。真实支付是异步的：前端跳转到微信/支付宝页面 → 用户付款 → 渠道回调通知 → 后端更新状态。但两者的状态更新逻辑完全一致（都走事务更新支付单+订单）。

### Q10：如果要做对账怎么做？

定时任务每天凌晨：1）从渠道下载对账单（微信/支付宝都提供对账文件下载接口）；2）和本地 payment_order 表逐笔比对；3）金额/状态不一致的标记为异常，人工处理。对账是支付系统的最后一道防线。

---

## 十五、项目结构

```
backend/service/payment/
├── payment.go                          # 入口
├── etc/payment-api.yaml                # 配置（DB + Redis + 支付过期时间 + 回调URL）
└── internal/
    ├── config/config.go                # 配置结构体（含 PaymentConfig）
    ├── channel/
    │   ├── channel.go                  # PayChannel 接口 + Registry（Strategy 模式核心）
    │   ├── mock.go                     # Mock 渠道（开发测试）
    │   ├── wechat.go                   # 微信支付（预留，含接入指南）
    │   └── alipay.go                   # 支付宝（预留，含接入指南）
    ├── handler/
    │   ├── routes.go                   # 6 个接口路由
    │   ├── createpaymenthandler.go     # 创建支付单
    │   ├── getpaymentstatushandler.go  # 查询状态
    │   ├── paymentnotifyhandler.go     # 支付回调（无需鉴权）
    │   ├── mockpayhandler.go           # 模拟支付（无需鉴权）
    │   ├── refundhandler.go            # 退款
    │   └── getuserpaymentshandler.go   # 支付记录
    ├── logic/
    │   ├── createpaymentlogic.go       # 创建支付（7步校验+落库）
    │   ├── paymentnotifylogic.go       # 回调处理（8步安全校验，核心）
    │   ├── mockpaylogic.go             # Mock 支付（复用回调逻辑）
    │   ├── refundlogic.go              # 退款（渠道退款+事务+库存回滚）
    │   ├── getpaymentstatuslogic.go    # 查询状态
    │   └── getuserpaymentslogic.go     # 支付记录
    ├── middleware/authmiddleware.go
    ├── svc/servicecontext.go
    └── types/types.go

backend/model/
├── paymentordermodel.go                # 支付单（FindByPaymentNo, UpdatePaySuccess, UpdateStatus）
├── paymentordermodel_gen.go            # 支付单（goctl 生成）
├── paymentrefundmodel.go               # 退款单（FindByRefundNo, UpdateStatus）
├── paymentrefundmodel_gen.go           # 退款单（goctl 生成）
└── sql/payment.sql                     # 建表 SQL
```

---

## 十六、一句话总结

支付系统的核心是**安全和幂等**。创建支付时用 Redis SETNX + 订单状态双重防重复；回调处理时用 Redis 幂等锁 + DB 状态机 + 终态检查三层保障；金额全程用分避免浮点精度问题；渠道验签防伪造回调；Strategy 模式让渠道扩展零侵入。整个系统的设计原则是：宁可多校验一次，不可少校验一次。
