package impact

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	githubPRRe = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/pull/(\d+)`)
	httpClient = &http.Client{Timeout: 45 * time.Second}
)

type prFile struct {
	Path   string
	Status string
	Patch  string
}

// ResolveChangedFiles 从 GitLab MR、GitHub PR、口述变更或文件列表解析变更。
func ResolveChangedFiles(req AnalyzeRequest) ([]prFile, string, error) {
	gitlabURL := strings.TrimSpace(req.GitLabMRURL)
	if gitlabURL != "" {
		files, err := fetchGitLabMRFiles(gitlabURL)
		return files, "gitlab_mr", err
	}
	if desc := strings.TrimSpace(req.ChangeDescription); desc != "" {
		return []prFile{{Path: "口述变更", Status: "description", Patch: desc}}, "description", nil
	}
	prURL := strings.TrimSpace(req.GitHubPRURL)
	if prURL != "" {
		files, err := fetchGitHubPRFiles(prURL)
		return files, "github_pr", err
	}
	repo := strings.TrimSpace(req.Repo)
	base := strings.TrimSpace(req.BaseRef)
	head := strings.TrimSpace(req.HeadRef)
	if repo != "" && base != "" && head != "" {
		files, err := fetchGitHubCompare(repo, base, head)
		return files, "github_compare", err
	}
	if len(req.ChangedFiles) > 0 {
		var out []prFile
		for _, f := range req.ChangedFiles {
			f = strings.TrimSpace(f)
			if f != "" {
				out = append(out, prFile{Path: f, Status: "manual"})
			}
		}
		if len(out) == 0 {
			return nil, "manual", fmt.Errorf("changed_files empty")
		}
		return out, "manual", nil
	}
	return nil, "", fmt.Errorf("需提供 gitlab_mr_url、change_description 或 changed_files 之一")
}

func fetchGitHubPRFiles(prURL string) ([]prFile, error) {
	m := githubPRRe.FindStringSubmatch(prURL)
	if len(m) < 4 {
		return nil, fmt.Errorf("无法解析 GitHub PR 链接: %s", prURL)
	}
	owner, repo, numStr := m[1], m[2], m[3]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PR number: %s", numStr)
	}
	token := githubToken()
	var all []prFile
	page := 1
	for {
		u := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/files?per_page=100&page=%d", owner, repo, num, page)
		body, more, err := githubGET(u, token)
		if err != nil {
			return nil, err
		}
		var items []struct {
			Filename string `json:"filename"`
			Status   string `json:"status"`
			Patch    string `json:"patch"`
		}
		if err := json.Unmarshal(body, &items); err != nil {
			return nil, fmt.Errorf("parse PR files: %w", err)
		}
		for _, it := range items {
			all = append(all, prFile{Path: it.Filename, Status: it.Status, Patch: it.Patch})
		}
		if !more || len(items) == 0 {
			break
		}
		page++
		if page > 10 {
			break
		}
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("PR 无变更文件或无权访问（请配置 GITHUB_TOKEN）")
	}
	return all, nil
}

func fetchGitHubCompare(repo, base, head string) ([]prFile, error) {
	token := githubToken()
	u := fmt.Sprintf("https://api.github.com/repos/%s/compare/%s...%s", repo, base, head)
	body, _, err := githubGET(u, token)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Files []struct {
			Filename string `json:"filename"`
			Status   string `json:"status"`
			Patch    string `json:"patch"`
		} `json:"files"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	var out []prFile
	for _, f := range resp.Files {
		out = append(out, prFile{Path: f.Filename, Status: f.Status, Patch: f.Patch})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("compare 无变更文件")
	}
	return out, nil
}

func githubGET(u, token string) ([]byte, bool, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", "api-test-platform")
	req.Header.Set("Accept", "application/vnd.github+json")
	setGitHubAuth(req, token)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode != 200 {
		return nil, false, fmt.Errorf("GitHub API %s: HTTP %d %s", u, resp.StatusCode, strings.TrimSpace(string(b)))
	}
	more := strings.Contains(resp.Header.Get("Link"), `rel="next"`)
	return b, more, nil
}

func setGitHubAuth(req *http.Request, token string) {
	if token == "" {
		return
	}
	if strings.HasPrefix(token, "ghp_") || strings.HasPrefix(token, "github_pat_") || strings.HasPrefix(token, "gho_") {
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		req.Header.Set("Authorization", "token "+token)
	}
}

func githubToken() string {
	return strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
}
