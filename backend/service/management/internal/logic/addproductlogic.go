// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"database/sql"

	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/management/internal/svc"
	"github.com/ErizJ/JMall/backend/service/management/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddProductLogic {
	return &AddProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddProductLogic) AddProduct(req *types.AddProductReq) (resp *types.AddProductResp, err error) {
	if _, insertErr := l.svcCtx.ProductModel.Insert(l.ctx, &model.Product{
		ProductName:         req.ProductName,
		CategoryId:          req.CategoryID,
		ProductTitle:        req.ProductTitle,
		ProductIntro:        req.ProductIntro,
		ProductPicture:      sql.NullString{String: req.ProductPicture, Valid: true},
		ProductPrice:        req.ProductPrice,
		ProductSellingPrice: req.ProductSellingPrice,
		ProductNum:          req.ProductNum,
		ProductIsPromotion:  req.ProductIsPromotion,
	}); insertErr != nil {
		return nil, insertErr
	}

	// Invalidate product list caches
	_ = l.svcCtx.Cache.Del(l.ctx,
		"jmall:products:all",
		"jmall:products:hot:7",
		"jmall:products:promotion:7",
		"jmall:product:phone:7",
		"jmall:product:shell:7",
		"jmall:product:charger:7",
	)

	return &types.AddProductResp{Code: "200"}, nil
}
