package docsrepo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type GitHubRepoSettings struct {
	OwnerRepo string // Everest-Ventures-Group/edgen-product-docs
	Ref       string
	LocalRoot string
}

func GitHubPRDSettingsFromEnv() GitHubRepoSettings {
	ref := strings.TrimSpace(os.Getenv("GITHUB_PRD_REF"))
	if ref == "" {
		ref = "main"
	}
	repo := strings.TrimSpace(os.Getenv("GITHUB_PRD_REPO"))
	if repo == "" {
		repo = "Everest-Ventures-Group/edgen-product-docs"
	}
	return GitHubRepoSettings{
		OwnerRepo: strings.Trim(repo, "/"),
		Ref:       ref,
		LocalRoot: strings.TrimSpace(os.Getenv("GITHUB_PRD_REPO_ROOT")),
	}
}

func (c GitHubRepoSettings) OK() bool {
	return c.OwnerRepo != ""
}

func githubToken() string {
	return strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
}

type ghContentItem struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"` // file | dir
	DownloadURL string `json:"download_url"`
}

var githubHTTP = &http.Client{Timeout: 45 * time.Second}

func githubListContents(ownerRepo, ref, path, token string) ([]ghContentItem, error) {
	encodedPath := encodeGitHubContentsPath(strings.Trim(path, "/"))
	u := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", ownerRepo, encodedPath)
	parsed, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	if ref != "" {
		q := parsed.Query()
		q.Set("ref", ref)
		parsed.RawQuery = q.Encode()
	}
	req, err := http.NewRequest("GET", parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "api-test-platform")
	req.Header.Set("Accept", "application/vnd.github+json")
	if token != "" {
		if strings.HasPrefix(token, "ghp_") || strings.HasPrefix(token, "github_pat_") || strings.HasPrefix(token, "gho_") {
			req.Header.Set("Authorization", "Bearer "+token)
		} else {
			req.Header.Set("Authorization", "token "+token)
		}
	}
	resp, err := githubHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub contents HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var items []ghContentItem
	if err := json.Unmarshal(body, &items); err != nil {
		var one ghContentItem
		if err2 := json.Unmarshal(body, &one); err2 == nil && one.Path != "" {
			return []ghContentItem{one}, nil
		}
		return nil, fmt.Errorf("parse GitHub contents: %w", err)
	}
	return items, nil
}

func githubWalkMarkdown(ownerRepo, ref, rootPath, token string) (text string, loaded []string, err error) {
	items, err := githubListContents(ownerRepo, ref, rootPath, token)
	if err != nil {
		return "", nil, err
	}
	for _, it := range items {
		if it.Type == "dir" {
			sub, subLoaded, err := githubWalkMarkdown(ownerRepo, ref, it.Path, token)
			if err != nil {
				continue
			}
			text = joinSection(text, sub)
			loaded = append(loaded, subLoaded...)
			continue
		}
		if !strings.HasSuffix(strings.ToLower(it.Name), ".md") {
			continue
		}
		content, _, err := fetchViaGitHubAPI(ownerRepo, ref, it.Path, token)
		if err != nil {
			continue
		}
		label := it.Path + " @ " + ref + " (GitHub)"
		loaded = append(loaded, label)
		text = joinSection(text, "## "+label+"\n\n"+content)
	}
	return text, loaded, nil
}

func normalizeVersionTag(version string) string {
	v := strings.TrimSpace(version)
	if v == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(v), "v") {
		v = v[1:]
	}
	parts := strings.Split(v, ".")
	for i, p := range parts {
		parts[i] = strings.TrimLeft(p, "0")
		if parts[i] == "" {
			parts[i] = "0"
		}
	}
	return "V" + strings.Join(parts, ".")
}

func versionMajorLine(version string) string {
	v := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(version)), "v")
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		return fmt.Sprintf("V%s.%s", parts[0], parts[1])
	}
	return normalizeVersionTag(version)
}

// LoadPRDFromProductDocs 从 edgen-product-docs 加载 PRD（GitHub 或本地 clone）。
func LoadPRDFromProductDocs(version, requirementID string) (text string, files []string, err error) {
	cfg := GitHubPRDSettingsFromEnv()
	if cfg.LocalRoot != "" {
		return loadPRDFromProductDocsLocal(cfg.LocalRoot, version, requirementID)
	}
	token := githubToken()
	if token == "" {
		return "", nil, fmt.Errorf("GITHUB_TOKEN 未配置，无法拉取 %s", cfg.OwnerRepo)
	}
	prdDir, err := resolveProductDocsPRDDir(cfg.OwnerRepo, cfg.Ref, version, requirementID, token)
	if err != nil {
		return "", nil, err
	}
	return githubWalkMarkdown(cfg.OwnerRepo, cfg.Ref, prdDir, token)
}

func loadPRDFromProductDocsLocal(root, version, requirementID string) (string, []string, error) {
	major := versionMajorLine(version)
	ver := normalizeVersionTag(version)
	prdRoot := filepath.Join(major, ver, "PRD")
	abs := filepath.Join(root, prdRoot)
	entries, err := os.ReadDir(abs)
	if err != nil {
		return "", nil, fmt.Errorf("本地 %s 不存在: %w", prdRoot, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	match := pickPRDFolder(names, version, requirementID)
	if match == "" {
		return "", nil, fmt.Errorf("未在 %s 下找到与 %q 匹配的需求目录", prdRoot, requirementID)
	}
	dir := filepath.Join(abs, match)
	return loadMarkdownTree(dir, filepath.Join(prdRoot, match))
}

func resolveProductDocsPRDDir(ownerRepo, ref, version, requirementID, token string) (string, error) {
	major := versionMajorLine(version)
	ver := normalizeVersionTag(version)
	prdRoot := major + "/" + ver + "/PRD"
	items, err := githubListContents(ownerRepo, ref, prdRoot, token)
	if err != nil {
		return "", fmt.Errorf("列出 GitHub %s/%s: %w", ownerRepo, prdRoot, err)
	}
	var names []string
	for _, it := range items {
		if it.Type == "dir" {
			names = append(names, it.Name)
		}
	}
	match := pickPRDFolder(names, version, requirementID)
	if match == "" {
		return "", fmt.Errorf("未在 %s 下找到与 %q 匹配的需求目录", prdRoot, requirementID)
	}
	return prdRoot + "/" + match, nil
}

func pickPRDFolder(names []string, version, requirementID string) string {
	rid := strings.ToLower(strings.TrimSpace(requirementID))
	ver := strings.ToLower(normalizeVersionTag(version))
	// 精确：V2.7.2-Brief页面、brief
	var scored []struct {
		name  string
		score int
	}
	for _, n := range names {
		low := strings.ToLower(n)
		score := 0
		if low == rid || strings.Contains(low, rid) {
			score += 10
		}
		if strings.Contains(low, ver) {
			score += 3
		}
		// brief -> brief页面
		if rid == "brief" && strings.Contains(low, "brief") {
			score += 15
		}
		if rid == "tracker" && strings.Contains(low, "tracker") {
			score += 15
		}
		if rid == "invest" && strings.Contains(low, "invest") {
			score += 15
		}
		if score > 0 {
			scored = append(scored, struct {
				name  string
				score int
			}{n, score})
		}
	}
	if len(scored) == 0 {
		return ""
	}
	best := scored[0]
	for _, s := range scored[1:] {
		if s.score > best.score {
			best = s
		}
	}
	return best.name
}
