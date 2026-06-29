package runner

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"api-test-platform/internal/store"
)

var (
	varPattern            = regexp.MustCompile(`\{\{(\w+)\}\}`)
	unresolvedVarPattern  = regexp.MustCompile(`\{\{(\w+)\}\}`)
)

// buildRunVars 合并 environments.variables 与 base_url 列。
func buildRunVars(env *store.Environment) map[string]string {
	vars := parseVars(env.Variables)
	base := strings.TrimRight(strings.TrimSpace(env.BaseURL), "/")
	if base == "" {
		for _, key := range []string{
			"edgen_url", "trex_url", "quest_trex_url", "anchor_url", "quest_edgen_url",
			"base_url_edgen", "base_url_trex", "base_url_quest",
			"base_url_anchor", "base_url_openreplay",
		} {
			if v := strings.TrimRight(strings.TrimSpace(vars[key]), "/"); v != "" {
				base = v
				break
			}
		}
	}
	if base != "" {
		vars["base_url"] = base
	}
	delete(vars, "base_url_example")
	return vars
}

func parseVars(raw string) map[string]string {
	m := map[string]string{}
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return m
	}
	// 兼容 variables 存成 JSON 字符串值
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		var nested map[string]any
		if err2 := json.Unmarshal([]byte(raw), &nested); err2 == nil {
			for k, v := range nested {
				m[k] = fmt.Sprint(v)
			}
		}
	}
	return m
}

func substitute(s string, vars map[string]string) string {
	if s == "" {
		return s
	}
	return varPattern.ReplaceAllStringFunc(s, func(m string) string {
		key := strings.Trim(m, "{}")
		if v, ok := vars[key]; ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
		// *_url / base_url_* 未单独配置时，回退到当前环境的 base_url
		if (strings.HasSuffix(key, "_url") || strings.HasPrefix(key, "base_url")) &&
			strings.TrimSpace(vars["base_url"]) != "" {
			return strings.TrimSpace(vars["base_url"])
		}
		return m
	})
}

func validateNoUnresolved(s, label string) error {
	matches := unresolvedVarPattern.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var keys []string
	for _, m := range matches {
		if len(m) < 2 || seen[m[1]] {
			continue
		}
		seen[m[1]] = true
		keys = append(keys, m[1])
	}
	return fmt.Errorf("%s 含未替换变量 {{%s}}，请在当前运行环境的 variables 中配置（JSON 字段）",
		label, strings.Join(keys, "}}, {{"))
}

func ensureFullURLTemplate(tpl, path string) string {
	tpl = strings.TrimSpace(tpl)
	if tpl != "" {
		return tpl
	}
	p := strings.TrimSpace(path)
	if strings.Contains(p, "{{base_url") || strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		return p
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return "{{base_url}}" + p
}
