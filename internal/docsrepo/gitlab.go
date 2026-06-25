package docsrepo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	gitlabTreeRe = regexp.MustCompile(`(?i)^(https?)://([^/]+)/(.+?)/-/tree/([^/]+)/(.+?)(?:\?.*)?$`)
	gitlabBlobRe = regexp.MustCompile(`(?i)^(https?)://([^/]+)/(.+?)/-/blob/([^/]+)/(.+?)(?:\?.*)?$`)
	gitlabHTTP   = &http.Client{Timeout: 45 * time.Second}
)

type gitLabRef struct {
	BaseURL     string
	ProjectPath string
	Ref         string
	RepoPath    string // tree 目录或 blob 文件在仓库内的路径
	IsFile      bool
}

// LoadTestCasesFromGitLabURL 从 GitLab 网页 tree/blob 链接加载 test-docs 用例 JSON。
// 例: https://gitlab.com/group/proj/-/tree/branch/test-docs/v2.7.2/brief
func LoadTestCasesFromGitLabURL(rawURL string) (*LoadedTestCases, error) {
	ref, err := parseGitLabRepoURL(rawURL)
	if err != nil {
		return nil, err
	}
	if ref.IsFile {
		token := gitLabToken()
		content, err := gitLabFetchRaw(ref, ref.RepoPath, token)
		if err != nil {
			return nil, err
		}
		return loadedFromGitLabJSON(ref, ref.RepoPath, content)
	}
	return loadTestCasesFromGitLabTree(ref.BaseURL, ref.ProjectPath, ref.Ref, ref.RepoPath)
}

func loadedFromGitLabJSON(ref *gitLabRef, jsonPath string, content []byte) (*LoadedTestCases, error) {
	var cases []map[string]any
	if err := json.Unmarshal(content, &cases); err != nil {
		return nil, fmt.Errorf("parse test cases json from GitLab: %w", err)
	}
	dir := jsonPath
	if i := strings.LastIndex(dir, "/"); i >= 0 {
		dir = dir[:i]
	}
	ver, rid := parseTestDocsMeta(dir)
	name := rid
	return &LoadedTestCases{
		Version:         ver,
		RequirementID:   rid,
		RequirementName: name,
		JSONPath:        jsonPath + " @ " + ref.Ref + " (GitLab)",
		CaseCount:       len(cases),
		CasesJSON:       string(content),
		Summary:         summarizeCases(cases),
	}, nil
}

func parseGitLabRepoURL(raw string) (*gitLabRef, error) {
	raw = strings.TrimSpace(raw)
	if m := gitlabTreeRe.FindStringSubmatch(raw); len(m) >= 6 {
		return &gitLabRef{
			BaseURL:     m[1] + "://" + m[2],
			ProjectPath: decodeGitLabPath(m[3]),
			Ref:         m[4],
			RepoPath:    strings.Trim(decodeGitLabPath(m[5]), "/"),
			IsFile:      false,
		}, nil
	}
	if m := gitlabBlobRe.FindStringSubmatch(raw); len(m) >= 6 {
		path := strings.Trim(decodeGitLabPath(m[5]), "/")
		if !strings.HasSuffix(strings.ToLower(path), ".json") {
			return nil, fmt.Errorf("blob 链接须指向 .json 文件，或改用 /-/tree/ 目录链接")
		}
		return &gitLabRef{
			BaseURL:     m[1] + "://" + m[2],
			ProjectPath: decodeGitLabPath(m[3]),
			Ref:         m[4],
			RepoPath:    path,
			IsFile:      true,
		}, nil
	}
	return nil, fmt.Errorf("无法解析 GitLab 链接（支持 /-/tree/<ref>/<path> 或 /-/blob/<ref>/<file.json>）")
}

func gitLabFindTestCaseJSON(ref *gitLabRef, token string) (string, []byte, error) {
	encodedProj := url.PathEscape(strings.Trim(ref.ProjectPath, "/"))
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/repository/tree?path=%s&ref=%s&per_page=100",
		strings.TrimRight(ref.BaseURL, "/"),
		encodedProj,
		url.QueryEscape(ref.RepoPath),
		url.QueryEscape(ref.Ref),
	)
	body, err := gitLabGET(apiURL, token)
	if err != nil {
		return "", nil, err
	}
	var items []struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &items); err != nil {
		return "", nil, fmt.Errorf("parse GitLab tree: %w", err)
	}
	var jsonPath string
	for _, it := range items {
		if it.Type != "blob" {
			continue
		}
		if strings.HasSuffix(it.Name, ".json") && strings.Contains(it.Name, "测试用例") {
			jsonPath = it.Path
			break
		}
	}
	if jsonPath == "" {
		for _, it := range items {
			if it.Type == "blob" && strings.HasSuffix(it.Name, ".json") {
				jsonPath = it.Path
				break
			}
		}
	}
	if jsonPath == "" {
		return "", nil, fmt.Errorf("目录 %s 下未找到测试用例 JSON（ref=%s）", ref.RepoPath, ref.Ref)
	}
	content, err := gitLabFetchRaw(ref, jsonPath, token)
	return jsonPath, content, err
}

func gitLabFetchRaw(ref *gitLabRef, filePath, token string) ([]byte, error) {
	encodedProj := url.PathEscape(strings.Trim(ref.ProjectPath, "/"))
	encodedFile := url.PathEscape(filePath)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/repository/files/%s/raw?ref=%s",
		strings.TrimRight(ref.BaseURL, "/"),
		encodedProj,
		encodedFile,
		url.QueryEscape(ref.Ref),
	)
	return gitLabGET(apiURL, token)
}

func gitLabGET(apiURL, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "api-test-platform")
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}
	resp, err := gitlabHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode != 200 {
		msg := strings.TrimSpace(string(b))
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			return nil, fmt.Errorf("GitLab 鉴权失败 HTTP %d（请配置 GITLAB_TOKEN，权限 read_api/read_repository）: %s", resp.StatusCode, msg)
		}
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("GitLab 资源未找到 HTTP 404（检查分支/ref 与路径）: %s", msg)
		}
		return nil, fmt.Errorf("GitLab API HTTP %d: %s", resp.StatusCode, msg)
	}
	return b, nil
}

func gitLabToken() string {
	if t := strings.TrimSpace(os.Getenv("GITLAB_TOKEN")); t != "" {
		return t
	}
	return strings.TrimSpace(os.Getenv("PRIVATE_TOKEN"))
}

func decodeGitLabPath(p string) string {
	if dec, err := url.PathUnescape(p); err == nil {
		return dec
	}
	return p
}

func parseTestDocsMeta(dirPath string) (version, requirementID string) {
	parts := strings.Split(strings.Trim(strings.TrimSpace(dirPath), "/"), "/")
	if len(parts) >= 3 && parts[0] == "test-docs" {
		return parts[1], parts[2]
	}
	return "", ""
}
