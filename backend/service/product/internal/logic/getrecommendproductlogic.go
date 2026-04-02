// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"sort"
	"time"

	"github.com/ErizJ/JMall/backend/service/product/internal/svc"
	"github.com/ErizJ/JMall/backend/service/product/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetRecommendProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRecommendProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRecommendProductLogic {
	return &GetRecommendProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRecommendProductLogic) GetRecommendProduct(req *types.GetRecommendProductReq) (resp *types.GetRecommendProductResp, err error) {
	const cacheKey = "jmall:product:recommend:personal"

	var result []types.ProductItem
	if cacheErr := l.svcCtx.Cache.Get(l.ctx, cacheKey, &result); cacheErr == nil {
		return &types.GetRecommendProductResp{Code: "200", Products: result}, nil
	}

	categories, catErr := l.svcCtx.CategoryModel.FindAll(l.ctx)
	if catErr != nil {
		return nil, catErr
	}

	if len(categories) == 0 {
		return &types.GetRecommendProductResp{Code: "200", Products: []types.ProductItem{}}, nil
	}

	// Sort by CategoryHot DESC and pick top category
	sort.Slice(categories, func(i, j int) bool {
		hotI := int64(0)
		if categories[i].CategoryHot.Valid {
			hotI = categories[i].CategoryHot.Int64
		}
		hotJ := int64(0)
		if categories[j].CategoryHot.Valid {
			hotJ = categories[j].CategoryHot.Int64
		}
		return hotI > hotJ
	})
	topCategoryId := categories[0].CategoryId

	products, queryErr := l.svcCtx.ProductModel.FindTopHotByCategory(l.ctx, topCategoryId, 7)
	if queryErr != nil {
		return nil, queryErr
	}

	result = productsToItems(products)
	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, result, 5*time.Minute)
	return &types.GetRecommendProductResp{Code: "200", Products: result}, nil
}
