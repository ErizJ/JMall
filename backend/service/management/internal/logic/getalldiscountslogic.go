// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/ErizJ/JMall/backend/service/management/internal/svc"
	"github.com/ErizJ/JMall/backend/service/management/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllDiscountsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAllDiscountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllDiscountsLogic {
	return &GetAllDiscountsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAllDiscountsLogic) GetAllDiscounts() (resp *types.GetAllDiscountsResp, err error) {
	rows, err := l.svcCtx.CombinationProductModel.FindAll(l.ctx)
	if err != nil {
		return nil, err
	}

	combos := make([]types.CombinationItem, 0, len(rows))
	for _, row := range rows {
		combos = append(combos, types.CombinationItem{
			ID:                  row.Id,
			MainProductID:       row.MainProductId,
			ViceProductID:       row.ViceProductId,
			AmountThreshold:     float64(row.AmountThreshold.Int64),
			PriceReductionRange: float64(row.PriceReductionRange.Int64),
		})
	}

	return &types.GetAllDiscountsResp{
		Code:         "200",
		Combinations: combos,
	}, nil
}
