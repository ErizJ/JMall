// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/service/cart/internal/svc"
	"github.com/ErizJ/JMall/backend/service/cart/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCartLogic {
	return &GetCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCartLogic) GetCart(req *types.GetCartReq) (resp *types.GetCartResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	cacheKey := fmt.Sprintf("jmall:cart:user:%d", userID)

	var items []types.CartItem
	if cacheErr := l.svcCtx.Cache.Get(l.ctx, cacheKey, &items); cacheErr == nil {
		return &types.GetCartResp{Code: "200", Items: items}, nil
	}

	rows, err := l.svcCtx.ShoppingcartModel.FindByUserId(l.ctx, userID)
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

	productMap := make(map[int64]*struct {
		name    string
		img     string
		price   float64
		maxNum  int64
	}, len(products))
	for _, p := range products {
		maxNum := int64(math.Floor(float64(p.ProductNum) / 2))
		productMap[p.ProductId] = &struct {
			name    string
			img     string
			price   float64
			maxNum  int64
		}{
			name:   p.ProductName,
			img:    p.ProductPicture.String,
			price:  p.ProductSellingPrice,
			maxNum: maxNum,
		}
	}

	items = make([]types.CartItem, 0, len(rows))
	for _, row := range rows {
		p := productMap[row.ProductId]
		if p == nil {
			continue
		}
		items = append(items, types.CartItem{
			ID:          row.Id,
			UserID:      row.UserId,
			ProductID:   row.ProductId,
			ProductName: p.name,
			ProductImg:  p.img,
			Price:       p.price,
			Num:         row.Num,
			MaxNum:      p.maxNum,
		})
	}

	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, items, 2*time.Minute)
	return &types.GetCartResp{
		Code:  "200",
		Items: items,
	}, nil
}
