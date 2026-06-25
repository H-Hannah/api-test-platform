package docsrepo

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// LoadedRequirementPackage 需求文档 + 后端技术方案 + 测试用例。
type LoadedRequirementPackage struct {
	Version         string   `json:"version"`
	RequirementID   string   `json:"requirement_id"`
	RequirementName string   `json:"requirement_name"`
	Ref             string   `json:"ref,omitempty"`
	Source          string   `json:"source"`
	PRDSource       string   `json:"prd_source,omitempty"`
	BeTechSource    string   `json:"be_tech_source,omitempty"`
	TcSource        string   `json:"tc_source,omitempty"`
	PRDText         string   `json:"prd_text"`
	BeTechText      string   `json:"be_tech_text"`
	UIDesignText    string   `json:"ui_design_text,omitempty"`
	CasesJSON       string   `json:"cases_json,omitempty"`
	CaseCount       int      `json:"case_count"`
	CasesSummary    string   `json:"cases_summary,omitempty"`
	FilesLoaded     []string `json:"files_loaded"`
	Warnings        []string `json:"warnings,omitempty"`
}

type reqSourceEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

// LoadRequirementPackage 从 edgen-product-docs + osp-wiki + qa-doc-generator 加载三类文档。
func LoadRequirementPackage(version, requirementID, ref string) (*LoadedRequirementPackage, error) {
	ver := strings.TrimSpace(version)
	rid := strings.TrimSpace(requirementID)
	if ver == "" || rid == "" {
		return nil, fmt.Errorf("version and requirement_id required")
	}
	if strings.TrimSpace(ref) == "" {
		return nil, fmt.Errorf("请先选择 qa-doc-generator 分支")
	}

	out := &LoadedRequirementPackage{
		Version:         ver,
		RequirementID:   rid,
		RequirementName: rid,
		Ref:             ref,
		Source:          "multi",
	}

	// 1. PRD — GitHub edgen-product-docs（失败则回退 qa-doc-generator requirements/）
	if prd, files, err := LoadPRDFromProductDocs(ver, rid); err != nil {
		if fb, fbFiles, fbErr := loadPRDFallbackQADoc(ver, rid, ref); fbErr == nil && fb != "" {
			out.PRDText = fb
			out.PRDSource = "qa-doc-generator/requirements (fallback)"
			out.FilesLoaded = append(out.FilesLoaded, fbFiles...)
			out.Warnings = append(out.Warnings, "需求文档: GitHub 未命中，已用 qa-doc-generator 回退")
		} else {
			out.Warnings = append(out.Warnings, "需求文档: "+err.Error())
		}
	} else {
		out.PRDText = prd
		out.PRDSource = prdSourceLabel()
		out.FilesLoaded = append(out.FilesLoaded, files...)
	}

	// 2. 后端设计 — GitLab osp-wiki features/（失败则回退 qa-doc-generator）
	if be, files, err := LoadBackendTechFromOSWiki(ver, rid); err != nil {
		if fb, fbFiles, fbErr := loadBEFallbackQADoc(ver, rid, ref); fbErr == nil && fb != "" {
			out.BeTechText = fb
			out.BeTechSource = "qa-doc-generator/requirements (fallback)"
			out.FilesLoaded = append(out.FilesLoaded, fbFiles...)
			out.Warnings = append(out.Warnings, "后端技术方案: osp-wiki 未命中，已用 qa-doc-generator 回退")
		} else {
			out.Warnings = append(out.Warnings, "后端技术方案: "+err.Error())
		}
	} else {
		out.BeTechText = be
		out.BeTechSource = beSourceLabel()
		out.FilesLoaded = append(out.FilesLoaded, files...)
	}

	// 3. 测试用例 — qa-doc-generator test-docs/
	tc, err := LoadTestCasesBySelector(ver, rid, ref)
	if err != nil {
		out.Warnings = append(out.Warnings, "测试用例: "+err.Error())
	} else {
		out.CasesJSON = tc.CasesJSON
		out.CaseCount = tc.CaseCount
		out.CasesSummary = tc.Summary
		out.TcSource = "qa-doc-generator/test-docs"
		if tc.RequirementName != "" {
			out.RequirementName = tc.RequirementName
		}
	}

	out.Warnings = append(out.Warnings, missingDocWarnings(out)...)
	return out, nil
}

func prdSourceLabel() string {
	cfg := GitHubPRDSettingsFromEnv()
	if cfg.LocalRoot != "" {
		return "edgen-product-docs (local)"
	}
	return cfg.OwnerRepo + " @ " + cfg.Ref
}

func beSourceLabel() string {
	cfg := GitLabBackendSettingsFromEnv()
	return cfg.ProjectPath + "/" + cfg.TreePath + " @ " + cfg.Ref
}

func loadPRDFallbackQADoc(version, requirementID, ref string) (string, []string, error) {
	pkg, err := loadQADocRequirementSlice(version, requirementID, ref)
	if err != nil {
		return "", nil, err
	}
	return pkg.PRDText, filterLoadedPRD(pkg), nil
}

func loadBEFallbackQADoc(version, requirementID, ref string) (string, []string, error) {
	pkg, err := loadQADocRequirementSlice(version, requirementID, ref)
	if err != nil {
		return "", nil, err
	}
	return pkg.BeTechText, filterLoadedBE(pkg), nil
}

func loadQADocRequirementSlice(version, requirementID, ref string) (*LoadedRequirementPackage, error) {
	pkgPath := fmt.Sprintf("requirements/%s/%s", version, requirementID)
	root := strings.TrimSpace(os.Getenv("DOCS_REPO_ROOT"))
	if root != "" {
		return loadLocalRequirementDocs(root, pkgPath)
	}
	cfg := GitLabDocsSettingsFromEnv()
	if !cfg.OK() {
		return nil, fmt.Errorf("DOCS_REPO_ROOT 与 GITLAB_DOCS_PROJECT 均未配置")
	}
	return loadQADocRequirementFromGitLab(cfg, version, requirementID, ref, pkgPath)
}

func filterLoadedPRD(p *LoadedRequirementPackage) []string {
	// FilesLoaded 混合了 prd/be，回退场景全部返回即可
	return p.FilesLoaded
}

func filterLoadedBE(p *LoadedRequirementPackage) []string {
	return p.FilesLoaded
}

func loadLocalRequirementDocs(root, pkgPath string) (*LoadedRequirementPackage, error) {
	abs, rel, err := resolveUnderRoot(root, pkgPath)
	if err != nil {
		return nil, err
	}
	out := &LoadedRequirementPackage{}
	var sources []reqSourceEntry
	if b, err := os.ReadFile(filepath.Join(abs, "sources.json")); err == nil {
		var sf struct {
			Sources []reqSourceEntry `json:"sources"`
		}
		_ = json.Unmarshal(b, &sf)
		sources = sf.Sources
	}
	sourceTypeByFile := map[string]string{}
	for _, s := range sources {
		p := strings.TrimPrefix(strings.TrimSpace(s.Path), "./")
		if p != "" {
			sourceTypeByFile[p] = strings.ToLower(strings.TrimSpace(s.Type))
			sourceTypeByFile[filepath.Base(p)] = strings.ToLower(strings.TrimSpace(s.Type))
		}
	}
	_ = filepath.Walk(abs, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			return nil
		}
		relFile, _ := filepath.Rel(abs, path)
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		text := string(b)
		loaded := rel + "/" + relFile
		out.FilesLoaded = append(out.FilesLoaded, loaded)
		sourceType := sourceTypeByFile[relFile]
		if sourceType == "" {
			sourceType = sourceTypeByFile[info.Name()]
		}
		kind := classifyDocFile(info.Name(), relFile, sourceType)
		switch kind {
		case "be":
			out.BeTechText = joinSection(out.BeTechText, "## "+loaded+"\n\n"+text)
		case "ui":
			out.UIDesignText = joinSection(out.UIDesignText, "## "+loaded+"\n\n"+text)
		default:
			out.PRDText = joinSection(out.PRDText, "## "+loaded+"\n\n"+text)
		}
		return nil
	})
	if strings.TrimSpace(out.PRDText) == "" && strings.TrimSpace(out.BeTechText) == "" {
		pkg, err := LoadPackage(root, pkgPath)
		if err == nil {
			out.PRDText = pkg.PRDText
			out.BeTechText = pkg.BeTechText
			out.UIDesignText = pkg.UIDesignText
			out.FilesLoaded = append(out.FilesLoaded, pkg.FilesLoaded...)
		}
	}
	if strings.TrimSpace(out.BeTechText) == "" {
		be, loaded, err := loadLocalBackendTech(root, pkgPath)
		if err == nil && be != "" {
			out.BeTechText = be
			out.FilesLoaded = append(out.FilesLoaded, loaded...)
		}
	}
	return out, nil
}

func loadLocalBackendTech(root, pkgPath string) (text string, loaded []string, err error) {
	abs, rel, err := resolveUnderRoot(root, pkgPath)
	if err != nil {
		return "", nil, err
	}
	techDir := filepath.Join(abs, "tech")
	if st, err := os.Stat(techDir); err == nil && st.IsDir() {
		return loadMarkdownTree(techDir, rel+"/tech")
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return "", nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		if !looksLikeBackendTech(name) {
			continue
		}
		b, err := os.ReadFile(filepath.Join(abs, e.Name()))
		if err != nil {
			continue
		}
		loaded = append(loaded, rel+"/"+e.Name())
		text = joinSection(text, "## "+rel+"/"+e.Name()+"\n\n"+string(b))
	}
	return text, loaded, nil
}

func loadQADocRequirementFromGitLab(cfg GitLabDocsSettings, version, requirementID, ref, pkgPath string) (*LoadedRequirementPackage, error) {
	token := gitLabToken()
	glRef := &gitLabRef{
		BaseURL:     cfg.BaseURL,
		ProjectPath: cfg.ProjectPath,
		Ref:         ref,
		RepoPath:    pkgPath,
	}
	entries, err := gitLabWalkTree(glRef, pkgPath, token)
	if err != nil {
		return nil, fmt.Errorf("加载 qa-doc-generator requirements 失败: %w", err)
	}
	var sources []reqSourceEntry
	for _, e := range entries {
		if e.Name == "sources.json" {
			b, err := gitLabFetchRaw(glRef, e.Path, token)
			if err == nil {
				var sf struct {
					Sources []reqSourceEntry `json:"sources"`
				}
				_ = json.Unmarshal(b, &sf)
				sources = sf.Sources
			}
			break
		}
	}
	sourceTypeByFile := map[string]string{}
	for _, s := range sources {
		p := strings.TrimPrefix(strings.TrimSpace(s.Path), "./")
		if p != "" {
			sourceTypeByFile[p] = strings.ToLower(strings.TrimSpace(s.Type))
			sourceTypeByFile[filepath.Base(p)] = strings.ToLower(strings.TrimSpace(s.Type))
		}
	}

	out := &LoadedRequirementPackage{
		Version:         version,
		RequirementID:   requirementID,
		RequirementName: requirementID,
		Ref:             ref,
	}
	for _, e := range entries {
		if !strings.HasSuffix(strings.ToLower(e.Name), ".md") {
			continue
		}
		b, err := gitLabFetchRaw(glRef, e.Path, token)
		if err != nil {
			continue
		}
		text := string(b)
		loaded := e.Path + " @ " + ref
		out.FilesLoaded = append(out.FilesLoaded, loaded)
		sourceType := sourceTypeByFile[e.Path]
		if sourceType == "" {
			sourceType = sourceTypeByFile[e.Name]
		}
		kind := classifyDocFile(e.Name, e.Path, sourceType)
		switch kind {
		case "be":
			out.BeTechText = joinSection(out.BeTechText, "## "+loaded+"\n\n"+text)
		case "ui":
			out.UIDesignText = joinSection(out.UIDesignText, "## "+loaded+"\n\n"+text)
		default:
			out.PRDText = joinSection(out.PRDText, "## "+loaded+"\n\n"+text)
		}
	}
	return out, nil
}

func missingDocWarnings(p *LoadedRequirementPackage) []string {
	var w []string
	if strings.TrimSpace(p.PRDText) == "" {
		w = append(w, "未找到需求文档（edgen-product-docs）")
	}
	if strings.TrimSpace(p.BeTechText) == "" {
		w = append(w, "未找到后端技术方案（osp-wiki features/）")
	}
	if strings.TrimSpace(p.CasesJSON) == "" {
		w = append(w, "未找到测试用例 JSON（qa-doc-generator test-docs/）")
	}
	return w
}

func classifyDocFile(name, repoPath, sourceType string) string {
	switch strings.ToLower(strings.TrimSpace(sourceType)) {
	case "technical_spec", "backend", "be", "api_design":
		return "be"
	case "ui_design", "design":
		return "ui"
	case "product_requirement", "prd":
		return "prd"
	}
	lower := strings.ToLower(name)
	pathLower := strings.ToLower(repoPath)
	if strings.Contains(pathLower, "/tech/") || looksLikeBackendTech(lower) {
		return "be"
	}
	if strings.Contains(pathLower, "/design/") || strings.Contains(lower, "-design") {
		return "ui"
	}
	return "prd"
}

func looksLikeBackendTech(lowerName string) bool {
	return strings.Contains(lowerName, "service-design") ||
		strings.Contains(lowerName, "java-service") ||
		strings.Contains(lowerName, "api-design") ||
		strings.Contains(lowerName, "backend") ||
		strings.Contains(lowerName, "技术方案")
}

type gitLabTreeItem struct {
	Name string
	Path string
	Type string
}

func gitLabWalkTree(ref *gitLabRef, treePath, token string) ([]gitLabTreeItem, error) {
	entries, err := gitLabListTreePage(ref, treePath, token)
	if err != nil {
		return nil, err
	}
	var out []gitLabTreeItem
	for _, e := range entries {
		if e.Type == "blob" {
			out = append(out, e)
			continue
		}
		if e.Type == "tree" {
			sub, err := gitLabWalkTree(ref, e.Path, token)
			if err != nil {
				return nil, err
			}
			out = append(out, sub...)
		}
	}
	return out, nil
}

func gitLabListTreePage(ref *gitLabRef, treePath, token string) ([]gitLabTreeItem, error) {
	encodedProj := url.PathEscape(strings.Trim(ref.ProjectPath, "/"))
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/repository/tree?path=%s&ref=%s&per_page=100",
		strings.TrimRight(ref.BaseURL, "/"),
		encodedProj,
		url.QueryEscape(treePath),
		url.QueryEscape(ref.Ref),
	)
	body, err := gitLabGET(apiURL, token)
	if err != nil {
		return nil, err
	}
	var items []struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("parse GitLab tree: %w", err)
	}
	out := make([]gitLabTreeItem, 0, len(items))
	for _, it := range items {
		out = append(out, gitLabTreeItem{Name: it.Name, Path: it.Path, Type: it.Type})
	}
	return out, nil
}
