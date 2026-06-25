package ai

import (
	"strings"
)

func EvaluateTestDataGate(datasets []TestDataDatasetAI, envKeys []string) (bool, []string) {
	var reasons []string
	if len(datasets) == 0 {
		return false, []string{"datasets 为空"}
	}
	seenKey := map[string]bool{}
	for i, ds := range datasets {
		key := strings.TrimSpace(ds.DatasetKey)
		if key == "" {
			reasons = append(reasons, fmtDataReason(i, "缺少 dataset_key"))
		} else if seenKey[key] {
			reasons = append(reasons, fmtDataReason(i, "dataset_key 重复: "+key))
		}
		seenKey[key] = true

		if strings.TrimSpace(ds.Name) == "" {
			reasons = append(reasons, fmtDataReason(i, "缺少 name"))
		}
		ot := strings.TrimSpace(ds.ObtainType)
		if ot != "env" && ot != "fixture" && ot != "manual" && ot != "setup" {
			reasons = append(reasons, fmtDataReason(i, "obtain_type 无效: "+ot))
		}
		if len(ds.TcRefs) == 0 && len(ds.ApiBindings) == 0 {
			reasons = append(reasons, fmtDataReason(i, "tc_refs 与 api_bindings 至少填一项"))
		}
		if ot == "fixture" && strings.TrimSpace(ds.BodyOverride) == "" && len(ds.Variables) == 0 {
			reasons = append(reasons, fmtDataReason(i, "fixture 类型需 body_override 或 variables"))
		}
	}
	if len(envKeys) == 0 {
		reasons = append(reasons, "env_keys 为空，建议至少包含 token 相关键")
	}
	return len(reasons) == 0, reasons
}

func fmtDataReason(idx int, msg string) string {
	return "数据集#" + itoa(idx+1) + ": " + msg
}
