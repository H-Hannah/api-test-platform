package ai

import (
	"strings"
)

// ServiceBaseVarKey 将微服务名映射为环境变量键，如 anchor -> base_url_anchor。
func ServiceBaseVarKey(service string) string {
	s := normalizeServiceName(service)
	if s == "" {
		return "base_url"
	}
	return "base_url_" + s
}

func normalizeServiceName(service string) string {
	s := strings.ToLower(strings.TrimSpace(service))
	s = strings.ReplaceAll(s, "-", "_")
	switch s {
	case "", "www", "api", "app", "m":
		return ""
	case "badge", "persona", "portal":
		return "trex"
	case "quests":
		return "quest"
	case "openreplay", "replay":
		return "openreplay"
	case "edgen", "ospprotocol":
		return "edgen"
	default:
		return s
	}
}

func serviceFromFolderPath(path []string) string {
	if len(path) < 2 {
		return ""
	}
	return normalizeServiceName(path[len(path)-1])
}

func ensureLeadingSlash(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "/"
	}
	if strings.Contains(p, "{{base_url") {
		if i := strings.Index(p, "}}"); i >= 0 {
			rest := strings.TrimSpace(p[i+2:])
			if rest != "" && !strings.HasPrefix(rest, "/") {
				return p[:i+2] + "/" + rest
			}
		}
		return p
	}
	if !strings.HasPrefix(p, "/") {
		return "/" + p
	}
	return p
}

// PathOnly 去掉模板前缀，供列表展示。
func PathOnly(p string) string {
	p = strings.TrimSpace(p)
	if strings.Contains(p, "{{base_url") {
		if i := strings.Index(p, "}}"); i >= 0 {
			return ensureLeadingSlash(p[i+2:])
		}
	}
	return ensureLeadingSlash(p)
}

// BuildFullURLTemplate 生成可执行 URL 模板。
func BuildFullURLTemplate(service, path string) string {
	path = PathOnly(path)
	if strings.Contains(path, "{{base_url") {
		return path
	}
	key := ServiceBaseVarKey(service)
	return "{{" + key + "}}" + path
}

// BuildStepRequestPath 场景步骤存库用的请求路径（含服务 base 变量）。
func BuildStepRequestPath(service, path string) string {
	return BuildFullURLTemplate(service, path)
}

func resolveServiceForItem(records []RawRecord, item AIAPIItem) string {
	if r := matchRecord(records, item); r != nil && r.Service != "" {
		return r.Service
	}
	if s := serviceFromFolderPath(item.FolderPath); s != "" {
		return s
	}
	return ""
}

func resolveServiceForStep(records []RawRecord, step AIScenarioStep) string {
	item := AIAPIItem{Method: step.Method, Path: step.Path}
	if r := matchRecord(records, item); r != nil && r.Service != "" {
		return r.Service
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
