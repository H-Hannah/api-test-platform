package docsrepo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// PackageMeta 来自需求包 meta.json。
type PackageMeta struct {
	Version         string `json:"version"`
	RequirementID   string `json:"requirement_id"`
	RequirementName string `json:"requirement_name"`
}

func LoadPackageMeta(repoRoot, packagePath string) (PackageMeta, error) {
	var meta PackageMeta
	root, rel, err := resolveUnderRoot(repoRoot, packagePath)
	if err != nil {
		return meta, err
	}
	dir := root
	if st, err := os.Stat(root); err == nil && !st.IsDir() {
		dir = filepath.Dir(root)
		rel = filepath.Dir(rel)
	}
	b, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	if err != nil {
		// 从路径推断 version / id
		parts := strings.Split(strings.Trim(rel, "/"), "/")
		if len(parts) >= 2 {
			meta.Version = parts[len(parts)-2]
			meta.RequirementID = parts[len(parts)-1]
		}
		return meta, nil
	}
	_ = json.Unmarshal(b, &meta)
	return meta, nil
}
