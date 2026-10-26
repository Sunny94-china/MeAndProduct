-- 2026-09-23 MeAndProduct P3 T10 — P1 基础 10 表之上的 P3 字段补齐（非 DELIMITER 单语句，兼容 golang-migrate 驱动）：
-- * reports 新增 hide_target + updated_at
-- * founder_profiles 补齐 one_liner / bio_md / bio_html / location / founding_stage / founding_started_at / skills / interests / is_looking_for_cofounder / looking_for_roles / is_visible
-- * products 补齐 logo_url / cover_url / github_url / twitter_url / producthunt_url / pricing_url / sort_order / launched_at
-- * tags 补齐 founder_count / product_count + uk_tag_slug
-- * links 补齐 label + uk_source_type_link
-- 幂等：利用 PREPARE + information_schema 判空，ADD/DROP INDEX/COLUMN 重复执行安全。

SET @mp_p3t10_db = DATABASE();

-- 1. reports.hide_target
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='reports' AND COLUMN_NAME='hide_target');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `reports` ADD COLUMN `hide_target` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '处理时是否隐藏目标（下架/屏蔽）' AFTER `handler_note`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 2. reports.updated_at
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='reports' AND COLUMN_NAME='updated_at');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `reports` ADD COLUMN `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER `created_at`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 3. founder_profiles.one_liner
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='one_liner');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `founder_profiles` ADD COLUMN `one_liner` VARCHAR(200) NULL COMMENT '一句话自我介绍' AFTER `avatar_url`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 4. founder_profiles.location
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='location');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `founder_profiles` ADD COLUMN `location` VARCHAR(100) NULL COMMENT '所在地' AFTER `one_liner`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 5. founder_profiles.bio_md
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='bio_md');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `founder_profiles` ADD COLUMN `bio_md` MEDIUMTEXT NULL COMMENT '个人介绍（Markdown 原文）' AFTER `location`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 6. founder_profiles.bio_html
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='bio_html');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `founder_profiles` ADD COLUMN `bio_html` MEDIUMTEXT NULL COMMENT '个人介绍（渲染后 HTML，服务端生成）' AFTER `bio_md`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 7. founder_profiles.founding_stage
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='founding_stage');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `founder_profiles` ADD COLUMN `founding_stage` VARCHAR(50) NULL COMMENT '创业阶段：idea/validation/traction/scale' AFTER `custom_links`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 8. founder_profiles.founding_started_at
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='founding_started_at');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `founder_profiles` ADD COLUMN `founding_started_at` DATETIME NULL COMMENT '开始创业时间' AFTER `founding_stage`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 9. founder_profiles.skills
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='skills');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `founder_profiles` ADD COLUMN `skills` JSON NULL COMMENT '技能标签数组'",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 10. founder_profiles.interests
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='interests');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `founder_profiles` ADD COLUMN `interests` JSON NULL COMMENT '感兴趣的方向'",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 11. founder_profiles.is_looking_for_cofounder
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='is_looking_for_cofounder');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `founder_profiles` ADD COLUMN `is_looking_for_cofounder` TINYINT(1) NOT NULL DEFAULT 0 AFTER `interests`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 12. founder_profiles.looking_for_roles
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='looking_for_roles');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `founder_profiles` ADD COLUMN `looking_for_roles` JSON NULL COMMENT '寻找的合伙人角色' AFTER `is_looking_for_cofounder`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 13. founder_profiles.is_visible
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='is_visible');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `founder_profiles` ADD COLUMN `is_visible` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否公开可见' AFTER `updated_at`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 14. products.logo_url
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='logo_url');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `products` ADD COLUMN `logo_url` VARCHAR(500) NULL COMMENT '产品 Logo 图片' AFTER `name`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 15. products.cover_url
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='cover_url');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `products` ADD COLUMN `cover_url` VARCHAR(500) NULL COMMENT '产品封面大图' AFTER `logo_url`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 16. products.github_url
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='github_url');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `products` ADD COLUMN `github_url` VARCHAR(500) NULL COMMENT 'GitHub 仓库主页' AFTER `official_url`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 17. products.twitter_url
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='twitter_url');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `products` ADD COLUMN `twitter_url` VARCHAR(500) NULL COMMENT 'Twitter/X 主页' AFTER `github_url`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 18. products.producthunt_url
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='producthunt_url');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `products` ADD COLUMN `producthunt_url` VARCHAR(500) NULL COMMENT 'Product Hunt 主页' AFTER `twitter_url`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 19. products.pricing_url
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='pricing_url');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `products` ADD COLUMN `pricing_url` VARCHAR(500) NULL COMMENT '定价页' AFTER `status`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 20. products.sort_order
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='sort_order');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `products` ADD COLUMN `sort_order` INT NOT NULL DEFAULT 0 COMMENT '创始人后台排序，越大越靠前' AFTER `favorite_count`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 21. products.launched_at
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='launched_at');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `products` ADD COLUMN `launched_at` DATETIME NULL COMMENT '上线时间' AFTER `sort_order`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 22. tags.founder_count
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='tags' AND COLUMN_NAME='founder_count');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `tags` ADD COLUMN `founder_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联创始人数量（冗余计数）' AFTER `is_active`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 23. tags.product_count
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='tags' AND COLUMN_NAME='product_count');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `tags` ADD COLUMN `product_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联产品数量（冗余计数）' AFTER `founder_count`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 24. tags.uk_tag_slug
SET @v_exists = (SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='tags' AND INDEX_NAME='uk_tag_slug');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `tags` ADD UNIQUE KEY `uk_tag_slug` (`slug`)",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 25. links.label
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='links' AND COLUMN_NAME='label');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `links` ADD COLUMN `label` VARCHAR(200) NULL COMMENT '用户自定义显示名' AFTER `link_type`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 26. links.uk_source_type_link
SET @v_exists = (SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='links' AND INDEX_NAME='uk_source_type_link');
SET @mp_p3t10_sql = IF(@v_exists=0,
  "ALTER TABLE `links` ADD UNIQUE KEY `uk_source_type_link` (`source_type`,`source_id`,`link_type`,`url`)",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
