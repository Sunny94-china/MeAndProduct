-- 2026-09-23 P3 T10 字段补齐回滚（非 DELIMITER，PREPARE 幂等判空，重复执行安全）
SET @mp_p3t10_db = DATABASE();

-- 26. links.uk_source_type_link
SET @v_exists = (SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='links' AND INDEX_NAME='uk_source_type_link');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `links` DROP INDEX `uk_source_type_link`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 25. links.label
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='links' AND COLUMN_NAME='label');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `links` DROP COLUMN `label`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 24. tags.uk_tag_slug
SET @v_exists = (SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='tags' AND INDEX_NAME='uk_tag_slug');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `tags` DROP INDEX `uk_tag_slug`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 23. tags.product_count
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='tags' AND COLUMN_NAME='product_count');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `tags` DROP COLUMN `product_count`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 22. tags.founder_count
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='tags' AND COLUMN_NAME='founder_count');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `tags` DROP COLUMN `founder_count`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 21. products.launched_at
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='launched_at');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `products` DROP COLUMN `launched_at`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 20. products.sort_order
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='sort_order');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `products` DROP COLUMN `sort_order`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 19. products.pricing_url
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='pricing_url');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `products` DROP COLUMN `pricing_url`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 18. products.producthunt_url
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='producthunt_url');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `products` DROP COLUMN `producthunt_url`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 17. products.twitter_url
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='twitter_url');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `products` DROP COLUMN `twitter_url`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 16. products.github_url
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='github_url');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `products` DROP COLUMN `github_url`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 15. products.cover_url
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='cover_url');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `products` DROP COLUMN `cover_url`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 14. products.logo_url
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='products' AND COLUMN_NAME='logo_url');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `products` DROP COLUMN `logo_url`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 13. founder_profiles.is_visible
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='is_visible');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `founder_profiles` DROP COLUMN `is_visible`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 12. founder_profiles.looking_for_roles
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='looking_for_roles');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `founder_profiles` DROP COLUMN `looking_for_roles`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 11. founder_profiles.is_looking_for_cofounder
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='is_looking_for_cofounder');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `founder_profiles` DROP COLUMN `is_looking_for_cofounder`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 10. founder_profiles.interests
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='interests');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `founder_profiles` DROP COLUMN `interests`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 9. founder_profiles.skills
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='skills');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `founder_profiles` DROP COLUMN `skills`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 8. founder_profiles.founding_started_at
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='founding_started_at');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `founder_profiles` DROP COLUMN `founding_started_at`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 7. founder_profiles.founding_stage
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='founding_stage');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `founder_profiles` DROP COLUMN `founding_stage`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 6. founder_profiles.bio_html
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='bio_html');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `founder_profiles` DROP COLUMN `bio_html`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 5. founder_profiles.bio_md
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='bio_md');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `founder_profiles` DROP COLUMN `bio_md`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 4. founder_profiles.location
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='location');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `founder_profiles` DROP COLUMN `location`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 3. founder_profiles.one_liner
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='founder_profiles' AND COLUMN_NAME='one_liner');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `founder_profiles` DROP COLUMN `one_liner`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 2. reports.updated_at
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='reports' AND COLUMN_NAME='updated_at');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `reports` DROP COLUMN `updated_at`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 1. reports.hide_target
SET @v_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@mp_p3t10_db AND TABLE_NAME='reports' AND COLUMN_NAME='hide_target');
SET @mp_p3t10_sql = IF(@v_exists=1,
  "ALTER TABLE `reports` DROP COLUMN `hide_target`",
  'SELECT 1');
PREPARE stmt FROM @mp_p3t10_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
