-- +migrate Up
-- 00001_init — MeAndProduct 初始化 10 张表

-- ================================================================
-- 1. 用户账户表
-- ================================================================
CREATE TABLE IF NOT EXISTS users (
  id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  username        VARCHAR(50)  NOT NULL COMMENT '登录用户名，URL友好',
  email           VARCHAR(100) NOT NULL COMMENT '邮箱',
  password_hash   VARCHAR(255) NOT NULL COMMENT 'bcrypt哈希',
  `role`          ENUM('founder','admin') NOT NULL DEFAULT 'founder',
  status          ENUM('pending','active','suspended') NOT NULL DEFAULT 'pending'
                  COMMENT '待审核(自助申请)/正常/封禁',
  avatar_url      VARCHAR(500) NULL,
  register_source ENUM('self','invite') NOT NULL COMMENT '注册来源',
  invite_code_id  BIGINT UNSIGNED NULL COMMENT '使用的邀请码ID',
  last_login_at   DATETIME NULL,
  created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at      DATETIME NULL,
  UNIQUE KEY uk_username (username),
  UNIQUE KEY uk_email (email),
  KEY idx_role_status (`role`, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户账户表';


-- ================================================================
-- 2. 创业者名片主表
-- ================================================================
CREATE TABLE IF NOT EXISTS founder_profiles (
  id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id         BIGINT UNSIGNED NOT NULL,
  display_name    VARCHAR(100) NOT NULL COMMENT '显示名称',
  slug            VARCHAR(100) NOT NULL COMMENT 'URL短标识',
  title           VARCHAR(100) NULL COMMENT '头衔',
  bio             VARCHAR(500) NULL COMMENT '一句话简介',
  bio_long        TEXT NULL COMMENT '详细介绍（Markdown）',
  location        VARCHAR(100) NULL,
  avatar_url      VARCHAR(500) NULL,
  cover_url       VARCHAR(500) NULL COMMENT '名片封面图',
  contact_email   VARCHAR(100) NULL COMMENT '公开联系邮箱',
  website_url     VARCHAR(500) NULL COMMENT '个人官网',
  github_url      VARCHAR(500) NULL,
  twitter_url     VARCHAR(500) NULL,
  linkedin_url    VARCHAR(500) NULL,
  wechat_id       VARCHAR(100) NULL,
  zhihu_url       VARCHAR(500) NULL,
  xiaohongshu_url VARCHAR(500) NULL,
  producthunt_url VARCHAR(500) NULL,
  custom_links    JSON NULL COMMENT '自定义外链：[{"name":"x","url":"y"}]',
  view_count      INT UNSIGNED NOT NULL DEFAULT 0,
  favorite_count  INT UNSIGNED NOT NULL DEFAULT 0,
  is_visible      TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否对外显示',
  created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at      DATETIME NULL,
  UNIQUE KEY uk_user_id (user_id),
  UNIQUE KEY uk_slug (slug),
  KEY idx_visible (is_visible),
  KEY idx_view_count (view_count DESC),
  KEY idx_favorite_count (favorite_count DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='创业者名片';


-- ================================================================
-- 3. 产品表
-- ================================================================
CREATE TABLE IF NOT EXISTS products (
  id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  founder_id      BIGINT UNSIGNED NOT NULL,
  name            VARCHAR(200) NOT NULL,
  slug            VARCHAR(200) NOT NULL COMMENT '同创业者下唯一',
  logo_url        VARCHAR(500) NULL,
  cover_url       VARCHAR(500) NULL,
  tagline         VARCHAR(200) NOT NULL COMMENT '一句话标语',
  `description`   TEXT NULL COMMENT '产品描述（Markdown）',
  official_url    VARCHAR(500) NOT NULL COMMENT '产品官网',
  github_url      VARCHAR(500) NULL,
  producthunt_url VARCHAR(500) NULL,
  twitter_url     VARCHAR(500) NULL,
  `status`        ENUM('idea','beta','launched','growing','archived')
                  NOT NULL DEFAULT 'idea' COMMENT '产品阶段',
  launch_date     DATE NULL,
  pricing_url     VARCHAR(500) NULL,
  custom_links    JSON NULL,
  view_count      INT UNSIGNED NOT NULL DEFAULT 0,
  favorite_count  INT UNSIGNED NOT NULL DEFAULT 0,
  sort_order      INT NOT NULL DEFAULT 0 COMMENT '展示排序',
  is_visible      TINYINT(1) NOT NULL DEFAULT 1,
  created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at      DATETIME NULL,
  UNIQUE KEY uk_founder_slug (founder_id, slug),
  KEY idx_founder_id (founder_id),
  KEY idx_status (`status`),
  KEY idx_view_count (view_count DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品';


-- ================================================================
-- 4. 标签分类表
-- ================================================================
CREATE TABLE IF NOT EXISTS tags (
  id            BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  name          VARCHAR(50) NOT NULL,
  slug          VARCHAR(50) NOT NULL,
  `type`        ENUM('industry','stage','role','other') NOT NULL DEFAULT 'industry',
  icon_url      VARCHAR(500) NULL,
  sort_order    INT NOT NULL DEFAULT 0,
  founder_count INT UNSIGNED NOT NULL DEFAULT 0,
  product_count INT UNSIGNED NOT NULL DEFAULT 0,
  is_active     TINYINT(1) NOT NULL DEFAULT 1,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_name (name),
  UNIQUE KEY uk_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签';


-- ================================================================
-- 5 & 6. 多对多：创业者-标签 / 产品-标签
-- ================================================================
CREATE TABLE IF NOT EXISTS founder_tags (
  founder_id BIGINT UNSIGNED NOT NULL,
  tag_id     BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (founder_id, tag_id),
  KEY idx_tag (tag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS product_tags (
  product_id BIGINT UNSIGNED NOT NULL,
  tag_id     BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (product_id, tag_id),
  KEY idx_tag (tag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- ================================================================
-- 7. 收藏表
-- ================================================================
CREATE TABLE IF NOT EXISTS favorites (
  id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id     BIGINT UNSIGNED NOT NULL,
  target_type ENUM('founder','product') NOT NULL,
  target_id   BIGINT UNSIGNED NOT NULL,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_user_target (user_id, target_type, target_id),
  KEY idx_user (user_id),
  KEY idx_target (target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='收藏';


-- ================================================================
-- 8. 邀请码表
-- ================================================================
CREATE TABLE IF NOT EXISTS invite_codes (
  id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  code        VARCHAR(32) NOT NULL,
  created_by  BIGINT UNSIGNED NOT NULL COMMENT '发起人',
  max_uses    INT NOT NULL DEFAULT 1,
  used_count  INT NOT NULL DEFAULT 0,
  expires_at  DATETIME NULL COMMENT 'NULL=永不过期',
  status      ENUM('active','expired','used','revoked') NOT NULL DEFAULT 'active',
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  used_at     DATETIME NULL,
  UNIQUE KEY uk_code (code),
  KEY idx_created_by (created_by),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邀请码';


-- ================================================================
-- 9. 举报表
-- ================================================================
CREATE TABLE IF NOT EXISTS reports (
  id            BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  reporter_id   BIGINT UNSIGNED NULL COMMENT 'NULL=匿名',
  target_type   ENUM('founder','product') NOT NULL,
  target_id     BIGINT UNSIGNED NOT NULL,
  reason_type   ENUM('spam','fraud','inappropriate','copyright','other') NOT NULL,
  reason_detail TEXT NOT NULL,
  contact_info  VARCHAR(200) NULL,
  `status`      ENUM('pending','reviewed','rejected','resolved') NOT NULL DEFAULT 'pending',
  handler_id    BIGINT UNSIGNED NULL,
  handler_note  TEXT NULL,
  handled_at    DATETIME NULL,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_target (target_type, target_id),
  KEY idx_status (`status`),
  KEY idx_reporter (reporter_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='举报';


-- ================================================================
-- 10. 通用外链表（点击统计用）
-- ================================================================
CREATE TABLE IF NOT EXISTS links (
  id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  source_type ENUM('founder','product') NOT NULL,
  source_id   BIGINT UNSIGNED NOT NULL,
  link_type   VARCHAR(50) NOT NULL COMMENT 'github/official/custom...',
  url         VARCHAR(500) NOT NULL,
  click_count INT UNSIGNED NOT NULL DEFAULT 0,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_source (source_type, source_id),
  KEY idx_link_type (link_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外链点击统计';


-- ================================================================
-- 初始种子：默认管理员（用户名=admin, 邮箱=admin@local, 密码=Admin@123 需自行修改）
-- 密码使用 bcrypt 哈希，此处占位；建议首次启动后通过程序/CLI 修改。
-- 为避免明文，不在这里 INSERT password_hash。应用层提供 cli 命令或在 P2 Auth 阶段处理。
-- ================================================================
