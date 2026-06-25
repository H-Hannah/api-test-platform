package impact

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// gitlabMRRe 匹配 GitLab MR 链接（含自建实例）
// 例: https://gitlab.com/group/project/-/merge_requests/42
//     https://git.example.com/a/b/c/-/merge_requests/7
var gitlabMRRe = regexp.MustCompile(`(?i)^(https?)://([^/]+)/(.+?)/-/merge_requests/(\d+)(?:[/?#].*)?$`)

func fetchGitLabMRFiles(mrURL string) ([]prFile, error) {
	base, project, iid, err := parseGitLabMRURL(mrURL)
	if err != nil {
		return nil, err
	}
	token := gitlabToken()
	encoded := encodeGitLabProject(project)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d/changes", strings.TrimRight(base, "/"), encoded, iid)

	body, err := gitlabGET(apiURL, token)
	if err != nil {
		if token == "" {
			return nil, fmt.Errorf("%w（私有仓库请配置 GITLAB_TOKEN）", err)
		}
		return nil, err
	}

	var resp struct {
		Changes []struct {
			OldPath     string `json:"old_path"`
			NewPath     string `json:"new_path"`
			Diff        string `json:"diff"`
			NewFile     bool   `json:"new_file"`
			DeletedFile bool   `json:"deleted_file"`
			RenamedFile bool   `json:"renamed_file"`
		} `json:"changes"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse GitLab MR changes: %w", err)
	}
	if len(resp.Changes) == 0 {
		return nil, fmt.Errorf("MR !%d 无变更文件或无权访问", iid)
	}

	var out []prFile
	for _, c := range resp.Changes {
		path := c.NewPath
		status := "modified"
		if c.NewFile {
			status = "added"
		}
		if c.DeletedFile {
			status = "deleted"
			path = c.OldPath
		}
		if c.RenamedFile {
			status = "renamed"
		}
		if path == "" {
			path = c.OldPath
		}
		out = append(out, prFile{Path: path, Status: status, Patch: c.Diff})
	}
	return out, nil
}

type gitlabMRMeta struct {
	Title        string
	Description  string
	SourceBranch string
	TargetBranch string
}

func fetchGitLabMRMeta(mrURL string) (*gitlabMRMeta, error) {
	base, project, iid, err := parseGitLabMRURL(mrURL)
	if err != nil {
		return nil, err
	}
	token := gitlabToken()
	encoded := encodeGitLabProject(project)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d", strings.TrimRight(base, "/"), encoded, iid)
	body, err := gitlabGET(apiURL, token)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		SourceBranch string `json:"source_branch"`
		TargetBranch string `json:"target_branch"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse GitLab MR meta: %w", err)
	}
	return &gitlabMRMeta{
		Title:        resp.Title,
		Description:  resp.Description,
		SourceBranch: resp.SourceBranch,
		TargetBranch: resp.TargetBranch,
	}, nil
}

func parseGitLabMRURL(raw string) (baseURL, projectPath string, iid int, err error) {
	raw = strings.TrimSpace(raw)
	m := gitlabMRRe.FindStringSubmatch(raw)
	if len(m) < 5 {
		return "", "", 0, fmt.Errorf("无法解析 GitLab MR 链接: %s（期望格式: https://gitlab.example.com/group/project/-/merge_requests/123）", raw)
	}
	baseURL = m[1] + "://" + m[2]
	projectPath = strings.Trim(strings.TrimSpace(m[3]), "/")
	if dec, e := url.PathUnescape(projectPath); e == nil {
		projectPath = dec
	}
	iid, err = strconv.Atoi(m[4])
	if err != nil || iid <= 0 {
		return "", "", 0, fmt.Errorf("invalid MR iid: %s", m[4])
	}
	return baseURL, projectPath, iid, nil
}

func encodeGitLabProject(path string) string {
	return url.PathEscape(strings.Trim(path, "/"))
}

func gitlabGET(apiURL, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "api-test-platform")
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode != 200 {
		msg := strings.TrimSpace(string(b))
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			return nil, fmt.Errorf("GitLab API 鉴权失败 HTTP %d: %s", resp.StatusCode, msg)
		}
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("GitLab MR 未找到 HTTP 404（检查链接与 Token 权限 read_api/read_repository）")
		}
		return nil, fmt.Errorf("GitLab API %s: HTTP %d %s", apiURL, resp.StatusCode, msg)
	}
	return b, nil
}

func gitlabToken() string {
	if t := strings.TrimSpace(os.Getenv("GITLAB_TOKEN")); t != "" {
		return t
	}
	// 兼容部分团队命名
	return strings.TrimSpace(os.Getenv("PRIVATE_TOKEN"))
}
