package logic

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/ErizJ/JMall/backend/ctxutil"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/recommendation/internal/svc"
	"github.com/ErizJ/JMall/backend/service/recommendation/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// ============================================================
// 猜你喜欢推荐引擎
//
// 架构：召回层(Recall) → 排序层(Rank) → 重排层(Re-rank)
//
// 召回策略（多路并行）：
//   1. 用户行为召回 — 基于用户近期浏览/点击/加购/购买行为的偏好分类
//   2. ItemCF 召回  — 基于商品相似度表，找用户交互过商品的相似商品
//   3. UserCF 召回  — 找行为相似的用户，推荐他们喜欢但当前用户没看过的商品
//   4. 热门兜底召回 — 全站热销商品（冷启动兜底）
//
// 排序层：
//   综合评分 = 策略基础分×0.3 + 用户偏好匹配分×0.3 + 热度分×0.2 + 多样性扰动×0.2
//
// 重排层：
//   - 分类打散（同分类不连续超过2个）
//   - 已购去重
//   - 分页支持
// ============================================================

const (
	defaultPageSize = 20
	maxPageSize     = 50
	behaviorDays    = 30 // 行为数据回溯天数
)

type GuessYouLikeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGuessYouLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GuessYouLikeLogic {
	return &GuessYouLikeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// candidate 召回候选商品
type candidate struct {
	productId int64
	reason    string
	score     float64 // 召回策略给出的基础分
}

func (l *GuessYouLikeLogic) GuessYouLike(req *types.GuessYouLikeReq) (*types.GuessYouLikeResp, error) {
	userID, ctxErr := ctxutil.UserIDFromCtx(l.ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > maxPageSize {
		pageSize = defaultPageSize
	}

	// ========== 1. 尝试读缓存 ==========
	cacheKey := fmt.Sprintf("jmall:recommend:guess:%d:%d:%d", userID, page, pageSize)
	var cached types.GuessYouLikeResp
	if err := l.svcCtx.Cache.Get(l.ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	// ========== 2. 获取用户行为画像 ==========
	userScores, _ := l.svcCtx.UserBehaviorModel.FindUserProductBehaviors(l.ctx, userID, behaviorDays)
	preferredCats, _ := l.svcCtx.UserBehaviorModel.FindUserPreferredCategories(l.ctx, userID, behaviorDays, 5)
	recentProductIds, _ := l.svcCtx.UserBehaviorModel.FindRecentProductIds(l.ctx, userID, 50)

	// 用户偏好分类集合
	preferredCatSet := make(map[int64]bool, len(preferredCats))
	for _, cid := range preferredCats {
		preferredCatSet[cid] = true
	}

	// 用户对商品的行为评分 map
	userProductScoreMap := make(map[int64]float64, len(userScores))
	for _, us := range userScores {
		userProductScoreMap[us.ProductId] = us.Score
	}

	isNewUser := len(userScores) == 0

	// ========== 3. 多路召回 ==========
	allCandidates := make([]candidate, 0, 100)

	if !isNewUser {
		// 策略1: 用户行为偏好召回 — 推荐用户偏好分类下的高热度商品
		allCandidates = append(allCandidates, l.recallByUserPreference(preferredCats, recentProductIds)...)

		// 策略2: ItemCF 召回 — 基于商品相似度
		allCandidates = append(allCandidates, l.recallByItemCF(recentProductIds)...)

		// 策略3: UserCF 召回 — 基于用户相似度
		allCandidates = append(allCandidates, l.recallByUserCF(userID, recentProductIds)...)
	}

	// 策略4: 热门兜底召回（所有用户都走，保证推荐列表不为空）
	allCandidates = append(allCandidates, l.recallByHotSelling(recentProductIds)...)

	// ========== 4. 排序层 — 综合评分 ==========
	ranked := l.rank(allCandidates, preferredCatSet, userProductScoreMap)

	// ========== 5. 重排层 — 去重 + 分类打散 + 分页 ==========
	reranked := l.rerank(ranked)

	// 分页
	start := (page - 1) * pageSize
	end := start + pageSize
	hasMore := false
	if start >= len(reranked) {
		reranked = nil
	} else {
		if end < len(reranked) {
			hasMore = true
		}
		if end > len(reranked) {
			end = len(reranked)
		}
		reranked = reranked[start:end]
	}

	// ========== 6. 填充商品详情 ==========
	result := l.fillProductDetails(reranked)

	resp := types.GuessYouLikeResp{
		Code:            "200",
		Recommendations: result,
		HasMore:         hasMore,
	}

	// 写缓存 3 分钟
	_ = l.svcCtx.Cache.Set(l.ctx, cacheKey, resp, 3*time.Minute)

	return &resp, nil
}

// ==================== 召回策略1: 用户行为偏好召回 ====================
// 根据用户偏好分类，推荐该分类下热度最高的商品

func (l *GuessYouLikeLogic) recallByUserPreference(preferredCats []int64, excludeIds []int64) []candidate {
	if len(preferredCats) == 0 {
		return nil
	}

	products, err := l.svcCtx.ProductModel.FindByCategoryIds(l.ctx, preferredCats, excludeIds, 30)
	if err != nil {
		l.Logger.Errorf("recallByUserPreference error: %v", err)
		return nil
	}

	results := make([]candidate, 0, len(products))
	for _, p := range products {
		hot := float64(0)
		if p.ProductHot.Valid {
			hot = float64(p.ProductHot.Int64)
		}
		results = append(results, candidate{
			productId: p.ProductId,
			reason:    "猜你喜欢",
			score:     80 + hot*0.5, // 偏好召回基础分80
		})
	}
	return results
}

// ==================== 召回策略2: ItemCF 召回 ====================
// 基于商品相似度表，找用户最近交互商品的相似商品

func (l *GuessYouLikeLogic) recallByItemCF(recentProductIds []int64) []candidate {
	if len(recentProductIds) == 0 {
		return nil
	}

	// 取最近交互的前10个商品做 ItemCF
	seedIds := recentProductIds
	if len(seedIds) > 10 {
		seedIds = seedIds[:10]
	}

	similarities, err := l.svcCtx.ProductSimilarityModel.FindSimilarProductsByIds(l.ctx, seedIds, 30)
	if err != nil {
		l.Logger.Errorf("recallByItemCF error: %v", err)
		return nil
	}

	results := make([]candidate, 0, len(similarities))
	for _, sim := range similarities {
		results = append(results, candidate{
			productId: sim.SimilarProductId,
			reason:    "相似商品推荐",
			score:     70 + sim.Score*30, // ItemCF 基础分70 + 相似度加成
		})
	}
	return results
}

// ==================== 召回策略3: UserCF 召回 ====================
// 找行为相似的用户，推荐他们喜欢但当前用户没看过的商品

func (l *GuessYouLikeLogic) recallByUserCF(userID int64, excludeProductIds []int64) []candidate {
	similarUsers, err := l.svcCtx.UserBehaviorModel.FindSimilarUsers(l.ctx, userID, behaviorDays, 10)
	if err != nil || len(similarUsers) == 0 {
		return nil
	}

	behaviors, err := l.svcCtx.UserBehaviorModel.FindProductsByUsers(l.ctx, similarUsers, excludeProductIds, behaviorDays, 30)
	if err != nil {
		l.Logger.Errorf("recallByUserCF error: %v", err)
		return nil
	}

	results := make([]candidate, 0, len(behaviors))
	for _, b := range behaviors {
		weight := model.BehaviorWeight[b.BehaviorType]
		results = append(results, candidate{
			productId: b.ProductId,
			reason:    "和你口味相似的人也在看",
			score:     60 + weight*5, // UserCF 基础分60 + 行为权重加成
		})
	}
	return results
}

// ==================== 召回策略4: 热门兜底召回 ====================
// 全站热销 Top N，冷启动兜底

func (l *GuessYouLikeLogic) recallByHotSelling(excludeIds []int64) []candidate {
	products, err := l.svcCtx.ProductModel.FindTopHot(l.ctx, 30)
	if err != nil {
		l.Logger.Errorf("recallByHotSelling error: %v", err)
		return nil
	}

	excludeSet := make(map[int64]bool, len(excludeIds))
	for _, id := range excludeIds {
		excludeSet[id] = true
	}

	results := make([]candidate, 0, len(products))
	for _, p := range products {
		if excludeSet[p.ProductId] || p.ProductNum <= 0 {
			continue
		}
		hot := float64(0)
		if p.ProductHot.Valid {
			hot = float64(p.ProductHot.Int64)
		}
		results = append(results, candidate{
			productId: p.ProductId,
			reason:    "热门推荐",
			score:     40 + hot*0.3, // 热门兜底基础分40
		})
	}
	return results
}

// ==================== 排序层 ====================
// 综合评分 = 策略基础分×0.3 + 用户偏好匹配分×0.3 + 热度分×0.2 + 多样性分×0.2

type rankedCandidate struct {
	productId  int64
	categoryId int64
	reason     string
	finalScore float64
}

func (l *GuessYouLikeLogic) rank(candidates []candidate, preferredCatSet map[int64]bool, userProductScoreMap map[int64]float64) []rankedCandidate {
	if len(candidates) == 0 {
		return nil
	}

	// 去重：同一商品保留最高分的候选
	bestMap := make(map[int64]candidate, len(candidates))
	for _, c := range candidates {
		if existing, ok := bestMap[c.productId]; !ok || c.score > existing.score {
			bestMap[c.productId] = c
		}
	}

	// 批量获取商品详情用于排序
	productIds := make([]int64, 0, len(bestMap))
	for pid := range bestMap {
		productIds = append(productIds, pid)
	}
	products, err := l.svcCtx.ProductModel.FindByIds(l.ctx, productIds)
	if err != nil {
		l.Logger.Errorf("rank FindByIds error: %v", err)
		return nil
	}

	productMap := make(map[int64]*model.Product, len(products))
	for _, p := range products {
		productMap[p.ProductId] = p
	}

	// 找最大热度用于归一化
	maxHot := float64(1)
	for _, p := range products {
		if p.ProductHot.Valid && float64(p.ProductHot.Int64) > maxHot {
			maxHot = float64(p.ProductHot.Int64)
		}
	}

	ranked := make([]rankedCandidate, 0, len(bestMap))
	for pid, c := range bestMap {
		p, ok := productMap[pid]
		if !ok || p.ProductNum <= 0 {
			continue
		}

		// 1. 策略基础分（归一化到 0-100）
		strategyScore := math.Min(c.score, 100)

		// 2. 用户偏好匹配分
		preferenceScore := float64(0)
		if preferredCatSet[p.CategoryId] {
			preferenceScore = 80
		}
		// 如果用户之前对该商品有行为，额外加分
		if userScore, ok := userProductScoreMap[pid]; ok {
			preferenceScore += math.Min(userScore*5, 20)
		}

		// 3. 热度分（归一化）
		hot := float64(0)
		if p.ProductHot.Valid {
			hot = float64(p.ProductHot.Int64)
		}
		hotScore := (hot / maxHot) * 100

		// 4. 多样性分（随机扰动，增加推荐多样性）
		diversityScore := rand.Float64() * 30

		// 综合评分
		finalScore := strategyScore*0.3 + preferenceScore*0.3 + hotScore*0.2 + diversityScore*0.2

		ranked = append(ranked, rankedCandidate{
			productId:  pid,
			categoryId: p.CategoryId,
			reason:     c.reason,
			finalScore: finalScore,
		})
	}

	// 按综合评分降序
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].finalScore > ranked[j].finalScore
	})

	return ranked
}

// ==================== 重排层 ====================
// 分类打散：同分类商品不连续超过2个，提升推荐多样性
//
// 实现：两阶段
//   - 第一轮（scatter）：遍历候选列表，严格限制同分类连续数 ≤ 2，
//     无法放入的商品跳过（defer）
//   - 第二轮（fill）：将第一轮跳过的商品按原顺序追加到末尾
//     两轮状态完全独立，避免跨轮状态污染

func (l *GuessYouLikeLogic) rerank(ranked []rankedCandidate) []rankedCandidate {
	if len(ranked) == 0 {
		return nil
	}

	result := make([]rankedCandidate, 0, len(ranked))
	used := make(map[int64]bool, len(ranked))

	// 第一轮：分类打散
	var scatterCat int64 = -1
	scatterCount := 0
	for _, r := range ranked {
		if r.categoryId == scatterCat && scatterCount >= 2 {
			continue // 本轮跳过，留给第二轮补充
		}
		result = append(result, r)
		used[r.productId] = true
		if r.categoryId == scatterCat {
			scatterCount++
		} else {
			scatterCat = r.categoryId
			scatterCount = 1
		}
	}

	// 第二轮：追加第一轮未入选的商品（状态独立，不受第一轮影响）
	for _, r := range ranked {
		if !used[r.productId] {
			result = append(result, r)
		}
	}

	return result
}

// ==================== 填充商品详情 ====================

func (l *GuessYouLikeLogic) fillProductDetails(ranked []rankedCandidate) []types.RecommendItem {
	if len(ranked) == 0 {
		return []types.RecommendItem{}
	}

	productIds := make([]int64, 0, len(ranked))
	for _, r := range ranked {
		productIds = append(productIds, r.productId)
	}

	products, err := l.svcCtx.ProductModel.FindByIds(l.ctx, productIds)
	if err != nil {
		l.Logger.Errorf("fillProductDetails error: %v", err)
		return []types.RecommendItem{}
	}

	productMap := make(map[int64]*model.Product, len(products))
	for _, p := range products {
		productMap[p.ProductId] = p
	}

	results := make([]types.RecommendItem, 0, len(ranked))
	for _, r := range ranked {
		p, ok := productMap[r.productId]
		if !ok {
			continue
		}
		picture := ""
		if p.ProductPicture.Valid {
			picture = p.ProductPicture.String
		}
		sales := int64(0)
		if p.ProductSales.Valid {
			sales = p.ProductSales.Int64
		}
		hot := int64(0)
		if p.ProductHot.Valid {
			hot = p.ProductHot.Int64
		}

		results = append(results, types.RecommendItem{
			ProductID:           p.ProductId,
			ProductName:         p.ProductName,
			CategoryID:          p.CategoryId,
			ProductTitle:        p.ProductTitle,
			ProductPicture:      picture,
			ProductPrice:        p.ProductPrice,
			ProductSellingPrice: p.ProductSellingPrice,
			ProductSales:        sales,
			ProductHot:          hot,
			RecommendReason:     r.reason,
			Score:               math.Round(r.finalScore*100) / 100,
		})
	}
	return results
}
