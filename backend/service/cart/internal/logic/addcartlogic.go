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

type AddCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCartLogic {
	return &AddCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddCartLogic) AddCart(req *types.AddCartReq) (resp *types.AddCartResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	product, err := l.svcCtx.ProductModel.FindOne(l.ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	addNum := req.Num
	if addNum <= 0 {
		addNum = 1
	}

	existing, err := l.svcCtx.ShoppingcartModel.FindByUserAndProduct(l.ctx, userID, req.ProductID)
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}

	maxNum := int64(math.Floor(float64(product.ProductNum) / 2))

	if existing != nil {
		newNum := existing.Num + addNum
		if newNum > maxNum {
			return &types.AddCartResp{Code: "003"}, nil
		}
		if updateErr := l.svcCtx.ShoppingcartModel.UpdateNumByUserAndProduct(l.ctx, userID, req.ProductID, newNum); updateErr != nil {
			return nil, updateErr
		}
	} else {
		if addNum > maxNum {
			addNum = maxNum
		}
		if _, insertErr := l.svcCtx.ShoppingcartModel.Insert(l.ctx, &model.Shoppingcart{
			UserId:    userID,
			ProductId: req.ProductID,
			Num:       addNum,
		}); insertErr != nil {
			return nil, insertErr
		}
	}

	// Increment hot scores only after successfully modifying the cart
	_ = l.svcCtx.CategoryModel.IncrCategoryHot(l.ctx, product.CategoryId)
	_ = l.svcCtx.ProductModel.IncrProductHot(l.ctx, req.ProductID)

	_ = l.svcCtx.Cache.Del(l.ctx, fmt.Sprintf("jmall:cart:user:%d", userID))

	return &types.AddCartResp{Code: "200"}, nil
}
