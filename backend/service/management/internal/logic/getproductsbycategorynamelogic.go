// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/ErizJ/JMall/backend/service/management/internal/svc"
	"github.com/ErizJ/JMall/backend/service/management/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductsByCategoryNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProductsByCategoryNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductsByCategoryNameLogic {
	return &GetProductsByCategoryNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductsByCategoryNameLogic) GetProductsByCategoryName(req *types.GetProductsByCategoryNameReq) (resp *types.GetProductsByCategoryNameResp, err error) {
	category, err := l.svcCtx.CategoryModel.FindOneByCategoryName(l.ctx, req.CategoryName)
	if err != nil {
		return nil, err
	}

	products, err := l.svcCtx.ProductModel.FindByCategory(l.ctx, category.CategoryId)
	if err != nil {
		return nil, err
	}

	items := make([]types.MgmtProductItem, 0, len(products))
	for _, p := range products {
		items = append(items, types.MgmtProductItem{
			ProductID:           p.ProductId,
			ProductName:         p.ProductName,
			CategoryID:          p.CategoryId,
			ProductTitle:        p.ProductTitle,
			ProductPicture:      p.ProductPicture.String,
			ProductPrice:        p.ProductPrice,
			ProductSellingPrice: p.ProductSellingPrice,
			ProductNum:          p.ProductNum,
			ProductSales:        p.ProductSales.Int64,
			ProductIsPromotion:  p.ProductIsPromotion,
		})
	}

	return &types.GetProductsByCategoryNameResp{
		Code:     "200",
		Products: items,
	}, nil
}
