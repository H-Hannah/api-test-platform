package runner

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"api-test-platform/internal/store"
)

type Service struct {
	store *store.Store
}

func New(st *store.Store) *Service {
	return &Service{store: st}
}

type StepResult struct {
	StepOrder        int
	Name             string
	Status           string
	RequestSnapshot  map[string]any
	ResponseSnapshot map[string]any
	AssertionResults []AssertionResult
	DurationMS       int64
	ErrorMessage     string
}

func (s *Service) RunAPI(apiID, envID, datasetID int64) (*store.Run, error) {
	api, err := s.store.GetAPI(apiID)
	if err != nil {
		return nil, err
	}
	env, err := s.store.GetEnvironment(envID)
	if err != nil {
		return nil, err
	}

	vars := buildRunVars(env)
	body := api.Body
	headers := api.Headers
	assertions := assertionInputsFromStore(api.Assertions)

	if datasetID > 0 {
		ds, err := s.store.GetTestDataset(datasetID)
		if err != nil {
			return nil, err
		}
		if !store.DatasetBelongsToAPI(ds, api) {
			return nil, fmt.Errorf("dataset %d does not belong to api %d", datasetID, apiID)
		}
		applyDatasetToRun(ds, vars, &body, &headers)
		if dsAssertions := ParseAssertionListJSON(ds.Assertions); len(dsAssertions) > 0 {
			assertions = dsAssertions
		}
	}
	if len(assertions) == 0 {
		return nil, fmt.Errorf("no assertions configured: select a test case with assertions or add api-level assertions")
	}

	runID, err := s.store.CreateRun(&store.Run{APIID: &apiID, EnvID: envID, Status: "running"})
	if err != nil {
		return nil, err
	}

	fullTpl := ensureFullURLTemplate(api.FullURLTemplate, api.Path)
	step := store.ScenarioStep{
		Name: api.Name, Method: api.Method, Path: api.Path,
		Headers: headers, Body: body,
	}
	res := s.executeStep(step, vars, assertions, fullTpl)

	st := &store.RunStep{
		RunID: runID, StepOrder: 1, Name: api.Name, Status: res.Status,
		RequestSnapshot:  mustJSON(res.RequestSnapshot),
		ResponseSnapshot: mustJSON(res.ResponseSnapshot),
		AssertionResults: mustJSON(res.AssertionResults),
		DurationMS:       res.DurationMS,
		ErrorMessage:     res.ErrorMessage,
	}
	_, _ = s.store.CreateRunStep(st)

	status := "passed"
	if res.Status != "passed" {
		status = "failed"
	}
	summary, _ := json.Marshal(map[string]any{"total": 1, "passed": status == "passed"})
	_ = s.store.FinishRun(runID, status, string(summary))
	return s.store.GetRun(runID)
}

func (s *Service) RunScenario(scenarioID, envID int64) (*store.Run, error) {
	sc, err := s.store.GetScenario(scenarioID)
	if err != nil {
		return nil, err
	}
	if envID == 0 && sc.EnvID != nil {
		envID = *sc.EnvID
	}
	env, err := s.store.GetEnvironment(envID)
	if err != nil {
		return nil, err
	}

	runID, err := s.store.CreateRun(&store.Run{ScenarioID: &scenarioID, EnvID: envID, Status: "running"})
	if err != nil {
		return nil, err
	}

	vars := buildRunVars(env)

	passed, failed := 0, 0
	for _, step := range sc.Steps {
		assertions := ParseAssertionListJSON(step.Assertions)
		res := s.executeStep(step, vars, assertions, "")
		if res.Status == "passed" {
			passed++
		} else {
			failed++
		}
		st := &store.RunStep{
			RunID: runID, StepOrder: step.StepOrder, Name: step.Name, Status: res.Status,
			RequestSnapshot:  mustJSON(res.RequestSnapshot),
			ResponseSnapshot: mustJSON(res.ResponseSnapshot),
			AssertionResults: mustJSON(res.AssertionResults),
			DurationMS:       res.DurationMS,
			ErrorMessage:     res.ErrorMessage,
		}
		_, _ = s.store.CreateRunStep(st)
		if res.Status != "passed" {
			break
		}
	}

	status := "passed"
	if failed > 0 {
		status = "failed"
	}
	summary, _ := json.Marshal(map[string]any{"passed": passed, "failed": failed})
	_ = s.store.FinishRun(runID, status, string(summary))
	return s.store.GetRun(runID)
}

func (s *Service) executeStep(step store.ScenarioStep, vars map[string]string, assertions []AssertionInput, fullURLTemplate string) StepResult {
	res := StepResult{StepOrder: step.StepOrder, Name: step.Name, Status: "failed"}

	if len(assertions) == 0 {
		res.ErrorMessage = "no assertions configured"
		return res
	}

	url := substitute(ResolveRequestURL(step.Path, fullURLTemplate, vars), vars)
	method := strings.ToUpper(strings.TrimSpace(step.Method))
	bodyStr := substitute(strings.TrimSpace(step.Body), vars)
	headers := parseHeaders(step.Headers, vars)

	if err := validateNoUnresolved(url, "请求 URL"); err != nil {
		res.ErrorMessage = err.Error()
		return res
	}
	if err := validateNoUnresolved(bodyStr, "请求 Body"); err != nil {
		res.ErrorMessage = err.Error()
		return res
	}
	for k, v := range headers {
		if err := validateNoUnresolved(v, "请求头 "+k); err != nil {
			res.ErrorMessage = err.Error()
			return res
		}
	}

	var bodyReader io.Reader
	if bodyStr != "" && method != "GET" && method != "HEAD" {
		bodyReader = strings.NewReader(bodyStr)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		res.ErrorMessage = err.Error()
		return res
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res.RequestSnapshot = map[string]any{
		"method": method, "url": url, "headers": headers, "body": bodyStr,
	}

	client := &http.Client{Timeout: 60 * time.Second}
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start).Milliseconds()
	res.DurationMS = duration

	if err != nil {
		res.ErrorMessage = err.Error()
		return res
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	res.ResponseSnapshot = buildResponseSnapshot(resp.StatusCode, respBody, 51200)

	applyExtracts(step.ExtractRules, respBody, vars)
	res.AssertionResults = EvalAssertions(assertions, resp.StatusCode, respBody, duration)

	if AllPassed(res.AssertionResults) {
		res.Status = "passed"
	}
	return res
}

func applyExtracts(rulesJSON string, body []byte, vars map[string]string) {
	var rules []struct {
		Var      string `json:"var"`
		JSONPath string `json:"jsonPath"`
	}
	if rulesJSON == "" {
		return
	}
	_ = json.Unmarshal([]byte(rulesJSON), &rules)
	for _, r := range rules {
		if r.Var == "" || r.JSONPath == "" {
			continue
		}
		val := gjsonGet(body, r.JSONPath)
		if val != "" {
			vars[r.Var] = val
		}
	}
}

func gjsonGet(body []byte, path string) string {
	return jsonPathGet(body, path).String()
}

func parseHeaders(raw string, vars map[string]string) map[string]string {
	out := map[string]string{}
	var list []struct {
		Name    string `json:"name"`
		Value   string `json:"value"`
		Enabled bool   `json:"enabled"`
	}
	if raw == "" {
		return out
	}
	_ = json.Unmarshal([]byte(raw), &list)
	for _, h := range list {
		if h.Enabled && h.Name != "" {
			out[h.Name] = substitute(h.Value, vars)
		}
	}
	return out
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func assertionInputsFromStore(list []store.Assertion) []AssertionInput {
	out := make([]AssertionInput, len(list))
	for i, a := range list {
		expr := a.Expression
		if a.Type == "json_path" {
			expr = NormalizeJSONPathExpr(expr)
		}
		out[i] = AssertionInput{Type: a.Type, Expression: expr, Operator: a.Operator, Expected: a.Expected}
	}
	return out
}
