package logic

import (
	"context"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/service/payment/internal/svc"
	"github.com/ErizJ/JMall/backend/service/payment/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserPaymentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserPaymentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserPaymentsLogic {
	return &GetUserPaymentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserPaymentsLogic) GetUserPayments(req *types.GetUserPaymentsReq) (resp *types.GetUserPaymentsResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	payments, findErr := l.svcCtx.PaymentOrderModel.FindByUserId(l.ctx, userID)
	if findErr != nil {
		return nil, findErr
	}

	var items []types.PaymentItem
	for _, p := range payments {
		items = append(items, types.PaymentItem{
			PaymentNo: p.PaymentNo,
			OrderID:   p.OrderId,
			Amount:    p.Amount,
			Channel:   p.Channel,
			Status:    p.Status,
			CreatedAt: p.CreatedAt,
		})
	}

	return &types.GetUserPaymentsResp{
		Code:     "200",
		Payments: items,
	}, nil
}
