package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/service/order/internal/svc"
	"github.com/ErizJ/JMall/backend/service/order/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderLogic) GetOrder(req *types.GetOrderReq) (resp *types.GetOrderResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	cacheKey := fmt.Sprintf("jmall:orders:user:%d", userID)

	var groups []types.OrderGroup
	if cacheErr := l.svcCtx.Cache.Get(l.ctx, cacheKey, &groups); cacheErr == nil {
		return &types.GetOrderResp{Code: "200", Orders: groups}, nil
	}

	rows, err := l.svcCtx.OrdersModel.FindByUserId(l.ctx, userID)
	if err != nil {
		return nil, err
	}

	// Batch fetch products
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

	// Group by order_id (preserve insertion order)
	orderMap := make(map[int64]*types.OrderGroup)
	var orderIDs []int64
	for _, row := range rows {
		p := productMap[row.ProductId]
		item := types.OrderItem{
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
		}

		g, exists := orderMap[row.OrderId]
		if !exists {
			g = &types.OrderGroup{
				OrderID:   row.OrderId,
				UserID:    row.UserId,
				Status:    row.Status,
				OrderTime: time.Unix(row.OrderTime, 0).Format("2006-01-02 15:04:05"),
			}
			orderMap[row.OrderId] = g
			orderIDs = append(orderIDs, row.OrderId)
		}
		g.Items = append(g.Items, item)
		g.ItemCount += row.ProductNum
		g.TotalAmount += row.ProductPrice * float64(row.ProductNum)
	}

	groups = make([]types.OrderGroup, 0, len(orderIDs))
	for _, oid := range orderIDs {
		groups = append(groups, *orderMap[oid])
	}

	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, groups, 2*time.Minute)
	return &types.GetOrderResp{Code: "200", Orders: groups}, nil
}
