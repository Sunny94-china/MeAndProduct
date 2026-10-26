-- 2026-09-28 MeAndProduct — 修复 tags 表 id=1~6 种子 name 字段的双重 UTF-8 编码乱码。
-- 根因：历史写入时把已是 UTF-8 的字节误当作 Latin-1 再用 UTF-8 连接二次编码，导致 name HEX 出现 C3/C2/C5 前缀异常。
-- 幂等策略：仅对 HEX 匹配乱码特征的行执行 CONVERT 逆向解码；已修复的行 HEX 无该特征，会被 WHERE 过滤掉，重复执行无损。

SET @mp_tags_db = DATABASE();

-- ============================================================
-- 1) 逐行安全修复 id=1~6：仅当该行 HEX 仍含乱码特征（C3/C2/C5 前缀）时才解码
--    标准逆向：CONVERT(BINARY CONVERT(name USING latin1) USING utf8mb4)
--    注意：golang-migrate 在 multiStatements=true 下不支持 DELIMITER，
--    因此用 PREPARE/EXECUTE 拼接动态 SQL 的方式实现条件执行（与 00003 一致）。
-- ============================================================

SET @mp_fix_sql = CONCAT(
  'UPDATE `tags` ',
  'SET `name` = CONVERT(BINARY CONVERT(`name` USING latin1) USING utf8mb4) ',
  'WHERE `id` BETWEEN 1 AND 6 ',
  '  AND HEX(`name`) REGEXP ''^(..)*(C3|C2|C5)(..)*$'' ',
  '  AND OCTET_LENGTH(`name`) > CHAR_LENGTH(`name`) * 2'
);

PREPARE stmt FROM @mp_fix_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 2) 可选：同步修复 tags.slug（若 slug 也出现类似乱码，通常不需要，
--    目前 6 条种子 slug 均为纯 ASCII，所以仅做 no-op 占位，避免误改）。
-- ============================================================

SELECT 1;
