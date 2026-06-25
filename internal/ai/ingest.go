package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"api-test-platform/internal/runner"
	"api-test-platform/internal/store"
)

type IngestService struct {
	ai    *Client
	store *store.Store
}

func NewIngestService(ai *Client, st *store.Store) *IngestService {
	return &IngestService{ai: ai, store: st}
}

func (s *IngestService) Ingest(req IngestRequest) (*IngestResponse, error) {
	req.ProductID = store.ResolveProductID(req.ProductID)
	if len(req.Records) == 0 {
		return nil, fmt.Errorf("records required")
	}
	mode := req.Mode
	if mode == "" {
		mode = "api"
	}

	tree, err := s.store.BuildFolderTree(req.ProductID)
	if err != nil {
		return nil, err
	}
	paths, err := s.store.FlatFolderPaths(req.ProductID)
	if err != nil {
		return nil, err
	}

	prompt := buildIngestPrompt(mode, req.Records, tree, paths, req.Hint)
	raw, err := s.ai.Complete(prompt)
	if err != nil {
		return nil, err
	}

	var result IngestAIResult
	if err := ParseJSON(raw, &result); err != nil {
		// retry once
		raw2, err2 := s.ai.Complete(prompt + "\n\n上次输出无法解析，请只输出合法 JSON。")
		if err2 != nil {
			return nil, fmt.Errorf("%v; retry: %v; raw: %s", err, err2, truncate(raw, 500))
		}
		if err := ParseJSON(raw2, &result); err != nil {
			return nil, fmt.Errorf("AI JSON parse failed: %w", err)
		}
	}

	out := &IngestResponse{}
	createdFolders := map[string]int64{}

	switch mode {
	case "scenario":
		if result.Scenario == nil {
			return nil, fmt.Errorf("AI did not return scenario")
		}
		saved, folders, err := s.saveScenario(req, *result.Scenario, createdFolders)
		if err != nil {
			return nil, err
		}
		out.Scenario = saved
		out.Folders = folders
	default:
		if len(result.APIs) == 0 {
			return nil, fmt.Errorf("AI did not return apis")
		}
		for _, item := range result.APIs {
			saved, folders, err := s.saveAPI(req, item, createdFolders)
			if err != nil {
				return nil, err
			}
			out.APIs = append(out.APIs, *saved)
			out.Folders = appendUniqueFolders(out.Folders, folders)
		}
	}

	return out, nil
}

func (s *IngestService) saveAPI(req IngestRequest, item AIAPIItem, cache map[string]int64) (*SavedAPI, []CreatedFolder, error) {
	folderID, folderPath, folders, err := s.resolveFolder(req.ProductID, item.FolderPath, cache)
	if err != nil {
		return nil, nil, err
	}

	svc := resolveServiceForItem(req.Records, item)
	pathOnly := PathOnly(item.Path)
	fullTpl := BuildFullURLTemplate(svc, item.Path)

	headersJSON, _ := json.Marshal(resolveAPIHeaders(item, req.Records))
	api := &store.APIDefinition{
		ProductID:       req.ProductID,
		FolderID:        folderID,
		Name:            item.Name,
		Method:          strings.ToUpper(item.Method),
		Path:            pathOnly,
		FullURLTemplate: fullTpl,
		Headers:         string(headersJSON),
		Body:            resolveAPIBody(item, req.Records),
		BodyType:        defaultBodyType(item.BodyType),
		Description:     item.Description,
		AIRemark:        item.AIRemark,
		SourceRecord:    matchSourceRecordJSON(req.Records, item),
	}
	id, err := s.store.CreateAPI(api)
	if err != nil {
		return nil, nil, err
	}

	assertions := make([]store.Assertion, len(item.Assertions))
	for i, a := range item.Assertions {
		expr := a.Expression
		if a.Type == "json_path" {
			expr = runner.NormalizeJSONPathExpr(expr)
		}
		assertions[i] = store.Assertion{
			Type: a.Type, Expression: expr,
			Operator: defaultOp(a.Operator), Expected: a.Expected, Enabled: true,
		}
	}
	if err := s.store.CreateAssertions(id, assertions); err != nil {
		return nil, nil, err
	}

	return &SavedAPI{ID: id, Name: api.Name, FolderID: folderID, FolderPath: folderPath}, folders, nil
}

func (s *IngestService) saveScenario(req IngestRequest, sc AIScenario, cache map[string]int64) (*SavedScenario, []CreatedFolder, error) {
	folderID, folderPath, folders, err := s.resolveFolder(req.ProductID, sc.FolderPath, cache)
	if err != nil {
		return nil, nil, err
	}

	var envID *int64
	if req.EnvID > 0 {
		envID = &req.EnvID
	}
	scenario := &store.Scenario{
		ProductID:   req.ProductID,
		FolderID:    folderID,
		Name:        sc.Name,
		Description: sc.Description,
		EnvID:       envID,
	}
	sid, err := s.store.CreateScenario(scenario)
	if err != nil {
		return nil, nil, err
	}

	for i, step := range sc.Steps {
		headersJSON, _ := json.Marshal(resolveStepHeaders(step, req.Records))
		extractJSON, _ := json.Marshal(step.ExtractRules)
		assertJSON, _ := json.Marshal(step.Assertions)
		svc := resolveServiceForStep(req.Records, step)
		stepPath := BuildStepRequestPath(svc, step.Path)
		st := &store.ScenarioStep{
			ScenarioID:   sid,
			StepOrder:    i + 1,
			Name:         step.Name,
			Method:       strings.ToUpper(step.Method),
			Path:         stepPath,
			Headers:      string(headersJSON),
			Body:         resolveStepBody(step, req.Records),
			ExtractRules: string(extractJSON),
			Assertions:   string(assertJSON),
		}
		if err := s.store.CreateScenarioStep(st); err != nil {
			return nil, nil, err
		}
	}

	return &SavedScenario{
		ID: sid, Name: sc.Name, FolderID: folderID,
		FolderPath: folderPath, StepCount: len(sc.Steps),
	}, folders, nil
}

func (s *IngestService) resolveFolder(productID int64, path []string, cache map[string]int64) (int64, string, []CreatedFolder, error) {
	key := strings.Join(path, "/")
	if key == "" {
		return 0, "", nil, nil
	}
	if id, ok := cache[key]; ok {
		return id, key, nil, nil
	}
	id, fullPath, err := s.store.EnsureFolderPath(productID, path)
	if err != nil {
		return 0, "", nil, err
	}
	cache[key] = id
	var created []CreatedFolder
	if id > 0 {
		created = append(created, CreatedFolder{ID: id, Path: fullPath})
	}
	return id, fullPath, created, nil
}

func defaultBodyType(t string) string {
	if t == "" {
		return "json"
	}
	return t
}

func defaultOp(op string) string {
	if op == "" {
		return "eq"
	}
	return op
}

func matchSourceRecordJSON(records []RawRecord, item AIAPIItem) string {
	method := strings.ToUpper(item.Method)
	for _, r := range records {
		if !strings.EqualFold(r.Method, method) {
			continue
		}
		rPath := r.Path
		if rPath == "" {
			rPath = pathFromURL(r.URL)
		}
		if rPath != "" && (rPath == item.Path || strings.Contains(rPath, item.Path) || strings.Contains(item.Path, rPath)) {
			b, _ := json.Marshal(map[string]string{
				"url": r.URL, "path": rPath, "host": r.Host, "service": r.Service,
			})
			return string(b)
		}
		if r.URL != "" && strings.Contains(r.URL, item.Path) {
			b, _ := json.Marshal(map[string]string{
				"url": r.URL, "path": rPath, "host": r.Host, "service": r.Service,
			})
			return string(b)
		}
	}
	return ""
}

func pathFromURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if i := strings.Index(raw, "://"); i >= 0 {
		raw = raw[i+3:]
	}
	if j := strings.Index(raw, "/"); j >= 0 {
		return raw[j:]
	}
	return ""
}

func appendUniqueFolders(dst []CreatedFolder, add []CreatedFolder) []CreatedFolder {
	seen := map[int64]bool{}
	for _, f := range dst {
		seen[f.ID] = true
	}
	for _, f := range add {
		if !seen[f.ID] {
			dst = append(dst, f)
			seen[f.ID] = true
		}
	}
	return dst
}
