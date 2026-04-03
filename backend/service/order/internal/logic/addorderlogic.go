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

// 库存扣减 Lua 脚本（原子操作）
// KEYS[1] = jmall:stock:{product_id}
// ARGV[1] = 扣减数量
// 返回值：1=成功 0=库存不足 -1=key不存在
//
// 为什么用 Lua 而不是先 GET 再 DECRBY？
// → GET + DECRBY 不是原子操作，高并发下两个请求同时 GET 到库存=1，都认为够用，都去扣减，导致超卖
// → Lua 脚本在 Redis 中是原子执行的，整个 check-and-decrement 不会被其他命令打断
const luaDecrStock = `
local stock = redis.call('GET', KEYS[1])
if stock == false then
    return -1
end
if tonumber(stock) < tonumber(ARGV[1]) then
    return 0
end
redis.call('DECRBY', KEYS[1], ARGV[1])
return 1
`

// 库存回滚 Lua 脚本
const luaIncrStock = `
local stock = redis.call('GET', KEYS[1])
if stock == false then
    return -1
end
redis.call('INCRBY', KEYS[1], ARGV[1])
return 1
`

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

// AddOrder 创建订单
//
// 主流实践对照：
//
//  1. 接口幂等（防重复提交）：
//     Redis SETNX jmall:order:submit:{user_id}，TTL=5s
//     用户快速双击只会创建一个订单
//
//  2. 金额服务端校验：
//     不信任前端传的 product_price，从数据库查询真实售价
//     防止前端篡改价格
//
//  3. Redis 预扣库存（Lua 原子操作）：
//     先在 Redis 中原子扣减库存，成功后再写数据库
//     失败则快速返回，不打数据库
//     数据库事务中再做最终扣减（双重保障）
//
//  4. 库存回滚：
//     如果数据库事务失败，回滚 Redis 中已扣减的库存
//
//  5. 订单超时关闭：
//     通过 Redis key TTL 实现延迟检查（简化版）
//     生产环境建议用 Kafka 延迟消息或定时任务扫描
func (l *AddOrderLogic) AddOrder(req *types.AddOrderReq) (resp *types.AddOrderResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	if len(req.Items) == 0 {
		return &types.AddOrderResp{Code: "002"}, nil
	}

	// ========== 1. 防重复提交（接口幂等） ==========
	// 同一用户 5 秒内只能提交一次订单
	submitKey := fmt.Sprintf("jmall:order:submit:%d", userID)
	if submitErr := l.svcCtx.Cache.SetNX(l.ctx, submitKey, "1", 5*time.Second); submitErr != nil {
		return &types.AddOrderResp{Code: "012"}, nil // 请勿重复提交
	}

	// ========== 2. 服务端校验商品价格 ==========
	// 批量查询商品信息，用数据库真实售价替代前端传入的价格
	productIDs := make([]int64, 0, len(req.Items))
	itemNumMap := make(map[int64]int64, len(req.Items))
	for _, item := range req.Items {
		productIDs = append(productIDs, item.ProductID)
		itemNumMap[item.ProductID] = item.ProductNum
	}

	products, prodErr := l.svcCtx.ProductModel.FindByIds(l.ctx, productIDs)
	if prodErr != nil {
		return nil, prodErr
	}
	if len(products) != len(req.Items) {
		return &types.AddOrderResp{Code: "013"}, nil // 部分商品不存在
	}

	// 构建商品信息 map（使用服务端真实价格）
	type productInfo struct {
		price float64
		stock int64
	}
	productMap := make(map[int64]*productInfo, len(products))
	for _, p := range products {
		productMap[p.ProductId] = &productInfo{
			price: p.ProductSellingPrice,
			stock: p.ProductNum,
		}
	}

	// ========== 3. Redis 预扣库存（Lua 原子操作） ==========
	deductedProducts := make([]int64, 0, len(req.Items))
	for _, item := range req.Items {
		stockKey := fmt.Sprintf("jmall:stock:%d", item.ProductID)

		p := productMap[item.ProductID]
		if p == nil {
			l.rollbackStock(deductedProducts, itemNumMap)
			return &types.AddOrderResp{Code: "013"}, nil
		}

		// 尝试 Lua 扣减
		ret, err := l.tryDecrStock(stockKey, item.ProductNum, p.stock)
		if err != nil {
			l.rollbackStock(deductedProducts, itemNumMap)
			return nil, err
		}
		if ret == 0 {
			l.rollbackStock(deductedProducts, itemNumMap)
			return &types.AddOrderResp{Code: "014"}, nil // 库存不足
		}

		deductedProducts = append(deductedProducts, item.ProductID)
	}

	// ========== 4. 生成订单号 + 数据库事务 ==========
	orderId := time.Now().UnixMilli()*1000 + int64(rand.Intn(1000))

	txErr := l.svcCtx.OrdersModel.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		txOrders := l.svcCtx.OrdersModel.WithSession(session)
		txCart := l.svcCtx.ShoppingcartModel.WithSession(session)
		txProduct := l.svcCtx.ProductModel.WithSession(session)

		for _, item := range req.Items {
			p := productMap[item.ProductID]

			// 写入订单（使用服务端真实价格，不信任前端）
			if _, insertErr := txOrders.Insert(ctx, &model.Orders{
				OrderId:      orderId,
				UserId:       userID,
				ProductId:    item.ProductID,
				ProductNum:   item.ProductNum,
				ProductPrice: p.price, // 服务端价格
				OrderTime:    time.Now().Unix(),
				Status:       0, // 待支付
			}); insertErr != nil {
				return insertErr
			}

			// 数据库扣减库存（最终一致性保障）
			if decrErr := txProduct.DecrStock(ctx, item.ProductID, item.ProductNum); decrErr != nil {
				return decrErr
			}
		}

		// 清空购物车中对应商品
		for _, item := range req.Items {
			if delErr := txCart.DeleteByUserAndProduct(ctx, userID, item.ProductID); delErr != nil {
				return delErr
			}
		}
		return nil
	})
	if txErr != nil {
		// 数据库事务失败，回滚 Redis 库存
		l.rollbackStock(deductedProducts, itemNumMap)
		return nil, txErr
	}

	// ========== 5. 设置订单超时关闭标记 ==========
	// 30 分钟后如果还是待支付状态，应该关闭订单并回滚库存
	// 这里用 Redis key 做标记，配合定时任务或延迟队列检查
	expireKey := fmt.Sprintf("jmall:order:expire:%d", orderId)
	_ = l.svcCtx.Cache.Set(l.ctx, expireKey, "1", 30*time.Minute)

	// 清理缓存
	_ = l.svcCtx.Cache.Del(l.ctx,
		fmt.Sprintf("jmall:orders:user:%d", userID),
		fmt.Sprintf("jmall:cart:user:%d", userID),
	)

	return &types.AddOrderResp{Code: "200", OrderID: orderId}, nil
}

// rollbackStock 回滚 Redis 中已扣减的库存
func (l *AddOrderLogic) rollbackStock(productIDs []int64, numMap map[int64]int64) {
	for _, pid := range productIDs {
		stockKey := fmt.Sprintf("jmall:stock:%d", pid)
		_, _ = l.svcCtx.Cache.Eval(l.ctx, luaIncrStock, []string{stockKey}, numMap[pid])
	}
}

// tryDecrStock 尝试通过 Lua 原子扣减 Redis 库存
// 返回值：1=成功 0=库存不足
// 如果 Redis 中没有库存 key，从 DB 加载后重试一次
func (l *AddOrderLogic) tryDecrStock(stockKey string, num int64, dbStock int64) (int64, error) {
	result, err := l.svcCtx.Cache.Eval(l.ctx, luaDecrStock, []string{stockKey}, num)
	if err != nil {
		// Redis 错误（非 Lua 返回值），尝试初始化后重试
		_ = l.svcCtx.Cache.Set(l.ctx, stockKey, dbStock, 10*time.Minute)
		result, err = l.svcCtx.Cache.Eval(l.ctx, luaDecrStock, []string{stockKey}, num)
		if err != nil {
			return -1, err
		}
	}

	ret, _ := result.(int64)
	if ret == -1 {
		// Lua 返回 -1 表示 key 不存在，从 DB 加载后重试
		_ = l.svcCtx.Cache.Set(l.ctx, stockKey, dbStock, 10*time.Minute)
		result, err = l.svcCtx.Cache.Eval(l.ctx, luaDecrStock, []string{stockKey}, num)
		if err != nil {
			return -1, err
		}
		ret, _ = result.(int64)
		if ret == -1 {
			// 仍然失败，当作库存不足处理
			return 0, nil
		}
	}

	return ret, nil
}
