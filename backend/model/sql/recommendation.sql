-- ============================================================
-- JMall 推荐系统数据表
-- ============================================================

-- 1. 用户行为日志表（记录浏览、点击、加购、购买、收藏等行为）
DROP TABLE IF EXISTS `user_behavior`;
CREATE TABLE `user_behavior` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `product_id` int NOT NULL,
  `category_id` int NOT NULL,
  `behavior_type` tinyint NOT NULL COMMENT '1=浏览 2=点击 3=加购 4=购买 5=收藏',
  `behavior_time` bigint NOT NULL COMMENT '行为发生时间戳(ms)',
  PRIMARY KEY (`id`),
  INDEX `idx_user_time` (`user_id`, `behavior_time` DESC),
  INDEX `idx_product_behavior` (`product_id`, `behavior_type`),
  INDEX `idx_behavior_type_time` (`behavior_type`, `behavior_time` DESC)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '用户行为日志';

-- 2. 商品相似度表（离线计算，ItemCF 结果存储）
DROP TABLE IF EXISTS `product_similarity`;
CREATE TABLE `product_similarity` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `product_id` int NOT NULL,
  `similar_product_id` int NOT NULL,
  `score` double NOT NULL DEFAULT 0 COMMENT '相似度分数 0~1',
  `updated_at` bigint NOT NULL COMMENT '更新时间戳(ms)',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uk_product_pair` (`product_id`, `similar_product_id`),
  INDEX `idx_product_score` (`product_id`, `score` DESC)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '商品相似度（ItemCF离线计算）';
