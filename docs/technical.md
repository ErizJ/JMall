# JMall 技术文档

> 本文档详细描述 JMall 后端（go-zero REST 微服务）的所有业务流程、缓存策略与数据库交互细节。

---

## 目录

1. [整体架构](#1-整体架构)
2. [基础设施](#2-基础设施)
3. [认证机制](#3-认证机制)
4. [缓存策略总览](#4-缓存策略总览)
5. [用户服务（user，端口 8881）](#5-用户服务)
6. [商品服务（product，端口 8882）](#6-商品服务)
7. [购物车服务（cart，端口 8883）](#7-购物车服务)
8. [订单服务（order，端口 8884）](#8-订单服务)
9. [收藏服务（collect，端口 8885）](#9-收藏服务)
10. [管理服务（management，端口 8886）](#10-管理服务)
11. [支付服务（payment，端口 8887）](#11-支付服务)
12. [热度系统](#12-热度系统)
13. [数据库表结构](#13-数据库表结构)
14. [AI 智能助手服务（aichat，端口 8888）](#14-ai-智能助手服务)
15. [智能凑单推荐服务（recommendation，端口 8889）](#15-智能凑单推荐服务)
16. [猜你喜欢推荐系统](#16-猜你喜欢推荐系统)

---

## 1. 整体架构

```
客户端（浏览器）
      │
      ▼
  Nginx / Vue CLI Dev Server（:8080）
      │
      ├─ /api/users/*      → user-api      :8881
      ├─ /api/products/*   → product-api   :8882
      ├─ /api/cart/*       → cart-api      :8883
      ├─ /api/orders/*     → order-api     :8884
      ├─ /api/collect/*    → collect-api   :8885
      ├─ /api/management/* → management-api :8886
      ├─ /api/payment/*      → payment-api       :8887
      ├─ /api/aichat/*       → aichat-api        :8888
      └─ /api/recommend/*    → recommendation-api :8889
                  │                │
            ┌─────┴─────┐         │
            ▼           ▼         ▼
          MySQL 8.0   Redis 7   豆包大模型 API
          (storedb)   (DB 0)    (ark.cn-beijing.volces.com)
```

**9 个独立 go-zero REST 服务**，共用同一个 MySQL 数据库和同一个 Redis 实例。其中 aichat 服务额外对接豆包大模型 API，通过 MCP 工具协议查询数据库后由大模型生成自然语言回复。recommendation 服务提供智能凑满减推荐和猜你喜欢个性化推荐功能。每个服务有自己的 `ServiceContext`，持有数据库 Model 和 Redis Client。

---

## 2. 基础设施

### MySQL

数据库名：`storedb`，字符集 `utf8mb4`。

| 表名 | 用途 |
|------|------|
| `users` | 用户账号 |
| `sysmanager` | 管理员账号（与 users 独立） |
| `product` | 商品信息 |
| `product_picture` | 商品附加图片 |
| `category` | 商品分类 |
| `combination_product` | 搭配购组合（满减规则） |
| `shoppingcart` | 购物车 |
| `orders` | 订单行项（一笔逻辑订单对应多行，含 status 字段） |
| `collect` | 收藏夹 |
| `carousel` | 首页轮播图 |
| `payment_order` | 支付单（支付流水） |
| `payment_refund` | 退款单 |
| `user_behavior` | 用户行为日志（浏览/点击/加购/购买/收藏） |
| `product_similarity` | 商品相似度（ItemCF 离线计算结果） |

### Redis

所有 key 以 `jmall:` 为命名空间前缀。值均为 JSON 序列化后的字节串，通过 `cache.Client`（`backend/cache/cache.go`）封装的 `Set/Get/Del` 读写。

```go
// cache.go 核心接口
func (c *Client) Set(ctx, key, value, ttl) error   // JSON 序列化后写入
func (c *Client) Get(ctx, key, dest) error          // 读取并反序列化；redis.Nil 表示 miss
func (c *Client) Del(ctx, keys...) error            // 删除一个或多个 key
func (c *Client) SetNX(ctx, key, value, ttl) error  // 仅当 key 不存在时写入（幂等/分布式锁）
func (c *Client) Eval(ctx, script, keys, args) (interface{}, error) // 执行 Lua 脚本（原子操作）
```

---

## 3. 认证机制

### JWT 生成（登录时）

```
POST /users/login
  → 验证账号密码
  → jwt.NewWithClaims(HS256, MapClaims{
        "userId":   user.UserId,
        "userName": user.UserName,
        "iat":      now.Unix(),
        "exp":      now.Unix() + config.Auth.ExpireSeconds,   // 默认 86400s = 24h
    })
  → 签名密钥：config.Auth.Secret（yaml 中配置，默认 "jmall-secret-key-change-in-production"）
  → 返回 token 字符串
```

### JWT 验证（AuthMiddleware）

每个需要认证的服务都有独立的 `internal/middleware/authmiddleware.go`，逻辑完全相同：

```
请求到达
  → 读取 Header: Authorization: Bearer <token>
  → 若缺失或格式错误 → 返回 {code: "401"}，终止
  → jwt.Parse(token, HS256, secret)
  → 若签名无效或过期 → 返回 {code: "401"}，终止
  → 从 claims 取出 userId（float64 类型，JWT MapClaims 的默认数字类型）
  → context.WithValue(ctx, ctxutil.CtxKeyUserID, userId)
  → 调用下一个 Handler
```

### 从 Context 提取 UserID

所有需要用到当前登录用户 ID 的 Logic 都调用：

```go
// backend/ctxutil/userid.go
userID, err := ctxutil.UserIDFromCtx(l.ctx)
// 内部处理 float64 / int64 两种类型，统一转为 int64
```

> **安全保证**：所有受保护接口的 userID 来自 JWT 上下文，而非请求体。客户端无法伪造其他用户的 ID。

---

## 4. 缓存策略总览

采用标准 **Cache-Aside（旁路缓存）** 模式：读时先查 Redis，miss 则查 DB 并回填；写/删时直接操作 DB，再主动删除对应 cache key。

| Cache Key | TTL | 写入时机 | 失效时机 |
|-----------|-----|----------|----------|
| `jmall:user:detail:{userId}` | 5 min | GetUserDetail（DB miss 时） | UpdateUser、DeleteUser |
| `jmall:cart:user:{userId}` | 2 min | GetCart（DB miss 时） | AddCart、DeleteCart、UpdateCart、AddOrder |
| `jmall:orders:user:{userId}` | 2 min | GetOrder（DB miss 时） | AddOrder、DeleteOrder |
| `jmall:collect:user:{userId}` | 2 min | GetCollect（DB miss 时） | AddCollect、DeleteCollect |
| `jmall:products:all` | 10 min | GetAllProduct | AddProduct、DeleteProduct、UpdateProduct、SetCategoryHotZero |
| `jmall:categories` | 10 min | GetCategory | SetCategoryHotZero（product 服务） |
| `jmall:carousel` | 10 min | GetCarousel | 从不主动失效 |
| `jmall:products:hot:7` | 5 min | GetHotProduct / GetAllUserRecommend | AddCollect、AddProduct、DeleteProduct、UpdateProduct、SetCategoryHotZero |
| `jmall:products:promotion:7` | 5 min | GetPromotionProduct | AddProduct、DeleteProduct、UpdateProduct、SetCategoryHotZero |
| `jmall:product:recommend:personal` | 5 min | GetRecommendProduct | AddCollect、SetCategoryHotZero |
| `jmall:products:category:{categoryId}` | 10 min | GetProductByCategory | 未主动失效（依赖 TTL） |
| `jmall:products:promo:{categoryId}` | 5 min | GetPromoProduct | 未主动失效（依赖 TTL） |
| `jmall:product:detail:{productId}` | 5 min | GetProductDetail | DeleteProduct、UpdateProduct |
| `jmall:product:pictures:{productId}` | 5 min | GetProductPictures | DeleteProduct |
| `jmall:product:phone:7` | 5 min | GetPhoneList | AddProduct、DeleteProduct、UpdateProduct、SetCategoryHotZero |
| `jmall:product:shell:7` | 5 min | GetProtectingShellList | 同上 |
| `jmall:product:charger:7` | 5 min | GetChargerList | 同上 |
| `jmall:order:submit:{userId}` | 5 s | AddOrder（SETNX 防重复提交） | TTL 自动过期 |
| `jmall:recommend:fillup:{userId}:{totalHash}` | 2 min | FillUp（凑单推荐） | 购物车变更时自动过期（TTL） |
| `jmall:recommend:guess:{userId}:{page}` | 3 min | GuessYouLike（猜你喜欢） | TTL 自动过期 |
| `jmall:stock:{productId}` | 10 min | AddOrder（Lua 原子预扣库存） | 退款/取消/删除订单时清理 |
| `jmall:order:expire:{orderId}` | 30 min | AddOrder（订单超时标记） | TTL 自动过期 |
| `jmall:payment:lock:{orderId}` | 30 min | CreatePayment（SETNX 防重复支付） | 支付成功/失败/过期后清理 |
| `jmall:payment:notify:{paymentNo}` | 24 h | PaymentNotify（SETNX 回调幂等锁） | 事务失败时清理允许重试 |
| `jmall:payment:user:{userId}` | - | GetUserPayments | 支付/退款后清理 |
| `jmall:refund:lock:{paymentNo}` | 30 s | Refund（SETNX 退款防重复提交） | 退款完成/失败后清理 |

---

## 5. 用户服务

### 5.1 注册 `POST /users/register`

```
输入: { userName, password, userPhoneNumber? }

1. 格式校验
   - userNameRegex: ^[a-zA-Z][a-zA-Z0-9_]{4,15}$
   - passwordRegex: ^[a-zA-Z]\w{5,17}$
   - 不合法 → code "002"

2. 检查用户名是否已存在
   - UsersModel.FindOneByUserName(userName)
   - 已存在 → code "003"

3. 密码哈希
   - bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)

4. 写入数据库
   - UsersModel.Insert({ UserName, Password(哈希), UserPhoneNumber })

5. 返回 { code: "200" }
```

### 5.2 登录 `POST /users/login`

```
输入: { userName, password }

1. 格式校验（同注册）→ code "002" on fail

2. 查询用户
   - UsersModel.FindOneByUserName(userName)
   - 不存在 → code "002"

3. 密码验证
   - bcrypt.CompareHashAndPassword(storedHash, inputPassword)
   - 不匹配 → code "002"

4. 生成 JWT
   - Claims: { userId, userName, iat, exp }
   - Algorithm: HS256
   - Secret: config.Auth.Secret

5. 返回 { code: "200", userId, userName, token }
```

### 5.3 登出 `POST /users/logout`

```
JWT 为无状态设计，服务端不维护会话列表。
直接返回 { code: "200" }。
```

> Token 在客户端删除后自然失效，服务端无任何操作。

### 5.4 检查用户名 `POST /users/findUserName`

```
输入: { userName }

1. 格式校验 → code "002"
2. UsersModel.FindOneByUserName(userName)
3. 存在 → code "003"（已占用）
4. 不存在 → code "200"（可用）
```

### 5.5 获取用户详情 `POST /users/getDetails` 🔒

```
输入: 无（userId 从 JWT context 读取）

1. ctxutil.UserIDFromCtx(ctx) → userId

2. 读取缓存
   key: jmall:user:detail:{userId}
   hit  → 直接返回缓存的 { code, userId, userName, userPhoneNumber }
   miss → 继续

3. DB 查询
   UsersModel.FindOne(userId)

4. 回填缓存（TTL 5 min）

5. 返回 { code: "200", userId, userName, userPhoneNumber }
```

### 5.6 更新用户信息 `POST /users/updateUser` 🔒

```
输入: { userName?, userPhoneNumber? }

1. ctxutil.UserIDFromCtx(ctx) → userId
2. UsersModel.FindOne(userId)
3. 按字段非空更新（userName、userPhoneNumber）
4. UsersModel.Update(user)
5. Cache.Del("jmall:user:detail:{userId}")
6. 返回 { code: "200" }
```

### 5.7 删除用户 `POST /users/deleteUserById` 🔒

```
1. ctxutil.UserIDFromCtx(ctx) → userId
2. UsersModel.Delete(userId)
3. Cache.Del("jmall:user:detail:{userId}")
4. 返回 { code: "200" }
```

### 5.8 判断是否为管理员 `POST /users/isManager` 🔒

```
1. ctxutil.UserIDFromCtx(ctx) → userId
2. UsersModel.FindOne(userId) → 得到 userName
3. SysManagerModel.FindOneBySysname(userName)
   - 找到 → { isManager: true }
   - ErrNotFound → { isManager: false }
```

> 管理员判定逻辑：users 表中的 user_name 与 sysmanager 表中的 sysname 相同，即为管理员。两张表密码独立，互不关联。

---

## 6. 商品服务

商品服务所有接口**无需登录**，均为只读（管理操作在 management 服务）。

### 6.1 获取全部商品 `POST /product/getAllProduct`

```
Cache key: jmall:products:all (TTL 10 min)

hit  → 返回缓存
miss → ProductModel.FindAll()
     → 序列化写入缓存
     → 返回 []ProductItem
```

### 6.2 获取分类列表 `POST /product/getCategory`

```
Cache key: jmall:categories (TTL 10 min)

hit  → 返回缓存
miss → CategoryModel.FindAll()
     → 写入缓存
     → 返回 []CategoryItem{ categoryId, categoryName, categoryHot }
```

### 6.3 获取热门商品 `POST /product/getHotProduct`

```
Cache key: jmall:products:hot:7 (TTL 5 min)

hit  → 返回缓存
miss → ProductModel.FindTopHot(limit=7)
       SQL: SELECT ... FROM product ORDER BY product_hot DESC LIMIT 7
     → 写入缓存
     → 返回 []ProductItem (top 7)
```

### 6.4 获取促销推荐 `POST /product/getPromoProduct`

```
输入: { categoryName }

Cache key: jmall:products:promo:{categoryId} (TTL 5 min)

hit  → 返回缓存
miss → CategoryModel.FindOneByCategoryName(categoryName) → categoryId
     → ProductModel.FindTopHotByCategory(categoryId, limit=7)
       SQL: SELECT ... WHERE category_id=? ORDER BY product_hot DESC LIMIT 7
     → 写入缓存
     → 返回 []ProductItem
```

### 6.5 获取特价促销商品 `POST /product/getPromotionProduct`

```
Cache key: jmall:products:promotion:7 (TTL 5 min)

hit  → 返回缓存
miss → ProductModel.FindByIsPromotion(limit=7)
       SQL: SELECT ... WHERE product_isPromotion > 0 ORDER BY product_isPromotion DESC LIMIT 7
     → 写入缓存
```

### 6.6 个性化推荐 `POST /product/getRecommendProduct`

```
Cache key: jmall:product:recommend:personal (TTL 5 min)

hit  → 返回缓存
miss → CategoryModel.FindAll()
     → 按 CategoryHot 降序排序，取第一个 → topCategory
     → ProductModel.FindTopHotByCategory(topCategory.CategoryId, 7)
     → 写入缓存
     → 返回 []ProductItem
```

> 推荐逻辑：哪个分类被加购/收藏次数最多（CategoryHot 最高），就推荐该分类下热度最高的 7 件商品。

### 6.7 按分类获取商品 `POST /product/getProductByCategory`

```
输入: { categoryId }

Cache key: jmall:products:category:{categoryId} (TTL 10 min)

miss → ProductModel.FindByCategory(categoryId)
       SQL: SELECT ... WHERE category_id=? ORDER BY product_sales DESC
```

### 6.8 商品详情 `POST /product/getProductDetail`

```
输入: { productId }

Cache key: jmall:product:detail:{productId} (TTL 5 min)

miss → ProductModel.FindOne(productId)
```

### 6.9 商品图片 `POST /product/getProductPictures`

```
输入: { productId }

Cache key: jmall:product:pictures:{productId} (TTL 5 min)

miss → ProductPictureModel.FindByProductId(productId)
       SQL: SELECT ... FROM product_picture WHERE product_id=?
     → 返回 []PictureItem{ productId, imgPath, intro }
```

### 6.10 搜索商品 `POST /product/getProductBySearch`

```
输入: { keyword }

无缓存（搜索词无限多，不适合缓存）

→ ProductModel.FindBySearch(keyword)
  SQL: SELECT ... FROM product
       WHERE product_name LIKE ?
          OR product_title LIKE ?
          OR product_intro LIKE ?
  参数: "%keyword%"（三个均相同）
```

### 6.11 分类快捷列表（手机/手机壳/充电器）

三个独立接口，逻辑完全相同：

| 接口 | Cache Key | 硬编码 category_id |
|------|-----------|-------------------|
| `GET /product/getPhoneList` | `jmall:product:phone:7` | 1 |
| `GET /product/getProtectingShellList` | `jmall:product:shell:7` | 5 |
| `GET /product/getChargerList` | `jmall:product:charger:7` | 7 |

```
Cache hit  → 返回缓存
Cache miss → ProductModel.FindTopHotByCategory(categoryId, limit=7)
           → 写入缓存（TTL 5 min）
```

---

## 7. 购物车服务

### 7.1 加入购物车 `POST /user/shoppingCart/addShoppingCart` 🔒

```
输入: { productId, num }

1. ctxutil.UserIDFromCtx → userId

2. ProductModel.FindOne(productId)
   - 不存在 → code "002"

3. 计算上限
   maxNum = floor(product.ProductNum / 2)
   （库存的一半，防止超买）

4. ShoppingcartModel.FindByUserAndProduct(userId, productId)

   ┌─ 已存在（更新数量）──────────────────────────────────┐
   │  addNum = max(req.Num, 1)                           │
   │  newNum = existing.Num + addNum                     │
   │  if newNum > maxNum → code "003"（超出限制）         │
   │  ShoppingcartModel.UpdateNumByUserAndProduct(        │
   │      userId, productId, newNum)                     │
   └─────────────────────────────────────────────────────┘
   ┌─ 不存在（新增）──────────────────────────────────────┐
   │  addNum = min(max(req.Num, 1), maxNum)              │
   │  ShoppingcartModel.Insert({                         │
   │      UserId, ProductId, Num: addNum })              │
   └─────────────────────────────────────────────────────┘

5. 更新成功后 → 热度追踪
   CategoryModel.IncrCategoryHot(product.CategoryId)
   ProductModel.IncrProductHot(productId)

6. Cache.Del("jmall:cart:user:{userId}")

7. 返回 { code: "200" }
```

### 7.2 查看购物车 `POST /user/shoppingCart/getShoppingCart` 🔒

```
1. ctxutil.UserIDFromCtx → userId

2. Cache key: jmall:cart:user:{userId} (TTL 2 min)
   hit  → 返回缓存的 []CartItem
   miss → 继续

3. ShoppingcartModel.FindByUserId(userId)
   SQL: SELECT ... FROM shoppingcart WHERE user_id=?

4. 收集所有 productId → productIDs []int64

5. 批量查询商品（消除 N+1）
   ProductModel.FindByIds(productIDs)
   SQL: SELECT ... FROM product WHERE product_id IN (?,?,?...)
   → 构建 map[productId → {name, img, price, productNum}]

6. 组装响应
   for each cart row:
     maxNum = floor(product.ProductNum / 2)
     CartItem{ id, userId, productId, productName,
               productImg, price, num, maxNum }

7. 写入缓存（TTL 2 min）

8. 返回 []CartItem
```

### 7.3 更新购物车数量 `POST /user/shoppingCart/updateShoppingCart` 🔒

```
输入: { productId, num }

1. ctxutil.UserIDFromCtx → userId
2. num < 1 → code "002"

3. ShoppingcartModel.FindByUserAndProduct(userId, productId)
   - ErrNotFound → code "004"（商品不在购物车中）

4. ProductModel.FindOne(productId) → 得到 maxNum
   maxNum = floor(product.ProductNum / 2)

5. 校验
   - num == existing.Num → code "003"（数量未变化）
   - num > maxNum       → code "003"（超出限制）

6. ShoppingcartModel.UpdateNumByUserAndProduct(userId, productId, num)

7. Cache.Del("jmall:cart:user:{userId}")

8. 返回 { code: "200" }
```

### 7.4 删除购物车商品 `POST /user/shoppingCart/deleteShoppingCart` 🔒

```
输入: { productId }

1. ctxutil.UserIDFromCtx → userId
2. ShoppingcartModel.DeleteByUserAndProduct(userId, productId)
3. Cache.Del("jmall:cart:user:{userId}")
4. 返回 { code: "200" }
```

### 7.5 检查商品是否在购物车 `POST /user/shoppingCart/isExistShoppingCart` 🔒

```
输入: { productId }

1. ctxutil.UserIDFromCtx → userId
2. ShoppingcartModel.FindByUserAndProduct(userId, productId)
   - 找到 → { isExist: true }
   - ErrNotFound → { isExist: false }
```

---

## 8. 订单服务

### 8.1 创建订单 `POST /user/order/addOrder` 🔒

```
输入: { items: [{ productId, productNum, productPrice }] }
注意: productPrice 由前端传入但服务端不信任，会用 DB 真实售价覆盖

1. ctxutil.UserIDFromCtx → userId
2. len(items) == 0 → code "002"

3. 防重复提交（Redis SETNX）
   key: jmall:order:submit:{userId}，TTL=5s
   SETNX 失败 → code "012"（请勿重复提交）

4. 服务端价格校验
   ProductModel.FindByIds(所有 productId)
   → 用 product.ProductSellingPrice 替代前端传的 productPrice
   → len(products) != len(items) → code "013"（部分商品不存在）

5. Redis 预扣库存（Lua 原子操作）
   for each item:
     key: jmall:stock:{productId}
     Lua 脚本: GET → 比较 → DECRBY（原子执行）
     返回 -1（key 不存在）→ 从 DB 加载库存到 Redis，重试一次
     返回 0（库存不足）→ 回滚已扣减的商品 Redis 库存 → code "014"
     返回 1（成功）→ 记录到 deductedProducts

   Lua 脚本（为什么不用 GET + DECRBY 分开调用）：
   GET + DECRBY 不是原子操作，高并发下两个请求同时 GET 到库存=1，
   都认为够用，都去扣减，导致库存变成 -1（超卖）。
   Lua 在 Redis 中是单线程原子执行的。

6. 生成订单号
   orderId = time.Now().UnixMilli() * 1000 + rand.Intn(1000)

7. 数据库事务 ────────────────────────────────────────────┐
   OrdersModel.TransactCtx(ctx, func(ctx, session) {     │
     txOrders  = OrdersModel.WithSession(session)         │
     txCart    = ShoppingcartModel.WithSession(session)    │
     txProduct = ProductModel.WithSession(session)         │
                                                           │
     for each item:                                        │
       txOrders.Insert(&Orders{                            │
         OrderId, UserId, ProductId, ProductNum,           │
         ProductPrice: 服务端真实价格,                      │
         OrderTime, Status: 0（待支付）                    │
       })                                                  │
       txProduct.DecrStock(productId, num)                 │
         SQL: UPDATE product SET product_num =             │
              product_num - ? WHERE product_id = ?         │
              AND product_num >= ?                          │
         （WHERE product_num >= ? 防止超卖）               │
                                                           │
     for each item:                                        │
       txCart.DeleteByUserAndProduct(userId, productId)     │
                                                           │
   })  ──────────────────────────────────────────────────── ┘
   事务失败 → rollbackStock（Lua INCRBY 回滚 Redis 库存）

8. 设置订单超时标记
   key: jmall:order:expire:{orderId}，TTL=30min
   （配合定时任务关闭超时未支付订单）

9. Cache.Del("jmall:orders:user:{userId}", "jmall:cart:user:{userId}")

10. 返回 { code: "200", order_id: orderId }
```

> **双重库存保障**：Redis Lua 是快速拦截层（高并发下不打 DB），DB `WHERE product_num >= ?` 是最终一致性保障（即使 Redis 数据丢失也不会超卖）。事务失败时 Redis 库存自动回滚。

### 8.2 查看我的订单列表 `POST /user/order/getOrder` 🔒

```
1. ctxutil.UserIDFromCtx → userId

2. Cache key: jmall:orders:user:{userId} (TTL 2 min)
   hit  → 返回缓存的 []OrderGroup
   miss → 继续

3. OrdersModel.FindByUserId(userId)
   SQL: SELECT ... FROM orders WHERE user_id=? ORDER BY order_time DESC

4. 收集所有 productId → 批量查询
   ProductModel.FindByIds(allProductIds)
   → 构建 productMap

5. 按 order_id 分组（保持插入顺序）
   for each row:
     找到或创建 OrderGroup{ orderId, userId, status, orderTime }
     追加 OrderItem 到 group.Items
     累加 group.ItemCount += productNum
     累加 group.TotalAmount += productPrice * productNum

6. 写入缓存（TTL 2 min）
7. 返回 []OrderGroup
```

> **响应格式**：返回的是按 `order_id` 分组后的订单列表，每个 `OrderGroup` 包含：
> - 订单头信息：`order_id`、`user_id`、`status`、`order_time`
> - 计算字段：`item_count`（总件数）、`total_amount`（总金额）
> - 商品行数组：`items[]`，每项包含 `product_id`、`product_name`、`product_img`、`product_num`、`product_price`

### 8.3 订单详情 `POST /order/getDetails` 🔒

```
输入: { orderId }

1. OrdersModel.FindByOrderId(orderId)
   - 无记录 → code "002"
2. 收集 productId → ProductModel.FindByIds(productIds)
3. 组装单个 OrderGroup（含 items、item_count、total_amount）
4. 返回 { code: "200", order: OrderGroup }
```

### 8.4 删除订单 `POST /order/deleteOrderById` 🔒

```
输入: { orderId }

1. ctxutil.UserIDFromCtx → userId

2. OrdersModel.FindByOrderId(orderId)
   - 无记录 → code "002"

3. 权限检查
   rows[0].UserId != userId → error("forbidden")

4. 状态检查
   status == 1（已支付）→ code "005"（需先退款）

5. 库存回滚（仅待支付订单）
   status == 0 → 事务内：
     for each item: ProductModel.IncrStock(productId, num)
     OrdersModel.DeleteByOrderId(orderId)
     → 清理 Redis 库存缓存 jmall:stock:{productId}
     → 清理支付防重锁 jmall:payment:lock:{orderId}

   status == 2 或 3 → 直接删除（库存已在取消/退款时回滚）

6. Cache.Del("jmall:orders:user:{userId}")
7. 返回 { code: "200" }
```

### 8.5 订单状态流转

```
0 待支付 ──支付成功──→ 1 已支付 ──退款──→ 3 已退款
   │                      │
   ├──支付失败/过期──→ 2 已取消    └──→ (不可删除，需先退款)
   │
   └──用户删除──→ (回滚库存后删除)
```

---

## 9. 收藏服务

### 9.1 添加收藏 `POST /user/collect/addCollect` 🔒

```
输入: { productId }

1. ctxutil.UserIDFromCtx → userId

2. ProductModel.FindOne(productId)
   - 不存在 → code "002"

3. 幂等性检查（防止重复收藏导致热度重复计数）
   CollectModel.FindByUserAndProduct(userId, productId)
   - 已收藏 → 直接返回 { code: "200" }，不做任何写操作

4. 热度追踪
   CategoryModel.IncrCategoryHot(product.CategoryId)
   ProductModel.IncrProductHot(productId)

5. CollectModel.Insert({
     UserId:      userId,
     ProductId:   productId,
     Category:    product.CategoryId,
     CollectTime: time.Now(),
   })

6. 失效相关缓存
   Cache.Del(
     "jmall:collect:user:{userId}",
     "jmall:products:hot:7",
     "jmall:product:recommend:personal",
   )

7. 返回 { code: "200" }
```

> **幂等保证**：同一用户对同一商品收藏两次，热度只计一次，DB 只写一行。

### 9.2 获取收藏列表 `POST /user/collect/getCollect` 🔒

```
1. ctxutil.UserIDFromCtx → userId

2. Cache key: jmall:collect:user:{userId} (TTL 2 min)
   hit  → 返回缓存
   miss → 继续

3. CollectModel.FindByUserId(userId)
   SQL: SELECT ... FROM collect WHERE user_id=?

4. CategoryModel.FindOne(collect.Category) → categoryName（逐行查询）

5. 组装 []CollectItem{
     id, userId, productId,
     category(分类名字符串),
     collectTime("2006-01-02 15:04:05"),
   }

6. 写入缓存（TTL 2 min）
7. 返回 []CollectItem
```

### 9.3 删除收藏 `POST /user/collect/deleteCollect` 🔒

```
输入: { productId }

1. ctxutil.UserIDFromCtx → userId
2. CollectModel.DeleteByUserAndProduct(userId, productId)
3. Cache.Del("jmall:collect:user:{userId}")
4. 返回 { code: "200" }
```

---

## 10. 管理服务

所有接口均需携带管理员 JWT（与普通用户 token 结构相同，但 `IsManager` 会额外校验 sysmanager 表）。

### 10.1 轮播图 `POST /resources/carousel`（公开）

```
Cache key: jmall:carousel (TTL 10 min)

miss → CarouselModel.FindAll()
     → 返回 []CarouselItem{ carouselId, imgPath, describes }
```

### 10.2 获取全部订单 `POST /management/getAllOrders` 🔒

```
无缓存（管理员实时查看）

OrdersModel.FindAllWithDetails()
SQL:
  SELECT o.id, o.order_id, o.user_id, o.product_id,
         o.product_num, o.product_price, o.order_time, o.status,
         u.user_name,
         p.product_name,
         COALESCE(p.product_picture, '') AS product_picture
  FROM orders o
  JOIN users u   ON o.user_id   = u.user_id
  JOIN product p ON o.product_id = p.product_id
  ORDER BY o.order_time DESC

→ 返回 []MgmtOrderItem（含用户名、商品名、商品图片、订单状态，单次 JOIN 查询，无 N+1）
```

> **前端分组**：管理端返回的仍然是扁平的订单行列表（因为管理员需要看到每一行的详细信息），
> 前端 `OrdersManage.vue` 在 `computed` 中按 `order_id` 分组为订单卡片展示。

### 10.3 按用户名查订单 `POST /management/getOrdersByUserName` 🔒

```
输入: { userName }

1. UsersModel.FindOneByUserName(userName)
   - ErrNotFound → 返回空列表 { code: "200", orders: [] }

2. OrdersModel.FindByUserId(userId)

3. 收集所有 productId → ProductModel.FindByIds(productIds)
   → 构建 productMap（含商品名和商品图片）

4. 组装 []MgmtOrderItem（含 user_name、product_name、product_picture、status）
```

### 10.4 获取全部用户 `POST /management/getAllUsers` 🔒

```
UsersModel.FindAll()
SQL: SELECT ... FROM users

→ 返回 []MgmtUserItem{ userId, userName, phoneNumber }
```

### 10.5 商品管理

#### 添加商品 `POST /management/addProduct` 🔒

```
输入: { productName, categoryId, productTitle, productIntro,
        productPicture, productPrice, productSellingPrice,
        productNum, productIsPromotion }

1. ProductModel.Insert(全部字段)

2. 批量失效商品列表缓存（6 个 key）：
   jmall:products:all
   jmall:products:hot:7
   jmall:products:promotion:7
   jmall:product:recommend:personal
   jmall:product:phone:7
   jmall:product:shell:7
   jmall:product:charger:7

3. 返回 { code: "200" }
```

#### 删除商品 `POST /product/deleteProductById` 🔒

```
输入: { productId }

1. ProductModel.Delete(productId)

2. 批量失效（9 个 key）：
   jmall:product:detail:{productId}
   jmall:product:pictures:{productId}
   + 上述 7 个列表缓存

3. 返回 { code: "200" }
```

#### 更新商品 `POST /product/updateProduct` 🔒

```
输入: { productId, productName?, productTitle?, productIntro?,
        productPicture?, productPrice?, productNum?,
        productIsPromotion? }

1. ProductModel.FindOne(productId) → 获取当前数据
2. 按字段非零值 patch（零值不覆盖，即无法将价格或库存改为 0）
3. ProductModel.Update(product)
4. 失效 detail 缓存 + 7 个列表缓存
5. 返回 { code: "200" }
```

### 10.6 搭配购（组合商品）管理

#### 获取全部折扣组合 `POST /management/getAllDiscounts` 🔒

```
CombinationProductModel.FindAll()
SQL: SELECT ... FROM combination_product

→ 返回 []CombinationItem{
    id, mainProductId, viceProductId,
    amountThreshold,      // 满N件
    priceReductionRange,  // 减M元
  }
```

#### 添加组合 `POST /management/addProductCombination` 🔒

```
输入: { mainProductId, viceProductId, amountThreshold, priceReductionRange }

CombinationProductModel.Insert(全部字段)
返回 { code: "200" }
```

#### 删除组合 `POST /management/deleteProductCombinationById` 🔒

```
输入: { id }

CombinationProductModel.Delete(id)
返回 { code: "200" }
```

### 10.7 按分类名查商品 `POST /management/getProductsByCategoryName` 🔒

```
输入: { categoryName }

1. CategoryModel.FindOneByCategoryName(categoryName)
2. ProductModel.FindByCategory(categoryId)
   SQL: SELECT ... WHERE category_id=? ORDER BY product_sales DESC
3. 返回 []MgmtProductItem
```

### 10.8 重置分类热度 `POST /management/setCategoryHotZero` 🔒

```
CategoryModel.ResetAllCategoryHot()
SQL: UPDATE category SET category_hot = 0

（management 服务版本仅重置 DB，不失效缓存；
 product 服务版本同时失效 8 个缓存 key——见下方注意事项）

返回 { code: "200" }
```

---

## 11. 支付服务

支付服务管理完整的支付生命周期：创建支付单 → 渠道下单 → 回调/确认 → 订单联动 → 退款。

### 11.1 渠道抽象（Strategy 模式）

```go
type PayChannel interface {
    Name() string
    CreatePayment(ctx, req) (*PayResponse, error)
    QueryPayment(ctx, paymentNo) (success, tradeNo, error)
    Refund(ctx, req) (*RefundResponse, error)
    VerifyNotify(ctx, params) (bool, error)
}
```

```
channel.Registry（全局注册中心）
├── "mock"    → MockChannel（init() 自动注册，开发测试用）
├── "wechat"  → TODO: WechatChannel
└── "alipay"  → TODO: AlipayChannel
```

业务逻辑通过 `channel.Get(name)` 获取实例，完全不感知具体渠道。新增渠道只需实现接口 + `init()` 注册，logic 层零改动。

### 11.2 创建支付单 `POST /payment/create` 🔒

```
输入: { order_id, channel }

1. ctxutil.UserIDFromCtx → userId

2. channel.Get(req.Channel)
   不存在 → code "002"

3. OrdersModel.FindByOrderId(orderId)
   不存在 → code "003"

4. 校验归属
   orderItems[0].UserId != userId → code "004"

5. 校验订单状态
   status != 0（待支付）→ code "011"

6. 防重复支付（Redis SETNX）
   key: jmall:payment:lock:{orderId}，TTL=支付过期时间（默认 30min）
   SETNX 失败 → code "005"

7. 计算金额
   sum(item.ProductPrice * item.ProductNum) → 转为分（int64）

8. 生成支付流水号
   paymentNo = "PAY" + UnixMilli + 3位随机数

9. 调用渠道预下单
   channel.CreatePayment(paymentNo, orderId, amount, notifyUrl)
   失败 → 释放 Redis 锁 → code "006"

10. 写入 payment_order 表
    失败 → 释放 Redis 锁

11. 返回 { code: "200", payment_no, pay_url }
```

### 11.3 支付回调 `POST /payment/notify`（无需鉴权）

```
输入: { payment_no, channel_trade_no, status, amount, paid_time, sign }

安全校验：
  1. 渠道验签 channel.VerifyNotify() → 防伪造回调
  2. 金额校验 req.Amount == payment.Amount → 防金额篡改

三层幂等保障：
  第一层: Redis SETNX jmall:payment:notify:{paymentNo} TTL=24h
         → O(1) 快速拦截重复回调
  第二层: SQL UPDATE ... WHERE status IN (0, 1)
         → 已成功的单不会被重复更新
  第三层: MySQL 事务
         → 支付单 + 订单原子更新

支付成功流程：
  1. Redis SETNX 幂等锁
     已存在 → 直接返回 200
  2. 查询支付单 → 终态检查
  3. 渠道验签 + 金额校验
  4. 过期检查 → 过期则关闭支付单 + 取消订单 + 回滚库存
  5. 事务 {
       UPDATE payment_order SET status=2 WHERE status IN (0,1)
       UPDATE orders SET status=1
     }
     失败 → 删除幂等锁（允许渠道重试）
  6. 清理 jmall:payment:lock:{orderId} + 用户缓存

支付失败流程：
  1. 更新支付单 status=3（失败）
  2. 事务 {
       UPDATE orders SET status=2（已取消）
       UPDATE product SET product_num = product_num + ?（回滚库存）
     }
  3. 清理 Redis 库存缓存 jmall:stock:{productId}
  4. 清理防重锁 + 用户缓存
```

> **为什么 Redis + DB 双重幂等？** Redis SETNX 是第一道防线，微信/支付宝短时间内可能重试多次回调，Redis O(1) 快速拦截。但 Redis 非持久化，宕机恢复后 key 可能丢失，DB 的 `WHERE status IN (0,1)` 是最终一致性保障。

### 11.4 Mock 支付确认 `POST /payment/mock/pay`（无需鉴权）

```
输入: { payment_no }

模拟用户在第三方支付页面完成支付。
内部走和真实回调完全相同的逻辑：
  幂等检查 → 状态校验 → 过期检查（含库存回滚）
  → 事务更新 → 清理锁和缓存

与 PaymentNotify 的区别：跳过验签和金额校验（Mock 渠道不需要）。
```

### 11.5 查询支付状态 `POST /payment/status` 🔒

```
输入: { payment_no }

PaymentOrderModel.FindByPaymentNo(paymentNo)
→ 返回 { paymentNo, orderId, amount, channel, status, paidTime }
```

### 11.6 用户支付记录 `POST /payment/list` 🔒

```
输入: { user_id }（实际从 JWT context 取 userId）

PaymentOrderModel.FindByUserId(userId)
→ 返回 []PaymentItem{ paymentNo, orderId, amount, channel, status, createdAt }
```

### 11.7 退款 `POST /payment/refund` 🔒

```
输入: { payment_no, refund_amount, reason }

1. 查询支付单 + 校验归属

2. 校验状态
   status != PaymentStatusSuccess → code "008"

3. 退款幂等（Redis SETNX）
   key: jmall:refund:lock:{paymentNo}，TTL=30s
   SETNX 失败 → code "016"

4. 校验退款金额
   amount <= 0 || amount > payment.Amount → 清锁 → code "009"

5. 调用渠道退款
   channel.Refund() → 失败则清锁 → code "010"

6. 事务 {
     INSERT payment_refund（退款单）
     UPDATE payment_order SET status=5（已退款）
     UPDATE orders SET status=3（已退款）
     for each orderItem:
       UPDATE product SET product_num = product_num + ?（回滚库存）
   }
   失败 → 清锁

7. 清理缓存
   jmall:orders:user:{userId}
   jmall:payment:user:{userId}
   jmall:refund:lock:{paymentNo}
   jmall:stock:{productId}（每个涉及商品）

8. 返回 { code: "200", refund_no }
```

> **所有失败路径都会清理退款幂等锁**，确保用户可以重试。

### 11.8 从 Mock 升级到真实支付

微信支付：
1. 申请商户号 → 获取 AppID、MchID、APIKey、证书
2. 实现 `WechatChannel`（模板在 `channel/wechat.go`）
3. 推荐 SDK：`github.com/wechatpay-apiv3/wechatpay-go`
4. `init()` 中注册 → 前端传 `channel: "wechat"` 即可

支付宝：
1. 开放平台创建应用 → 获取 AppID、私钥、支付宝公钥
2. 实现 `AlipayChannel`（模板在 `channel/alipay.go`）
3. 推荐 SDK：`github.com/smartwalle/alipay/v3`
4. `init()` 中注册 → 前端传 `channel: "alipay"` 即可

**业务逻辑层（logic/）完全不需要改动。**

---

## 12. 热度系统

### 热度来源

```
用户操作             → 触发热度增量
─────────────────────────────────────────────
AddCart（成功写入）  → CategoryHot+1, ProductHot+1
AddCollect（首次）   → CategoryHot+1, ProductHot+1
```

两个计数器分别存在 `category.category_hot` 和 `product.product_hot` 字段中。

### 热度用途

| 热度字段 | 用于哪些接口 |
|----------|-------------|
| `product_hot` | GetHotProduct、GetPromoProduct、GetAllUserRecommend、GetPhoneList、GetProtectingShellList、GetChargerList |
| `category_hot` | GetRecommendProduct（选热度最高的分类，再从该分类取热商品） |

### 热度重置

`SetCategoryHotZero` 将 `category_hot` 全部归零，会影响个性化推荐的结果（推荐将退化为热度相同时按 ID 顺序的首个分类）。

---

## 13. 数据库表结构

### users

| 字段 | 类型 | 说明 |
|------|------|------|
| `user_id` | INT AUTO_INCREMENT | 主键 |
| `user_name` | VARCHAR | 唯一用户名 |
| `password` | VARCHAR | bcrypt 哈希 |
| `user_phone_number` | VARCHAR | 手机号（可为空） |

### sysmanager

| 字段 | 类型 | 说明 |
|------|------|------|
| `sys_id` | INT AUTO_INCREMENT | 主键 |
| `sysname` | VARCHAR | 管理员用户名（对应 users.user_name） |
| `syspassword` | VARCHAR | 管理员密码（独立，未使用） |
| `user_phone_number` | VARCHAR | 可为空 |

### category

| 字段 | 类型 | 说明 |
|------|------|------|
| `category_id` | INT AUTO_INCREMENT | 主键 |
| `category_name` | VARCHAR | 分类名 |
| `category_hot` | INT NULL | 热度计数器 |

种子数据：1=手机, 2=电视机, 3=笔记本, 4=平板, 5=手机壳, 6=耳机, 7=充电器

### product

| 字段 | 类型 | 说明 |
|------|------|------|
| `product_id` | INT AUTO_INCREMENT | 主键 |
| `product_name` | VARCHAR | 商品名 |
| `category_id` | INT | 外键 → category |
| `product_title` | VARCHAR | 副标题 |
| `product_intro` | TEXT | 简介 |
| `product_picture` | VARCHAR NULL | 主图路径 |
| `product_price` | DECIMAL | 原价 |
| `product_selling_price` | DECIMAL | 售价 |
| `product_num` | INT | 库存 |
| `product_sales` | INT NULL | 销量 |
| `product_isPromotion` | INT | 特价标志（>0 为特价） |
| `product_hot` | INT NULL | 热度计数器 |

### orders

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INT AUTO_INCREMENT | 行 ID（非业务 ID） |
| `order_id` | BIGINT | 逻辑订单号（同一笔订单多行共享） |
| `user_id` | INT | 外键 → users |
| `product_id` | INT | 外键 → product |
| `product_num` | INT | 购买数量 |
| `product_price` | DECIMAL | 下单时价格快照（服务端校验后写入） |
| `order_time` | BIGINT | 下单时间（unix 秒） |
| `status` | TINYINT DEFAULT 0 | 0=待支付 1=已支付 2=已取消 3=已退款 |

### shoppingcart

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INT AUTO_INCREMENT | 主键 |
| `user_id` | INT | 外键 → users |
| `product_id` | INT | 外键 → product |
| `num` | INT | 数量 |

### collect

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INT AUTO_INCREMENT | 主键 |
| `user_id` | INT | 外键 → users |
| `product_id` | INT | 外键 → product |
| `category` | INT | 冗余存储分类 ID |
| `collect_time` | DATETIME | 收藏时间 |

### combination_product

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INT AUTO_INCREMENT | 主键 |
| `main_product_id` | INT | 主商品 |
| `vice_product_id` | INT | 搭配商品 |
| `amount_threshold` | INT NULL | 满 N 件触发 |
| `price_reduction_range` | INT NULL | 减 M 元 |


### payment_order

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | BIGINT AUTO_INCREMENT | 主键 |
| `payment_no` | VARCHAR(64) UNIQUE | 支付流水号（全局唯一，对外交互用） |
| `order_id` | BIGINT | 关联业务订单 ID |
| `user_id` | BIGINT | 用户 ID |
| `amount` | BIGINT | 支付金额（单位：分，避免浮点精度问题） |
| `channel` | VARCHAR(32) | 支付渠道：mock / wechat / alipay |
| `channel_trade_no` | VARCHAR(128) | 第三方交易号（回调时回填） |
| `status` | TINYINT DEFAULT 0 | 0=待支付 1=支付中 2=成功 3=失败 4=已关闭 5=已退款 |
| `expire_time` | BIGINT | 支付过期时间（unix 秒） |
| `paid_time` | BIGINT | 实际支付时间（unix 秒） |
| `notify_url` | VARCHAR(256) | 回调通知 URL |
| `extra` | TEXT | 扩展字段（JSON） |
| `created_at` | BIGINT | 创建时间 |
| `updated_at` | BIGINT | 更新时间 |

索引：`uk_payment_no`（唯一）、`idx_order_id`、`idx_user_id`、`idx_status`、`idx_expire_time`

### payment_refund

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | BIGINT AUTO_INCREMENT | 主键 |
| `refund_no` | VARCHAR(64) UNIQUE | 退款流水号 |
| `payment_no` | VARCHAR(64) | 关联支付流水号 |
| `order_id` | BIGINT | 关联订单 ID |
| `user_id` | BIGINT | 用户 ID |
| `refund_amount` | BIGINT | 退款金额（分） |
| `reason` | VARCHAR(256) | 退款原因 |
| `channel` | VARCHAR(32) | 退款渠道 |
| `channel_refund_no` | VARCHAR(128) | 第三方退款单号 |
| `status` | TINYINT DEFAULT 0 | 0=退款中 1=退款成功 2=退款失败 |
| `created_at` | BIGINT | 创建时间 |
| `updated_at` | BIGINT | 更新时间 |

索引：`uk_refund_no`（唯一）、`idx_payment_no`、`idx_order_id`


---

## 14. AI 智能助手服务

AI 智能助手是一个嵌入在商城所有页面右下角的悬浮聊天组件，用户可以通过自然语言与之对话，查询商品信息、价格、促销活动、组合优惠等。后端通过 MCP（Model Context Protocol）工具协议让豆包大模型能够实时查询数据库，将结构化数据与自然语言生成能力结合，为用户提供智能购物助手体验。

### 14.1 整体架构

```
用户输入消息
    │
    ▼
前端 AiChat.vue 组件
    │
    ├─ Mock 模式 → Axios POST /api/aichat/chat → Mock 拦截器（本地关键词匹配）
    │
    └─ 正常模式 → fetch SSE POST /api/aichat/stream
                      │
                      ▼ (Nginx/DevServer proxy rewrite: /api → /)
                aichat-api :8888
                      │
                      ▼
              ┌───────────────┐
              │  ChatStream   │
              │    Logic      │
              └───────┬───────┘
                      │
          ┌───────────┼───────────┐
          ▼           ▼           ▼
    豆包大模型 API   MCP 工具层   MySQL
    (chat/completions) (tools.go)  (storedb)
          │           │
          │    ┌──────┴──────┐
          │    │ 工具调用结果  │
          │    └──────┬──────┘
          │           │
          ▼           ▼
    模型生成最终回复（流式 SSE）
          │
          ▼
    前端逐字渲染
```

核心流程分为三个阶段：

1. **意图理解**：用户消息 + System Prompt + MCP 工具定义 → 发送给豆包大模型
2. **数据查询**：模型判断需要调用哪些工具 → 后端执行工具查询数据库 → 结果回传给模型
3. **回复生成**：模型基于工具返回的真实数据生成自然语言回复 → 流式推送给前端

### 14.2 后端服务结构

```
backend/service/aichat/
├── aichat.go                          # 服务入口
├── etc/
│   └── aichat-api.yaml                # 配置文件
└── internal/
    ├── config/
    │   └── config.go                  # 配置结构体（含 DoubaoConfig）
    ├── handler/
    │   ├── routes.go                  # 路由注册
    │   ├── chathandler.go             # 非流式接口 Handler
    │   └── chatstreamhandler.go       # 流式接口 Handler
    ├── logic/
    │   ├── chatlogic.go               # 非流式聊天逻辑
    │   ├── chatstreamlogic.go         # 流式聊天逻辑（SSE）
    │   └── doubao.go                  # 豆包 API 客户端（callDoubao / streamDoubao）
    ├── mcp/
    │   └── tools.go                   # MCP 工具定义与执行
    ├── svc/
    │   └── servicecontext.go          # 服务上下文（DB Model + Cache）
    └── types/
        └── types.go                   # 请求/响应类型
```

### 14.3 配置

```yaml
# etc/aichat-api.yaml
Name: aichat-api
Host: 0.0.0.0
Port: 8888

DB:
  DataSource: root:root@tcp(localhost:3306)/storedb?charset=utf8mb4&parseTime=True&loc=Local

Auth:
  Secret: jmall-secret-key-change-in-production
  ExpireSeconds: 86400

Cache:
  Addr: localhost:6379
  Password: ""
  DB: 0

Doubao:
  ApiKey: "your-doubao-api-key-here"       # 豆包 API Key（火山引擎控制台获取）
  Model: "doubao-1-5-pro-256k-250115"      # 模型 ID（Endpoint ID）
  BaseUrl: "https://ark.cn-beijing.volces.com/api/v3"  # 豆包 API 基础 URL
```

Docker 部署时通过环境变量 `DOUBAO_API_KEY` 注入，`docker-entrypoint.sh` 会自动替换 YAML 中的 `ApiKey` 字段。

### 14.4 API 接口

#### 14.4.1 非流式聊天 `POST /aichat/chat`

```
输入: { message: "有什么手机推荐？" }

1. 构建消息列表
   messages = [
     { role: "system", content: systemPrompt },
     { role: "user",   content: req.Message },
   ]

2. 附加 MCP 工具定义（7 个工具，见 14.5 节）

3. 调用豆包 API（非流式）
   POST https://ark.cn-beijing.volces.com/api/v3/chat/completions
   {
     model: "doubao-1-5-pro-256k-250115",
     messages: [...],
     tools: [...],
     stream: false
   }

4. 检查响应中是否有 tool_calls
   ├─ 无 tool_calls（finish_reason == "stop"）
   │  → 直接返回 { code: "200", reply: choice.message.content }
   │
   └─ 有 tool_calls
      → 解析每个 tool_call 的 function.name 和 function.arguments
      → 调用 mcp.ExecuteTool() 执行数据库查询
      → 将工具结果以 { role: "tool", content: 查询结果JSON, tool_call_id: id } 追加到 messages
      → 重新调用豆包 API（最多循环 3 轮）

5. 返回 { code: "200", reply: "最终自然语言回复" }
```

#### 14.4.2 流式聊天 `POST /aichat/stream`（SSE）

```
输入: { message: "现在有什么促销活动？" }

响应头:
  Content-Type: text/event-stream
  Cache-Control: no-cache
  Connection: keep-alive

1. 构建 messages + tools（同非流式）

2. 第一阶段：工具调用（非流式，可能多轮）
   调用 callDoubao()（stream=false）检查是否需要工具调用
   ├─ 有 tool_calls
   │  → 向前端推送思考状态：data: {"thinking":"正在查询商品信息..."}
   │  → 执行工具 → 结果追加到 messages → 再次调用 callDoubao()
   │  → 最多循环 3 轮
   │
   └─ 无 tool_calls → 进入第二阶段

3. 第二阶段：流式生成（SSE）
   调用 streamDoubao()（stream=true）
   → 豆包 API 返回 SSE 流
   → 逐 chunk 解析 delta.content
   → 转发给前端：data: {"content":"每个文字片段"}
   → 前端实时渲染

4. 结束标记
   data: [DONE]
```

> **为什么工具调用阶段用非流式？** 工具调用需要完整的 `tool_calls` JSON 才能解析执行，流式模式下 `tool_calls` 的 `arguments` 字段会被分片传输，需要拼接后才能使用。为简化实现，工具调用阶段使用非流式请求，仅最终文本生成阶段使用流式，兼顾了实时性和可靠性。

### 14.5 MCP 工具协议

MCP（Model Context Protocol）工具层是连接大模型和数据库的桥梁。通过 OpenAI 兼容的 Function Calling 协议，将数据库查询能力以工具定义的形式暴露给大模型，模型根据用户意图自主决定调用哪些工具。

#### 工具定义

每个工具以 JSON Schema 格式定义，包含名称、描述和参数结构：

| 工具名 | 描述 | 参数 | 对应 DB 操作 |
|--------|------|------|-------------|
| `search_products` | 根据关键词搜索商品 | `keyword: string` | `ProductModel.FindBySearch(keyword)` |
| `get_categories` | 获取所有商品分类 | 无 | `CategoryModel.FindAll()` |
| `get_product_detail` | 根据 ID 获取商品详情 | `product_id: int` | `ProductModel.FindOne(id)` |
| `get_products_by_category` | 按分类获取商品列表 | `category_id: int` | `ProductModel.FindByCategory(id)` |
| `get_hot_products` | 获取热门商品排行 | `limit?: int`（默认 10） | `ProductModel.FindTopHot(limit)` |
| `get_promotion_products` | 获取促销商品列表 | `limit?: int`（默认 10） | `ProductModel.FindByIsPromotion(limit)` |
| `get_combination_discounts` | 获取组合优惠/满减信息 | 无 | `CombinationProductModel.FindAll()` + 关联查询商品名 |

#### 工具执行流程

```
豆包模型返回 tool_calls:
[
  {
    "id": "call_abc123",
    "type": "function",
    "function": {
      "name": "search_products",
      "arguments": "{\"keyword\":\"手机\"}"
    }
  }
]
    │
    ▼
mcp.ExecuteTool(ctx, svcCtx, "search_products", "{\"keyword\":\"手机\"}")
    │
    ▼
execSearchProducts()
    │
    ├─ 解析 arguments JSON → keyword = "手机"
    ├─ svcCtx.ProductModel.FindBySearch(ctx, "手机")
    │  SQL: SELECT ... FROM product
    │       WHERE product_name LIKE '%手机%'
    │          OR product_title LIKE '%手机%'
    │          OR product_intro LIKE '%手机%'
    ├─ 组装精简结果（只返回模型需要的字段）
    └─ 返回 JSON: [{"id":1,"name":"Redmi K30","price":1999,"selling_price":1599,...}]
    │
    ▼
结果以 tool message 追加到对话历史:
{
  "role": "tool",
  "content": "[{\"id\":1,\"name\":\"Redmi K30\",...}]",
  "tool_call_id": "call_abc123"
}
    │
    ▼
再次调用豆包 API → 模型基于真实数据生成自然语言回复
```

#### 工具返回数据格式

工具返回的 JSON 经过精简，只包含模型生成回复所需的关键字段，避免传输冗余数据：

```json
// search_products / get_products_by_category 返回格式
[
  {
    "id": 1,
    "name": "Redmi K30",
    "price": 1999,
    "selling_price": 1599,
    "stock": 100,
    "sales": 50,
    "is_promotion": 1
  }
]

// get_combination_discounts 返回格式（关联查询了商品名）
[
  {
    "main_product_id": 1,
    "main_product_name": "Redmi K30",
    "vice_product_id": 6,
    "vice_product_name": "小米USB充电器30W",
    "amount_threshold": 2,
    "price_reduction": 20
  }
]
```

### 14.6 豆包大模型 API 客户端

#### 非流式调用 `callDoubao()`

```
POST {BaseUrl}/chat/completions
Headers:
  Content-Type: application/json
  Authorization: Bearer {ApiKey}

Body:
{
  "model": "doubao-1-5-pro-256k-250115",
  "messages": [
    { "role": "system", "content": "你是 JMall 商城的 AI 智能购物助手..." },
    { "role": "user", "content": "有什么手机推荐？" }
  ],
  "tools": [
    { "type": "function", "function": { "name": "search_products", ... } },
    ...
  ],
  "stream": false
}

Response:
{
  "id": "chatcmpl-xxx",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "...",           // 文本回复（无工具调用时）
      "tool_calls": [...]         // 工具调用（需要查询数据时）
    },
    "finish_reason": "stop"       // "stop" = 文本完成, "tool_calls" = 需要执行工具
  }]
}
```

#### 流式调用 `streamDoubao()`

```
请求同上，但 "stream": true

响应为 SSE 流：
data: {"id":"chatcmpl-xxx","choices":[{"delta":{"content":"为"},"index":0}]}
data: {"id":"chatcmpl-xxx","choices":[{"delta":{"content":"你"},"index":0}]}
data: {"id":"chatcmpl-xxx","choices":[{"delta":{"content":"推荐"},"index":0}]}
...
data: [DONE]

后端逐行解析 → 提取 delta.content → 封装为前端 SSE 格式 → 转发
```

流式模式下 tool_calls 的 arguments 会被分片传输，后端通过 `toolCallArgBuilders` map 按 index 拼接完整参数。

### 14.7 System Prompt

```
你是 JMall 商城的 AI 智能购物助手。你可以帮助用户：
1. 搜索和查询商品信息（名称、价格、库存等）
2. 查看商品分类
3. 了解热门商品和促销活动
4. 查询组合优惠和满减信息
5. 提供购物建议和商品推荐

请用友好、专业的语气回答用户问题。当需要查询商品信息时，请使用提供的工具函数。
回答时请使用中文，并尽量提供具体的商品信息（如价格、库存等）。
如果用户问的问题与购物无关，请礼貌地引导用户回到购物相关话题。
```

System Prompt 定义了模型的角色边界和行为规范，确保模型：
- 知道自己是购物助手，会主动使用工具查询真实数据
- 用中文回复，提供具体的价格和库存信息
- 对非购物话题进行礼貌引导

### 14.8 前端实现

#### 组件结构

`AiChat.vue` 是一个全局悬浮组件，挂载在 `App.vue` 中，所有页面可见。

```
App.vue
└── <AiChat />
    ├── 悬浮按钮（FAB）── 右下角固定定位，点击展开/收起聊天窗口
    ├── 聊天窗口
    │   ├── Header（标题 + 最小化按钮）
    │   ├── Body（消息列表 + 欢迎消息 + 快捷建议）
    │   └── Footer（输入框 + 发送按钮）
    └── 加载指示器（打字动画 / 思考状态文字）
```

#### 双模式发送策略

组件根据环境变量 `VUE_APP_USE_MOCK` 自动切换发送模式：

```
sendMessage()
    │
    ├─ isMock === true
    │  → sendMockMessage()
    │  → this.$axios.post('/api/aichat/chat', { message })
    │  → Axios 拦截器捕获 → 返回 mock 数据
    │  → 逐字打字动画渲染（每 3 个字符暂停 30ms）
    │
    └─ isMock === false
       → sendStreamMessage()
       → fetch('/api/aichat/stream', { method: 'POST', body: { message } })
       → ReadableStream reader 逐 chunk 读取
       → 解析 SSE data 行
       │  ├─ {"thinking":"..."} → 显示思考状态
       │  ├─ {"content":"..."}  → 追加到消息气泡
       │  ├─ {"error":"..."}    → 显示错误信息
       │  └─ [DONE]             → 结束
       → 实时渲染，自动滚动到底部
```

> **为什么 Mock 模式不用 SSE？** Mock 系统基于 Axios 请求拦截器实现，通过 `config.adapter` 直接返回数据，不发出真实网络请求。而 `fetch` API 不经过 Axios，无法被拦截。因此 Mock 模式下改用 Axios 调用非流式接口 `/api/aichat/chat`，并通过前端逐字动画模拟流式效果。

#### Mock 智能回复

Mock 模式下的 AI 回复通过关键词匹配实现，覆盖以下场景：

| 用户输入关键词 | Mock 行为 |
|---------------|----------|
| 热门、推荐、火 | 返回前 5 个商品，含价格和优惠信息 |
| 促销、打折、优惠、便宜 | 筛选原价 > 售价的商品，计算折扣率 |
| 分类、类别、有什么 | 列出所有分类及各分类商品数量 |
| 手机 | 筛选 category_id=1 的商品 |
| 电视 | 筛选 category_id=2 的商品 |
| 充电、配件 | 筛选 category_id=7 的商品 |
| 价格、多少钱、贵 | 尝试匹配具体商品名，返回详细价格信息 |
| 其他 | 在商品名/标题/简介中模糊搜索，无匹配则返回功能引导 |

Mock 数据来源于 `frontend/src/mock/data.js` 中的 `products` 和 `categories` 数组，与其他 Mock 接口共享同一份数据。

### 14.9 请求路由与代理

#### 开发环境（vue.config.js devServer proxy）

```
前端请求: POST /api/aichat/stream
    → pathRewrite: '^/api' → ''
    → 转发到: http://localhost:8888/aichat/stream
```

#### Docker 环境（nginx.conf）

```nginx
location /api/aichat/ {
    rewrite ^/api/(.*)$ /$1 break;
    proxy_pass http://aichat:8888;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_buffering off;       # SSE 必须关闭缓冲
    proxy_cache off;           # 禁用缓存
    proxy_read_timeout 300s;   # 长连接超时（流式响应可能持续较久）
}
```

> SSE 代理的关键配置：`proxy_buffering off` 确保 Nginx 不缓冲后端响应，每个 chunk 立即转发给客户端；`proxy_read_timeout 300s` 防止长时间的流式响应被 Nginx 超时断开。

### 14.10 ServiceContext

```go
type ServiceContext struct {
    Config                  config.Config
    Cache                   *cache.Client
    ProductModel            model.ProductModel           // 商品查询
    CategoryModel           model.CategoryModel          // 分类查询
    CombinationProductModel model.CombinationProductModel // 组合优惠查询
}
```

aichat 服务复用了 product、category、combination_product 三个 Model，直接查询同一个 MySQL 数据库。不需要通过 RPC 调用其他微服务，避免了额外的网络开销。

### 14.11 多轮工具调用机制

模型可能需要多次工具调用才能回答一个复杂问题。例如用户问"手机分类下有什么促销商品"，模型可能：

```
第 1 轮: 调用 get_categories → 获取分类列表，找到"手机"的 category_id=1
第 2 轮: 调用 get_products_by_category(category_id=1) → 获取手机列表
第 3 轮: 无工具调用 → 基于数据生成最终回复
```

后端限制最多 3 轮工具调用，防止模型陷入无限循环。每轮的对话历史完整保留：

```
messages 演变过程:

[system, user]
    → 第 1 轮调用后: [system, user, assistant(tool_calls), tool(结果)]
    → 第 2 轮调用后: [system, user, assistant(tool_calls), tool(结果), assistant(tool_calls), tool(结果)]
    → 最终生成:      [system, user, ..., assistant(最终文本回复)]
```

### 14.12 部署配置

#### docker-compose.yml

```yaml
aichat:
  build:
    context: ./backend
    dockerfile: Dockerfile
    args:
      SERVICE: aichat
  container_name: jmall-aichat
  restart: unless-stopped
  ports:
    - "8888:8888"
  environment:
    DB_SOURCE: root:root@tcp(mysql:3306)/storedb?charset=utf8mb4&parseTime=True&loc=Local
    REDIS_ADDR: redis:6379
    DOUBAO_API_KEY: ${DOUBAO_API_KEY:-your-doubao-api-key-here}
  depends_on:
    mysql:
      condition: service_healthy
    redis:
      condition: service_healthy
```

#### docker-entrypoint.sh 注入逻辑

```sh
# AI chat service specific: inject Doubao API key
if [ -n "$DOUBAO_API_KEY" ]; then
  sed -i "s|ApiKey:.*|ApiKey: ${DOUBAO_API_KEY}|" "$CONFIG"
fi
```

#### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DOUBAO_API_KEY` | 豆包 API Key | `your-doubao-api-key-here`（需替换） |
| `DB_SOURCE` | MySQL 连接串 | 同其他服务 |
| `REDIS_ADDR` | Redis 地址 | 同其他服务 |

### 14.13 典型交互示例

```
用户: 有什么手机推荐？

→ 后端发送给豆包:
  messages: [system, user("有什么手机推荐？")]
  tools: [search_products, get_categories, ...]

→ 豆包返回 tool_calls:
  [{ function: { name: "search_products", arguments: '{"keyword":"手机"}' } }]

→ 后端执行 search_products("手机"):
  SQL: SELECT ... FROM product WHERE product_name LIKE '%手机%' OR ...
  返回: [
    {"id":1,"name":"Redmi K30","price":1999,"selling_price":1599,"stock":100,"sales":50},
    {"id":2,"name":"Redmi K30 5G","price":2599,"selling_price":2599,"stock":80,"sales":30},
    {"id":3,"name":"小米CC9 Pro","price":2799,"selling_price":2599,"stock":60,"sales":20}
  ]

→ 工具结果追加到 messages，再次调用豆包

→ 豆包生成最终回复（流式）:
  "为你推荐以下手机：
   1. **Redmi K30** - 120Hz流速屏，售价 ¥1599（原价 ¥1999，立省 ¥400），库存充足
   2. **Redmi K30 5G** - 双模5G，售价 ¥2599，库存 80 件
   3. **小米CC9 Pro** - 1亿像素五摄，售价 ¥2599（原价 ¥2799），库存 60 件
   需要了解某款手机的详细信息吗？"

→ 前端 SSE 逐字渲染
```


---

## 15. 智能凑单推荐服务

智能凑单推荐是购物车页面的核心增值功能。当用户购物车总价未达到满减门槛时，系统自动分析差额、购物车商品分类、全站热度数据，通过三种推荐策略综合评分，推荐最合适的"凑单"商品，帮助用户以最小额外消费达到满减优惠，提升客单价和转化率。

### 15.1 整体架构

```
购物车页面 ShoppingCart.vue
    │
    └── <FillUpRecommend /> 组件
            │
            ▼
    POST /api/recommend/fillup
            │ (Nginx/DevServer proxy rewrite: /api → /)
            ▼
    recommendation-api :8889
            │
            ▼
    AuthMiddleware（JWT 验证）
            │
            ▼
    FillUpHandler → FillUpLogic.FillUp()
            │
            ├─ 1. ShoppingcartModel.FindByUserId()        → 获取购物车
            ├─ 2. ProductModel.FindByIds()                 → 批量查商品详情
            ├─ 3. CombinationProductModel.FindAll()        → 获取满减规则
            ├─ 4. Cache.Get()                              → 尝试读缓存
            │
            ├─ 5. 三策略并行收集候选商品
            │     ├─ strategyPriceGap()                    → 差额精准推荐
            │     ├─ strategyAssociated()                   → 关联商品推荐
            │     └─ strategyHotSelling()                   → 热销商品推荐
            │
            ├─ 6. deduplicateAndRank()                     → 去重 + 综合评分排序
            ├─ 7. Cache.Set()                              → 写缓存（TTL 2min）
            │
            ▼
    JSON Response → 前端渲染推荐列表
```

### 15.2 后端服务结构

```
backend/service/recommendation/
├── recommendation.go                          # 服务入口
├── etc/
│   └── recommendation-api.yaml                # 配置文件（端口 8889）
└── internal/
    ├── config/
    │   └── config.go                          # 配置结构体
    ├── handler/
    │   ├── routes.go                          # 路由注册（含 AuthMiddleware）
    │   ├── filluphandler.go                   # 凑单推荐 Handler
    │   ├── guessyoulikehandler.go             # 猜你喜欢 Handler
    │   └── reportbehaviorhandler.go           # 行为上报 Handler
    ├── logic/
    │   ├── filluplogic.go                     # 凑单推荐算法（3 策略 + 评分排序）
    │   ├── guessyoulikelogic.go               # 猜你喜欢推荐引擎（4 路召回 + 排序 + 重排）
    │   ├── reportbehaviorlogic.go             # 用户行为上报
    │   └── itemcf.go                          # ItemCF 离线计算引擎
    ├── middleware/
    │   └── authmiddleware.go                  # JWT 认证中间件
    ├── svc/
    │   └── servicecontext.go                  # 服务上下文（依赖注入）
    └── types/
        └── types.go                           # 请求/响应类型定义
```

### 15.3 ServiceContext

```go
type ServiceContext struct {
    Config                  config.Config
    Cache                   *cache.Client
    AuthMiddleware          rest.Middleware
    ProductModel            model.ProductModel            // 商品查询（含新增的价格区间、分类批量查询）
    ShoppingcartModel       model.ShoppingcartModel       // 购物车查询
    OrdersModel             model.OrdersModel             // 订单历史
    CollectModel            model.CollectModel            // 收藏记录
    CombinationProductModel model.CombinationProductModel // 搭配购组合 / 满减规则
    CategoryModel           model.CategoryModel           // 商品分类
    UserBehaviorModel       model.UserBehaviorModel       // 用户行为日志（猜你喜欢）
    ProductSimilarityModel  model.ProductSimilarityModel  // 商品相似度（ItemCF）
}
```

> 注入了 8 个 Model。`UserBehaviorModel` 和 `ProductSimilarityModel` 是猜你喜欢推荐系统新增的，分别用于用户行为采集和 ItemCF 相似度查询。

### 15.4 API 接口

#### `POST /recommend/fillup` 🔒

```
输入: { user_id: int64 }（实际 userId 从 JWT context 取，请求体中的 user_id 不被信任）

输出:
{
  "code": "200",
  "cart_total": 7797.0,          // 购物车当前总价（售价 × 数量）
  "nearest_rule": {
    "threshold": 10000,          // 最近的未达标满减门槛
    "reduction": 1000            // 对应的减免金额
  },
  "gap": 2203.0,                 // 还差多少钱达到门槛
  "recommendations": [           // 推荐商品列表（最多 12 件，按综合评分降序）
    {
      "product_id": 25,
      "product_name": "小米USB充电器60W快充版（6口）",
      "category_id": 7,
      "product_title": "6口输出，USB-C输出接口",
      "product_picture": "public/imgs/accessory/charger-60w.png",
      "product_price": 129,
      "product_selling_price": 129,
      "product_sales": 0,
      "product_hot": 46,
      "recommend_reason": "关联配件推荐",   // 推荐理由标签
      "score": 78.5                         // 综合评分
    }
  ]
}
```

#### 响应码

| code | 含义 |
|------|------|
| `"200"` | 成功 |
| `"401"` | 未登录 / Token 无效 |

### 15.5 满减规则解析

满减规则从 `combination_product` 表动态提取，而非硬编码。

```
getPromotionTiers()
    │
    ▼
CombinationProductModel.FindAll()
    │
    ▼
遍历所有记录，提取 (amountThreshold, priceReductionRange) 去重
    │
    ▼
按 threshold 升序排列 → []promotionTier

示例（基于种子数据）:
  combination_product 表中存在:
    (amountThreshold=2000, priceReductionRange=200)  ← 多条记录
    (amountThreshold=3000, priceReductionRange=300)  ← 多条记录
  去重后得到两档:
    Tier 1: 满 2000 减 200
    Tier 2: 满 3000 减 300
```

#### 最近档位匹配

```
findNearestTier(tiers, cartTotal)
    │
    ▼
从低到高遍历档位:
  cartTotal < tier.Threshold → 返回该档位 + gap
  全部满足 → 返回 nil（已达到所有满减，无需凑单）

示例:
  cartTotal = 1800 → 命中 Tier 1，gap = 200
  cartTotal = 2500 → 命中 Tier 2，gap = 500
  cartTotal = 3500 → 返回 nil（已满足所有档位）
```

> **兜底机制**：如果 `combination_product` 表为空或查询失败，使用默认规则 `[{2000, 200}, {3000, 300}]`。

### 15.6 三种推荐策略（核心算法）

三种策略并行收集候选商品，每个候选商品携带策略基础分和推荐理由，最终由综合评分算法统一排序。

#### 策略 1：差额精准推荐（strategyPriceGap）

**目标**：推荐价格接近差额的商品，让用户加一件就能凑到满减门槛。

```
输入: gap（差额）, excludeIds（购物车已有商品 ID）

1. 计算价格区间
   minPrice = max(gap × 0.5, 1)    // 最低不低于 1 元
   maxPrice = gap × 1.5

   示例: gap=200 → 价格区间 [100, 300]

2. 查询商品
   ProductModel.FindByPriceRange(minPrice, maxPrice, excludeIds, limit=20)
   SQL: SELECT ... FROM product
        WHERE product_selling_price >= ? AND product_selling_price <= ?
          AND product_num > 0
          AND product_id NOT IN (购物车商品)
        ORDER BY product_hot DESC, product_sales DESC
        LIMIT 20

3. 计算策略基础分
   for each product:
     priceDiff = |product.SellingPrice - gap|
     priceScore = max(0, 100 - priceDiff/gap × 100)
     // 价格越接近 gap，分数越高（满分 100）

   示例:
     gap=200, 商品价格=199 → priceDiff=1, score=99.5
     gap=200, 商品价格=129 → priceDiff=71, score=64.5
     gap=200, 商品价格=300 → priceDiff=100, score=50.0

4. 返回 []scoredProduct{ product, reason="差额精准推荐", score }
```

**适用场景**：差额较小（几十到几百元），用户倾向于"加一件小东西就够了"。

#### 策略 2：关联商品推荐（strategyAssociated）

**目标**：推荐与购物车商品有逻辑关联的商品，提升购买合理性。

```
输入: cartProductIds, cartCategoryIds, excludeIds

分两个子策略:

── a) 搭配购推荐（combination_product 表）──────────────────
for each cartProductId:
  CombinationProductModel.FindByMainProductId(productId)
  → 获取搭配商品 ID（vice_product_id）
  → 排除已在购物车的
  → ProductModel.FindOne(viceProductId)
  → 检查库存 > 0
  → score = 90（搭配购给最高基础分）
  → reason = "搭配购推荐"

── b) 关联品类推荐 ─────────────────────────────────────────
关联品类映射（硬编码，基于商城实际商品关系）:
  手机(1)   → 保护套(5), 保护膜(6), 充电器(7), 充电宝(8)
  保护套(5) → 手机(1)
  保护膜(6) → 手机(1)
  充电器(7) → 手机(1)
  充电宝(8) → 手机(1)

从购物车分类 → 查找关联分类 → 去重
ProductModel.FindByCategoryIds(relatedCatIds, excludeIds, limit=15)
SQL: SELECT ... FROM product
     WHERE category_id IN (关联分类)
       AND product_num > 0
       AND product_id NOT IN (购物车商品)
     ORDER BY product_hot DESC, product_sales DESC
     LIMIT 15

for each product:
  score = 70 + product_hot × 0.5   // 基础 70 分 + 热度加成
  reason = "关联配件推荐"
```

**适用场景**：用户买了手机，推荐手机壳、充电器等配件；或者管理员在 `combination_product` 表配置了搭配购组合。

**避免推荐无关商品的机制**：
- 关联品类映射是手动维护的，不会出现"买手机推荐洗衣机"
- 搭配购来自管理员配置，业务合理性由运营保证
- 排除购物车已有商品，避免重复推荐

#### 策略 3：热销商品推荐（strategyHotSelling）

**目标**：全站热销商品兜底，保证推荐列表不为空，同时利用从众心理提升转化。

```
输入: excludeIds

ProductModel.FindTopHot(limit=20)
SQL: SELECT ... FROM product ORDER BY product_hot DESC LIMIT 20

for each product:
  排除购物车已有 + 库存为 0 的
  score = 50 + product_hot × 0.3   // 基础 50 分 + 热度加成
  reason = "热销推荐"
```

**适用场景**：当差额推荐和关联推荐结果不足时，热销商品作为兜底保证推荐列表有内容。

#### 三策略优先级对比

| 策略 | 基础分范围 | 适用场景 | 数据来源 |
|------|-----------|---------|---------|
| 搭配购推荐 | 90 | 管理员配置的组合商品 | `combination_product` 表 |
| 差额精准推荐 | 0–100 | 价格接近差额的商品 | `product` 表按价格区间查询 |
| 关联配件推荐 | 70+ | 购物车商品的关联品类 | `product` 表按分类查询 |
| 热销推荐 | 50+ | 全站热门兜底 | `product` 表按热度排序 |

> 搭配购基础分最高（90），因为这是运营人员精心配置的组合，业务价值最高。差额精准推荐的分数范围最大（0–100），价格完美匹配时可以超过搭配购。热销推荐基础分最低（50+），作为兜底策略。

### 15.7 综合评分排序算法（deduplicateAndRank）

三种策略收集的候选商品可能有重复（同一商品被多个策略命中），需要去重后按综合评分排序。

#### 去重规则

```
遍历所有候选商品:
  if product_id 已出现:
    保留分数更高的那个（替换）
  else:
    加入去重列表
```

#### 综合评分公式

```
finalScore = strategyScore × 0.4 + priceMatchScore × 0.4 + hotScore × 0.2
```

三个维度：

| 维度 | 权重 | 计算方式 | 含义 |
|------|------|---------|------|
| 策略基础分 | 40% | 各策略赋予的原始分数 | 推荐来源的可信度 |
| 价格匹配分 | 40% | 商品价格与差额的接近程度 | 凑单效率（加一件就够） |
| 热度分 | 20% | 归一化到 0–100（相对于候选池最高热度） | 大众认可度 |

#### 价格匹配分计算

```
if gap > 0:
  priceDiff = |product.SellingPrice - gap|
  priceMatchScore = max(0, 100 - priceDiff/gap × 80)

  // 惩罚机制：价格超过差额 2 倍的商品大幅降权
  if product.SellingPrice > gap × 2:
    priceMatchScore × = 0.3

else:
  priceMatchScore = 50   // 已满足满减时，价格匹配分统一给 50
```

**为什么要惩罚高价商品？** 凑单的核心目标是"以最小额外消费达到满减"。如果差额是 200 元，推荐一个 2000 元的商品虽然也能凑到，但违背了用户"省钱"的初衷。乘以 0.3 的惩罚系数让这类商品排到后面。

#### 热度归一化

```
maxHot = 候选池中最高的 product_hot 值（至少为 1，避免除零）
hotScore = (product.ProductHot / maxHot) × 100
```

#### 排序示例

假设 gap = 200，候选池中有以下商品：

| 商品 | 策略 | 策略分 | 售价 | 价格匹配分 | 热度 | 热度分 | 综合分 |
|------|------|-------|------|-----------|------|-------|-------|
| USB充电器60W | 关联配件 | 93 | 129 | 71.6 | 46 | 100 | 85.8 |
| 小米MIX3保护壳 | 差额精准 | 93.5 | 12.9 | 25.2 | 2 | 4.3 | 48.3 |
| 小米电视4A 32寸 | 热销 | 77.3 | 799 | 0 (×0.3) | 91 | 100 | 50.9 |

最终排序：USB充电器60W > 小米电视4A > 小米MIX3保护壳

> 充电器虽然策略分和差额精准推荐的保护壳接近，但价格匹配分和热度分都更高，综合评分胜出。电视虽然热度最高，但价格远超差额（799 > 200×2），价格匹配分被惩罚为 0，综合分被拉低。

### 15.8 缓存策略

```
Cache Key: jmall:recommend:fillup:{userId}:{cartTotal}
TTL: 2 分钟
写入时机: FillUp() 计算完成后
失效时机: TTL 自动过期
```

**为什么用 `cartTotal` 而非购物车内容哈希？**

购物车总价变化意味着购物车内容发生了变化（增删商品或修改数量），此时差额和推荐结果都会不同，旧缓存自然失效。使用总价作为 key 的一部分，比计算购物车内容哈希更简单高效，且 2 分钟的短 TTL 保证了数据新鲜度。

**缓存穿透防护**：购物车为空时直接返回空列表，不查询数据库也不写缓存。已达到所有满减档位时同样直接返回，避免无意义的推荐计算。

### 15.9 新增 Model 方法

为支持推荐算法，在 `ProductModel` 中新增了两个查询方法：

#### FindByPriceRange

```go
func (m *customProductModel) FindByPriceRange(
    ctx context.Context,
    minPrice, maxPrice float64,
    excludeIds []int64,
    limit int,
) ([]*Product, error)
```

```sql
SELECT ... FROM product
WHERE product_selling_price >= ?
  AND product_selling_price <= ?
  AND product_num > 0
  AND product_id NOT IN (?, ?, ...)   -- 排除购物车已有商品
ORDER BY product_hot DESC, product_sales DESC
LIMIT ?
```

#### FindByCategoryIds

```go
func (m *customProductModel) FindByCategoryIds(
    ctx context.Context,
    categoryIds []int64,
    excludeIds []int64,
    limit int,
) ([]*Product, error)
```

```sql
SELECT ... FROM product
WHERE category_id IN (?, ?, ...)
  AND product_num > 0
  AND product_id NOT IN (?, ?, ...)   -- 排除购物车已有商品
ORDER BY product_hot DESC, product_sales DESC
LIMIT ?
```

两个方法都：
- 过滤库存为 0 的商品（`product_num > 0`）
- 支持排除指定商品 ID（购物车已有的）
- 按热度 + 销量降序排列
- 支持 `excludeIds` 为空的情况（不加 NOT IN 子句）

### 15.10 完整调用链

```
用户打开购物车页面
    │
    ▼
ShoppingCart.vue 渲染 <FillUpRecommend /> 组件
    │
    ▼
FillUpRecommend.mounted() → fetchRecommendations()
    │
    ▼
POST /api/recommend/fillup { user_id: 1 }
    │ (vue.config.js proxy: /api/recommend → http://localhost:8889)
    ▼
recommendation-api :8889
    │
    ├─ AuthMiddleware
    │   → 解析 JWT → 注入 userId 到 context
    │
    ├─ FillUpHandler
    │   → 解析请求体 → 调用 FillUpLogic
    │
    └─ FillUpLogic.FillUp()
        │
        ├─ 1. ShoppingcartModel.FindByUserId(userId)
        │     → 获取购物车行 [{productId, num}, ...]
        │
        ├─ 2. ProductModel.FindByIds(cartProductIds)
        │     → 批量获取商品详情（消除 N+1）
        │     → 计算 cartTotal = Σ(sellingPrice × num)
        │     → 收集 cartCategoryIds
        │
        ├─ 3. getPromotionTiers()
        │     → CombinationProductModel.FindAll()
        │     → 提取去重满减档位 [{2000,200}, {3000,300}]
        │
        ├─ 4. findNearestTier(tiers, cartTotal)
        │     → 找到最近未达标档位 + 计算 gap
        │     → 全部达标 → 返回空推荐
        │
        ├─ 5. Cache.Get(jmall:recommend:fillup:{userId}:{cartTotal})
        │     → hit → 直接返回缓存结果
        │     → miss → 继续计算
        │
        ├─ 6. 三策略并行收集
        │     ├─ strategyPriceGap(gap, excludeIds)
        │     │   → ProductModel.FindByPriceRange()
        │     │   → 计算价格匹配分
        │     │
        │     ├─ strategyAssociated(cartProductIds, categoryIds, excludeIds)
        │     │   → CombinationProductModel.FindByMainProductId() × N
        │     │   → ProductModel.FindByCategoryIds(relatedCatIds)
        │     │
        │     └─ strategyHotSelling(excludeIds)
        │         → ProductModel.FindTopHot(20)
        │
        ├─ 7. deduplicateAndRank(candidates, gap, threshold)
        │     → 去重（保留高分）
        │     → 综合评分 = 策略分×0.4 + 价格匹配×0.4 + 热度×0.2
        │     → 降序排列，截取前 12 件
        │
        ├─ 8. Cache.Set(key, results, 2min)
        │
        └─ 9. 返回 FillUpResp
              {code, cartTotal, nearestRule, gap, recommendations}
    │
    ▼
前端 FillUpRecommend.vue 渲染
    ├─ 满减进度条（gap > 0 时显示）
    ├─ 推荐商品网格列表
    └─ "加入购物车" 按钮
        │
        ├─ isExistShoppingCart → 不在 → addShoppingCart
        │                     → 已在 → Vuex addShoppingCartNum
        ├─ $emit('cartUpdated') → 父组件刷新购物车
        └─ setTimeout → fetchRecommendations()（500ms 后刷新推荐）
```

### 15.11 前端实现

#### 组件结构

`FillUpRecommend.vue` 嵌入在 `ShoppingCart.vue` 的购物车列表下方，仅在购物车有商品时显示。

```
ShoppingCart.vue
├── 购物车商品列表
├── 底部结算栏
├── <FillUpRecommend @cartUpdated="reloadCart" />   ← 新增
│   ├── 满减进度条（promo-progress）
│   │   ├── 文字提示："还差 ¥XX 即可享受满N减M优惠"
│   │   └── el-progress 进度条（百分比 = cartTotal / threshold × 100）
│   ├── 已满足提示（promo-achieved，gap=0 时显示）
│   └── 推荐商品列表（recommend-list）
│       └── recommend-item × N
│           ├── 商品图片 + 名称 + 副标题
│           ├── 售价 + 推荐理由标签（el-tag）
│           └── "加入购物车" 按钮
└── 满减助手抽屉（原有功能，保留）
```

#### 响应式更新机制

```
watch: {
  getTotalPrice() {           // 监听 Vuex 中购物车总价变化
    this.fetchRecommendations()  // 自动重新获取推荐
  }
}
```

当用户在购物车中修改商品数量、删除商品、或通过推荐列表加入新商品时，Vuex 中的 `getTotalPrice` 会变化，触发推荐列表自动刷新。这实现了"动态更新差额"的需求。

#### 加入购物车交互流程

```
用户点击推荐商品的"加入购物车"按钮
    │
    ├─ 1. POST /api/user/shoppingCart/isExistShoppingCart
    │     检查商品是否已在购物车
    │
    ├─ 2a. 不在购物车（code "002"）
    │      → POST /api/user/shoppingCart/addShoppingCart
    │      → 成功 → $emit('cartUpdated') → 父组件重新加载购物车
    │
    ├─ 2b. 已在购物车
    │      → Vuex addShoppingCartNum(productId) → 本地数量 +1
    │      → $emit('cartUpdated')
    │
    ├─ 3. 从推荐列表中移除已添加的商品（即时反馈）
    │
    └─ 4. 500ms 后重新 fetchRecommendations()
          （等待后端购物车缓存更新，获取新的差额和推荐）
```

### 15.12 请求路由与代理

#### 开发环境（vue.config.js）

```javascript
'/api/recommend': {
  target: 'http://localhost:8889/',
  changeOrigin: true,
  pathRewrite: { '^/api': '' }
}
```

前端请求 `POST /api/recommend/fillup` → 代理到 `http://localhost:8889/recommend/fillup`

#### Docker 环境（nginx.conf）

```nginx
location /api/recommend/ {
    rewrite ^/api/(.*)$ /$1 break;
    proxy_pass http://recommendation:8889;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

### 15.13 部署配置

#### docker-compose.yml

```yaml
recommendation:
  build:
    context: ./backend
    dockerfile: Dockerfile
    args:
      SERVICE: recommendation
  container_name: jmall-recommendation
  restart: unless-stopped
  ports:
    - "8889:8889"
  environment:
    DB_SOURCE: root:root@tcp(mysql:3306)/storedb?charset=utf8mb4&parseTime=True&loc=Local
    REDIS_ADDR: redis:6379
  depends_on:
    mysql:
      condition: service_healthy
    redis:
      condition: service_healthy
```

#### 启动命令（本地开发）

```bash
cd backend
go run service/recommendation/recommendation.go -f service/recommendation/etc/recommendation-api.yaml
```

### 15.14 性能分析

#### 数据库查询次数（单次请求）

| 步骤 | 查询 | 次数 |
|------|------|------|
| 获取购物车 | `FindByUserId` | 1 |
| 批量查商品 | `FindByIds` | 1 |
| 获取满减规则 | `FindAll`（combination_product） | 1 |
| 策略1: 差额推荐 | `FindByPriceRange` | 1 |
| 策略2a: 搭配购 | `FindByMainProductId` × N + `FindOne` × M | N+M（N=购物车商品数） |
| 策略2b: 关联品类 | `FindByCategoryIds` | 1 |
| 策略3: 热销 | `FindTopHot` | 1 |
| **总计** | | **6 + N + M**（缓存 miss 时） |

典型场景（购物车 3 件商品，每件有 1 个搭配）：6 + 3 + 3 = 12 次查询。

#### 优化措施

1. **Redis 缓存**：2 分钟 TTL，相同购物车总价的重复请求直接返回缓存，0 次 DB 查询
2. **批量查询**：`FindByIds` 消除购物车商品的 N+1 问题
3. **结果截断**：每个策略最多返回 15–20 个候选，最终结果截取前 12 个
4. **库存过滤**：SQL 层面 `product_num > 0`，避免推荐无库存商品

### 15.15 进阶优化方向

#### 个性化推荐（基于用户历史）

`ServiceContext` 已注入 `OrdersModel` 和 `CollectModel`，可扩展第四种策略：

```
strategyPersonalized(userId, excludeIds):
  1. OrdersModel.FindByUserId(userId) → 获取历史购买的商品分类
  2. CollectModel.FindByUserId(userId) → 获取收藏的商品分类
  3. 合并分类偏好 → 按频次排序
  4. ProductModel.FindByCategoryIds(偏好分类, excludeIds, limit)
  5. score = 80 + 偏好频次加成
  6. reason = "猜你喜欢"
```

#### A/B 测试方案

在 `FillUpLogic` 中根据用户 ID 分桶：

```go
if userID % 2 == 0 {
    // A 组：当前综合评分公式
    finalScore = strategyScore*0.4 + priceMatch*0.4 + hot*0.2
} else {
    // B 组：提高价格匹配权重
    finalScore = strategyScore*0.3 + priceMatch*0.5 + hot*0.2
}
```

通过对比两组的凑单转化率（推荐商品被加入购物车的比例）和满减达成率，确定最优评分公式。

#### 推荐效果评估

在推荐商品被加入购物车时埋点记录 `recommend_reason`，统计指标：

| 指标 | 计算方式 | 目标 |
|------|---------|------|
| 推荐点击率 | 加入购物车次数 / 推荐展示次数 | > 5% |
| 策略转化率 | 各策略被采纳次数 / 各策略展示次数 | 对比优化 |
| 满减达成率 | 凑单后达到满减的用户数 / 看到推荐的用户数 | > 30% |
| 客单价提升 | 有推荐的订单均价 - 无推荐的订单均价 | 正向提升 |

---

## 16. 猜你喜欢推荐系统

猜你喜欢是首页底部的个性化推荐模块，参考淘宝/京东/拼多多的主流实现方式，采用工业级的"多路召回 → 排序 → 重排"三层流水线架构。系统通过采集用户的浏览、点击、加购、购买、收藏等行为数据，结合协同过滤算法和热度模型，为每个用户生成个性化的商品推荐列表。

### 16.1 整体架构

```
首页 Home.vue
    │
    └── <GuessYouLike /> 组件（瀑布流 + 无限滚动）
            │
            ├─ POST /api/recommend/guessYouLike    → 获取推荐列表
            └─ POST /api/recommend/reportBehavior   → 上报点击行为
            │
            ▼ (Nginx/DevServer proxy rewrite: /api → /)
    recommendation-api :8889
            │
            ▼
    AuthMiddleware（JWT 验证）
            │
            ▼
    GuessYouLikeLogic.GuessYouLike()
            │
            ├─ 1. Cache.Get()                              → 尝试读缓存（3min TTL）
            │
            ├─ 2. 用户画像构建
            │     ├─ UserBehaviorModel.FindUserProductBehaviors()   → 行为加权评分
            │     ├─ UserBehaviorModel.FindUserPreferredCategories() → 偏好分类
            │     └─ UserBehaviorModel.FindRecentProductIds()        → 已交互商品
            │
            ├─ 3. 多路召回（Recall Layer）
            │     ├─ recallByUserPreference()              → 用户偏好召回
            │     ├─ recallByItemCF()                      → ItemCF 召回
            │     ├─ recallByUserCF()                      → UserCF 召回
            │     └─ recallByHotSelling()                  → 热门兜底召回
            │
            ├─ 4. 排序层（Rank Layer）
            │     └─ rank()                                → 综合评分排序
            │
            ├─ 5. 重排层（Re-rank Layer）
            │     └─ rerank()                              → 分类打散 + 分页
            │
            ├─ 6. fillProductDetails()                     → 填充商品详情
            ├─ 7. Cache.Set()                              → 写缓存（TTL 3min）
            │
            ▼
    JSON Response → 前端瀑布流渲染
```

### 16.2 数据流：从用户行为到推荐结果

```
用户行为采集                          离线计算                        在线推荐
─────────────                    ──────────                    ──────────
浏览商品详情 ─┐                                                
点击推荐商品 ─┤                                                
加入购物车   ─┼─→ POST /recommend/reportBehavior               
购买商品     ─┤       │                                        
收藏商品     ─┘       ▼                                        
                 user_behavior 表                              
                      │                                        
                      ├──────────→ ItemCF 离线计算（定时任务）  
                      │               │                        
                      │               ▼                        
                      │          product_similarity 表         
                      │               │                        
                      └───────────────┼──→ 多路召回            
                                      │       │                
                                      │       ▼                
                                      │    排序 → 重排         
                                      │       │                
                                      │       ▼                
                                      │    Redis 缓存 3min     
                                      │       │                
                                      └───────┼──→ 推荐结果    
```

#### 行为类型与权重

| 行为类型 | 编码 | 权重 | 说明 |
|---------|------|------|------|
| 浏览 | 1 | 1.0 | 用户打开商品详情页时自动上报 |
| 点击 | 2 | 2.0 | 用户点击推荐列表中的商品时上报 |
| 加购 | 3 | 3.0 | 加入购物车时上报 |
| 收藏 | 5 | 4.0 | 收藏商品时上报 |
| 购买 | 4 | 5.0 | 下单成功时上报 |

权重越高表示用户对该商品的兴趣越强。购买权重最高（5.0），因为这是最强的正向信号。

### 16.3 API 接口

#### 16.3.1 猜你喜欢 `POST /recommend/guessYouLike` 🔒

```
输入: { page: 1, page_size: 20 }
  - page: 页码，默认 1
  - page_size: 每页数量，默认 20，最大 50

输出:
{
  "code": "200",
  "recommendations": [
    {
      "product_id": 9,
      "product_name": "小米电视4A 32英寸",
      "category_id": 2,
      "product_title": "人工智能系统，高清液晶屏",
      "product_picture": "public/imgs/appliance/MiTv-4A-32.png",
      "product_price": 799,
      "product_selling_price": 799,
      "product_sales": 0,
      "product_hot": 91,
      "recommend_reason": "猜你喜欢",
      "score": 82.35
    }
  ],
  "has_more": true
}
```

#### 16.3.2 上报用户行为 `POST /recommend/reportBehavior` 🔒

```
输入: {
  product_id: 1,
  category_id: 1,
  behavior_type: 1    // 1=浏览 2=点击 3=加购 4=购买 5=收藏
}

输出: { "code": "200" }
```

行为上报是异步的，即使写入失败也返回成功，不影响用户体验。前端在以下时机自动上报：
- 打开商品详情页 → 上报浏览（behavior_type=1）
- 点击推荐列表中的商品 → 上报点击（behavior_type=2）

### 16.4 四路召回策略（Recall Layer）

召回层的目标是从全量商品中快速筛选出用户可能感兴趣的候选集。四种策略并行执行，各自返回带基础分的候选商品列表。

#### 策略 1：用户偏好召回（recallByUserPreference）

**原理**：根据用户近 30 天的行为数据，计算用户偏好的商品分类，推荐该分类下热度最高的商品。

```
1. UserBehaviorModel.FindUserPreferredCategories(userId, 30天, limit=5)
   SQL: SELECT category_id FROM user_behavior
        WHERE user_id=? AND behavior_time > ?
        GROUP BY category_id
        ORDER BY SUM(行为权重) DESC
        LIMIT 5

2. ProductModel.FindByCategoryIds(偏好分类, 排除已交互商品, limit=30)

3. 基础分 = 80 + product_hot × 0.5
   推荐理由 = "猜你喜欢"
```

**适用场景**：有行为数据的老用户。如果用户最近频繁浏览手机，就推荐手机分类下的热门商品。

#### 策略 2：ItemCF 召回（recallByItemCF）

**原理**：基于商品相似度表（离线计算），找到用户最近交互过的商品的相似商品。核心假设是"喜欢商品 A 的用户也可能喜欢与 A 相似的商品 B"。

```
1. 取用户最近交互的前 10 个商品作为种子

2. ProductSimilarityModel.FindSimilarProductsByIds(种子商品, limit=30)
   SQL: SELECT * FROM product_similarity
        WHERE product_id IN (种子商品)
        ORDER BY score DESC
        LIMIT 30

3. 基础分 = 70 + similarity_score × 30
   推荐理由 = "相似商品推荐"
```

**适用场景**：有行为数据 + product_similarity 表有数据（需要离线计算）。

#### 策略 3：UserCF 召回（recallByUserCF）

**原理**：找到与当前用户行为相似的其他用户（共同交互商品数最多），推荐这些相似用户喜欢但当前用户没看过的商品。核心假设是"行为相似的用户有相似的偏好"。

```
1. UserBehaviorModel.FindSimilarUsers(userId, 30天, limit=10)
   SQL: SELECT b2.user_id
        FROM user_behavior b1
        JOIN user_behavior b2 ON b1.product_id = b2.product_id
             AND b2.user_id != b1.user_id
        WHERE b1.user_id=? AND b1.behavior_time > ? AND b2.behavior_time > ?
        GROUP BY b2.user_id
        ORDER BY COUNT(DISTINCT b1.product_id) DESC
        LIMIT 10

2. UserBehaviorModel.FindProductsByUsers(相似用户, 排除已交互商品, 30天, limit=30)

3. 基础分 = 60 + 行为权重 × 5
   推荐理由 = "和你口味相似的人也在看"
```

**适用场景**：有行为数据且系统中有足够多的用户行为交叉。

#### 策略 4：热门兜底召回（recallByHotSelling）

**原理**：全站热销商品 Top N，利用从众心理。这是冷启动的兜底策略，保证任何用户都能看到推荐内容。

```
1. ProductModel.FindTopHot(limit=30)
   SQL: SELECT ... FROM product ORDER BY product_hot DESC LIMIT 30

2. 排除已交互商品 + 库存为 0 的商品

3. 基础分 = 40 + product_hot × 0.3
   推荐理由 = "热门推荐"
```

**适用场景**：所有用户（新用户冷启动时只有这一路召回有结果）。

#### 四路召回优先级对比

| 策略 | 基础分范围 | 数据依赖 | 冷启动可用 |
|------|-----------|---------|-----------|
| 用户偏好召回 | 80+ | user_behavior 表 | ✗ |
| ItemCF 召回 | 70–100 | user_behavior + product_similarity | ✗ |
| UserCF 召回 | 60–85 | user_behavior（多用户交叉） | ✗ |
| 热门兜底 | 40–70 | product.product_hot | ✓ |

### 16.5 排序层（Rank Layer）

排序层对召回层返回的所有候选商品进行统一评分和排序。

#### 去重

同一商品可能被多个召回策略命中（如一个商品既是用户偏好分类下的热门，又出现在 ItemCF 结果中）。去重规则：保留分数最高的那个候选。

#### 综合评分公式

```
finalScore = strategyScore × 0.3
           + preferenceScore × 0.3
           + hotScore × 0.2
           + diversityScore × 0.2
```

| 维度 | 权重 | 计算方式 | 含义 |
|------|------|---------|------|
| 策略基础分 | 30% | 各召回策略赋予的原始分数（上限 100） | 召回来源的可信度 |
| 用户偏好匹配分 | 30% | 商品分类在用户偏好中 → 80 分；用户对该商品有历史行为 → 额外加分 | 个性化程度 |
| 热度分 | 20% | product_hot 归一化到 0–100（相对于候选池最高热度） | 大众认可度 |
| 多样性分 | 20% | 随机扰动 0–30 | 增加推荐多样性，避免每次结果完全相同 |

#### 偏好匹配分计算

```
preferenceScore = 0

if 商品分类 ∈ 用户偏好分类集合:
    preferenceScore = 80

if 用户对该商品有历史行为:
    preferenceScore += min(行为评分 × 5, 20)
```

### 16.6 重排层（Re-rank Layer）

重排层对排序后的结果进行最终调整，提升推荐列表的多样性和用户体验。

#### 分类打散算法

**目标**：避免推荐列表中连续出现同一分类的商品（如连续 5 个手机），提升视觉多样性。

**规则**：同一分类的商品不连续超过 2 个。

```
算法（三轮扫描）：

第 1 轮（严格打散）：
  遍历排序后的候选列表
  if 当前商品分类 == 上一个商品分类 && 连续计数 >= 2:
    跳过（留到下一轮）
  else:
    加入结果列表

第 2 轮（放宽限制）：
  遍历未被选中的商品，不限制连续数量

第 3 轮（兜底）：
  确保所有商品都能入选
```

#### 分页

```
start = (page - 1) × pageSize
end = start + pageSize
hasMore = end < len(reranked)
result = reranked[start:end]
```

### 16.7 ItemCF 离线计算引擎

ItemCF（Item-based Collaborative Filtering）是推荐系统中最经典的协同过滤算法之一。核心思想：如果两个商品被很多相同用户交互过，则它们相似。

#### 算法：余弦相似度

```
sim(A, B) = |users(A) ∩ users(B)| / sqrt(|users(A)| × |users(B)|)
```

其中 `users(A)` 是近 30 天内与商品 A 有过交互的用户集合。

#### 计算流程

```
ComputeItemCF(ctx, svcCtx)
    │
    ├─ 1. ProductModel.FindAll() → 获取所有商品
    │
    ├─ 2. 构建倒排索引
    │     for each product:
    │       UserBehaviorModel.FindUsersByProduct(productId, 30天)
    │       productUsers[productId] = {userId1, userId2, ...}
    │
    ├─ 3. 两两计算余弦相似度
    │     for i in range(products):
    │       for j in range(i+1, products):
    │         intersection = |productUsers[i] ∩ productUsers[j]|
    │         if intersection == 0: continue
    │         sim = intersection / sqrt(|users_i| × |users_j|)
    │         → 双向写入 (i→j, j→i)
    │
    ├─ 4. 批量写入 product_similarity 表
    │     BatchUpsert（每 100 条一批）
    │     INSERT ... ON DUPLICATE KEY UPDATE
    │
    └─ 5. 日志记录处理的商品数量
```

#### 调度方式

- 定时任务（推荐）：每天凌晨 2:00 执行一次
- 管理后台手动触发：可扩展一个管理接口调用 `ComputeItemCF()`
- 首次部署：手动执行一次初始化相似度数据

### 16.8 数据库表结构

#### user_behavior（用户行为日志）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | BIGINT AUTO_INCREMENT | 主键 |
| `user_id` | INT | 用户 ID |
| `product_id` | INT | 商品 ID |
| `category_id` | INT | 商品分类 ID（冗余存储，避免 JOIN） |
| `behavior_type` | TINYINT | 1=浏览 2=点击 3=加购 4=购买 5=收藏 |
| `behavior_time` | BIGINT | 行为发生时间戳（毫秒） |

索引：
- `idx_user_time`：`(user_id, behavior_time DESC)` — 查询用户最近行为
- `idx_product_behavior`：`(product_id, behavior_type)` — ItemCF 倒排索引
- `idx_behavior_type_time`：`(behavior_type, behavior_time DESC)` — 按行为类型查询

#### product_similarity（商品相似度）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | BIGINT AUTO_INCREMENT | 主键 |
| `product_id` | INT | 商品 A |
| `similar_product_id` | INT | 商品 B（与 A 相似） |
| `score` | DOUBLE | 相似度分数 0~1 |
| `updated_at` | BIGINT | 更新时间戳（毫秒） |

索引：
- `uk_product_pair`：`(product_id, similar_product_id)` UNIQUE — 防止重复写入
- `idx_product_score`：`(product_id, score DESC)` — 按相似度降序查询

### 16.9 缓存策略

| Cache Key | TTL | 写入时机 | 失效时机 |
|-----------|-----|----------|----------|
| `jmall:recommend:guess:{userId}:{page}` | 3 min | GuessYouLike 计算完成后 | TTL 自动过期 |

**为什么 TTL 是 3 分钟？** 推荐结果不需要实时更新，3 分钟的缓存在用户体验和系统负载之间取得平衡。用户刷新页面或翻页时，如果缓存未过期则直接返回，避免重复计算。

**缓存 key 包含 page**：不同页码的推荐结果独立缓存，避免用户翻页时重新计算全部结果。

### 16.10 冷启动解决方案

#### 新用户冷启动

新用户没有行为数据，前三路召回（用户偏好、ItemCF、UserCF）均返回空结果。此时热门兜底召回保证推荐列表不为空。

```
新用户请求 GuessYouLike
    │
    ├─ recallByUserPreference → 空（无行为数据）
    ├─ recallByItemCF         → 空（无行为数据）
    ├─ recallByUserCF         → 空（无行为数据）
    └─ recallByHotSelling     → 全站热销 Top 30 ✓
    │
    ▼
排序 → 重排 → 返回热门商品列表
```

随着用户浏览、点击商品，行为数据逐渐积累，推荐结果会越来越个性化。

#### 新商品冷启动

新上架的商品没有用户交互数据，不会出现在 ItemCF 和 UserCF 的结果中。解决方案：

1. 通过 `product_hot` 字段的自然增长（加购/收藏时自增），逐步进入热门兜底池
2. 下一次 ItemCF 离线计算时，如果有用户与新商品交互，新商品会被纳入相似度计算
3. 可扩展：新商品上架后的前 N 天给予额外曝光权重（boost）

### 16.11 前端实现

#### GuessYouLike.vue 组件

```
<GuessYouLike />
├── 标题栏（"猜你喜欢" + "换一批"按钮）
├── 商品网格（5列瀑布流）
│   └── 商品卡片 × N
│       ├── 商品图片（lazy loading）
│       ├── 推荐理由标签（左上角渐变色）
│       ├── 商品名称 + 副标题
│       └── 售价 + 已购人数
├── 加载中提示
├── "已经到底了"提示
└── 空状态提示
```

#### 无限滚动

```javascript
handleScroll() {
  const scrollTop = document.documentElement.scrollTop
  const clientHeight = document.documentElement.clientHeight
  const scrollHeight = document.documentElement.scrollHeight
  // 距离底部 200px 时触发加载下一页
  if (scrollTop + clientHeight >= scrollHeight - 200) {
    this.fetchRecommendations()  // page 自动递增
  }
}
```

#### 行为上报

用户点击推荐列表中的商品时，异步上报点击行为（不阻塞页面跳转）：

```javascript
reportClick(item) {
  this.$axios.post('/api/recommend/reportBehavior', {
    product_id: item.product_id,
    category_id: item.category_id,
    behavior_type: 2,  // 点击
  }).catch(() => {})   // 静默处理，不影响用户体验
}
```

商品详情页打开时，自动上报浏览行为（在 `Details.vue` 中）：

```javascript
getDetails(val) {
  this.$axios.post('/api/product/getDetails', { productID: val })
    .then((res) => {
      this.productDetails = res.data.Product[0]
      this.reportBehavior(this.productDetails, 1)  // 上报浏览
    })
}
```

#### 响应式布局

```css
.guess-grid { grid-template-columns: repeat(5, 1fr); }  /* 默认 5 列 */

@media (max-width: 1200px) { repeat(4, 1fr); }  /* 中屏 4 列 */
@media (max-width: 900px)  { repeat(3, 1fr); }  /* 小屏 3 列 */
@media (max-width: 600px)  { repeat(2, 1fr); }  /* 手机 2 列 */
```

### 16.12 完整调用链

```
用户打开首页（已登录）
    │
    ▼
Home.vue 渲染 <GuessYouLike /> 组件
    │
    ▼
GuessYouLike.mounted() → fetchRecommendations()
    │
    ▼
POST /api/recommend/guessYouLike { page: 1, page_size: 20 }
    │ (vue.config.js proxy: /api/recommend → http://localhost:8889)
    ▼
recommendation-api :8889
    │
    ├─ AuthMiddleware → 解析 JWT → 注入 userId 到 context
    │
    ├─ GuessYouLikeHandler → 解析请求体 → 调用 GuessYouLikeLogic
    │
    └─ GuessYouLikeLogic.GuessYouLike()
        │
        ├─ 1. Cache.Get("jmall:recommend:guess:1:1")
        │     → hit → 直接返回缓存结果
        │     → miss → 继续计算
        │
        ├─ 2. 构建用户画像
        │     ├─ FindUserProductBehaviors(userId, 30天)
        │     │   → [{productId:1, categoryId:1, score:12.0}, ...]
        │     ├─ FindUserPreferredCategories(userId, 30天, 5)
        │     │   → [1, 2, 7]（手机、电视、充电器）
        │     └─ FindRecentProductIds(userId, 50)
        │         → [1, 9, 25, 3, ...]
        │
        ├─ 3. 四路召回
        │     ├─ recallByUserPreference([1,2,7], exclude)
        │     │   → [{productId:4, reason:"猜你喜欢", score:84}, ...]
        │     ├─ recallByItemCF([1,9,25,3,...])
        │     │   → [{productId:10, reason:"相似商品推荐", score:92}, ...]
        │     ├─ recallByUserCF(userId, exclude)
        │     │   → [{productId:18, reason:"和你口味相似的人也在看", score:70}, ...]
        │     └─ recallByHotSelling(exclude)
        │         → [{productId:9, reason:"热门推荐", score:67}, ...]
        │
        ├─ 4. 排序
        │     → 去重（保留高分）
        │     → 批量 FindByIds 获取商品详情
        │     → 计算综合评分
        │     → 降序排列
        │
        ├─ 5. 重排
        │     → 分类打散（同类不连续 > 2 个）
        │     → 分页截取 [0:20]
        │
        ├─ 6. 填充商品详情
        │     → FindByIds → 组装 RecommendItem
        │
        ├─ 7. Cache.Set("jmall:recommend:guess:1:1", resp, 3min)
        │
        └─ 8. 返回 GuessYouLikeResp
              { code, recommendations, has_more }
    │
    ▼
前端 GuessYouLike.vue 渲染瀑布流
    │
    ├─ 用户滚动到底部 → page++ → fetchRecommendations()
    │
    └─ 用户点击商品卡片
        ├─ reportClick(item) → POST /recommend/reportBehavior（异步）
        └─ router-link → /goods/details?productID=xxx
```

### 16.13 性能分析

#### 数据库查询次数（单次请求，缓存 miss）

| 步骤 | 查询 | 次数 |
|------|------|------|
| 用户画像 | FindUserProductBehaviors + FindUserPreferredCategories + FindRecentProductIds | 3 |
| 召回1: 用户偏好 | FindByCategoryIds | 1 |
| 召回2: ItemCF | FindSimilarProductsByIds | 1 |
| 召回3: UserCF | FindSimilarUsers + FindProductsByUsers | 2 |
| 召回4: 热门 | FindTopHot | 1 |
| 排序: 批量查商品 | FindByIds | 1 |
| 填充详情 | FindByIds | 1 |
| **总计** | | **10**（缓存 miss 时） |

#### 优化措施

1. **Redis 缓存 3 分钟**：相同用户 + 页码的重复请求直接返回缓存
2. **批量查询**：FindByIds 消除 N+1 问题
3. **召回数量限制**：每路召回最多 30 个候选，总候选不超过 120 个
4. **离线计算**：ItemCF 相似度离线计算，在线查询只是简单的索引查找
5. **行为数据窗口**：只回溯 30 天，避免扫描过多历史数据

### 16.14 与凑单推荐的对比

| 维度 | 凑单推荐（FillUp） | 猜你喜欢（GuessYouLike） |
|------|-------------------|------------------------|
| 触发场景 | 购物车页面 | 首页底部 |
| 推荐目标 | 帮用户凑满减 | 发现用户可能感兴趣的商品 |
| 核心约束 | 价格接近差额 | 无价格约束 |
| 召回策略 | 差额精准 + 关联商品 + 热销 | 用户偏好 + ItemCF + UserCF + 热销 |
| 个性化程度 | 低（基于购物车内容） | 高（基于用户历史行为） |
| 数据依赖 | 购物车 + 满减规则 | user_behavior + product_similarity |
| 缓存 TTL | 2 分钟 | 3 分钟 |
| 结果数量 | 最多 12 件 | 分页，每页 20 件 |

### 16.15 扩展方向

#### 接入简单机器学习模型

当前的综合评分公式是手工设计的线性加权，可以升级为 LR（逻辑回归）或 LightGBM 模型：

```
特征工程:
  - 用户特征: 注册天数、历史购买次数、偏好分类分布
  - 商品特征: 价格、分类、热度、销量、是否促销
  - 交叉特征: 用户对该分类的历史行为次数、用户历史平均客单价与商品价格的比值

训练数据:
  - 正样本: 用户点击/加购/购买的推荐商品
  - 负样本: 用户曝光但未点击的推荐商品

模型输出:
  - 预测用户点击该商品的概率 → 替代当前的 finalScore
```

#### 引入消息队列

当行为上报 QPS 增长到影响 MySQL 写入性能时：

```
当前: 前端 → POST /reportBehavior → MySQL INSERT
升级: 前端 → POST /reportBehavior → Redis LIST LPUSH
      → 异步消费者 → 批量 INSERT MySQL（每秒一批）
```

Redis LIST 作为轻量级消息队列，无需引入 Kafka 等重型中间件。

#### 实时特征更新

当前用户画像在每次请求时实时计算（查询 user_behavior 表）。如果行为数据量增长，可以改为：

```
行为上报时 → 异步更新 Redis 中的用户画像缓存
推荐请求时 → 直接读取 Redis 中的用户画像（O(1)）
```
