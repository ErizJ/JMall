// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"time"

	"github.com/ErizJ/JMall/backend/service/management/internal/svc"
	"github.com/ErizJ/JMall/backend/service/management/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCarouselLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCarouselLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCarouselLogic {
	return &GetCarouselLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCarouselLogic) GetCarousel() (resp *types.GetCarouselResp, err error) {
	const cacheKey = "jmall:carousel"

	var items []types.CarouselItem
	if cacheErr := l.svcCtx.Cache.Get(l.ctx, cacheKey, &items); cacheErr == nil {
		return &types.GetCarouselResp{Code: "200", Carousel: items}, nil
	}

	rows, err := l.svcCtx.CarouselModel.FindAll(l.ctx)
	if err != nil {
		return nil, err
	}

	items = make([]types.CarouselItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, types.CarouselItem{
			CarouselID: row.CarouselId,
			ImgPath:    row.ImgPath,
			Describes:  row.Describes,
		})
	}

	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, items, 10*time.Minute)
	return &types.GetCarouselResp{
		Code:     "200",
		Carousel: items,
	}, nil
}
