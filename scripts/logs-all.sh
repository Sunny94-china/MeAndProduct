#!/usr/bin/env bash
# ==============================================================================
# MeAndProduct —— 日志查看（实时跟踪，Ctrl+C 退出）
# ==============================================================================
# 用法：
#   ./scripts/logs-all.sh               # 所有服务日志，最近 200 行 + follow
#   ./scripts/logs-all.sh backend       # 只看后端
#   ./scripts/logs-all.sh nginx mysql   # 多个 service
#   ./scripts/logs-all.sh --no-follow   # 只打印不 follow（末尾 500 行）
# ==============================================================================
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

COMPOSE_FILE="docker-compose.yml"
ENV_FILE=".env"
FOLLOW=1
TAIL=200
SERVICES=()

for a in "$@"; do
  case "${a}" in
    --no-follow) FOLLOW=0 ;;
    --tail=*)    TAIL="${a#--tail=}" ;;
    -*) : ;;
    *) SERVICES+=("${a}") ;;
  esac
done

if [[ ! -f "${ENV_FILE}" ]]; then
  :
  ENV_FLAG=()
else
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  ENV_FLAG=(--env-file "${ENV_FILE}")
fi

ARGS=()
(( FOLLOW == 1 )) && ARGS+=("--follow")
ARGS+=("--tail=${TAIL}" "--timestamps")

# shellcheck disable=SC2086
exec docker compose -f "${COMPOSE_FILE}" ${ENV_FLAG[@]+"${ENV_FLAG[@]}"} logs "${ARGS[@]}" "${SERVICES[@]+"${SERVICES[@]}"}"
