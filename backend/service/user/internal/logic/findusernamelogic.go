// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"regexp"

	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/user/internal/svc"
	"github.com/ErizJ/JMall/backend/service/user/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type FindUserNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFindUserNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindUserNameLogic {
	return &FindUserNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FindUserNameLogic) FindUserName(req *types.CheckUserNameReq) (resp *types.CheckUserNameResp, err error) {
	userNameRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{4,15}$`)
	if !userNameRegex.MatchString(req.UserName) {
		return &types.CheckUserNameResp{Code: "002"}, nil
	}

	_, findErr := l.svcCtx.UsersModel.FindOneByUserName(l.ctx, req.UserName)
	if findErr == nil {
		// username exists — taken
		return &types.CheckUserNameResp{Code: "003"}, nil
	}
	if findErr != model.ErrNotFound {
		return nil, findErr
	}

	// not found — available
	return &types.CheckUserNameResp{Code: "200"}, nil
}
