package logic

import (
	"context"
	"fmt"

	"github.com/ErizJ/JMall/backend/service/seckill/internal/svc"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetSeckillActivityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSeckillActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSeckillActivityLogic {
	return &GetSeckillActivityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetSeckillActivity 获取秒杀活动详情（含实时库存）
func (l *GetSeckillActivityLogic) GetSeckillActivity(req *types.GetSeckillActivityReq) (resp *types.SeckillActivityResp, err error) {
	activity, actErr := l.svcCtx.SeckillActivityModel.FindOne(l.ctx, req.ActivityID)
	if actErr != nil {
		return &types.SeckillActivityResp{Code: "ACTIVITY_NOT_FOUND"}, nil
	}

	// 尝试从 Redis 读取实时库存（比 DB 更准确）
	stock := activity.AvailableStock
	stockKey := fmt.Sprintf("seckill:stock:%d", req.ActivityID)
	var redisStock int64
	if getErr := l.svcCtx.Cache.Get(l.ctx, stockKey, &redisStock); getErr == nil {
		stock = redisStock
	}

	return &types.SeckillActivityResp{
		Code:          "200",
		ActivityID:    activity.Id,
		Title:         activity.Title,
		ProductID:     activity.ProductId,
		SeckillPrice:  activity.SeckillPrice,
		OriginalPrice: activity.OriginalPrice,
		Stock:         stock,
		LimitPerUser:  activity.LimitPerUser,
		StartTime:     activity.StartTime,
		EndTime:       activity.EndTime,
		Status:        int(activity.Status),
	}, nil
}
