package docsrepo

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GitLabDocsSettings qa-doc-generator 在 GitLab 上的定位。
type GitLabDocsSettings struct {
	BaseURL     string
	ProjectPath string
	Ref         string
}

// TestDocsCatalog 测试用例目录（版本或需求列表）。
type TestDocsCatalog struct {
	Source       string   `json:"source"`
	Ref          string   `json:"ref,omitempty"`
	Project      string   `json:"project,omitempty"`
	Versions     []string `json:"versions,omitempty"`
	Requirements []string `json:"requirements,omitempty"`
}

func GitLabDocsSettingsFromEnv() GitLabDocsSettings {
	base := strings.TrimSpace(os.Getenv("GITLAB_BASE_URL"))
	if base == "" {
		base = "https://gitlab.com"
	}
	ref := strings.TrimSpace(os.Getenv("GITLAB_DOCS_REF"))
	if ref == "" {
		ref = strings.TrimSpace(os.Getenv("GITLAB_DOCS_BRANCH"))
	}
	if ref == "" {
		ref = "main"
	}
	return GitLabDocsSettings{
		BaseURL:     strings.TrimRight(base, "/"),
		ProjectPath: strings.Trim(strings.TrimSpace(os.Getenv("GITLAB_DOCS_PROJECT")), "/"),
		Ref:         ref,
	}
}

func (c GitLabDocsSettings) OK() bool {
	return c.ProjectPath != ""
}

// ListTestDocsCatalog version 为空时列出版本；否则列出该版本下的 requirement_id。
func ListTestDocsCatalog(version, ref string) (*TestDocsCatalog, error) {
	cfg := GitLabDocsSettingsFromEnv()
	if cfg.OK() {
		if strings.TrimSpace(ref) == "" {
			return nil, fmt.Errorf("请先选择分支")
		}
		return listTestDocsFromGitLab(cfg, ref, version)
	}
	root := strings.TrimSpace(os.Getenv("DOCS_REPO_ROOT"))
	if root == "" {
		return nil, fmt.Errorf("请配置 GITLAB_DOCS_PROJECT（推荐）或 DOCS_REPO_ROOT")
	}
	return listTestDocsFromLocal(root, version)
}

func listTestDocsFromGitLab(cfg GitLabDocsSettings, ref, version string) (*TestDocsCatalog, error) {
	token := gitLabToken()
	treePath := "test-docs"
	if v := strings.TrimSpace(version); v != "" {
		treePath = "test-docs/" + v
	}
	names, err := gitLabListTreeDirs(cfg, ref, treePath, token)
	if err != nil {
		return nil, err
	}
	out := &TestDocsCatalog{
		Source:  "gitlab",
		Ref:     ref,
		Project: cfg.ProjectPath,
	}
	if version == "" {
		out.Versions = sortVersions(names)
	} else {
		out.Requirements = sortStrings(names)
	}
	return out, nil
}

func listTestDocsFromLocal(root, version string) (*TestDocsCatalog, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(root, "test-docs")
	if v := strings.TrimSpace(version); v != "" {
		dir = filepath.Join(dir, v)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取本地 test-docs 失败: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	out := &TestDocsCatalog{Source: "local"}
	if version == "" {
		out.Versions = sortVersions(names)
	} else {
		out.Requirements = sortStrings(names)
	}
	return out, nil
}

// LoadTestCasesBySelector 从 qa-doc-generator 的 test-docs/<version>/<requirement_id> 加载。
func LoadTestCasesBySelector(version, requirementID, ref string) (*LoadedTestCases, error) {
	ver := strings.TrimSpace(version)
	rid := strings.TrimSpace(requirementID)
	if ver == "" || rid == "" {
		return nil, fmt.Errorf("version and requirement_id required")
	}
	cfg := GitLabDocsSettingsFromEnv()
	if cfg.OK() {
		if strings.TrimSpace(ref) == "" {
			return nil, fmt.Errorf("请先选择分支")
		}
		repoPath := fmt.Sprintf("test-docs/%s/%s", ver, rid)
		return loadTestCasesFromGitLabTree(cfg.BaseURL, cfg.ProjectPath, ref, repoPath)
	}
	root := strings.TrimSpace(os.Getenv("DOCS_REPO_ROOT"))
	if root != "" {
		return LoadTestCases(root, ver, rid)
	}
	return nil, fmt.Errorf("请配置 GITLAB_DOCS_PROJECT 或 DOCS_REPO_ROOT")
}

func loadTestCasesFromGitLabTree(baseURL, projectPath, ref, repoPath string) (*LoadedTestCases, error) {
	glRef := &gitLabRef{
		BaseURL:     strings.TrimRight(baseURL, "/"),
		ProjectPath: projectPath,
		Ref:         ref,
		RepoPath:    strings.Trim(repoPath, "/"),
		IsFile:      false,
	}
	token := gitLabToken()
	jsonPath, content, err := gitLabFindTestCaseJSON(glRef, token)
	if err != nil {
		return nil, err
	}
	var cases []map[string]any
	if err := json.Unmarshal(content, &cases); err != nil {
		return nil, fmt.Errorf("parse test cases json from GitLab: %w", err)
	}
	ver, rid := parseTestDocsMeta(glRef.RepoPath)
	name := rid
	return &LoadedTestCases{
		Version:         ver,
		RequirementID:   rid,
		RequirementName: name,
		JSONPath:        jsonPath + " @ " + ref + " (GitLab)",
		CaseCount:       len(cases),
		CasesJSON:       string(content),
		Summary:         summarizeCases(cases),
	}, nil
}

func gitLabListTreeDirs(cfg GitLabDocsSettings, ref, treePath, token string) ([]string, error) {
	encodedProj := url.PathEscape(strings.Trim(cfg.ProjectPath, "/"))
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/repository/tree?path=%s&ref=%s&per_page=100",
		strings.TrimRight(cfg.BaseURL, "/"),
		encodedProj,
		url.QueryEscape(treePath),
		url.QueryEscape(ref),
	)
	body, err := gitLabGET(apiURL, token)
	if err != nil {
		return nil, err
	}
	var items []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("parse GitLab tree: %w", err)
	}
	var dirs []string
	for _, it := range items {
		if it.Type == "tree" {
			dirs = append(dirs, it.Name)
		}
	}
	return dirs, nil
}

func sortStrings(ss []string) []string {
	out := append([]string(nil), ss...)
	sort.Strings(out)
	return out
}

func sortVersions(ss []string) []string {
	out := append([]string(nil), ss...)
	sort.Slice(out, func(i, j int) bool {
		return compareVersion(out[i], out[j]) > 0
	})
	return out
}

func compareVersion(a, b string) int {
	pa := parseVersionParts(a)
	pb := parseVersionParts(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			return pa[i] - pb[i]
		}
	}
	return strings.Compare(a, b)
}

func parseVersionParts(v string) [3]int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.Split(v, ".")
	var out [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		var n int
		fmt.Sscanf(parts[i], "%d", &n)
		out[i] = n
	}
	return out
}
