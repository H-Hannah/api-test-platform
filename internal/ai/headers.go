package ai

import (
	"strings"
)

// resolveAPIHeaders：AI 若返回空 headers，则从匹配录制记录回填（仅去掉对执行无用的头）。
func resolveAPIHeaders(item AIAPIItem, records []RawRecord) []HeaderKV {
	if usefulHeaders(item.Headers) {
		return sanitizeHeaderKVs(item.Headers)
	}
	rec := matchRecord(records, item)
	if rec == nil {
		return defaultHeadersForMethod(item.Method, item.Body)
	}
	fromRec := headerKVsFromMap(rec.RequestHeaders)
	if len(fromRec) == 0 {
		return defaultHeadersForMethod(item.Method, item.Body)
	}
	return sanitizeHeaderKVs(fromRec)
}

func resolveStepHeaders(step AIScenarioStep, records []RawRecord) []HeaderKV {
	if usefulHeaders(step.Headers) {
		return sanitizeHeaderKVs(step.Headers)
	}
	item := AIAPIItem{Method: step.Method, Path: step.Path}
	rec := matchRecord(records, item)
	if rec == nil {
		return defaultHeadersForMethod(step.Method, step.Body)
	}
	fromRec := headerKVsFromMap(rec.RequestHeaders)
	if len(fromRec) == 0 {
		return defaultHeadersForMethod(step.Method, step.Body)
	}
	return sanitizeHeaderKVs(fromRec)
}

func usefulHeaders(list []HeaderKV) bool {
	for _, h := range list {
		if h.Enabled && strings.TrimSpace(h.Name) != "" {
			return true
		}
	}
	return false
}

func headerKVsFromMap(m map[string]string) []HeaderKV {
	if len(m) == 0 {
		return nil
	}
	out := make([]HeaderKV, 0, len(m))
	for k, v := range m {
		if strings.TrimSpace(k) == "" {
			continue
		}
		out = append(out, HeaderKV{Name: k, Value: v, Enabled: true})
	}
	return out
}

func sanitizeHeaderKVs(list []HeaderKV) []HeaderKV {
	out := make([]HeaderKV, 0, len(list))
	for _, h := range list {
		if !h.Enabled || strings.TrimSpace(h.Name) == "" {
			continue
		}
		name := h.Name
		if shouldDropHeader(name) {
			continue
		}
		val := strings.TrimSpace(h.Value)
		if strings.EqualFold(name, "Authorization") {
			val = normalizeAuthorization(val)
		}
		out = append(out, HeaderKV{Name: name, Value: val, Enabled: true})
	}
	return out
}

// shouldDropHeader 仅丢弃对接口自动化执行无意义的浏览器/传输层头。
func shouldDropHeader(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	switch lower {
	case "cookie",
		"user-agent",
		"accept-encoding",
		"connection",
		"content-length",
		"transfer-encoding",
		"keep-alive",
		"proxy-connection",
		"upgrade",
		"te":
		return true
	}
	if strings.HasPrefix(lower, "sec-ch-ua") || strings.HasPrefix(lower, "sec-fetch-") {
		return true
	}
	if lower == "sec-purpose" || lower == "sec-gpc" {
		return true
	}
	return false
}

func normalizeAuthorization(val string) string {
	if val == "" {
		return "Bearer {{token}}"
	}
	lower := strings.ToLower(val)
	if strings.HasPrefix(lower, "bearer ") && len(val) > 20 {
		return "Bearer {{token}}"
	}
	return val
}

func defaultHeadersForMethod(method, body string) []HeaderKV {
	m := strings.ToUpper(method)
	if body != "" && (m == "POST" || m == "PUT" || m == "PATCH") {
		return []HeaderKV{
			{Name: "Content-Type", Value: "application/json", Enabled: true},
			{Name: "Accept", Value: "application/json", Enabled: true},
		}
	}
	if m == "POST" || m == "PUT" || m == "PATCH" {
		return []HeaderKV{
			{Name: "Content-Type", Value: "application/json", Enabled: true},
			{Name: "Accept", Value: "application/json", Enabled: true},
		}
	}
	return []HeaderKV{{Name: "Accept", Value: "application/json", Enabled: true}}
}
