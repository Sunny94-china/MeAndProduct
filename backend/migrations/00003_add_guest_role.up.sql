-- 2026-09-24 MeAndProduct — 给 users.role ENUM 增加 'guest' 值（浏览用户角色）。
-- 幂等：用 information_schema.COLUMN_TYPE 检查 ENUM 是否已经包含 'guest'，包含则跳过。

SET @mp_guest_db = DATABASE();

SET @mp_role_enum = (
  SELECT COLUMN_TYPE
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=@mp_guest_db
    AND TABLE_NAME='users'
    AND COLUMN_NAME='role'
);

SET @mp_has_guest = (@mp_role_enum IS NOT NULL AND @mp_role_enum LIKE '%guest%');

SET @mp_guest_sql = IF(@mp_has_guest=1,
  'SELECT 1',
  "ALTER TABLE `users` MODIFY COLUMN `role` ENUM('founder','admin','guest') NOT NULL DEFAULT 'founder' COMMENT '角色'"
);

PREPARE stmt FROM @mp_guest_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
