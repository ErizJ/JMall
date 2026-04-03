package logic

import (
	"context"

	"github.com/ErizJ/JMall/backend/service/payment/internal/svc"
	"github.com/ErizJ/JMall/backend/service/payment/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaymentStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPaymentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentStatusLogic {
	return &GetPaymentStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPaymentStatusLogic) GetPaymentStatus(req *types.GetPaymentStatusReq) (resp *types.GetPaymentStatusResp, err error) {
	payment, findErr := l.svcCtx.PaymentOrderModel.FindByPaymentNo(l.ctx, req.PaymentNo)
	if findErr != nil {
		return &types.GetPaymentStatusResp{Code: "404"}, nil
	}

	return &types.GetPaymentStatusResp{
		Code:      "200",
		PaymentNo: payment.PaymentNo,
		OrderID:   payment.OrderId,
		Amount:    payment.Amount,
		Channel:   payment.Channel,
		Status:    payment.Status,
		PaidTime:  payment.PaidTime,
	}, nil
}
