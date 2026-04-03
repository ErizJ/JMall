package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/payment/internal/svc"
	"github.com/ErizJ/JMall/backend/service/payment/internal/types"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/zeromicro/go-zero/core/logx"
)

type MockPayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMockPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MockPayLogic {
	return &MockPayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// MockPay 模拟用户完成支付
//
// 这个接口模拟了"用户在第三方支付页面完成支付"的行为。
// 在真实场景中，这个动作由微信/支付宝回调触发。
// Mock 模式下，前端调用此接口来模拟支付成功。
//
// 内部直接复用回调处理逻辑，保证行为一致。
func (l *MockPayLogic) MockPay(req *types.MockPayReq) (resp *types.MockPayResp, err error) {
	// 1. 幂等检查
	idempotentKey := fmt.Sprintf("jmall:payment:notify:%s", req.PaymentNo)
	lockErr := l.svcCtx.Cache.SetNX(l.ctx, idempotentKey, "1", 24*time.Hour)
	if lockErr != nil {
		return &types.MockPayResp{Code: "200"}, nil // 已处理
	}

	// 2. 查询支付单
	payment, findErr := l.svcCtx.PaymentOrderModel.FindByPaymentNo(l.ctx, req.PaymentNo)
	if findErr != nil {
		_ = l.svcCtx.Cache.Del(l.ctx, idempotentKey)
		return &types.MockPayResp{Code: "404"}, nil
	}

	// 3. 状态检查
	if payment.Status != model.PaymentStatusPending && payment.Status != model.PaymentStatusPaying {
		return &types.MockPayResp{Code: "008"}, nil
	}

	// 4. 检查过期
	if payment.ExpireTime > 0 && time.Now().Unix() > payment.ExpireTime {
		_ = l.svcCtx.PaymentOrderModel.UpdateStatus(l.ctx, req.PaymentNo, model.PaymentStatusClosed, time.Now().Unix())

		// 过期关闭时回滚库存并取消订单
		orderItems, findOrderErr := l.svcCtx.OrdersModel.FindByOrderId(l.ctx, payment.OrderId)
		if findOrderErr == nil && len(orderItems) > 0 && orderItems[0].Status == 0 {
			_ = l.svcCtx.OrdersModel.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
				txOrders := l.svcCtx.OrdersModel.WithSession(session)
				if statusErr := txOrders.UpdateStatusByOrderId(ctx, payment.OrderId, 2); statusErr != nil {
					return statusErr
				}
				txProduct := l.svcCtx.ProductModel.WithSession(session)
				for _, oi := range orderItems {
					if incrErr := txProduct.IncrStock(ctx, oi.ProductId, oi.ProductNum); incrErr != nil {
						return incrErr
					}
				}
				return nil
			})
			for _, oi := range orderItems {
				_ = l.svcCtx.Cache.Del(l.ctx, fmt.Sprintf("jmall:stock:%d", oi.ProductId))
			}
		}

		_ = l.svcCtx.Cache.Del(l.ctx, idempotentKey)
		return &types.MockPayResp{Code: "007"}, nil
	}

	now := time.Now()
	channelTradeNo := fmt.Sprintf("MOCK_TRADE_%s_%d", req.PaymentNo, now.UnixMilli())

	// 5. 事务：更新支付单 + 订单状态
	txErr := l.svcCtx.PaymentOrderModel.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		txPayment := l.svcCtx.PaymentOrderModel.WithSession(session)
		if updateErr := txPayment.UpdatePaySuccess(ctx, req.PaymentNo, channelTradeNo, now.Unix(), now.Unix()); updateErr != nil {
			return updateErr
		}

		txOrders := l.svcCtx.OrdersModel.WithSession(session)
		if statusErr := txOrders.UpdateStatusByOrderId(ctx, payment.OrderId, 1); statusErr != nil {
			return statusErr
		}

		return nil
	})
	if txErr != nil {
		_ = l.svcCtx.Cache.Del(l.ctx, idempotentKey)
		l.Errorf("mock pay transaction failed: %v", txErr)
		return &types.MockPayResp{Code: "500"}, nil
	}

	// 6. 清理锁和缓存
	lockKey := fmt.Sprintf("jmall:payment:lock:%d", payment.OrderId)
	_ = l.svcCtx.Cache.Del(l.ctx, lockKey)
	_ = l.svcCtx.Cache.Del(l.ctx,
		fmt.Sprintf("jmall:orders:user:%d", payment.UserId),
		fmt.Sprintf("jmall:payment:user:%d", payment.UserId),
	)

	return &types.MockPayResp{Code: "200"}, nil
}
