# T-Rex Portal Badge 测试验证指南

参考页面：[Badges & Reputation - T-Rex Portal](https://www.trex.xyz/portal/badge)

页面模块与建议 API 分组对应关系：

| Portal 导航 | 建议 folder_path | 典型接口语义 |
|-------------|------------------|--------------|
| BADGE | `T-Rex/Portal/Badge` | 勋章列表、详情、即将上线 |
| PERSONA | `T-Rex/Portal/Persona` | 5D Persona、链上身份 |
| QUEST | `T-Rex/Portal/Quest` | 任务/成就 |
| Auth / Login | `T-Rex/Auth` | session、登录、用户信息 |

---

## 一、curl 离线验证（无需打开浏览器）

```bash
cd api-test-platform
./scripts/seed.sh
go run ./cmd/server          # 终端 1

# 终端 2 — 需配置 .env 中 AI_API_KEY
./scripts/integration-test-trex.sh
```

或手动：

```bash
export API_TOKEN=TEST123

curl -sS -X POST http://localhost:8080/api/v1/ai/ingest \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d @fixtures/ingest-trex-badge-api.json | python3 -m json.tool

curl -sS -H "Authorization: Bearer $API_TOKEN" \
  http://localhost:8080/api/v1/products/1/folders/tree | python3 -m json.tool
```

**预期**：`apis[].folder_path` 含 `T-Rex/Portal/Badge` 或复用 seed 中已有 `T-Rex/Portal/Badge` 节点。

---

## 二、Chrome 插件真机录制（推荐）

### 1. 准备

1. 启动平台：`go run ./cmd/server`
2. 重载扩展：`api-recorder/dist`
3. 数据处理页 → **配置平台**
   - 地址：`http://localhost:8080`
   - Token：与 `.env` 中 `API_TOKEN` 一致
   - 默认产品：**Trex**（已 seed T-Rex 目录）

### 2. 在 T-Rex 页面录制

1. 打开 <https://www.trex.xyz/portal/badge>
2. 点击扩展 → **开始录制**
3. 建议操作（产生 XHR/Fetch）：
   - 未登录：点击 **Login or Sign up for TREX**（会跳转 `/auth/portal-login`）
   - 完成钱包/账号登录后回到 Badge 页
   - 浏览 **New & Upcoming Badges**、切换底部 Tab（PERSONA / QUEST 等）
4. **保存** 录制 → 进入数据处理页

> 插件只录制 **XHR / Fetch**。若页面接口走 GraphQL 或 WebSocket，需在 Network 面板确认是否有可录制的 HTTP 请求。

### 3. AI 入库

1. 勾选 Badge 相关请求（URL 含 `badge`、`portal`、`auth` 等）
2. 点击 **AI 保存接口**
3. 产品选 **Trex**，业务说明填：

   ```text
   T-Rex Portal Badge 勋章模块 https://www.trex.xyz/portal/badge
   ```

4. 成功后查看抽屉中的 `folder_path`

**场景模式**（登录 → 拉用户信息 → 拉 Badge 列表）：

1. 按时间顺序勾选 ≥2 条请求
2. **AI 保存场景**，hint 同上

### 4. 平台侧核对

```bash
curl -sS -H "Authorization: Bearer TEST123" \
  http://localhost:8080/api/v1/products/1/folders/tree | python3 -m json.tool

curl -sS -H "Authorization: Bearer TEST123" \
  http://localhost:8080/api/v1/products/1/apis | python3 -m json.tool
```

---

## 三、环境与真实 API 对齐

seed 中已增加 `trex-dev` 环境：

- `base_url`: `https://api.trex.xyz`（**请按抓包实际 API 域名修改**）
- 变量：`token`、`walletAddress`

在 Web 平台（待开发）或后续执行时，将 `{{token}}` 替换为登录后真实 JWT。

---

## 四、验收清单

- [ ] `folders/tree` 下出现 `T-Rex/Portal/Badge`（或 AI 合理子路径）
- [ ] Badge 列表/详情接口带有 `status_code` + `json_path` + `duration_ms` 断言
- [ ] 场景步骤含 `extract_rules`（如从 login 响应取 `token`）
- [ ] 再次入库同类接口时，AI **复用**已有目录而非新建重复节点
- [ ] 插件「配置平台」连接测试通过

---

## 五、常见问题

**Q: 录制列表为空？**  
A: 确认已 attach debugger 且无其他 DevTools 占用；页面需有 `/api` 或业务 XHR 请求。

**Q: AI 分组到错误模块？**  
A: 在入库对话框填写更具体的 `hint`，或先在 seed 中维护准确保留目录。

**Q: 真实 API 域名不是 api.trex.xyz？**  
A: 以录制到的 `url` 为准，入库后可在平台编辑 `path`；并更新 `environments.trex-dev.base_url`。
