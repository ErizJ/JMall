-- ============================================================
-- 秒杀模块数据库表结构
-- 适用于 storedb 数据库
-- ============================================================

-- 1. 秒杀活动表
DROP TABLE IF EXISTS `seckill_activity`;
CREATE TABLE `seckill_activity` (
  `id`              bigint       NOT NULL AUTO_INCREMENT,
  `title`           varchar(128) NOT NULL COMMENT '活动标题',
  `product_id`      int          NOT NULL COMMENT '关联商品ID',
  `seckill_price`   bigint       NOT NULL COMMENT '秒杀价（单位：分）',
  `original_price`  bigint       NOT NULL COMMENT '原价（单位：分）',
  `total_stock`     int          NOT NULL COMMENT '秒杀总库存',
  `available_stock` int          NOT NULL COMMENT '剩余库存',
  `limit_per_user`  int          NOT NULL DEFAULT 1 COMMENT '每人限购数量',
  `start_time`      bigint       NOT NULL COMMENT '开始时间（unix秒）',
  `end_time`        bigint       NOT NULL COMMENT '结束时间（unix秒）',
  `status`          tinyint      NOT NULL DEFAULT 0 COMMENT '0=未开始 1=进行中 2=已结束 3=已取消',
  `created_at`      bigint       NOT NULL,
  `updated_at`      bigint       NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_product_id` (`product_id`),
  INDEX `idx_start_time` (`start_time`),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀活动表';

-- 2. 秒杀订单表（与普通订单关联，方便独立管理和防重）
DROP TABLE IF EXISTS `seckill_order`;
CREATE TABLE `seckill_order` (
  `id`            bigint  NOT NULL AUTO_INCREMENT,
  `activity_id`   bigint  NOT NULL COMMENT '秒杀活动ID',
  `order_id`      bigint  NOT NULL COMMENT '关联业务订单ID（orders.order_id）',
  `user_id`       bigint  NOT NULL COMMENT '用户ID',
  `product_id`    int     NOT NULL COMMENT '商品ID',
  `seckill_price` bigint  NOT NULL COMMENT '秒杀成交价（分）',
  `num`           int     NOT NULL DEFAULT 1 COMMENT '购买数量',
  `created_at`    bigint  NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_activity_user` (`activity_id`, `user_id`) COMMENT '同一活动同一用户唯一',
  INDEX `idx_order_id` (`order_id`),
  INDEX `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀订单表';
