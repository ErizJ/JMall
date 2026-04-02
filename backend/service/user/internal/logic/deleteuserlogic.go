// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/service/user/internal/svc"
	"github.com/ErizJ/JMall/backend/service/user/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteUserLogic) DeleteUser(req *types.DeleteUserReq) (resp *types.DeleteUserResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	if deleteErr := l.svcCtx.UsersModel.Delete(l.ctx, userID); deleteErr != nil {
		return nil, deleteErr
	}

	_ = l.svcCtx.Cache.Del(l.ctx, fmt.Sprintf("jmall:user:detail:%d", userID))

	return &types.DeleteUserResp{Code: "200"}, nil
}
