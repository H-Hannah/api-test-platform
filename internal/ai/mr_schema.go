package ai

import (
	"encoding/json"
	"strconv"
)

// MRAPIItem MR 变更中的接口条目。
type MRAPIItem struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Note   string `json:"note,omitempty"`
}

// MRExtra AI 识别的额外风险项。
type MRExtra struct {
	API  string `json:"api"`
	Risk string `json:"risk"`
}

func jsonMarshal(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

// MRVerifyTCRequest 阶段 C：按 Git 测试用例核对 MR。
type MRVerifyTCRequest struct {
	ProductID       int64       `json:"product_id"`
	MRTag           string      `json:"mr_tag"`
	Version         string      `json:"version"`
	RequirementID   string      `json:"requirement_id"`
	RequirementName string      `json:"requirement_name,omitempty"`
	ChangedAPIs     []MRAPIItem `json:"changed_apis"`
	CasesJSON       string      `json:"cases_json"`
	BeTechText      string      `json:"be_tech_text,omitempty"`
}

// MRVerifyTCResult AI + 规则核对结果。
type MRVerifyTCResult struct {
	Verdict           string          `json:"verdict"`
	Summary           string          `json:"summary"`
	Covered           []MRTCMatch     `json:"covered"`
	Gaps              []MRTCGap       `json:"gaps"`
	Extras            []MRExtra       `json:"extras"`
	Suggestions       []string        `json:"suggestions"`
	TCCount           int             `json:"tc_count"`
	MRAPICount        int             `json:"mr_api_count"`
	PlatformAPICount  int             `json:"platform_api_count"`
	LocalTCMatched    int             `json:"local_tc_matched,omitempty"`
}

type MRTCMatch struct {
	TCID     string `json:"tc_id"`
	TCTitle  string `json:"tc_title"`
	API      string `json:"api"`
	Note     string `json:"note,omitempty"`
}

type MRTCGap struct {
	Type       string `json:"type"` // mr_no_tc | tc_no_mr | platform_missing
	TCID       string `json:"tc_id,omitempty"`
	TCTitle    string `json:"tc_title,omitempty"`
	API        string `json:"api,omitempty"`
	Action     string `json:"action"`
}

// MRIngestAPIsRequest 从后端设计 + MR 变更生成平台接口定义。
type MRIngestAPIsRequest struct {
	ProductID       int64       `json:"product_id"`
	MRTag           string      `json:"mr_tag"`
	UserStory       string      `json:"user_story,omitempty"`
	TCRef           string      `json:"tc_ref,omitempty"`
	BeTechText      string      `json:"be_tech_text"`
	ChangedAPIs     []MRAPIItem `json:"changed_apis"`
	FolderPath      []string    `json:"folder_path,omitempty"`
	Hint            string      `json:"hint,omitempty"`
}
