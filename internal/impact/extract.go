package impact

import (
	"regexp"
	"strings"
)

var (
	apiPathRe = regexp.MustCompile(`["'](/v\d+[^"'\s]*)["']`)
	routeRe   = regexp.MustCompile(`(?i)(GET|POST|PUT|DELETE|PATCH)\s+["'](/[^"'\s]+)["']`)
	mrPathRe  = regexp.MustCompile(`(?i)^(GET|POST|PUT|DELETE|PATCH)?[:/\s]*(/v?\d*[^\s#]+)`)
)

// ExtractInferredAPIs 从 patch、显式 MR 列表提取 API path。
func ExtractInferredAPIs(files []prFile, explicit []APIItem) []string {
	seen := map[string]bool{}
	var out []string

	add := func(method, path string) {
		path = normalizePath(path)
		if path == "" {
			return
		}
		key := strings.ToUpper(strings.TrimSpace(method)) + " " + path
		if method == "" {
			key = path
		}
		if !seen[key] {
			seen[key] = true
			out = append(out, key)
		}
	}

	for _, item := range explicit {
		add(item.Method, item.Path)
	}

	corpus := strings.Builder{}
	for _, f := range files {
		corpus.WriteString(f.Path)
		corpus.WriteByte('\n')
		corpus.WriteString(f.Patch)
		corpus.WriteByte('\n')
	}
	all := corpus.String()

	for _, m := range routeRe.FindAllStringSubmatch(all, -1) {
		if len(m) >= 3 {
			add(m[1], m[2])
		}
	}
	for _, m := range apiPathRe.FindAllStringSubmatch(all, -1) {
		if len(m) >= 2 {
			add("", m[1])
		}
	}
	return out
}

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, ":")
	if p == "" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func parseAPIKey(key string) (method, path string) {
	key = strings.TrimSpace(key)
	if i := strings.IndexByte(key, ' '); i > 0 {
		return strings.ToUpper(key[:i]), normalizePath(key[i+1:])
	}
	return "", normalizePath(key)
}

func pathMatches(inferred, apiPath string) bool {
	apiPath = normalizePath(apiPath)
	infMethod, infPath := parseAPIKey(inferred)
	if infPath == "" {
		return false
	}
	if infMethod == "" {
		return strings.Contains(apiPath, infPath) || strings.Contains(infPath, apiPath)
	}
	return strings.EqualFold(infPath, apiPath) || strings.HasPrefix(apiPath, infPath)
}

func corpusFromFiles(files []prFile) string {
	var b strings.Builder
	for _, f := range files {
		b.WriteString(f.Path)
		b.WriteByte('\n')
		b.WriteString(f.Patch)
		b.WriteByte('\n')
	}
	return b.String()
}
