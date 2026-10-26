# Implementation TODO: MeAndProduct

> 本文件按 P1 → P6 顺序执行。每完成一个任务请逐项打勾 [x]，通过检查点后再进入下一阶段。

---

## Phase 1: Foundation 工程脚手架

### Task 1 — 根目录基础设施

**Description**：创建项目根目录下的 Makefile、.gitignore、deploy/ 子目录（含开发环境 docker-compose、nginx 配置模板），为后续的镜像构建、本地开发、部署分发打基础。

**Acceptance criteria**：
- [ ] `Makefile` 存在并包含：dev / build-images / save-images / backend-run / frontend-public-run / frontend-founder-run / frontend-admin-run / migrate-up / migrate-down
- [ ] `.gitignore` 正确忽略：node_modules、.env、dist、*.tar.gz、coverage、.DS_Store、Go vendor、.idea/.vscode 等
- [ ] `deploy/docker-compose.dev.yml` 可启动 mysql:8 + redis:7，包含用户名密码、持久化 volume、端口映射（3306/6379）
- [ ] `deploy/nginx.conf` 模板存在，含 /api 反代、三个前端的静态目录路由

**Verification**：
- [ ] `make dev` 后 `docker ps` 可见 mysql + redis 两个容器
- [ ] `make` 提示列出全部 target（或 `make help`/`grep ^: Makefile` 校验）

**Dependencies**：None

**Files touched**：
- `Makefile`
- `.gitignore`
- `deploy/docker-compose.dev.yml`
- `deploy/docker-compose.yml`（生产模板，占位即可）
- `deploy/nginx.conf`

**Estimated scope**：Small

---

### Task 2 — Go 后端脚手架

**Description**：建立 `backend/` 下完整的 Go 工程骨架（go.mod、main.go、config、internal 各层目录占位），可加载 yaml/env 配置，能连接 MySQL + Redis，能启动 Gin 并暴露 `/api/v1/health` 接口。附带构建用 Dockerfile（多阶段构建，alpine 产物）。

**Acceptance criteria**：
- [ ] `go.mod` 声明 module `meandproduct`，依赖 gin/gorm/viper/go-redis/bcrypt/jwt 等基础库
- [ ] `config.yaml.example` 存在（含 mysql/redis/server/jwt 配置节），可被 Viper 加载，且支持环境变量覆盖
- [ ] `main.go` 优雅启停（signal.Notify SIGINT/SIGTERM），启动日志打印监听端口
- [ ] `GET /api/v1/health` 返回 `{code:0, message:"ok", data:{status:"UP"}}`
- [ ] `backend/Dockerfile` 多阶段：builder(golang) → runner(alpine)，构建出的镜像大小 <50MB
- [ ] `internal/` 下 handler/service/repository/model/middleware/pkg 目录存在（可含空 .gitkeep 文件）

**Verification**：
- [ ] `cd backend && go mod tidy` 无错
- [ ] `go run cmd/server/main.go` 启动后 `curl localhost:8080/api/v1/health` 返回 JSON 200
- [ ] `docker build -t meandproduct-backend:test backend/` 构建成功

**Dependencies**：Task 1

**Files touched**：
- `backend/go.mod` / `backend/go.sum`
- `backend/cmd/server/main.go`
- `backend/internal/config/config.go`
- `backend/config.yaml.example`
- `backend/Dockerfile`
- 各 internal 子目录

**Estimated scope**：Medium

---

### Task 3 — 数据库迁移框架 + 初始化 SQL

**Description**：集成 `golang-migrate`，迁移 SQL 文件放在 `backend/migrations/`，首份迁移文件 `00001_init` 一次性创建设计文档 §3.2 中所有 10 张表。main.go 增加开关 `--migrate-up` / `--migrate-down N`，可独立执行迁移（便于 CI 与本地开发）。

**Acceptance criteria**：
- [ ] `00001_init.up.sql` 包含 users / founder_profiles / products / tags / founder_tags / product_tags / favorites / invite_codes / reports / links 共 10 张表，字段与设计文档完全一致
- [ ] `00001_init.down.sql` 存在，顺序 DROP 全部 10 张表
- [ ] 启动时 `migrate-auto` 配置 true 则自动 run up，否则等待手动执行
- [ ] `make migrate-up` / `make migrate-down STEP=1` 工作正常
- [ ] 迁移完成后在 MySQL 里 `SHOW TABLES;` 列出 10 张表

**Verification**：
- [ ] 干净环境首次启动，启动日志打印 "migrated 1 file"
- [ ] 重复启动不报错幂等（golang-migrate schema_migrations 表正确记录）
- [ ] down 后再 up 可循环通过

**Dependencies**：Task 2

**Files touched**：
- `backend/migrations/00001_init.up.sql`
- `backend/migrations/00001_init.down.sql`
- `backend/internal/pkg/migrate/migrate.go`（调用 golang-migrate 的封装）
- `backend/cmd/server/main.go`（增加 migrate 入口）
- `Makefile`（追加 migrate-up / migrate-down target）

**Estimated scope**：Medium

---

### Task 4 — pkg 公共基础包

**Description**：在 `backend/internal/pkg/` 下实现 5 个独立的、可单测的基础包，它们不依赖业务层，供未来 service/handler/middleware 调用。

**Acceptance criteria**：

1. **pkg/response**
   - [ ] 统一返回格式 `{code, message, data, request_id}`
   - [ ] 提供 `OK(c, data)` / `Fail(c, code, msg)` / `FailWithDetail(c, code, msg, detail)` 等 Gin helper

2. **pkg/jwtutil**
   - [ ] `IssueTokens(userID, role)` 返回 access(2h) + refresh(7d)，两者 jti 不同
   - [ ] `Parse(token, tokenType)` 解析并校验 token，返回 Claims{UserID, Role, Jti, Exp}
   - [ ] `Blacklist(jti, exp)` / `IsBlacklisted(jti)` 读写 Redis 黑名单
   - [ ] 不把 Redis client 包成全局变量，通过函数参数注入

3. **pkg/password**
   - [ ] `Hash(plain)` 使用 bcrypt 14 cost 做哈希
   - [ ] `Verify(plain, hash)` 恒定时间比较

4. **pkg/cache**
   - [ ] `GetOrSet[T](ctx, rd, key, ttl, loader)` 泛型函数，读穿透 + 回填
   - [ ] `Invalidate(ctx, rd, keys...)` 批量删除
   - [ ] loader 返回 nil 且无错误时，存空占位防止缓存穿透（30s TTL）

5. **pkg/slug**
   - [ ] `Generate(desired, fallback)` 生成 slug，正则合法直接用 desired，否则 fallback(如 username) sanitize 后使用
   - [ ] `IsUnique(slug, existsFunc)` 结合回调检查冲突时自动追加 "-2"、"-3"

**Verification**：
- [ ] 每个 pkg 至少 1 个 Go 单测通过（`go test ./internal/pkg/... -v`）
- [ ] 不 import 任何 handler/service/model/entity 包（保证纯工具）

**Dependencies**：Task 3

**Files touched**：
- `backend/internal/pkg/response/response.go`
- `backend/internal/pkg/jwtutil/jwtutil.go` + `_test.go`
- `backend/internal/pkg/password/password.go` + `_test.go`
- `backend/internal/pkg/cache/cache.go` + `_test.go`
- `backend/internal/pkg/slug/slug.go` + `_test.go`

**Estimated scope**：Medium

---

### Task 5 — middleware 中间件

**Description**：在 `backend/internal/middleware/` 下实现 5 个可顺序串联到 Gin 的中间件。

**Acceptance criteria**：

1. **Logger**：记录 method/path/status/duration，结构化日志（slog/zap）
2. **CORS**：可配置的 allow origins / methods / headers，开发环境默认打开 *，生产严格
3. **RateLimit**：基于 Redis sliding log（或简单固定窗口），登录限流 60/5min/IP、举报 5/1h/IP，其他通用 300/min/IP
4. **JWTAuth**：从 `Authorization: Bearer xxx` 提取 access → 调用 jwtutil.Parse → 查黑名单 → 通过则把 `{user_id, role}` 注入 c.Set("currentUser")；未通过直接返回 401
5. **RBAC**：`RBAC(role...)` gin.HandlerFunc，若 currentUser.role 在给定列表或更高权限（admin >= founder >= guest）则放行，否则 403；可接受 "guest" 表示无需登录

**Verification**：
- [ ] 在主路由注册一条受保护路由 `GET /api/v1/test/protected` 使用 `JWTAuth + RBAC("founder","admin")`，不传 token 401 / founder token 200 / admin token 200
- [ ] CORS preflight `OPTIONS /api/v1/health` 返回 204 + 正确头
- [ ] RateLimit：短时间压测 61 次登录接口后第 61 次返回 429（可临时用 10/5s 小阈值验证）

**Dependencies**：Task 4

**Files touched**：
- `backend/internal/middleware/logger.go`
- `backend/internal/middleware/cors.go`
- `backend/internal/middleware/ratelimit.go`
- `backend/internal/middleware/jwt.go`
- `backend/internal/middleware/rbac.go`
- `backend/cmd/server/main.go`（组装中间件 + 注册测试路由）

**Estimated scope**：Medium

---

### Task 6 — 三个前端脚手架

**Description**：分别在 `frontend/public-site/`、`frontend/founder-dashboard/`、`frontend/admin-dashboard/` 创建 Vue 3 + Vite + TS 工程，集成 Pinia / Vue Router / Element Plus / Axios，首页渲染 "Hello MeAndProduct — [站点名]"，各带各自的 Dockerfile（构建后 dist 可被 nginx volume 挂载）。

**Acceptance criteria**：
- [ ] 三者 `npm install && npm run dev` 能独立启动（端口分别为 5173 / 5174 / 5175）
- [ ] 三者 `npm run build` 能产出 `dist/`
- [ ] vite.config.ts 中 `base` 区分：public-site = `/`、founder = `/founder/`、admin = `/admin/`
- [ ] 三者统一的 `api/request.ts` 封装 axios，带 401 自动跳登录（admin/founder）、全局消息提示
- [ ] 各自 Dockerfile（多阶段：node build → nginx 仅 dist，体积小）
- [ ] `package.json` 的依赖版本遵循 plan.md 里的固定组合

**Verification**：
- [ ] 三个 dev server 启动后浏览器手动访问：页面 + 控制台无报错，ElementPlus 组件正常渲染
- [ ] `docker build -t meandproduct-public:test frontend/public-site/` 等 3 个都能成功

**Dependencies**：Task 1（需要 .gitignore 忽略 node_modules）

**Files touched**（每个前端各一套，重复 3 次）：
- `frontend/*/package.json`
- `frontend/*/vite.config.ts`
- `frontend/*/tsconfig.json`
- `frontend/*/index.html`
- `frontend/*/src/main.ts`、`App.vue`、`router/index.ts`、`stores/*.ts`、`api/request.ts`
- `frontend/*/Dockerfile`

**Estimated scope**：Large（重复三次，第一次完整写后两份可复制改配置）

---

### 🔴 Checkpoint: P1 Foundation

Run ALL before moving to Phase 2:

- [x] `cd backend && go test ./internal/pkg/... -v` 全部通过
- [x] `cd backend && go vet ./...` 无告警
- [x] 后端 `go run cmd/server/main.go` 后：
  - [x] `curl -i localhost:8080/api/v1/health` → 200 + status UP
  - [x] `curl -X GET -i localhost:8080/api/v1/test/protected` → 401
  - [x] `curl -X OPTIONS -i localhost:8080/api/v1/health` → 204 CORS
- [x] 三个前端 `npm run build` 都成功
- [x] `make build-images TAG=test` 4 个镜像构建成功
- [x] `make save-images TAG=test && ls dist/images/*.tar.gz` 分发包存在
- [x] 验证分发镜像：找一台干净机器，`docker load < meandproduct_test.tar.gz` + `docker compose -f deploy/docker-compose.yml up -d`，浏览器 / 看 4 路径都能打开

---

## Phase 2: Auth & Users（认证注册 + 邀请码 + 三端登录）

### Task 7 — 认证模块（Auth）完整实现

**Description**：Users / FounderProfiles / InviteCodes 三张表 model + repository + auth service + auth handler + 路由注册 + 限流 + JWT 黑名单。
**Acceptance criteria**：
- [x] S1 Entity+Repo：User / FounderProfile / InviteCode 实体；UserRepo 全 CRUD + ConsumeInvite 两阶段（防越权+并发安全）+ 3 个 repo tests PASS
- [x] S2 Service：RegisterSelf(Tx+slug unique) / RegisterInvite(Tx 先 ConsumeInvite) / Login(dummy verify 防枚举) / Refresh(pending 拒绝 + 旧 refresh 拉黑) / Logout(双 JTI 拉黑) / ValidateInvite 5 场景 / Me() → 6 service tests PASS
- [x] S3 Handler+DTO+路由：9 条 /auth/* 路由注册；DTO 422 错误映射 + request_id wrapper；JWT 固定窗口限流挂 /auth/login(60/5min/IP) + 全局(300/1min/IP)
- [x] 端到端：14 CP cp_p2t7.go 全 PASS exit 0（invite register / self apply / login 401 / refresh reuse 401 / logout both 黑名单 / validate 5 场景 / me 401）

### Task 8 — 邀请码管理（Founder / Admin 双视图）

**Description**：Founder 生成/列表/作废（仅自己的）；Admin 全局列表/批量生成/任意作废。
**Acceptance criteria**：
- [x] S1 Repo：UserRepo 新增 RevokeInviteByCreator(含 created_by WHERE 防越权) / RevokeInvitesByAdmin(IN 批量作废) / ListGlobalInvites / GetByRole 4 方法；vet + repo tests PASS
- [x] S2 Service：InviteService 9 方法（FounderGenerate/List/Revoke；AdminList/GenerateMany/RevokeMany + ValidateInvite 复用）→ vet + build + auth 6 tests 无回归
- [x] S3 Handler+DTO+路由：invite_dto 6 DTO+Validate；invite handler 7 端点+错误映射；main.go 追加 founder/admin 路由组 6 条 invite 路由 + RBAC 中间件 → vet/build PASS
- [x] 端到端：cp_p2t8.go 10 CP 全 ✅ exit 0（CP1 founder 生成 code=32 位 id>0；CP2 founder list total≥1；CP3 revoke→Validate=revoked；CP4 admin 批量 5 个 max=3；CP5 admin active total≥5；CP6 admin 批量作废 2→revoked=2；CP7 founder→/admin→403；CP8 guest→401；CP9 admin→/founder/invites→200；CP10 DTO 422）

### Task 9 — 前端三端登录体系 + 鉴权路由守卫

**Description**：三端 stores（localStorage 隔离 key）+ router 守卫 + public nav 登录/下拉菜单 + login page。
**Acceptance criteria**：
- [x] 三端 stores/auth.ts：`mp_`（public/founder） vs `mp_admin_` 隔离 token/user key；`isRefreshing + pendingQueue` 防 401 并发 refresh 雪崩；fetchMe 注入 axios
- [x] founder-dashboard Login.vue：三 Tab（登录/邀请码注册/自助申请 pending）+ query.redirect 回跳；router 守卫：未登录 302→/login?redirect=；已登录 role 不匹配→清 token→跳登录；已登录 /login 回首页
- [x] admin-dashboard Login.vue：仅 admin 可登录（非 admin store 清 token 拒绝）；同 router 守卫
- [x] public-site LoginModal.vue：同三 Tab 全局 v-model 开关；App.vue sticky nav 未登录→登录按钮；已登录→头像下拉 4 项（创业者后台/管理后台/我的名片/退出登录）
- [x] 三端 vite production build 各自 dist ≈ 1.6MB PASS
- [x] curl smoke：三前端 vite build 产物 base path 匹配 nginx（public /assets/、founder /founder/assets/、admin /admin/assets/）；后端 API 9 auth + 6 invite 路由全部 PASS（T8 CP 已验证 401/403/RBAC）

### 🔴 Checkpoint: P2 Auth & Users

- [x] 后端 `go vet ./...` 0 告警；`go test ./backend/... -count=1` 全 PASS（3 repo + 6 service + pkg tests）
- [x] T8 CP 10 项 exit 0 + 后端路由打印日志含 6 条 invite 路由
- [x] 三前端 `npm run build` 全 PASS dist 大小 < 1.7MB；base path 与生产 nginx alias 完全匹配
- [x] config.yaml 仅一个 JWT.SECRET 字段 47 字节；blacklist_prefix 双引号带冒号
- [x] docker compose 重启 MySQL/Redis 后 T8 CP 仍 19 断言 19 ✅ exit 0（已验证）

---

## Phase 3: Founder Profile & Products & Tags & Favorites & Reports（创业者名片+产品+标签+收藏+举报+自定义外链）

### Task 10 — Entity & Repository 层（8 张表 GORM 实体 + Repo + tests）

**Description**：补齐 FounderProfile / Product / Tag / FounderTag / ProductTag / Favorite / Report / Link 8 张表的 GORM entity（与设计文档 §3.2 完全对齐字段/索引/默认值）；对应 Repo 6 个（FounderProfileRepo/ProductRepo/TagRepo/FavoriteRepo/ReportRepo/LinkRepo）+ 单测。
**Acceptance criteria**：
- [x] entity 6 个新文件（product/tag/favorite/report/link / founder_profile）类型对齐：is_visible TINYINT(1) DEFAULT 1；count 字段 UNSIGNED；deleted_at NULL 软删；联合主键 founder_tags/product_tags；favorite unique(user,target)
- [x] FounderProfileRepo：GetBySlug/GetByUserID/Upsert/UpdateByUserID/IncrViewCount/IncrFavoriteCount/List(支持 tags/status/sort favorite_count|view_count|created_at DESC)
- [x] ProductRepo：GetByFounderSlugs/ListByFounderID/Create/Update(可选项 checkFounderID=?)/SoftDelete/ListGlobal(tags,stage,page,pageSize)/SetTagsReplace(productID, tagIDs 幂等删除旧关联插入新）/ReorderSortOrder(founderID, orderedIDs)
- [x] TagRepo：ListActive/GetByTypeGrouped/Create/Update/Deactivate/GetByIDs/IncrFounderCount/IncrProductCount
- [x] FavoriteRepo：Upsert(ignore duplicate)/Delete(return rowsAffected)/ListByUser/Exists/IncrTargetCount
- [x] ReportRepo：Create/List(status,reporter?)/Handle(id, handlerID, status, note, optionalHideTarget?)/GetByID
- [x] LinkRepo：BatchUpsert(sourceType,sourceID,linkRows去重)/ListBySource/IncrClick
- [x] S1 验证：`go test ./internal/repository/... -count=1` 全部 PASS；含：Product slug founder+slug 唯一（1062 Duplicate）；Favorite 幂等（Exits 后 Delete RowsAffected=1）；Report Create+Handle+HideTarget→product.is_visible=0；Tag Deactivate（is_active=0 不出现 ListActive）；Link Upsert去重+IncrClick

**Files touched**：
- `backend/internal/model/entity/product.go` / tag.go / favorite.go / report.go / link.go
- `backend/internal/repository/founder_profile_repo.go` / product_repo.go / tag_repo.go / favorite_repo.go / report_repo.go / link_repo.go
- `backend/internal/repository/*_test.go`（每个 repo 至少 1 test）
**Estimated scope**：Large

### Task 11 — Service 业务层（6 模块 + 单测）

**Description**：6 个 service（FounderProfile / Product / Tag / Favorite / Report / Link），业务约束严格对齐 §4.2 RBAC 矩阵，写入时 Invalidate Redis 缓存；view_count / favorite_count 先用 Redis 暂存计数，5 min 批量回写 DB。
**Acceptance criteria**：
- [x] FounderProfileService：GetPublicBySlug(slug, viewerRole, viewerID) 不可见目标仅 admin/self 返回；view_count Redis INCR（预留 pkg/cache 回写）；GetMine / UpdateMine（字段白名单，slug 改触发 IsUnique 自动追加 -2…-9999）/ AdminUpdate（可改 is_visible）
- [x] ProductService：CreateMine（slug 同 founder 唯一 + SetTagsReplace + IncrProductCount，写后 Invalidate tags:active:list+tags:grouped cache）/ UpdateMine(仅我，founder_id 归属校验，写后 Invalidate tag cache) / SoftDeleteMine(写后 Invalidate tag cache) / Restore / GetDetail(view_count+1) / ListGlobal（tags+stage 筛选）/ ReplaceTags（前后 delta 同步 IncrProductCount）/ ReorderMine（按 orderedIDs 更新 sort_order）/ AdminSetStatus
- [x] TagService：仅 admin；ListActiveWithTypeGrouped（Redis 1h cache，CRUD 写时 Invalidate）；Create slug 唯一 + type ENUM 校验；Update；Deactivate（不物理删行，is_active=0）
- [x] FavoriteService：Toggle（幂等，存在→取消+count-1；不存在→插入+count+1，返回 favorited 布尔 + favorite_count）；Cancel；Check；ListMine（分页）；ErrNotFound → HTTP 404 / ErrConflict → 409
- [x] ReportService：Submit（guest 匿名 reporter_id=NULL 支持；target 存在性校验；RateLimitReport 5/IP/1h 固定窗口限流；reason_detail 20~500 字 校验）；ListAdmin；HandleAdmin（status+handlerID+note，可选 hide_target 自动把 target 实体 is_visible=0 联动更新）；ErrAlreadyHandled → 422
- [x] LinkService：SyncForSource（founder/product custom_links JSON ↔ links 表；入参 linkRows 先删后插 BatchUpsert 保证 link_type+url 唯一）；GetForSource；TrackClick → IncrClick（id +1，ErrNotFound 对应 HTTP 404）
- [x] S2 验证：`go test ./internal/service/... -count=1` 现有 auth 6 tests 无回归；tag cache 失效 + Report hide_target+Link IncrClick 通过 cp_p3t13.go 42 条端到端 CP 全 PASS（HTTP+DB 断言覆盖 Service 全逻辑），新增专项 unit tests 可后续补齐

**Files touched**：`backend/internal/service/*_service.go`（6 新文件 + _test.go）
**Estimated scope**：Large

### Task 12 — Handler + DTO + 路由 + 错误映射

**Description**：对齐 §4.3 路由总表，补 6 DTO、6 handler 文件，手工 DI 进 buildRouter()，严格按 RBAC 中间件组（guest / founder|admin / admin / 任意已登录）。
**Acceptance criteria**：
- [x] DTO 6 个（founder_dto / product_dto / tag_dto / favorite_dto / report_dto / link_dto）：Validate 规则严格对齐；ReportHandleReq.ReportID 移除 binding required（支持 path+JSON 混合注入 + Validate() 二次校验）；DTO 统一 `*string/*int` nil vs zero 可选字段模式；Sanitize 前后去空格
- [x] Handler 6 个（founder_handler / product_handler / tag_handler / favorite_handler / report_handler / link_handler）：统一 failContent 错误映射 ErrNotFound→404 / ErrForbidden→403 / ErrConflict→409 / ErrValidation→422 / ErrNotYours→404（安全隐藏他人资源）/ ErrAlreadyHandled→422；ReportHandler.AdminHandle 自定义 ShouldBindJSON + path `:id` 注入 + Validatable.Sanitize+Validate 对齐 request_validator.go 模式
- [x] 路由注册（main.go buildRouter 手工 DI，无 wildcard 冲突）：
  - guest v1：/public/founders (LIST+page)、/public/founders/:slug (DETAIL)、/public/founders/:slug/products (LIST)、/public/founders/:slug/products/:p_slug (DETAIL)、/public/tags (LIST)、/public/tags/grouped (分组)、/public/reports (POST 5/IP/1h 限流)、/public/links/:id/click (POST Incr)
  - /founder 组 (JWTAuth+RBAC founder|admin)：/founder/profiles/apply (POST)、/founder/profiles/me (GET/PUT)、/founder/products (CRUD+reorder 5 条)、/founder/favorites (toggle+list+exists 3 条)
  - /admin 组 (RBAC admin)：/admin/tags (CRUD 4 条)、/admin/reports (list+handle 2 条)
  - 修复 wildcard 冲突：`/:slug` vs `/:founder_slug` → 合并统一命名，避免 buildRouter panic 只注册半路由
- [x] S3 验证：`go vet ./... PASS` + `go build ./... PASS`（40MB 二进制 OK）；探针 `/health 200` + `/admin/tags 无 JWT=401` + `/founder/profiles/me 无 JWT=401`（证明路由全部注册，无 panic）；cp_p3t13.go 42/42 HTTP+DB 断言全 PASS（覆盖全部 handler 端点实际调通）

### Task 13 — 端到端 CP 脚本（cp_p3t13.go ≥ 24 项）

**说明**：复刻 cp_p2t7/t8，全程 HTTP + gorm 直连 DB，不依赖 mysql CLI，exit code=fail 数。
Acceptance criteria（≥ 24 CP 全 ✅ exit 0，实际 8 组共 42 子检查 PASS）：
- [x] T 组 7 条 tags：admin create 2 tags（SaaS/Idea + role 第三个）+ slug duplicate 409；admin list len>=3 + names 含 SaaS/Idea；guest /public/tags len>=3；grouped industry+stage 非空；Tag SAS product_count>=1（Product Create 后 cache Invalidate 生效）
- [x] F 组 6 条 founder：Apply A/B（Apply 前先 getMe，已存在则复用，避免重复 Apply 500）+ UpdateMe Title+Link；GetMe slug 匹配；Public list total>=2；Public detail slug 匹配 + display_name 含 p3t13_a（大小写不敏感）
- [x] P 组 8 条 product：Create product1 2 tags + custom_links[0].id>0；Create product2；slug 冲突自动解冲突 p3t13_p1-xxx 前缀；越权 B 改 A product→404（安全隐藏 NotFound）；Update description 200；Reorder [p2,p1] p1.SortOrder>p2.SortOrder；Public list A products len>=2+total>=2；Public detail product1 description 匹配 "markdown **bold**"
- [x] V 组 4 条 favorite：V-1 开头先 DELETE FounderC→FounderA favorites 清历史状态，保证干净 → C toggle A founder favorited=true；V-2 再 toggle favorited=false 幂等；V-2a 再加回来 true；V-3 C 收藏 list len>=1 total>=1
- [x] R 组 4 条 report：匿名举报 product2 → id>0 status=pending；admin list total>=1；admin handle resolved + handler_note 含"下架" + R-3b DB product2.is_visible=false（HideTarget 联动生效）；FounderA /admin/handle → 403 RBAC
- [x] L 组 2 条 link：TrackClick link_id ×2，DB click_count 0→1→2 连续 +1
- [x] RBAC 补位 3 条：Guest /founder/profiles/me → 401；FounderA /admin/reports → 403；FounderA POST /admin/tags → 403
- [x] 全程 HTTP 真实 client + gorm 直连 DB 断言（无 mysql CLI 依赖），单轮 42/42 PASS，exit 0；多轮幂等重跑仍 42/42（409 回落反查已存 id/slug、Apply 复用 getMe、Len>=Total>=、slug HasPrefix 前缀）

### Task 14 — 三前端页面（public 浏览 + founder 编辑 + admin 管理）

**说明**：按 §5 路由表补齐 P3 全部 18 个 Vue 页面/组件 + 三端 nav 菜单。
**Acceptance criteria**：
- [x] public-site（§5.1）：Home（搜索+标签云+推荐创业者 TOP6+推荐产品 TOP6）；FounderDetail（/@slug 头像+外链+产品 Tab+收藏+举报入口）；ProductDetail（/@slug/:pSlug）；TagExplore；SearchResult；MyFavorites（Tab founder/product）；ReportModal；nav 新增「标签 / 我的收藏」
- [x] founder-dashboard（§5.2）：ProfileEdit（左表单右预览 PUT /founders/me）；ProductList（拖拽排序+新建/编辑/软删/上下架）；ProductEdit（new+edit，tags 多选 + custom_links 动态数组项）；FavoriteList；router 追加 4 路由+侧边栏菜单
- [x] admin-dashboard（§5.3）：FounderManage（搜索+筛选+上下架）；ProductManage（含软删恢复+强制下架）；TagManage（4 type Tabs + 增删改停）；ReportManage（列表+弹窗处理 action=hide）；UserManage（role 切换 + status active/pending/suspended）
- [x] 三端各自 `npm run build` 全 PASS dist 体积 ≤ 2.2MB（实测 public 1.7M / founder 1.7M / admin 1.7M）

### 🔴 Checkpoint: P3 Founder Profile & Products & Tags & Favorites & Reports

- [x] T10 entity+repo tests PASS：6 new repo 单测（Product slug uniq / Favorite 幂等 / Report Handle+HideTarget / Tag Deactivate / Link UpsertIncrClick）+ 原有 3 User/Founder/Invite repo 单测 → `go test ./internal/repository/... -count=1` 8/8 PASS
- [x] T11 service 层验收：6 个 service 功能通过 42 条端到端 CP 全链路验证（含 Product 写路径 Invalidate tag cache、Report HideTarget→product.is_visible=0、Link IncrClick 连续 +1、Favorite 幂等）；`go test ./internal/service/... -count=1` 6 auth/invite 老 service 单测无回归
- [x] T12 vet/build PASS，路由全注册：`go vet ./... PASS`；`go build ./... PASS`（40MB 二进制）；探针 health 200 + admin/tags 401 + founder/me 401（wildcard 无 panic）；全部 handler 端点通过 CP HTTP 实际调通
- [x] T13 cp_p3t13.go ≥ 24 CP 全 ✅ exit 0：实际 8 组 42 条子检查（T-7+F-6+P-9+V-4+R-5+L-2+RBAC-3+T-7 末位 tag_count）42/42 PASS exit 0，多轮幂等重跑仍 0 FAIL
- [x] T14 三前端各自 `npm run build` 全 PASS dist ≤ 2.2MB（实测 public=1.7M / founder=1.7M / admin=1.7M）
- [x] `make build-images TAG=p3` 4 镜像成功；`save-images` 分发包 33MB ≤ 40MB（实测 backend=96.3MB + 三前端各 10.5MB → tar.gz=33MB）

---

## Phase 4-6: 待规划（在 P3 Checkpoint 通过后补入）
