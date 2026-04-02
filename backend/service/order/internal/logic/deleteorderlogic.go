// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"errors"
	"fmt"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/service/order/internal/svc"
	"github.com/ErizJ/JMall/backend/service/order/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOrderLogic {
	return &DeleteOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteOrderLogic) DeleteOrder(req *types.DeleteOrderReq) (resp *types.DeleteOrderResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	// Fetch order rows to verify ownership before deleting
	rows, err := l.svcCtx.OrdersModel.FindByOrderId(l.ctx, req.OrderID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return &types.DeleteOrderResp{Code: "002"}, nil
	}
	// Ensure the requesting user owns this order
	if rows[0].UserId != userID {
		return nil, errors.New("forbidden")
	}

	if err := l.svcCtx.OrdersModel.DeleteByOrderId(l.ctx, req.OrderID); err != nil {
		return nil, err
	}

	_ = l.svcCtx.Cache.Del(l.ctx, fmt.Sprintf("jmall:orders:user:%d", userID))

	return &types.DeleteOrderResp{Code: "200"}, nil
}
