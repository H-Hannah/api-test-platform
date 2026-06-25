package docsrepo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LoadedPackage 从 Git 需求包目录加载的 PRD / UI / 技术设计 / BDD 文本。
type LoadedPackage struct {
	PackagePath     string      `json:"package_path"`
	PRDText         string      `json:"prd_text"`
	UIDesignText    string      `json:"ui_design_text"`
	FeTechText      string      `json:"fe_tech_text"`
	BeTechText      string      `json:"be_tech_text"`
	BDDText         string      `json:"bdd_text"`
	FigmaURL        string      `json:"figma_url"`
	FilesLoaded     []string    `json:"files_loaded"`
	Meta            PackageMeta `json:"meta"`
}

// LoadPackageOptions 扩展加载选项（阶段 B）。
type LoadPackageOptions struct {
	PackagePath  string
	FeTechPaths  []string
	BeTechPaths  []string
	IncludeBDD   bool
}

type sourcesFile struct {
	Sources []struct {
		Type     string `json:"type"`
		Path     string `json:"path"`
		Required bool   `json:"required"`
	} `json:"sources"`
}

var figmaURLRe = regexp.MustCompile(`https://[^\s\)"']*figma\.com[^\s\)"']*`)

// LoadPackage 从 repoRoot 下读取需求包（目录或单个 .md 文件）。
func LoadPackage(repoRoot, packagePath string) (*LoadedPackage, error) {
	return LoadPackageWithOptions(repoRoot, LoadPackageOptions{PackagePath: packagePath})
}

// LoadPackageWithOptions 加载需求包并可附带前后端技术方案、BDD 目录。
func LoadPackageWithOptions(repoRoot string, opt LoadPackageOptions) (*LoadedPackage, error) {
	rel := strings.Trim(strings.TrimSpace(opt.PackagePath), `/`)
	if rel == "" {
		return nil, fmt.Errorf("package_path required")
	}
	full, _, err := resolveUnderRoot(repoRoot, rel)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(full)
	if err != nil {
		return nil, fmt.Errorf("read package: %w", err)
	}
	out := &LoadedPackage{PackagePath: rel}
	meta, _ := LoadPackageMeta(repoRoot, rel)
	out.Meta = meta
	if info.IsDir() {
		out, err = loadDir(full, rel, out)
		if err != nil {
			return nil, err
		}
	} else {
		out, err = loadSingleFile(full, rel, out)
		if err != nil {
			return nil, err
		}
	}
	if t, loaded, err := LoadFiles(repoRoot, opt.FeTechPaths); err != nil {
		return nil, err
	} else if t != "" {
		out.FeTechText = t
		out.FilesLoaded = append(out.FilesLoaded, loaded...)
	}
	if t, loaded, err := LoadFiles(repoRoot, opt.BeTechPaths); err != nil {
		return nil, err
	} else if t != "" {
		out.BeTechText = t
		out.FilesLoaded = append(out.FilesLoaded, loaded...)
	}
	if opt.IncludeBDD {
		t, loaded, err := LoadBDDDir(repoRoot, rel)
		if err != nil {
			return nil, err
		}
		out.BDDText = t
		out.FilesLoaded = append(out.FilesLoaded, loaded...)
	}
	return out, nil
}

func loadSingleFile(full, rel string, out *LoadedPackage) (*LoadedPackage, error) {
	if !strings.HasSuffix(strings.ToLower(full), ".md") {
		return nil, fmt.Errorf("not a markdown file: %s", rel)
	}
	b, err := os.ReadFile(full)
	if err != nil {
		return nil, err
	}
	text := string(b)
	out.FilesLoaded = append(out.FilesLoaded, rel)
	if strings.Contains(rel, "/design/") || strings.Contains(rel, "design/") {
		out.UIDesignText = text
	} else {
		out.PRDText = text
	}
	out.FigmaURL = firstFigmaURL(text)
	return out, nil
}

func loadDir(dir, rel string, out *LoadedPackage) (*LoadedPackage, error) {
	sourcesPath := filepath.Join(dir, "sources.json")
	if b, err := os.ReadFile(sourcesPath); err == nil {
		var sf sourcesFile
		if err := json.Unmarshal(b, &sf); err != nil {
			return nil, fmt.Errorf("parse sources.json: %w", err)
		}
		for _, s := range sf.Sources {
			p := strings.TrimSpace(s.Path)
			if p == "" {
				continue
			}
			fp := filepath.Join(dir, filepath.FromSlash(p))
			b, err := os.ReadFile(fp)
			if err != nil {
				if s.Required {
					return nil, fmt.Errorf("required source %s: %w", p, err)
				}
				continue
			}
			text := string(b)
			loaded := rel + "/" + strings.TrimPrefix(p, "./")
			out.FilesLoaded = append(out.FilesLoaded, loaded)
			switch strings.ToLower(strings.TrimSpace(s.Type)) {
			case "design":
				out.UIDesignText = joinSection(out.UIDesignText, "## "+loaded+"\n\n"+text)
				if out.FigmaURL == "" {
					out.FigmaURL = firstFigmaURL(text)
				}
			case "tech":
				if strings.Contains(strings.ToLower(loaded), "fe") || strings.Contains(strings.ToLower(loaded), "front") || strings.Contains(strings.ToLower(loaded), "前端") {
					out.FeTechText = joinSection(out.FeTechText, "## "+loaded+"\n\n"+text)
				} else {
					out.BeTechText = joinSection(out.BeTechText, "## "+loaded+"\n\n"+text)
				}
			default:
				out.PRDText = joinSection(out.PRDText, "## "+loaded+"\n\n"+text)
			}
		}
		return out, nil
	}
	// 回退：扫描 product/ 与 design/
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}
		relFile, _ := filepath.Rel(dir, path)
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		text := string(b)
		loaded := rel + "/" + relFile
		out.FilesLoaded = append(out.FilesLoaded, loaded)
		switch {
		case strings.HasPrefix(relFile, "design"+string(os.PathSeparator)) || strings.Contains(relFile, "/design/"):
			out.UIDesignText = joinSection(out.UIDesignText, "## "+loaded+"\n\n"+text)
			if out.FigmaURL == "" {
				out.FigmaURL = firstFigmaURL(text)
			}
		case strings.HasPrefix(relFile, "tech"+string(os.PathSeparator)) || strings.Contains(relFile, "/tech/"):
			if strings.Contains(strings.ToLower(relFile), "fe") || strings.Contains(strings.ToLower(relFile), "front") || strings.Contains(strings.ToLower(relFile), "前端") {
				out.FeTechText = joinSection(out.FeTechText, "## "+loaded+"\n\n"+text)
			} else {
				out.BeTechText = joinSection(out.BeTechText, "## "+loaded+"\n\n"+text)
			}
		default:
			out.PRDText = joinSection(out.PRDText, "## "+loaded+"\n\n"+text)
		}
		return nil
	})
	if out.PRDText == "" && out.UIDesignText == "" {
		return nil, fmt.Errorf("no markdown found under %s", rel)
	}
	return out, nil
}

func joinSection(base, part string) string {
	part = strings.TrimSpace(part)
	if part == "" {
		return base
	}
	if strings.TrimSpace(base) == "" {
		return part
	}
	return base + "\n\n---\n\n" + part
}

func firstFigmaURL(text string) string {
	m := figmaURLRe.FindString(text)
	return strings.TrimRight(m, ".,;)")
}
