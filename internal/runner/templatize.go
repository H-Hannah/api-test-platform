package runner

import (
	"sort"
	"strings"

	"api-test-platform/internal/store"
)

// RunVars 导出运行期环境变量（供入库模板化等场景）。
func RunVars(env *store.Environment) map[string]string {
	if env == nil {
		return map[string]string{}
	}
	return buildRunVars(env)
}

type templatizePair struct {
	key string
	val string
	n   int
}

func pairsFromVars(vars map[string]string) []templatizePair {
	list := make([]templatizePair, 0, len(vars))
	for k, v := range vars {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		list = append(list, templatizePair{k, v, len(v)})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].n > list[j].n })
	return list
}

func envVarKeyRank(k string) int {
	if k == "base_url" {
		return 0
	}
	if strings.HasSuffix(k, "_url") {
		return 2
	}
	return 1
}

func pairsFromEnvironments(envs []*store.Environment) []templatizePair {
	bestKey := map[string]string{}
	bestLen := map[string]int{}
	for _, env := range envs {
		if env == nil {
			continue
		}
		for k, v := range buildRunVars(env) {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			if existing, ok := bestKey[v]; !ok || envVarKeyRank(k) > envVarKeyRank(existing) {
				bestKey[v] = k
				bestLen[v] = len(v)
			}
		}
	}
	list := make([]templatizePair, 0, len(bestKey))
	for v, k := range bestKey {
		list = append(list, templatizePair{k, v, bestLen[v]})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].n > list[j].n })
	return list
}

func applyTemplatizePairs(s string, list []templatizePair) string {
	out := s
	for _, p := range list {
		if strings.Contains(out, p.val) {
			out = strings.ReplaceAll(out, p.val, "{{"+p.key+"}}")
			continue
		}
		trimmed := strings.TrimRight(p.val, "/")
		if trimmed != p.val && strings.Contains(out, trimmed) {
			out = strings.ReplaceAll(out, trimmed, "{{"+p.key+"}}")
		}
	}
	return out
}

// Templatize 扫描 vars 中的值，在 s 里将匹配到的片段替换为 {{key}}（最长优先）。
func Templatize(s string, vars map[string]string) string {
	if s == "" || len(vars) == 0 {
		return s
	}
	return applyTemplatizePairs(s, pairsFromVars(vars))
}

// TemplatizeFromEnvironments 合并多个运行环境的变量值后做最长匹配替换（任意环境录制的 URL 均可识别）。
func TemplatizeFromEnvironments(s string, envs []*store.Environment) string {
	if s == "" || len(envs) == 0 {
		return s
	}
	list := pairsFromEnvironments(envs)
	if len(list) == 0 {
		return s
	}
	return applyTemplatizePairs(s, list)
}
