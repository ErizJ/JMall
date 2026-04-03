// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"time"

	"github.com/ErizJ/JMall/backend/service/order/internal/svc"
	"github.com/ErizJ/JMall/backend/service/order/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderDetailLogic {
	return &GetOrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderDetailLogic) GetOrderDetail(req *types.GetOrderDetailReq) (resp *types.GetOrderDetailResp, err error) {
	rows, err := l.svcCtx.OrdersModel.FindByOrderId(l.ctx, req.OrderID)
	if err != nil {
		return nil, err
	}

	// Batch fetch products to avoid N+1 queries
	productIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		productIDs = append(productIDs, row.ProductId)
	}
	products, err := l.svcCtx.ProductModel.FindByIds(l.ctx, productIDs)
	if err != nil {
		return nil, err
	}
	productMap := make(map[int64]struct{ name, img string }, len(products))
	for _, p := range products {
		productMap[p.ProductId] = struct{ name, img string }{p.ProductName, p.ProductPicture.String}
	}

	orders := make([]types.OrderItem, 0, len(rows))
	for _, row := range rows {
		p := productMap[row.ProductId]
		orders = append(orders, types.OrderItem{
			ID:           row.Id,
			OrderID:      row.OrderId,
			UserID:       row.UserId,
			ProductID:    row.ProductId,
			ProductName:  p.name,
			ProductImg:   p.img,
			ProductNum:   row.ProductNum,
			ProductPrice: row.ProductPrice,
			OrderTime:    time.Unix(row.OrderTime, 0).Format("2006-01-02 15:04:05"),
			Status:       row.Status,
		})
	}

	return &types.GetOrderDetailResp{
		Code:   "200",
		Orders: orders,
	}, nil
}
