package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"api-test-platform/internal/store"
)

type MRService struct {
	ai     *Client
	ingest *IngestService
	store  *store.Store
	repo   string
}

func NewMRService(ai *Client, ingest *IngestService, st *store.Store, docsRepoRoot string) *MRService {
	return &MRService{ai: ai, ingest: ingest, store: st, repo: docsRepoRoot}
}

func (s *MRService) VerifyMRWithTC(req MRVerifyTCRequest) (*MRVerifyTCResult, error) {
	productID := req.ProductID
	if productID < 0 {
		return nil, fmt.Errorf("invalid product_id")
	}
	mrTag := strings.TrimSpace(req.MRTag)
	if mrTag == "" {
		return nil, fmt.Errorf("mr_tag required")
	}
	if len(req.ChangedAPIs) == 0 {
		return nil, fmt.Errorf("changed_apis required")
	}
	casesJSON := strings.TrimSpace(req.CasesJSON)
	if casesJSON == "" {
		return nil, fmt.Errorf("cases_json required：请先从 GitLab 链接加载测试用例")
	}

	var platformAPIs []map[string]string
	list, err := s.store.ListAPIsFiltered(store.APIListFilter{
		ProductID: productID,
		MRTag:     mrTag,
	})
	if err != nil {
		return nil, err
	}
	for _, a := range list {
		platformAPIs = append(platformAPIs, map[string]string{
			"method": a.Method,
			"path":   a.Path,
			"name":   a.Name,
			"ready":  fmt.Sprintf("%v", a.ScenarioReady),
			"tc_ref": a.TCRef,
		})
	}
	changed := make([]map[string]string, 0, len(req.ChangedAPIs))
	for _, c := range req.ChangedAPIs {
		changed = append(changed, map[string]string{
			"method": c.Method,
			"path":   c.Path,
			"note":   c.Note,
		})
	}

	var cases []map[string]any
	_ = json.Unmarshal([]byte(casesJSON), &cases)
	tcSummary := docsrepoSummarize(cases)

	prompt := buildMRVerifyTCPrompt(req, changed, platformAPIs, tcSummary)
	raw, err := s.ai.Complete(prompt)
	if err != nil {
		return nil, err
	}
	var result MRVerifyTCResult
	if err := ParseJSON(raw, &result); err != nil {
		return nil, fmt.Errorf("verify TC JSON parse failed: %w", err)
	}
	result.TCCount = len(cases)
	result.MRAPICount = len(req.ChangedAPIs)
	result.PlatformAPICount = len(platformAPIs)
	result.LocalTCMatched = localMatchMRToTC(req.ChangedAPIs, cases)
	if result.Verdict == "" {
		if len(result.Gaps) == 0 {
			result.Verdict = "pass"
		} else {
			result.Verdict = "gap"
		}
	}
	return &result, nil
}

func (s *MRService) IngestAPIsFromMR(req MRIngestAPIsRequest) (*IngestResponse, error) {
	req.ProductID = store.ResolveProductID(req.ProductID)
	mrTag := strings.TrimSpace(req.MRTag)
	if mrTag == "" {
		return nil, fmt.Errorf("mr_tag required")
	}
	be := strings.TrimSpace(req.BeTechText)
	if be == "" {
		return nil, fmt.Errorf("be_tech_text required")
	}
	if len(req.ChangedAPIs) == 0 {
		return nil, fmt.Errorf("changed_apis required")
	}

	changed := make([]map[string]string, 0, len(req.ChangedAPIs))
	for _, c := range req.ChangedAPIs {
		changed = append(changed, map[string]string{
			"method": c.Method,
			"path":   c.Path,
			"note":   c.Note,
		})
	}
	prompt := buildMRIngestPrompt(be, changed, req.FolderPath, req.Hint)
	raw, err := s.ai.Complete(prompt)
	if err != nil {
		return nil, err
	}
	var result IngestAIResult
	if err := ParseJSON(raw, &result); err != nil {
		raw2, err2 := s.ai.Complete(prompt + "\n\n上次输出无法解析，请只输出合法 JSON。")
		if err2 != nil {
			return nil, fmt.Errorf("%v; retry: %v", err, err2)
		}
		if err := ParseJSON(raw2, &result); err != nil {
			return nil, fmt.Errorf("ingest MR JSON parse failed: %w", err)
		}
	}
	if len(result.APIs) == 0 {
		return nil, fmt.Errorf("AI 未返回 apis")
	}

	out := &IngestResponse{}
	createdFolders := map[string]int64{}
	ingestReq := IngestRequest{ProductID: req.ProductID, Mode: "api", Records: []RawRecord{}}

	for _, item := range result.APIs {
		if len(item.FolderPath) == 0 && len(req.FolderPath) > 0 {
			item.FolderPath = req.FolderPath
		}
		saved, folders, err := s.ingest.saveAPI(ingestReq, item, createdFolders)
		if err != nil {
			return nil, err
		}
		_ = s.store.AppendMRTags(saved.ID, mrTag)
		us := strings.TrimSpace(req.UserStory)
		tc := strings.TrimSpace(req.TCRef)
		bdd := ""
		if tc != "" {
			bdd = "TC:" + tc
		}
		if us != "" || tc != "" {
			apiRow, _ := s.store.GetAPI(saved.ID)
			if apiRow != nil {
				if us == "" {
					us = apiRow.UserStory
				}
				if tc == "" {
					tc = apiRow.TCRef
				}
				_ = s.store.UpdateAPIMeta(saved.ID, us, bdd, tc, apiRow.MRTags)
			}
		}
		out.APIs = append(out.APIs, *saved)
		out.Folders = appendUniqueFolders(out.Folders, folders)
	}
	return out, nil
}

func docsrepoSummarize(cases []map[string]any) string {
	var parts []string
	for i, c := range cases {
		if i >= 50 {
			parts = append(parts, fmt.Sprintf("...共 %d 条", len(cases)))
			break
		}
		title, _ := c["用例标题"].(string)
		p, _ := c["优先级"].(string)
		rid, _ := c["需求ID"].(string)
		bdd, _ := c["BDD场景"].(string)
		parts = append(parts, fmt.Sprintf("TC%03d [%s] %s req=%s bdd=%s", i+1, p, title, rid, bdd))
	}
	return strings.Join(parts, "\n")
}

func localMatchMRToTC(changed []MRAPIItem, cases []map[string]any) int {
	if len(changed) == 0 || len(cases) == 0 {
		return 0
	}
	var corpus strings.Builder
	for _, c := range cases {
		b, _ := json.Marshal(c)
		corpus.Write(b)
	}
	all := corpus.String()
	n := 0
	for _, api := range changed {
		path := strings.TrimSpace(api.Path)
		if path == "" {
			continue
		}
		if strings.Contains(all, path) {
			n++
		}
	}
	return n
}
