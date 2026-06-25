# 精准测试（Impact Testing）

根据 **代码变更**（GitLab MR 或口述变更）结合 **qa-doc-generator 测试用例**，由 AI 推荐应执行的用例与平台接口，支持一键跑推荐子集，并将报告提交到 MR 评论区。

## 页面流程

```
选择 test-docs 分支 → 版本 → 需求 → 加载用例
填写 GitLab MR 链接（或口述变更说明）
→ AI 精准分析
→ 查看推荐用例 / 接口 / 变更文件
→ （MR 模式）预览并提交到 MR 评论
→ （可选）选择运行环境后执行推荐接口
```

### 变更来源

| 模式 | 说明 |
|------|------|
| **GitLab MR**（推荐） | 粘贴 `https://gitlab.com/group/project/-/merge_requests/123`，平台拉取 diff 分析 |
| **口述变更** | 无 MR 时描述配置、开关、逻辑变更，生成测试点解读 |

### 加载测试用例

在页面选择 `qa-doc-generator` 分支、版本与需求 ID，调用：

```http
GET /api/v1/docs/testcases/branches
GET /api/v1/docs/testcases/catalog?ref=<branch>
GET /api/v1/docs/testcases?ref=<branch>&version=v2.7.2&requirement_id=brief
```

服务端 `.env` 需配置 `GITLAB_DOCS_PROJECT` 与 `GITLAB_TOKEN`（读私有仓）。

## GitLab Token 权限

发表 MR 评论需要 **写权限**，创建 Token 时除 `read_api`、`read_repository` 外，需勾选 **`api`**（或等效的 write/api 范围）：

1. GitLab 网页 **头像 → Preferences → Access tokens**
2. 勾选 `read_api`、`read_repository`、`api`
3. 写入 `.env`：`GITLAB_TOKEN=glpat-xxx`
4. 重启服务

自建 GitLab 无需单独配置域名：MR 完整链接里已包含实例地址。

## 推荐逻辑

- 从 MR diff 提取变更文件与路由
- 用例：与变更路径、模块关键词匹配 → 打分，**仅展示得分 > 1** 的强相关用例
- 接口：path 命中、目录关键词、场景就绪加分
- `use_ai: true`（页面默认开启）时附加 AI 变更解读 / 测试点解读

## API

```http
POST /api/v1/impact/analyze
POST /api/v1/impact/preview-mr-comment
POST /api/v1/impact/post-mr-comment
POST /api/v1/impact/run-plan
```

### analyze 示例（MR 模式）

```json
{
  "product_id": 2,
  "version": "v2.7.2",
  "requirement_id": "brief",
  "gitlab_mr_url": "https://gitlab.com/org/backend/-/merge_requests/42",
  "cases_json": "...",
  "use_ai": true
}
```

### post-mr-comment 示例

将当前分析结果格式化为 Markdown 并 POST 到 GitLab MR Notes API：

```json
{
  "gitlab_mr_url": "https://gitlab.com/org/backend/-/merge_requests/42",
  "version": "v2.7.2",
  "requirement_id": "brief",
  "tc_docs_branch": "beta_20260618_v272",
  "result": { "...": "analyze 返回的完整结果对象" }
}
```

响应：

```json
{
  "note_id": 12345,
  "note_url": "https://gitlab.com/org/backend/-/merge_requests/42#note_12345",
  "markdown": "## 精准测试报告\n\n..."
}
```

`preview-mr-comment` 请求体相同，仅生成 `markdown` 不发表，供页面预览。

MR 评论内容包含：用例来源、摘要、AI 解读、推荐用例表、推荐接口表、变更文件列表、缺口提示。

### run-plan 示例

```json
{
  "product_id": 2,
  "env_id": 3,
  "api_ids": [12, 15],
  "scenario_ids": [2]
}
```

## 与 MR 核查的关系

| MR 核查 | 精准测试 |
|---------|----------|
| MR 接口列表 ↔ TC 文档覆盖 | 代码变更 ↔ TC + 平台接口 |
| AI 核对缺口 | AI 解读 + 规则打分 |
| 偏提测门禁 | 偏执行范围收敛 + MR 评论留痕 |

两者互补：MR 核查看「测全了吗」，精准测试看「这次改动了什么、最少跑哪些」。
