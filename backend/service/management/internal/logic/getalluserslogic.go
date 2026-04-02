// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/ErizJ/JMall/backend/service/management/internal/svc"
	"github.com/ErizJ/JMall/backend/service/management/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAllUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllUsersLogic {
	return &GetAllUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAllUsersLogic) GetAllUsers() (resp *types.GetAllUsersResp, err error) {
	rows, err := l.svcCtx.UsersModel.FindAll(l.ctx)
	if err != nil {
		return nil, err
	}

	users := make([]types.MgmtUserItem, 0, len(rows))
	for _, row := range rows {
		users = append(users, types.MgmtUserItem{
			UserID:      row.UserId,
			UserName:    row.UserName,
			PhoneNumber: row.UserPhoneNumber.String,
		})
	}

	return &types.GetAllUsersResp{
		Code:  "200",
		Users: users,
	}, nil
}
