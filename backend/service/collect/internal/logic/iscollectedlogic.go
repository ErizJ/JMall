package logic

import (
	"context"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/collect/internal/svc"
	"github.com/ErizJ/JMall/backend/service/collect/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsCollectedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIsCollectedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsCollectedLogic {
	return &IsCollectedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IsCollectedLogic) IsCollected(req *types.IsCollectedReq) (resp *types.IsCollectedResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	_, err = l.svcCtx.CollectModel.FindByUserAndProduct(l.ctx, userID, req.ProductID)
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}

	return &types.IsCollectedResp{
		Code:        "200",
		IsCollected: err == nil,
	}, nil
}
