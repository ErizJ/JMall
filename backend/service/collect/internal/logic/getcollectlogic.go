// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/service/collect/internal/svc"
	"github.com/ErizJ/JMall/backend/service/collect/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCollectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCollectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCollectLogic {
	return &GetCollectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCollectLogic) GetCollect(req *types.GetCollectReq) (resp *types.GetCollectResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	cacheKey := fmt.Sprintf("jmall:collect:user:%d", userID)

	var collects []types.CollectItem
	if cacheErr := l.svcCtx.Cache.Get(l.ctx, cacheKey, &collects); cacheErr == nil {
		return &types.GetCollectResp{Code: "200", Collects: collects}, nil
	}

	rows, err := l.svcCtx.CollectModel.FindByUserId(l.ctx, userID)
	if err != nil {
		return nil, err
	}

	collects = make([]types.CollectItem, 0, len(rows))
	for _, row := range rows {
		collects = append(collects, types.CollectItem{
			ID:          row.Id,
			UserID:      row.UserId,
			ProductID:   row.ProductId,
			Category:    fmt.Sprintf("%d", row.Category),
			CollectTime: time.Unix(row.CollectTime, 0).Format("2006-01-02 15:04:05"),
		})
	}

	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, collects, 2*time.Minute)
	return &types.GetCollectResp{
		Code:     "200",
		Collects: collects,
	}, nil
}
