package ai

// TestDataGenerateRequest AI 生成测试数据规格 + 数据集草案。
type TestDataGenerateRequest struct {
	ProductID       int64    `json:"product_id"`
	Version         string   `json:"version"`
	RequirementID   string   `json:"requirement_id"`
	RequirementName string   `json:"requirement_name"`
	PrdText         string   `json:"prd_text"`
	BeTechText      string   `json:"be_tech_text"`
	CasesJSON       string   `json:"cases_json,omitempty"`
	BDDText         string   `json:"bdd_text,omitempty"`
	APIHints        []string `json:"api_hints,omitempty"` // "GET /v2/foo"
	Hint            string   `json:"hint,omitempty"`
}

// TestDataCollectionAI 业务域测试集（如 Tracker、Brief）。
type TestDataCollectionAI struct {
	CollectionKey string              `json:"collection_key"`
	Name          string              `json:"name"`
	Description   string              `json:"description,omitempty"`
	Datasets      []TestDataDatasetAI `json:"datasets"`
}

type TestDataDatasetAI struct {
	DatasetKey      string            `json:"dataset_key"`
	CollectionKey   string            `json:"collection_key,omitempty"`
	CollectionName  string            `json:"collection_name,omitempty"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	TcRefs          []string          `json:"tc_refs"`
	ApiBindings     []string          `json:"api_bindings"`
	Variables       map[string]string `json:"variables"`
	HeadersOverride []HeaderKV        `json:"headers_override"`
	BodyOverride    string            `json:"body_override"`
	ObtainType      string            `json:"obtain_type"`
	ObtainNote      string            `json:"obtain_note"`
	Owner           string            `json:"owner"`
	Tags            []string          `json:"tags"`
}

type TestDataGenerateAIResult struct {
	Version         string                 `json:"version"`
	RequirementID   string                 `json:"requirement_id"`
	RequirementName string                 `json:"requirement_name"`
	EnvKeys         []string               `json:"env_keys"`
	Collections     []TestDataCollectionAI `json:"collections,omitempty"`
	Datasets        []TestDataDatasetAI    `json:"datasets"`
	CoverageNotes   string                 `json:"coverage_notes,omitempty"`
	GitOutputHint   string                 `json:"git_output_hint,omitempty"`
}

type TestDataGenerateResponse struct {
	Version         string                 `json:"version"`
	RequirementID   string                 `json:"requirement_id"`
	RequirementName string                 `json:"requirement_name"`
	EnvKeys         []string               `json:"env_keys"`
	Collections     []TestDataCollectionAI `json:"collections,omitempty"`
	Datasets        []TestDataDatasetAI    `json:"datasets"`
	SpecYAML        string              `json:"spec_yaml"`
	GitYAMLPath     string              `json:"git_yaml_path"`
	GitOutputHint   string              `json:"git_output_hint"`
	CoverageNotes   string              `json:"coverage_notes,omitempty"`
	Stats           TestDataStats       `json:"stats"`
	GatePassed      bool                `json:"gate_passed"`
	GateReasons     []string            `json:"gate_reasons"`
}

type TestDataStats struct {
	TotalDatasets    int            `json:"total_datasets"`
	TotalCollections int            `json:"total_collections"`
	ByObtain         map[string]int `json:"by_obtain"`
	EnvKeyCount      int            `json:"env_key_count"`
}

// TestDataImportRequest 将 AI 生成结果导入平台 SQLite。
type TestDataImportRequest struct {
	ProductID       int64               `json:"product_id"`
	Version         string              `json:"version"`
	RequirementID   string              `json:"requirement_id"`
	RequirementName string              `json:"requirement_name"`
	SpecYAML        string              `json:"spec_yaml"`
	EnvKeys         []string            `json:"env_keys"`
	Datasets        []TestDataDatasetAI `json:"datasets"`
}

type TestDataImportResponse struct {
	Imported   int     `json:"imported"`
	SpecID     int64   `json:"spec_id"`
	DatasetIDs []int64 `json:"dataset_ids"`
}

type ImportEnvKeysRequest struct {
	Keys        []string `json:"keys"`
	Placeholder string   `json:"placeholder,omitempty"`
}

type ImportEnvKeysResponse struct {
	Added int `json:"added"`
	Total int `json:"total"`
}
