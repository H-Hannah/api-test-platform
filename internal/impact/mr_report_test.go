package impact

import (
	"strings"
	"testing"
)

func TestBuildMRCommentMarkdown(t *testing.T) {
	res := &AnalyzeResult{
		Source:              "gitlab_mr",
		Summary:             "变更 3 个文件",
		AISummary:           "本次修改了绑定逻辑。",
		RecommendedTCReason: "与 MR 匹配；共推荐 1 条。",
		RecommendedTCs: []RecommendedTC{
			{TCID: "TC001", Title: "查询绑定", Priority: "P0", Score: 1.82},
		},
		RecommendedAPIs: []RecommendedAPI{
			{Method: "GET", Path: "/v2/bindings", Name: "查询绑定", ScenarioReady: true, Score: 2.1},
		},
		ChangedFiles: []ChangedFile{
			{Path: "internal/bind/handler.go", Status: "modified"},
		},
	}
	md := BuildMRCommentMarkdown(MRCommentContext{
		Version: "v2.7.2", RequirementID: "brief", TCDocsBranch: "beta_x",
	}, res)
	for _, want := range []string{"精准测试报告", "TC001", "查询绑定", "变更解读", "推荐接口", "/v2/bindings", "变更文件", "handler.go"} {
		if !strings.Contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}
