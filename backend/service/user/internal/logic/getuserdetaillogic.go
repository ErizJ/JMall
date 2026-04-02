// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/user/internal/svc"
	"github.com/ErizJ/JMall/backend/service/user/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserDetailLogic {
	return &GetUserDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserDetailLogic) GetUserDetail(req *types.GetUserDetailReq) (resp *types.GetUserDetailResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	cacheKey := fmt.Sprintf("jmall:user:detail:%d", userID)
	var cached types.GetUserDetailResp
	if cacheErr := l.svcCtx.Cache.Get(l.ctx, cacheKey, &cached); cacheErr == nil {
		return &cached, nil
	}

	user, findErr := l.svcCtx.UsersModel.FindOne(l.ctx, userID)
	if findErr != nil {
		if findErr == model.ErrNotFound {
			return &types.GetUserDetailResp{Code: "002"}, nil
		}
		return nil, findErr
	}

	phoneNumber := ""
	if user.UserPhoneNumber.Valid {
		phoneNumber = user.UserPhoneNumber.String
	}

	result := &types.GetUserDetailResp{
		Code:        "200",
		UserID:      user.UserId,
		UserName:    user.UserName,
		PhoneNumber: phoneNumber,
	}
	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, result, 5*time.Minute)
	return result, nil
}
