package docsrepo

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 20 * time.Second}

// FetchURLResult 拉取远程文档的结果。
type FetchURLResult struct {
	URL        string `json:"url"`
	RawURL     string `json:"raw_url"`
	Content    string `json:"content"`
	Truncated  bool   `json:"truncated"`
}

// MaxFetchBytes 单文件最大拉取字节（约 400KB）。
const MaxFetchBytes = 400_000

// FetchDocURL 拉取远程文档内容，支持 GitHub blob/raw、Figma（只返回链接元数据）。
func FetchDocURL(rawURL string) (*FetchURLResult, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("url required")
	}
	// Figma 链接：不拉取，直接返回 URL（在 prompt 里作为设计稿引用）
	if strings.Contains(rawURL, "figma.com") {
		return &FetchURLResult{URL: rawURL, RawURL: rawURL, Content: ""}, nil
	}

	// GitHub blob 链接：若配置了 token，优先走 GitHub API（支持 private repo）。
	if m := githubBlobRe.FindStringSubmatch(rawURL); m != nil {
		token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
		if token != "" {
			ownerRepo := m[1]
			ref := m[2]
			filePath := normalizeGitHubFilePath(m[3])
			content, usedURL, err := fetchViaGitHubAPI(ownerRepo, ref, filePath, token)
			if err == nil {
				return &FetchURLResult{URL: rawURL, RawURL: usedURL, Content: content}, nil
			}
			// 已配置 token 时直接返回 API 错误，避免误报 raw 404
			return nil, fmt.Errorf("GitHub API 拉取失败: %w（请检查 GITHUB_TOKEN 是否有效、是否对该仓库有读权限）", err)
		}
	}

	fetchURL := toRawURL(rawURL)
	resp, err := doRequest(fetchURL, "")
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", fetchURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		// GitHub 对 private repo 未授权访问经常返回 404（包括 blob 和 raw）。
		if resp.StatusCode == 404 && strings.Contains(fetchURL, "raw.githubusercontent.com/") {
			return nil, fmt.Errorf("fetch %s: HTTP 404（可能是 GitHub 私有仓库：请配置环境变量 GITHUB_TOKEN，或改用本地文档仓 DOCS_REPO_ROOT）", fetchURL)
		}
		return nil, fmt.Errorf("fetch %s: HTTP %d", fetchURL, resp.StatusCode)
	}
	lr := io.LimitReader(resp.Body, int64(MaxFetchBytes)+1)
	body, err := io.ReadAll(lr)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", fetchURL, err)
	}
	truncated := len(body) > MaxFetchBytes
	if truncated {
		body = body[:MaxFetchBytes]
	}
	return &FetchURLResult{
		URL:       rawURL,
		RawURL:    fetchURL,
		Content:   string(body),
		Truncated: truncated,
	}, nil
}

func toRawURL(u string) string {
	// github.com/.../blob/... → raw.githubusercontent.com
	if m := githubBlobRe.FindStringSubmatch(u); m != nil {
		return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", m[1], m[2], m[3])
	}
	// 已经是 raw URL 或其它
	return u
}

func doRequest(fetchURL string, token string) (*http.Response, error) {
	req, err := http.NewRequest("GET", fetchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "api-test-platform")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return httpClient.Do(req)
}

func normalizeGitHubFilePath(p string) string {
	p = strings.TrimPrefix(strings.TrimSpace(p), "/")
	// blob URL 里路径常为百分号编码，API 需要解码后的真实路径
	if dec, err := url.PathUnescape(p); err == nil {
		p = dec
	}
	// 去掉 ?plain=1 等查询串（若误贴在路径末尾）
	if i := strings.IndexByte(p, '?'); i >= 0 {
		p = p[:i]
	}
	return p
}

func encodeGitHubContentsPath(filePath string) string {
	parts := strings.Split(filePath, "/")
	for i, seg := range parts {
		parts[i] = url.PathEscape(seg)
	}
	return strings.Join(parts, "/")
}

func fetchViaGitHubAPI(ownerRepo, ref, filePath, token string) (content string, usedURL string, err error) {
	// https://api.github.com/repos/{owner}/{repo}/contents/{path}?ref={ref}
	// Accept: application/vnd.github.raw 直接返回 raw 内容。
	encodedPath := encodeGitHubContentsPath(filePath)
	u := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", ownerRepo, encodedPath)
	parsed, err := url.Parse(u)
	if err != nil {
		return "", u, err
	}
	q := parsed.Query()
	if strings.TrimSpace(ref) != "" {
		q.Set("ref", ref)
	}
	parsed.RawQuery = q.Encode()
	usedURL = parsed.String()

	req, err := http.NewRequest("GET", usedURL, nil)
	if err != nil {
		return "", usedURL, err
	}
	req.Header.Set("User-Agent", "api-test-platform")
	// classic: token xxx；fine-grained / PAT: Bearer xxx，两种都兼容
	if strings.HasPrefix(token, "ghp_") || strings.HasPrefix(token, "github_pat_") || strings.HasPrefix(token, "gho_") {
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.raw")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", usedURL, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		msg := strings.TrimSpace(string(body))
		if msg != "" {
			return "", usedURL, fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
		}
		return "", usedURL, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	lr := io.LimitReader(resp.Body, int64(MaxFetchBytes)+1)
	body, err := io.ReadAll(lr)
	if err != nil {
		return "", usedURL, err
	}
	if len(body) > MaxFetchBytes {
		body = body[:MaxFetchBytes]
	}
	return string(body), usedURL, nil
}
