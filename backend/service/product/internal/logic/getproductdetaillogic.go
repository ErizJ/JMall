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

type GetProductDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductDetailLogic {
	return &GetProductDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductDetailLogic) GetProductDetail(req *types.GetProductDetailReq) (resp *types.GetProductDetailResp, err error) {
	cacheKey := fmt.Sprintf("jmall:product:detail:%d", req.ProductID)

	var item types.ProductItem
	if cacheErr := l.svcCtx.Cache.Get(l.ctx, cacheKey, &item); cacheErr == nil {
		return &types.GetProductDetailResp{Code: "200", Product: item}, nil
	}

	product, queryErr := l.svcCtx.ProductModel.FindOne(l.ctx, req.ProductID)
	if queryErr != nil {
		if queryErr == model.ErrNotFound {
			return &types.GetProductDetailResp{Code: "002"}, nil
		}
		return nil, queryErr
	}

	item = productToItem(product)
	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, item, 5*time.Minute)
	return &types.GetProductDetailResp{Code: "200", Product: item}, nil
}
