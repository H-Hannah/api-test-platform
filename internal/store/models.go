package store

type Product struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type Folder struct {
	ID        int64    `json:"id"`
	ProductID int64    `json:"product_id"`
	ParentID  int64    `json:"parent_id"`
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	CreatedAt string   `json:"created_at"`
	Children  []Folder `json:"children,omitempty"`
}

type Environment struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	BaseURL   string `json:"base_url"`
	Variables string `json:"variables"`
	IsDefault bool   `json:"is_default"`
	CreatedAt string `json:"created_at"`
}

type APIDefinition struct {
	ID              int64  `json:"id"`
	ProductID       int64  `json:"product_id"`
	FolderID        int64  `json:"folder_id"`
	FolderPath      string `json:"folder_path,omitempty"`
	Name            string `json:"name"`
	Method          string `json:"method"`
	Path            string `json:"path"`
	FullURLTemplate string `json:"full_url_template"`
	Headers         string `json:"headers"`
	Body            string `json:"body"`
	BodyType        string `json:"body_type"`
	Description     string `json:"description"`
	AIRemark        string `json:"ai_remark"`
	SourceRecord    string `json:"source_record,omitempty"`
	UserStory       string `json:"user_story"`
	BDDRef          string `json:"bdd_ref"`
	TCRef           string `json:"tc_ref"`
	MRTags          string `json:"mr_tags"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	Assertions      []Assertion `json:"assertions,omitempty"`
	AssertionCount  int    `json:"assertion_count,omitempty"`
	ScenarioReady   bool   `json:"scenario_ready,omitempty"`
}

// APICoverage 产品维度场景覆盖统计（MR / BDD 驱动）。
type APICoverage struct {
	Total          int            `json:"total"`
	WithUserStory  int            `json:"with_user_story"`
	WithBDD        int            `json:"with_bdd"`
	WithTC         int            `json:"with_tc"`
	WithAssertions int            `json:"with_assertions"`
	ScenarioReady  int            `json:"scenario_ready"`
	GapNoUS        int            `json:"gap_no_user_story"`
	GapNoBDD       int            `json:"gap_no_bdd"`
	GapNoTC        int            `json:"gap_no_tc"`
	GapNoAssert    int            `json:"gap_no_assertions"`
	ByMR           map[string]int `json:"by_mr"`
}

type Assertion struct {
	ID         int64  `json:"id,omitempty"`
	APIID      int64  `json:"api_id,omitempty"`
	Type       string `json:"type"`
	Expression string `json:"expression"`
	Operator   string `json:"operator"`
	Expected   string `json:"expected"`
	Enabled    bool   `json:"enabled"`
}

type Scenario struct {
	ID          int64  `json:"id"`
	ProductID   int64  `json:"product_id"`
	FolderID    int64  `json:"folder_id"`
	FolderPath  string `json:"folder_path,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	EnvID       *int64 `json:"env_id"`
	CreatedAt   string `json:"created_at"`
	Steps       []ScenarioStep `json:"steps,omitempty"`
}

type ScenarioStep struct {
	ID           int64  `json:"id,omitempty"`
	ScenarioID   int64  `json:"scenario_id,omitempty"`
	StepOrder    int    `json:"step_order"`
	Name         string `json:"name"`
	APIID        *int64 `json:"api_id"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	Headers      string `json:"headers"`
	Body         string `json:"body"`
	ExtractRules string `json:"extract_rules"`
	Assertions   string `json:"assertions"`
}

type Run struct {
	ID         int64   `json:"id"`
	ScenarioID *int64  `json:"scenario_id"`
	APIID      *int64  `json:"api_id"`
	EnvID      int64   `json:"env_id"`
	Status     string  `json:"status"`
	StartedAt  string  `json:"started_at"`
	FinishedAt *string `json:"finished_at"`
	Summary    string  `json:"summary"`
	Steps      []RunStep `json:"steps,omitempty"`
}

type RunStep struct {
	ID               int64  `json:"id"`
	RunID            int64  `json:"run_id"`
	StepOrder        int    `json:"step_order"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	RequestSnapshot  string `json:"request_snapshot"`
	ResponseSnapshot string `json:"response_snapshot"`
	AssertionResults string `json:"assertion_results"`
	DurationMS       int64  `json:"duration_ms"`
	ErrorMessage     string `json:"error_message"`
}

// TestDataset 可注入执行的测试数据（变量覆盖、body 覆盖）。
type TestDataset struct {
	ID              int64  `json:"id"`
	ProductID       int64  `json:"product_id"`
	Version         string `json:"version"`
	RequirementID   string `json:"requirement_id"`
	DatasetKey      string `json:"dataset_key"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	TcRefs          string `json:"tc_refs"`
	ApiBindings     string `json:"api_bindings"`
	Variables       string `json:"variables"`
	HeadersOverride string `json:"headers_override"`
	BodyOverride    string `json:"body_override"`
	ObtainType      string `json:"obtain_type"`
	ObtainNote      string `json:"obtain_note"`
	Owner           string `json:"owner"`
	Tags            string `json:"tags"`
	Source          string `json:"source"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// TestDataSpec 需求包维度的数据规格文档（YAML）。
type TestDataSpec struct {
	ID              int64  `json:"id"`
	ProductID       int64  `json:"product_id"`
	Version         string `json:"version"`
	RequirementID   string `json:"requirement_id"`
	RequirementName string `json:"requirement_name"`
	SpecYAML        string `json:"spec_yaml"`
	EnvKeys         string `json:"env_keys"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// FolderTreeNode is sent to AI for classification context.
type FolderTreeNode struct {
	ID       int64            `json:"id"`
	Name     string           `json:"name"`
	Path     string           `json:"path"`
	Children []FolderTreeNode `json:"children,omitempty"`
}
