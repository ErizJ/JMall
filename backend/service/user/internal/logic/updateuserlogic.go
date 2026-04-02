// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/user/internal/svc"
	"github.com/ErizJ/JMall/backend/service/user/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UpdateUserReq) (resp *types.UpdateUserResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	user, findErr := l.svcCtx.UsersModel.FindOne(l.ctx, userID)
	if findErr != nil {
		if findErr == model.ErrNotFound {
			return &types.UpdateUserResp{Code: "002"}, nil
		}
		return nil, findErr
	}

	if req.UserName != "" {
		user.UserName = req.UserName
	}
	if req.PhoneNumber != "" {
		user.UserPhoneNumber = sql.NullString{
			String: req.PhoneNumber,
			Valid:  true,
		}
	}

	if updateErr := l.svcCtx.UsersModel.Update(l.ctx, user); updateErr != nil {
		return nil, updateErr
	}

	_ = l.svcCtx.Cache.Del(l.ctx, fmt.Sprintf("jmall:user:detail:%d", userID))

	return &types.UpdateUserResp{Code: "200"}, nil
}
