package impact

import (
	"regexp"
	"strings"
)

var diffFileRe = regexp.MustCompile(`^\+\+\+ b/(.+)$`)

// parseDiffFiles 从 unified diff 提取变更文件路径。
func parseDiffFiles(diff string) []prFile {
	seen := map[string]bool{}
	var out []prFile
	for _, line := range strings.Split(diff, "\n") {
		line = strings.TrimSpace(line)
		if m := diffFileRe.FindStringSubmatch(line); len(m) == 2 {
			p := strings.TrimSpace(m[1])
			if p != "" && !seen[p] {
				seen[p] = true
				out = append(out, prFile{Path: p, Status: "diff"})
			}
		}
	}
	if len(out) == 0 {
		// fallback: diff --git a/foo b/foo
		gitRe := regexp.MustCompile(`^diff --git a/.+ b/(.+)$`)
		for _, line := range strings.Split(diff, "\n") {
			if m := gitRe.FindStringSubmatch(strings.TrimSpace(line)); len(m) == 2 {
				p := m[1]
				if !seen[p] {
					seen[p] = true
					out = append(out, prFile{Path: p, Status: "diff"})
				}
			}
		}
	}
	return out
}

func fileModule(path string) string {
	p := strings.ToLower(strings.ReplaceAll(path, "\\", "/"))
	parts := strings.Split(p, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return p
}

func fileKeywords(path string) []string {
	p := strings.ToLower(strings.ReplaceAll(path, "\\", "/"))
	seen := map[string]bool{}
	var keys []string
	for _, seg := range strings.FieldsFunc(p, func(r rune) bool {
		return r == '/' || r == '_' || r == '-' || r == '.'
	}) {
		seg = strings.TrimSpace(seg)
		if len(seg) < 3 {
			continue
		}
		switch seg {
		case "internal", "pkg", "src", "api", "handler", "service", "controller", "routes", "go", "ts", "tsx", "vue", "js":
			continue
		}
		if !seen[seg] {
			seen[seg] = true
			keys = append(keys, seg)
		}
	}
	return keys
}

func toChangedFiles(files []prFile) []ChangedFile {
	out := make([]ChangedFile, 0, len(files))
	for _, f := range files {
		out = append(out, ChangedFile{
			Path:   f.Path,
			Status: f.Status,
			Module: fileModule(f.Path),
		})
	}
	return out
}
