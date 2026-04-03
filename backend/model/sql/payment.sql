-- ============================================================
-- 支付模块数据库表结构
-- 适用于 storedb 数据库
-- ============================================================

-- 1. 给 orders 表增加状态字段（当前订单表没有状态，支付需要联动）
ALTER TABLE `orders`
  ADD COLUMN `status` tinyint NOT NULL DEFAULT 0 COMMENT '0=待支付 1=已支付 2=已取消 3=已退款'
  AFTER `order_time`;

-- 2. 支付单表（核心表）
-- 设计理由：
--   - payment_no 是全局唯一的支付流水号，用于对外交互（给支付渠道、给前端）
--   - order_id 关联业务订单，一个订单可能多次发起支付（前一次超时/失败后重新支付）
--   - channel 标识支付渠道，方便路由和对账
--   - channel_trade_no 是第三方返回的交易号，用于退款和对账
--   - status 使用 tinyint 而非 enum，方便扩展
--   - amount 使用 bigint 存储分为单位，避免浮点精度问题（生产级必须）
--   - expire_time 支付单过期时间，超时自动关闭
DROP TABLE IF EXISTS `payment_order`;
CREATE TABLE `payment_order` (
  `id`               bigint       NOT NULL AUTO_INCREMENT,
  `payment_no`       varchar(64)  NOT NULL COMMENT '支付流水号（全局唯一）',
  `order_id`         bigint       NOT NULL COMMENT '关联业务订单ID',
  `user_id`          bigint       NOT NULL COMMENT '用户ID',
  `amount`           bigint       NOT NULL COMMENT '支付金额（单位：分）',
  `channel`          varchar(32)  NOT NULL COMMENT '支付渠道: mock/wechat/alipay',
  `channel_trade_no` varchar(128) NOT NULL DEFAULT '' COMMENT '第三方交易号',
  `status`           tinyint      NOT NULL DEFAULT 0 COMMENT '0=待支付 1=支付中 2=支付成功 3=支付失败 4=已关闭 5=已退款',
  `expire_time`      bigint       NOT NULL DEFAULT 0 COMMENT '支付过期时间（unix秒）',
  `paid_time`        bigint       NOT NULL DEFAULT 0 COMMENT '实际支付时间（unix秒）',
  `notify_url`       varchar(256) NOT NULL DEFAULT '' COMMENT '回调通知URL',
  `extra`            text         COMMENT '扩展字段（JSON），存储渠道特有参数',
  `created_at`       bigint       NOT NULL COMMENT '创建时间（unix秒）',
  `updated_at`       bigint       NOT NULL COMMENT '更新时间（unix秒）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payment_no` (`payment_no`),
  INDEX `idx_order_id` (`order_id`),
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_status` (`status`),
  INDEX `idx_expire_time` (`expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付单表';

-- 3. 退款单表
-- 设计理由：
--   - 退款与支付单 1:N 关系（部分退款场景）
--   - refund_no 全局唯一，对外交互用
--   - 独立 status 管理退款生命周期
DROP TABLE IF EXISTS `payment_refund`;
CREATE TABLE `payment_refund` (
  `id`                bigint       NOT NULL AUTO_INCREMENT,
  `refund_no`         varchar(64)  NOT NULL COMMENT '退款流水号（全局唯一）',
  `payment_no`        varchar(64)  NOT NULL COMMENT '关联支付流水号',
  `order_id`          bigint       NOT NULL COMMENT '关联业务订单ID',
  `user_id`           bigint       NOT NULL COMMENT '用户ID',
  `refund_amount`     bigint       NOT NULL COMMENT '退款金额（单位：分）',
  `reason`            varchar(256) NOT NULL DEFAULT '' COMMENT '退款原因',
  `channel`           varchar(32)  NOT NULL COMMENT '退款渠道（与支付渠道一致）',
  `channel_refund_no` varchar(128) NOT NULL DEFAULT '' COMMENT '第三方退款单号',
  `status`            tinyint      NOT NULL DEFAULT 0 COMMENT '0=退款中 1=退款成功 2=退款失败',
  `created_at`        bigint       NOT NULL COMMENT '创建时间（unix秒）',
  `updated_at`        bigint       NOT NULL COMMENT '更新时间（unix秒）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_refund_no` (`refund_no`),
  INDEX `idx_payment_no` (`payment_no`),
  INDEX `idx_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='退款单表';
