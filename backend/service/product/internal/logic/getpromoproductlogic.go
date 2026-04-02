// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/product/internal/svc"
	"github.com/ErizJ/JMall/backend/service/product/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPromoProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPromoProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPromoProductLogic {
	return &GetPromoProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPromoProductLogic) GetPromoProduct(req *types.GetPromoProductReq) (resp *types.GetPromoProductResp, err error) {
	category, catErr := l.svcCtx.CategoryModel.FindOneByCategoryName(l.ctx, req.CategoryName)
	if catErr != nil {
		if catErr == model.ErrNotFound {
			return &types.GetPromoProductResp{Code: "200", Products: []types.ProductItem{}}, nil
		}
		return nil, catErr
	}

	cacheKey := fmt.Sprintf("jmall:products:promo:%d", category.CategoryId)

	var result []types.ProductItem
	if cacheErr := l.svcCtx.Cache.Get(l.ctx, cacheKey, &result); cacheErr == nil {
		return &types.GetPromoProductResp{Code: "200", Products: result}, nil
	}

	products, queryErr := l.svcCtx.ProductModel.FindTopHotByCategory(l.ctx, category.CategoryId, 7)
	if queryErr != nil {
		return nil, queryErr
	}

	result = productsToItems(products)
	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, result, 5*time.Minute)
	return &types.GetPromoProductResp{Code: "200", Products: result}, nil
}
