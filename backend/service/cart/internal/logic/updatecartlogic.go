// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"math"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/cart/internal/svc"
	"github.com/ErizJ/JMall/backend/service/cart/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCartLogic {
	return &UpdateCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateCartLogic) UpdateCart(req *types.UpdateCartReq) (resp *types.UpdateCartResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	if req.Num < 1 {
		return &types.UpdateCartResp{Code: "003"}, nil
	}

	item, err := l.svcCtx.ShoppingcartModel.FindByUserAndProduct(l.ctx, userID, req.ProductID)
	if err == model.ErrNotFound {
		return &types.UpdateCartResp{Code: "004"}, nil
	}
	if err != nil {
		return nil, err
	}

	if item.Num == req.Num {
		return &types.UpdateCartResp{Code: "003"}, nil
	}

	product, err := l.svcCtx.ProductModel.FindOne(l.ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	maxNum := int64(math.Floor(float64(product.ProductNum) / 2))
	if req.Num > maxNum {
		return &types.UpdateCartResp{Code: "003"}, nil
	}

	if err := l.svcCtx.ShoppingcartModel.UpdateNumByUserAndProduct(l.ctx, userID, req.ProductID, req.Num); err != nil {
		return nil, err
	}

	_ = l.svcCtx.Cache.Del(l.ctx, fmt.Sprintf("jmall:cart:user:%d", userID))

	return &types.UpdateCartResp{Code: "200"}, nil
}
