# API Test Platform

轻量接口自动化平台（Go + SQLite），配合 Chrome 录制插件使用。

## 功能（Phase 1）

- 产品（**Trex**）+ **树形目录**自动/手动管理
- **AI 入库** `POST /api/v1/ai/ingest`：录制数据 → 接口定义 + 断言 + **自动分组到 folder_path 树**
- 单接口 / 线性场景执行（status_code、json_path、duration_ms 断言）
- 执行报告

## 快速启动

```bash
cd api-test-platform
cp .env.example .env
# 编辑 .env 设置 API_TOKEN、AI_API_KEY

go mod tidy

# 1) 初始化库：环境 + 示例目录树（AI 分组复用测试用）
./scripts/seed.sh

# 2) 启动服务（另开终端）
go run ./cmd/server

# 3) 集成测试（需 AI_API_KEY；无 Key 时可 SKIP_AI=1 只测基础接口）
./scripts/integration-test.sh
```

默认：`http://localhost:8081`，数据库 `./data/platform.db`

### 阶段 D · 测试数据

后端技术方案 + 用例/BDD → `test-data/` YAML + 平台数据集（执行注入）。详见 [docs/TESTDATA-PHASE-D.md](docs/TESTDATA-PHASE-D.md)。

> BDD 与测试用例生成请在本地通过 Cursor Skill / qa-doc-generator 完成，平台仅加载 Git 用例做精准测试。

### 阶段 C · MR 提测

按 Git **测试用例** 核对 MR；新接口用「设计+MR」AI 入库，已有接口用录制。详见 [docs/PHASE-C-MR.md](docs/PHASE-C-MR.md)。

### 精准测试

GitHub PR / diff → 推荐用例与接口执行子集。详见 [docs/IMPACT-TESTING.md](docs/IMPACT-TESTING.md)。

### Edgen 全链路演示（接口场景 + 精准测试）

```bash
go run ./cmd/seed-edgen-demo   # 灌入 US-EDGEN-042 平台绑定示例
go run ./cmd/server
```

Web 顶栏选 **Edgen / PROD**，按 [docs/EDGEN-WALKTHROUGH.md](docs/EDGEN-WALKTHROUGH.md) 走 **接口定义 → 精准测试**。示例 MR 号：`MR-EDGEN-DEMO`。

## Web 管理界面（Vue3 + Element Plus）

MeterSphere 风格三页：**接口定义**（目录树 + 列表 + 详情抽屉）、**测试场景**、**测试报告**。侧栏品牌图标与 Chrome 插件共用 `api-recorder/src/assets/icon*.png`（构建时自动同步）。

```bash
# 构建前端（产物由 Go 服务托管）
./scripts/build-web.sh

# 开发模式（热更新，代理到 8081）
cd web && npm install && npm run dev
```

浏览器打开 `http://localhost:8081`，首次进入**登录页**填入 `TEST123`（与 `.env` 中 `API_TOKEN` 一致），验证通过后进入管理界面。

### Seed 内容

| 产品 (id) | 环境 | 预置目录示例 |
|-----------|------|----------------|
| Trex (1) | BETA, PRE, PROD | `T-Rex/Portal/Badge` 等 |
| Edgen (2) | BETA, PRE, PROD | `Edgen/API`、`Edgen/Auth` |
| example (3) | BETA, PRE, PROD | `example/Demo`、`example/API` |

预置目录用于验证 AI **优先复用**已有路径，而非全部新建。

### 环境与多微服务 URL

- **BETA / PRE / PROD**：部署档，执行时选一次即可。
- 每个环境的 `variables` JSON 含多套服务地址，例如 **Trex BETA**：
  - `base_url_trex` → `https://api.trex.beta.dipbit.xyz`
  - `base_url_quest` → `https://api.quests.beta.dipbit.xyz`
  - `base_url_anchor` → `https://anchor.trex.beta.dipbit.xyz`
- **Edgen** 另有 `base_url_openreplay`（如 `https://openreplay.ospprotocol.xyz`）。
- 入库时按录制 `service` 生成 `full_url_template`；跨服务场景执行时只选一次 BETA/PRE/PROD。
- 域名与 **token**：在 Web **环境管理**（侧栏 / 顶栏「管理」）按项目维护，无需改 SQLite。
- `token` 写入 `variables.token`，对应接口头里的 `Bearer {{token}}`；未配置时执行会提示缺变量。
- 域名变更：环境管理页编辑，或 migration `007_real_env_urls.sql`。

## 鉴权

请求头：`Authorization: Bearer <API_TOKEN>`

## 核心 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/ai/ingest` | 插件主入口，AI 分类入库 |
| GET | `/api/v1/products/{id}/folders/tree` | 目录树 |
| GET | `/api/v1/products/{id}/apis?folder_id=` | 接口列表 |
| GET/POST | `/api/v1/products/{id}/environments` | 环境列表 / 新建 |
| GET/PUT/DELETE | `/api/v1/environments/{id}` | 环境详情 / 更新 / 删除 |
| POST | `/api/v1/apis/{id}/run` | 单接口执行 |
| POST | `/api/v1/scenarios/{id}/run` | 场景执行 |

### AI 入库 curl 示例

**单条接口（应归入 `广告/Campaign` 类目录）：**

```bash
export API_TOKEN=TEST123
curl -sS -X POST http://localhost:8080/api/v1/ai/ingest \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d @fixtures/ingest-api-ad.json | jq .
```

**登录线性场景（应归入 `用户中心/认证` 类目录）：**

```bash
curl -sS -X POST http://localhost:8080/api/v1/ai/ingest \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d @fixtures/ingest-scenario-login.json | jq .
```

**查看分组结果：**

```bash
curl -sS -H "Authorization: Bearer $API_TOKEN" \
  http://localhost:8080/api/v1/products/1/folders/tree | jq .

curl -sS -H "Authorization: Bearer $API_TOKEN" \
  http://localhost:8080/api/v1/products/1/apis | jq .
```

响应含 `folder_path`、`folder_id`，以及 `folders_created`（新建节点时）。

## AI 自动分组说明

入库时服务端会将**已有目录树**和**路径列表**传给大模型；模型为每个接口返回 `folder_path: ["模块","子模块"]`，服务端通过 `EnsureFolderPath` 自动创建缺失节点并关联 `folder_id`。

## T-Rex Portal Badge 验证

参考真实产品页 [trex.xyz/portal/badge](https://www.trex.xyz/portal/badge)：

```bash
./scripts/seed.sh
go run ./cmd/server
./scripts/integration-test-trex.sh   # 或阅读 docs/VERIFY-TREX.md 用插件真机录制
```

Fixtures：`fixtures/ingest-trex-badge-api.json`、`fixtures/ingest-trex-portal-scenario.json`
