// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/ErizJ/JMall/backend/service/product/internal/svc"
	"github.com/ErizJ/JMall/backend/service/product/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductByCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProductByCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductByCategoryLogic {
	return &GetProductByCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductByCategoryLogic) GetProductByCategory(req *types.GetProductByCategoryReq) (resp *types.GetProductByCategoryResp, err error) {
	cacheKey := fmt.Sprintf("jmall:products:category:%d", req.CategoryID)

	var result []types.ProductItem
	if cacheErr := l.svcCtx.Cache.Get(l.ctx, cacheKey, &result); cacheErr == nil {
		return &types.GetProductByCategoryResp{Code: "200", Products: result}, nil
	}

	products, queryErr := l.svcCtx.ProductModel.FindByCategory(l.ctx, req.CategoryID)
	if queryErr != nil {
		return nil, queryErr
	}

	result = productsToItems(products)
	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, result, 10*time.Minute)
	return &types.GetProductByCategoryResp{Code: "200", Products: result}, nil
}
