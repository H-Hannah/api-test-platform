package impact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// PostGitLabMRNote 在 MR 下创建评论（Markdown）。
func PostGitLabMRNote(mrURL, body string) (noteID int64, err error) {
	mrURL = strings.TrimSpace(mrURL)
	body = strings.TrimSpace(body)
	if mrURL == "" {
		return 0, fmt.Errorf("gitlab_mr_url required")
	}
	if body == "" {
		return 0, fmt.Errorf("comment body empty")
	}
	base, project, iid, err := parseGitLabMRURL(mrURL)
	if err != nil {
		return 0, err
	}
	token := gitlabToken()
	if token == "" {
		return 0, fmt.Errorf("GITLAB_TOKEN 未配置")
	}
	encoded := encodeGitLabProject(project)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d/notes",
		strings.TrimRight(base, "/"), encoded, iid)

	payload, _ := json.Marshal(map[string]string{"body": body})
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "api-test-platform")
	req.Header.Set("PRIVATE-TOKEN", token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		msg := strings.TrimSpace(string(b))
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			return 0, fmt.Errorf("GitLab 无权限发表评论 HTTP %d（Token 需 api 写权限）: %s", resp.StatusCode, msg)
		}
		return 0, fmt.Errorf("GitLab 发表评论失败 HTTP %d: %s", resp.StatusCode, msg)
	}
	var note struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(b, &note); err != nil {
		return 0, fmt.Errorf("parse note response: %w", err)
	}
	return note.ID, nil
}

// MRNoteWebURL 构造 MR 评论锚点链接。
func MRNoteWebURL(mrURL string, noteID int64) string {
	mrURL = strings.TrimSpace(mrURL)
	if mrURL == "" || noteID <= 0 {
		return mrURL
	}
	if i := strings.Index(mrURL, "#"); i >= 0 {
		mrURL = mrURL[:i]
	}
	return fmt.Sprintf("%s#note_%d", mrURL, noteID)
}
