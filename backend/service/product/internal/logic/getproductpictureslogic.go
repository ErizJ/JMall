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

type GetProductPicturesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProductPicturesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductPicturesLogic {
	return &GetProductPicturesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductPicturesLogic) GetProductPictures(req *types.GetProductPicturesReq) (resp *types.GetProductPicturesResp, err error) {
	cacheKey := fmt.Sprintf("jmall:product:pictures:%d", req.ProductID)

	var items []types.ProductPicture
	if cacheErr := l.svcCtx.Cache.Get(l.ctx, cacheKey, &items); cacheErr == nil {
		return &types.GetProductPicturesResp{Code: "200", Pictures: items}, nil
	}

	pictures, queryErr := l.svcCtx.ProductPictureModel.FindByProductId(l.ctx, req.ProductID)
	if queryErr != nil {
		return nil, queryErr
	}

	items = make([]types.ProductPicture, 0, len(pictures))
	for _, p := range pictures {
		pictureURL := ""
		if p.ProductPicture.Valid {
			pictureURL = p.ProductPicture.String
		}
		introStr := ""
		if p.Intro.Valid {
			introStr = p.Intro.String
		}
		items = append(items, types.ProductPicture{
			ID:             p.Id,
			ProductID:      p.ProductId,
			ProductPicture: pictureURL,
			Intro:          introStr,
		})
	}

	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, items, 5*time.Minute)
	return &types.GetProductPicturesResp{Code: "200", Pictures: items}, nil
}
