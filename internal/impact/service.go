package impact

import (
	"fmt"
	"strings"

	"api-test-platform/internal/store"
)

type AICompleter interface {
	Complete(prompt string) (string, error)
}

type Service struct {
	store *store.Store
	repo  string
	ai    AICompleter
}

func NewService(st *store.Store, docsRepoRoot string, ai AICompleter) *Service {
	return &Service{store: st, repo: docsRepoRoot, ai: ai}
}

func (s *Service) Analyze(req AnalyzeRequest) (*AnalyzeResult, error) {
	if req.ProductID < 0 {
		return nil, fmt.Errorf("invalid product_id")
	}

	files, source, err := ResolveChangedFiles(req)
	if err != nil {
		return nil, err
	}

	casesJSON := strings.TrimSpace(req.CasesJSON)
	if casesJSON == "" {
		return nil, fmt.Errorf("cases_json required：请先从 GitLab 链接加载测试用例")
	}

	tcRows, err := parseTestCases(casesJSON)
	if err != nil {
		return nil, err
	}

	inferred := ExtractInferredAPIs(files, req.ChangedAPIs)
	changed := toChangedFiles(files)

	platformAPIs, err := s.loadPlatformAPIs(req)
	if err != nil {
		return nil, err
	}

	recTCs := scoreTestCases(tcRows, files, inferred, minRecommendTCScore)
	recAPIs := scoreAPIs(platformAPIs, files, inferred, 0.25)

	scenarioInputs, err := s.loadScenarioAPIRefs(req.ProductID)
	if err != nil {
		return nil, err
	}
	recScenarios := scoreScenarios(scenarioInputs, recAPIs)

	plan := buildRunPlan(recTCs, recAPIs, recScenarios)
	gaps := buildGaps(files, inferred, recTCs, platformAPIs)

	ready := 0
	for _, a := range recAPIs {
		if a.ScenarioReady {
			ready++
		}
	}

	result := &AnalyzeResult{
		Source:               source,
		ChangedFiles:         changed,
		InferredAPIs:         inferred,
		RecommendedTCs:      recTCs,
		RecommendedTCReason: buildRecommendedTCReason(source, inferred, recTCs),
		RecommendedAPIs:     recAPIs,
		RecommendedScenarios: recScenarios,
		Gaps:                 gaps,
		RunPlan:              plan,
		Summary:              buildSummary(changed, inferred, recTCs, recAPIs),
		Stats: ImpactStats{
			ChangedFileCount: len(changed),
			InferredAPICount: len(inferred),
			RecommendedTC:    len(recTCs),
			RecommendedAPI:     len(recAPIs),
			PlatformReadyAPI:   ready,
		},
	}

	if req.UseAI && s.ai != nil {
		if summary, err := s.aiSummary(req, files, result); err == nil && summary != "" {
			result.AISummary = summary
		}
	}
	return result, nil
}

func (s *Service) PostMRComment(req PostMRCommentRequest) (*PostMRCommentResult, error) {
	mrURL := strings.TrimSpace(req.GitLabMRURL)
	if mrURL == "" {
		return nil, fmt.Errorf("gitlab_mr_url required")
	}
	if req.Result == nil {
		return nil, fmt.Errorf("result required：请先完成精准分析")
	}
	md := buildMRCommentMarkdown(req)
	noteID, err := PostGitLabMRNote(mrURL, md)
	if err != nil {
		return nil, err
	}
	return &PostMRCommentResult{
		NoteID:   noteID,
		NoteURL:  MRNoteWebURL(mrURL, noteID),
		Markdown: md,
	}, nil
}

func (s *Service) PreviewMRComment(req PostMRCommentRequest) (*PostMRCommentResult, error) {
	if strings.TrimSpace(req.GitLabMRURL) == "" {
		return nil, fmt.Errorf("gitlab_mr_url required")
	}
	if req.Result == nil {
		return nil, fmt.Errorf("result required：请先完成精准分析")
	}
	md := buildMRCommentMarkdown(req)
	return &PostMRCommentResult{Markdown: md}, nil
}

func buildMRCommentMarkdown(req PostMRCommentRequest) string {
	ctx := MRCommentContext{
		GitLabMRURL:   req.GitLabMRURL,
		Version:       req.Version,
		RequirementID: req.RequirementID,
		TCDocsBranch:  req.TCDocsBranch,
	}
	return BuildMRCommentMarkdown(ctx, req.Result)
}

func (s *Service) loadPlatformAPIs(req AnalyzeRequest) ([]apiRow, error) {
	filter := store.APIListFilter{ProductID: req.ProductID}
	if mr := strings.TrimSpace(req.MRTag); mr != "" {
		filter.MRTag = mr
	}
	list, err := s.store.ListAPIsFiltered(filter)
	if err != nil {
		return nil, err
	}
	// MR 标签过滤结果为空时，回退全量产品接口以便推荐
	if len(list) == 0 && filter.MRTag != "" {
		list, err = s.store.ListAPIsFiltered(store.APIListFilter{ProductID: req.ProductID})
		if err != nil {
			return nil, err
		}
	}
	rows := make([]apiRow, 0, len(list))
	for _, a := range list {
		rows = append(rows, apiRow{
			ID: a.ID, Method: a.Method, Path: a.Path, Name: a.Name,
			FolderPath: a.FolderPath, ScenarioReady: a.ScenarioReady, TCRef: a.TCRef,
		})
	}
	return rows, nil
}

func (s *Service) loadScenarioAPIRefs(productID int64) ([]struct {
	ID     int64
	Name   string
	APIIDs []int64
}, error) {
	list, err := s.store.ListScenarios(productID)
	if err != nil {
		return nil, err
	}
	var out []struct {
		ID     int64
		Name   string
		APIIDs []int64
	}
	for _, sc := range list {
		full, err := s.store.GetScenario(sc.ID)
		if err != nil {
			continue
		}
		var ids []int64
		for _, st := range full.Steps {
			if st.APIID != nil && *st.APIID > 0 {
				ids = append(ids, *st.APIID)
			}
		}
		out = append(out, struct {
			ID     int64
			Name   string
			APIIDs []int64
		}{ID: sc.ID, Name: sc.Name, APIIDs: ids})
	}
	return out, nil
}

func buildRecommendedTCReason(source string, inferred []string, tcs []RecommendedTC) string {
	if len(tcs) == 0 {
		return "无得分>1的强相关用例"
	}
	var parts []string
	switch source {
	case "description":
		parts = append(parts, "与口述变更中的接口路径、关键词匹配")
	case "gitlab_mr":
		parts = append(parts, "与 MR 代码 diff 中的变更文件、推断接口匹配")
	default:
		parts = append(parts, "与本次变更推断的模块/接口匹配")
	}
	if len(inferred) > 0 {
		show := inferred
		if len(show) > 4 {
			show = append(append([]string{}, show[:4]...), "…")
		}
		parts = append(parts, "涉及接口："+strings.Join(show, "、"))
	}
	p0, p1 := 0, 0
	for _, tc := range tcs {
		switch strings.ToUpper(tc.Priority) {
		case "P0":
			p0++
		case "P1":
			p1++
		}
	}
	if p0+p1 > 0 {
		parts = append(parts, fmt.Sprintf("含 P0 %d 条、P1 %d 条", p0, p1))
	}
	parts = append(parts, fmt.Sprintf("共推荐 %d 条（相关分>1）", len(tcs)))
	return strings.Join(parts, "；") + "。"
}

func buildSummary(changed []ChangedFile, inferred []string, tcs []RecommendedTC, apis []RecommendedAPI) string {
	if len(changed) == 1 && changed[0].Status == "description" {
		return fmt.Sprintf("口述变更；推断 %d 个接口；推荐 %d 条用例、%d 个平台接口",
			len(inferred), len(tcs), len(apis))
	}
	return fmt.Sprintf("变更 %d 个文件，推断 %d 个接口；推荐执行 %d 条用例、%d 个平台接口（其中场景就绪 %d）",
		len(changed), len(inferred), len(tcs), len(apis), countReady(apis))
}

func countReady(apis []RecommendedAPI) int {
	n := 0
	for _, a := range apis {
		if a.ScenarioReady {
			n++
		}
	}
	return n
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
