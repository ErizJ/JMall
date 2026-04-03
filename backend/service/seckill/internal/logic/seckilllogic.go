package logic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/svc"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/types"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

// Lua 脚本：原子扣库存 + 限购校验（一次 RTT 完成两个检查）
//
// KEYS[1] = seckill:stock:{activity_id}       — 库存
// KEYS[2] = seckill:bought:{activity_id}:{uid} — 用户已购数量
// ARGV[1] = 本次购买数量
// ARGV[2] = 限购上限
//
// 返回: 1=成功, 0=库存不足, -1=超出限购
const luaSeckillDecrStock = `
local bought = tonumber(redis.call('GET', KEYS[2]) or '0')
if bought + tonumber(ARGV[1]) > tonumber(ARGV[2]) then
    return -1
end
local stock = tonumber(redis.call('GET', KEYS[1]))
if stock == nil or stock < tonumber(ARGV[1]) then
    return 0
end
redis.call('DECRBY', KEYS[1], ARGV[1])
redis.call('INCRBY', KEYS[2], ARGV[1])
if bought == 0 then
    redis.call('EXPIRE', KEYS[2], 86400)
end
return 1
`

// Lua 脚本：回滚库存 + 限购计数
const luaSeckillRollback = `
redis.call('INCRBY', KEYS[1], ARGV[1])
local bought = tonumber(redis.call('GET', KEYS[2]) or '0')
if bought > 0 then
    redis.call('DECRBY', KEYS[2], ARGV[1])
end
return 1
`

// 本地售罄标记（进程内缓存，售罄后不再访问 Redis）
var (
	soldOutMap sync.Map // map[int64]bool
)

type SeckillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSeckillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SeckillLogic {
	return &SeckillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Seckill 秒杀抢购核心逻辑
//
// 流程：
//  1. 本地售罄快速拦截（零网络开销）
//  2. Redis 读取活动信息 + 时间窗口校验
//  3. Redis Lua 原子操作：扣库存 + 限购校验
//  4. 生成排队令牌，写入 Redis
//  5. 投递 Kafka 异步下单
//  6. 返回令牌，前端轮询结果
func (l *SeckillLogic) Seckill(req *types.SeckillReq) (resp *types.SeckillResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	// ========== 1. 本地售罄快速拦截 ==========
	if v, ok := soldOutMap.Load(req.ActivityID); ok && v.(bool) {
		return &types.SeckillResp{Code: "SOLD_OUT", Msg: "已售罄"}, nil
	}

	// ========== 2. 活动校验（优先 Redis 缓存） ==========
	activity, actErr := l.getActivity(req.ActivityID)
	if actErr != nil {
		return &types.SeckillResp{Code: "ACTIVITY_NOT_FOUND", Msg: "活动不存在"}, nil
	}

	now := time.Now().Unix()
	if now < activity.StartTime {
		return &types.SeckillResp{Code: "NOT_STARTED", Msg: "活动未开始"}, nil
	}
	if now > activity.EndTime {
		return &types.SeckillResp{Code: "ENDED", Msg: "活动已结束"}, nil
	}

	// ========== 3. Redis Lua 原子操作：扣库存 + 限购 ==========
	stockKey := fmt.Sprintf("seckill:stock:%d", req.ActivityID)
	boughtKey := fmt.Sprintf("seckill:bought:%d:%d", req.ActivityID, userID)

	result, evalErr := l.svcCtx.Cache.Eval(l.ctx, luaSeckillDecrStock,
		[]string{stockKey, boughtKey},
		1,                      // 购买数量
		activity.LimitPerUser,  // 限购上限
	)
	if evalErr != nil {
		l.Errorf("seckill lua eval error: %v", evalErr)
		return &types.SeckillResp{Code: "SYSTEM_ERROR", Msg: "系统繁忙，请重试"}, nil
	}

	ret, _ := result.(int64)
	switch ret {
	case -1:
		return &types.SeckillResp{Code: "LIMIT_EXCEEDED", Msg: fmt.Sprintf("每人限购%d件", activity.LimitPerUser)}, nil
	case 0:
		// 售罄 → 设置本地标记，后续请求直接拦截
		soldOutMap.Store(req.ActivityID, true)
		return &types.SeckillResp{Code: "SOLD_OUT", Msg: "已售罄"}, nil
	}

	// ========== 4. 生成排队令牌 ==========
	token := uuid.New().String()
	tokenKey := fmt.Sprintf("seckill:token:%s", token)
	tokenData := types.SeckillMessage{
		Token:        token,
		UserID:       userID,
		ActivityID:   req.ActivityID,
		ProductID:    activity.ProductId,
		SeckillPrice: activity.SeckillPrice,
		Num:          1,
		Timestamp:    now,
	}
	if setErr := l.svcCtx.Cache.Set(l.ctx, tokenKey, tokenData, 5*time.Minute); setErr != nil {
		l.rollbackStock(req.ActivityID, userID, 1)
		return nil, setErr
	}

	// ========== 5. 投递 Kafka 异步下单 ==========
	if sendErr := l.svcCtx.KafkaProducer.Send(l.ctx, fmt.Sprintf("%d", req.ActivityID), tokenData); sendErr != nil {
		l.Errorf("kafka send error: %v", sendErr)
		l.rollbackStock(req.ActivityID, userID, 1)
		_ = l.svcCtx.Cache.Del(l.ctx, tokenKey)
		return &types.SeckillResp{Code: "SYSTEM_ERROR", Msg: "系统繁忙，请重试"}, nil
	}

	// ========== 6. 返回排队令牌 ==========
	return &types.SeckillResp{
		Code:  "200",
		Token: token,
		Msg:   "排队中，请稍候查询结果",
	}, nil
}

// getActivity 从 Redis 缓存获取活动信息，miss 时回源 DB
func (l *SeckillLogic) getActivity(activityID int64) (*model.SeckillActivity, error) {
	cacheKey := fmt.Sprintf("seckill:activity:%d", activityID)
	var activity model.SeckillActivity
	if err := l.svcCtx.Cache.Get(l.ctx, cacheKey, &activity); err == nil {
		return &activity, nil
	}

	// Cache miss → 查 DB
	dbActivity, dbErr := l.svcCtx.SeckillActivityModel.FindOne(l.ctx, activityID)
	if dbErr != nil {
		return nil, dbErr
	}

	// 写回缓存，TTL = 活动结束后 1 小时
	ttl := time.Duration(dbActivity.EndTime-time.Now().Unix()+3600) * time.Second
	if ttl < time.Minute {
		ttl = time.Minute
	}
	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, dbActivity, ttl)

	return dbActivity, nil
}

// rollbackStock 回滚 Redis 中的库存和限购计数
func (l *SeckillLogic) rollbackStock(activityID, userID, num int64) {
	stockKey := fmt.Sprintf("seckill:stock:%d", activityID)
	boughtKey := fmt.Sprintf("seckill:bought:%d:%d", activityID, userID)
	_, err := l.svcCtx.Cache.Eval(l.ctx, luaSeckillRollback,
		[]string{stockKey, boughtKey},
		num,
	)
	if err != nil {
		l.Errorf("rollback seckill stock failed: activity=%d, user=%d, err=%v", activityID, userID, err)
	}
	// 清除本地售罄标记（库存回来了）
	soldOutMap.Delete(activityID)
}
