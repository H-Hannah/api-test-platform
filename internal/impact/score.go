package impact

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// minRecommendTCScore 推荐用例最低相关分（仅列出得分 > 1 的用例）。
const minRecommendTCScore = 1.0

type tcRow struct {
	Index    int
	ID       string
	Title    string
	Priority string
	Modules  []string
	BDD      string
	Corpus   string
}

func parseTestCases(casesJSON string) ([]tcRow, error) {
	var cases []map[string]any
	if err := json.Unmarshal([]byte(casesJSON), &cases); err != nil {
		return nil, fmt.Errorf("parse cases json: %w", err)
	}
	rows := make([]tcRow, 0, len(cases))
	for i, c := range cases {
		title, _ := c["用例标题"].(string)
		pri, _ := c["优先级"].(string)
		bdd, _ := c["BDD场景"].(string)
		rid, _ := c["需求ID"].(string)
		var mods []string
		if raw, ok := c["模块"].([]any); ok {
			for _, m := range raw {
				mods = append(mods, fmt.Sprint(m))
			}
		}
		b, _ := json.Marshal(c)
		id := fmt.Sprintf("TC%03d", i+1)
		rows = append(rows, tcRow{
			Index: i + 1, ID: id, Title: title, Priority: pri,
			Modules: mods, BDD: bdd, Corpus: string(b) + " " + rid,
		})
	}
	return rows, nil
}

func scoreTestCases(rows []tcRow, files []prFile, inferred []string, minScore float64) []RecommendedTC {
	corpus := strings.ToLower(corpusFromFiles(files))
	var out []RecommendedTC
	for _, tc := range rows {
		score := 0.0
		var reasons []string
		low := strings.ToLower(tc.Corpus + " " + tc.Title + " " + tc.BDD)

		for _, inf := range inferred {
			_, p := parseAPIKey(inf)
			if p != "" && strings.Contains(low, strings.ToLower(p)) {
				score += 0.45
				reasons = append(reasons, "用例含接口 "+p)
			}
		}
		for _, f := range files {
			for _, kw := range fileKeywords(f.Path) {
				if strings.Contains(low, kw) {
					score += 0.12
					reasons = append(reasons, "模块关键词 "+kw+" ↔ 文件 "+f.Path)
					break
				}
			}
			if strings.Contains(corpus, strings.ToLower(f.Path)) {
				score += 0.05
			}
		}
		matchedKw := 0
		for _, kw := range keywordsFromDescription(corpus) {
			if matchedKw >= 4 {
				break
			}
			if strings.Contains(low, kw) {
				boost := 0.18
				if strings.Contains(strings.ToLower(tc.Title), kw) {
					boost = 0.42
				}
				score += boost
				reasons = append(reasons, "口述含「"+kw+"」")
				matchedKw++
			}
		}
		switch strings.ToUpper(strings.TrimSpace(tc.Priority)) {
		case "P0":
			score += 0.15
		case "P1":
			score += 0.08
		}
		if score <= minScore {
			continue
		}
		out = append(out, RecommendedTC{
			TCID: tc.ID, Title: tc.Title, Priority: tc.Priority,
			Score: roundScore(score), Reasons: dedupe(reasons),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}

type apiRow struct {
	ID            int64
	Method        string
	Path          string
	Name          string
	FolderPath    string
	ScenarioReady bool
	TCRef         string
}

func scoreAPIs(apis []apiRow, files []prFile, inferred []string, minScore float64) []RecommendedAPI {
	corpus := corpusFromFiles(files)
	var out []RecommendedAPI
	for _, api := range apis {
		score := 0.0
		var reasons []string
		full := strings.ToUpper(api.Method) + " " + api.Path

		for _, inf := range inferred {
			if pathMatches(inf, api.Path) {
				score += 0.55
				reasons = append(reasons, "命中变更接口 "+inf)
			}
		}
		if strings.Contains(corpus, api.Path) {
			score += 0.25
			reasons = append(reasons, "变更描述含 path "+api.Path)
		}
		lowName := strings.ToLower(api.FolderPath + " " + api.Name)
		for _, kw := range keywordsFromDescription(corpus) {
			if len(kw) >= 3 && strings.Contains(lowName, kw) {
				score += 0.12
				reasons = append(reasons, "口述含「"+kw+"」")
				break
			}
		}
		for _, f := range files {
			for _, kw := range fileKeywords(f.Path) {
				low := strings.ToLower(api.FolderPath + " " + api.Name + " " + api.Path)
				if strings.Contains(low, kw) {
					score += 0.1
					reasons = append(reasons, "目录/名称含 "+kw)
					break
				}
			}
		}
		if api.ScenarioReady {
			score += 0.1
			reasons = append(reasons, "场景就绪")
		}
		_ = full
		if score < minScore {
			continue
		}
		out = append(out, RecommendedAPI{
			APIID: api.ID, Method: api.Method, Path: api.Path, Name: api.Name,
			ScenarioReady: api.ScenarioReady, Score: roundScore(score), Reasons: dedupe(reasons),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}

func scoreScenarios(scenarios []struct {
	ID   int64
	Name string
	APIIDs []int64
}, recAPIs []RecommendedAPI) []RecommendedScenario {
	if len(recAPIs) == 0 {
		return nil
	}
	recSet := map[int64]float64{}
	for _, a := range recAPIs {
		recSet[a.APIID] = a.Score
	}
	var out []RecommendedScenario
	for _, sc := range scenarios {
		score := 0.0
		var reasons []string
		for _, aid := range sc.APIIDs {
			if s, ok := recSet[aid]; ok {
				score += s * 0.5
				reasons = append(reasons, fmt.Sprintf("含推荐接口 #%d", aid))
			}
		}
		if score < 0.25 {
			continue
		}
		out = append(out, RecommendedScenario{
			ScenarioID: sc.ID, Name: sc.Name, Score: roundScore(score), Reasons: dedupe(reasons),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}

func buildRunPlan(tcs []RecommendedTC, apis []RecommendedAPI, scenarios []RecommendedScenario) RunPlan {
	plan := RunPlan{}
	for _, tc := range tcs {
		switch strings.ToUpper(tc.Priority) {
		case "P0":
			plan.P0TCIDs = append(plan.P0TCIDs, tc.TCID)
		case "P1":
			plan.P1TCIDs = append(plan.P1TCIDs, tc.TCID)
		}
	}
	for _, a := range apis {
		plan.APIIDs = append(plan.APIIDs, a.APIID)
	}
	for _, s := range scenarios {
		plan.ScenarioIDs = append(plan.ScenarioIDs, s.ScenarioID)
	}
	return plan
}

func buildGaps(files []prFile, inferred []string, tcs []RecommendedTC, apis []apiRow) []ImpactGap {
	var gaps []ImpactGap
	tcCorpus := strings.Builder{}
	for _, tc := range tcs {
		tcCorpus.WriteString(strings.ToLower(tc.Title))
	}
	for _, inf := range inferred {
		_, p := parseAPIKey(inf)
		if p == "" {
			continue
		}
		if !strings.Contains(tcCorpus.String(), strings.ToLower(p)) {
			gaps = append(gaps, ImpactGap{
				Type: "code_no_tc", API: inf, Action: "补充覆盖该接口的用例",
			})
		}
		found := false
		for _, a := range apis {
			if pathMatches(inf, a.Path) {
				found = true
				break
			}
		}
		if !found {
			gaps = append(gaps, ImpactGap{
				Type: "platform_missing", API: inf, Action: "平台入库或录制",
			})
		}
	}
	for _, f := range files {
		if f.Status == "description" {
			low := strings.ToLower(f.Patch)
			if strings.Contains(low, "配置") || strings.Contains(low, "config") ||
				strings.Contains(low, "redis") || strings.Contains(low, "nacos") ||
				strings.Contains(low, "开关") || strings.Contains(low, "环境变量") {
				gaps = append(gaps, ImpactGap{
					Type: "config_change", File: "口述变更", Action: "核对配置类、环境依赖相关用例与数据集",
				})
			}
			continue
		}
		if strings.Contains(strings.ToLower(f.Path), "migration") || strings.HasSuffix(f.Path, ".sql") {
			gaps = append(gaps, ImpactGap{
				Type: "schema_change", File: f.Path, Action: "检查数据类用例与测试数据集",
			})
		}
	}
	return gaps
}

func dedupe(items []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range items {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func roundScore(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
