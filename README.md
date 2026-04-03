# JMall

小米风格电商平台，前端 Vue 2 + Element UI，后端 go-zero 微服务。

## 技术栈

| 层 | 技术 |
|---|---|
| 前端 | Vue 2.6 · JavaScript · Vue CLI 4 · Element UI 2.x · Vuex 3 |
| 后端 | Go 1.23 · go-zero 1.10（REST） |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7 |
| 消息队列 | Apache Kafka（秒杀异步下单、削峰填谷） |

---

## 各服务监听端口

| 服务 | 端口 | 说明 |
|------|------|------|
| MySQL | 3306 | 主数据存储，数据库名 `storedb` |
| Redis | 6379 | 缓存，key 前缀 `jmall:` |
| Kafka | 9092 | 消息队列（秒杀异步下单） |
| user-api | 8881 | 用户服务（注册/登录/信息） |
| product-api | 8882 | 商品服务（列表/搜索/详情） |
| cart-api | 8883 | 购物车服务 |
| order-api | 8884 | 订单服务 |
| collect-api | 8885 | 收藏服务 |
| management-api | 8886 | 管理后台服务 |
| payment-api | 8887 | 支付服务（支付/回调/退款） |
| aichat-api | 8888 | AI 智能助手服务（豆包大模型 + MCP） |
| seckill-api | 8889 | 秒杀服务（Redis Lua 预扣库存 + Kafka 异步下单） |
| recommendation-api | 8889 (容器内) / 8890 (宿主机) | 推荐服务（猜你喜欢、相似商品） |
| 前端 | 8080 | Nginx（Docker）/ Vue CLI Dev Server（本地） |

---

## 方式一：Docker 一键部署（推荐）

> 前提：已安装 [Docker Desktop](https://www.docker.com/products/docker-desktop/) 或 Docker + Docker Compose v2。

```bash
git clone https://github.com/ErizJ/JMall.git
cd JMall
docker compose up --build
```

启动后会自动完成以下工作：

1. 启动 MySQL 8.0，自动执行 `storeDB.sql`、`payment.sql`、`seckill.sql`、`recommendation.sql` 初始化数据库
2. 启动 Redis 7
3. 启动 Kafka（含 Zookeeper）
4. 等待 MySQL、Redis、Kafka 健康检查通过后，启动 10 个后端微服务
5. 构建前端（`npm run build`），用 Nginx 托管静态文件并反向代理 API 请求

访问地址：[http://localhost:8080](http://localhost:8080)

```bash
# 停止服务（保留数据）
docker compose down

# 停止服务并删除数据卷（完全重置）
docker compose down -v
```

### Docker 架构说明

```
浏览器 → Nginx(:8080)
              │
              ├─ 静态文件（Vue 构建产物）
              │
              └─ /api/* 反向代理 ──┬─ /api/users/*           → user:8881
                                   ├─ /api/product/*          → product:8882
                                   ├─ /api/user/shoppingCart/* → cart:8883
                                   ├─ /api/user/order/*       → order:8884
                                   ├─ /api/order/*            → order:8884
                                   ├─ /api/user/collect/*     → collect:8885
                                   ├─ /api/resources/*        → management:8886
                                   ├─ /api/management/*       → management:8886
                                   ├─ /api/payment/*          → payment:8887
                                   ├─ /api/aichat/*           → aichat:8888
                                   ├─ /api/seckill/*          → seckill:8889
                                   └─ /api/recommend/*        → recommendation:8889
```

前端 Dockerfile 采用两阶段构建：
- 构建阶段：`node:16-alpine` 执行 `npm ci && npm run build`
- 运行阶段：`nginx:1.25-alpine` 托管 `dist/` 静态文件 + 反向代理

后端 Dockerfile 同样两阶段构建：
- 构建阶段：`golang:1.23-alpine` 编译指定服务
- 运行阶段：`alpine:3.20`，通过 `docker-entrypoint.sh` 注入环境变量到 yaml 配置

---

## 方式二：本地开发启动

### 前提

- Go 1.23+
- Node.js 16+（推荐 16.x，兼容 Vue CLI 4）
- MySQL 8.0（本地安装或 Docker 容器）
- Redis 7（本地安装或 Docker 容器）
- Kafka（秒杀服务需要，本地安装或 Docker 容器）

> 如果只想用 Docker 跑中间件，不想本地装 MySQL/Redis/Kafka：
> ```bash
> docker compose up mysql redis zookeeper kafka -d
> ```

### 第一步：初始化数据库

```bash
mysql -u root -p < backend/model/sql/storeDB.sql
mysql -u root -p storedb < backend/model/sql/payment.sql
mysql -u root -p storedb < backend/model/sql/seckill.sql
mysql -u root -p storedb < backend/model/sql/recommendation.sql
```

默认连接信息：`root:root@tcp(localhost:3306)/storedb`，可在各服务的 `etc/<name>-api.yaml` 中修改。

### 第二步：启动后端（10 个服务）

```bash
cd backend

# 一次性启动所有服务（后台运行）
go run service/user/user.go             -f service/user/etc/user-api.yaml &
go run service/product/product.go       -f service/product/etc/product-api.yaml &
go run service/cart/cart.go             -f service/cart/etc/cart-api.yaml &
go run service/order/order.go           -f service/order/etc/order-api.yaml &
go run service/collect/collect.go       -f service/collect/etc/collect-api.yaml &
go run service/management/management.go -f service/management/etc/management-api.yaml &
go run service/payment/payment.go       -f service/payment/etc/payment-api.yaml &
go run service/aichat/aichat.go         -f service/aichat/etc/aichat-api.yaml &
go run service/seckill/seckill.go       -f service/seckill/etc/seckill-api.yaml &
go run service/recommendation/recommendation.go -f service/recommendation/etc/recommendation-api.yaml &
```

> 秒杀服务（seckill）依赖 Kafka，启动前请确保 Kafka 已在 `localhost:9092` 运行。

每个服务的配置文件在 `backend/service/<name>/etc/<name>-api.yaml`，可修改数据库连接串、Redis 地址、JWT 密钥等。

### 第三步：启动前端

```bash
cd frontend
npm install
npm run serve
```

Vue CLI Dev Server 运行在 `http://localhost:8080`，通过 `vue.config.js` 中的 proxy 配置将 `/api` 请求按路径前缀转发到对应的后端服务端口。

> Node.js 17+ 需要设置环境变量：`NODE_OPTIONS=--openssl-legacy-provider npm run serve`

---

## 前端环境切换：真实后端 vs Mock 数据

前端支持两种运行模式，通过不同的 npm 命令一键切换：

| 命令 | 模式 | 需要后端？ | 数据来源 | 适用场景 |
|------|------|-----------|----------|----------|
| `npm run serve` | 真实后端 | 是 | MySQL + Redis | 联调、集成测试、正式开发 |
| `npm run mock` | Mock 数据 | 否 | `src/mock/data.js` | 前端独立开发、UI 调试、演示 |

### 使用真实后端

确保后端 7 个服务和 MySQL、Redis 都已启动（参考上方"方式一"或"方式二"），然后：

```bash
cd frontend
npm run serve
```

请求通过 `vue.config.js` 中的 proxy 转发到各后端服务端口（8881-8887）。

### 使用 Mock 数据（无需后端）

不需要启动任何后端服务或数据库：

```bash
cd frontend
npm run mock
```

原理：`npm run mock` 等价于 `vue-cli-service serve --mode mock`，会加载 `.env.mock` 文件中的 `VUE_APP_USE_MOCK=true`。`main.js` 检测到该变量后，注册一个 Axios 请求拦截器，所有 `/api/*` 请求在发出前就被拦截并返回 `src/mock/data.js` 中的模拟数据，不会产生任何真实网络请求。

Mock 模式下的行为：
- 任意用户名密码均可登录，内置测试用户自动获得管理员权限
- 商品、购物车、订单、收藏、支付等全流程可正常操作
- 商品图片使用 picsum.photos 占位图
- 数据存在内存中，刷新页面后重置
- 浏览器控制台会输出 `[Mock] Mock 模式已启用` 提示

### 自定义 Mock 数据

编辑 `frontend/src/mock/data.js` 可修改模拟数据（商品列表、分类、用户信息等）。编辑 `frontend/src/mock/index.js` 可调整接口响应逻辑。修改后热更新自动生效，无需重启。

---

## 项目结构

```
JMall/
├── docker-compose.yml            # Docker 编排（一键启动全部服务）
│
├── frontend/
│   ├── Dockerfile                # 前端镜像（Node 构建 + Nginx 运行）
│   ├── nginx.conf                # Nginx 配置（静态托管 + API 反向代理）
│   ├── vue.config.js             # Vue CLI 配置（开发代理）
│   ├── public/                   # 静态资源（index.html、favicon）
│   └── src/
│       ├── main.js               # 入口（Element UI、Axios 拦截器、全局组件）
│       ├── Global.js             # 全局变量与工具方法
│       ├── store/                # Vuex 状态管理（user, shoppingCart）
│       ├── router/               # Vue Router 路由配置
│       ├── views/                # 页面组件
│       │   ├── Home.vue          # 首页（轮播图、推荐、促销）
│       │   ├── Goods.vue         # 全部商品（分类、搜索、分页）
│       │   ├── Details.vue       # 商品详情
│       │   ├── ShoppingCart.vue  # 购物车
│       │   ├── ConfirmOrder.vue  # 确认订单
│       │   ├── Payment.vue       # 支付页面
│       │   ├── Order.vue         # 我的订单
│       │   ├── Collect.vue       # 我的收藏
│       │   └── Manager.vue       # 系统管理
│       ├── components/           # 通用组件（MyList, MyLogin, MyRegister 等）
│       ├── mock/                 # Mock 拦截器 + 模拟数据（npm run mock 时启用）
│       └── assets/               # 图片、CSS 资源
│
└── backend/
    ├── Dockerfile                # 后端镜像（Go 编译 + Alpine 运行）
    ├── docker-entrypoint.sh      # 容器启动脚本（注入环境变量）
    ├── api/                      # .api 接口定义文件（goctl 输入）
    ├── model/                    # 数据库模型 + SQL 初始化脚本
    │   └── sql/
    │       ├── storeDB.sql       # 建表 + 种子数据
    │       ├── payment.sql       # 支付相关表
    │       ├── seckill.sql       # 秒杀活动表 + 秒杀订单表
    │       └── recommendation.sql # 用户行为日志 + 商品相似度表
    ├── cache/                    # Redis client 封装
    ├── ctxutil/                  # JWT context 工具
    ├── middleware/               # 共享 Auth 中间件
    ├── kafka/                    # Kafka Producer/Consumer 封装
    └── service/
        ├── user/                 # 用户服务 :8881
        ├── product/              # 商品服务 :8882
        ├── cart/                 # 购物车服务 :8883
        ├── order/                # 订单服务 :8884
        ├── collect/              # 收藏服务 :8885
        ├── management/           # 管理后台服务 :8886
        ├── payment/              # 支付服务 :8887
        ├── aichat/               # AI 智能助手服务 :8888
        ├── seckill/              # 秒杀服务 :8889（Redis Lua + Kafka）
        └── recommendation/       # 推荐服务 :8889
```

---

## API 概览

所有接口均为 `POST`，Content-Type: `application/json`。
需要登录的接口须在 Header 中携带 `Authorization: Bearer <token>`。

| 服务 | 接口 | 是否需要登录 |
|------|------|-------------|
| user | `/users/register` `/users/login` `/users/logout` `/users/findUserName` | 否 |
| user | `/users/getDetails` `/users/updateUser` `/users/deleteUserById` `/users/isManager` | 是 |
| product | `/product/getAllProduct` `/product/getCategory` `/product/getHotProduct` `/product/getProductBySearch` `/product/getDetails` `/product/getDetailsPicture` | 否 |
| cart | `/user/shoppingCart/addShoppingCart` `/user/shoppingCart/getShoppingCart` `/user/shoppingCart/updateShoppingCart` `/user/shoppingCart/deleteShoppingCart` `/user/shoppingCart/isExistShoppingCart` | 是 |
| order | `/user/order/addOrder` `/user/order/getOrder` `/order/getDetails` `/order/deleteOrderById` | 是 |
| collect | `/user/collect/addCollect` `/user/collect/getCollect` `/user/collect/deleteCollect` | 是 |
| management | `/management/getAllOrders` `/management/getOrdersByUserName` `/management/getAllUsers` `/resources/carousel` 等 | 是（管理员） |
| payment | `/payment/create` `/payment/status` `/payment/refund` `/payment/list` | 是 |
| payment | `/payment/notify` `/payment/mock/pay` | 否 |
| seckill | `/seckill/buy` `/seckill/result` | 是 |
| seckill | `/seckill/activity` `/seckill/activities` | 否 |
| recommendation | `/recommend/fillup` `/recommend/guessYouLike` `/recommend/reportBehavior` | 是 |

---

## 技术文档

详细的业务流程、缓存策略、数据库交互细节见 [docs/technical.md](docs/technical.md)。

秒杀系统的完整架构设计（Redis Lua 原子扣库存、Kafka 削峰填谷、三层防超卖等）见 [docs/seckill-architecture.md](docs/seckill-architecture.md)。

---

## 原始项目参考

- 前端基于 Vue 2 + Element UI 开发，参考：[JMall-Vue](https://github.com/ErizJ/JMall-Vue)
- 原始后端：[JMall-Server](https://github.com/ErizJ/JMall-Server)（Koa.js），现已用 go-zero 重构
