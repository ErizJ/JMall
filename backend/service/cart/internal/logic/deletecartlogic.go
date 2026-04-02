// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/service/cart/internal/svc"
	"github.com/ErizJ/JMall/backend/service/cart/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCartLogic {
	return &DeleteCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCartLogic) DeleteCart(req *types.DeleteCartReq) (resp *types.DeleteCartResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	if err := l.svcCtx.ShoppingcartModel.DeleteByUserAndProduct(l.ctx, userID, req.ProductID); err != nil {
		return nil, err
	}

	_ = l.svcCtx.Cache.Del(l.ctx, fmt.Sprintf("jmall:cart:user:%d", userID))

	return &types.DeleteCartResp{Code: "200"}, nil
}
