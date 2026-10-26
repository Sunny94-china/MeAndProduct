-- Down：移除 guest 枚举（把 role 回退到 founder/admin 两值，前提是所有 guest 行必须先被处理为 founder）。
-- 为了避免数据丢失，这里只做一个 no-op 的安全 down；真要降级先手动把 users.role='guest' 改成其他值再跑。

SET @mp_guest_db = DATABASE();

SET @mp_role_enum = (
  SELECT COLUMN_TYPE
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=@mp_guest_db AND TABLE_NAME='users' AND COLUMN_NAME='role'
);

SET @mp_has_guest = (@mp_role_enum IS NOT NULL AND @mp_role_enum LIKE '%guest%');

SET @mp_guest_count = IF(@mp_has_guest=1,
  (SELECT COUNT(*) FROM `users` WHERE `role`='guest'),
  0
);

-- 仅当没有 guest 用户时才允许 down
SET @mp_guest_sql = IF(@mp_has_guest=1 AND @mp_guest_count=0,
  "ALTER TABLE `users` MODIFY COLUMN `role` ENUM('founder','admin') NOT NULL DEFAULT 'founder' COMMENT '角色'",
  'SELECT 1'
);

PREPARE stmt FROM @mp_guest_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
