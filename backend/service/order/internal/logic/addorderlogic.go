// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/order/internal/svc"
	"github.com/ErizJ/JMall/backend/service/order/internal/types"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddOrderLogic {
	return &AddOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddOrderLogic) AddOrder(req *types.AddOrderReq) (resp *types.AddOrderResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	if len(req.Items) == 0 {
		return &types.AddOrderResp{Code: "002"}, nil
	}

	// Safe order ID: millisecond timestamp * 1000 + random 3-digit suffix (stays well within int64)
	orderId := time.Now().UnixMilli()*1000 + int64(rand.Intn(1000))

	txErr := l.svcCtx.OrdersModel.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		txOrders := l.svcCtx.OrdersModel.WithSession(session)
		txCart := l.svcCtx.ShoppingcartModel.WithSession(session)

		for _, item := range req.Items {
			if _, insertErr := txOrders.Insert(ctx, &model.Orders{
				OrderId:      orderId,
				UserId:       userID,
				ProductId:    item.ProductID,
				ProductNum:   item.ProductNum,
				ProductPrice: item.ProductPrice,
				OrderTime:    time.Now().Unix(),
			}); insertErr != nil {
				return insertErr
			}
		}

		for _, item := range req.Items {
			if delErr := txCart.DeleteByUserAndProduct(ctx, userID, item.ProductID); delErr != nil {
				return delErr
			}
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	_ = l.svcCtx.Cache.Del(l.ctx,
		fmt.Sprintf("jmall:orders:user:%d", userID),
		fmt.Sprintf("jmall:cart:user:%d", userID),
	)

	return &types.AddOrderResp{Code: "200", OrderID: orderId}, nil
}
