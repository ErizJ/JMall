package logic

import (
	"context"
	"time"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/recommendation/internal/svc"
	"github.com/ErizJ/JMall/backend/service/recommendation/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ReportBehaviorLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReportBehaviorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportBehaviorLogic {
	return &ReportBehaviorLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReportBehaviorLogic) ReportBehavior(req *types.ReportBehaviorReq) (*types.ReportBehaviorResp, error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	// 校验行为类型
	if req.BehaviorType < model.BehaviorView || req.BehaviorType > model.BehaviorCollect {
		return &types.ReportBehaviorResp{Code: "002"}, nil
	}

	behavior := &model.UserBehavior{
		UserId:       userID,
		ProductId:    req.ProductID,
		CategoryId:   req.CategoryID,
		BehaviorType: req.BehaviorType,
		BehaviorTime: time.Now().UnixMilli(),
	}

	if _, err := l.svcCtx.UserBehaviorModel.Insert(l.ctx, behavior); err != nil {
		l.Logger.Errorf("ReportBehavior insert error: %v", err)
		// 行为上报失败不影响用户体验，静默处理
		return &types.ReportBehaviorResp{Code: "200"}, nil
	}

	return &types.ReportBehaviorResp{Code: "200"}, nil
}
