-- +migrate Down
-- 00001_init — 回滚全部 10 张表（顺序与创建相反，避免外键问题）

DROP TABLE IF EXISTS links;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS invite_codes;
DROP TABLE IF EXISTS favorites;
DROP TABLE IF EXISTS product_tags;
DROP TABLE IF EXISTS founder_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS founder_profiles;
DROP TABLE IF EXISTS users;
