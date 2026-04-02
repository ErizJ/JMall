// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/ErizJ/JMall/backend/service/management/internal/svc"
	"github.com/ErizJ/JMall/backend/service/management/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCombinationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCombinationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCombinationLogic {
	return &DeleteCombinationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCombinationLogic) DeleteCombination(req *types.DeleteCombinationReq) (resp *types.DeleteCombinationResp, err error) {
	if err := l.svcCtx.CombinationProductModel.Delete(l.ctx, req.ID); err != nil {
		return nil, err
	}
	return &types.DeleteCombinationResp{Code: "200"}, nil
}
