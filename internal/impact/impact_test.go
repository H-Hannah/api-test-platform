package impact

import (
	"strings"
	"testing"
)

func TestParseDiffFiles(t *testing.T) {
	diff := `diff --git a/internal/foo.go b/internal/foo.go
+++ b/internal/foo.go
@@ -1 +1 @@
+GET "/v2/platform/bindings"
`
	files := parseDiffFiles(diff)
	if len(files) != 1 || files[0].Path != "internal/foo.go" {
		t.Fatalf("unexpected files: %+v", files)
	}
}

func TestExtractInferredAPIs(t *testing.T) {
	files := []prFile{{
		Path:  "routes.go",
		Patch: `router.GET("/v2/platform/bind/TWITTER", handler)`,
	}}
	apis := ExtractInferredAPIs(files, []APIItem{{Method: "GET", Path: "/v2/foo"}})
	if len(apis) < 2 {
		t.Fatalf("expected >=2 apis, got %v", apis)
	}
}

func TestParseGitLabMRURL(t *testing.T) {
	base, proj, iid, err := parseGitLabMRURL("https://gitlab.com/my-group/my-project/-/merge_requests/42")
	if err != nil {
		t.Fatal(err)
	}
	if base != "https://gitlab.com" || proj != "my-group/my-project" || iid != 42 {
		t.Fatalf("unexpected: %s %s %d", base, proj, iid)
	}
}

func TestBuildChangeDigestForAI(t *testing.T) {
	files := []prFile{
		{Path: "routes.go", Status: "modified", Patch: "+router.GET(\"/v2/foo\", h)\n"},
		{Path: "readme.md", Status: "added", Patch: ""},
	}
	digest := buildChangeDigestForAI(files)
	if !strings.Contains(digest, "routes.go") || !strings.Contains(digest, "/v2/foo") {
		t.Fatalf("unexpected digest: %s", digest)
	}
}

func TestScoreFromDescription(t *testing.T) {
	cases := `[{"用例标题":"Redis配置校验","优先级":"P0","模块":["配置"],"步骤":[{"操作":"GET /v2/platform/bind/TWITTER","预期":"200"}]}]`
	rows, err := parseTestCases(cases)
	if err != nil {
		t.Fatal(err)
	}
	files := []prFile{{Path: "口述变更", Status: "description", Patch: "改了 Redis 连接配置和 /v2/platform/bind/TWITTER 超时"}}
	inferred := ExtractInferredAPIs(files, nil)
	rec := scoreTestCases(rows, files, inferred, minRecommendTCScore)
	if len(rec) == 0 {
		t.Fatal("expected description-based recommendation with score > 1")
	}
}

func TestScoreTestCases(t *testing.T) {
	cases := `[{"用例标题":"查询Twitter绑定","优先级":"P0","模块":["平台","绑定"],"步骤":[{"操作":"GET /v2/platform/bind/TWITTER","预期":"200"}]}]`
	rows, err := parseTestCases(cases)
	if err != nil {
		t.Fatal(err)
	}
	files := []prFile{
		{Path: "internal/platform/bind_handler.go"},
		{Path: "internal/platform/bind_service.go"},
		{Path: "pkg/platform/bindings.go"},
	}
	inferred := []string{"GET /v2/platform/bind/TWITTER"}
	rec := scoreTestCases(rows, files, inferred, minRecommendTCScore)
	if len(rec) == 0 {
		t.Fatal("expected recommendations")
	}
}
