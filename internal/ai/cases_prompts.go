package ai

import (
	"encoding/json"
	"fmt"

	"api-test-platform/internal/store"
)

type IngestCasesAIResult struct {
	Datasets []AICaseDataset `json:"datasets"`
}

type AICaseDataset struct {
	DatasetKey      string        `json:"dataset_key"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Variables       map[string]string `json:"variables"`
	HeadersOverride []HeaderKV    `json:"headers_override"`
	BodyOverride    string        `json:"body_override"`
	Assertions      []AIAssertion `json:"assertions"`
	Tags            []string      `json:"tags"`
}

func buildApiCasesPrompt(api *store.APIDefinition, record RawRecord, hint string) string {
	apiJSON, _ := json.Marshal(map[string]any{
		"id":          api.ID,
		"name":        api.Name,
		"method":      api.Method,
		"path":        api.Path,
		"headers":     api.Headers,
		"body":        api.Body,
		"description": api.Description,
	})
	recordJSON, _ := json.Marshal(record)

	return fmt.Sprintf(`你是资深接口测试工程师。根据已入库的接口定义，生成**单接口多案例**测试数据集。

### 目标
为同一个 API 设计多条可执行用例（test_datasets），覆盖成功、空结果、参数错误/边界等典型场景；由你根据接口语义自由发挥，不必固定条数。

### 用户提示（可为空）
%s

### 已入库接口定义
%s

### 录制流量（含响应时用于推断成功路径断言；无响应则按接口语义推断）
%s

### 输出 JSON Schema（不要 markdown，不要代码块）
{
  "datasets": [
    {
      "dataset_key": "success",
      "name": "查询成功",
      "description": "基于录制响应校验主路径",
      "variables": {},
      "headers_override": [],
      "body_override": "",
      "assertions": [
        {"type":"status_code","expression":"200","operator":"eq","expected":"200"},
        {"type":"json_path","expression":"$.code","operator":"eq","expected":"0"},
        {"type":"duration_ms","expression":"3000","operator":"lt","expected":""}
      ],
      "tags": []
    }
  ]
}

### 规则
1. datasets 至少 2 条，建议 2～5 条；必须包含一条「成功/主路径」用例（有录制响应时尽量贴合，否则按接口语义推断）。
2. 其余用例由你推断：如空列表、无数据、参数缺失/非法、权限不足等（按接口类型选择合理的子集）。
3. 用 variables / body_override / headers_override 区分请求差异；若接口 path 含 {{var}}，用 variables 覆盖。
4. 每条 dataset 必须有独立 assertions（status_code + 至少 1 条 json_path + duration_ms）。
5. assertions[].expected 必须是 JSON 字符串；json_path 针对响应体 JSON 根（$.code、$.data 等）。
6. 推断的负例/空态在 tags 中加入 "ai-inferred" 与 "draft"（待人工确认）。
7. dataset_key 用小写英文+连字符，如 success、empty-result、invalid-param。
8. 仅返回一个 JSON 对象，无其他文字。`,
		hint, string(apiJSON), string(recordJSON))
}
