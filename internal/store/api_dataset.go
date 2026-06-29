package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// ParseAPIRequirementID 从 api-{id} 解析接口 ID。
func ParseAPIRequirementID(rid string) int64 {
	rid = strings.TrimSpace(rid)
	var id int64
	if _, err := fmt.Sscanf(rid, "api-%d", &id); err != nil {
		return 0
	}
	return id
}

func APIRequirementID(apiID int64) string {
	return fmt.Sprintf("api-%d", apiID)
}

// APIBinding 返回 METHOD path 绑定键。
func APIBinding(api *APIDefinition) string {
	if api == nil {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(api.Method)) + " " + strings.TrimSpace(api.Path)
}

// APIDefinitionFingerprint 接口请求模板的稳定指纹（method/path/headers/body/url）。
func APIDefinitionFingerprint(api *APIDefinition) string {
	if api == nil {
		return ""
	}
	payload := strings.Join([]string{
		strings.ToUpper(strings.TrimSpace(api.Method)),
		strings.TrimSpace(api.Path),
		strings.TrimSpace(api.FullURLTemplate),
		strings.TrimSpace(api.Headers),
		strings.TrimSpace(api.Body),
	}, "\n")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:16])
}

// DatasetBelongsToAPI 判断数据集是否归属该接口（requirement_id 或精确 api_bindings）。
func DatasetBelongsToAPI(ds *TestDataset, api *APIDefinition) bool {
	if ds == nil || api == nil {
		return false
	}
	if ds.RequirementID == APIRequirementID(api.ID) {
		return true
	}
	return datasetMatchesBinding(*ds, APIBinding(api))
}

func datasetMatchesBinding(ds TestDataset, binding string) bool {
	binding = strings.TrimSpace(binding)
	if binding == "" {
		return false
	}
	var bindings []string
	_ = json.Unmarshal([]byte(ds.ApiBindings), &bindings)
	for _, b := range bindings {
		if strings.EqualFold(strings.TrimSpace(b), binding) {
			return true
		}
	}
	return false
}

// DatasetStaleAgainstAPI 用例相对当前接口定义是否过期。
func DatasetStaleAgainstAPI(api *APIDefinition, ds *TestDataset) (bool, string) {
	if api == nil || ds == nil {
		return false, ""
	}
	fp := APIDefinitionFingerprint(api)
	if ds.ApiFingerprint != "" {
		if ds.ApiFingerprint != fp {
			return true, "接口定义已变更，用例尚未同步"
		}
		return false, ""
	}
	// 兼容旧数据：无指纹时回退时间比较
	if datasetStaleAgainstAPI(api.UpdatedAt, ds.UpdatedAt) {
		return true, "接口定义已更新，用例尚未同步"
	}
	return false, ""
}
