package ai

import (
	"strings"

	"api-test-platform/internal/runner"
	"api-test-platform/internal/store"
)

func stripURLVarPrefix(p string) (string, bool) {
	p = strings.TrimSpace(p)
	if !strings.HasPrefix(p, "{{") {
		return p, false
	}
	if i := strings.Index(p, "}}"); i >= 0 {
		return strings.TrimSpace(p[i+2:]), true
	}
	return p, false
}

func ensureLeadingSlash(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "/"
	}
	if rest, ok := stripURLVarPrefix(p); ok {
		if rest != "" && !strings.HasPrefix(rest, "/") {
			return p[:strings.Index(p, "}}")+2] + "/" + rest
		}
		return p
	}
	if !strings.HasPrefix(p, "/") {
		return "/" + p
	}
	return p
}

// PathOnly 去掉 {{var}} 前缀，供列表展示 pathname。
func PathOnly(p string) string {
	p = strings.TrimSpace(p)
	if rest, ok := stripURLVarPrefix(p); ok {
		return ensureLeadingSlash(rest)
	}
	return ensureLeadingSlash(p)
}

// buildURLTemplate 用全部运行环境变量扫描录制 URL，将匹配值替换为 {{key}}。
func buildURLTemplate(records []RawRecord, item AIAPIItem, envs []*store.Environment) (fullTpl, pathOnly string) {
	rec := matchRecord(records, item)
	raw := recordFullURL(rec, item)
	if raw == "" {
		p := PathOnly(item.Path)
		return p, p
	}
	fullTpl = runner.TemplatizeFromEnvironments(raw, envs)
	if fullTpl == "" {
		fullTpl = raw
	}
	return fullTpl, PathOnly(fullTpl)
}

func recordFullURL(rec *RawRecord, item AIAPIItem) string {
	if rec == nil {
		return ""
	}
	if u := strings.TrimSpace(rec.URL); u != "" {
		return u
	}
	host := strings.TrimSpace(rec.Host)
	p := strings.TrimSpace(rec.Path)
	if p == "" {
		p = strings.TrimSpace(item.Path)
	}
	if host == "" || p == "" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return "https://" + host + p
}

func templatizeHeaderKVs(list []HeaderKV, envs []*store.Environment) []HeaderKV {
	if len(envs) == 0 {
		return list
	}
	out := make([]HeaderKV, len(list))
	for i, h := range list {
		out[i] = HeaderKV{
			Name:    h.Name,
			Value:   runner.TemplatizeFromEnvironments(h.Value, envs),
			Enabled: h.Enabled,
		}
	}
	return out
}

func templatizeBody(body string, envs []*store.Environment) string {
	if body == "" || len(envs) == 0 {
		return body
	}
	return runner.TemplatizeFromEnvironments(body, envs)
}

func pathFromURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if i := strings.Index(raw, "://"); i >= 0 {
		raw = raw[i+3:]
	}
	if j := strings.Index(raw, "/"); j >= 0 {
		return raw[j:]
	}
	return ""
}

func matchRecord(records []RawRecord, item AIAPIItem) *RawRecord {
	method := strings.ToUpper(item.Method)
	itemPath := PathOnly(item.Path)
	for i := range records {
		r := &records[i]
		if !strings.EqualFold(r.Method, method) {
			continue
		}
		rPath := r.Path
		if rPath == "" {
			rPath = pathFromURL(r.URL)
		}
		rPath = PathOnly(rPath)
		if rPath != "" && (rPath == itemPath || strings.Contains(rPath, itemPath) || strings.Contains(itemPath, rPath)) {
			return r
		}
		if r.URL != "" && itemPath != "" && strings.Contains(r.URL, itemPath) {
			return r
		}
	}
	return nil
}
