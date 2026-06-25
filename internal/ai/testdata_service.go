package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"api-test-platform/internal/store"
)

type TestDataService struct {
	ai    *Client
	store *store.Store
}

func NewTestDataService(ai *Client, st *store.Store) *TestDataService {
	return &TestDataService{ai: ai, store: st}
}

func (s *TestDataService) Generate(req TestDataGenerateRequest) (*TestDataGenerateResponse, error) {
	req.ProductID = store.ResolveProductID(req.ProductID)
	be := strings.TrimSpace(req.BeTechText)
	prd := strings.TrimSpace(req.PrdText)
	if prd == "" && be == "" {
		return nil, fmt.Errorf("prd_text 或 be_tech_text 至少填一项（建议从 GitLab 加载三类文档）")
	}
	if strings.TrimSpace(req.Version) == "" {
		return nil, fmt.Errorf("version required")
	}
	if strings.TrimSpace(req.RequirementID) == "" {
		return nil, fmt.Errorf("requirement_id required")
	}
	if strings.TrimSpace(req.RequirementName) == "" {
		req.RequirementName = req.RequirementID
	}

	prompt := buildTestDataGeneratePrompt(req)
	raw, err := s.ai.Complete(prompt)
	if err != nil {
		return nil, err
	}
	var result TestDataGenerateAIResult
	if err := ParseJSON(raw, &result); err != nil {
		raw2, err2 := s.ai.Complete(prompt + "\n\n上次输出无法解析，请只输出合法 JSON。")
		if err2 != nil {
			return nil, fmt.Errorf("%v; retry: %v", err, err2)
		}
		if err := ParseJSON(raw2, &result); err != nil {
			return nil, fmt.Errorf("test data JSON parse failed: %w", err)
		}
	}
	if strings.TrimSpace(result.Version) == "" {
		result.Version = req.Version
	}
	if strings.TrimSpace(result.RequirementID) == "" {
		result.RequirementID = req.RequirementID
	}
	if strings.TrimSpace(result.RequirementName) == "" {
		result.RequirementName = req.RequirementName
	}

	collections := flattenTestDataCollections(&result)

	passed, reasons := EvaluateTestDataGate(result.Datasets, result.EnvKeys)
	yaml := RenderTestDataYAML(result.Version, result.RequirementID, result.RequirementName,
		result.EnvKeys, collections, result.Datasets, result.CoverageNotes)
	yamlPath := SuggestTestDataGitPath(result.Version, result.RequirementID)
	hint := strings.TrimSpace(result.GitOutputHint)
	if hint == "" {
		hint = yamlPath
	}

	return &TestDataGenerateResponse{
		Version:         result.Version,
		RequirementID:   result.RequirementID,
		RequirementName: result.RequirementName,
		EnvKeys:         result.EnvKeys,
		Collections:     collections,
		Datasets:        result.Datasets,
		SpecYAML:        yaml,
		GitYAMLPath:     yamlPath,
		GitOutputHint:   hint,
		CoverageNotes:   result.CoverageNotes,
		Stats:           BuildTestDataStats(result.Datasets, result.EnvKeys, collections),
		GatePassed:      passed,
		GateReasons:     reasons,
	}, nil
}

func (s *TestDataService) Import(req TestDataImportRequest) (*TestDataImportResponse, error) {
	req.ProductID = store.ResolveProductID(req.ProductID)
	if strings.TrimSpace(req.Version) == "" || strings.TrimSpace(req.RequirementID) == "" {
		return nil, fmt.Errorf("version and requirement_id required")
	}
	if len(req.Datasets) == 0 {
		return nil, fmt.Errorf("datasets required")
	}
	envKeysJSON, _ := json.Marshal(req.EnvKeys)
	specYAML := strings.TrimSpace(req.SpecYAML)
	if specYAML == "" {
		specYAML = RenderTestDataYAML(req.Version, req.RequirementID, req.RequirementName,
			req.EnvKeys, nil, req.Datasets, "")
	}
	specID, err := s.store.UpsertTestDataSpec(&store.TestDataSpec{
		ProductID:       req.ProductID,
		Version:         req.Version,
		RequirementID:   req.RequirementID,
		RequirementName: req.RequirementName,
		SpecYAML:        specYAML,
		EnvKeys:         string(envKeysJSON),
	})
	if err != nil {
		return nil, err
	}

	var ids []int64
	for _, ds := range req.Datasets {
		row := datasetAIToStore(req.ProductID, req.Version, req.RequirementID, ds)
		id, err := s.store.UpsertTestDataset(&store.TestDataset{
			ProductID:       row.ProductID,
			Version:         row.Version,
			RequirementID:   row.RequirementID,
			DatasetKey:      row.DatasetKey,
			Name:            row.Name,
			Description:     row.Description,
			TcRefs:          row.TcRefs,
			ApiBindings:     row.ApiBindings,
			Variables:       row.Variables,
			HeadersOverride: row.HeadersOverride,
			BodyOverride:    row.BodyOverride,
			ObtainType:      row.ObtainType,
			ObtainNote:      row.ObtainNote,
			Owner:           row.Owner,
			Tags:            row.Tags,
			Source:          row.Source,
		})
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return &TestDataImportResponse{
		Imported:   len(ids),
		SpecID:     specID,
		DatasetIDs: ids,
	}, nil
}

func (s *TestDataService) ImportEnvKeys(envID int64, req ImportEnvKeysRequest) (*ImportEnvKeysResponse, error) {
	if envID <= 0 {
		return nil, fmt.Errorf("env_id required")
	}
	if len(req.Keys) == 0 {
		return nil, fmt.Errorf("keys required")
	}
	added, err := s.store.MergeEnvVarKeys(envID, req.Keys, req.Placeholder)
	if err != nil {
		return nil, err
	}
	return &ImportEnvKeysResponse{Added: added, Total: len(req.Keys)}, nil
}
