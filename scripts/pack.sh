#!/usr/bin/env bash
# 本机打包（有源码）→ 无源码发布包，对齐 docs/superpowers/ai-one-click-deploy-spec.md
#
#   make pack TAG=v1.0.0
#   bash scripts/pack.sh v1.0.0
#
# 默认平台 linux/amd64（Apple Silicon 也打 amd64，供常见服务器使用）
# 产出：dist/meandproduct-<TAG>.tar.gz
# 包内：*-web.tar.gz + docker-compose.yml + .env.example + deploy.md
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

TAG="${1:-${TAG:-latest}}"
PLATFORM="${DOCKER_PLATFORM:-linux/amd64}"
export TAG
export DOCKER_DEFAULT_PLATFORM="${PLATFORM}"
BACKEND_IMAGE="${BACKEND_IMAGE:-meandproduct-backend:${TAG}}"
PUBLIC_IMAGE="${PUBLIC_IMAGE:-meandproduct-public:${TAG}}"
ADMIN_IMAGE="${ADMIN_IMAGE:-meandproduct-admin:${TAG}}"
export BACKEND_IMAGE PUBLIC_IMAGE ADMIN_IMAGE

STAGE="dist/pack/meandproduct"
ARCHIVE="dist/meandproduct-${TAG}.tar.gz"
WEB_TAR="meandproduct-${TAG}-web.tar.gz"
WEB_TAR_PATH="dist/release/${WEB_TAR}"

log_info()  { printf '\033[1;34m[INFO]\033[0m %s\n' "$*"; }
log_ok()    { printf '\033[1;32m[ OK ]\033[0m %s\n' "$*"; }
log_error() { printf '\033[1;31m[FAIL]\033[0m %s\n' "$*"; }

if ! docker compose version >/dev/null 2>&1; then
  log_error "需要 Docker 和 docker compose v2"
  exit 1
fi

if [[ "${SKIP_BUILD:-0}" != "1" ]]; then
  log_info "构建镜像 TAG=${TAG} platform=${PLATFORM}"
  docker compose -f docker-compose.yml -f docker-compose.build.yml build
else
  log_info "SKIP_BUILD=1，跳过构建"
fi

for img in "${BACKEND_IMAGE}" "${PUBLIC_IMAGE}" "${ADMIN_IMAGE}"; do
  if ! docker image inspect "${img}" >/dev/null 2>&1; then
    log_error "镜像不存在：${img}"
    exit 1
  fi
done

log_info "导出镜像 → ${WEB_TAR_PATH}"
mkdir -p dist/release
docker save "${BACKEND_IMAGE}" "${PUBLIC_IMAGE}" "${ADMIN_IMAGE}" \
  | gzip > "${WEB_TAR_PATH}"

log_info "组装无源码发布目录"
rm -rf "${STAGE}"
mkdir -p "${STAGE}"
cp "${WEB_TAR_PATH}" "${STAGE}/"
cp docker-compose.yml "${STAGE}/"
cp docs/deploy.md "${STAGE}/deploy.md"

cat > "${STAGE}/.env.example" <<EOF
# 由 make pack TAG=${TAG} 生成（platform=${PLATFORM}）
# 服务器：cp .env.example .env → 改配置 → 见 deploy.md

TAG=${TAG}
BACKEND_IMAGE=${BACKEND_IMAGE}
PUBLIC_IMAGE=${PUBLIC_IMAGE}
ADMIN_IMAGE=${ADMIN_IMAGE}

# 只绑 127.0.0.1，给已有 Nginx 反代；不要用 80/443
API_PORT=18080
PUBLIC_PORT=18081
ADMIN_PORT=18082

MYSQL_HOST=host.docker.internal
MYSQL_PORT=3306
MYSQL_USER=meandproduct
MYSQL_PASSWORD=change_me_app_pwd_456
MYSQL_DB=meandproduct

REDIS_ADDR=host.docker.internal:6379
REDIS_PASSWORD=
REDIS_DB=11

JWT_SECRET=please_change_me_to_a_long_random_string_at_least_32_chars
JWT_BLACKLIST_PREFIX=mp:jwt:blacklist:

CORS_ALLOW_ORIGINS=*

ADMIN_USERNAME=admin
ADMIN_EMAIL=admin@local
ADMIN_PASSWORD=Admin@123
EOF

printf '%s\n' "${TAG}" > "${STAGE}/VERSION"
printf '%s\n' "${PLATFORM}" > "${STAGE}/PLATFORM"

tar -C dist/pack -czf "${ARCHIVE}" meandproduct
log_ok "发布包：${ARCHIVE}（$(du -h "${ARCHIVE}" | awk '{print $1}')）"
echo
echo "本机打包完成（platform=${PLATFORM}）。人工上传到服务器后："
echo "  tar -xzf meandproduct-${TAG}.tar.gz && cd meandproduct"
echo "  cp .env.example .env    # 改 MySQL/Redis/JWT/管理员密码"
echo "  gzip -dc ${WEB_TAR} | docker load"
echo "  docker compose up -d"
echo "  # Nginx 片段见包内 deploy.md"
echo
echo "服务器若是 arm64：DOCKER_PLATFORM=linux/arm64 make pack TAG=${TAG}"
