package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/payment/internal/channel"
	"github.com/ErizJ/JMall/backend/service/payment/internal/svc"
	"github.com/ErizJ/JMall/backend/service/payment/internal/types"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/zeromicro/go-zero/core/logx"
)

type PaymentNotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPaymentNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PaymentNotifyLogic {
	return &PaymentNotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// PaymentNotify 支付回调处理
//
// 主流安全实践：
//
//  1. 渠道验签：
//     调用 channel.VerifyNotify() 验证回调签名
//     防止伪造回调攻击（生产环境必须）
//
//  2. 金额校验：
//     回调中的实际支付金额必须与支付单金额一致
//     防止金额篡改攻击（如：支付 1 分钱但回调声称支付了 1000 元）
//
//  3. 三层幂等保障：
//     第一层：Redis SETNX（O(1) 快速拦截重复回调）
//     第二层：DB 状态机（UPDATE WHERE status IN (0,1)，已成功的单不会被重复更新）
//     第三层：MySQL 事务（支付单 + 订单原子更新）
//
//  4. 失败可重试：
//     事务失败时删除 Redis 幂等锁，允许渠道重新回调
func (l *PaymentNotifyLogic) PaymentNotify(req *types.PaymentNotifyReq) (resp *types.PaymentNotifyResp, err error) {
	// ========== 1. Redis 幂等锁（第一道防线） ==========
	idempotentKey := fmt.Sprintf("jmall:payment:notify:%s", req.PaymentNo)
	lockErr := l.svcCtx.Cache.SetNX(l.ctx, idempotentKey, "1", 24*time.Hour)
	if lockErr != nil {
		l.Infof("duplicate notify for payment_no=%s, skip", req.PaymentNo)
		return &types.PaymentNotifyResp{Code: "200"}, nil
	}

	// ========== 2. 查询支付单 ==========
	payment, findErr := l.svcCtx.PaymentOrderModel.FindByPaymentNo(l.ctx, req.PaymentNo)
	if findErr != nil {
		_ = l.svcCtx.Cache.Del(l.ctx, idempotentKey)
		l.Errorf("payment not found: %s, err: %v", req.PaymentNo, findErr)
		return &types.PaymentNotifyResp{Code: "404"}, nil
	}

	// ========== 3. 终态检查 ==========
	if payment.Status == model.PaymentStatusSuccess ||
		payment.Status == model.PaymentStatusRefund {
		return &types.PaymentNotifyResp{Code: "200"}, nil
	}

	// ========== 4. 渠道验签（安全校验） ==========
	ch, chErr := channel.Get(payment.Channel)
	if chErr != nil {
		_ = l.svcCtx.Cache.Del(l.ctx, idempotentKey)
		return &types.PaymentNotifyResp{Code: "002"}, nil
	}
	signParams := map[string]string{
		"payment_no":       req.PaymentNo,
		"channel_trade_no": req.ChannelTradeNo,
		"status":           req.Status,
		"sign":             req.Sign,
	}
	valid, verifyErr := ch.VerifyNotify(l.ctx, signParams)
	if verifyErr != nil || !valid {
		_ = l.svcCtx.Cache.Del(l.ctx, idempotentKey)
		l.Errorf("notify sign verification failed for payment_no=%s", req.PaymentNo)
		return &types.PaymentNotifyResp{Code: "401"}, nil // 签名验证失败
	}

	// ========== 5. 金额校验 ==========
	// 回调金额必须与支付单金额一致，防止金额篡改攻击
	if req.Amount > 0 && req.Amount != payment.Amount {
		_ = l.svcCtx.Cache.Del(l.ctx, idempotentKey)
		l.Errorf("amount mismatch: notify=%d, payment=%d, payment_no=%s",
			req.Amount, payment.Amount, req.PaymentNo)
		return &types.PaymentNotifyResp{Code: "015"}, nil // 金额不匹配
	}

	// ========== 6. 过期检查 ==========
	if payment.ExpireTime > 0 && time.Now().Unix() > payment.ExpireTime {
		// 事务内原子更新：支付单关闭 + 订单取消 + 库存回滚
		orderItems, findOrderErr := l.svcCtx.OrdersModel.FindByOrderId(l.ctx, payment.OrderId)
		if findOrderErr == nil && len(orderItems) > 0 && orderItems[0].Status == 0 {
			_ = l.svcCtx.PaymentOrderModel.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
				txPayment := l.svcCtx.PaymentOrderModel.WithSession(session)
				if updateErr := txPayment.UpdateStatus(ctx, req.PaymentNo, model.PaymentStatusClosed, time.Now().Unix()); updateErr != nil {
					return updateErr
				}
				txOrders := l.svcCtx.OrdersModel.WithSession(session)
				if statusErr := txOrders.UpdateStatusByOrderId(ctx, payment.OrderId, 2); statusErr != nil {
					return statusErr
				}
				txProduct := l.svcCtx.ProductModel.WithSession(session)
				for _, item := range orderItems {
					if incrErr := txProduct.IncrStock(ctx, item.ProductId, item.ProductNum); incrErr != nil {
						return incrErr
					}
				}
				return nil
			})
			for _, item := range orderItems {
				_ = l.svcCtx.Cache.Del(l.ctx, fmt.Sprintf("jmall:stock:%d", item.ProductId))
			}
		} else {
			// 订单状态已不是待支付，只关闭支付单
			_ = l.svcCtx.PaymentOrderModel.UpdateStatus(l.ctx, req.PaymentNo, model.PaymentStatusClosed, time.Now().Unix())
		}

		_ = l.svcCtx.Cache.Del(l.ctx, idempotentKey)
		return &types.PaymentNotifyResp{Code: "007"}, nil
	}

	now := time.Now().Unix()

	if req.Status == "success" {
		// ========== 7. 事务：更新支付单 + 订单状态 ==========
		txErr := l.svcCtx.PaymentOrderModel.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			txPayment := l.svcCtx.PaymentOrderModel.WithSession(session)
			if updateErr := txPayment.UpdatePaySuccess(ctx, req.PaymentNo, req.ChannelTradeNo, req.PaidTime, now); updateErr != nil {
				return updateErr
			}

			txOrders := l.svcCtx.OrdersModel.WithSession(session)
			if statusErr := txOrders.UpdateStatusByOrderId(ctx, payment.OrderId, 1); statusErr != nil {
				return statusErr
			}

			return nil
		})
		if txErr != nil {
			// 事务失败，删除幂等锁允许渠道重试
			_ = l.svcCtx.Cache.Del(l.ctx, idempotentKey)
			l.Errorf("notify transaction failed: %v", txErr)
			return &types.PaymentNotifyResp{Code: "500"}, nil
		}
	} else {
		// 支付失败：事务内原子更新支付单状态 + 取消订单 + 回滚库存
		orderItems, findOrderErr := l.svcCtx.OrdersModel.FindByOrderId(l.ctx, payment.OrderId)
		if findOrderErr == nil && len(orderItems) > 0 && orderItems[0].Status == 0 {
			_ = l.svcCtx.PaymentOrderModel.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
				txPayment := l.svcCtx.PaymentOrderModel.WithSession(session)
				if updateErr := txPayment.UpdateStatus(ctx, req.PaymentNo, model.PaymentStatusFailed, now); updateErr != nil {
					return updateErr
				}
				txOrders := l.svcCtx.OrdersModel.WithSession(session)
				if statusErr := txOrders.UpdateStatusByOrderId(ctx, payment.OrderId, 2); statusErr != nil {
					return statusErr
				}
				txProduct := l.svcCtx.ProductModel.WithSession(session)
				for _, item := range orderItems {
					if incrErr := txProduct.IncrStock(ctx, item.ProductId, item.ProductNum); incrErr != nil {
						return incrErr
					}
				}
				return nil
			})
			// 清理库存缓存
			for _, item := range orderItems {
				_ = l.svcCtx.Cache.Del(l.ctx, fmt.Sprintf("jmall:stock:%d", item.ProductId))
			}
		} else {
			// 订单状态已不是待支付，只更新支付单状态
			_ = l.svcCtx.PaymentOrderModel.UpdateStatus(l.ctx, req.PaymentNo, model.PaymentStatusFailed, now)
		}
	}

	// ========== 8. 清理锁和缓存 ==========
	lockKey := fmt.Sprintf("jmall:payment:lock:%d", payment.OrderId)
	_ = l.svcCtx.Cache.Del(l.ctx, lockKey)
	_ = l.svcCtx.Cache.Del(l.ctx,
		fmt.Sprintf("jmall:orders:user:%d", payment.UserId),
		fmt.Sprintf("jmall:payment:user:%d", payment.UserId),
	)

	return &types.PaymentNotifyResp{Code: "200"}, nil
}
