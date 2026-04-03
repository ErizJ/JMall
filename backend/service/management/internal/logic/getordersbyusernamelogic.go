// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"time"

	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/management/internal/svc"
	"github.com/ErizJ/JMall/backend/service/management/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrdersByUserNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrdersByUserNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrdersByUserNameLogic {
	return &GetOrdersByUserNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrdersByUserNameLogic) GetOrdersByUserName(req *types.GetOrdersByUserNameReq) (resp *types.GetOrdersByUserNameResp, err error) {
	user, err := l.svcCtx.UsersModel.FindOneByUserName(l.ctx, req.UserName)
	if err == model.ErrNotFound {
		return &types.GetOrdersByUserNameResp{Code: "200", Orders: []types.MgmtOrderItem{}}, nil
	}
	if err != nil {
		return nil, err
	}

	rows, err := l.svcCtx.OrdersModel.FindByUserId(l.ctx, user.UserId)
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

	orders := make([]types.MgmtOrderItem, 0, len(rows))
	for _, row := range rows {
		p := productMap[row.ProductId]
		orders = append(orders, types.MgmtOrderItem{
			ID:           row.Id,
			OrderID:      row.OrderId,
			UserID:       row.UserId,
			UserName:     user.UserName,
			ProductID:    row.ProductId,
			ProductName:  p.name,
			ProductImg:   p.img,
			ProductNum:   row.ProductNum,
			ProductPrice: row.ProductPrice,
			OrderTime:    time.Unix(row.OrderTime, 0).Format("2006-01-02 15:04:05"),
			Status:       row.Status,
		})
	}

	return &types.GetOrdersByUserNameResp{
		Code:   "200",
		Orders: orders,
	}, nil
}
