#!/usr/bin/env bash
# 可选辅助：在无源码发布目录执行 load + up -d（等价于 docs/deploy.md 中的命令）。
# 规范主路径是手写：gzip -dc *-web.tar.gz | docker load && docker compose up -d
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${ROOT_DIR}"

log_info()  { printf '\033[1;34m[INFO]\033[0m %s\n' "$*"; }
log_ok()    { printf '\033[1;32m[ OK ]\033[0m %s\n' "$*"; }
log_warn()  { printf '\033[1;33m[WARN]\033[0m %s\n' "$*"; }
log_error() { printf '\033[1;31m[FAIL]\033[0m %s\n' "$*"; }

if ! docker compose version >/dev/null 2>&1; then
  log_error "需要 Docker 和 docker compose v2"
  exit 1
fi

if [[ -f docker-compose.build.yml ]] || [[ -d backend ]] || [[ -d frontend ]]; then
  log_error "检测到源码或 build 编排。服务器应使用 make pack 打出的无源码包。"
  exit 1
fi

if [[ ! -f .env ]]; then
  cp .env.example .env
  log_warn "已生成 .env，请编辑后再执行本脚本（或按 deploy.md 手动操作）"
  exit 1
fi

# shellcheck disable=SC1091
source .env

TAG="${TAG:-latest}"
API_PORT="${API_PORT:-18080}"
PUBLIC_PORT="${PUBLIC_PORT:-18081}"
ADMIN_PORT="${ADMIN_PORT:-18082}"

for p in "${API_PORT}" "${PUBLIC_PORT}" "${ADMIN_PORT}"; do
  if [[ "${p}" == "80" || "${p}" == "443" ]]; then
    log_error "端口 ${p} 会和已有 Nginx 冲突"
    exit 1
  fi
done

WEB_TAR=""
if [[ -f "meandproduct-${TAG}-web.tar.gz" ]]; then
  WEB_TAR="meandproduct-${TAG}-web.tar.gz"
elif ls meandproduct-*-web.tar.gz >/dev/null 2>&1; then
  WEB_TAR="$(ls meandproduct-*-web.tar.gz | head -n1)"
elif [[ -f "images/meandproduct-${TAG}-web.tar.gz" ]]; then
  WEB_TAR="images/meandproduct-${TAG}-web.tar.gz"
fi

if [[ -z "${WEB_TAR}" ]]; then
  log_error "找不到 meandproduct-*-web.tar.gz，请先 gzip -dc … | docker load"
  exit 1
fi

log_info "docker load ← ${WEB_TAR}"
gzip -dc "${WEB_TAR}" | docker load

log_info "docker compose up -d"
docker compose --env-file .env up -d --remove-orphans

HEALTH_URL="http://127.0.0.1:${API_PORT}/api/v1/health"
HTTP_CODE="000"
for _ in 1 2 3 4 5 6 7 8 9 10; do
  HTTP_CODE="$(curl -sS -o /dev/null -w '%{http_code}' "${HEALTH_URL}" 2>/dev/null || echo 000)"
  [[ "${HTTP_CODE}" == "200" ]] && break
  sleep 3
done

if [[ "${HTTP_CODE}" == "200" ]]; then
  log_ok "健康检查通过：${HEALTH_URL}"
else
  log_warn "健康检查 ${HTTP_CODE}。日志：docker compose logs -f backend"
fi

echo "下一步按 deploy.md 配置服务器已有 Nginx（端口 ${API_PORT}/${PUBLIC_PORT}/${ADMIN_PORT}）"
