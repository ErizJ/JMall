// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"time"

	"github.com/ErizJ/JMall/backend/service/product/internal/svc"
	"github.com/ErizJ/JMall/backend/service/product/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPhoneListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPhoneListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPhoneListLogic {
	return &GetPhoneListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPhoneListLogic) GetPhoneList() (resp *types.GetAllProductResp, err error) {
	const cacheKey = "jmall:product:phone:7"

	var result []types.ProductItem
	if cacheErr := l.svcCtx.Cache.Get(l.ctx, cacheKey, &result); cacheErr == nil {
		return &types.GetAllProductResp{Code: "200", Products: result}, nil
	}

	// category_id = 1 (phones)
	products, queryErr := l.svcCtx.ProductModel.FindTopHotByCategory(l.ctx, 1, 7)
	if queryErr != nil {
		return nil, queryErr
	}

	result = productsToItems(products)
	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, result, 5*time.Minute)
	return &types.GetAllProductResp{Code: "200", Products: result}, nil
}
