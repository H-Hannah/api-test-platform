package docsrepo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveUnderRoot(repoRoot, relPath string) (abs, rel string, err error) {
	root, err := filepath.Abs(strings.TrimSpace(repoRoot))
	if err != nil {
		return "", "", err
	}
	if root == "" {
		return "", "", fmt.Errorf("DOCS_REPO_ROOT 未配置")
	}
	rel = strings.Trim(strings.TrimSpace(relPath), `/`)
	if rel == "" {
		return "", "", fmt.Errorf("path required")
	}
	abs, err = filepath.Abs(filepath.Join(root, rel))
	if err != nil {
		return "", "", err
	}
	if !strings.HasPrefix(abs, root+string(os.PathSeparator)) && abs != root {
		return "", "", fmt.Errorf("path escapes repo root")
	}
	return abs, rel, nil
}

// LoadFiles 读取文档仓内多个相对路径并拼接（支持目录，递归 .md）。
func LoadFiles(repoRoot string, paths []string) (text string, loaded []string, err error) {
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		abs, rel, e := resolveUnderRoot(repoRoot, p)
		if e != nil {
			return "", loaded, e
		}
		st, e := os.Stat(abs)
		if e != nil {
			return "", loaded, fmt.Errorf("read %s: %w", p, e)
		}
		if st.IsDir() {
			t, l, e := loadMarkdownTree(abs, rel)
			if e != nil {
				return "", loaded, e
			}
			text = joinSection(text, t)
			loaded = append(loaded, l...)
			continue
		}
		b, e := os.ReadFile(abs)
		if e != nil {
			return "", loaded, fmt.Errorf("read %s: %w", p, e)
		}
		loaded = append(loaded, rel)
		text = joinSection(text, "## "+rel+"\n\n"+string(b))
	}
	return text, loaded, nil
}

func loadMarkdownTree(dir, rel string) (text string, loaded []string, err error) {
	err = filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}
		relFile, _ := filepath.Rel(dir, path)
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		loaded = append(loaded, rel+"/"+relFile)
		text = joinSection(text, "## "+rel+"/"+relFile+"\n\n"+string(b))
		return nil
	})
	return text, loaded, err
}

// LoadBDDDir 加载需求包下 bdd/ 目录（feature + contract + matrix）。
func LoadBDDDir(repoRoot, packagePath string) (text string, files []string, err error) {
	abs, rel, err := resolveUnderRoot(repoRoot, packagePath)
	if err != nil {
		return "", nil, err
	}
	bddDir := filepath.Join(abs, "bdd")
	if st, err := os.Stat(bddDir); err != nil || !st.IsDir() {
		return "", nil, fmt.Errorf("bdd/ 目录不存在: %s/bdd", rel)
	}
	_ = filepath.Walk(bddDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return nil
		}
		lower := strings.ToLower(info.Name())
		if !strings.HasSuffix(lower, ".md") && !strings.HasSuffix(lower, ".feature") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		relFile, _ := filepath.Rel(abs, path)
		loaded := rel + "/" + relFile
		files = append(files, loaded)
		text = joinSection(text, "## "+loaded+"\n\n"+string(b))
		return nil
	})
	if text == "" {
		return "", files, fmt.Errorf("bdd/ 下无 .feature 或 .md 文件")
	}
	return text, files, nil
}
