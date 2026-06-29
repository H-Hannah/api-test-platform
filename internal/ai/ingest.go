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

	if mode == "api_cases" {
		return s.ingestAPICases(req)
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

	envs := s.ingestAllEnvironments()
	fullTpl, pathOnly := buildURLTemplate(req.Records, item, envs)

	headersJSON, _ := json.Marshal(templatizeHeaderKVs(resolveAPIHeaders(item, req.Records), envs))
	body := templatizeBody(resolveAPIBody(item, req.Records), envs)
	api := &store.APIDefinition{
		ProductID:       req.ProductID,
		FolderID:        folderID,
		Name:            item.Name,
		Method:          strings.ToUpper(item.Method),
		Path:            pathOnly,
		FullURLTemplate: fullTpl,
		Headers:         string(headersJSON),
		Body:            body,
		BodyType:        defaultBodyType(item.BodyType),
		Description:     item.Description,
		AIRemark:        item.AIRemark,
	}
	id, err := s.store.CreateAPI(api)
	if err != nil {
		return nil, nil, err
	}

	return &SavedAPI{ID: id, Name: api.Name, FolderID: folderID, FolderPath: folderPath}, folders, nil
}

func (s *IngestService) GenerateAPICases(apiID int64, hint string) (*IngestResponse, error) {
	api, err := s.store.GetAPI(apiID)
	if err != nil {
		return nil, fmt.Errorf("api not found: %w", err)
	}
	return s.ingestAPICases(IngestRequest{
		ProductID: api.ProductID,
		ApiID:     apiID,
		Mode:      "api_cases",
		Records:   []RawRecord{syntheticRecordFromAPI(api)},
		Hint:      hint,
	})
}

func syntheticRecordFromAPI(api *store.APIDefinition) RawRecord {
	rec := RawRecord{
		Method:      api.Method,
		Path:        api.Path,
		URL:         strings.TrimSpace(api.FullURLTemplate),
		RequestBody: api.Body,
		StatusCode:  200,
	}
	if rec.URL == "" {
		rec.URL = api.Path
	}
	var headers []HeaderKV
	_ = json.Unmarshal([]byte(api.Headers), &headers)
	for _, h := range headers {
		if !h.Enabled || h.Name == "" {
			continue
		}
		if rec.RequestHeaders == nil {
			rec.RequestHeaders = map[string]string{}
		}
		rec.RequestHeaders[h.Name] = h.Value
	}
	return rec
}

func (s *IngestService) ingestAPICases(req IngestRequest) (*IngestResponse, error) {
	if req.ApiID <= 0 {
		return nil, fmt.Errorf("api_id required for api_cases mode")
	}
	api, err := s.store.GetAPI(req.ApiID)
	if err != nil {
		return nil, fmt.Errorf("api not found: %w", err)
	}

	record := pickCaseRecord(req.Records, api)
	if record == nil {
		return nil, fmt.Errorf("no matching record for api")
	}

	prompt := buildApiCasesPrompt(api, *record, req.Hint)
	raw, err := s.ai.Complete(prompt)
	if err != nil {
		return nil, err
	}

	var result IngestCasesAIResult
	if err := ParseJSON(raw, &result); err != nil {
		raw2, err2 := s.ai.Complete(prompt + "\n\n上次输出无法解析，请只输出合法 JSON。")
		if err2 != nil {
			return nil, fmt.Errorf("%v; retry: %v; raw: %s", err, err2, truncate(raw, 500))
		}
		if err := ParseJSON(raw2, &result); err != nil {
			return nil, fmt.Errorf("AI JSON parse failed: %w", err)
		}
	}
	if len(result.Datasets) == 0 {
		return nil, fmt.Errorf("AI did not return datasets")
	}

	envs := s.ingestAllEnvironments()
	binding := strings.ToUpper(strings.TrimSpace(api.Method)) + " " + strings.TrimSpace(api.Path)
	out := &IngestResponse{}
	for _, item := range result.Datasets {
		saved, err := s.saveCaseDataset(req.ProductID, api.ID, binding, item, envs)
		if err != nil {
			return nil, err
		}
		out.Datasets = append(out.Datasets, *saved)
	}
	return out, nil
}

func pickCaseRecord(records []RawRecord, api *store.APIDefinition) *RawRecord {
	item := AIAPIItem{Method: api.Method, Path: api.Path}
	if rec := matchRecord(records, item); rec != nil {
		return rec
	}
	if len(records) == 1 {
		return &records[0]
	}
	return nil
}

func (s *IngestService) saveCaseDataset(productID, apiID int64, binding string, item AICaseDataset, envs []*store.Environment) (*SavedDataset, error) {
	api, err := s.store.GetAPI(apiID)
	if err != nil {
		return nil, err
	}
	varsJSON, _ := json.Marshal(item.Variables)
	if len(item.Variables) == 0 {
		varsJSON = []byte("{}")
	}
	headers := templatizeHeaderKVs(item.HeadersOverride, envs)
	headersJSON, _ := json.Marshal(headers)
	body := templatizeBody(item.BodyOverride, envs)
	assertions := assertionsToRunnerJSON(item.Assertions)
	tagsJSON, _ := json.Marshal(normalizeCaseTags(item.Tags))

	key := strings.TrimSpace(item.DatasetKey)
	if key == "" {
		key = slugDatasetKey(item.Name)
	}
	name := strings.TrimSpace(item.Name)
	if name == "" {
		name = key
	}

	ds := &store.TestDataset{
		ProductID:       productID,
		Version:         "recorder",
		RequirementID:   fmt.Sprintf("api-%d", apiID),
		DatasetKey:      key,
		Name:            name,
		Description:     strings.TrimSpace(item.Description),
		TcRefs:          "[]",
		ApiBindings:     mustJSONArray([]string{binding}),
		Variables:       string(varsJSON),
		HeadersOverride: string(headersJSON),
		BodyOverride:    body,
		ObtainType:      "recorder",
		Owner:           "qa",
		Tags:            string(tagsJSON),
		Source:          "ai",
		Assertions:      assertions,
		ApiFingerprint:  store.APIDefinitionFingerprint(api),
	}
	id, err := s.store.UpsertTestDataset(ds)
	if err != nil {
		return nil, err
	}
	return &SavedDataset{ID: id, Name: name, DatasetKey: key, ApiID: apiID}, nil
}

func normalizeCaseTags(tags []string) []string {
	out := make([]string, 0, len(tags)+2)
	seen := map[string]bool{}
	hasInferred := false
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
		if t == "ai-inferred" {
			hasInferred = true
		}
	}
	if hasInferred {
		for _, extra := range []string{"ai-inferred", "draft"} {
			if !seen[extra] {
				seen[extra] = true
				out = append(out, extra)
			}
		}
	}
	return out
}

func assertionsToRunnerJSON(list []AIAssertion) string {
	if len(list) == 0 {
		return "[]"
	}
	out := make([]runner.AssertionInput, len(list))
	for i, a := range list {
		expr := a.Expression
		if a.Type == "json_path" {
			expr = runner.NormalizeJSONPathExpr(expr)
		}
		out[i] = runner.AssertionInput{
			Type: a.Type, Expression: expr,
			Operator: defaultOp(a.Operator), Expected: a.Expected,
		}
	}
	b, _ := json.Marshal(out)
	return string(b)
}

func slugDatasetKey(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	s := strings.Trim(b.String(), "-")
	if s == "" {
		return "case"
	}
	return s
}

func mustJSONArray(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
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
		envs := s.ingestAllEnvironments()
		headersJSON, _ := json.Marshal(templatizeHeaderKVs(resolveStepHeaders(step, req.Records), envs))
		extractJSON, _ := json.Marshal(step.ExtractRules)
		assertJSON, _ := json.Marshal(step.Assertions)
		item := AIAPIItem{Method: step.Method, Path: step.Path}
		stepPath, _ := buildURLTemplate(req.Records, item, envs)
		st := &store.ScenarioStep{
			ScenarioID:   sid,
			StepOrder:    i + 1,
			Name:         step.Name,
			Method:       strings.ToUpper(step.Method),
			Path:         stepPath,
			Headers:      string(headersJSON),
			Body:         templatizeBody(resolveStepBody(step, req.Records), envs),
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
	rec := matchRecord(records, item)
	if rec == nil {
		return ""
	}
	rPath := rec.Path
	if rPath == "" {
		rPath = pathFromURL(rec.URL)
	}
	b, _ := json.Marshal(map[string]string{
		"url": rec.URL, "path": rPath, "host": rec.Host, "service": rec.Service,
	})
	return string(b)
}

func (s *IngestService) ingestAllEnvironments() []*store.Environment {
	envs, err := s.store.ListAllEnvironments()
	if err != nil || len(envs) == 0 {
		return nil
	}
	out := make([]*store.Environment, len(envs))
	for i := range envs {
		out[i] = &envs[i]
	}
	return out
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
