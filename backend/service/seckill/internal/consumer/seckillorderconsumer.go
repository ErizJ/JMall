package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/ErizJ/JMall/backend/cache"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Lua 脚本：回滚库存 + 限购计数
const luaSeckillRollback = `
redis.call('INCRBY', KEYS[1], ARGV[1])
local bought = tonumber(redis.call('GET', KEYS[2]) or '0')
if bought > 0 then
    redis.call('DECRBY', KEYS[2], ARGV[1])
end
return 1
`

// SeckillOrderConsumer 消费 seckill-order topic，执行实际下单。
//
// 核心保障：
//  1. 幂等：Redis SETNX + MySQL 唯一索引 (activity_id, user_id)
//  2. 不丢单：框架层最多重试 3 次，全部失败才跳过
//  3. 不超卖：MySQL WHERE product_num >= num 最终防线
//
// 重试设计：
//   框架（kafka/consumer.go）对同一条消息最多调用 Consume 3 次。
//   Consume 返回 error 表示"可重试"，返回 nil 表示"已处理完毕（成功或不可恢复的失败）"。
//   事务失败时返回 error 触发重试，不回滚 Redis（重试可能成功）。
//   只在确定放弃时（token 过期、格式错误）才回滚 Redis 并返回 nil。
type SeckillOrderConsumer struct {
	cache                *cache.Client
	ordersModel          model.OrdersModel
	productModel         model.ProductModel
	seckillOrderModel    model.SeckillOrderModel
	seckillActivityModel model.SeckillActivityModel
}

func NewSeckillOrderConsumer(
	cache *cache.Client,
	ordersModel model.OrdersModel,
	productModel model.ProductModel,
	seckillOrderModel model.SeckillOrderModel,
	seckillActivityModel model.SeckillActivityModel,
) *SeckillOrderConsumer {
	return &SeckillOrderConsumer{
		cache:                cache,
		ordersModel:          ordersModel,
		productModel:         productModel,
		seckillOrderModel:    seckillOrderModel,
		seckillActivityModel: seckillActivityModel,
	}
}

// Consume processes a single seckill order message from Kafka.
//
// 返回值语义（配合框架重试）：
//   - nil  → 已处理完毕（成功 or 不可恢复失败），框架提交 offset
//   - error → 可重试的临时失败，框架会再次调用 Consume（最多 3 次）
func (c *SeckillOrderConsumer) Consume(ctx context.Context, key, value string) error {
	var msg types.SeckillMessage
	if err := json.Unmarshal([]byte(value), &msg); err != nil {
		logx.Errorf("unmarshal seckill msg failed: %v", err)
		return nil // 格式错误，不可恢复，丢弃
	}

	// ========== 1. 幂等校验（Redis SETNX） ==========
	idempotentKey := fmt.Sprintf("seckill:idempotent:%s", msg.Token)
	if err := c.cache.SetNX(ctx, idempotentKey, "1", 24*time.Hour); err != nil {
		logx.Infof("duplicate seckill msg, token=%s, skip", msg.Token)
		return nil // 已处理过，跳过
	}

	// ========== 2. Token 有效性校验（只读不删） ==========
	tokenKey := fmt.Sprintf("seckill:token:%s", msg.Token)
	var tokenData types.SeckillMessage
	if err := c.cache.Get(ctx, tokenKey, &tokenData); err != nil {
		// token 过期（5min TTL），不可恢复
		logx.Errorf("seckill token expired: %s", msg.Token)
		c.rollbackRedisStock(ctx, msg)
		c.writeResult(ctx, msg.Token, 2, 0, "令牌已过期，请重新抢购")
		return nil // 不可恢复，不重试
	}

	// ========== 3. MySQL 事务下单 ==========
	orderID := time.Now().UnixMilli()*1000 + int64(rand.Intn(1000))

	txErr := c.seckillOrderModel.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		txOrders := c.ordersModel.WithSession(session)
		txProduct := c.productModel.WithSession(session)
		txSeckillOrder := c.seckillOrderModel.WithSession(session)
		txSeckillActivity := c.seckillActivityModel.WithSession(session)

		// 3a. 插入 orders 表
		if _, err := txOrders.Insert(ctx, &model.Orders{
			OrderId:      orderID,
			UserId:       msg.UserID,
			ProductId:    msg.ProductID,
			ProductNum:   msg.Num,
			ProductPrice: float64(msg.SeckillPrice) / 100,
			OrderTime:    time.Now().Unix(),
			Status:       0,
		}); err != nil {
			return fmt.Errorf("insert order: %w", err)
		}

		// 3b. 扣减商品总库存（WHERE product_num >= num 防超卖）
		if err := txProduct.DecrStock(ctx, msg.ProductID, msg.Num); err != nil {
			return fmt.Errorf("decr product stock: %w", err)
		}

		// 3c. 扣减秒杀活动库存（与 Redis 保持一致，防止 Redis 重启后 WarmUp 用旧值超卖）
		if err := txSeckillActivity.DecrStock(ctx, msg.ActivityID, msg.Num); err != nil {
			return fmt.Errorf("decr seckill stock: %w", err)
		}

		// 3d. 插入秒杀订单关联（唯一索引 activity_id+user_id 防重）
		if _, err := txSeckillOrder.Insert(ctx, &model.SeckillOrder{
			ActivityId:   msg.ActivityID,
			OrderId:      orderID,
			UserId:       msg.UserID,
			ProductId:    msg.ProductID,
			SeckillPrice: msg.SeckillPrice,
			Num:          msg.Num,
			CreatedAt:    time.Now().Unix(),
		}); err != nil {
			return fmt.Errorf("insert seckill order: %w", err)
		}

		return nil
	})

	if txErr != nil {
		logx.Errorf("seckill tx failed: token=%s, err=%v", msg.Token, txErr)
		// 删除幂等锁，让重试能通过步骤 1
		_ = c.cache.Del(ctx, idempotentKey)
		// 不回滚 Redis 库存！框架会立即重试，重试可能成功。
		// 只在框架放弃后（3 次都失败）由 OnExhausted 回滚。
		return txErr // 返回 error → 触发框架重试
	}

	// ========== 4. 事务成功 ==========
	_ = c.cache.Del(ctx, tokenKey) // 删除 token（一次性消费）
	c.writeResult(ctx, msg.Token, 1, orderID, "下单成功")
	logx.Infof("seckill order ok: token=%s, order=%d, user=%d", msg.Token, orderID, msg.UserID)
	return nil
}

// OnExhausted 当框架重试耗尽时调用，做最终清理。
// 回滚 Redis 库存 + 写入失败结果。
func (c *SeckillOrderConsumer) OnExhausted(ctx context.Context, key, value string) {
	var msg types.SeckillMessage
	if err := json.Unmarshal([]byte(value), &msg); err != nil {
		return
	}
	logx.Errorf("seckill order exhausted retries: token=%s, user=%d", msg.Token, msg.UserID)
	c.rollbackRedisStock(ctx, msg)
	c.writeResult(ctx, msg.Token, 2, 0, "下单失败，请重新抢购")
}

// writeResult 写入秒杀结果到 Redis，供前端轮询
func (c *SeckillOrderConsumer) writeResult(ctx context.Context, token string, status int, orderID int64, msg string) {
	resultKey := fmt.Sprintf("seckill:result:%s", token)
	result := map[string]interface{}{
		"status":   status,
		"order_id": orderID,
		"msg":      msg,
	}
	if err := c.cache.Set(ctx, resultKey, result, 30*time.Minute); err != nil {
		logx.Errorf("write seckill result failed: token=%s, err=%v", token, err)
	}
}

// rollbackRedisStock 回滚 Redis 中的库存和限购计数
func (c *SeckillOrderConsumer) rollbackRedisStock(ctx context.Context, msg types.SeckillMessage) {
	stockKey := fmt.Sprintf("seckill:stock:%d", msg.ActivityID)
	boughtKey := fmt.Sprintf("seckill:bought:%d:%d", msg.ActivityID, msg.UserID)
	_, err := c.cache.Eval(ctx, luaSeckillRollback,
		[]string{stockKey, boughtKey},
		msg.Num,
	)
	if err != nil {
		logx.Errorf("rollback redis stock failed: activity=%d, user=%d, err=%v",
			msg.ActivityID, msg.UserID, err)
	}
}
