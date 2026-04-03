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
      ├─ /api/payment/*   → payment-api   :8887
      └─ /api/aichat/*    → aichat-api    :8888
                  │                │
            ┌─────┴─────┐         │
            ▼           ▼         ▼
          MySQL 8.0   Redis 7   豆包大模型 API
          (storedb)   (DB 0)    (ark.cn-beijing.volces.com)
```

**8 个独立 go-zero REST 服务**，共用同一个 MySQL 数据库和同一个 Redis 实例。其中 aichat 服务额外对接豆包大模型 API，通过 MCP 工具协议查询数据库后由大模型生成自然语言回复。每个服务有自己的 `ServiceContext`，持有数据库 Model 和 Redis Client。

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
