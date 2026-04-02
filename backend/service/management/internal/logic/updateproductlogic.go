// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ErizJ/JMall/backend/service/management/internal/svc"
	"github.com/ErizJ/JMall/backend/service/management/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductLogic {
	return &UpdateProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProductLogic) UpdateProduct(req *types.UpdateProductReq) (resp *types.UpdateProductResp, err error) {
	product, err := l.svcCtx.ProductModel.FindOne(l.ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	if req.ProductName != "" {
		product.ProductName = req.ProductName
	}
	if req.ProductTitle != "" {
		product.ProductTitle = req.ProductTitle
	}
	if req.ProductIntro != "" {
		product.ProductIntro = req.ProductIntro
	}
	if req.ProductPicture != "" {
		product.ProductPicture = sql.NullString{String: req.ProductPicture, Valid: true}
	}
	if req.ProductPrice != nil {
		product.ProductPrice = *req.ProductPrice
	}
	if req.ProductSellingPrice != nil {
		product.ProductSellingPrice = *req.ProductSellingPrice
	}
	if req.ProductNum != nil {
		product.ProductNum = *req.ProductNum
	}
	if req.ProductIsPromotion != nil {
		product.ProductIsPromotion = *req.ProductIsPromotion
	}

	if err := l.svcCtx.ProductModel.Update(l.ctx, product); err != nil {
		return nil, err
	}

	// Invalidate product caches
	_ = l.svcCtx.Cache.Del(l.ctx,
		fmt.Sprintf("jmall:product:detail:%d", req.ProductID),
		"jmall:products:all",
		"jmall:products:hot:7",
		"jmall:products:promotion:7",
		"jmall:product:phone:7",
		"jmall:product:shell:7",
		"jmall:product:charger:7",
	)

	return &types.UpdateProductResp{Code: "200"}, nil
}
