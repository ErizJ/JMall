// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/ErizJ/JMall/backend/service/product/internal/svc"
	"github.com/ErizJ/JMall/backend/service/product/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductBySearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProductBySearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductBySearchLogic {
	return &GetProductBySearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductBySearchLogic) GetProductBySearch(req *types.SearchProductReq) (resp *types.SearchProductResp, err error) {
	products, queryErr := l.svcCtx.ProductModel.FindBySearch(l.ctx, req.Keyword)
	if queryErr != nil {
		return nil, queryErr
	}

	return &types.SearchProductResp{Code: "200", Products: productsToItems(products)}, nil
}
