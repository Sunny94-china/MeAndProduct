-- Down：理论上无法回滚「解码」操作（一旦修复 name 已变为正确 UTF-8，无法再还原为乱码字节）。
-- 为严格幂等且避免数据损坏，此处做 no-op：仅当 name 仍有乱码特征时才反向编码（即 up 的逆操作），
-- 但正常情况下 up 已执行完后不会命中，因此 down 基本是 no-op，符合 migration 安全性。

SET @mp_tags_db = DATABASE();

SET @mp_rollback_sql = CONCAT(
  'UPDATE `tags` ',
  'SET `name` = CONVERT(CONVERT(BINARY `name` USING utf8mb4) USING latin1) ',
  'WHERE `id` BETWEEN 1 AND 6 ',
  '  AND HEX(`name`) NOT REGEXP ''^(..)*(C3|C2|C5)(..)*$'' ',
  '  AND OCTET_LENGTH(`name`) = CHAR_LENGTH(`name`) + (CHAR_LENGTH(`name`) - LENGTH(CAST(`name` AS CHAR CHARACTER SET latin1)))'
);

PREPARE stmt FROM @mp_rollback_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT 1;
