// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"time"

	"github.com/ErizJ/JMall/backend/service/management/internal/svc"
	"github.com/ErizJ/JMall/backend/service/management/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAllOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllOrdersLogic {
	return &GetAllOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAllOrdersLogic) GetAllOrders() (resp *types.GetAllOrdersResp, err error) {
	rows, err := l.svcCtx.OrdersModel.FindAllWithDetails(l.ctx)
	if err != nil {
		return nil, err
	}

	orders := make([]types.MgmtOrderItem, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, types.MgmtOrderItem{
			ID:           row.Id,
			OrderID:      row.OrderId,
			UserID:       row.UserId,
			UserName:     row.UserName,
			ProductID:    row.ProductId,
			ProductName:  row.ProductName,
			ProductNum:   row.ProductNum,
			ProductPrice: row.ProductPrice,
			OrderTime:    time.Unix(row.OrderTime, 0).Format("2006-01-02 15:04:05"),
		})
	}

	return &types.GetAllOrdersResp{
		Code:   "200",
		Orders: orders,
	}, nil
}
