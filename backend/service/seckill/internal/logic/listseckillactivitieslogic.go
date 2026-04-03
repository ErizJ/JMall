package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/ErizJ/JMall/backend/service/seckill/internal/svc"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListSeckillActivitiesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListSeckillActivitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSeckillActivitiesLogic {
	return &ListSeckillActivitiesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListSeckillActivities 获取当前进行中的秒杀活动列表
func (l *ListSeckillActivitiesLogic) ListSeckillActivities() (resp *types.ListSeckillActivitiesResp, err error) {
	now := time.Now().Unix()
	activities, findErr := l.svcCtx.SeckillActivityModel.FindOngoing(l.ctx, now)
	if findErr != nil {
		return nil, findErr
	}

	list := make([]types.SeckillActivityResp, 0, len(activities))
	for _, a := range activities {
		// 优先读 Redis 实时库存，与详情页保持一致
		stock := a.AvailableStock
		stockKey := fmt.Sprintf("seckill:stock:%d", a.Id)
		var redisStock int64
		if getErr := l.svcCtx.Cache.Get(l.ctx, stockKey, &redisStock); getErr == nil {
			stock = redisStock
		}
		list = append(list, types.SeckillActivityResp{
			Code:          "200",
			ActivityID:    a.Id,
			Title:         a.Title,
			ProductID:     a.ProductId,
			SeckillPrice:  a.SeckillPrice,
			OriginalPrice: a.OriginalPrice,
			Stock:         stock,
			LimitPerUser:  a.LimitPerUser,
			StartTime:     a.StartTime,
			EndTime:       a.EndTime,
			Status:        int(a.Status),
		})
	}

	return &types.ListSeckillActivitiesResp{
		Code:       "200",
		Activities: list,
	}, nil
}
