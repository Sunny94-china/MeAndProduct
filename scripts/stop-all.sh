#!/usr/bin/env bash
# 停止仓库根目录 compose 拉起的 3 个业务容器（不动服务器上的 MySQL / Redis）。
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

log_info()  { printf '\033[1;34m[INFO]\033[0m %s\n' "$*"; }
log_ok()    { printf '\033[1;32m[ OK ]\033[0m %s\n' "$*"; }
log_warn()  { printf '\033[1;33m[WARN]\033[0m %s\n' "$*"; }

ENV_FLAG=()
if [[ -f .env ]]; then
  ENV_FLAG=(--env-file .env)
fi

log_info "停止 meandproduct 业务容器..."
docker compose ${ENV_FLAG[@]+"${ENV_FLAG[@]}"} down --remove-orphans || true

# 顺带清掉旧版 init/nginx 残留卷（若还在）
for v_name in \
  meandproduct-public-dist-p3 \
  meandproduct-admin-dist-p3 \
  meandproduct-founder-dist-p3 \
  meandproduct-nginx-logs \
  meandproduct-mysql-data \
  meandproduct-redis-data; do
  if docker volume inspect "${v_name}" >/dev/null 2>&1; then
    if [[ "${1:-}" == "--purge-old-volumes" ]]; then
      docker volume rm "${v_name}" >/dev/null && log_ok "已删除旧卷: ${v_name}" || true
    else
      log_warn "发现旧卷 ${v_name}（本项目已不再使用）。需要时加 --purge-old-volumes 删除。"
    fi
  fi
done

log_ok "已停止。服务器上的 MySQL / Redis 不会被本脚本改动。"
echo "  重新启动： docker compose up -d --build"
