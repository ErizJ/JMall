package logic

import (
	"context"
	"errors"
	"fmt"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/service/order/internal/svc"
	"github.com/ErizJ/JMall/backend/service/order/internal/types"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOrderLogic {
	return &DeleteOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteOrderLogic) DeleteOrder(req *types.DeleteOrderReq) (resp *types.DeleteOrderResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	rows, err := l.svcCtx.OrdersModel.FindByOrderId(l.ctx, req.OrderID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return &types.DeleteOrderResp{Code: "002"}, nil
	}
	if rows[0].UserId != userID {
		return nil, errors.New("forbidden")
	}

	// 已支付订单不可删除，需要先退款
	if rows[0].Status == 1 {
		return &types.DeleteOrderResp{Code: "005"}, nil
	}

	// 待支付(0)的订单删除时需要回滚库存
	// 已取消(2)、已退款(3)的库存已经在对应流程中回滚过了
	needRollbackStock := rows[0].Status == 0

	if needRollbackStock {
		// 事务内：回滚库存 + 删除订单
		txErr := l.svcCtx.OrdersModel.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			txProduct := l.svcCtx.ProductModel.WithSession(session)
			for _, item := range rows {
				if incrErr := txProduct.IncrStock(ctx, item.ProductId, item.ProductNum); incrErr != nil {
					return incrErr
				}
			}

			txOrders := l.svcCtx.OrdersModel.WithSession(session)
			if delErr := txOrders.DeleteByOrderId(ctx, req.OrderID); delErr != nil {
				return delErr
			}
			return nil
		})
		if txErr != nil {
			return nil, txErr
		}

		// 清理 Redis 库存缓存
		for _, item := range rows {
			_ = l.svcCtx.Cache.Del(l.ctx, fmt.Sprintf("jmall:stock:%d", item.ProductId))
		}
		// 清理防重支付锁（如果有的话）
		_ = l.svcCtx.Cache.Del(l.ctx, fmt.Sprintf("jmall:payment:lock:%d", req.OrderID))
	} else {
		if err := l.svcCtx.OrdersModel.DeleteByOrderId(l.ctx, req.OrderID); err != nil {
			return nil, err
		}
	}

	_ = l.svcCtx.Cache.Del(l.ctx, fmt.Sprintf("jmall:orders:user:%d", userID))

	return &types.DeleteOrderResp{Code: "200"}, nil
}
