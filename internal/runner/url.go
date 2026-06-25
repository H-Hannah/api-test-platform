package runner

import (
	"strings"
)

// ResolveRequestURL 拼出最终请求 URL。
// 优先 fullURLTemplate；否则 path 已含 {{base_url_*}} 或绝对地址则直接替换变量；最后回退 base_url+path。
func ResolveRequestURL(path, fullURLTemplate string, vars map[string]string) string {
	if t := strings.TrimSpace(fullURLTemplate); t != "" {
		return substitute(t, vars)
	}
	p := strings.TrimSpace(path)
	if p == "" {
		return substitute(strings.TrimRight(vars["base_url"], "/"), vars)
	}
	if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		return substitute(p, vars)
	}
	if strings.Contains(p, "{{base_url") {
		return substitute(p, vars)
	}
	base := strings.TrimRight(vars["base_url"], "/")
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return substitute(base+p, vars)
}
