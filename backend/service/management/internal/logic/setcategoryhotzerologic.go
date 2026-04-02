package logic

import (
	"context"

	"github.com/ErizJ/JMall/backend/service/management/internal/svc"
	"github.com/ErizJ/JMall/backend/service/management/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type SetCategoryHotZeroLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetCategoryHotZeroLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetCategoryHotZeroLogic {
	return &SetCategoryHotZeroLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetCategoryHotZeroLogic) SetCategoryHotZero() (resp *types.SetCategoryHotZeroResp, err error) {
	if resetErr := l.svcCtx.CategoryModel.ResetAllCategoryHot(l.ctx); resetErr != nil {
		return nil, resetErr
	}
	return &types.SetCategoryHotZeroResp{Code: "200"}, nil
}
