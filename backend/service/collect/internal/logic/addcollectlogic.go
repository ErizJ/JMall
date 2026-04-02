// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/collect/internal/svc"
	"github.com/ErizJ/JMall/backend/service/collect/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddCollectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddCollectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCollectLogic {
	return &AddCollectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddCollectLogic) AddCollect(req *types.AddCollectReq) (resp *types.AddCollectResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	product, err := l.svcCtx.ProductModel.FindOne(l.ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	// Idempotency check: don't double-insert or double-increment hot scores
	existing, err := l.svcCtx.CollectModel.FindByUserAndProduct(l.ctx, userID, req.ProductID)
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}
	if existing != nil {
		return &types.AddCollectResp{Code: "200"}, nil
	}

	_ = l.svcCtx.CategoryModel.IncrCategoryHot(l.ctx, product.CategoryId)
	_ = l.svcCtx.ProductModel.IncrProductHot(l.ctx, req.ProductID)

	if _, insertErr := l.svcCtx.CollectModel.Insert(l.ctx, &model.Collect{
		UserId:      userID,
		ProductId:   req.ProductID,
		Category:    product.CategoryId,
		CollectTime: time.Now().Unix(),
	}); insertErr != nil {
		return nil, insertErr
	}

	_ = l.svcCtx.Cache.Del(l.ctx,
		fmt.Sprintf("jmall:collect:user:%d", userID),
		"jmall:products:hot:7",
		"jmall:product:recommend:personal",
	)

	return &types.AddCollectResp{Code: "200"}, nil
}
