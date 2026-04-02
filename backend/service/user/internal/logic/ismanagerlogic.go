// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/user/internal/svc"
	"github.com/ErizJ/JMall/backend/service/user/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type IsManagerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIsManagerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsManagerLogic {
	return &IsManagerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IsManagerLogic) IsManager(req *types.IsManagerReq) (resp *types.IsManagerResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	user, userErr := l.svcCtx.UsersModel.FindOne(l.ctx, userID)
	if userErr != nil {
		if userErr == model.ErrNotFound {
			return &types.IsManagerResp{Code: "200", IsManager: false}, nil
		}
		return nil, userErr
	}

	_, sysErr := l.svcCtx.SysManagerModel.FindOneBySysname(l.ctx, user.UserName)
	if sysErr == nil {
		return &types.IsManagerResp{Code: "200", IsManager: true}, nil
	}
	if sysErr == model.ErrNotFound {
		return &types.IsManagerResp{Code: "200", IsManager: false}, nil
	}

	return nil, sysErr
}
