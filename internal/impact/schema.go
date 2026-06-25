package impact

// AnalyzeRequest 精准测试：代码变更 → 推荐用例与接口。
type AnalyzeRequest struct {
	ProductID     int64    `json:"product_id"`
	Version       string   `json:"version"`
	RequirementID string   `json:"requirement_id"`
	MRTag         string   `json:"mr_tag,omitempty"`
	GitLabMRURL   string   `json:"gitlab_mr_url,omitempty"`
	GitHubPRURL   string   `json:"github_pr_url,omitempty"`
	Repo          string   `json:"repo,omitempty"`     // owner/repo，compare 模式
	BaseRef       string   `json:"base_ref,omitempty"` // main
	HeadRef       string   `json:"head_ref,omitempty"` // feature branch
	DiffText      string   `json:"diff_text,omitempty"`
	ChangedFiles  []string `json:"changed_files,omitempty"`
	ChangedAPIs       []APIItem `json:"changed_apis,omitempty"`
	ChangeDescription string    `json:"change_description,omitempty"` // 口述变更（配置、开关等无 MR 时）
	CasesJSON         string    `json:"cases_json,omitempty"`
	UseAI         bool     `json:"use_ai,omitempty"`
}

type APIItem struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Note   string `json:"note,omitempty"`
}

type ChangedFile struct {
	Path   string `json:"path"`
	Status string `json:"status,omitempty"`
	Module string `json:"module,omitempty"`
}

type RecommendedTC struct {
	TCID     string   `json:"tc_id"`
	Title    string   `json:"title"`
	Priority string   `json:"priority"`
	Score    float64  `json:"score"`
	Reasons  []string `json:"reasons"`
}

type RecommendedAPI struct {
	APIID         int64    `json:"api_id"`
	Method        string   `json:"method"`
	Path          string   `json:"path"`
	Name          string   `json:"name"`
	ScenarioReady bool     `json:"scenario_ready"`
	Score         float64  `json:"score"`
	Reasons       []string `json:"reasons"`
}

type RecommendedScenario struct {
	ScenarioID int64    `json:"scenario_id"`
	Name       string   `json:"name"`
	Score      float64  `json:"score"`
	Reasons    []string `json:"reasons"`
}

type ImpactGap struct {
	Type   string `json:"type"`
	File   string `json:"file,omitempty"`
	API    string `json:"api,omitempty"`
	TCID   string `json:"tc_id,omitempty"`
	Action string `json:"action"`
}

type RunPlan struct {
	P0TCIDs      []string `json:"p0_tc_ids"`
	P1TCIDs      []string `json:"p1_tc_ids"`
	APIIDs       []int64  `json:"api_ids"`
	ScenarioIDs  []int64  `json:"scenario_ids"`
}

type ImpactStats struct {
	ChangedFileCount int `json:"changed_file_count"`
	InferredAPICount int `json:"inferred_api_count"`
	RecommendedTC    int `json:"recommended_tc"`
	RecommendedAPI     int `json:"recommended_api"`
	PlatformReadyAPI   int `json:"platform_ready_api"`
}

type AnalyzeResult struct {
	Source              string                `json:"source"`
	ChangedFiles        []ChangedFile         `json:"changed_files"`
	InferredAPIs        []string              `json:"inferred_apis"`
	RecommendedTCs       []RecommendedTC       `json:"recommended_tcs"`
	RecommendedTCReason  string                `json:"recommended_tc_reason,omitempty"`
	RecommendedAPIs      []RecommendedAPI      `json:"recommended_apis"`
	RecommendedScenarios []RecommendedScenario `json:"recommended_scenarios"`
	Gaps                []ImpactGap           `json:"gaps"`
	RunPlan             RunPlan               `json:"run_plan"`
	Summary             string                `json:"summary"`
	AISummary           string                `json:"ai_summary,omitempty"`
	Stats               ImpactStats           `json:"stats"`
}

type RunPlanRequest struct {
	ProductID   int64   `json:"product_id"`
	EnvID       int64   `json:"env_id"`
	APIIDs      []int64 `json:"api_ids"`
	ScenarioIDs []int64 `json:"scenario_ids"`
	DatasetID   int64   `json:"dataset_id,omitempty"`
}

type RunPlanResult struct {
	Total   int              `json:"total"`
	Passed  int              `json:"passed"`
	Failed  int              `json:"failed"`
	Runs    []RunPlanItem    `json:"runs"`
}

type RunPlanItem struct {
	Kind       string `json:"kind"` // api | scenario
	TargetID   int64  `json:"target_id"`
	Name       string `json:"name"`
	RunID      int64  `json:"run_id,omitempty"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
}

// PostMRCommentRequest 将精准测试报告发表到 GitLab MR。
type PostMRCommentRequest struct {
	GitLabMRURL   string         `json:"gitlab_mr_url"`
	Version       string         `json:"version,omitempty"`
	RequirementID string         `json:"requirement_id,omitempty"`
	TCDocsBranch  string         `json:"tc_docs_branch,omitempty"`
	Result        *AnalyzeResult `json:"result"`
}

type PostMRCommentResult struct {
	NoteID   int64  `json:"note_id"`
	NoteURL  string `json:"note_url"`
	Markdown string `json:"markdown"`
}
