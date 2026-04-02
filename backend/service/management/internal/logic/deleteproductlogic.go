// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"

	"github.com/ErizJ/JMall/backend/service/management/internal/svc"
	"github.com/ErizJ/JMall/backend/service/management/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductLogic {
	return &DeleteProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteProductLogic) DeleteProduct(req *types.DeleteProductReq) (resp *types.DeleteProductResp, err error) {
	if err := l.svcCtx.ProductModel.Delete(l.ctx, req.ProductID); err != nil {
		return nil, err
	}

	// Invalidate product caches
	_ = l.svcCtx.Cache.Del(l.ctx,
		fmt.Sprintf("jmall:product:detail:%d", req.ProductID),
		fmt.Sprintf("jmall:product:pictures:%d", req.ProductID),
		"jmall:products:all",
		"jmall:products:hot:7",
		"jmall:products:promotion:7",
		"jmall:product:phone:7",
		"jmall:product:shell:7",
		"jmall:product:charger:7",
	)

	return &types.DeleteProductResp{Code: "200"}, nil
}
