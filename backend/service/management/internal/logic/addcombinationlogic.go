// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"database/sql"

	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/management/internal/svc"
	"github.com/ErizJ/JMall/backend/service/management/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddCombinationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddCombinationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCombinationLogic {
	return &AddCombinationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddCombinationLogic) AddCombination(req *types.AddCombinationReq) (resp *types.AddCombinationResp, err error) {
	if _, insertErr := l.svcCtx.CombinationProductModel.Insert(l.ctx, &model.CombinationProduct{
		MainProductId:       req.MainProductID,
		ViceProductId:       req.ViceProductID,
		AmountThreshold:     sql.NullInt64{Int64: int64(req.AmountThreshold), Valid: true},
		PriceReductionRange: sql.NullInt64{Int64: int64(req.PriceReductionRange), Valid: true},
	}); insertErr != nil {
		return nil, insertErr
	}
	return &types.AddCombinationResp{Code: "200"}, nil
}
