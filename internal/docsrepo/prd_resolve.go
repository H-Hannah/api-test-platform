package docsrepo

import (
	"fmt"
	"strings"
)

// PRDResolveResult 解析 PRD 输入后的内容。
type PRDResolveResult struct {
	Source      string `json:"source"`       // local | remote
	PackagePath string `json:"package_path"` // 本地相对路径（若适用）
	PRDText     string `json:"prd_text"`
	UIDesignText string `json:"ui_design_text"`
	FigmaURL    string `json:"figma_url"`
}

// ResolvePRD 优先从本地文档仓读取；未配置仓或路径不可解析时再尝试远程 URL。
func ResolvePRD(repoRoot, prdURL, packagePath, existingPRD string) (*PRDResolveResult, error) {
	if strings.TrimSpace(existingPRD) != "" {
		return &PRDResolveResult{Source: "inline", PRDText: existingPRD}, nil
	}

	rel := strings.TrimSpace(packagePath)
	if rel == "" {
		rel = ResolvePackagePath(prdURL)
	}

	root := strings.TrimSpace(repoRoot)
	if rel != "" && root != "" {
		pkg, err := LoadPackage(root, rel)
		if err != nil {
			return nil, fmt.Errorf("从本地文档仓读取失败（DOCS_REPO_ROOT=%s，路径=%s）: %w", root, rel, err)
		}
		out := &PRDResolveResult{
			Source:       "local",
			PackagePath:  rel,
			PRDText:      pkg.PRDText,
			UIDesignText: pkg.UIDesignText,
			FigmaURL:     pkg.FigmaURL,
		}
		if out.PRDText == "" && len(pkg.FilesLoaded) == 1 {
			// 单文件加载时 PRD 在 PRDText
			out.PRDText = pkg.PRDText
		}
		return out, nil
	}

	u := strings.TrimSpace(prdURL)
	if u == "" {
		if rel != "" && root == "" {
			return nil, fmt.Errorf("已解析路径 %q，但未配置 DOCS_REPO_ROOT：请在 .env 设置本地 clone 的 edgen-product-docs 目录后重启服务", rel)
		}
		return nil, fmt.Errorf("请提供 PRD 链接、本地相对路径，或粘贴 PRD 文本")
	}

	if rel != "" && root == "" {
		return nil, fmt.Errorf(
			"GitHub 链接已解析为 %q，但未配置 DOCS_REPO_ROOT。\n请执行：git clone git@github.com:Everest-Ventures-Group/edgen-product-docs.git\n然后在 .env 设置 DOCS_REPO_ROOT=/path/to/edgen-product-docs 并重启后端",
			rel,
		)
	}

	res, err := FetchDocURL(u)
	if err != nil {
		return nil, err
	}
	out := &PRDResolveResult{Source: "remote", PRDText: res.Content}
	if out.PRDText != "" {
		out.FigmaURL = firstFigmaURL(out.PRDText)
	}
	return out, nil
}
