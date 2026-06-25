# Edgen 平台 — 全链路演示（接口场景 + 精准测试）

本演示模拟接口回归与精准测试流程，数据由 `go run ./cmd/seed-edgen-demo` 灌入，**不依赖 AI** 即可走通 UI。

BDD 与测试用例请在本地通过 Cursor Skill / `qa-doc-generator` 生成并提交 Git；平台从 GitLab 加载用例做精准测试。

## 0. 准备

```bash
cd api-test-platform
go run ./cmd/seed-edgen-demo    # 迁移 + Edgen 演示数据
go run ./cmd/server             # 或已有服务
# Web: ./scripts/build-web.sh 后访问 http://localhost:8081
```

顶栏：**归属项目 = Edgen**，**运行环境 = PROD**。

在 **环境管理** → 编辑 **PROD**，填写真实 `token`（否则执行接口会失败）。

---

## 1. 接口定义（补场景、执行）

路径：**侧栏 → 接口定义**

| 步骤 | 操作 |
|------|------|
| 1 | 筛选 **MR 标签** = `MR-EDGEN-DEMO` |
| 2 | 打开 **获取 Twitter 平台绑定信息** → 场景 **就绪** → 点执行（需 token） |
| 3 | 打开 **分页查询平台绑定列表** → 场景 **缺口** → **追溯** 页补 US、BDD、TC 引用、保存 |
| 4 | 再执行，确认断言通过 |

**用插件追加新接口（可选）：**

```bash
export API_TOKEN=你的平台令牌
curl -sS -X POST http://localhost:8081/api/v1/ai/ingest \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d @fixtures/edgen/ingest-mr-demo.json
```

---

## 2. 精准测试（MR 变更分析）

路径：**侧栏 → 精准测试**

1. 从 `qa-doc-generator` 选择分支 / 版本 / 需求，**加载** 测试用例
2. 填写 GitLab MR 链接或口述变更
3. 点 **AI 分析**，查看推荐用例与接口
4. MR 模式下可 **提交到 MR 评论**

---

## 附录：演示数据说明

| 资源 | 说明 |
|------|------|
| `fixtures/edgen/bdd-platform-bind.feature` | 种子 BDD 文本（仅演示追溯字段，不在平台生成） |
| `fixtures/edgen/prd-platform-bind.md` | PRD 样例 |
| `MR-EDGEN-DEMO` | 预置 MR 标签，关联 2 条接口 |
