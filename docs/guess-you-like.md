# JMall 猜你喜欢：推荐系统原理与实现

> 本文档基于 JMall 项目真实代码，从推荐算法原理到工程实现完整讲解。
> 技术栈：Go (go-zero) + MySQL + Redis + Vue.js
> 目标：读完这篇文档，面试中能把推荐系统从算法到落地讲清楚。

---

## 一、推荐系统的本质问题

推荐系统要解决的核心问题：**在海量商品中，找到用户最可能感兴趣的那几个**。

电商场景下，用户不可能浏览所有商品。推荐系统就是一个"信息过滤器"：
- 输入：用户的历史行为（浏览、点击、加购、购买、收藏）
- 输出：一个按兴趣排序的商品列表

### 1.1 我们的推荐系统包含两个场景

| 场景 | 位置 | 目的 |
|------|------|------|
| 猜你喜欢 | 商品详情页底部 | 提升用户浏览深度，增加转化 |
| 智能凑单 | 购物车页面 | 凑满减，提升客单价 |

本文重点讲"猜你喜欢"，凑单推荐在最后一章简要介绍。

---

## 二、整体架构：召回 → 排序 → 重排

这是工业界推荐系统的标准三层架构，我们的实现完全遵循这个范式。

```
全部商品（~50个）
    │
    ▼
┌──────────────────────────────────────────┐
│            召回层 (Recall)                 │
│  从全量商品中快速筛选出候选集               │
│                                           │
│  策略1: 用户偏好召回 ──→ ~30个             │
│  策略2: ItemCF 召回  ──→ ~30个             │
│  策略3: UserCF 召回  ──→ ~30个             │
│  策略4: 热门兜底召回 ──→ ~30个             │
│                                           │
│  合并去重 ──→ ~60-80个候选                 │
└──────────────┬───────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────┐
│            排序层 (Rank)                   │
│  对候选集精细打分排序                       │
│                                           │
│  综合评分 = 策略基础分 × 0.3               │
│           + 用户偏好匹配分 × 0.3           │
│           + 热度分 × 0.2                   │
│           + 多样性扰动 × 0.2               │
│                                           │
│  按评分降序排列                            │
└──────────────┬───────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────┐
│            重排层 (Re-rank)                │
│  业务规则调整最终展示顺序                   │
│                                           │
│  · 分类打散（同分类不连续超过2个）          │
│  · 已购/已浏览去重                         │
│  · 分页截取                                │
│                                           │
│  输出 ──→ 20个/页                          │
└──────────────────────────────────────────┘
```

### 面试话术

> "我们的推荐系统采用经典的三层架构：召回层用四路策略并行从全量商品中
> 快速筛选候选集；排序层用多维度加权评分做精细排序；重排层做分类打散
> 和业务规则过滤。这个架构和工业界的做法是一致的，只是我们用规则打分
> 替代了机器学习模型。"

---

## 三、数据基础：用户行为采集

推荐系统的质量取决于数据质量。没有用户行为数据，推荐就是瞎猜。

### 3.1 行为类型与权重

我们采集 5 种用户行为，每种行为赋予不同权重：

| 行为 | 类型值 | 权重 | 含义 |
|------|--------|------|------|
| 浏览 | 1 | 1.0 | 用户看了一眼，兴趣最弱 |
| 点击 | 2 | 2.0 | 主动点击，有一定兴趣 |
| 加购 | 3 | 3.0 | 加入购物车，购买意向明确 |
| 收藏 | 4 | 4.0 | 收藏了，强烈兴趣但暂不购买 |
| 购买 | 5 | 5.0 | 已购买，兴趣最强 |

**为什么收藏权重比购买低？** 购买是最强的正反馈信号——用户用真金白银投票了。收藏说明感兴趣但还在犹豫。

### 3.2 行为上报机制

前端在用户交互时异步上报行为，不阻塞用户操作：

```javascript
// Details.vue — 用户进入商品详情页时上报"浏览"行为
reportBehavior(product, 1)  // behavior_type=1 浏览

// GuessYouLike.vue — 用户点击推荐商品时上报"点击"行为
reportClick(item) {
    this.$axios.post('/api/recommend/reportBehavior', {
        product_id: item.product_id,
        category_id: item.category_id,
        behavior_type: 2,  // 点击
    }).catch(() => {})  // 静默失败，不影响用户体验
}
```

后端写入 `user_behavior` 表：

```go
// reportbehaviorlogic.go
func (l *ReportBehaviorLogic) ReportBehavior(req) {
    behavior := &model.UserBehavior{
        UserId:       userID,
        ProductId:    req.ProductID,
        CategoryId:   req.CategoryID,
        BehaviorType: req.BehaviorType,
        BehaviorTime: time.Now().UnixMilli(),
    }
    l.svcCtx.UserBehaviorModel.Insert(l.ctx, behavior)
    // 插入失败也返回成功——行为上报不能影响用户体验
}
```

### 3.3 数据库表设计

```sql
CREATE TABLE `user_behavior` (
  `id`            bigint  NOT NULL AUTO_INCREMENT,
  `user_id`       int     NOT NULL,
  `product_id`    int     NOT NULL,
  `category_id`   int     NOT NULL,
  `behavior_type` tinyint NOT NULL COMMENT '1浏览 2点击 3加购 4购买 5收藏',
  `behavior_time` bigint  NOT NULL COMMENT '时间戳(ms)',
  PRIMARY KEY (`id`),
  INDEX `idx_user_time` (`user_id`, `behavior_time` DESC),
  INDEX `idx_product_behavior` (`product_id`, `behavior_type`)
) ENGINE=InnoDB COMMENT='用户行为日志';
```

索引设计的考量：
- `idx_user_time`：按用户查最近行为（猜你喜欢的核心查询）
- `idx_product_behavior`：按商品查行为用户（ItemCF 离线计算用）

---

## 四、召回层：四路策略详解

召回层的目标：从全量商品中快速筛选出"可能感兴趣"的候选集。
不追求精确，追求覆盖率——宁可多召回，不能漏掉好商品。

### 4.1 策略一：用户偏好召回

**原理：** 统计用户近 30 天的行为，找出偏好的商品分类，推荐该分类下的热门商品。

```
用户行为数据 → 按分类聚合加权评分 → 取 Top 5 偏好分类 → 推荐这些分类下的热门商品
```

**SQL 查询用户偏好分类：**
```sql
SELECT category_id
FROM user_behavior
WHERE user_id = ? AND behavior_time > ?  -- 近30天
GROUP BY category_id
ORDER BY SUM(CASE behavior_type
    WHEN 1 THEN 1  -- 浏览
    WHEN 2 THEN 2  -- 点击
    WHEN 3 THEN 3  -- 加购
    WHEN 4 THEN 5  -- 购买（权重最高）
    WHEN 5 THEN 4  -- 收藏
END) DESC
LIMIT 5
```

**代码实现：**
```go
func (l *GuessYouLikeLogic) recallByUserPreference(preferredCats, excludeIds []int64) []candidate {
    // 查询偏好分类下的热门商品（排除已看过的）
    products, _ := l.svcCtx.ProductModel.FindByCategoryIds(l.ctx, preferredCats, excludeIds, 30)

    results := make([]candidate, 0, len(products))
    for _, p := range products {
        results = append(results, candidate{
            productId: p.ProductId,
            reason:    "猜你喜欢",
            score:     80 + hot*0.5,  // 基础分 80 + 热度加成
        })
    }
    return results
}
```

**面试要点：** 这是最简单但最有效的策略。用户买了手机，大概率还会看手机壳、充电器。

### 4.2 策略二：ItemCF（基于物品的协同过滤）

**原理：** 如果两个商品被很多相同用户交互过，则它们相似。推荐用户交互过的商品的相似商品。

**数学公式（余弦相似度）：**
```
sim(A, B) = |users(A) ∩ users(B)| / sqrt(|users(A)| × |users(B)|)
```

举例：
- 商品 A 被用户 {1,2,3,4,5} 交互过
- 商品 B 被用户 {1,2,3,6,7} 交互过
- 交集 = {1,2,3}，|交集| = 3
- sim(A,B) = 3 / sqrt(5 × 5) = 3/5 = 0.6

**离线计算过程（itemcf.go）：**

```go
func ComputeItemCF(ctx context.Context, svcCtx *svc.ServiceContext) error {
    // 1. 构建商品→用户倒排索引
    //    productUsers[商品A] = {用户1, 用户2, 用户3, ...}
    productUsers := make(map[int64]map[int64]bool)
    for _, product := range allProducts {
        users := FindUsersByProduct(product.Id, 30天)
        productUsers[product.Id] = users
    }

    // 2. 两两计算余弦相似度
    for i, pidA := range productIds {
        for j := i+1; j < len(productIds); j++ {
            pidB := productIds[j]
            intersection := |usersA ∩ usersB|
            sim := intersection / sqrt(|usersA| × |usersB|)

            // 3. 双向写入相似度表
            BatchUpsert(pidA → pidB, sim)
            BatchUpsert(pidB → pidA, sim)
        }
    }
}
```

**相似度存储表：**
```sql
CREATE TABLE `product_similarity` (
  `product_id`         int    NOT NULL,
  `similar_product_id` int    NOT NULL,
  `score`              double NOT NULL COMMENT '相似度 0~1',
  `updated_at`         bigint NOT NULL,
  UNIQUE INDEX `uk_product_pair` (`product_id`, `similar_product_id`),
  INDEX `idx_product_score` (`product_id`, `score` DESC)
) ENGINE=InnoDB COMMENT='商品相似度（ItemCF离线计算）';
```

**在线召回：**
```go
func (l *GuessYouLikeLogic) recallByItemCF(recentProductIds []int64) []candidate {
    // 取用户最近交互的前 10 个商品作为种子
    seedIds := recentProductIds[:10]

    // 查相似度表，找这些种子商品的相似商品
    similarities := FindSimilarProductsByIds(seedIds, 30)

    for _, sim := range similarities {
        results = append(results, candidate{
            productId: sim.SimilarProductId,
            reason:    "相似商品推荐",
            score:     70 + sim.Score*30,  // 基础分 70 + 相似度加成
        })
    }
}
```

**面试话术：**
> "ItemCF 的核心思想是：如果两个商品被很多相同用户交互过，它们就是相似的。
> 我们用余弦相似度离线计算商品两两相似度，结果存到 product_similarity 表。
> 在线服务时，取用户最近交互的商品作为种子，查相似度表召回候选。
> 离线计算可以用定时任务每天凌晨跑一次。"

### 4.3 策略三：UserCF（基于用户的协同过滤）

**原理：** 找到和当前用户行为相似的其他用户，推荐他们喜欢但当前用户没看过的商品。

```
当前用户 → 找行为相似的用户 → 他们交互过但我没看过的商品 → 推荐给我
```

**SQL 找相似用户（按共同交互商品数排序）：**
```sql
SELECT b2.user_id
FROM user_behavior b1
JOIN user_behavior b2
  ON b1.product_id = b2.product_id AND b2.user_id != b1.user_id
WHERE b1.user_id = ? AND b1.behavior_time > ? AND b2.behavior_time > ?
GROUP BY b2.user_id
ORDER BY COUNT(DISTINCT b2.product_id) DESC  -- 共同交互商品越多越相似
LIMIT 10
```

**代码实现：**
```go
func (l *GuessYouLikeLogic) recallByUserCF(userID int64, excludeProductIds []int64) []candidate {
    // 1. 找 10 个最相似的用户
    similarUsers := FindSimilarUsers(userID, 30天, 10)

    // 2. 查这些用户交互过但当前用户没看过的商品
    behaviors := FindProductsByUsers(similarUsers, excludeProductIds, 30天, 30)

    for _, b := range behaviors {
        weight := BehaviorWeight[b.BehaviorType]
        results = append(results, candidate{
            productId: b.ProductId,
            reason:    "和你口味相似的人也在看",
            score:     60 + weight*5,  // 基础分 60 + 行为权重加成
        })
    }
}
```

### 4.4 策略四：热门兜底召回

**原理：** 全站热销 Top N。这是冷启动的兜底策略——新用户没有行为数据时，推荐热门商品总不会太差。

```go
func (l *GuessYouLikeLogic) recallByHotSelling(excludeIds []int64) []candidate {
    products := FindTopHot(30)  // 全站热度 Top 30
    for _, p := range products {
        results = append(results, candidate{
            productId: p.ProductId,
            reason:    "热门推荐",
            score:     40 + hot*0.3,  // 基础分最低，只做兜底
        })
    }
}
```

### 4.5 四路召回的基础分设计

| 策略 | 基础分 | 设计意图 |
|------|--------|---------|
| 用户偏好 | 80 | 最了解用户，分最高 |
| ItemCF | 70 | 基于相似度，可信度高 |
| UserCF | 60 | 基于相似用户，间接推断 |
| 热门兜底 | 40 | 不个性化，分最低 |

基础分决定了不同策略的优先级。用户偏好召回的商品天然排在前面，热门兜底的排在后面。

---

## 五、排序层：多维度加权评分

召回层给出了候选集，排序层要精细排序。

### 5.1 评分公式

```
综合评分 = 策略基础分 × 0.3
         + 用户偏好匹配分 × 0.3
         + 热度分 × 0.2
         + 多样性扰动 × 0.2
```

### 5.2 各维度详解

**策略基础分（0.3 权重）：** 召回策略给出的分数，归一化到 0-100。

**用户偏好匹配分（0.3 权重）：**
```go
preferenceScore := 0
if 商品分类在用户偏好分类中 {
    preferenceScore = 80
}
if 用户之前对该商品有行为 {
    preferenceScore += min(行为评分 × 5, 20)  // 最多加 20 分
}
```

**热度分（0.2 权重）：** 商品热度归一化到 0-100。热门商品有加成，但权重不高，避免推荐全是爆款。

**多样性扰动（0.2 权重）：** `rand.Float64() * 30`，随机 0-30 分。
- 为什么要加随机？避免每次推荐结果完全一样
- 用户刷新"换一批"时能看到不同商品
- 20% 的权重不会颠覆排序，但能增加新鲜感

### 5.3 代码实现

```go
func (l *GuessYouLikeLogic) rank(candidates []candidate, ...) []rankedCandidate {
    // 1. 去重：同一商品保留最高分的候选
    bestMap := make(map[int64]candidate)
    for _, c := range candidates {
        if existing, ok := bestMap[c.productId]; !ok || c.score > existing.score {
            bestMap[c.productId] = c
        }
    }

    // 2. 批量获取商品详情（一次 DB 查询，避免 N+1）
    products := FindByIds(productIds)

    // 3. 计算综合评分
    for pid, c := range bestMap {
        strategyScore := min(c.score, 100)
        preferenceScore := ...  // 偏好匹配
        hotScore := (hot / maxHot) * 100  // 归一化
        diversityScore := rand.Float64() * 30  // 随机扰动

        finalScore := strategyScore*0.3 + preferenceScore*0.3 + hotScore*0.2 + diversityScore*0.2
    }

    // 4. 按综合评分降序排列
    sort.Slice(ranked, func(i, j int) bool {
        return ranked[i].finalScore > ranked[j].finalScore
    })
}
```

---

## 六、重排层：分类打散

排序层输出的列表可能出现"连续 5 个手机"的情况——用户体验很差。
重排层的核心任务是**分类打散**：同一分类的商品不连续超过 2 个。

### 6.1 算法

```
输入：[手机A, 手机B, 手机C, 手机D, 电视E, 电视F, 空调G]
                                ↓ 分类打散（连续 ≤ 2）
输出：[手机A, 手机B, 电视E, 手机C, 电视F, 空调G, 手机D]
```

两阶段实现：
1. 第一轮（scatter）：遍历候选列表，同分类连续数 ≤ 2 的放入结果，超过的跳过
2. 第二轮（fill）：把第一轮跳过的商品追加到末尾

```go
func (l *GuessYouLikeLogic) rerank(ranked []rankedCandidate) []rankedCandidate {
    result := make([]rankedCandidate, 0, len(ranked))
    used := make(map[int64]bool)

    // 第一轮：分类打散
    var lastCat int64 = -1
    catCount := 0
    for _, r := range ranked {
        if r.categoryId == lastCat && catCount >= 2 {
            continue  // 跳过，留给第二轮
        }
        result = append(result, r)
        used[r.productId] = true
        if r.categoryId == lastCat {
            catCount++
        } else {
            lastCat = r.categoryId
            catCount = 1
        }
    }

    // 第二轮：追加未入选的
    for _, r := range ranked {
        if !used[r.productId] {
            result = append(result, r)
        }
    }
    return result
}
```

---

## 七、冷启动问题

新用户没有行为数据，怎么推荐？

### 7.1 我们的解决方案

```go
isNewUser := len(userScores) == 0

if !isNewUser {
    // 老用户：四路召回全开
    recallByUserPreference(...)
    recallByItemCF(...)
    recallByUserCF(...)
}
// 所有用户都走热门兜底（新用户只走这一路）
recallByHotSelling(...)
```

- 新用户：只有热门兜底召回，推荐全站热销商品
- 随着用户产生行为（浏览、点击），逐渐切换到个性化推荐
- 这是一个**渐进式**的过程，不需要等用户积累大量数据

### 7.2 面试话术

> "冷启动我们分两种情况：新用户冷启动用热门兜底策略，推荐全站热销商品；
> 新商品冷启动靠 ItemCF 离线计算时的分类关联。随着用户行为积累，
> 个性化策略的权重会自然增大，热门兜底的权重自然降低。"

---

## 八、性能优化

### 8.1 Redis 缓存

```go
// 推荐结果缓存 3 分钟
cacheKey := fmt.Sprintf("jmall:recommend:guess:%d:%d:%d", userID, page, pageSize)
if err := l.svcCtx.Cache.Get(l.ctx, cacheKey, &cached); err == nil {
    return &cached, nil  // 缓存命中，直接返回
}

// ... 计算推荐 ...

l.svcCtx.Cache.Set(l.ctx, cacheKey, resp, 3*time.Minute)
```

为什么只缓存 3 分钟？
- 太长：用户行为变了但推荐没变，体验差
- 太短：缓存命中率低，等于没缓存
- 3 分钟是平衡点：同一用户短时间内多次请求走缓存，行为变化后自然更新

### 8.2 批量查询避免 N+1

```go
// ❌ 错误做法：循环中逐个查询
for _, candidate := range candidates {
    product := FindOne(candidate.productId)  // N 次 DB 查询
}

// ✅ 正确做法：收集 ID 后批量查询
productIds := collectIds(candidates)
products := FindByIds(productIds)  // 1 次 DB 查询
productMap := buildMap(products)
```

### 8.3 ItemCF 离线计算

相似度计算是 O(n²) 的，不能在线实时算。我们的做法：
- 离线：定时任务每天凌晨计算，结果写入 `product_similarity` 表
- 在线：直接查表，O(1) 时间复杂度

```
离线（每天凌晨）：ComputeItemCF() → product_similarity 表
在线（用户请求）：SELECT * FROM product_similarity WHERE product_id IN (?) ORDER BY score DESC
```

---

## 九、智能凑单推荐（FillUp）

这是推荐系统的第二个场景，放在购物车页面。

### 9.1 核心思路

```
购物车总价 1800 元 → 满 2000 减 200 → 差 200 元
                                        ↓
推荐价格在 100~300 元区间的商品，让用户加一件就能凑到满减
```

### 9.2 三策略并行

| 策略 | 说明 | 基础分 |
|------|------|--------|
| 差额精准推荐 | 价格在 [gap×0.5, gap×1.5] 区间 | 按价格接近度打分 |
| 关联商品推荐 | 购物车商品的搭配品（手机→手机壳） | 90（搭配购最高分） |
| 热销兜底 | 全站热销 | 50 |

### 9.3 评分公式

```
综合评分 = 策略基础分 × 0.4 + 价格匹配分 × 0.4 + 热度分 × 0.2
```

价格匹配分的权重最高（0.4），因为凑单场景下价格是否合适是最重要的。

### 9.4 关联品类映射

```go
var relatedCategoryMap = map[int64][]int64{
    1: {5, 6, 7, 8},  // 手机 → 保护套, 保护膜, 充电器, 充电宝
    5: {1},            // 保护套 → 手机
    6: {1},            // 保护膜 → 手机
    7: {1},            // 充电器 → 手机
    8: {1},            // 充电宝 → 手机
}
```

购物车里有手机 → 推荐手机壳、充电器等配件。这是基于业务知识的硬编码规则，简单但有效。

---

## 十、前端实现

### 10.1 猜你喜欢组件（GuessYouLike.vue）

- 瀑布流网格布局（5 列响应式）
- 无限滚动加载（距底部 200px 触发）
- "换一批"刷新（重置分页）
- 点击上报行为（异步不阻塞）

```javascript
handleScroll() {
    const scrollTop = document.documentElement.scrollTop
    const clientHeight = document.documentElement.clientHeight
    const scrollHeight = document.documentElement.scrollHeight
    if (scrollTop + clientHeight >= scrollHeight - 200) {
        this.fetchRecommendations()  // 触底加载下一页
    }
}
```

### 10.2 凑单推荐组件（FillUpRecommend.vue）

- 满减进度条（实时显示距满减还差多少）
- 一键加购（加购后自动刷新推荐）
- 展开/收起动画

---

## 十一、ItemCF vs UserCF 对比

面试常问：你为什么同时用了 ItemCF 和 UserCF？

| 维度 | ItemCF | UserCF |
|------|--------|--------|
| 核心思想 | 商品相似 → 推荐相似商品 | 用户相似 → 推荐相似用户喜欢的 |
| 适用场景 | 商品数 < 用户数（电商） | 用户数 < 商品数（社交） |
| 实时性 | 离线计算，更新慢 | 在线计算，实时性好 |
| 冷启动 | 新商品需要积累行为 | 新用户需要积累行为 |
| 可解释性 | "因为你看了 A，推荐相似的 B" | "和你口味相似的人也在看" |
| 我们的用法 | 离线算相似度，在线查表 | 在线实时计算 |

两者互补：ItemCF 擅长"看了又看"，UserCF 擅长"发现新品类"。

---

## 十二、面试高频问题

### Q1：你的推荐系统用了什么算法？

三层架构：召回用四路策略（用户偏好、ItemCF、UserCF、热门兜底），排序用多维度加权评分，重排用分类打散。没有用深度学习模型，用的是规则+协同过滤。

### Q2：ItemCF 的相似度怎么算的？

余弦相似度。构建商品→用户倒排索引，两两计算交集用户数除以各自用户数的几何平均。离线计算，每天跑一次，结果存 DB。

### Q3：冷启动怎么解决？

新用户走热门兜底策略。随着行为积累，个性化策略自然生效。不需要显式切换。

### Q4：推荐结果怎么保证多样性？

两个机制：排序层加 20% 权重的随机扰动，重排层做分类打散（同分类不连续超过 2 个）。

### Q5：为什么不用机器学习模型？

项目规模决定的。我们的商品量级是几十个，用户量级是几百个。这个数据量训练模型没有意义，规则+协同过滤已经足够好。如果数据量上去了（百万级商品、千万级用户），可以引入 LR/DeepFM 等模型替换排序层。

### Q6：推荐系统的效果怎么评估？

线上指标：点击率（CTR）、转化率、人均浏览深度、客单价。
离线指标：准确率、召回率、覆盖率、多样性。
我们目前主要看点击率和转化率。

### Q7：凑单推荐和猜你喜欢有什么区别？

目标不同。猜你喜欢的目标是"让用户多逛"，优化浏览深度和转化率。凑单推荐的目标是"让用户多买"，优化客单价。所以凑单推荐的排序公式中价格匹配分权重最高（0.4），而猜你喜欢更看重用户偏好。

### Q8：行为上报失败了怎么办？

静默处理。行为上报是"尽力而为"的，失败不影响用户体验。少了几条行为数据不会显著影响推荐质量。如果要做到不丢，可以用 Kafka 异步写入。

### Q9：为什么推荐结果只缓存 3 分钟？

平衡实时性和性能。太长用户行为变了推荐没变，太短缓存命中率低。3 分钟内同一用户多次请求走缓存，行为变化后自然更新。

### Q10：如果要升级这个推荐系统，你会怎么做？

1. 排序层引入 LR/GBDT 模型替代规则打分
2. 特征工程：加入用户画像（年龄、性别）、上下文特征（时间、设备）
3. 实时特征：用 Flink 实时计算用户近 5 分钟的行为特征
4. A/B 测试框架：对比不同策略的效果
5. 向量召回：用 Embedding 做语义相似度召回

---

## 十三、项目结构

```
backend/service/recommendation/
├── recommendation.go                    # 入口
├── etc/recommendation-api.yaml          # 配置
└── internal/
    ├── config/config.go
    ├── handler/
    │   ├── routes.go                    # 3 个接口路由
    │   ├── guessyoulikehandler.go       # 猜你喜欢
    │   ├── filluphandler.go             # 智能凑单
    │   └── reportbehaviorhandler.go     # 行为上报
    ├── logic/
    │   ├── guessyoulikelogic.go         # 核心：四路召回→排序→重排
    │   ├── filluplogic.go              # 凑单：三策略→评分→排序
    │   ├── reportbehaviorlogic.go       # 行为写入 DB
    │   └── itemcf.go                    # ItemCF 离线计算引擎
    ├── middleware/authmiddleware.go
    ├── svc/servicecontext.go
    └── types/types.go

backend/model/
├── userbehaviormodel.go                 # 用户行为（偏好分类、相似用户、行为评分）
├── productsimilaritymodel.go            # 商品相似度（ItemCF 结果查询、批量 UPSERT）
└── sql/recommendation.sql               # 建表 SQL
```

---

## 十四、一句话总结

猜你喜欢的本质是一个**多路召回 + 加权排序 + 分类打散**的推荐管道。四路召回保证覆盖率，加权评分保证相关性，分类打散保证多样性。没有用深度学习，但架构是工业级的——如果未来数据量上去了，只需要把排序层的规则打分替换成模型打分，其他层不用动。
