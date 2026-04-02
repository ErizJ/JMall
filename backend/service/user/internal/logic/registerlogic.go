// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"database/sql"

	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/user/internal/svc"
	"github.com/ErizJ/JMall/backend/service/user/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	if !userNameRegex.MatchString(req.UserName) {
		return &types.RegisterResp{Code: "002"}, nil
	}

	if !passwordRegex.MatchString(req.Password) {
		return &types.RegisterResp{Code: "002"}, nil
	}

	_, findErr := l.svcCtx.UsersModel.FindOneByUserName(l.ctx, req.UserName)
	if findErr == nil {
		return &types.RegisterResp{Code: "003"}, nil
	}
	if findErr != model.ErrNotFound {
		return nil, findErr
	}

	hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if hashErr != nil {
		return nil, hashErr
	}

	newUser := &model.Users{
		UserName: req.UserName,
		Password: string(hashedPassword),
		UserPhoneNumber: sql.NullString{
			String: req.PhoneNumber,
			Valid:  req.PhoneNumber != "",
		},
	}

	_, insertErr := l.svcCtx.UsersModel.Insert(l.ctx, newUser)
	if insertErr != nil {
		return nil, insertErr
	}

	return &types.RegisterResp{Code: "200"}, nil
}
