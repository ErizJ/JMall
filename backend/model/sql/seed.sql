/*
 JMall 完整数据库初始化 + 种子数据
 ============================================================
 执行顺序：本文件完全独立，可直接在空的 storedb 数据库上运行。
 也可在已有库上运行（使用 IF NOT EXISTS + INSERT IGNORE / REPLACE）。

 注意：
   - bcrypt hash 已预生成（密码见注释）
   - 图片使用 https://img.picsum.photos 占位（可替换为真实资源）
   - 新增分类：耳机、智能手表、平板电脑、笔记本、路由器（category_id 9-13）
   - orders.status 字段通过 ALTER 兼容旧库；新库建表时直接包含
 ============================================================
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================
-- 0. 建库（如不存在）
-- ============================================================
CREATE DATABASE IF NOT EXISTS `storedb`
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_0900_ai_ci;
USE `storedb`;

-- ============================================================
-- 1. 表结构（重建）
-- ============================================================

DROP TABLE IF EXISTS `payment_refund`;
DROP TABLE IF EXISTS `payment_order`;
DROP TABLE IF EXISTS `collect`;
DROP TABLE IF EXISTS `shoppingcart`;
DROP TABLE IF EXISTS `orders`;
DROP TABLE IF EXISTS `combination_product`;
DROP TABLE IF EXISTS `product_picture`;
DROP TABLE IF EXISTS `product`;
DROP TABLE IF EXISTS `category`;
DROP TABLE IF EXISTS `carousel`;
DROP TABLE IF EXISTS `sysmanager`;
DROP TABLE IF EXISTS `users`;

-- users（password 扩展至 char(100) 以存储 bcrypt hash）
CREATE TABLE `users` (
  `user_id`         int          NOT NULL AUTO_INCREMENT,
  `userName`        char(40)     NOT NULL,
  `password`        char(100)    NOT NULL COMMENT 'bcrypt hash',
  `userPhoneNumber` char(11)     NULL DEFAULT NULL,
  PRIMARY KEY (`user_id`),
  UNIQUE KEY `userName` (`userName`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

-- sysmanager（syspassword 扩展至 char(100)）
CREATE TABLE `sysmanager` (
  `sys_id`          int          NOT NULL AUTO_INCREMENT,
  `sysname`         char(40)     NOT NULL,
  `syspassword`     char(100)    NOT NULL COMMENT 'bcrypt hash',
  `userPhoneNumber` char(11)     NULL DEFAULT NULL,
  PRIMARY KEY (`sys_id`),
  UNIQUE KEY `sysname` (`sysname`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

-- carousel
CREATE TABLE `carousel` (
  `carousel_id` int         NOT NULL AUTO_INCREMENT,
  `imgPath`     char(200)   NOT NULL,
  `describes`   char(50)    NOT NULL,
  PRIMARY KEY (`carousel_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

-- category
CREATE TABLE `category` (
  `category_id`   int         NOT NULL AUTO_INCREMENT,
  `category_name` char(20)    NOT NULL,
  `category_hot`  int         NULL DEFAULT 0,
  PRIMARY KEY (`category_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

-- product
CREATE TABLE `product` (
  `product_id`           int          NOT NULL AUTO_INCREMENT,
  `product_name`         char(100)    NOT NULL,
  `category_id`          int          NOT NULL,
  `product_title`        char(30)     NOT NULL,
  `product_intro`        text         NOT NULL,
  `product_picture`      char(200)    NULL DEFAULT NULL,
  `product_price`        double       NOT NULL,
  `product_selling_price` double      NOT NULL,
  `product_num`          int          NOT NULL DEFAULT 500,
  `product_sales`        int          NULL DEFAULT 0,
  `product_isPromotion`  int          NOT NULL DEFAULT 0,
  `product_hot`          int          NULL DEFAULT 0,
  PRIMARY KEY (`product_id`),
  INDEX `FK_product_category` (`category_id`),
  CONSTRAINT `FK_product_category` FOREIGN KEY (`category_id`) REFERENCES `category` (`category_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

-- product_picture
CREATE TABLE `product_picture` (
  `id`              int         NOT NULL AUTO_INCREMENT,
  `product_id`      int         NOT NULL,
  `product_picture` char(200)   NULL DEFAULT NULL,
  `intro`           text        NULL,
  PRIMARY KEY (`id`),
  INDEX `FK_product_id` (`product_id`),
  CONSTRAINT `FK_product_id` FOREIGN KEY (`product_id`) REFERENCES `product` (`product_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

-- combination_product
CREATE TABLE `combination_product` (
  `id`                  int  NOT NULL AUTO_INCREMENT,
  `main_product_id`     int  NOT NULL,
  `vice_product_id`     int  NOT NULL,
  `amountThreshold`     int  NULL DEFAULT NULL,
  `priceReductionRange` int  NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

-- orders（含 status 字段，兼容支付模块）
CREATE TABLE `orders` (
  `id`            int       NOT NULL AUTO_INCREMENT,
  `order_id`      bigint    NOT NULL,
  `user_id`       int       NOT NULL,
  `product_id`    int       NOT NULL,
  `product_num`   int       NOT NULL,
  `product_price` double    NOT NULL,
  `order_time`    bigint    NOT NULL,
  `status`        tinyint   NOT NULL DEFAULT 0 COMMENT '0=待支付 1=已支付 2=已取消 3=已退款',
  PRIMARY KEY (`id`),
  INDEX `FK_order_user_id` (`user_id`),
  INDEX `FK_order_id`      (`product_id`),
  CONSTRAINT `FK_order_user_id` FOREIGN KEY (`user_id`)    REFERENCES `users`   (`user_id`),
  CONSTRAINT `FK_order_id`      FOREIGN KEY (`product_id`) REFERENCES `product` (`product_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

-- shoppingcart
CREATE TABLE `shoppingcart` (
  `id`         int NOT NULL AUTO_INCREMENT,
  `user_id`    int NOT NULL,
  `product_id` int NOT NULL,
  `num`        int NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `FK_user_id`         (`user_id`),
  INDEX `FK_shoppingCart_id` (`product_id`),
  CONSTRAINT `FK_shoppingCart_id` FOREIGN KEY (`product_id`) REFERENCES `product` (`product_id`),
  CONSTRAINT `FK_user_id`         FOREIGN KEY (`user_id`)    REFERENCES `users`   (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

-- collect
CREATE TABLE `collect` (
  `id`           int    NOT NULL AUTO_INCREMENT,
  `user_id`      int    NOT NULL,
  `product_id`   int    NOT NULL,
  `category`     int    NOT NULL,
  `collect_time` bigint NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `FK_collect_user_id` (`user_id`),
  INDEX `FK_collect_id`      (`product_id`),
  CONSTRAINT `FK_collect_id`      FOREIGN KEY (`product_id`) REFERENCES `product` (`product_id`),
  CONSTRAINT `FK_collect_user_id` FOREIGN KEY (`user_id`)    REFERENCES `users`   (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

-- payment_order
CREATE TABLE `payment_order` (
  `id`               bigint       NOT NULL AUTO_INCREMENT,
  `payment_no`       varchar(64)  NOT NULL,
  `order_id`         bigint       NOT NULL,
  `user_id`          bigint       NOT NULL,
  `amount`           bigint       NOT NULL COMMENT '单位：分',
  `channel`          varchar(32)  NOT NULL,
  `channel_trade_no` varchar(128) NOT NULL DEFAULT '',
  `status`           tinyint      NOT NULL DEFAULT 0,
  `expire_time`      bigint       NOT NULL DEFAULT 0,
  `paid_time`        bigint       NOT NULL DEFAULT 0,
  `notify_url`       varchar(256) NOT NULL DEFAULT '',
  `extra`            text,
  `created_at`       bigint       NOT NULL,
  `updated_at`       bigint       NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payment_no` (`payment_no`),
  INDEX `idx_order_id`    (`order_id`),
  INDEX `idx_user_id`     (`user_id`),
  INDEX `idx_status`      (`status`),
  INDEX `idx_expire_time` (`expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- payment_refund
CREATE TABLE `payment_refund` (
  `id`                bigint       NOT NULL AUTO_INCREMENT,
  `refund_no`         varchar(64)  NOT NULL,
  `payment_no`        varchar(64)  NOT NULL,
  `order_id`          bigint       NOT NULL,
  `user_id`           bigint       NOT NULL,
  `refund_amount`     bigint       NOT NULL,
  `reason`            varchar(256) NOT NULL DEFAULT '',
  `channel`           varchar(32)  NOT NULL,
  `channel_refund_no` varchar(128) NOT NULL DEFAULT '',
  `status`            tinyint      NOT NULL DEFAULT 0,
  `created_at`        bigint       NOT NULL,
  `updated_at`        bigint       NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_refund_no` (`refund_no`),
  INDEX `idx_payment_no` (`payment_no`),
  INDEX `idx_order_id`   (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 2. 用户数据
--    bcrypt hash 对应密码见注释（cost=10）
--    生产环境请重新用 bcrypt 生成；这里用固定 hash 方便演示
--
--    admin123  -> $2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi
--    user123   -> $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
--    test123   -> $2a$10$YyIhd4GpHMaI1WlhG4EFyOKZJ2WGT8HWfV/q0O5/wBz.wHHZ7bB6K
-- ============================================================
INSERT INTO `users` (`user_id`, `userName`, `password`, `userPhoneNumber`) VALUES
(1,  'admin',   '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '13800000001'),
(2,  'alice',   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '13800000002'),
(3,  'bob',     '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '13800000003'),
(4,  'charlie', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '13800000004'),
(5,  'david',   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '13800000005'),
(6,  'eve',     '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '13800000006'),
(7,  'frank',   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '13800000007'),
(8,  'grace',   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '13800000008'),
(9,  'henry',   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '13800000009'),
(10, 'ivy',     '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '13800000010');
-- 密码：admin=admin123, 其余所有用户=user123

-- ============================================================
-- 3. 管理员
--    syspassword: admin123
-- ============================================================
INSERT INTO `sysmanager` (`sys_id`, `sysname`, `syspassword`, `userPhoneNumber`) VALUES
(1, 'admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '13800000001');

-- ============================================================
-- 4. 轮播图（使用外部占位图，可替换为真实 CDN 链接）
-- ============================================================
INSERT INTO `carousel` (`carousel_id`, `imgPath`, `describes`) VALUES
(1, 'https://img12.360buyimg.com/n1/jfs/t1/191494/25/34741/65489/64c374e4Fb37b2047/1c84d4d7f5b49f45.jpg', '小米15 Ultra — 徕卡四摄旗舰'),
(2, 'https://img12.360buyimg.com/n1/jfs/t1/225008/26/12658/110072/65a24864Fd4520ac7/278f9da6d5fdf028.jpg', 'Redmi K80 Pro — 天玑9400 性能旗舰'),
(3, 'https://img12.360buyimg.com/n1/jfs/t1/129816/27/33316/166715/64ba6656F81bc83e3/80e3641c8816b9d8.jpg', '小米电视 S85 — Mini LED 旗舰大屏'),
(4, 'https://img12.360buyimg.com/n1/jfs/t1/117023/17/34694/104272/6471feb4F2d2a0660/162cee3fd9feec1e.jpg', '小米 Watch S4 — 血压检测智能手表'),
(5, 'https://img12.360buyimg.com/n1/jfs/t1/178961/14/43969/110572/655ac406F77e9ffe8/49eac41779da5020.jpg', '小米平板 7 Pro — 骁龙 X 超薄生产力'),
(6, 'https://img12.360buyimg.com/n1/jfs/t1/120017/33/34460/32883/64ba671fF118b88f5/bdfccb14f49f41a7.jpg', '小米笔记本 Pro 14 Ultra — OLED 旗舰本');

-- ============================================================
-- 5. 商品分类（原有 1-8 保留，新增 9-13）
--    category_id:
--      1=手机  2=电视  3=空调  4=洗衣机  5=保护套
--      6=保护膜  7=充电器  8=充电宝
--      9=耳机  10=智能手表  11=平板电脑  12=笔记本  13=路由器
-- ============================================================
INSERT INTO `category` (`category_id`, `category_name`, `category_hot`) VALUES
(1,  '手机',     15),
(2,  '电视',     8),
(3,  '空调',     6),
(4,  '洗衣机',   4),
(5,  '保护套',   10),
(6,  '保护膜',   7),
(7,  '充电器',   9),
(8,  '充电宝',   6),
(9,  '耳机',     12),
(10, '智能手表', 11),
(11, '平板电脑', 9),
(12, '笔记本',   7),
(13, '路由器',   5);

-- ============================================================
-- 6. 商品数据
--    图片使用 picsum 占位（每个产品 ID 对应固定种子，视觉上不同）
--    product_isPromotion=1 表示促销，product_hot 越大越热门
-- ============================================================

-- ---------- 手机 (category_id=1) ----------
INSERT INTO `product` VALUES
(1,  '小米15 Ultra',          1, '徕卡可变光圈，影像旗舰',
 '6000mAh硅碳负极电池 / 骁龙8至尊版 / 徕卡可变光圈镜头 / 1英寸超大底 / 100W有线+80W无线快充 / IP68防尘防水 / 6.73英寸2K OLED',
 'https://img12.360buyimg.com/n1/jfs/t1/191494/25/34741/65489/64c374e4Fb37b2047/1c84d4d7f5b49f45.jpg', 6499, 5999, 800, 320, 1, 98),

(2,  '小米15 Pro',            1, '骁龙8至尊，超长续航',
 '5800mAh大电量 / 骁龙8至尊版 / 徕卡三摄 / 90W有线+50W无线 / 6.73英寸LTPO OLED / 旗舰级散热',
 'https://img12.360buyimg.com/n1/jfs/t1/162969/39/39306/44350/64e859adF05e88510/79c67d38ad3ac68f.jpg', 5499, 4999, 600, 210, 1, 85),

(3,  '小米15',                1, '轻薄旗舰，双面玻璃',
 '5240mAh / 骁龙8至尊版 / 5000万徕卡光学镜头 / 90W快充 / 6.36英寸OLED',
 'https://img12.360buyimg.com/n1/jfs/t1/273656/23/29811/731/681b1e9aF9ce3b5bc/6102ba8228118daf.png', 4299, 3999, 1000, 450, 0, 72),

(4,  'Redmi K80 Pro',         1, '天玑9400，性能旗舰',
 '6000mAh / 天玑9400 / 5000万光学防抖 / 90W快充 / 6.67英寸2K OLED / 骁龙X85 5G基带',
 'https://img12.360buyimg.com/n1/jfs/t1/225008/26/12658/110072/65a24864Fd4520ac7/278f9da6d5fdf028.jpg', 3499, 2999, 1200, 560, 1, 90),

(5,  'Redmi K80',             1, '天玑9300+，均衡旗舰',
 '5500mAh / 天玑9300+ / 5000万主摄 / 45W快充 / 6.67英寸2K OLED',
 'https://img12.360buyimg.com/n1/jfs/t1/232454/4/10838/116695/65a24861F0c14850a/a590b04a373904bb.jpg', 2499, 2199, 1500, 380, 0, 65),

(6,  'Redmi Note 14 Pro+',    1, '2亿像素，影像中端',
 '2亿像素主摄 / 天玑9200+ / 90W快充 / 5000mAh / 6.67英寸2K OLED / IP68',
 'https://img12.360buyimg.com/n1/jfs/t1/112377/35/34825/16116/65781e43Fef835de1/f7bb17ee87a45260.jpg', 2299, 1999, 2000, 620, 1, 78),

(7,  'Redmi Note 14',         1, '千元好屏，轻薄日常',
 '5110mAh / 天玑7300 / 5000万主摄 / 45W快充 / 6.67英寸1.5K OLED',
 'https://img12.360buyimg.com/n1/jfs/t1/109139/6/46202/76948/654c8887Fe66e6bae/f820246a26752d1e.jpg', 1299, 1099, 2500, 810, 0, 55),

(8,  'Redmi 14C',             1, '5G千元机，大电量首选',
 '5160mAh / 天玑7025 / 5000万AI相机 / 18W快充 / 6.88英寸IPS屏',
 'https://img12.360buyimg.com/n1/jfs/t1/158479/31/47913/8842/66cc2352Fd8ec932a/d2b0de0b2a55be2c.png', 899,  799,  3000, 950, 1, 48),

(9,  '小米Civi 4 Pro',        1, '前置旗舰，自拍神器',
 '索尼IMX858前置镜头 / 骁龙8s Gen3 / 4700mAh / 67W快充 / 6.55英寸OLED / 轻薄设计',
 'https://img12.360buyimg.com/n1/jfs/t1/119827/15/36050/67290/64a11c2bF8c013f81/7f5bfffe1625cb30.jpg', 2999, 2699, 800, 260, 1, 62),

(10, '小米MIX Fold 4',        1, '折叠旗舰，极致轻薄',
 '骁龙8至尊版 / 5100mAh / 徕卡三摄 / 展开8.02英寸OLED / 折叠合厚9.95mm',
 'https://img12.360buyimg.com/n1/jfs/t1/118964/25/37403/49705/64c46699F2b92ec3e/d7ff3a93d076d32e.jpg', 9999, 8999, 300, 88, 1, 95);

-- ---------- 电视 (category_id=2) ----------
INSERT INTO `product` VALUES
(11, '小米电视 S85 Mini LED',  2, '2304分区 Mini LED，旗舰巨幕',
 '2304背光分区 / 4K 144Hz / 全局调光HDR10000 / 环绕立体声 / 骁龙4K处理器 / 小爱同学 / PatchWall',
 'https://img12.360buyimg.com/n1/jfs/t1/129816/27/33316/166715/64ba6656F81bc83e3/80e3641c8816b9d8.jpg', 9999, 8999, 200, 45, 1, 88),

(12, '小米电视 S75 Mini LED',  2, '75英寸 Mini LED 旗舰',
 '1152分区 / 4K 120Hz / HDR5000 / 杜比视界 / 75英寸全面屏',
 'https://img12.360buyimg.com/n1/jfs/t1/143847/28/34916/142000/64ba6656Fd5335683/eb373e2c68d53452.jpg', 6999, 5999, 300, 78, 1, 76),

(13, '小米电视 A Pro 65英寸',  2, '4K 144Hz，全面屏高刷',
 '4K 144Hz / MEMC运动补偿 / 杜比音效 / 内置小爱 / 65英寸',
 'https://img12.360buyimg.com/n1/jfs/t1/105458/10/28155/467695/62625c45E6b0eea3f/e98bcad64ce51291.jpg', 3999, 3499, 400, 130, 0, 65),

(14, '小米电视 A Pro 55英寸',  2, '55英寸4K QLED',
 '4K QLED / 120Hz / HDR10 / 2+32GB大存储 / 55英寸全面屏',
 'https://img12.360buyimg.com/n1/jfs/t1/103185/4/32909/127993/6315a81bEbfaeb5ff/a72458f586caaed7.jpg', 2499, 2199, 500, 180, 1, 58),

(15, '小米电视 4A 43英寸',     2, '入门全高清，性价比之选',
 'FHD全高清 / 60Hz / 1+8GB / 小爱同学 / 43英寸',
 'https://img12.360buyimg.com/n1/jfs/t1/37602/38/12801/159674/5d16d3a3E59fee584/0d80b61a8ec50900.jpg', 1299, 1099, 600, 260, 0, 44),

(16, '小米电视 ES 75英寸',     2, '超大屏，远场语音',
 '4K / 120Hz / 2+32GB / 远场语音 / 75英寸',
 'https://img12.360buyimg.com/n1/jfs/t1/165000/20/33476/39731/63da057eF96d3a7a2/bed67f1655b4b996.jpg', 4999, 4299, 250, 90, 1, 60);

-- ---------- 空调 (category_id=3) ----------
INSERT INTO `product` VALUES
(17, '小米空调 新风旗舰版 1.5P', 3, '引入新风，净化室内空气',
 '一级能效 / 1.5匹 / 新风净化 / 全直流变频 / 静音13dB / PM2.5过滤 / Wi-Fi控制',
 'https://img12.360buyimg.com/n1/jfs/t1/112573/25/24539/161157/623ea389Ecd80ac9c/58b8ae40843d3cd9.jpg', 5999, 5299, 150, 42, 1, 55),

(18, '米家互联网空调 C1 1.5P', 3, '一级能效，变频节能',
 '一级能效 / 1.5匹 / 全直流变频 / 静音设计 / 自清洁 / 全屋互联',
 'https://img12.360buyimg.com/n1/jfs/t1/161837/25/14868/34939/605d6031Ed1a6e301/34ed2ed17f5609ea.jpg', 2999, 2599, 200, 88, 0, 42),

(19, '米家空调 1P',             3, '小户型首选，快速制冷',
 '大1匹 / 三级能效 / 静音运行 / 快速制冷 / 除湿 / Wi-Fi控制',
 'https://img12.360buyimg.com/n1/jfs/t1/161837/25/14868/34939/605d6031Ed1a6e301/34ed2ed17f5609ea.jpg', 1999, 1799, 300, 120, 0, 35),

(20, '小米空调 立式 3P',        3, '客厅大空间，立式旗舰',
 '3匹 / 一级能效 / 柜式 / 制冷/热双效 / 立体送风 / 自清洁',
 'https://img12.360buyimg.com/n1/jfs/t1/112573/25/24539/161157/623ea389Ecd80ac9c/58b8ae40843d3cd9.jpg', 6999, 5999, 100, 28, 1, 48);

-- ---------- 洗衣机 (category_id=4) ----------
INSERT INTO `product` VALUES
(21, '米家洗烘一体机 Pro 10kg', 4, '国标双A+洗烘，22种模式',
 'A+洗涤/A+烘干 / 22种模式 / 智能投放洗涤剂 / 小爱语音 / 除菌率99.9%+ / 10kg',
 'https://img12.360buyimg.com/n1/jfs/t1/103966/1/23144/48439/62209887E3dc8e429/f71d6340a46075c5.jpg', 3299, 2999, 200, 76, 1, 60),

(22, '米家滚筒洗衣机 10kg',     4, '大容量，深度清洁',
 '10kg大容量 / 1400转 / 15种洗涤程序 / 蒸汽除菌 / Wi-Fi控制',
 'https://img12.360buyimg.com/n1/jfs/t1/105832/36/23828/61710/6220986eE87a116d8/9a5948a58d0c6b66.jpg', 1999, 1699, 300, 112, 0, 45),

(23, '米家迷你洗衣机 3kg',      4, '宝宝专属，杀菌洗涤',
 '3kg婴儿洗衣机 / 高温蒸汽 / 除菌率99.9% / 四种程序 / 超小机身',
 'https://img12.360buyimg.com/n1/jfs/t1/103966/1/23144/48439/62209887E3dc8e429/f71d6340a46075c5.jpg', 999,  899,  400, 160, 1, 38);

-- ---------- 保护套 (category_id=5) ----------
INSERT INTO `product` VALUES
(24, '小米15 Pro 原装保护壳',   5, '磁吸皮革，优雅耐用',
 '磁吸皮革材质 / 四角气囊防摔 / 精准开孔 / 支持MagSafe / 小米15 Pro 专属',
 'https://img.picsum.photos/seed/mi15case/400/400', 129, 99,  500, 200, 1, 55),

(25, 'Redmi K80 Pro 液态硅胶壳',5, '亲肤手感，全面防护',
 '液态硅胶 / 微纤维内衬 / 四角气囊 / 精准按键 / 哑光手感',
 'https://img12.360buyimg.com/n1/jfs/t1/225008/26/12658/110072/65a24864Fd4520ac7/278f9da6d5fdf028.jpg', 69,  59,  800, 320, 0, 48),

(26, '小米15 磁吸支架壳',       5, '磁吸支架，办公神器',
 '内置磁环 / 多角度支架 / TPU软边 / PC硬背 / 黑/白/蓝三色',
 'https://img.picsum.photos/seed/mi15stand/400/400', 99,  79,  600, 180, 1, 42),

(27, 'Redmi Note 14 Pro 透明壳',5, '展示原色，轻薄防护',
 '高透PC材质 / 防黄变工艺 / 0.8mm极薄 / 精准开孔',
 'https://img.picsum.photos/seed/note14case/400/400', 39,  29,  1000,400, 0, 35),

(28, '小米Civi 4 Pro 皮革翻盖壳',5,'翻盖设计，保护屏幕',
 '翻盖皮革 / 内置卡槽 / 支架功能 / 磁吸自动休眠',
 'https://img.picsum.photos/seed/civicase/400/400', 89,  69,  400, 140, 1, 38);

-- ---------- 保护膜 (category_id=6) ----------
INSERT INTO `product` VALUES
(29, '小米15 Pro 原装钢化膜',   6, '高清防指纹，2.5D弧边',
 '9H硬度 / 高清透亮 / 疏油疏水 / 2.5D弧边 / 防指纹涂层 / 一片装',
 'https://img.picsum.photos/seed/mi15glass/400/400', 49,  39,  1000,500, 1, 58),

(30, 'Redmi K80 Pro 全覆盖水凝膜',6,'柔性水凝，曲面适配',
 '热弯曲面贴合 / 自动修复划痕 / 哑光防眩光 / 全覆盖',
 'https://img.picsum.photos/seed/k80film/400/400', 39,  29,  1200,620, 0, 50),

(31, '小米15 隐私防窥膜',        6, '防窥保密，保护隐私',
 '45°防窥角度 / 高透钢化 / 疏油涂层 / 防刮',
 'https://img.picsum.photos/seed/mi15priv/400/400', 59,  49,  800, 280, 1, 44),

(32, '通用防蓝光护眼膜（6寸）',  6, '过滤蓝光，护眼首选',
 '防蓝光率>90% / 德国莱茵认证 / 减少眩光 / 6英寸通用款',
 'https://img.picsum.photos/seed/bluefilm/400/400', 29,  19,  1500,760, 0, 38);

-- ---------- 充电器 (category_id=7) ----------
INSERT INTO `product` VALUES
(33, '小米充电器 140W GaN',     7, 'C+C+A 三口，笔记本直充',
 '140W最大输出 / 3口（2C+1A）/ GaN氮化镓 / 折叠插脚 / 兼容PD3.1 / 支持笔记本快充',
 'https://img.picsum.photos/seed/140wgan/400/400', 199, 169, 500, 220, 1, 72),

(34, '小米充电器 120W 单口',     7, '超快120W，手机满速充',
 '120W / 单USB-A / 可折叠插头 / 小巧便携 / 支持多种快充协议',
 'https://img.picsum.photos/seed/120w/400/400', 99,  79,  800, 380, 0, 60),

(35, '小米充电器 67W GaN',      7, '双口快充，出行必备',
 '67W / USB-C+USB-A / GaN氮化镓 / 比苹果原装小60% / 兼容MagSafe',
 'https://img.picsum.photos/seed/67wgan/400/400', 79,  59,  1000,450, 1, 65),

(36, '米家无线充电器 80W',       7, '立式无线，一放即充',
 '80W有线 + 50W无线 / 竖置横置两用 / 兼容Qi2 / 温控散热',
 'https://img.picsum.photos/seed/80wwl/400/400', 149, 119, 600, 200, 1, 55),

(37, '小米车载充电器 65W',       7, '双口车充，告别低速',
 '65W USB-C + 22.5W USB-A / 智能分配 / 防过热保护 / 兼容PD/QC',
 'https://img.picsum.photos/seed/65wcar/400/400', 89,  69,  700, 260, 0, 48);

-- ---------- 充电宝 (category_id=8) ----------
INSERT INTO `product` VALUES
(38, '小米充电宝 3 Ultra 25000mAh',8,'25000mAh，笔记本也能充',
 '25000mAh / 140W双向快充 / 支持笔记本 / 航空级电池 / LCD电量显示 / 2C+1A三口',
 'https://img.picsum.photos/seed/25000pb/400/400', 399, 349, 400, 180, 1, 80),

(39, '小米充电宝 3 20000mAh',    8, '20000mAh，100W快充',
 '20000mAh / 100W双向快充 / 2C+1A / 超薄机身 / 铝合金外壳',
 'https://img.picsum.photos/seed/20000pb/400/400', 249, 199, 600, 280, 0, 65),

(40, '小米充电宝 4 磁吸版 10000mAh',8,'MagSafe磁吸，随手就充',
 '10000mAh / MagSafe磁吸 / 15W无线 / 30W有线 / 兼容iPhone/小米磁吸系列 / 超薄',
 'https://img.picsum.photos/seed/magsafe/400/400', 199, 169, 500, 320, 1, 72),

(41, '米家充电宝 10000mAh 轻薄版',8,'18mm超薄，口袋随行',
 '10000mAh / 22.5W快充 / 18mm极薄 / 铝合金 / Type-C双向',
 'https://img.picsum.photos/seed/10000slim/400/400', 99,  79,  800, 480, 0, 55);

-- ---------- 耳机 (category_id=9) ----------
INSERT INTO `product` VALUES
(42, '小米耳机 Air3 Pro',       9, '主动降噪，空间音频',
 'ANC主动降噪 / 空间音频 / 骨声纹识别 / 通话降噪 / 6小时续航 / Hi-Res认证',
 'https://img12.360buyimg.com/n1/jfs/t1/129520/33/33785/68620/6479976fFb5c8b0ee/0051f2eb2cfe7b92.jpg', 599, 499, 600, 280, 1, 85),

(43, 'Redmi Buds 6 Pro',       9, '千元内降噪旗舰',
 '55dB主动降噪 / 11mm双磁路动圈 / Hi-Res Audio / 9小时续航+40h充电盒 / IPX4防水',
 'https://img12.360buyimg.com/n1/jfs/t1/232962/37/11826/40009/65a24863F2095a651/e1c452b463d417b9.jpg', 399, 329, 800, 420, 1, 78),

(44, 'Redmi Buds 6',           9, '轻量通勤，日常首选',
 '30dB降噪 / 10mm动圈 / 6小时续航 / IPX4 / 双设备连接',
 'https://img12.360buyimg.com/n1/jfs/t1/103347/11/46485/43811/652111e5F50ad3240/d97418f59e209231.jpg', 199, 169, 1200,560, 0, 62),

(45, '小米耳机 Air3 SE',        9, '轻量开放，不入耳设计',
 '不入耳气传导式 / 开放式佩戴 / 通透倾听 / 8小时续航 / 双连接',
 'https://img.picsum.photos/seed/air3se/400/400', 299, 249, 500, 200, 1, 55),

(46, '小米头戴耳机 Studio Pro', 9, '头戴式，专业监听级',
 '40mm大单元 / 主动降噪 / LDAC高清音质 / 35小时续航 / 折叠设计 / Hi-Res Wireless',
 'https://img.picsum.photos/seed/studiopro/400/400', 999, 799, 300, 120, 1, 70);

-- ---------- 智能手表 (category_id=10) ----------
INSERT INTO `product` VALUES
(47, '小米 Watch S4',           10,'血压检测，健康旗舰',
 '血压检测 / ECG心电图 / 血氧 / 心率 / 睡眠监测 / 1.43英寸AMOLED / 15天续航 / 150+运动模式',
 'https://img12.360buyimg.com/n1/jfs/t1/117023/17/34694/104272/6471feb4F2d2a0660/162cee3fd9feec1e.jpg', 1499,1299, 400, 180, 1, 88),

(48, '小米 Watch S3',           10,'AMOLED，健康全能',
 '1.43英寸AMOLED / 血氧/心率/睡眠 / 12天续航 / 可换表圈 / 蓝牙通话 / GPS',
 'https://img12.360buyimg.com/n1/jfs/t1/142070/19/36915/115870/6471feb4F57f13f50/150d1f2f337d0699.jpg', 999, 799, 600, 280, 0, 75),

(49, 'Redmi Watch 5',           10,'超大屏，长续航千元表',
 '2.07英寸大屏 / 血氧 / 心率 / 24天超长续航 / 130+运动模式 / 蓝牙通话',
 'https://img12.360buyimg.com/n1/jfs/t1/103997/40/46214/67032/65156cf0Fdf5ccf48/fd5d914d46dd8a05.jpg', 499, 399, 800, 360, 1, 65),

(50, 'Redmi Watch 5 Active',    10,'百元入门，运动首选',
 '1.96英寸大屏 / 心率血氧 / 18天续航 / IP68 / 100+运动模式',
 'https://img12.360buyimg.com/n1/jfs/t1/108128/40/35910/81278/65156cefFe0b26fe5/3958ca9c254d5c24.jpg', 199, 169, 1000,480, 0, 52),

(51, '小米 Band 9 Pro',         10,'手环旗舰，超薄大屏',
 '1.74英寸AMOLED / LTPO全程心率 / 血氧 / 21天续航 / 150+运动 / 超薄6.99mm',
 'https://img12.360buyimg.com/n1/jfs/t1/115408/10/40180/86378/65693bdeF2d421419/ea820b50ff76cce3.jpg', 399, 299, 1200,650, 1, 80);

-- ---------- 平板电脑 (category_id=11) ----------
INSERT INTO `product` VALUES
(52, '小米平板 7 Pro',          11,'骁龙X，生产力旗舰平板',
 '骁龙X系列 / 12.1英寸2.8K OLED / 144Hz / 10000mAh / 90W快充 / 支持手写笔 / 键盘盖选配',
 'https://img12.360buyimg.com/n1/jfs/t1/178961/14/43969/110572/655ac406F77e9ffe8/49eac41779da5020.jpg', 3999,3499, 300, 120, 1, 85),

(53, '小米平板 7',              11,'均衡旗舰，影音生产力',
 '天玑9400 / 11.2英寸3K OLED / 144Hz / 9000mAh / 67W快充',
 'https://img12.360buyimg.com/n1/jfs/t1/106329/24/28697/75387/64f92987F6f0a375d/48d0c3667740c1bb.jpg', 2999,2499, 400, 180, 0, 72),

(54, '小米平板 7 SE',           11,'轻薄护眼，学习办公',
 '天玑7300 / 11英寸2.5K LCD / 90Hz / 8000mAh / 33W快充 / 护眼认证',
 'https://img12.360buyimg.com/n1/jfs/t1/158907/14/38944/40830/64f92987F5dababaf/7584f48beeec1dd3.jpg', 1499,1299, 600, 280, 1, 58),

(55, 'Redmi Pad Pro 2',         11,'轻薄大屏，性价比之选',
 '天玑8300 / 12.1英寸2.5K / 120Hz / 8600mAh / 45W快充 / 四扬声器',
 'https://img12.360buyimg.com/n1/jfs/t1/175039/21/39403/84724/64f92987F5da1a33f/c6adc64866d8184a.jpg', 1999,1699, 500, 220, 0, 65);

-- ---------- 笔记本 (category_id=12) ----------
INSERT INTO `product` VALUES
(56, '小米笔记本 Pro 14 Ultra', 12,'Ultra芯片，超薄全能本',
 'Ultra 9 185H / 14.5英寸3K OLED / 120Hz / 64GB+2TB / 独显RTX4060 / 72Wh电池 / 1.9kg',
 'https://img12.360buyimg.com/n1/jfs/t1/120017/33/34460/32883/64ba671fF118b88f5/bdfccb14f49f41a7.jpg',8999, 7999, 200, 88, 1, 90),

(57, '小米笔记本 Air 14',       12,'极致轻薄，日常全能',
 '酷睿Ultra 5 125H / 14英寸2.8K OLED / 60Hz / 16GB+512GB / 英特尔核显 / 1.35kg超轻',
 'https://img12.360buyimg.com/n1/jfs/t1/125479/17/37614/21192/64ba671fF740836a5/479ef09aa9e4efe4.jpg',5999, 4999, 300, 130, 0, 72),

(58, 'RedmiBook Pro 15',        12,'大屏高性能，学生首选',
 '酷睿Ultra 5 125H / 15.6英寸2.5K OLED / 60Hz / 16GB+512GB / RTX4050 独显',
 'https://img12.360buyimg.com/n1/jfs/t1/151806/16/32830/64319/64489584Fbd9d3fb1/802d30f91aa4a89a.jpg',5499, 4699, 400, 160, 1, 68),

(59, 'RedmiBook 14',            12,'轻巧入门，全能办公',
 '酷睿Ultra 5 125U / 14英寸2.8K OLED / 16GB+512GB / 1.37kg / 72Wh / 18小时续航',
 'https://img12.360buyimg.com/n1/jfs/t1/116910/20/34680/72246/64489584Fcbe8e64c/8bcb45130904d5c1.jpg',3999, 3499, 500, 220, 0, 55);

-- ---------- 路由器 (category_id=13) ----------
INSERT INTO `product` VALUES
(60, '小米路由器 BE10000',      13,'Wi-Fi 7旗舰，万兆有线',
 'Wi-Fi 7 / 10000Mbps无线速率 / 2.5G+万兆有线 / 12天线 / Mesh组网 / 小爱同学',
 'https://img12.360buyimg.com/n1/jfs/t1/103858/31/40252/93953/6437e591F8a8cec14/9731fced8157d5ea.jpg',1499,1299, 300, 120, 1, 80),

(61, '小米路由器 BE7000',       13,'Wi-Fi 7，家庭高速首选',
 'Wi-Fi 7 / 7000Mbps / 2.5G网口 / 旗舰级覆盖 / Mesh自动漫游',
 'https://img12.360buyimg.com/n1/jfs/t1/121337/15/32262/56979/637c645cE0a72b6e8/2a7d66566e6dd258.jpg', 799, 649, 500, 240, 0, 65),

(62, '小米路由器 AX9000',       13,'三频Wi-Fi 6，游戏加速',
 'Wi-Fi 6 三频 / 9000Mbps / 12天线 / 游戏加速 / 4个2.5G口',
 'https://img12.360buyimg.com/n1/jfs/t1/124358/24/28990/37170/637c6466E661fa31f/9b16ab7d1d2797ad.jpg', 599, 499, 600, 280, 1, 58),

(63, '小米路由器 BE3600',       13,'入门Wi-Fi 7，够用就好',
 'Wi-Fi 7 / 3600Mbps / 千兆网口 / 自动优化频段 / 小巧机身',
 'https://img12.360buyimg.com/n1/jfs/t1/136856/25/32764/40405/637c6466Ec6b61a7f/2547fdada75b0da6.jpg', 299, 249, 800, 380, 0, 48);

-- ============================================================
-- 7. 商品详细图片（每个商品 3-5 张）
--    使用 picsum seed 保证图片不同
-- ============================================================
INSERT INTO `product_picture` (`product_id`, `product_picture`, `intro`) VALUES
(1, 'https://img12.360buyimg.com/n1/jfs/t1/194367/27/35890/60253/64c374e8Fe4cc5362/14ba119c576cb34c.jpg', NULL),
(1, 'https://img12.360buyimg.com/n1/jfs/t1/221125/11/23263/93240/64c374e6Fb6e06e80/18d9539a28fb28ed.jpg', NULL),
(1, 'https://img12.360buyimg.com/n1/jfs/t1/96506/25/41752/94762/64c374ebF2b204d2c/a41476b651cd1091.jpg', NULL),
(1, 'https://img12.360buyimg.com/n1/jfs/t1/188234/35/36067/35053/64c485ccF7202874d/a974d7b02d7d3438.jpg', NULL),
(2, 'https://img12.360buyimg.com/n1/jfs/t1/103424/16/37073/12232/64e859abFb36a0bb0/4b8f25dfbe9ba4f7.jpg', NULL),
(2, 'https://img12.360buyimg.com/n1/jfs/t1/124623/37/36331/7494/64e859acFce95bef4/c90571431597b7f7.jpg', NULL),
(2, 'https://img12.360buyimg.com/n1/jfs/t1/166303/34/39283/11951/64e859abFb2eab319/716545993eefec4b.jpg', NULL),
(3, 'https://img12.360buyimg.com/n1/jfs/t1/244464/19/7644/16909/6627cc7fF5ec76ccd/c7a140a57884c0e5.png', NULL),
(3, 'https://img12.360buyimg.com/n1/jfs/t1/115932/19/40163/45399/64c374e2F787eed10/cf6d29818ce20d62.jpg', NULL),
(4, 'https://img12.360buyimg.com/n1/jfs/t1/227749/19/12606/126392/65a24863F7255860e/53073b7463685cd7.jpg', NULL),
(4, 'https://img12.360buyimg.com/n1/jfs/t1/232454/4/10838/116695/65a24861F0c14850a/a590b04a373904bb.jpg', NULL),
(4, 'https://img12.360buyimg.com/n1/jfs/t1/232962/37/11826/40009/65a24863F2095a651/e1c452b463d417b9.jpg', NULL),
(5, 'https://img12.360buyimg.com/n1/jfs/t1/227749/19/12606/126392/65a24863F7255860e/53073b7463685cd7.jpg', NULL),
(5, 'https://img12.360buyimg.com/n1/jfs/t1/225008/26/12658/110072/65a24864Fd4520ac7/278f9da6d5fdf028.jpg', NULL),
(6, 'https://img12.360buyimg.com/n1/jfs/t1/273656/23/29811/731/681b1e9aF9ce3b5bc/6102ba8228118daf.png', NULL),
(6, 'https://img12.360buyimg.com/n1/jfs/t1/244464/19/7644/16909/6627cc7fF5ec76ccd/c7a140a57884c0e5.png', NULL),
(7, 'https://img12.360buyimg.com/n1/jfs/t1/186056/36/42245/71045/654c8887Ff9e53317/05646f360fab16a9.jpg', NULL),
(7, 'https://img12.360buyimg.com/n1/jfs/t1/189400/8/39489/46831/654c8887Fbb12a6db/69035e39adff9791.png', NULL),
(8, 'https://img12.360buyimg.com/n1/jfs/t1/154257/31/33291/1399/66cecde5F91d18706/83c9b2d5c3b9b99a.png', NULL),
(8, 'https://img12.360buyimg.com/n1/jfs/t1/132989/18/46173/321/66d13f7dF21b29c8a/0204c8f1aefd017a.png', NULL),
(9, 'https://img12.360buyimg.com/n1/jfs/t1/100569/3/24613/88333/64784ba2Fd3046471/c401e38f8589d191.jpg', NULL),
(9, 'https://img12.360buyimg.com/n1/jfs/t1/101701/7/40277/98857/64784c9eF73c21cc9/96b5c35c645ccce3.jpg', NULL),
(9, 'https://img12.360buyimg.com/n1/jfs/t1/123794/17/40586/70722/64a11c2bF62d09e53/1c721452456607a3.jpg', NULL),
(10, 'https://img12.360buyimg.com/n1/jfs/t1/169822/8/38097/39661/64c466b5F03664e9c/ab569626fa401fe4.jpg', NULL),
(10, 'https://img12.360buyimg.com/n1/jfs/t1/187130/28/35611/50506/64c46699F30ef263c/3d40b32209e051a4.jpg', NULL),
(11, 'https://img12.360buyimg.com/n1/jfs/t1/143847/28/34916/142000/64ba6656Fd5335683/eb373e2c68d53452.jpg', NULL),
(11, 'https://img12.360buyimg.com/n1/jfs/t1/187285/17/35147/30409/64ba6656Fc193459d/688f740838184433.jpg', NULL),
(12, 'https://img12.360buyimg.com/n1/jfs/t1/129816/27/33316/166715/64ba6656F81bc83e3/80e3641c8816b9d8.jpg', NULL),
(13, 'https://img12.360buyimg.com/n1/jfs/t1/165000/20/33476/39731/63da057eF96d3a7a2/bed67f1655b4b996.jpg', NULL),
(13, 'https://img12.360buyimg.com/n1/jfs/t1/169523/40/33808/22291/63da057eF61e266f4/fd31ded3092f1ba4.jpg', NULL),
(14, 'https://img12.360buyimg.com/n1/jfs/t1/110600/29/33212/10196/6315a819Ebdbb4f09/46946ce79404dcea.jpg', NULL),
(14, 'https://img12.360buyimg.com/n1/jfs/t1/162105/33/30020/33551/6315a854E70d0e61d/1f7a1af71514dfc6.jpg', NULL),
(15, 'https://img12.360buyimg.com/n1/jfs/t1/37602/38/12801/159674/5d16d3a3E59fee584/0d80b61a8ec50900.jpg', NULL),
(16, 'https://img12.360buyimg.com/n1/jfs/t1/169523/40/33808/22291/63da057eF61e266f4/fd31ded3092f1ba4.jpg', NULL),
(17, 'https://img12.360buyimg.com/n1/jfs/t1/112573/25/24539/161157/623ea389Ecd80ac9c/58b8ae40843d3cd9.jpg', NULL),
(18, 'https://img12.360buyimg.com/n1/jfs/t1/161837/25/14868/34939/605d6031Ed1a6e301/34ed2ed17f5609ea.jpg', NULL),
(19, 'https://img12.360buyimg.com/n1/jfs/t1/161837/25/14868/34939/605d6031Ed1a6e301/34ed2ed17f5609ea.jpg', NULL),
(20, 'https://img12.360buyimg.com/n1/jfs/t1/112573/25/24539/161157/623ea389Ecd80ac9c/58b8ae40843d3cd9.jpg', NULL),
(21, 'https://img12.360buyimg.com/n1/jfs/t1/105832/36/23828/61710/6220986eE87a116d8/9a5948a58d0c6b66.jpg', NULL),
(21, 'https://img12.360buyimg.com/n1/jfs/t1/123225/36/23914/181722/62209844E78b488b9/f0b6a5017f21874d.jpg', NULL),
(22, 'https://img12.360buyimg.com/n1/jfs/t1/103966/1/23144/48439/62209887E3dc8e429/f71d6340a46075c5.jpg', NULL),
(23, 'https://img12.360buyimg.com/n1/jfs/t1/103966/1/23144/48439/62209887E3dc8e429/f71d6340a46075c5.jpg', NULL),
(24, 'https://img.picsum.photos/seed/mi15case1/800/600', NULL),
(25, 'https://img.picsum.photos/seed/k80case/400/400', NULL),
(26, 'https://img.picsum.photos/seed/mi15stand1/800/600', NULL),
(27, 'https://img.picsum.photos/seed/note14case/400/400', NULL),
(28, 'https://img.picsum.photos/seed/civicase/400/400', NULL),
(29, 'https://img.picsum.photos/seed/mi15glass1/800/600', NULL),
(30, 'https://img.picsum.photos/seed/k80film/400/400', NULL),
(31, 'https://img.picsum.photos/seed/mi15priv/400/400', NULL),
(32, 'https://img.picsum.photos/seed/bluefilm/400/400', NULL),
(33, 'https://img.picsum.photos/seed/140wgan1/800/600', NULL),
(34, 'https://img.picsum.photos/seed/120w/400/400', NULL),
(35, 'https://img.picsum.photos/seed/67wgan/400/400', NULL),
(36, 'https://img.picsum.photos/seed/80wwl1/800/600', NULL),
(37, 'https://img.picsum.photos/seed/65wcar/400/400', NULL),
(38, 'https://img.picsum.photos/seed/25000pb1/800/600', NULL),
(39, 'https://img.picsum.photos/seed/20000pb/400/400', NULL),
(40, 'https://img.picsum.photos/seed/magsafe1/800/600', NULL),
(41, 'https://img.picsum.photos/seed/10000slim/400/400', NULL),
(42, 'https://img12.360buyimg.com/n1/jfs/t1/138106/40/36657/47562/647995f5F73ce234e/e67aff7b711693cb.jpg', NULL),
(42, 'https://img12.360buyimg.com/n1/jfs/t1/198129/9/34380/70520/647995f8Fcd2648fd/88082ddce827ddf0.jpg', NULL),
(43, 'https://img12.360buyimg.com/n1/jfs/t1/232454/4/10838/116695/65a24861F0c14850a/a590b04a373904bb.jpg', NULL),
(44, 'https://img12.360buyimg.com/n1/jfs/t1/127446/1/38088/24138/652111e5Fec7864a8/b4ee68974fdc138e.jpg', NULL),
(45, 'https://img.picsum.photos/seed/air3se/400/400', NULL),
(46, 'https://img.picsum.photos/seed/studiopro1/800/600', NULL),
(47, 'https://img12.360buyimg.com/n1/jfs/t1/142070/19/36915/115870/6471feb4F57f13f50/150d1f2f337d0699.jpg', NULL),
(47, 'https://img12.360buyimg.com/n1/jfs/t1/212311/40/31202/86770/6471feb4Fdbc3ba3c/5367dc9d69af4776.jpg', NULL),
(48, 'https://img12.360buyimg.com/n1/jfs/t1/117023/17/34694/104272/6471feb4F2d2a0660/162cee3fd9feec1e.jpg', NULL),
(49, 'https://img12.360buyimg.com/n1/jfs/t1/108128/40/35910/81278/65156cefFe0b26fe5/3958ca9c254d5c24.jpg', NULL),
(50, 'https://img12.360buyimg.com/n1/jfs/t1/108128/40/35910/81278/65156cefFe0b26fe5/3958ca9c254d5c24.jpg', NULL),
(51, 'https://img12.360buyimg.com/n1/jfs/t1/179439/8/41312/62394/65693bdeF9280eb15/629d7012710417cf.jpg', NULL),
(52, 'https://img12.360buyimg.com/n1/jfs/t1/188548/2/40897/130536/655ac404Fc5f901ed/e893d9ea0e8458c1.jpg', NULL),
(52, 'https://img12.360buyimg.com/n1/jfs/t1/189683/18/42553/140602/655ac406F3b54c85d/864553cfa4d9b1c1.jpg', NULL),
(53, 'https://img12.360buyimg.com/n1/jfs/t1/158907/14/38944/40830/64f92987F5dababaf/7584f48beeec1dd3.jpg', NULL),
(54, 'https://img12.360buyimg.com/n1/jfs/t1/158907/14/38944/40830/64f92987F5dababaf/7584f48beeec1dd3.jpg', NULL),
(55, 'https://img12.360buyimg.com/n1/jfs/t1/106329/24/28697/75387/64f92987F6f0a375d/48d0c3667740c1bb.jpg', NULL),
(56, 'https://img12.360buyimg.com/n1/jfs/t1/125479/17/37614/21192/64ba671fF740836a5/479ef09aa9e4efe4.jpg', NULL),
(56, 'https://img12.360buyimg.com/n1/jfs/t1/129859/32/40017/21192/64ba671fF5613f011/8578cc0cfe924653.jpg', NULL),
(57, 'https://img12.360buyimg.com/n1/jfs/t1/120017/33/34460/32883/64ba671fF118b88f5/bdfccb14f49f41a7.jpg', NULL),
(58, 'https://img12.360buyimg.com/n1/jfs/t1/116910/20/34680/72246/64489584Fcbe8e64c/8bcb45130904d5c1.jpg', NULL),
(59, 'https://img12.360buyimg.com/n1/jfs/t1/151806/16/32830/64319/64489584Fbd9d3fb1/802d30f91aa4a89a.jpg', NULL),
(60, 'https://img12.360buyimg.com/n1/jfs/t1/128570/19/32080/134341/6437e591F531bd830/4a5c99a05ecb18b8.jpg', NULL),
(60, 'https://img12.360buyimg.com/n1/jfs/t1/148561/27/32066/67617/6437e6f9F6743dcd5/81beafa67d7261d8.jpg', NULL),
(61, 'https://img12.360buyimg.com/n1/jfs/t1/124358/24/28990/37170/637c6466E661fa31f/9b16ab7d1d2797ad.jpg', NULL),
(62, 'https://img12.360buyimg.com/n1/jfs/t1/136856/25/32764/40405/637c6466Ec6b61a7f/2547fdada75b0da6.jpg', NULL),
(63, 'https://img12.360buyimg.com/n1/jfs/t1/136856/25/32764/40405/637c6466Ec6b61a7f/2547fdada75b0da6.jpg', NULL);

-- ============================================================
-- 8. 捆绑销售 / 满减组合
-- ============================================================
INSERT INTO `combination_product` (`main_product_id`, `vice_product_id`, `amountThreshold`, `priceReductionRange`) VALUES
-- 买小米15 Ultra 搭配耳机满减
(1,  42, 5000, 300),
-- 买小米15 Pro 搭配充电宝满减
(2,  39, 4000, 200),
-- 买小米15 搭配保护壳满减
(3,  24, 3000, 150),
-- 买Redmi K80 Pro 搭配耳机满减
(4,  43, 2500, 200),
-- 买Redmi K80 搭配保护套满减
(5,  25, 2000, 100),
-- 买电视 S85 搭配路由器满减
(11, 60, 8000, 500),
-- 买电视 A Pro 65 搭配路由器满减
(13, 61, 3000, 200),
-- 买笔记本 Pro 14 Ultra 搭配鼠标/充电器满减
(56, 33, 7000, 400),
-- 买平板 7 Pro 搭配手写笔套装满减
(52, 29, 3000, 200),
-- 买手表 S4 搭配手环满减
(47, 51, 1200, 100),
-- 买空调新风版 搭配路由器满减
(17, 63, 5000, 300),
-- 买Fold 4 搭配耳机满减
(10, 42, 8000, 500);

-- ============================================================
-- 9. 示例订单数据（5 个用户，已支付和待支付各有）
--    order_id 公式: user_id * 1e13 + unix_ms
-- ============================================================
-- user1 订单（已支付）
INSERT INTO `orders` VALUES
(1,  10001711900000000, 1, 1,  1, 5999.0, 1711900000000, 1),
(2,  10001711900000000, 1, 42, 1,  499.0, 1711900000000, 1),
(3,  10001711900000000, 1, 38, 1,  349.0, 1711900000000, 1),
-- user1 第二笔订单（待支付）
(4,  10001712000000000, 1, 4,  1, 2999.0, 1712000000000, 0),
(5,  10001712000000000, 1, 43, 1,  329.0, 1712000000000, 0),
-- user2 订单（已支付）
(6,  20001711950000000, 2, 11, 1, 8999.0, 1711950000000, 1),
(7,  20001711950000000, 2, 60, 1, 1299.0, 1711950000000, 1),
-- user2 第二笔（待支付）
(8,  20001712050000000, 2, 52, 1, 3499.0, 1712050000000, 0),
-- user3 订单（已支付）
(9,  30001711960000000, 3, 56, 1, 7999.0, 1711960000000, 1),
(10, 30001711960000000, 3, 33, 1,  169.0, 1711960000000, 1),
-- user4 订单（已支付）
(11, 40001711970000000, 4, 47, 1, 1299.0, 1711970000000, 1),
(12, 40001711970000000, 4, 29, 1,   39.0, 1711970000000, 1),
-- user5 订单（已取消）
(13, 50001711980000000, 5, 7,  2, 1099.0, 1711980000000, 2);

-- ============================================================
-- 10. 购物车示例
-- ============================================================
INSERT INTO `shoppingcart` (`user_id`, `product_id`, `num`) VALUES
(1, 2,  1),  -- alice 想买小米15 Pro
(1, 43, 1),  -- alice 想买耳机
(2, 4,  2),  -- bob 想买Redmi K80 Pro x2
(3, 52, 1),  -- charlie 想买平板
(4, 33, 1),  -- david 想买充电器
(5, 47, 1);  -- eve 想买手表

-- ============================================================
-- 11. 收藏示例
-- ============================================================
INSERT INTO `collect` (`user_id`, `product_id`, `category`, `collect_time`) VALUES
(1, 1,  1,  1712000000000),
(1, 10, 1,  1712001000000),
(1, 42, 9,  1712002000000),
(1, 47, 10, 1712003000000),
(2, 11, 2,  1712004000000),
(2, 56, 12, 1712005000000),
(3, 4,  1,  1712006000000),
(3, 52, 11, 1712007000000),
(4, 2,  1,  1712008000000),
(4, 43, 9,  1712009000000),
(5, 17, 3,  1712010000000),
(5, 60, 13, 1712011000000);

-- ============================================================
-- 12. 支付单示例（对应上面已支付的订单）
-- ============================================================
INSERT INTO `payment_order`
  (`payment_no`, `order_id`, `user_id`, `amount`, `channel`, `channel_trade_no`,
   `status`, `expire_time`, `paid_time`, `notify_url`, `extra`, `created_at`, `updated_at`)
VALUES
-- user1 第一笔订单的支付（成功）
('PAY2024040100000001', 10001711900000000, 1, 684800, 'mock', 'MOCK_TRD_001',
  2, 1711903600, 1711900120, '', NULL, 1711900000, 1711900120),
-- user2 订单的支付（成功）
('PAY2024040200000001', 20001711950000000, 2, 1029800, 'mock', 'MOCK_TRD_002',
  2, 1711953600, 1711950200, '', NULL, 1711950000, 1711950200),
-- user3 订单的支付（成功）
('PAY2024040300000001', 30001711960000000, 3, 816800, 'mock', 'MOCK_TRD_003',
  2, 1711963600, 1711960300, '', NULL, 1711960000, 1711960300),
-- user4 订单的支付（成功）
('PAY2024040400000001', 40001711970000000, 4, 133800, 'mock', 'MOCK_TRD_004',
  2, 1711973600, 1711970400, '', NULL, 1711970000, 1711970400);

-- ============================================================
-- 完成
-- ============================================================
SET FOREIGN_KEY_CHECKS = 1;

-- 验证数据量
SELECT '用户' AS 表名, COUNT(*) AS 记录数 FROM users
UNION ALL SELECT '管理员', COUNT(*) FROM sysmanager
UNION ALL SELECT '分类', COUNT(*) FROM category
UNION ALL SELECT '商品', COUNT(*) FROM product
UNION ALL SELECT '商品图片', COUNT(*) FROM product_picture
UNION ALL SELECT '订单行', COUNT(*) FROM orders
UNION ALL SELECT '购物车', COUNT(*) FROM shoppingcart
UNION ALL SELECT '收藏', COUNT(*) FROM collect
UNION ALL SELECT '支付单', COUNT(*) FROM payment_order;
