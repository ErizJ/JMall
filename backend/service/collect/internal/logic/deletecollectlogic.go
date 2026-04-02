// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/service/collect/internal/svc"
	"github.com/ErizJ/JMall/backend/service/collect/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCollectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCollectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCollectLogic {
	return &DeleteCollectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCollectLogic) DeleteCollect(req *types.DeleteCollectReq) (resp *types.DeleteCollectResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	if err := l.svcCtx.CollectModel.DeleteByUserAndProduct(l.ctx, userID, req.ProductID); err != nil {
		return nil, err
	}

	_ = l.svcCtx.Cache.Del(l.ctx, fmt.Sprintf("jmall:collect:user:%d", userID))

	return &types.DeleteCollectResp{Code: "200"}, nil
}
