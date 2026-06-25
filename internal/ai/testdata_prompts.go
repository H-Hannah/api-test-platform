package ai

import "fmt"

func buildTestDataGeneratePrompt(req TestDataGenerateRequest) string {
	prdBlock := req.PrdText
	if prdBlock == "" {
		prdBlock = "（未提供需求文档）"
	}
	casesBlock := req.CasesJSON
	if casesBlock == "" {
		casesBlock = "（未提供测试用例 JSON，请根据需求与后端设计推断数据场景）"
	}
	beBlock := req.BeTechText
	if beBlock == "" {
		beBlock = "（未提供后端技术方案）"
	}
	apiBlock := "（未提供接口清单，从后端设计中提取）"
	if len(req.APIHints) > 0 {
		apiBlock = fmt.Sprintf("%v", req.APIHints)
	}
	return fmt.Sprintf(`你是资深测试开发工程师。综合**需求文档、后端研发技术方案、测试用例**三份输入，设计**可落地执行的测试数据规格**。

## 需求包
- version: %s
- requirement_id: %s
- requirement_name: %s

## 用户补充
%s

## 需求文档（PRD / 产品说明）
%s

## 后端研发技术方案（API、错误码、领域模型、数据约束）
%s

## 测试用例 JSON
%s

## 已知接口清单（可选）
%s

## 任务
1. **先识别业务域/子系统**（如 Tracker、Brief、Portfolio、Chat 等），为每个域创建独立 **测试集 collections**。
2. 每个测试集包含多条 **datasets**（正常流、边界、异常、权限、空态、setup 前置等）。
3. 测试集命名示例：collection_key=tracker, name="Tracker 测试集"；collection_key=brief, name="Brief 测试集"。
4. 单需求若只涉及一个域，也须输出 1 个 collection；跨域需求须拆分多个 collection。
5. 每条 dataset 须映射到至少 1 条 TC（tc_refs）和/或接口（api_bindings 格式 "METHOD /path"）。
6. variables 使用 {{var_name}} 引用环境变量；敏感项只列 env_keys 键名，值用 {{token}} 等占位。
7. obtain_type: env|fixture|manual|setup；owner: qa|backend|fe。
8. dataset_key 建议格式：{collection_key}-001（如 tracker-001、brief-002）。

## 输出要求
只输出一个 JSON 对象，不要 markdown 代码块。优先使用 collections 结构；若无 collections 则平铺 datasets。

## JSON Schema
{
  "version": "%s",
  "requirement_id": "%s",
  "requirement_name": "%s",
  "env_keys": ["token", "user_id_bound"],
  "collections": [
    {
      "collection_key": "tracker",
      "name": "Tracker 测试集",
      "description": "Tracker 创建、汇报、Push 相关数据",
      "datasets": [
        {
          "dataset_key": "tracker-001",
          "collection_key": "tracker",
          "collection_name": "Tracker 测试集",
          "name": "已创建 Tracker-正常查询",
          "description": "...",
          "tc_refs": ["TC003"],
          "api_bindings": ["GET /v2/trackers/{{id}}"],
          "variables": {"tracker_id": "{{tracker_id}}"},
          "headers_override": [],
          "body_override": "",
          "obtain_type": "env",
          "obtain_note": "BETA 配置 tracker_id",
          "owner": "qa",
          "tags": ["happy-path"]
        }
      ]
    }
  ],
  "datasets": [],
  "coverage_notes": "各测试集覆盖说明与缺口",
  "git_output_hint": "test-data/%s/%s/data-spec.yaml"
}`,
		req.Version, req.RequirementID, req.RequirementName, req.Hint,
		truncateText(prdBlock, 10000),
		truncateText(beBlock, 12000),
		truncateText(casesBlock, 12000),
		apiBlock,
		req.Version, req.RequirementID, req.RequirementName,
		req.Version, req.RequirementID)
}
