package docsrepo

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DocsBranches 可选分支列表。
type DocsBranches struct {
	Source   string       `json:"source"`
	Project  string       `json:"project,omitempty"`
	Default  string       `json:"default,omitempty"`
	Branches []BranchItem `json:"branches"`
}

type BranchItem struct {
	Name      string `json:"name"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// ListDocsBranches 列出 qa-doc-generator 可用分支（GitLab 远程或本地 git）。
func ListDocsBranches() (*DocsBranches, error) {
	cfg := GitLabDocsSettingsFromEnv()
	if cfg.OK() {
		return listGitLabBranches(cfg)
	}
	root := strings.TrimSpace(os.Getenv("DOCS_REPO_ROOT"))
	if root == "" {
		return nil, fmt.Errorf("请配置 GITLAB_DOCS_PROJECT 或 DOCS_REPO_ROOT")
	}
	return listLocalGitBranches(root)
}

func listGitLabBranches(cfg GitLabDocsSettings) (*DocsBranches, error) {
	token := gitLabToken()
	defaultRef, _ := gitLabDefaultBranch(cfg, token)
	var all []BranchItem
	page := 1
	for {
		encodedProj := url.PathEscape(strings.Trim(cfg.ProjectPath, "/"))
		apiURL := fmt.Sprintf("%s/api/v4/projects/%s/repository/branches?per_page=100&page=%d&sort=updated_desc",
			strings.TrimRight(cfg.BaseURL, "/"), encodedProj, page)
		body, err := gitLabGET(apiURL, token)
		if err != nil {
			return nil, err
		}
		var items []struct {
			Name   string `json:"name"`
			Commit struct {
				CommittedDate time.Time `json:"committed_date"`
			} `json:"commit"`
		}
		if err := json.Unmarshal(body, &items); err != nil {
			return nil, fmt.Errorf("parse GitLab branches: %w", err)
		}
		if len(items) == 0 {
			break
		}
		for _, it := range items {
			all = append(all, BranchItem{
				Name:      it.Name,
				UpdatedAt: it.Commit.CommittedDate.Format(time.RFC3339),
			})
		}
		if len(items) < 100 || page >= 5 {
			break
		}
		page++
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("未获取到分支（检查 GITLAB_DOCS_PROJECT 与 Token）")
	}
	sort.SliceStable(all, func(i, j int) bool {
		return branchRank(all[i].Name) < branchRank(all[j].Name)
	})
	def := strings.TrimSpace(os.Getenv("GITLAB_DOCS_REF"))
	if def == "" {
		def = defaultRef
	}
	if def == "" {
		def = all[0].Name
	}
	return &DocsBranches{
		Source:   "gitlab",
		Project:  cfg.ProjectPath,
		Default:  def,
		Branches: all,
	}, nil
}

func gitLabDefaultBranch(cfg GitLabDocsSettings, token string) (string, error) {
	encodedProj := url.PathEscape(strings.Trim(cfg.ProjectPath, "/"))
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s", strings.TrimRight(cfg.BaseURL, "/"), encodedProj)
	body, err := gitLabGET(apiURL, token)
	if err != nil {
		return "", err
	}
	var resp struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	return resp.DefaultBranch, nil
}

func branchRank(name string) int {
	low := strings.ToLower(name)
	switch {
	case strings.HasPrefix(low, "beta_"), strings.HasPrefix(low, "release_"):
		return 0
	case low == "main" || low == "master":
		return 1
	case strings.HasPrefix(low, "develop"), strings.HasPrefix(low, "dev"):
		return 2
	default:
		return 3
	}
}

func listLocalGitBranches(root string) (*DocsBranches, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	out, err := exec.Command("git", "-C", root, "branch", "-a", "--format=%(refname:short)").Output()
	if err != nil {
		return nil, fmt.Errorf("读取本地 git 分支失败: %w", err)
	}
	seen := map[string]bool{}
	var branches []BranchItem
	for _, line := range strings.Split(string(out), "\n") {
		name := strings.TrimSpace(line)
		if name == "" || strings.Contains(name, "HEAD") {
			continue
		}
		name = strings.TrimPrefix(name, "origin/")
		if seen[name] {
			continue
		}
		seen[name] = true
		branches = append(branches, BranchItem{Name: name})
	}
	if len(branches) == 0 {
		branches = []BranchItem{{Name: "main"}}
	}
	sort.Slice(branches, func(i, j int) bool {
		return branchRank(branches[i].Name) < branchRank(branches[j].Name)
	})
	cur, _ := exec.Command("git", "-C", root, "rev-parse", "--abbrev-ref", "HEAD").Output()
	def := strings.TrimSpace(string(cur))
	if def == "" {
		def = branches[0].Name
	}
	return &DocsBranches{Source: "local", Default: def, Branches: branches}, nil
}
