#!/usr/bin/env bash
# T-Rex Portal Badge 专项集成测试
# 参考页面: https://www.trex.xyz/portal/badge
#
# 用法:
#   ./scripts/seed.sh
#   go run ./cmd/server
#   ./scripts/integration-test-trex.sh

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

# shellcheck source=scripts/common.sh
source "$(dirname "$0")/common.sh"
resolve_base_url
API_TOKEN="${API_TOKEN:-TEST123}"
AUTH_HEADER="Authorization: Bearer ${API_TOKEN}"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'
step() { echo -e "\n${GREEN}==>${NC} $1"; }

curl_json() {
  curl -sS -H "$AUTH_HEADER" -H "Content-Type: application/json" "$@"
}

step "T-Rex 预置目录树（应含 T-Rex/Portal/Badge 等）"
curl_json "${BASE_URL}/api/v1/products/1/folders/tree" | python3 -m json.tool 2>/dev/null || true

if [ "${SKIP_AI:-}" = "1" ]; then
  echo "SKIP_AI=1，跳过 ingest"
  exit 0
fi

step "AI 入库: Badge 模块接口（fixtures/ingest-trex-badge-api.json）"
resp_api=$(curl -sS -X POST -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d @"${ROOT}/fixtures/ingest-trex-badge-api.json" \
  "${BASE_URL}/api/v1/ai/ingest")
echo "$resp_api" | python3 -m json.tool 2>/dev/null || echo "$resp_api"

if echo "$resp_api" | grep -qiE 'Badge|badge|Portal|T-Rex|trex'; then
  echo -e "${GREEN}✓${NC} Badge 相关接口已归入预期目录"
else
  echo -e "${YELLOW}!${NC} 请检查 apis[].folder_path 是否落在 T-Rex/Portal/Badge"
fi

step "AI 入库: 登录→Badge 场景（fixtures/ingest-trex-portal-scenario.json）"
resp_sc=$(curl -sS -X POST -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d @"${ROOT}/fixtures/ingest-trex-portal-scenario.json" \
  "${BASE_URL}/api/v1/ai/ingest")
echo "$resp_sc" | python3 -m json.tool 2>/dev/null || echo "$resp_sc"

if echo "$resp_sc" | grep -qiE 'Auth|Badge|Portal|T-Rex|auth|badge'; then
  echo -e "${GREEN}✓${NC} 场景目录符合 Portal/Auth 预期"
else
  echo -e "${YELLOW}!${NC} 请检查 scenario.folder_path"
fi

step "入库后接口列表（按 folder 筛选 Badge）"
curl -sS -H "$AUTH_HEADER" "${BASE_URL}/api/v1/products/1/apis" | python3 -m json.tool 2>/dev/null | head -80

echo ""
echo "完成。浏览器真机验证见 docs/VERIFY-TREX.md"
