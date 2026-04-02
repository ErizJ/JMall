# JMall

小米风格电商平台，使用 go-zero 重构后端。

## 技术栈

| 层 | 技术 |
|---|---|
| 前端 | Vue 3 · TypeScript · Vite · Element Plus · Pinia |
| 后端 | Go 1.23 · go-zero 1.10（REST） |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7 |

---

## 中间件依赖

| 中间件 | 版本 | 用途 | 默认端口 |
|--------|------|------|----------|
| MySQL | 8.0 | 主数据存储，数据库名 `storedb` | 3306 |
| Redis | 7.x | 缓存（用户信息、商品列表、购物车等） | 6379 |

### MySQL 配置

- 数据库名：`storedb`
- 默认账号：`root` / 密码：`root`（可在 yaml 配置文件中修改）
- 字符集：`utf8mb4`
- 初始化 SQL：`backend/model/sql/storeDB.sql`（含建表 + 种子数据）

### Redis 配置

- 无密码（默认）
- DB 编号：`0`
- 所有服务共用同一个 Redis 实例，key 前缀为 `jmall:`

### 各服务监听端口

| 服务 | 端口 |
|------|------|
| user-api | 8881 |
| product-api | 8882 |
| cart-api | 8883 |
| order-api | 8884 |
| collect-api | 8885 |
| management-api | 8886 |
| 前端（Nginx / Dev） | 8080 |

---

## 快速启动（Docker 一键部署）

> 前提：已安装 [Docker Desktop](https://www.docker.com/products/docker-desktop/) 或 Docker + Docker Compose v2。

```bash
git clone https://github.com/ErizJ/JMall.git
cd JMall
docker compose up --build
```

- MySQL 数据库会在首次启动时自动执行 `storeDB.sql` 完成初始化。
- 所有后端服务会等待 MySQL 和 Redis 健康检查通过后才启动。
- 前端访问地址：[http://localhost:8080](http://localhost:8080)

停止并清理：

```bash
docker compose down          # 保留数据卷
docker compose down -v       # 同时删除 MySQL / Redis 数据卷
```

---

## 本地开发启动

### 前提

- Go 1.23+
- Node.js 20+
- MySQL 8.0（本地或容器）
- Redis 7（本地或容器）

### 1. 初始化数据库

```bash
mysql -u root -p < backend/model/sql/storeDB.sql
```

### 2. 启动所有后端服务

```bash
cd backend
go run service/user/user.go       -f service/user/etc/user-api.yaml &
go run service/product/product.go -f service/product/etc/product-api.yaml &
go run service/cart/cart.go       -f service/cart/etc/cart-api.yaml &
go run service/order/order.go     -f service/order/etc/order-api.yaml &
go run service/collect/collect.go -f service/collect/etc/collect-api.yaml &
go run service/management/management.go -f service/management/etc/management-api.yaml &
```

配置文件路径：`backend/service/<name>/etc/<name>-api.yaml`，可在其中修改数据库连接串和 Redis 地址。

### 3. 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端 dev server 运行在 `http://localhost:8080`，并通过 Vite proxy 将 `/api` 转发到后端。

---

## 项目结构

```
JMall/
├── docker-compose.yml         # 一键启动编排文件
├── frontend/
│   ├── Dockerfile
│   ├── nginx.conf             # 生产环境 Nginx 反代配置
│   └── src/
│       ├── api/               # Axios 请求层
│       ├── stores/            # Pinia 状态管理（user, cart）
│       ├── views/             # 页面组件
│       ├── components/        # 通用 / 业务组件
│       ├── router/            # Vue Router 配置
│       ├── types/             # TypeScript 类型
│       └── utils/             # Axios 实例 / 工具函数
│
└── backend/
    ├── Dockerfile             # 多服务共用，ARG SERVICE 指定目标
    ├── docker-entrypoint.sh   # 容器启动脚本（注入环境变量）
    ├── api/                   # .api 接口定义文件（goctl 输入）
    ├── ctxutil/               # 共享工具（JWT context 提取）
    ├── cache/                 # Redis client 封装
    ├── model/                 # go-zero 生成的数据库模型 + 自定义方法
    │   └── sql/storeDB.sql    # 建表 + 种子数据
    ├── middleware/            # 共享 Auth 中间件
    └── service/
        ├── user/              # 用户服务（注册 / 登录 / 信息）
        ├── product/           # 商品服务（列表 / 搜索 / 详情）
        ├── cart/              # 购物车服务
        ├── order/             # 订单服务
        ├── collect/           # 收藏服务
        └── management/        # 管理后台服务
```

---

## API 概览

所有接口均为 `POST`，Content-Type: `application/json`。
需要登录的接口须在 Header 中携带 `Authorization: Bearer <token>`。

| 服务 | 接口 | 是否需要登录 |
|------|------|-------------|
| user | `/users/register` `/users/login` `/users/logout` `/users/findUserName` | 否 |
| user | `/users/getDetails` `/users/updateUser` `/users/deleteUserById` `/users/isManager` | 是 |
| product | `/products/getAll` `/products/getByCategory` `/products/search` `/products/getHot` `/products/getPromotion` `/products/getDetails` | 否 |
| cart | `/cart/add` `/cart/get` `/cart/update` `/cart/delete` `/cart/isExist` | 是 |
| order | `/orders/add` `/orders/get` `/orders/getDetail` `/orders/delete` | 是 |
| collect | `/collect/add` `/collect/get` `/collect/delete` | 是 |
| management | `/management/getAllOrders` `/management/getOrdersByUserName` `/management/getAllUsers` `/management/deleteUser` `/management/getAllProducts` `/management/getProductsByCategory` `/management/addProduct` `/management/updateProduct` `/management/deleteProduct` `/management/setCategoryHotZero` | 是（管理员） |

---

## 原始项目参考

- 原始前端：[JMall-Vue](https://github.com/ErizJ/JMall-Vue)（Vue 2 + Element UI）
- 原始后端：[JMall-Server](https://github.com/ErizJ/JMall-Server)（Koa.js）
