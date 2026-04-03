-- ============================================================
-- 性能优化：补充缺失的数据库索引
-- ============================================================

-- orders 表：order_id 和 status 是高频查询字段，缺少索引
ALTER TABLE `orders`
  ADD INDEX `idx_order_id` (`order_id`),
  ADD INDEX `idx_status` (`status`);

-- shoppingcart 表：user_id + product_id 联合查询（FindByUserAndProduct）缺少复合索引
ALTER TABLE `shoppingcart`
  ADD UNIQUE INDEX `uk_user_product` (`user_id`, `product_id`);

-- collect 表：user_id + product_id 联合查询（FindByUserAndProduct）缺少复合索引
ALTER TABLE `collect`
  ADD UNIQUE INDEX `uk_user_product` (`user_id`, `product_id`);

-- combination_product 表：main_product_id 查询缺少索引
ALTER TABLE `combination_product`
  ADD INDEX `idx_main_product_id` (`main_product_id`);

-- product 表：product_selling_price 用于价格区间查询（推荐凑单）
ALTER TABLE `product`
  ADD INDEX `idx_selling_price` (`product_selling_price`),
  ADD INDEX `idx_product_hot` (`product_hot` DESC);
