package ai

import (
	"encoding/json"
	"strconv"
	"strings"
)

// IngestRequest from Chrome extension.
type IngestRequest struct {
	ProductID  int64         `json:"product_id"`
	EnvID      int64         `json:"env_id,omitempty"`
	Mode       string        `json:"mode"` // api | scenario | api_cases
	ApiID      int64         `json:"api_id,omitempty"`
	Records    []RawRecord   `json:"records"`
	Hint       string        `json:"hint,omitempty"` // optional business hint from user
}

type RawRecord struct {
	URL             string            `json:"url"`
	Path            string            `json:"path,omitempty"`   // pathname + query，入库 path 依据
	Host            string            `json:"host,omitempty"`
	Service         string            `json:"service,omitempty"` // 从 host/路径推断的微服务名
	Method          string            `json:"method"`
	RequestHeaders  map[string]string `json:"requestHeaders"`
	RequestBody     string            `json:"requestBody"`
	ResponseHeaders map[string]string `json:"responseHeaders"`
	ResponseBody    string            `json:"responseBody"`
	StatusCode      int               `json:"statusCode"`
	Timestamp       int64             `json:"timestamp"`
	RequestID       string            `json:"requestId,omitempty"`
}

// IngestAIResult is the strict JSON contract from LLM.
type IngestAIResult struct {
	APIs     []AIAPIItem     `json:"apis"`
	Scenario *AIScenario     `json:"scenario,omitempty"`
}

type AIAPIItem struct {
	Name       string          `json:"name"`
	Method     string          `json:"method"`
	Path       string          `json:"path"`
	Headers    []HeaderKV      `json:"headers"`
	Body       string          `json:"body"`
	BodyType   string          `json:"body_type"`
	Description string         `json:"description"`
	AIRemark   string          `json:"ai_remark"`
	FolderPath []string        `json:"folder_path"` // AI auto classification
	Assertions []AIAssertion   `json:"assertions"`
}

type AIScenario struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	FolderPath  []string       `json:"folder_path"`
	Steps       []AIScenarioStep `json:"steps"`
}

type AIScenarioStep struct {
	Name         string        `json:"name"`
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	Headers      []HeaderKV    `json:"headers"`
	Body         string        `json:"body"`
	ExtractRules []ExtractRule `json:"extract_rules"`
	Assertions   []AIAssertion `json:"assertions"`
}

type HeaderKV struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

type AIAssertion struct {
	Type       string `json:"type"` // status_code | json_path | duration_ms
	Expression string `json:"expression"`
	Operator   string `json:"operator"`
	Expected   string `json:"expected"`
}

// UnmarshalJSON 兼容 LLM 将 expected 输出为 bool/number 的情况，统一转为 string。
func (a *AIAssertion) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type       string          `json:"type"`
		Expression string          `json:"expression"`
		Operator   string          `json:"operator"`
		Expected   json.RawMessage `json:"expected"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	a.Type = raw.Type
	a.Expression = raw.Expression
	a.Operator = raw.Operator
	a.Expected = coerceAssertionExpected(raw.Expected)
	return nil
}

func coerceAssertionExpected(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	s := strings.TrimSpace(string(raw))
	if s == "null" {
		return ""
	}
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		return strconv.FormatBool(b)
	}
	var n json.Number
	if err := json.Unmarshal(raw, &n); err == nil {
		return n.String()
	}
	return strings.Trim(s, `"`)
}

type ExtractRule struct {
	Var      string `json:"var"`
	JSONPath string `json:"jsonPath"`
}

// IngestResponse returned to plugin.
type IngestResponse struct {
	APIs      []SavedAPI      `json:"apis,omitempty"`
	Scenario  *SavedScenario  `json:"scenario,omitempty"`
	Datasets  []SavedDataset  `json:"datasets,omitempty"`
	Folders   []CreatedFolder `json:"folders_created,omitempty"`
}

type SavedDataset struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	DatasetKey string `json:"dataset_key"`
	ApiID      int64  `json:"api_id"`
}

type SavedAPI struct {
	ID         int64    `json:"id"`
	Name       string   `json:"name"`
	FolderID   int64    `json:"folder_id"`
	FolderPath string   `json:"folder_path"`
}

type SavedScenario struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	FolderID   int64  `json:"folder_id"`
	FolderPath string `json:"folder_path"`
	StepCount  int    `json:"step_count"`
}

type CreatedFolder struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
}
