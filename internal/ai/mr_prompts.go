package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

func buildMRVerifyTCPrompt(req MRVerifyTCRequest, changed, platform []map[string]string, tcSummary string) string {
	return fmt.Sprintf(`你是测试专家。根据 **Git 测试用例（TC）** 核对一次 Merge Request 的接口变更是否测全，不要用 BDD。

## 需求
- 版本: %s
- requirement_id: %s
- 需求名: %s
- MR: %s

## MR 接口变更列表
%s

## 测试平台已关联本 MR 的接口
%s

## 后端技术设计（节选，可选）
%s

## Git 测试用例摘要（共 %d 条 JSON）
%s

## 任务
1. 判断 MR 中每个接口变更是否至少有 1 条相关 TC 覆盖（步骤或标题中出现 path/接口语义）。
2. 找出 **MR 有但 TC 未覆盖** 的接口（type=mr_no_tc）。
3. 找出 **TC 涉及但 MR 未包含** 的接口能力（type=tc_no_mr，可能是漏提测）。
4. 对比平台已入库接口，建议 platform_missing 需 AI 入库或录制。
5. verdict: pass（MR⊆TC 且无明显漏测）| gap | risk

## 输出 JSON（不要 markdown）
{
  "verdict": "pass|gap|risk",
  "summary": "一句话",
  "covered": [{"tc_id":"TC001","tc_title":"","api":"GET /v2/...","note":""}],
  "gaps": [{"type":"mr_no_tc|tc_no_mr|platform_missing","tc_id":"","tc_title":"","api":"GET /v2/...","action":"补用例/补MR/平台入库"}],
  "extras": [{"api":"","risk":""}],
  "suggestions": ["..."]
}`,
		req.Version, req.RequirementID, req.RequirementName, req.MRTag,
		jsonMarshal(changed), jsonMarshal(platform),
		truncateText(req.BeTechText, 8000), countCasesJSON(req.CasesJSON), tcSummary)
}

func buildMRIngestPrompt(beTech string, changed []map[string]string, folder []string, hint string) string {
	return fmt.Sprintf(`你是 API 测试工程师。根据**后端接口设计文档**与 **MR 变更接口列表**，生成可入库的接口定义（无需真实 HTTP 录制）。

## 后端技术设计
%s

## MR 变更接口
%s

## 目标目录（folder_path 数组）
%v

## 补充
%s

## 输出 JSON（不要 markdown）
{
  "apis": [
    {
      "name": "接口中文名",
      "method": "GET",
      "path": "/v2/foo/{{id}}",
      "headers": [{"name":"Authorization","value":"Bearer {{token}}","enabled":true}],
      "body": "",
      "body_type": "json",
      "description": "",
      "ai_remark": "来自 MR+设计文档",
      "folder_path": ["模块","子模块"],
      "assertions": [
        {"type":"status_code","expression":"","operator":"eq","expected":"200"},
        {"type":"json_path","expression":"$.code","operator":"eq","expected":"0"}
      ]
    }
  ]
}

要求：为 MR 中每个接口至少 1 条；path 用 pathname+query 模板；断言覆盖 code/关键字段。`,
		beTech, jsonMarshal(changed), folder, hint)
}

func countCasesJSON(raw string) int {
	var cases []any
	if err := json.Unmarshal([]byte(raw), &cases); err != nil {
		return 0
	}
	return len(cases)
}

func truncateText(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "\n...(truncated)"
}
