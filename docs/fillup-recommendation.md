# JMall 凑满减推荐：原理与实现

> 本文档基于 JMall 项目真实代码，从凑单推荐的业务逻辑到算法实现完整讲解。
> 技术栈：Go (go-zero) + MySQL + Redis + Vue.js
> 目标：读完这篇文档，面试中能把凑满减推荐从业务到技术讲清楚。

---

## 一、凑满减推荐解决什么问题？

用户购物车里有 1800 元的商品，满 2000 减 200。差 200 元。

传统做法：用户自己去翻商品列表，找一个 200 元左右的商品加进去。
智能凑单：系统自动推荐价格合适的商品，用户一键加购就能凑到满减。

核心价值：**提升客单价**。用户本来只想花 1800，凑单后花了 2000 但省了 200，实际多花了 0 元却多买了一件商品。商家多卖了一件，用户觉得自己赚了。双赢。

### 1.1 和"猜你喜欢"的区别

| 维度 | 猜你喜欢 | 凑满减推荐 |
|------|---------|-----------|
| 目标 | 让用户多逛（提升浏览深度） | 让用户多买（提升客单价） |
| 位置 | 商品详情页底部 | 购物车页面 |
| 核心因素 | 用户兴趣偏好 | 价格是否合适 |
| 排序权重 | 偏好匹配 0.3 > 热度 0.2 | 价格匹配 0.4 = 策略分 0.4 |
| 触发条件 | 随时 | 购物车有商品且未达满减门槛 |

---

## 二、整体流程

```
用户打开购物车页面
    │
    ▼
前端调用 POST /recommend/fillup
    │
    ▼
┌──────────────────────────────────────────────────┐
│                 FillUpLogic                        │
│                                                    │
│  1. 读购物车 → 计算总价 1800 元                    │
│  2. 读满减规则 → 满2000减200，满3000减300          │
│  3. 找最近未达标档位 → 满2000减200，差200元         │
│  4. 三策略并行召回候选商品：                        │
│     ├── 差额精准推荐（价格 100~300 元）            │
│     ├── 关联商品推荐（搭配购 + 同品类配件）        │
│     └── 热销兜底推荐（全站热销）                   │
│  5. 去重 + 综合评分排序                            │
│  6. 返回 Top 12 推荐商品                           │
└──────────────────────────────────────────────────┘
    │
    ▼
前端展示：满减进度条 + 推荐商品列表 + 一键加购
```

---

## 三、满减规则解析

### 3.1 数据来源

满减规则存在 `combination_product` 表中：

```sql
CREATE TABLE `combination_product` (
  `id`                  int NOT NULL AUTO_INCREMENT,
  `main_product_id`     int NOT NULL COMMENT '主商品ID',
  `vice_product_id`     int NOT NULL COMMENT '搭配商品ID',
  `amountThreshold`     int NULL COMMENT '满减门槛（元）',
  `priceReductionRange` int NULL COMMENT '减免金额（元）',
  PRIMARY KEY (`id`)
);
```

示例数据：
```
main=1(Redmi K30), vice=9(小米电视),  threshold=2000, reduction=200
main=2(Redmi K30 5G), vice=8(Redmi 7A), threshold=3000, reduction=300
```

这张表同时存了两个信息：商品搭配关系 + 满减规则。

### 3.2 提取满减档位

从表中提取去重的满减档位，缓存 10 分钟：

```go
func (l *FillUpLogic) getPromotionTiers() []promotionTier {
    // 优先读缓存
    const tiersCacheKey = "jmall:promotion:tiers"
    if cached := cache.Get(tiersCacheKey); cached != nil {
        return cached
    }

    // 从 combination_product 表提取去重档位
    combos := CombinationProductModel.FindAll()
    tierMap := make(map[float64]float64)  // threshold → reduction
    for _, c := range combos {
        tierMap[c.AmountThreshold] = max(tierMap[c.AmountThreshold], c.PriceReductionRange)
    }

    // 按门槛升序排列：[{2000, 200}, {3000, 300}]
    sort.Slice(tiers, func(i, j int) { return tiers[i].Threshold < tiers[j].Threshold })

    cache.Set(tiersCacheKey, tiers, 10*time.Minute)
    return tiers
}
```

### 3.3 找最近未达标档位

```go
func (l *FillUpLogic) findNearestTier(tiers []promotionTier, cartTotal float64) (*promotionTier, float64) {
    for i := range tiers {
        if cartTotal < tiers[i].Threshold {
            gap := tiers[i].Threshold - cartTotal
            return &tiers[i], gap  // 返回最近的未达标档位和差额
        }
    }
    return nil, 0  // 已达到所有档位，无需凑单
}
```

例子：
- 购物车 1800 元 → 最近档位 2000，差 200 元
- 购物车 2500 元 → 已过 2000 档，最近档位 3000，差 500 元
- 购物车 3500 元 → 已达所有档位，不推荐凑单

---

## 四、三策略并行召回

### 4.1 策略一：差额精准推荐

**核心思路：** 推荐价格在 `[差额×0.5, 差额×1.5]` 区间的商品。用户加一件就能凑到满减。

```
差额 200 元 → 推荐价格 100~300 元的商品
```

为什么是 0.5~1.5 倍而不是精确等于差额？
- 精确匹配太少，可能找不到商品
- 略低于差额的也有价值（加两件凑到）
- 略高于差额的也可以（多花一点但凑到了）

```go
func (l *FillUpLogic) strategyPriceGap(gap float64, excludeIds []int64) []scoredProduct {
    minPrice := gap * 0.5
    maxPrice := gap * 1.5
    if minPrice < 1 { minPrice = 1 }

    // SQL: WHERE selling_price >= 100 AND selling_price <= 300 AND product_num > 0
    products := ProductModel.FindByPriceRange(minPrice, maxPrice, excludeIds, 20)

    for _, p := range products {
        // 评分：价格越接近差额，分数越高
        priceDiff := abs(p.SellingPrice - gap)
        priceScore := max(0, 100 - priceDiff/gap*100)
        // 差额200，商品200元 → priceScore=100（满分）
        // 差额200，商品100元 → priceScore=50
        // 差额200，商品300元 → priceScore=50
    }
}
```

**评分逻辑：** 价格越接近差额，分数越高。这是凑单场景最重要的信号。

### 4.2 策略二：关联商品推荐

**核心思路：** 推荐和购物车商品有关联的商品。分两个子策略：

**a) 搭配购推荐（combination_product 表）**

```go
// 查购物车中每个商品的搭配商品
for _, cartProductId := range cartProductIds {
    combos := CombinationProductModel.FindByMainProductId(cartProductId)
    for _, c := range combos {
        viceProductIds = append(viceProductIds, c.ViceProductId)
    }
}
// 批量查询搭配商品详情（避免 N+1）
viceProducts := ProductModel.FindByIds(viceProductIds)
// 搭配购给最高分 90
```

**b) 关联品类推荐**

买手机 → 推荐手机壳、充电器、充电宝。这是基于业务知识的硬编码映射：

```go
var relatedCategoryMap = map[int64][]int64{
    1: {5, 6, 7, 8},  // 手机 → 保护套, 保护膜, 充电器, 充电宝
    5: {1},            // 保护套 → 手机
    6: {1},            // 保护膜 → 手机
    7: {1},            // 充电器 → 手机
    8: {1},            // 充电宝 → 手机
}
```

```go
// 找购物车商品分类的关联品类
relatedCatIds := make(map[int64]bool)
for _, cartCategoryId := range cartCategoryIds {
    for _, relatedCatId := range relatedCategoryMap[cartCategoryId] {
        relatedCatIds[relatedCatId] = true
    }
}
// 查关联品类下的商品，基础分 70 + 热度加成
products := ProductModel.FindByCategoryIds(catIds, excludeIds, 15)
```

### 4.3 策略三：热销兜底推荐

全站热销 Top 20，保证推荐列表不为空：

```go
func (l *FillUpLogic) strategyHotSelling(excludeIds []int64) []scoredProduct {
    products := ProductModel.FindTopHot(20)
    // 基础分 50 + 热度加成（最低分，只做兜底）
}
```

### 4.4 三策略的基础分设计

| 策略 | 基础分 | 设计意图 |
|------|--------|---------|
| 差额精准 | 0~100（按价格接近度） | 价格最合适的排最前 |
| 搭配购 | 90 | 商家配置的搭配，可信度高 |
| 关联品类 | 70 + 热度加成 | 品类相关但不如搭配购精准 |
| 热销兜底 | 50 + 热度加成 | 不个性化，分最低 |

---

## 五、排序层：综合评分

### 5.1 评分公式

```
综合评分 = 策略基础分 × 0.4 + 价格匹配分 × 0.4 + 热度分 × 0.2
```

和"猜你喜欢"的公式对比：

| 维度 | 猜你喜欢权重 | 凑满减权重 | 原因 |
|------|-------------|-----------|------|
| 策略基础分 | 0.3 | 0.4 | 凑单更依赖策略质量 |
| 用户偏好/价格匹配 | 0.3 | 0.4 | 凑单场景价格是第一要素 |
| 热度 | 0.2 | 0.2 | 相同 |
| 多样性扰动 | 0.2 | 0（无） | 凑单不需要随机性，要稳定 |

**关键区别：凑满减没有随机扰动。** 因为凑单推荐要稳定——用户刷新页面看到的应该是同样的推荐（除非购物车变了）。而猜你喜欢需要"换一批"的新鲜感。

### 5.2 价格匹配分详解

```go
var priceMatchScore float64
if gap > 0 {
    priceDiff := math.Abs(p.SellingPrice - gap)
    priceMatchScore = math.Max(0, 100 - priceDiff/gap*80)

    // 如果商品价格超过差额 2 倍，大幅扣分
    if p.SellingPrice > gap*2 {
        priceMatchScore *= 0.3
    }
} else {
    priceMatchScore = 50  // 已达满减，给中间分
}
```

例子（差额 200 元）：
| 商品价格 | priceDiff | priceMatchScore | 说明 |
|---------|-----------|-----------------|------|
| 200 元 | 0 | 100 | 完美匹配 |
| 150 元 | 50 | 80 | 接近，高分 |
| 100 元 | 100 | 60 | 还行 |
| 50 元 | 150 | 40 | 偏低 |
| 500 元 | 300 | 0 → ×0.3 | 超过 2 倍，大幅扣分 |

**为什么超过 2 倍要扣分？** 差 200 元推荐一个 500 元的商品，用户会觉得"我只想凑 200，你让我多花 500？"体验很差。

### 5.3 去重逻辑

三个策略可能召回同一个商品（比如一个热销手机壳同时命中差额精准和热销兜底）。去重时保留分数最高的：

```go
// O(n) 去重，用 map 索引替代内层循环
indexMap := make(map[int64]int)  // productId → index
for _, c := range candidates {
    if idx, exists := indexMap[c.product.ProductId]; exists {
        if c.score > unique[idx].score {
            unique[idx] = c  // 保留高分的
        }
    } else {
        indexMap[c.product.ProductId] = len(unique)
        unique = append(unique, c)
    }
}
```

---

## 六、缓存设计

### 6.1 推荐结果缓存

```go
// 缓存 key 设计：用户ID + 购物车总价（分为单位取整）
cacheKey := fmt.Sprintf("jmall:recommend:fillup:%d:%d", userID, int64(math.Round(cartTotal*100)))
```

为什么用购物车总价做缓存 key 的一部分？
- 同一用户，购物车总价变了（加了商品/删了商品），推荐结果应该变
- 总价没变，推荐结果可以复用

为什么用 `math.Round(cartTotal*100)` 转成分？
- 浮点数有精度问题：1800.00 和 1800.0000001 是不同的 key
- 转成分（整数）避免浮点截断导致的缓存碰撞

缓存 TTL = 2 分钟。比猜你喜欢（3 分钟）更短，因为购物车变化更频繁。

### 6.2 满减规则缓存

```go
const tiersCacheKey = "jmall:promotion:tiers"
cache.Set(tiersCacheKey, tiers, 10*time.Minute)
```

满减规则是准静态数据（运营配置后很少变），缓存 10 分钟。避免每次请求都查 `combination_product` 表。

---

## 七、前端实现

### 7.1 组件位置

`FillUpRecommend` 组件放在购物车页面底部：

```html
<!-- ShoppingCart.vue -->
<div class="cart-wrap" v-if="getShoppingCart.length > 0">
    <FillUpRecommend @cartUpdated="reloadCart" />
</div>
```

购物车有商品时才显示。加购后触发 `cartUpdated` 事件，购物车页面重新加载数据。

### 7.2 满减进度条

```
┌──────────────────────────────────────────────────┐
│ [满减] 还差 ¥200.00 享满2000减200  [████████░░] [去凑单 ▼] │
└──────────────────────────────────────────────────┘
```

进度计算：
```javascript
progressPercent() {
    if (this.nearestRule.threshold <= 0) return 100
    return Math.min(100, Math.round((this.cartTotal / this.nearestRule.threshold) * 100))
}
// 购物车 1800 元，门槛 2000 → 进度 90%
```

### 7.3 一键加购

```javascript
async addToCart(productId) {
    // 1. 检查购物车是否已有该商品
    const checkRes = await this.$axios.post('/api/user/shoppingCart/isExistShoppingCart', {
        user_id: userId, product_id: productId
    })

    if (checkRes.data.code === '002') {
        // 不存在 → 新增
        await this.$axios.post('/api/user/shoppingCart/addShoppingCart', {
            user_id: userId, product_id: productId, num: 1
        })
    } else {
        // 已存在 → 数量+1
        this.addShoppingCartNum(productId)
    }

    // 2. 加购成功后，从推荐列表中移除该商品
    this.recommendations = this.recommendations.filter(r => r.product_id !== productId)

    // 3. 重新请求推荐（购物车变了，推荐也要变）
    this.fetchRecommendations()
}
```

加购后自动刷新推荐——因为购物车总价变了，差额变了，推荐的商品也应该变。

### 7.4 响应式更新

```javascript
watch: {
    getTotalPrice() {
        this.fetchRecommendations()  // 购物车总价变化时自动刷新推荐
    }
}
```

用户在购物车里改数量、删商品，总价变化 → 自动重新请求推荐。

---

## 八、完整数据流

```
用户打开购物车
    │
    ▼
FillUpRecommend.mounted() → fetchRecommendations()
    │
    ▼
POST /recommend/fillup
    │
    ▼
FillUpLogic.FillUp():
    │
    ├── 1. ShoppingcartModel.FindByUserId(userID)
    │      → [{productId:1, num:1}, {productId:9, num:1}]
    │
    ├── 2. ProductModel.FindByIds([1, 9])
    │      → [{id:1, price:1599}, {id:9, price:799}]
    │      → cartTotal = 1599 + 799 = 2398
    │
    ├── 3. getPromotionTiers()
    │      → [{threshold:2000, reduction:200}, {threshold:3000, reduction:300}]
    │      → 2398 > 2000 → 已过第一档
    │      → 2398 < 3000 → 最近档位 3000，差 602 元
    │
    ├── 4a. strategyPriceGap(gap=602)
    │       → 价格 301~903 元的商品
    │       → [{id:4, name:"Redmi 8", price:699, score:84}]
    │
    ├── 4b. strategyAssociated(cart=[1,9], categories=[手机,电视])
    │       → 搭配购：combination_product 表查 main=1 的搭配
    │       → 关联品类：手机→保护套/充电器，电视→无
    │       → [{id:19, name:"K20保护壳", price:39, score:90}]
    │
    ├── 4c. strategyHotSelling()
    │       → 全站热销 Top 20
    │       → [{id:16, name:"空调", price:2599, score:52}]
    │
    ├── 5. deduplicateAndRank(gap=602, threshold=3000)
    │      → 综合评分 = 策略分×0.4 + 价格匹配×0.4 + 热度×0.2
    │      → Redmi 8 (699元): 策略84×0.4 + 价格84×0.4 + 热度×0.2 = 高分
    │      → K20保护壳 (39元): 策略90×0.4 + 价格低(差太远)×0.4 = 中分
    │
    └── 6. 返回 Top 12
    │
    ▼
前端渲染：进度条（差602元到满3000减300）+ 推荐列表
```

---

## 九、性能优化

| 优化点 | 做法 | 效果 |
|--------|------|------|
| 满减规则缓存 | Redis 缓存 10 分钟 | 避免每次查 combination_product 表 |
| 推荐结果缓存 | Redis 缓存 2 分钟，key 含总价 | 短时间内重复请求走缓存 |
| 批量查询 | `FindByIds` 一次查多个商品 | 避免 N+1 查询 |
| 搭配商品批量查 | 先收集所有 viceProductId，再一次 FindByIds | 避免循环中逐个查 |
| 结果数量限制 | 最多返回 12 个 | 减少数据传输量 |
| 去重用 map 索引 | `map[int64]int` 替代内层循环 | O(n²) → O(n) |

---

## 十、和业界方案的对比

| 维度 | 我们的实现 | 美团/京东 |
|------|-----------|----------|
| 满减规则 | 从 DB 读取，缓存 10min | 运营后台配置，实时生效 |
| 召回策略 | 3 路（差额+关联+热销） | 10+ 路（含用户画像、实时行为） |
| 排序模型 | 规则加权评分 | LR/GBDT/DNN 模型 |
| 关联品类 | 硬编码映射表 | 基于购买数据挖掘的关联规则 |
| 实时性 | 购物车变化时重新请求 | 实时流计算 + 增量更新 |
| 个性化 | 无（不考虑用户偏好） | 结合用户画像个性化排序 |

我们的方案适合中小规模电商，简单有效。如果要升级，可以：
1. 排序层引入模型（LR/GBDT）
2. 关联品类从硬编码改为数据挖掘（频繁项集/Apriori 算法）
3. 加入用户偏好因子（结合猜你喜欢的用户画像）

---

## 十一、面试高频问题

### Q1：凑满减推荐的核心思路是什么？

三步：1）算出购物车离最近满减门槛差多少钱；2）三策略并行召回候选商品（差额精准、关联搭配、热销兜底）；3）按"策略分×0.4 + 价格匹配分×0.4 + 热度×0.2"综合评分排序。价格匹配是最重要的因素——推荐的商品价格要接近差额。

### Q2：为什么价格匹配分权重这么高（0.4）？

凑单场景下，用户的核心诉求是"加一件就能凑到满减"。如果推荐一个 2000 元的商品来凑 200 元的差额，用户会觉得荒谬。价格合适是凑单推荐的第一要素。

### Q3：差额精准推荐为什么用 0.5~1.5 倍区间？

精确匹配太少。0.5 倍是因为用户可以加两件凑到；1.5 倍是因为多花一点也能接受。这个区间在实践中能覆盖大部分场景。

### Q4：搭配购推荐的数据从哪来？

`combination_product` 表，由运营在后台配置。比如"买 Redmi K30 搭配小米电视"。这是人工策划的搭配关系，可信度最高，所以基础分给了 90（最高）。

### Q5：关联品类映射为什么是硬编码？

因为我们的商品品类只有 13 个，关联关系很明确（手机→配件）。硬编码简单直接，维护成本低。如果品类扩展到几百个，应该改用数据挖掘（比如 Apriori 算法从购买数据中自动发现关联规则）。

### Q6：为什么凑满减没有随机扰动（猜你喜欢有）？

凑单推荐要稳定。用户在购物车页面反复查看，每次看到的推荐应该一样（除非购物车变了）。随机扰动会让用户困惑："刚才看到的那个商品怎么没了？"

### Q7：缓存 key 为什么包含购物车总价？

购物车总价变了（加了商品/删了商品/改了数量），差额就变了，推荐结果也应该变。用总价做 key 的一部分，确保购物车变化时缓存自动失效。

### Q8：如果用户已经达到所有满减档位怎么办？

`findNearestTier` 返回 nil，直接返回空推荐列表。前端组件检测到 `recommendations.length === 0 && gap === 0` 时不显示。

### Q9：加购后为什么要重新请求推荐？

因为购物车总价变了。比如差 200 元时推荐了一个 199 元的商品，用户加购后总价变了，可能已经达到满减门槛了，或者差额变成了 1 元，推荐列表应该更新。

### Q10：如果要做个性化凑单推荐，怎么改？

在排序公式中加入用户偏好因子。比如：
```
综合评分 = 策略分×0.3 + 价格匹配×0.3 + 用户偏好×0.2 + 热度×0.2
```
用户偏好可以复用猜你喜欢的用户画像（偏好分类、行为评分）。这样同样价格的两个商品，用户更感兴趣的那个排前面。

---

## 十二、项目结构

```
backend/service/recommendation/internal/logic/
└── filluplogic.go          # 凑单推荐核心逻辑
    ├── FillUp()             # 入口：读购物车→算差额→三策略召回→排序
    ├── getPromotionTiers()  # 满减规则解析（缓存 10min）
    ├── findNearestTier()    # 找最近未达标档位
    ├── strategyPriceGap()   # 策略1：差额精准推荐
    ├── strategyAssociated() # 策略2：搭配购 + 关联品类
    ├── strategyHotSelling() # 策略3：热销兜底
    └── deduplicateAndRank() # 去重 + 综合评分排序

backend/model/
├── combinationproductmodel.go  # 满减/搭配数据（FindAll, FindByMainProductId）
├── productmodel.go             # 商品查询（FindByPriceRange, FindByCategoryIds, FindTopHot）
└── shoppingcartmodel.go        # 购物车（FindByUserId）

frontend/src/components/
└── FillUpRecommend.vue     # 满减进度条 + 推荐列表 + 一键加购
```

---

## 十三、一句话总结

凑满减推荐的核心是**价格导向的三策略召回**：差额精准推荐找价格合适的商品，关联搭配推荐找品类相关的商品，热销兜底保证列表不为空。排序时价格匹配分权重最高（0.4），因为凑单场景下"加一件就能凑到"是用户的第一诉求。
