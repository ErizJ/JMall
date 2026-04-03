package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// WarmUp 预热秒杀活动数据到 Redis。
// 在服务启动时调用，将即将开始和进行中的活动库存加载到 Redis。
// 这样秒杀开始时 Redis 已有数据，避免冷启动。
func WarmUp(ctx context.Context, svcCtx *svc.ServiceContext) {
	now := time.Now().Unix()

	// 加载即将开始的活动（1 小时内）
	upcoming, err := svcCtx.SeckillActivityModel.FindUpcoming(ctx, now)
	if err != nil {
		logx.Errorf("warmup: find upcoming activities failed: %v", err)
		return
	}

	// 加载进行中的活动
	ongoing, err := svcCtx.SeckillActivityModel.FindOngoing(ctx, now)
	if err != nil {
		logx.Errorf("warmup: find ongoing activities failed: %v", err)
		return
	}

	activities := make([]*model.SeckillActivity, 0, len(upcoming)+len(ongoing))
	seen := make(map[int64]bool)
	for _, a := range upcoming {
		if !seen[a.Id] {
			activities = append(activities, a)
			seen[a.Id] = true
		}
	}
	for _, a := range ongoing {
		if !seen[a.Id] {
			activities = append(activities, a)
			seen[a.Id] = true
		}
	}

	for _, a := range activities {
		ttl := time.Duration(a.EndTime-now+3600) * time.Second
		if ttl < time.Minute {
			ttl = time.Minute
		}

		// 缓存活动信息
		activityKey := fmt.Sprintf("seckill:activity:%d", a.Id)
		if err := svcCtx.Cache.Set(ctx, activityKey, a, ttl); err != nil {
			logx.Errorf("warmup: cache activity %d failed: %v", a.Id, err)
			continue
		}

		// 缓存库存（用 SetNX 避免覆盖已有的实时库存）
		stockKey := fmt.Sprintf("seckill:stock:%d", a.Id)
		_ = svcCtx.Cache.SetNX(ctx, stockKey, a.AvailableStock, ttl)

		logx.Infof("warmup: activity %d [%s] loaded, stock=%d", a.Id, a.Title, a.AvailableStock)
	}

	logx.Infof("warmup: %d activities loaded", len(activities))
}
