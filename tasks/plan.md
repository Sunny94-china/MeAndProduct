# Implementation Plan: MeAndProduct 创业者名片平台

## Overview
按设计文档 §10 的 6 阶段实施计划，从工程脚手架（P1）到部署打磨（P6），采用「垂直切片 + 底层先行」策略。每阶段结束设检查点，确认上一阶段稳定再进入下一阶段。

本计划覆盖 **P1 工程脚手架阶段**，即：把项目从空仓库变成三端（Go 后端 + 三 Vue 前端）都能独立运行、基础中间件就位、MySQL 迁移框架可用的状态。

## Architecture Decisions
- **Monorepo 根目录**：`backend/` 放 Go，`frontend/{public-site,founder-dashboard,admin-dashboard}/` 放三个 Vue 项目，`deploy/` 放 docker compose 和 nginx 配置。
- **Go 三层架构**：handler/service/repository，通过构造函数依赖注入（不含 DI 框架，手工 wire 即可）。
- **迁移方案**：`golang-migrate/migrate`，原生 SQL 文件，`migrations/00001_*.up.sql` 风格。
- **缓存策略**：GORM repository 层内封装 pkg/cache 的 GetOrSet，不跨层传 redis client。
- **部署策略**：严格镜像分发包 —— 本地 `make build-images` → `make save-images` 导出 `.tar.gz` → 生产 `docker load` + `docker compose up`，**严禁生产服务器接触源码**。
- **三前端差异化端口**：本地开发 public(5173) / founder(5174) / admin(5175)，生产 nginx 按 path 路由。

## Phase & Task List

### Phase 1: Foundation 工程脚手架

- [ ] **Task 1**：根目录基础设施 —— Makefile / .gitignore / deploy/docker-compose.dev.yml / deploy/nginx.conf
- [ ] **Task 2**：Go 后端脚手架 —— go.mod + 目录骨架 + main.go + Viper 配置 + MySQL/Redis 连接 + Dockerfile
- [ ] **Task 3**：数据库迁移框架 + 初始化 SQL（00001_init，全部 10 张表）+ 自动执行迁移开关
- [ ] **Task 4**：pkg 公共基础包 —— response / jwtutil / password / cache / slug
- [ ] **Task 5**：middleware 中间件 —— Logger / CORS / RateLimit / JWT Auth / RBAC
- [ ] **Task 6**：三前端脚手架（public-site / founder-dashboard / admin-dashboard）+ 各自 Dockerfile

#### 🔴 Checkpoint: P1 Foundation
- [ ] `make build-images TAG=test` 全部镜像构建成功（4 个）
- [ ] `make dev` 启动 mysql + redis
- [ ] 后端独立 `make backend-run` 启动成功，`GET /api/v1/health` 返回 200
- [ ] 三前端独立 `npm run dev` 启动成功，浏览器访问显示 "Hello MeAndProduct"
- [ ] `make save-images TAG=test` → 解压 → `docker load` → `docker compose up` 全栈跑通（health + 前端首页）

### Phase 2: Auth & Users（认证注册 + 邀请码）

- [ ] **Task 7**：认证模块（Auth）完整实现 —— 用户/邀请码 model + repository + auth service + auth handler + 路由注册
  - [ ] **Slice 1** 数据模型 & Repository：User / FounderProfile / InviteCode 实体；UserRepo / InviteRepo（GORM）+ 单元测试
  - [ ] **Slice 2** Auth Service 业务层：`RegisterSelf` / `RegisterInvite` / `Login` / `Refresh` / `Logout` / `ValidateInvite` / `Me` + 邀请码自动使用计数
  - [ ] **Slice 3** Auth Handler & 路由：DTO 参数绑定 + 校验 + 401/403/422/409 错误映射 + `/auth/*` 全部 7 条路由注册 + curl 端到端验证 + 限流

- [ ] **Task 8**：邀请码管理（邀请码 CRUD）
  - founder：生成 / 自己列表 / 作废自己的
  - admin：全局列表 / 批量生成 / 任意作废

- [ ] **Task 9**：前端三端登录页 + 邀请码校验 + 鉴权路由守卫
  - founder-dashboard：登录/注册/邀请码校验 → 登录后跳 profile-edit
  - admin-dashboard：独立管理员登录 → 跳 dashboard
  - public-site：顶部登录入口（跳 founder-dashboard /login）+ 已登录显示收藏入口

#### 🔴 Checkpoint: P2 Auth & Users

- [x] 后端 `go vet ./...` 全部通过；单测 `go test ./... -count=1` 全部 PASS
- [x] curl CP + cp_p2t7.go 14 项认证注册 CP 全 PASS（invite register / self apply / login 401 / refresh reuse 401 / logout 双 401 / validate 5 场景 / me 401）
- [x] cp_p2t8.go 10 项邀请码管理 CP 全 PASS（Founder 生成/列表/作废 → Admin 批量 5 条→5/全局列表/批量作废 2 → 403/401/200 → DTO 422），19/19 子断言 ✅ EXIT=0
- [x] `POST /auth/login` → 正确密码 200 双 token；错误密码 401；pending/suspended 403（T8 CP 已覆盖 admin/founder）
- [x] `POST /auth/refresh` → 旧 refresh → 新 access + 旧 refresh 黑名单（T7 CP 验证）
- [x] JWT 限流生效（RateLimitGeneral 300/min/IP + RateLimitLogin 60/5min/IP 挂/auth/login）
- [x] docker compose 重启 MySQL/Redis 后 JWT + 业务 API 依然 PASS；三前端 build 全 PASS dist <1.7MB
- [x] 前端三端 stores/auth.ts + router 守卫 + login 页面 + public nav 登录/下拉菜单 build 成功（三端 vite build 全 PASS；base path 与 nginx alias 匹配）

### Phase 3: Founder Profile & Products & Tags & Favorites（创业者名片+产品+标签+收藏+举报+自定义外链）

本阶段覆盖设计文档 §3.2 表 2-10（founder_profiles / products / tags / founder_tags / product_tags / favorites / reports / links）+ §4.3.2 Founder、§4.3.3 Product、§4.3.4 Tag、§4.3.5 Favorite、§4.3.7 Report、§5 对应路由。采用垂直切片：先后端 5 个模块（T10→T11→T12→T13→T14），再三前端页面（T15：public 推荐/详情 + T16：founder/admin 编辑）。

---

#### T10 — 数据模型 & Repository 层（S1: Entity + Repo + tests）
> 说明：补齐 founder_profiles / products / tags / product_tags / founder_tags / favorites / reports / links 8 张表的 GORM 实体 + 对应 Repo CRUD + 单测（真实 MySQL + miniredis 预留给计数）
Acceptance Criteria：
- [x] entity：`model/entity/product.go` `Product` / `Tag` / `Favorite` / `Report` / `Link` / FounderProfile / FounderTag / ProductTag 共 8 个 GORM 实体，字段类型、索引、默认值与设计文档 §3.2.3-3.2.9 完全一致（is_visible 默认 1、count 字段 UNSIGNED、deleted_at NULL 软删、links联合uk + product_tags/favorites联合unique 等）
- [x] `FounderProfileRepo`：`GetBySlug(slug) / GetByUserID(uid) / Upsert(profile) / UpdateByUserID(uid, fields map) / IncrViewCount(id, delta) / IncrFavoriteCount(id, delta) / List(status, tags, page, pageSize)（支持 sort 字段：favorite_count DESC / view_count DESC / created_at DESC）
- [x] `ProductRepo`：`GetByFounderSlugs(founderSlug, productSlug) / ListByFounderID(fid, page, pageSize, sort_order) / Create(p) / Update(id, fields, checkFounder=?) / SoftDelete(id, checkFounder=?) / ListGlobal(tags, stage, page, pageSize) / SetTagsReplace(productID, tagIDs []uint64 幂等重写关联) / ReorderSortOrder(founderID, orderedIDs)`
- [x] `TagRepo`：`ListActive() / GetByTypeGrouped() / Create(t) / Update(id, fields) / Deactivate(id) / GetByIDs(ids) / IncrFounderCount(id,±1) / IncrProductCount(id,±1)`
- [x] `FavoriteRepo`：`Upsert(userID, targetType, targetID) / Delete(userID, targetType, targetID)（rowsAffected 回传）/ ListByUser(userID, filterType founder|product|all, page, pageSize) / Exists(userID, targetType, targetID) / IncrTargetCount(targetType, targetID, delta)`
- [x] `ReportRepo`：`Create(r) / List(status, reporterFilter, page, pageSize) / Handle(id, handlerID, status, note, HideTarget bool → 自动联动 target.is_visible=0) / GetByID(id)`
- [x] `LinkRepo`：`BatchUpsert(sourceType, sourceID, []Link 去重合并) / ListBySource(sourceType, sourceID) / IncrClick(id, delta) / GetByID(id)`
- [x] **S1 验证**：`go test ./internal/repository/... -count=1` 全部 8 PASS（覆盖：Product slug founder+slug 1062 Duplicate；Favorite 幂等 Upsert+Delete RowsAffected；Report Handle+HideTarget；Tag Deactivate；Link Upsert+IncrClick + 原有 User/Founder/Invite 3 repo 单测无回归）

#### T11 — Service 业务层（S2: service 6 模块 + 单测）
> 说明：service 是业务约束所在地，严格对齐设计文档 §4 RBAC 权限矩阵
Acceptance Criteria：
- [x] `FounderProfileService`：`GetPublicBySlug(slug, viewerRole, viewerID)` 隐藏 invisible=0（仅 admin/self 可见）+ view_count 计数（GREATEST 原子更新，防止负数）；`GetMine(userID)`；`UpdateMine(userID, fields)` 字段白名单（slug 改自动 IsUnique 冲突解 -2…-9999）；`Apply(userID, req)` 首次 Upsert；`AdminUpdate(adminID, founderUserID, fields)` 可改 is_visible
- [x] `ProductService`：`CreateMine(founderUserID, fields)` slug 同 founder 唯一 + SetTagsReplace + IncrProductCount（写后 Invalidate `tags:active:list` + `tags:grouped`）；`UpdateMine(founderUserID, productID, fields)` 强制归属 founder_id 校验；`SoftDeleteMine`（写后 Invalidate tag cache）+ 可选 Restore；`GetDetail(founderSlug, productSlug, viewer)` 详情；`ListByFounderID / ListGlobal`（tags+stage 筛选分页）；`ReplaceTags` 关联前后 delta 修正 tag.product_count；`ReorderMine(founderUserID, orderedIDs)` 按索引批量 UpdateColumn sort_order（ordered_ids[0] = 最小值）；`AdminSetStatus`
- [x] `TagService`：仅 admin；`ListActiveWithTypeGrouped()` 按 4 分组（Redis 1h cache `tags:active:list` + `tags:grouped`，CRUD 与 Product 写时 Invalidate）；`Create` slug 唯一 + type ENUM；`Update`；`Deactivate` 置 is_active=0 不物理删行
- [x] `FavoriteService`：`Toggle(userID, targetType, targetID)` 幂等，先查存在→取消+count-1，不存在→插入+count+1，返回 `{favorited,favorite_count,target_id,target_type}`；`ListMine(userID,page,pageSize)`；错误映射 ErrNotFound→404 / ErrConflict→409
- [x] `ReportService`：`Submit(ip, reporterID nullable, payload)` guest 匿名；target 存在校验；`rate:report:{ip}` Redis 固定窗口限流 5/1h；reason_detail 20~500；`ListAdmin`；`HandleAdmin(id,handlerID,status,note,HideTarget bool)` → HideTarget=true 时同步 target 实体（Founder/Product）`is_visible=0`；ErrAlreadyHandled→422
- [x] `LinkService`：`SyncForSource(sourceType,sourceID,linkRows)` 幂等，先删后插 BatchUpsert（保证 link_type+url 源唯一）；`GetForSource`；`TrackClick(id)` → `IncrClick(id,1)` + GetByID
- [x] **S2 验证**：`go test ./internal/service/... -count=1` 原有 6 auth/invite service tests 无回归；其余 6 个 Content Service 业务逻辑通过 cp_p3t13.go 42/42 端到端 HTTP+DB 断言覆盖（含 cache Invalidate、HideTarget 联动、IncrClick 连续 +1、Favorite 幂等），专项 service unit tests 可后续补

#### T12 — Handler + 路由 + DTO（S3: handler/DTO/路由注册 + vet/build）
> 说明：对齐 §4.3 路由总表，严格按 RBAC 路由组挂中间件
Acceptance Criteria：
- [x] DTO（`model/dto/`：founder_dto / product_dto / tag_dto / favorite_dto / report_dto / link_dto，每个 `Sanitize()` 去空格 + `Validate()` 映射 422 detail）
  - Founder Apply/Update：display_name 2~100，bio 0~500，URL http(s) 前缀，custom_links 每项 link_type≤20 url≤500
  - Product Create/Update：name 2~200 tagline 1~200 official_url 必填 status∈6 tag_ids 去重≤20
  - Tag Create：name 1~50 slug 1~50 type∈4（industry/stage/role/other）ENUM 校验
  - Favorite Toggle：target_type∈founder/product target_id>0
  - Report Submit/Handle：Submit reason_type ENUM reason_detail 20~500；Handle ReportID 移除 binding required（支持 path+JSON 混合注入 Validate 二次校验）；Status∈reviewed/rejected/resolved
- [x] Handler 6 个：`handler/founder_handler.go` / `product_handler.go` / `tag_handler.go` / `favorite_handler.go` / `report_handler.go` / `link_handler.go`；统一 failContent 错误映射 `{ErrNotFound:404, ErrForbidden:403, ErrConflict:409, ErrValidation:422, ErrNotYours:404（安全隐藏他人资源）, ErrAlreadyHandled:422}`；ReportHandler.AdminHandle 自定义 ShouldBindJSON + path `:id` 注入 + Validatable 接口 Sanitize/Validate 调用对齐 request_validator.go 模式
- [x] 路由注册（main.go buildRouter 手工 DI，修复 wildcard `/:a` vs `/:b` 同名前缀冲突 panic）：
  - **guest public**：`/public/founders` LIST+page；`/public/founders/:slug` DETAIL；`/public/founders/:slug/products` LIST；`/public/founders/:slug/products/:p_slug` DETAIL；`/public/tags` LIST；`/public/tags/grouped` 分组；`/public/reports` POST（RateLimitReport 5/IP/1h）；`/public/links/:id/click` POST IncrClick
  - **/founder 组 (JWTAuth+RBAC founder|admin)**：`/founder/profiles/apply` POST；`/founder/profiles/me` GET/PUT；`/founder/products` CRUD + POST `/reorder`（5 条）；`/founder/favorites` POST toggle + GET list（3 条）
  - **/admin 组 (RBAC admin)**：`/admin/tags` CRUD 4 条；`/admin/reports` GET list + POST `/:id/handle` 2 条
- [x] S3 验证：`go vet ./... PASS`；`go build ./... PASS`（40MB 二进制）；探针 health 200 + admin/tags 401 + founder/me 401（wildcard panic 修复，路由全注册）；上述 25+ 条新路由全部通过 cp_p3t13.go 真实 HTTP 客户端调通且返回断言正确

#### T13 — 端到端 CP（Go 临时脚本 cp_p3t13.go）
> 说明：复刻 cp_p2t7/t8 模式，HTTP + gorm 直连 DB，绕开 mysql cli，每轮失败返回 exit code = fail 数
Acceptance Criteria（每项 1 条 ✅ 共 ≥ 24 项，exit 0，**实际 42/42 PASS**）：
- [x] **T 组 7 条 tags**：admin create 2 tags（sas/idea）+ 409 回落反查；duplicate sas → 409；admin list len>=3；public tags len>=3；grouped industry+stage；T-7 sas product_count>=1（Product 写后 tag cache Invalidate 生效）
- [x] **F 组 6 条 founders**：Apply A/B（先 getMe 复用，避免重复 Apply 500）+ UpdateMe A 改 Title+Link；GetMe A slug 匹配；Public list total>=2；Public detail display_name 小写含 p3t13_a
- [x] **P 组 8 条 products**：Create product1（2 tags + custom_links[0].id>0）+ Create product2；slug 冲突自动解 p3t13_p1-xxx；越权 B 改 A product → 404（NotFound 安全隐藏）；Update description 匹配；Reorder [p2,p1] → p1.SortOrder > p2.SortOrder；Public list len>=2+total>=2；Public detail 含 markdown **bold**
- [x] **V 组 4 条 favorites**：V-1 前 DELETE FounderC→FounderA favorites 清历史（确保干净空状态）→ V-1 toggle favorited=true；V-2 再 toggle=false 幂等；V-2a 加回=true；V-3 C 收藏 list len>=1 total>=1
- [x] **R 组 4 条 reports**：匿名 submit product2 → id>0 status=pending；admin list total>=1；handle resolved + handler_note 含"下架" + R-3b DB product2.is_visible=false（HideTarget 联动）；FounderA handle → 403
- [x] **L 组 2 条 links**：TrackClick link_id ×2 → click_count 0→1→2 连续 +1
- [x] **RBAC 3 条**：Guest /founder/profiles/me → 401；FounderA /admin/reports → 403；FounderA /admin/tags create → 403
- [x] **幂等性**：ensureFounderUser 返回 user_id（用于清理 C→A favorite）；409 回落从 list 反查已有 tag id；Apply 前 getMe 复用已有 profile；CP 断言 >= 而非 ==（历史数据叠加）；slug 按前缀匹配 HasPrefix；display_name 大小写不敏感

#### T14 — Profile / Product / Tags / Favorites / Reports / Links ↔ 三前端页面
> 说明：按设计文档 §5 页面结构，public + founder-dashboard + admin-dashboard 补全部 P3 页面（API 调用 + 路由守卫 + stores）
Acceptance Criteria：
1. **public-site**（§5.1）
   - [ ] Home.vue：/ 首页：搜索框 + 标签云 + 推荐创业者 TOP6（favorite_count）+ 推荐产品 TOP6 + 卡片样式
   - [ ] FounderDetail.vue（/@slug）：左侧头像+头衔+bio+外链+产品Tab+举报入口+收藏按钮（登录后）；右侧产品卡片列表
   - [ ] ProductDetail.vue（/@slug/:pSlug）：产品信息+外链；点击「创始人」回跳 @slug
   - [ ] TagExplore.vue（/tags）：按 4 种 type 分组 tag 卡片，点标签进列表
   - [ ] SearchResult.vue（/search?q=）：type=founder/product tab 切换
   - [ ] MyFavorites.vue（/favorites 需登录）：Tab founder 收藏夹 + product 收藏夹 + 取消收藏
   - [ ] ReportModal/Report.vue：`/report?target_type=&target_id=` 表单 reason_type 单选 + reason_detail textarea
   - [ ] App.vue nav：新增「标签」「我的收藏（登录后）」入口
2. **founder-dashboard**（§5.2）
   - [ ] ProfileEdit.vue（/profile/edit）：左表单右预览，提交 PUT /founder/me；保存后清缓存
   - [ ] ProductList.vue（/products）：拖拽排序（vuedraggable）+ 新建/编辑/软删/上下架 + 上下架切换 visible
   - [ ] ProductEdit.vue（/products/new 与 /products/:id/edit）：表单型编辑器，tags 多选标签，custom_links 动态数组项
   - [ ] FavoriteList.vue（/favorites Tab）：复用 public MyFavorites 同组件 founder 版
   - [ ] router：上述 4 条路由追加，TabBar 侧边栏新增菜单
3. **admin-dashboard**（§5.3）
   - [ ] FounderManage.vue（/founders）：搜索筛选+上下架切换+status 查看详情跳 /@slug
   - [ ] ProductManage.vue（/products）：同上，含软删/恢复/强制下架
   - [ ] TagManage.vue（/tags）：4 组 type tabs + 新增/编辑/停用
   - [ ] ReportManage.vue（/reports）：列表+弹窗处理（action 选项：仅备注 / 屏蔽被举报对象 is_visible=0）
   - [ ] UserManage.vue（/users 提前实现基础表格：role 改 founder↔admin↔guest 改 status active↔pending↔suspended）
- [ ] 三端 vite build 各自 产物大小 ≤ 2.2MB（build 全 PASS）

#### 🔴 Checkpoint: P3 Founder Profile & Products & Tags & Favorites & Reports
- [x] T10 entity+repo tests PASS：8 repo 单测（6 new + 2 legacy）全绿
- [x] T11 service 验收：6 Content Service 逻辑通过 CP 42/42 覆盖；原有 6 auth/invite service tests 无回归
- [x] T12 vet/build PASS，路由全注册：wildcard 冲突修复，25+ 条新路由全部 HTTP 实际调通
- [x] T13 cp_p3t13.go 42 CP 全 ✅，exit 0：8 组全绿，多轮幂等重跑仍 exit 0
- [ ] T14 三前端各自 `npm run build` 全 PASS dist 大小 ≤ 2.2MB
- [ ] `make build-images TAG=p3` 4 镜像构建成功，save-images 分发包 ≤ 40MB

### Phase 4: Public Site 浏览搜索 & 缓存 & 统计（待规划，在 P3 通过后推进）

### Phase 5: Growth Records & Recommendation Polish（待规划）

### Phase 6: Deployment Scripts Polish & Smoke（待规划）

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| golang-migrate 首次配置不兼容 MySQL 8 认证 | High | 初始化连接参数固定 `multiStatements=true&parseTime=true&charset=utf8mb4`，MySQL 容器固定 auth plugin |
| Vue 3 + Vite + Element Plus 多版本互斥 | Medium | 固定稳定版：Vue@3.4 + Vite@5.4 + ElementPlus@2.8 + Pinia@2.2 + VueRouter@4.4 |
| JWT 与 RBAC 对「guest / founder / admin」三角色边界不清 | Medium | 每个接口在代码里显式声明权限注解，且在 API 总表与代码中双写对齐 |
| 三前端 nginx 路径打包后静态资源 404 | Medium | 每个前端在 vite.config.ts 显式指定 `base`（public=/ , founder=/founder/, admin=/admin/），构建产物路径一致 |

## Open Questions
- （无，当前阶段设计文档已覆盖全部细节）
