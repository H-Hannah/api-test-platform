package docsrepo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadedTestCases 从 Git test-docs 加载的用例 JSON。
type LoadedTestCases struct {
	Version         string `json:"version"`
	RequirementID   string `json:"requirement_id"`
	RequirementName string `json:"requirement_name"`
	JSONPath        string `json:"json_path"`
	CaseCount       int    `json:"case_count"`
	CasesJSON       string `json:"cases_json"`
	Summary         string `json:"summary"`
}

// LoadTestCases 读取 test-docs/<version>/<requirement_id>/*-测试用例.json。
func LoadTestCases(repoRoot, version, requirementID string) (*LoadedTestCases, error) {
	root, err := filepath.Abs(strings.TrimSpace(repoRoot))
	if err != nil {
		return nil, err
	}
	if root == "" {
		return nil, fmt.Errorf("DOCS_REPO_ROOT 未配置")
	}
	ver := strings.TrimSpace(version)
	rid := strings.TrimSpace(requirementID)
	if ver == "" || rid == "" {
		return nil, fmt.Errorf("version and requirement_id required")
	}
	dir := filepath.Join(root, "test-docs", ver, rid)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if strings.Contains(root, "/path/to/") || !dirExists(root) {
			return nil, fmt.Errorf("test-docs 目录不存在: test-docs/%s/%s（DOCS_REPO_ROOT=%q 无效，请改为本机 qa-doc-generator 的绝对路径后重启服务）: %w",
				ver, rid, repoRoot, err)
		}
		return nil, fmt.Errorf("test-docs 目录不存在: %s (%w)", filepath.Join("test-docs", ver, rid), err)
	}
	var jsonFile string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".json") && strings.Contains(name, "测试用例") {
			jsonFile = filepath.Join(dir, name)
			break
		}
	}
	if jsonFile == "" {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				jsonFile = filepath.Join(dir, e.Name())
				break
			}
		}
	}
	if jsonFile == "" {
		return nil, fmt.Errorf("未找到测试用例 JSON: test-docs/%s/%s", ver, rid)
	}
	b, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}
	var cases []map[string]any
	if err := json.Unmarshal(b, &cases); err != nil {
		return nil, fmt.Errorf("parse test cases json: %w", err)
	}
	rel, _ := filepath.Rel(root, jsonFile)
	summary := summarizeCases(cases)
	meta, _ := LoadPackageMeta(root, filepath.Join("requirements", ver, rid))
	name := meta.RequirementName
	if name == "" {
		name = rid
	}
	return &LoadedTestCases{
		Version:         ver,
		RequirementID:   rid,
		RequirementName: name,
		JSONPath:        rel,
		CaseCount:       len(cases),
		CasesJSON:       string(b),
		Summary:         summary,
	}, nil
}

func summarizeCases(cases []map[string]any) string {
	var b strings.Builder
	pri := map[string]int{}
	for i, c := range cases {
		if i >= 40 {
			b.WriteString(fmt.Sprintf("\n... 另有 %d 条", len(cases)-40))
			break
		}
		title, _ := c["用例标题"].(string)
		p, _ := c["优先级"].(string)
		rid, _ := c["需求ID"].(string)
		pri[p]++
		b.WriteString(fmt.Sprintf("- [%s] %s (需求:%s)\n", p, title, rid))
	}
	b.WriteString(fmt.Sprintf("\n优先级: P0=%d P1=%d P2=%d P3=%d", pri["P0"], pri["P1"], pri["P2"], pri["P3"]))
	return b.String()
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}
