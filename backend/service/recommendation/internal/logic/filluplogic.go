package logic

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/recommendation/internal/svc"
	"github.com/ErizJ/JMall/backend/service/recommendation/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// 满减规则档位（从 combination_product 表提取的去重规则）
type promotionTier struct {
	Threshold float64
	Reduction float64
}

// 带评分的候选商品
type scoredProduct struct {
	product *model.Product
	reason  string
	score   float64
}

type FillUpLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFillUpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FillUpLogic {
	return &FillUpLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FillUpLogic) FillUp(req *types.FillUpReq) (resp *types.FillUpResp, err error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	// ========== 1. 获取购物车，计算总价 ==========
	cartItems, err := l.svcCtx.ShoppingcartModel.FindByUserId(l.ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(cartItems) == 0 {
		return &types.FillUpResp{Code: "200", Recommendations: []types.RecommendItem{}}, nil
	}

	// 获取购物车中所有商品详情
	cartProductIds := make([]int64, 0, len(cartItems))
	cartProductMap := make(map[int64]int64) // productId -> num
	for _, ci := range cartItems {
		cartProductIds = append(cartProductIds, ci.ProductId)
		cartProductMap[ci.ProductId] = ci.Num
	}
	cartProducts, err := l.svcCtx.ProductModel.FindByIds(l.ctx, cartProductIds)
	if err != nil {
		return nil, err
	}

	// 计算购物车总价（使用售价）
	var cartTotal float64
	cartCategoryIds := make(map[int64]bool)
	for _, p := range cartProducts {
		num := cartProductMap[p.ProductId]
		cartTotal += p.ProductSellingPrice * float64(num)
		cartCategoryIds[p.CategoryId] = true
	}

	// ========== 2. 获取满减规则，找到最近的未达标档位 ==========
	tiers := l.getPromotionTiers()
	nearestTier, gap := l.findNearestTier(tiers, cartTotal)

	if nearestTier == nil {
		// 已达到所有满减档位，无需凑单
		return &types.FillUpResp{
			Code:            "200",
			CartTotal:       cartTotal,
			NearestRule:     types.PromotionRule{Threshold: 0, Reduction: 0},
			Gap:             0,
			Recommendations: []types.RecommendItem{},
		}, nil
	}

	// ========== 3. 尝试读缓存 ==========
	cacheKey := fmt.Sprintf("jmall:recommend:fillup:%d:%.0f", userID, cartTotal)
	var cached []types.RecommendItem
	if cacheErr := l.svcCtx.Cache.Get(l.ctx, cacheKey, &cached); cacheErr == nil {
		return &types.FillUpResp{
			Code:            "200",
			CartTotal:       cartTotal,
			NearestRule:     types.PromotionRule{Threshold: nearestTier.Threshold, Reduction: nearestTier.Reduction},
			Gap:             gap,
			Recommendations: cached,
		}, nil
	}

	// ========== 4. 三策略并行收集候选商品 ==========
	excludeIds := cartProductIds
	candidates := make([]scoredProduct, 0, 60)

	// 策略1: 差额精准推荐 — 价格在 [gap*0.5, gap*1.5] 区间
	candidates = append(candidates, l.strategyPriceGap(gap, excludeIds)...)

	// 策略2: 关联商品推荐 — 同分类 + combination_product 表的搭配
	categoryIds := make([]int64, 0, len(cartCategoryIds))
	for cid := range cartCategoryIds {
		categoryIds = append(categoryIds, cid)
	}
	candidates = append(candidates, l.strategyAssociated(cartProductIds, categoryIds, excludeIds)...)

	// 策略3: 热销商品推荐 — 全站热销
	candidates = append(candidates, l.strategyHotSelling(excludeIds)...)

	// ========== 5. 去重 + 综合评分排序 ==========
	results := l.deduplicateAndRank(candidates, gap, nearestTier.Threshold)

	// 限制返回数量
	const maxResults = 12
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	// ========== 6. 写缓存 ==========
	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, results, 2*time.Minute)

	return &types.FillUpResp{
		Code:            "200",
		CartTotal:       cartTotal,
		NearestRule:     types.PromotionRule{Threshold: nearestTier.Threshold, Reduction: nearestTier.Reduction},
		Gap:             gap,
		Recommendations: results,
	}, nil
}

// ==================== 满减规则解析 ====================

func (l *FillUpLogic) getPromotionTiers() []promotionTier {
	// 优化：满减规则是准静态数据，缓存 10 分钟避免每次请求都查 DB
	const tiersCacheKey = "jmall:promotion:tiers"
	var cached []promotionTier
	if err := l.svcCtx.Cache.Get(l.ctx, tiersCacheKey, &cached); err == nil && len(cached) > 0 {
		return cached
	}

	combos, err := l.svcCtx.CombinationProductModel.FindAll(l.ctx)
	if err != nil || len(combos) == 0 {
		// 兜底默认规则
		return []promotionTier{
			{Threshold: 2000, Reduction: 200},
			{Threshold: 3000, Reduction: 300},
		}
	}

	// 从 combination_product 表提取去重的满减档位
	tierMap := make(map[float64]float64)
	for _, c := range combos {
		if c.AmountThreshold.Valid && c.PriceReductionRange.Valid {
			threshold := float64(c.AmountThreshold.Int64)
			reduction := float64(c.PriceReductionRange.Int64)
			if existing, ok := tierMap[threshold]; !ok || reduction > existing {
				tierMap[threshold] = reduction
			}
		}
	}

	tiers := make([]promotionTier, 0, len(tierMap))
	for t, r := range tierMap {
		tiers = append(tiers, promotionTier{Threshold: t, Reduction: r})
	}
	sort.Slice(tiers, func(i, j int) bool {
		return tiers[i].Threshold < tiers[j].Threshold
	})

	if len(tiers) == 0 {
		return []promotionTier{
			{Threshold: 2000, Reduction: 200},
			{Threshold: 3000, Reduction: 300},
		}
	}

	// 写入缓存，10 分钟过期
	_ = l.svcCtx.Cache.Set(l.ctx, tiersCacheKey, tiers, 10*time.Minute)
	return tiers
}

// findNearestTier 找到用户最接近但未达到的满减档位
func (l *FillUpLogic) findNearestTier(tiers []promotionTier, cartTotal float64) (*promotionTier, float64) {
	for i := range tiers {
		if cartTotal < tiers[i].Threshold {
			gap := tiers[i].Threshold - cartTotal
			return &tiers[i], gap
		}
	}
	return nil, 0 // 已达到所有档位
}

// ==================== 策略1: 差额精准推荐 ====================
// 核心思路：推荐价格在 [gap*0.5, gap*1.5] 区间的商品
// 让用户加一件就能凑到满减门槛，避免过度消费

func (l *FillUpLogic) strategyPriceGap(gap float64, excludeIds []int64) []scoredProduct {
	minPrice := gap * 0.5
	maxPrice := gap * 1.5
	// 至少从1元开始
	if minPrice < 1 {
		minPrice = 1
	}

	products, err := l.svcCtx.ProductModel.FindByPriceRange(l.ctx, minPrice, maxPrice, excludeIds, 20)
	if err != nil {
		l.Logger.Errorf("strategyPriceGap error: %v", err)
		return nil
	}

	results := make([]scoredProduct, 0, len(products))
	for _, p := range products {
		// 评分：价格越接近gap，分数越高（满分100）
		priceDiff := math.Abs(p.ProductSellingPrice - gap)
		priceScore := math.Max(0, 100-priceDiff/gap*100)
		results = append(results, scoredProduct{
			product: p,
			reason:  "差额精准推荐",
			score:   priceScore,
		})
	}
	return results
}

// ==================== 策略2: 关联商品推荐 ====================
// 核心思路：
//   a) 从 combination_product 表找购物车商品的搭配商品
//   b) 推荐购物车同分类下的其他商品（买手机推手机壳/充电器等关联品类）
// 关联品类映射：手机(1) → 保护套(5), 保护膜(6), 充电器(7), 充电宝(8)

var relatedCategoryMap = map[int64][]int64{
	1: {5, 6, 7, 8}, // 手机 → 配件
	2: {},            // 电视
	3: {},            // 空调
	4: {},            // 洗衣机
	5: {1},           // 保护套 → 手机
	6: {1},           // 保护膜 → 手机
	7: {1},           // 充电器 → 手机
	8: {1},           // 充电宝 → 手机
}

func (l *FillUpLogic) strategyAssociated(cartProductIds, cartCategoryIds, excludeIds []int64) []scoredProduct {
	results := make([]scoredProduct, 0, 20)
	excludeSet := make(map[int64]bool, len(excludeIds))
	for _, id := range excludeIds {
		excludeSet[id] = true
	}

	// a) combination_product 表的搭配商品 — 批量查询避免 N+1
	// 先收集所有需要的 vice_product_id，再一次性 FindByIds
	viceProductIds := make([]int64, 0, 10)
	for _, pid := range cartProductIds {
		combos, err := l.svcCtx.CombinationProductModel.FindByMainProductId(l.ctx, pid)
		if err != nil {
			continue
		}
		for _, c := range combos {
			if !excludeSet[c.ViceProductId] {
				viceProductIds = append(viceProductIds, c.ViceProductId)
			}
		}
	}
	if len(viceProductIds) > 0 {
		viceProducts, err := l.svcCtx.ProductModel.FindByIds(l.ctx, viceProductIds)
		if err == nil {
			for _, p := range viceProducts {
				if p.ProductNum <= 0 {
					continue
				}
				results = append(results, scoredProduct{
					product: p,
					reason:  "搭配购推荐",
					score:   90, // 搭配购给高分
				})
			}
		}
	}

	// b) 关联品类推荐
	relatedCatIds := make(map[int64]bool)
	for _, cid := range cartCategoryIds {
		if related, ok := relatedCategoryMap[cid]; ok {
			for _, rcid := range related {
				relatedCatIds[rcid] = true
			}
		}
	}
	if len(relatedCatIds) > 0 {
		catIds := make([]int64, 0, len(relatedCatIds))
		for cid := range relatedCatIds {
			catIds = append(catIds, cid)
		}
		products, err := l.svcCtx.ProductModel.FindByCategoryIds(l.ctx, catIds, excludeIds, 15)
		if err == nil {
			for _, p := range products {
				hot := int64(0)
				if p.ProductHot.Valid {
					hot = p.ProductHot.Int64
				}
				results = append(results, scoredProduct{
					product: p,
					reason:  "关联配件推荐",
					score:   70 + float64(hot)*0.5, // 基础70分 + 热度加成
				})
			}
		}
	}

	return results
}

// ==================== 策略3: 热销商品推荐 ====================
// 核心思路：全站热销 Top N，作为兜底策略保证推荐列表不为空

func (l *FillUpLogic) strategyHotSelling(excludeIds []int64) []scoredProduct {
	products, err := l.svcCtx.ProductModel.FindTopHot(l.ctx, 20)
	if err != nil {
		l.Logger.Errorf("strategyHotSelling error: %v", err)
		return nil
	}

	results := make([]scoredProduct, 0, len(products))
	for _, p := range products {
		if contains(excludeIds, p.ProductId) || p.ProductNum <= 0 {
			continue
		}
		hot := int64(0)
		if p.ProductHot.Valid {
			hot = p.ProductHot.Int64
		}
		results = append(results, scoredProduct{
			product: p,
			reason:  "热销推荐",
			score:   50 + float64(hot)*0.3, // 基础50分 + 热度加成
		})
	}
	return results
}

// ==================== 去重 + 综合评分排序 ====================
// 综合评分 = 策略基础分 * 0.4 + 价格匹配分 * 0.4 + 热度分 * 0.2
// 价格匹配分：商品价格越接近差额，分数越高
// 热度分：归一化到 0-100

func (l *FillUpLogic) deduplicateAndRank(candidates []scoredProduct, gap, threshold float64) []types.RecommendItem {
	// 优化：用 map 索引替代内层循环，O(n²) → O(n)
	indexMap := make(map[int64]int, len(candidates)) // productId -> index in unique
	unique := make([]scoredProduct, 0, len(candidates))

	for _, c := range candidates {
		if idx, exists := indexMap[c.product.ProductId]; exists {
			// 保留分数更高的
			if c.score > unique[idx].score {
				unique[idx] = c
			}
			continue
		}
		indexMap[c.product.ProductId] = len(unique)
		unique = append(unique, c)
	}

	// 计算综合评分
	maxHot := float64(1)
	for _, c := range unique {
		hot := float64(0)
		if c.product.ProductHot.Valid {
			hot = float64(c.product.ProductHot.Int64)
		}
		if hot > maxHot {
			maxHot = hot
		}
	}

	type ranked struct {
		item       types.RecommendItem
		finalScore float64
	}
	rankedList := make([]ranked, 0, len(unique))

	for _, c := range unique {
		p := c.product
		hot := float64(0)
		if p.ProductHot.Valid {
			hot = float64(p.ProductHot.Int64)
		}
		sales := int64(0)
		if p.ProductSales.Valid {
			sales = p.ProductSales.Int64
		}
		picture := ""
		if p.ProductPicture.Valid {
			picture = p.ProductPicture.String
		}

		// 价格匹配分：价格越接近gap越好，超过threshold的扣分
		var priceMatchScore float64
		if gap > 0 {
			priceDiff := math.Abs(p.ProductSellingPrice - gap)
			priceMatchScore = math.Max(0, 100-priceDiff/gap*80)
			// 如果商品价格超过差额太多（超过2倍），大幅扣分
			if p.ProductSellingPrice > gap*2 {
				priceMatchScore *= 0.3
			}
		} else {
			priceMatchScore = 50
		}

		// 热度归一化分
		hotScore := (hot / maxHot) * 100

		// 综合评分 = 策略分*0.4 + 价格匹配*0.4 + 热度*0.2
		finalScore := c.score*0.4 + priceMatchScore*0.4 + hotScore*0.2

		rankedList = append(rankedList, ranked{
			item: types.RecommendItem{
				ProductID:           p.ProductId,
				ProductName:         p.ProductName,
				CategoryID:          p.CategoryId,
				ProductTitle:        p.ProductTitle,
				ProductPicture:      picture,
				ProductPrice:        p.ProductPrice,
				ProductSellingPrice: p.ProductSellingPrice,
				ProductSales:        sales,
				ProductHot:          int64(hot),
				RecommendReason:     c.reason,
				Score:               math.Round(finalScore*100) / 100,
			},
			finalScore: finalScore,
		})
	}

	// 按综合评分降序排序
	sort.Slice(rankedList, func(i, j int) bool {
		return rankedList[i].finalScore > rankedList[j].finalScore
	})

	results := make([]types.RecommendItem, 0, len(rankedList))
	for _, r := range rankedList {
		results = append(results, r.item)
	}
	return results
}

// ==================== 工具函数 ====================

func contains(ids []int64, target int64) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
