# JMall

JMall 重构版 —— 小米风格电商平台

## 技术栈

| 层 | 技术 |
|---|---|
| 前端 | Vue 3 + TypeScript + Vite + Element Plus + Pinia |
| 后端 | Go + go-zero（REST API） |
| 数据库 | MySQL 8.0 |

## 项目结构

```
JMall/
├── frontend/          # Vue 3 前端
│   ├── src/
│   │   ├── api/       # API 请求层（axios）
│   │   ├── stores/    # Pinia 状态管理（user, cart）
│   │   ├── views/     # 页面组件
│   │   ├── components/# 通用/业务组件
│   │   ├── router/    # 路由配置
│   │   ├── types/     # TypeScript 类型定义
│   │   └── utils/     # 工具函数（axios 实例等）
│   └── vite.config.ts
│
└── backend/           # go-zero 后端
    ├── api/           # .api 接口定义文件
    ├── service/       # 各业务服务
    │   ├── user/      # 用户服务
    │   ├── product/   # 商品服务
    │   ├── cart/      # 购物车服务
    │   ├── order/     # 订单服务
    │   ├── collect/   # 收藏服务
    │   └── management/# 管理服务
    ├── model/         # go-zero 自动生成的数据库模型
    ├── middleware/     # 共享中间件（Auth）
    ├── config/        # 公共配置结构体
    └── model/sql/     # 数据库 DDL
```

## 快速开始

### 数据库

```bash
mysql -u root -p < backend/model/sql/storeDB.sql
```

### 后端（以 user 服务为例）

```bash
cd backend/service/user
go run user.go -f etc/user-api.yaml
```

### 前端

```bash
cd frontend
npm install
npm run dev
```

## 原始项目参考

- 原始前端: [JMall-Vue](https://github.com/ErizJ/JMall-Vue)（Vue 2 + Element UI）
- 原始后端: [JMall-Server](https://github.com/ErizJ/JMall-Server)（Koa.js）
