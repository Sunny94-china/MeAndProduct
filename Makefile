# MeAndProduct — Makefile
# Usage:
#   make dev                      # 本地开发启动 mysql + redis
#   make backend-run              # 本地启动 Go 后端
#   make frontend-public-run      # 前台 + 创业者工作台 dev
#   make frontend-admin-run       # 管理员后台 dev
#   make migrate-up               # 执行数据库迁移
#   make migrate-down STEP=1      # 回滚 STEP 步迁移
#   make build-images TAG=v0.1.0  # 有源码机器：构建 linux/amd64 镜像
#   make pack         TAG=v0.1.0  # 发布包 → dist/meandproduct-TAG.tar.gz
#   make up                        # 有源码本机试跑（含 build）
#   make help
#
# 运维文档：docs/deploy.md
# 服务器（无源码）：解压包 → 改 .env → docker load → docker compose up -d

TAG          ?= latest
PLATFORM     ?= linux/amd64
BACKEND_IMG   = meandproduct-backend:$(TAG)
PUBLIC_IMG    = meandproduct-public:$(TAG)
ADMIN_IMG     = meandproduct-admin:$(TAG)
STEP         ?= 1
export TAG BACKEND_IMAGE=$(BACKEND_IMG) PUBLIC_IMAGE=$(PUBLIC_IMG) ADMIN_IMAGE=$(ADMIN_IMG)
export DOCKER_DEFAULT_PLATFORM=$(PLATFORM)
export DOCKER_PLATFORM=$(PLATFORM)

.PHONY: help dev down
.PHONY: backend-run backend-build backend-test
.PHONY: frontend-public-run frontend-admin-run
.PHONY: migrate-up migrate-down
.PHONY: build-images save-images load-images push-images pack clean-dist
.PHONY: up down logs restart

help: ## 显示此帮助
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-26s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# ================================================================
# 本地开发环境（mysql + redis）
# ================================================================
dev: ## 本地启动 mysql + redis（开发环境）
	docker compose -f deploy/docker-compose.dev.yml up -d

down: ## 关闭本地开发 mysql + redis
	docker compose -f deploy/docker-compose.dev.yml down

# ================================================================
# Backend (Go)
# ================================================================
backend-run: ## 本地运行 Go 后端
	cd backend && go run cmd/server/main.go

backend-build: ## 本地编译 backend 二进制（Linux 多 arch，产出到 backend/dist/）
	@mkdir -p backend/dist
	@echo "==> build backend binaries (Linux amd64 / arm64, -s -w strip debug)"
	cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/server-linux-amd64 cmd/server/main.go
	cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/server-linux-arm64 cmd/server/main.go
	@echo "==> 把 config.yaml.example 和 migrations 一起放进 dist"
	cp backend/config.yaml.example backend/dist/config.yaml
	@rm -rf backend/dist/migrations && cp -R backend/migrations backend/dist/migrations

backend-test: ## 运行 Go 单元测试
	cd backend && go test -race -v ./...

backend-vet: ## go vet 静态检查
	cd backend && go vet ./...

# ================================================================
# Frontend (Vue 3)
# ================================================================
frontend-public-run: ## 运行前台（含创业者工作台）dev server (5173)
	cd frontend/public-site && npm install && npm run dev

frontend-admin-run: ## 运行管理员后台 dev server (5175)
	cd frontend/admin-dashboard && npm install && npm run dev

frontend-build-all: ## 本地构建前台和管理员后台（各输出 dist/，由 init 容器用）
	cd frontend/public-site && npm install --registry=https://registry.npmmirror.com && npm run build
	cd frontend/admin-dashboard && npm install --registry=https://registry.npmmirror.com && npm run build

# ================================================================
# DB Migrate
# ================================================================
migrate-up: ## 执行数据库迁移 up（全量）
	cd backend && go run cmd/server/main.go --migrate-up

migrate-down: ## 回滚数据库迁移（STEP=数字，默认1）
	cd backend && go run cmd/server/main.go --migrate-down $(STEP)

# ================================================================
# Docker 镜像（严格本地打包 → 镜像分发）
# ================================================================
build-images: ## 构建 3 个 linux/amd64 镜像，用法: make build-images TAG=v0.1.0
	@echo "==> Building TAG=$(TAG) platform=$(PLATFORM)"
	docker compose -f docker-compose.yml -f docker-compose.build.yml build
	@echo "==> Done: $(BACKEND_IMG), $(PUBLIC_IMG), $(ADMIN_IMG)"
	@docker images --format '{{.Repository}}:{{.Tag}}\t{{.Size}}' | grep "meandproduct.*:$(TAG)" || true

save-images: ## 导出 *-web.tar.gz（不含 compose）
	@mkdir -p dist/release
	@echo "==> Saving dist/release/meandproduct_$(TAG)-web.tar.gz"
	docker save $(BACKEND_IMG) $(PUBLIC_IMG) $(ADMIN_IMG) \
	  | gzip > dist/release/meandproduct-$(TAG)-web.tar.gz
	@echo "==> Done. Size: `du -h dist/release/meandproduct-$(TAG)-web.tar.gz | cut -f1`"

load-images: ## 加载 web 镜像 tar
	@echo "==> Loading dist/release/meandproduct-$(TAG)-web.tar.gz"
	gzip -dc dist/release/meandproduct-$(TAG)-web.tar.gz | docker load
	@echo "==> Done. 然后：docker compose up -d"

push-images: ## 可选：推镜像仓库（非规范默认路径）
	docker push $(BACKEND_IMG)
	docker push $(PUBLIC_IMG)
	docker push $(ADMIN_IMG)

pack: ## 打无源码发布包（amd64），用法: make pack TAG=v1.0.0
	bash scripts/pack.sh $(TAG)

clean-dist: ## 清理构建产物
	rm -rf dist/
	cd backend && rm -rf bin/ dist/
	-cd frontend/public-site && rm -rf dist/
	-cd frontend/admin-dashboard && rm -rf dist/
	-cd backend && rm -rf vendor/

# ================================================================
# 一键启动 / 停止 / 查看日志（依赖 + 服务 一体化）
# ================================================================
up: ## 有源码本机：build + 启动。服务器见 docs/deploy.md
	@bash ./scripts/start-all.sh

down: ## 一键停止服务（保留 mysql/redis 数据卷）
	@bash ./scripts/stop-all.sh

logs: ## 实时查看全部服务日志
	@bash ./scripts/logs-all.sh

restart: down up ## 重启整个 compose（down + up）
