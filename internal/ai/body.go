package ai

import (
	"encoding/json"
	"strings"
)

// resolveAPIBody：AI 未填 body 时从录制 requestBody 回填。
func resolveAPIBody(item AIAPIItem, records []RawRecord) string {
	if b := strings.TrimSpace(item.Body); b != "" && b != "{}" && b != "null" {
		return normalizeBodyString(b)
	}
	rec := matchRecord(records, item)
	if rec == nil || strings.TrimSpace(rec.RequestBody) == "" {
		return item.Body
	}
	return normalizeBodyString(rec.RequestBody)
}

func resolveStepBody(step AIScenarioStep, records []RawRecord) string {
	if b := strings.TrimSpace(step.Body); b != "" && b != "{}" {
		return normalizeBodyString(b)
	}
	item := AIAPIItem{Method: step.Method, Path: step.Path}
	rec := matchRecord(records, item)
	if rec == nil || strings.TrimSpace(rec.RequestBody) == "" {
		return step.Body
	}
	return normalizeBodyString(rec.RequestBody)
}

func normalizeBodyString(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if raw[0] == '{' || raw[0] == '[' {
		var v any
		if json.Unmarshal([]byte(raw), &v) == nil {
			b, err := json.Marshal(v)
			if err == nil {
				return string(b)
			}
		}
	}
	return raw
}
