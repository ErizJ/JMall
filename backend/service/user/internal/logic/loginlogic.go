// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"regexp"
	"time"

	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/user/internal/svc"
	"github.com/ErizJ/JMall/backend/service/user/internal/types"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

var (
	userNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{4,15}$`)
	passwordRegex = regexp.MustCompile(`^[a-zA-Z]\w{5,17}$`)
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	if !userNameRegex.MatchString(req.UserName) {
		return &types.LoginResp{Code: "002"}, nil
	}

	if !passwordRegex.MatchString(req.Password) {
		return &types.LoginResp{Code: "002"}, nil
	}

	user, err := l.svcCtx.UsersModel.FindOneByUserName(l.ctx, req.UserName)
	if err != nil {
		if err == model.ErrNotFound {
			return &types.LoginResp{Code: "002"}, nil
		}
		return nil, err
	}

	if bcryptErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); bcryptErr != nil {
		return &types.LoginResp{Code: "002"}, nil
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"userId":   user.UserId,
		"userName": user.UserName,
		"iat":      now.Unix(),
		"exp":      now.Add(time.Duration(l.svcCtx.Config.Auth.ExpireSeconds) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, tokenErr := token.SignedString([]byte(l.svcCtx.Config.Auth.Secret))
	if tokenErr != nil {
		return nil, tokenErr
	}

	return &types.LoginResp{
		Code:     "200",
		UserID:   user.UserId,
		UserName: user.UserName,
		Token:    tokenStr,
	}, nil
}
