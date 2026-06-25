# 测试数据（阶段 D）

依据 **需求文档 + 后端技术方案 + 测试用例** 三份输入，AI 生成按业务域划分的 **测试集**（如 Tracker 测试集、Brief 测试集）及可执行 **数据集**。

## 文档来源（三仓）

| 类型 | 仓库 | 配置 |
|------|------|------|
| 需求文档 PRD | [edgen-product-docs](https://github.com/Everest-Ventures-Group/edgen-product-docs) | `GITHUB_PRD_REPO` + `GITHUB_TOKEN` 或 `GITHUB_PRD_REPO_ROOT` |
| 后端技术方案 | [osp-wiki/features](https://gitlab.com/Keccak256-evg/opensocial/osp-wiki/-/tree/ospdev/features) | `GITLAB_BE_PROJECT` + `GITLAB_TOKEN` |
| 测试用例 | qa-doc-generator `test-docs/` | `GITLAB_DOCS_PROJECT` + 页面所选分支 |

GitHub / osp-wiki 未命中时，会回退读取 qa-doc-generator `requirements/` 下同名目录。

## 流程

```
qa-doc-generator：选择分支 / 版本 / 需求 → 加载三类文档
  → AI 生成测试集 + data-spec.yaml
  → 导入平台 + 导入环境变量键
  → 接口定义：选择数据集后执行
```

## 测试集（collections）

| 字段 | 说明 |
|------|------|
| `collection_key` | 业务域标识，如 `tracker`、`brief` |
| `name` | 测试集名称，如「Tracker 测试集」 |
| `datasets` | 该域下的多条数据集 |

## 数据集字段

| 字段 | 说明 |
|------|------|
| `variables` | 合并进运行环境变量（覆盖同名键） |
| `body_override` | 覆盖接口默认 body |
| `api_bindings` | `METHOD /path`，用于接口详情筛选 |
| `obtain_type` | env / fixture / manual / setup |

## API

```http
GET  /api/v1/docs/requirement-package?ref=&version=&requirement_id=
POST /api/v1/ai/testdata/generate
POST /api/v1/products/{id}/testdata/import
GET  /api/v1/products/{id}/testdata/datasets?version=&requirement_id=&api_id=
POST /api/v1/apis/{id}/run  { "env_id": 1, "dataset_id": 3 }
POST /api/v1/environments/{id}/import-var-keys  { "keys": ["user_id_bound"] }
```

## Git 路径

`test-data/<version>/<requirement_id>/data-spec.yaml`
