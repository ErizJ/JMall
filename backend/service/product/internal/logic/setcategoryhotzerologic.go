// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/ErizJ/JMall/backend/service/product/internal/svc"
	"github.com/ErizJ/JMall/backend/service/product/internal/types"
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

	// Invalidate all product and category caches
	_ = l.svcCtx.Cache.Del(l.ctx,
		"jmall:categories",
		"jmall:products:all",
		"jmall:products:hot:7",
		"jmall:products:promotion:7",
		"jmall:product:recommend:personal",
		"jmall:product:phone:7",
		"jmall:product:shell:7",
		"jmall:product:charger:7",
	)

	return &types.SetCategoryHotZeroResp{Code: "200"}, nil
}
