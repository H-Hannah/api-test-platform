#!/usr/bin/env bash
# 本地集成测试：验证 AI 入库与目录树分组（需先启动服务并配置 AI_API_KEY）
#
# 用法:
#   ./scripts/seed.sh          # 首次初始化库
#   go run ./cmd/server        # 另开终端启动
#   ./scripts/integration-test.sh
#
# 环境变量（可覆盖）:
#   BASE_URL   默认由 .env 的 ADDR 推导（:8081 -> http://localhost:8081）
#   API_TOKEN  默认 TEST123
#   SKIP_AI=1  跳过需调用大模型的 ingest 步骤

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

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

step() { echo -e "\n${GREEN}==>${NC} $1"; }
ok()   { echo -e "${GREEN}✓${NC} $1"; }
warn() { echo -e "${YELLOW}!${NC} $1"; }
fail() { echo -e "${RED}✗${NC} $1"; exit 1; }

curl_json() {
  local method="$1"
  local path="$2"
  local data="${3:-}"
  if [ "$method" = "GET" ]; then
    curl -sS -H "$AUTH_HEADER" "${BASE_URL}${path}"
    return
  fi
  if [ -n "$data" ]; then
    curl -sS -X "$method" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
      -d "$data" "${BASE_URL}${path}"
  else
    curl -sS -X "$method" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
      "${BASE_URL}${path}"
  fi
}

# --- 1. 健康检查 ---
step "1/7 健康检查 GET /health"
health=$(curl -sS "${BASE_URL}/health")
echo "$health" | grep -q '"status":"ok"' && ok "服务正常" || fail "服务未就绪: $health"

# --- 2. 产品列表 ---
step "2/7 产品列表 GET /api/v1/products"
products=$(curl_json GET "/api/v1/products")
echo "$products" | head -c 200
echo ""
echo "$products" | grep -q 'Trex' && ok "产品 Trex 存在" || fail "缺少 Trex"
echo "$products" | grep -q 'Edgen' && ok "产品 Edgen 存在" || fail "缺少 Edgen"
echo "$products" | grep -q 'example' && ok "产品 example 存在" || fail "缺少 example"

# --- 3. 环境列表 ---
step "3/7 环境列表 GET /api/v1/products/1/environments"
envs=$(curl_json GET "/api/v1/products/1/environments")
echo "$envs"
echo "$envs" | grep -q '"name":"dev"' && ok "Trex dev 环境已 seed" || warn "请先运行 ./scripts/seed.sh"

# --- 4. 入库前目录树 ---
step "4/7 入库前目录树 GET /api/v1/products/1/folders/tree"
tree_before=$(curl_json GET "/api/v1/products/1/folders/tree")
echo "$tree_before" | python3 -m json.tool 2>/dev/null || echo "$tree_before"
ok "已有预置目录（用户中心/广告 等），AI 应优先复用"

if [ "${SKIP_AI:-}" = "1" ]; then
  warn "SKIP_AI=1，跳过 AI ingest 步骤（5-6）"
  step "7/7 接口列表 GET /api/v1/products/1/apis"
  curl_json GET "/api/v1/products/1/apis" | python3 -m json.tool 2>/dev/null || true
  echo ""
  ok "基础接口测试完成（未测 AI）"
  exit 0
fi

if [ -z "${AI_API_KEY:-}" ]; then
  warn "未设置 AI_API_KEY，ingest 可能失败。可在 .env 中配置或 export AI_API_KEY=sk-xxx"
fi

# --- 5. AI 入库 - 广告接口（应归入 广告/Campaign 或类似路径）---
step "5/7 AI 入库（api 模式）POST /api/v1/ai/ingest"
echo "    fixture: fixtures/ingest-api-ad.json"
ingest_api=$(curl -sS -X POST \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d @"${ROOT}/fixtures/ingest-api-ad.json" \
  "${BASE_URL}/api/v1/ai/ingest")
echo "$ingest_api" | python3 -m json.tool 2>/dev/null || echo "$ingest_api"

if echo "$ingest_api" | grep -q '"error"'; then
  fail "API ingest 失败（检查 AI_API_KEY 与服务日志）"
fi
echo "$ingest_api" | grep -q '"apis"' && ok "接口入库成功" || fail "响应无 apis 字段"

# 检查是否创建了合理分组
if echo "$ingest_api" | grep -qiE '广告|Campaign|campaign'; then
  ok "folder_path 含广告/Campaign 相关分组（符合预期）"
else
  warn "folder_path 未明显匹配「广告」，请人工查看上方 JSON 中 folder_path 字段"
fi

# --- 6. AI 入库 - 登录场景（应归入 用户中心/认证）---
step "6/7 AI 入库（scenario 模式）POST /api/v1/ai/ingest"
echo "    fixture: fixtures/ingest-scenario-login.json"
ingest_sc=$(curl -sS -X POST \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d @"${ROOT}/fixtures/ingest-scenario-login.json" \
  "${BASE_URL}/api/v1/ai/ingest")
echo "$ingest_sc" | python3 -m json.tool 2>/dev/null || echo "$ingest_sc"

if echo "$ingest_sc" | grep -q '"error"'; then
  fail "Scenario ingest 失败"
fi
echo "$ingest_sc" | grep -q '"scenario"' && ok "场景入库成功" || fail "响应无 scenario 字段"

if echo "$ingest_sc" | grep -qiE '用户|认证|auth|login'; then
  ok "场景 folder_path 含用户/认证相关分组（符合预期）"
else
  warn "场景分组路径请人工确认"
fi

# --- 7. 入库后目录树 + 接口列表 ---
step "7/7 入库后目录树与接口列表"
tree_after=$(curl_json GET "/api/v1/products/1/folders/tree")
echo "--- folders/tree ---"
echo "$tree_after" | python3 -m json.tool 2>/dev/null || echo "$tree_after"

apis=$(curl_json GET "/api/v1/products/1/apis")
echo "--- apis ---"
echo "$apis" | python3 -m json.tool 2>/dev/null || echo "$apis"

scenarios=$(curl_json GET "/api/v1/products/1/scenarios")
echo "--- scenarios ---"
echo "$scenarios" | python3 -m json.tool 2>/dev/null || echo "$scenarios"

echo ""
ok "集成测试流程结束"
echo "提示: 再次执行 ingest 时，AI 应优先复用 tree_before 中已有路径"
