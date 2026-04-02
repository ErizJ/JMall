// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/cart/internal/svc"
	"github.com/ErizJ/JMall/backend/service/cart/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsExistCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIsExistCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsExistCartLogic {
	return &IsExistCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IsExistCartLogic) IsExistCart(req *types.IsExistCartReq) (resp *types.IsExistCartResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	_, err = l.svcCtx.ShoppingcartModel.FindByUserAndProduct(l.ctx, userID, req.ProductID)
	if err != nil {
		if err == model.ErrNotFound {
			return &types.IsExistCartResp{Code: "200", IsExist: false}, nil
		}
		return nil, err
	}
	return &types.IsExistCartResp{Code: "200", IsExist: true}, nil
}
