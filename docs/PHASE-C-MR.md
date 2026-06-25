# 阶段 C：MR 提测 — 按 TC 核查 + 接口入库

## 流程

```
后端提测（MR）
    → 填写 MR 号 + 变更 API 列表
    → 从 Git 加载 test-docs 用例（阶段 B 已提交）
    → AI 按 TC 核对 MR（不再使用 BDD）
    → 缺口：补 TC（Git）/ 补 MR / 平台「设计+MR」入库 或 录制
    → 接口定义：补 US + TC 追溯 + 断言 → 执行
```

## MR 核查（只看 TC）

- 页面：**MR 核查**
- 填写 `version` + `requirement_id` → **从 Git 加载 TC**
- 粘贴 MR diff 接口列表 → **AI 按测试用例核对 MR**

缺口类型：

| type | 含义 |
|------|------|
| `mr_no_tc` | MR 改了但用例未覆盖 |
| `tc_no_mr` | 用例要求但 MR 未包含 |
| `platform_missing` | 平台尚无接口场景 |

## 接口进平台

| 方式 | 何时用 |
|------|--------|
| **从设计文档 + MR 生成接口入库** | 新接口、有后端设计文档 |
| **Chrome 录制 ingest** | 线上已有接口、要真实 request/response |

## 场景就绪（平台）

**US + `tc_ref` + ≥1 断言**（`bdd_ref` 仅可选追溯）

`tc_ref` 示例：`TC001 | REQ-V2.7.0-CHAT-01 | @line-trend`

## API

```http
GET /api/v1/docs/testcases?version=v2.7.0&requirement_id=chat-voice-input
POST /api/v1/ai/mr/verify-tc
POST /api/v1/ai/mr/ingest-apis
```
