package logic

import (
	"context"
	"fmt"

	"github.com/ErizJ/JMall/backend/service/seckill/internal/svc"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type SeckillResultLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSeckillResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SeckillResultLogic {
	return &SeckillResultLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SeckillResult 查询秒杀下单结果（前端轮询调用）
//
// Redis key: seckill:result:{token}
// 值: {"status": 0|1|2, "order_id": xxx, "msg": "..."}
func (l *SeckillResultLogic) SeckillResult(req *types.SeckillResultReq) (resp *types.SeckillResultResp, err error) {
	if req.Token == "" {
		return &types.SeckillResultResp{Code: "002", Msg: "token 不能为空"}, nil
	}

	resultKey := fmt.Sprintf("seckill:result:%s", req.Token)
	var result struct {
		Status  int    `json:"status"`
		OrderID int64  `json:"order_id"`
		Msg     string `json:"msg"`
	}

	if getErr := l.svcCtx.Cache.Get(l.ctx, resultKey, &result); getErr != nil {
		// 没有结果 → 还在排队
		return &types.SeckillResultResp{
			Code:   "200",
			Status: 0,
			Msg:    "排队中，请稍候",
		}, nil
	}

	return &types.SeckillResultResp{
		Code:    "200",
		Status:  result.Status,
		OrderID: result.OrderID,
		Msg:     result.Msg,
	}, nil
}
