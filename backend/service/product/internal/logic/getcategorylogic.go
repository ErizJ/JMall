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

type GetCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCategoryLogic {
	return &GetCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCategoryLogic) GetCategory() (resp *types.GetCategoryResp, err error) {
	const cacheKey = "jmall:categories"

	var items []types.CategoryItem
	if cacheErr := l.svcCtx.Cache.Get(l.ctx, cacheKey, &items); cacheErr == nil {
		return &types.GetCategoryResp{Code: "200", Categories: items}, nil
	}

	categories, queryErr := l.svcCtx.CategoryModel.FindAll(l.ctx)
	if queryErr != nil {
		return nil, queryErr
	}

	items = make([]types.CategoryItem, 0, len(categories))
	for _, c := range categories {
		hotVal := int64(0)
		if c.CategoryHot.Valid {
			hotVal = c.CategoryHot.Int64
		}
		items = append(items, types.CategoryItem{
			CategoryID:   c.CategoryId,
			CategoryName: c.CategoryName,
			CategoryHot:  hotVal,
		})
	}

	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, items, 10*time.Minute)
	return &types.GetCategoryResp{Code: "200", Categories: items}, nil
}
