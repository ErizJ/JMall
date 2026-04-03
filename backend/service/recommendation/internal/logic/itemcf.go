package logic

import (
	"context"
	"math"
	"time"

	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/recommendation/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// ============================================================
// ItemCF 离线计算引擎
//
// 算法：基于用户行为的 Item-based Collaborative Filtering
// 核心思路：如果两个商品被很多相同用户交互过，则它们相似
//
// 相似度计算：余弦相似度
//   sim(A,B) = |users(A) ∩ users(B)| / sqrt(|users(A)| * |users(B)|)
//
// 调度方式：
//   - 可通过定时任务（cron）每天凌晨执行一次
//   - 也可通过管理后台手动触发
//   - 结果写入 product_similarity 表
// ============================================================

// ComputeItemCF 离线计算商品相似度并写入数据库
// 适合在定时任务中调用，如每天凌晨执行
func ComputeItemCF(ctx context.Context, svcCtx *svc.ServiceContext) error {
	logger := logx.WithContext(ctx)
	logger.Info("ItemCF computation started")

	// 1. 获取所有商品
	products, err := svcCtx.ProductModel.FindAll(ctx)
	if err != nil {
		return err
	}

	// 2. 构建商品→用户倒排索引
	// productUsers[productId] = set of userIds
	productUsers := make(map[int64]map[int64]bool)
	for _, p := range products {
		users, err := svcCtx.UserBehaviorModel.FindUsersByProduct(ctx, p.ProductId, 30)
		if err != nil {
			continue
		}
		userSet := make(map[int64]bool, len(users))
		for _, uid := range users {
			userSet[uid] = true
		}
		if len(userSet) > 0 {
			productUsers[p.ProductId] = userSet
		}
	}

	// 3. 计算两两商品的余弦相似度
	productIds := make([]int64, 0, len(productUsers))
	for pid := range productUsers {
		productIds = append(productIds, pid)
	}

	now := time.Now().UnixMilli()
	batch := make([]*model.ProductSimilarity, 0, 100)

	for i := 0; i < len(productIds); i++ {
		pidA := productIds[i]
		usersA := productUsers[pidA]

		for j := i + 1; j < len(productIds); j++ {
			pidB := productIds[j]
			usersB := productUsers[pidB]

			// 计算交集
			intersection := 0
			for uid := range usersA {
				if usersB[uid] {
					intersection++
				}
			}

			if intersection == 0 {
				continue
			}

			// 余弦相似度
			sim := float64(intersection) / math.Sqrt(float64(len(usersA))*float64(len(usersB)))

			// 双向写入
			batch = append(batch,
				&model.ProductSimilarity{ProductId: pidA, SimilarProductId: pidB, Score: sim, UpdatedAt: now},
				&model.ProductSimilarity{ProductId: pidB, SimilarProductId: pidA, Score: sim, UpdatedAt: now},
			)

			// 批量写入，每100条一批
			if len(batch) >= 100 {
				if err := svcCtx.ProductSimilarityModel.BatchUpsert(ctx, batch); err != nil {
					logger.Errorf("ItemCF batch upsert error: %v", err)
				}
				batch = batch[:0]
			}
		}
	}

	// 写入剩余数据
	if len(batch) > 0 {
		if err := svcCtx.ProductSimilarityModel.BatchUpsert(ctx, batch); err != nil {
			logger.Errorf("ItemCF final batch upsert error: %v", err)
		}
	}

	logger.Infof("ItemCF computation completed, processed %d products", len(productIds))
	return nil
}
