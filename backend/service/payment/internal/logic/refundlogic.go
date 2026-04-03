package logic

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/payment/internal/channel"
	"github.com/ErizJ/JMall/backend/service/payment/internal/svc"
	"github.com/ErizJ/JMall/backend/service/payment/internal/types"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundLogic {
	return &RefundLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Refund 申请退款
//
// 流程：
//  1. 查询支付单，校验状态必须是"支付成功"
//  2. 校验退款金额不超过支付金额
//  3. 调用支付渠道退款接口
//  4. 事务内：创建退款单 + 更新支付单状态 + 更新订单状态
func (l *RefundLogic) Refund(req *types.RefundReq) (resp *types.RefundResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	// 1. 查询支付单
	payment, findErr := l.svcCtx.PaymentOrderModel.FindByPaymentNo(l.ctx, req.PaymentNo)
	if findErr != nil {
		return &types.RefundResp{Code: "404"}, nil
	}

	// 2. 校验归属
	if payment.UserId != userID {
		return &types.RefundResp{Code: "004"}, nil
	}

	// 3. 校验状态：只有支付成功的单才能退款
	if payment.Status != model.PaymentStatusSuccess {
		return &types.RefundResp{Code: "008"}, nil // 状态不允许退款
	}

	// 3.5 退款幂等：防止重复提交退款
	refundLockKey := fmt.Sprintf("jmall:refund:lock:%s", req.PaymentNo)
	if lockErr := l.svcCtx.Cache.SetNX(l.ctx, refundLockKey, "1", 30*time.Second); lockErr != nil {
		return &types.RefundResp{Code: "016"}, nil // 退款处理中，请勿重复提交
	}

	// 4. 校验退款金额
	if req.RefundAmount <= 0 || req.RefundAmount > payment.Amount {
		_ = l.svcCtx.Cache.Del(l.ctx, refundLockKey)
		return &types.RefundResp{Code: "009"}, nil // 退款金额无效
	}

	// 5. 生成退款流水号
	now := time.Now()
	refundNo := fmt.Sprintf("REF%d%03d", now.UnixMilli(), rand.Intn(1000))

	// 6. 调用渠道退款
	ch, chErr := channel.Get(payment.Channel)
	if chErr != nil {
		_ = l.svcCtx.Cache.Del(l.ctx, refundLockKey)
		return &types.RefundResp{Code: "002"}, nil
	}

	refundResp, refundErr := ch.Refund(l.ctx, &channel.RefundRequest{
		PaymentNo:      payment.PaymentNo,
		RefundNo:       refundNo,
		ChannelTradeNo: payment.ChannelTradeNo,
		TotalAmount:    payment.Amount,
		RefundAmount:   req.RefundAmount,
		Reason:         req.Reason,
	})
	if refundErr != nil {
		_ = l.svcCtx.Cache.Del(l.ctx, refundLockKey)
		l.Errorf("channel refund failed: %v", refundErr)
		return &types.RefundResp{Code: "010"}, nil
	}

	// 7. 事务：创建退款单 + 更新支付单/订单状态 + 回滚库存
	txErr := l.svcCtx.PaymentOrderModel.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		txRefund := l.svcCtx.PaymentRefundModel.WithSession(session)
		if _, insertErr := txRefund.Insert(ctx, &model.PaymentRefund{
			RefundNo:        refundNo,
			PaymentNo:       payment.PaymentNo,
			OrderId:         payment.OrderId,
			UserId:          userID,
			RefundAmount:    req.RefundAmount,
			Reason:          req.Reason,
			Channel:         payment.Channel,
			ChannelRefundNo: refundResp.ChannelRefundNo,
			Status:          model.RefundStatusSuccess, // Mock 渠道直接成功
			CreatedAt:       now.Unix(),
			UpdatedAt:       now.Unix(),
		}); insertErr != nil {
			return insertErr
		}

		// 更新支付单状态为已退款
		txPayment := l.svcCtx.PaymentOrderModel.WithSession(session)
		if updateErr := txPayment.UpdateStatus(ctx, payment.PaymentNo, model.PaymentStatusRefund, now.Unix()); updateErr != nil {
			return updateErr
		}

		// 联动更新订单状态为已退款
		txOrders := l.svcCtx.OrdersModel.WithSession(session)
		if statusErr := txOrders.UpdateStatusByOrderId(ctx, payment.OrderId, 3); statusErr != nil {
			return statusErr
		}

		// 回滚库存：查询订单商品明细，逐个加回库存
		orderItems, findErr := txOrders.FindByOrderId(ctx, payment.OrderId)
		if findErr != nil {
			return findErr
		}
		txProduct := l.svcCtx.ProductModel.WithSession(session)
		for _, item := range orderItems {
			if incrErr := txProduct.IncrStock(ctx, item.ProductId, item.ProductNum); incrErr != nil {
				return incrErr
			}
		}

		return nil
	})
	if txErr != nil {
		_ = l.svcCtx.Cache.Del(l.ctx, refundLockKey)
		l.Errorf("refund transaction failed: %v", txErr)
		return nil, txErr
	}

	// 清理缓存（包括库存缓存，让下次从 DB 重新加载）
	cacheKeys := []string{
		fmt.Sprintf("jmall:orders:user:%d", userID),
		fmt.Sprintf("jmall:payment:user:%d", userID),
		refundLockKey,
	}
	// 清理退款涉及商品的库存缓存
	orderItems, _ := l.svcCtx.OrdersModel.FindByOrderId(l.ctx, payment.OrderId)
	for _, item := range orderItems {
		cacheKeys = append(cacheKeys, fmt.Sprintf("jmall:stock:%d", item.ProductId))
	}
	_ = l.svcCtx.Cache.Del(l.ctx, cacheKeys...)

	return &types.RefundResp{
		Code:     "200",
		RefundNo: refundNo,
	}, nil
}
