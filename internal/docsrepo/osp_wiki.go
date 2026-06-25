package docsrepo

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// GitLabBackendSettings osp-wiki 后端设计文档仓。
type GitLabBackendSettings struct {
	BaseURL     string
	ProjectPath string
	Ref         string
	TreePath    string
}

func GitLabBackendSettingsFromEnv() GitLabBackendSettings {
	base := strings.TrimSpace(os.Getenv("GITLAB_BASE_URL"))
	if base == "" {
		base = "https://gitlab.com"
	}
	ref := strings.TrimSpace(os.Getenv("GITLAB_BE_REF"))
	if ref == "" {
		ref = "ospdev"
	}
	tree := strings.TrimSpace(os.Getenv("GITLAB_BE_TREE"))
	if tree == "" {
		tree = "features"
	}
	proj := strings.TrimSpace(os.Getenv("GITLAB_BE_PROJECT"))
	if proj == "" {
		proj = "Keccak256-evg/opensocial/osp-wiki"
	}
	return GitLabBackendSettings{
		BaseURL:     strings.TrimRight(base, "/"),
		ProjectPath: strings.Trim(proj, "/"),
		Ref:         ref,
		TreePath:    strings.Trim(tree, "/"),
	}
}

func (c GitLabBackendSettings) OK() bool {
	return c.ProjectPath != ""
}

// LoadBackendTechFromOSWiki 从 osp-wiki features/ 加载与需求相关的后端设计 md。
func LoadBackendTechFromOSWiki(version, requirementID string) (text string, files []string, err error) {
	cfg := GitLabBackendSettingsFromEnv()
	if !cfg.OK() {
		return "", nil, fmt.Errorf("GITLAB_BE_PROJECT 未配置")
	}
	token := gitLabToken()
	glRef := &gitLabRef{
		BaseURL:     cfg.BaseURL,
		ProjectPath: cfg.ProjectPath,
		Ref:         cfg.Ref,
		RepoPath:    cfg.TreePath,
	}
	top, err := gitLabListTreePage(glRef, cfg.TreePath, token)
	if err != nil {
		return "", nil, err
	}
	var featureDirs []string
	for _, e := range top {
		if e.Type == "tree" {
			featureDirs = append(featureDirs, e.Name)
		}
	}
	matched := matchOSWikiFeatures(requirementID, featureDirs)
	if len(matched) == 0 {
		return "", nil, fmt.Errorf("osp-wiki features/ 下未找到与 %q 匹配的后端设计目录", requirementID)
	}
	for _, feat := range matched {
		featPath := cfg.TreePath + "/" + feat
		blobs, err := gitLabWalkTree(glRef, featPath, token)
		if err != nil {
			continue
		}
		for _, b := range blobs {
			if !strings.HasSuffix(strings.ToLower(b.Name), ".md") {
				continue
			}
			if !looksLikeBackendDoc(b.Name) {
				continue
			}
			content, err := gitLabFetchRaw(glRef, b.Path, token)
			if err != nil {
				continue
			}
			label := b.Path + " @ " + cfg.Ref + " (osp-wiki)"
			files = append(files, label)
			text = joinSection(text, "## "+label+"\n\n"+string(content))
		}
	}
	if strings.TrimSpace(text) == "" {
		return "", files, fmt.Errorf("已匹配目录 %v 但未读到后端设计 md", matched)
	}
	return text, files, nil
}

func looksLikeBackendDoc(name string) bool {
	low := strings.ToLower(name)
	if strings.Contains(low, "design") ||
		strings.Contains(low, "service") ||
		strings.Contains(low, "architecture") ||
		strings.Contains(low, "analysis") ||
		strings.Contains(low, "schema") ||
		strings.Contains(low, "plan") {
		return true
	}
	return looksLikeBackendTech(low)
}

var beFeatureAliases = map[string][]string{
	"brief":                    {"proactive-tracker", "proactive-agent"},
	"tracker":                  {"proactive-tracker", "edgen-tracker-dashboard", "guru-tracker"},
	"tracker-tab-catchup-push": {"proactive-tracker"},
	"invest":                   {"invest-plan", "financial-planner"},
	"plaid-integration":        {"plaid-user-assets"},
	"plaid":                    {"plaid-user-assets"},
	"financial-planner":        {"financial-planner"},
	"plan-page":                {"invest-plan", "financial-planner"},
	"goal-page":                {"financial-planner"},
}

func matchOSWikiFeatures(requirementID string, features []string) []string {
	rid := strings.ToLower(strings.TrimSpace(requirementID))
	if rid == "" {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	for _, alias := range beFeatureAliases[rid] {
		for _, f := range features {
			if strings.Contains(strings.ToLower(f), alias) {
				add(f)
			}
		}
	}
	type scored struct {
		name  string
		score int
	}
	var ranks []scored
	for _, f := range features {
		low := strings.ToLower(f)
		score := 0
		if strings.Contains(low, rid) {
			score += 20
		}
		for _, tok := range strings.Split(rid, "-") {
			if len(tok) >= 4 && strings.Contains(low, tok) {
				score += 5
			}
		}
		if score > 0 {
			ranks = append(ranks, scored{f, score})
		}
	}
	sort.Slice(ranks, func(i, j int) bool { return ranks[i].score > ranks[j].score })
	for _, r := range ranks {
		add(r.name)
	}
	if len(out) > 3 {
		out = out[:3]
	}
	return out
}
