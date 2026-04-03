package logic

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/payment/internal/channel"
	"github.com/ErizJ/JMall/backend/service/payment/internal/svc"
	"github.com/ErizJ/JMall/backend/service/payment/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePaymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreatePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentLogic {
	return &CreatePaymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreatePayment 创建支付单
//
// 核心流程：
//  1. 鉴权：从 context 提取 userID
//  2. 校验：检查订单是否存在、是否属于当前用户
//  3. 防重复：检查该订单是否已有进行中的支付单（Redis SETNX）
//  4. 计算金额：汇总订单商品总价，转换为分
//  5. 生成支付流水号：时间戳 + 随机数，全局唯一
//  6. 调用支付渠道：通过 Strategy 模式路由到具体渠道
//  7. 落库：写入 payment_order 表
//
// 防重复支付设计：
//   - Redis key: jmall:payment:lock:{order_id}，TTL = 支付过期时间
//   - SETNX 成功才允许创建，失败说明已有进行中的支付单
//   - 支付成功/失败/过期后删除 key
func (l *CreatePaymentLogic) CreatePayment(req *types.CreatePaymentReq) (resp *types.CreatePaymentResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	// 1. 校验支付渠道
	ch, chErr := channel.Get(req.Channel)
	if chErr != nil {
		return &types.CreatePaymentResp{Code: "002", PaymentNo: "", PayUrl: ""}, nil
	}

	// 2. 查询订单是否存在
	orderItems, findErr := l.svcCtx.OrdersModel.FindByOrderId(l.ctx, req.OrderID)
	if findErr != nil || len(orderItems) == 0 {
		return &types.CreatePaymentResp{Code: "003"}, nil // 订单不存在
	}

	// 3. 校验订单归属
	if orderItems[0].UserId != userID {
		return &types.CreatePaymentResp{Code: "004"}, nil // 非本人订单
	}

	// 3.5 校验订单状态：只有待支付的订单才能发起支付
	// 防止已支付/已退款的订单被重复支付（Redis 防重锁在支付完成后会被清理，
	// 所以必须在业务层做状态校验作为最终保障）
	if orderItems[0].Status != 0 {
		return &types.CreatePaymentResp{Code: "011"}, nil // 订单状态不允许支付
	}

	// 4. 防重复支付（Redis SETNX）
	lockKey := fmt.Sprintf("jmall:payment:lock:%d", req.OrderID)
	expireSec := l.svcCtx.Config.Payment.ExpireSeconds
	if expireSec <= 0 {
		expireSec = 1800 // 默认30分钟
	}
	lockErr := l.svcCtx.Cache.SetNX(l.ctx, lockKey, "1", time.Duration(expireSec)*time.Second)
	if lockErr != nil {
		// SETNX 失败，说明已有进行中的支付单
		return &types.CreatePaymentResp{Code: "005"}, nil // 重复支付
	}

	// 5. 计算订单总金额（转换为分）
	var totalAmountYuan float64
	for _, item := range orderItems {
		totalAmountYuan += item.ProductPrice * float64(item.ProductNum)
	}
	totalAmountFen := int64(math.Round(totalAmountYuan * 100))

	// 6. 生成支付流水号: PAY + 时间戳 + 3位随机数
	now := time.Now()
	paymentNo := fmt.Sprintf("PAY%d%06d", now.UnixMilli(), rand.Intn(1000000))

	// 7. 调用支付渠道创建预支付
	payResp, payErr := ch.CreatePayment(l.ctx, &channel.PayRequest{
		PaymentNo: paymentNo,
		OrderId:   req.OrderID,
		Amount:    totalAmountFen,
		Subject:   fmt.Sprintf("JMall订单-%d", req.OrderID),
		NotifyUrl: l.svcCtx.Config.Payment.NotifyUrl,
	})
	if payErr != nil {
		// 渠道调用失败，释放锁
		_ = l.svcCtx.Cache.Del(l.ctx, lockKey)
		l.Errorf("channel.CreatePayment failed: %v", payErr)
		return &types.CreatePaymentResp{Code: "006"}, nil
	}

	// 8. 落库
	_, insertErr := l.svcCtx.PaymentOrderModel.Insert(l.ctx, &model.PaymentOrder{
		PaymentNo:      paymentNo,
		OrderId:        req.OrderID,
		UserId:         userID,
		Amount:         totalAmountFen,
		Channel:        req.Channel,
		ChannelTradeNo: payResp.ChannelTradeNo,
		Status:         model.PaymentStatusPending,
		ExpireTime:     now.Unix() + expireSec,
		NotifyUrl:      l.svcCtx.Config.Payment.NotifyUrl,
		Extra:          payResp.Extra,
		CreatedAt:      now.Unix(),
		UpdatedAt:      now.Unix(),
	})
	if insertErr != nil {
		_ = l.svcCtx.Cache.Del(l.ctx, lockKey)
		l.Errorf("insert payment order failed: %v", insertErr)
		return nil, insertErr
	}

	return &types.CreatePaymentResp{
		Code:      "200",
		PaymentNo: paymentNo,
		PayUrl:    payResp.PayUrl,
	}, nil
}
