package runner

import (
	"encoding/json"
	"strings"

	"api-test-platform/internal/store"
)

func applyDatasetToRun(ds *store.TestDataset, vars map[string]string, body, headers *string) {
	mergeJSONMapIntoVars(ds.Variables, vars)
	if strings.TrimSpace(ds.BodyOverride) != "" {
		*body = ds.BodyOverride
	}
	if strings.TrimSpace(ds.HeadersOverride) != "" && ds.HeadersOverride != "[]" {
		*headers = mergeHeaderJSON(*headers, ds.HeadersOverride)
	}
}

func mergeJSONMapIntoVars(raw string, vars map[string]string) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return
	}
	m := map[string]string{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return
	}
	for k, v := range m {
		if strings.TrimSpace(k) == "" {
			continue
		}
		vars[k] = v
	}
}

func mergeHeaderJSON(baseRaw, overrideRaw string) string {
	base := parseHeaderList(baseRaw)
	over := parseHeaderList(overrideRaw)
	byName := map[string]headerItem{}
	for _, h := range base {
		if h.Name != "" {
			byName[h.Name] = h
		}
	}
	for _, h := range over {
		if h.Name != "" {
			byName[h.Name] = h
		}
	}
	out := make([]headerItem, 0, len(byName))
	for _, h := range byName {
		out = append(out, h)
	}
	b, _ := json.Marshal(out)
	return string(b)
}

type headerItem struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

func parseHeaderList(raw string) []headerItem {
	var list []headerItem
	_ = json.Unmarshal([]byte(strings.TrimSpace(raw)), &list)
	return list
}
