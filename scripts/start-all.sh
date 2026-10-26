#!/usr/bin/env bash
# 有源码本机试跑。服务器无源码部署见 docs/deploy.md，不要用本脚本。
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

log_info()  { printf '\033[1;34m[INFO]\033[0m %s\n' "$*"; }
log_ok()    { printf '\033[1;32m[ OK ]\033[0m %s\n' "$*"; }
log_warn()  { printf '\033[1;33m[WARN]\033[0m %s\n' "$*"; }
log_error() { printf '\033[1;31m[FAIL]\033[0m %s\n' "$*"; }

if ! docker compose version >/dev/null 2>&1; then
  log_error "需要 Docker 和 docker compose v2"
  exit 1
fi

if [[ ! -f docker-compose.build.yml ]]; then
  log_error "无 docker-compose.build.yml。服务器请按 docs/deploy.md：docker load + docker compose up -d"
  exit 1
fi

if [[ ! -f .env ]]; then
  cp .env.example .env
  log_warn "已生成 .env，请先改配置再执行"
  exit 1
fi

set -a
# shellcheck disable=SC1091
source .env
set +a

TAG="${TAG:-latest}"
export TAG
export DOCKER_DEFAULT_PLATFORM="${DOCKER_PLATFORM:-${DOCKER_DEFAULT_PLATFORM:-linux/amd64}}"
export BACKEND_IMAGE="${BACKEND_IMAGE:-meandproduct-backend:${TAG}}"
export PUBLIC_IMAGE="${PUBLIC_IMAGE:-meandproduct-public:${TAG}}"
export ADMIN_IMAGE="${ADMIN_IMAGE:-meandproduct-admin:${TAG}}"

log_info "本机构建并启动 platform=${DOCKER_DEFAULT_PLATFORM}"
docker compose -f docker-compose.yml -f docker-compose.build.yml up -d --build --remove-orphans

log_ok "API :${API_PORT:-18080}  前台 :${PUBLIC_PORT:-18081}  管理后台 :${ADMIN_PORT:-18082}"
echo "打服务器发布包：make pack TAG=${TAG}   说明：docs/deploy.md"
