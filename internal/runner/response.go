package runner

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

var jsonPathIndexBracket = regexp.MustCompile(`\[(\d+)\]`)

// buildResponseSnapshot 保存执行响应用于展示；JSON 响应解析到 json 字段，避免与 $.body 路径混淆。
func buildResponseSnapshot(statusCode int, respBody []byte, max int) map[string]any {
	snap := map[string]any{
		"status":      statusCode,
		"status_code": statusCode,
	}
	if len(respBody) == 0 {
		return snap
	}
	body := respBody
	truncated := false
	if len(body) > max {
		body = body[:max]
		truncated = true
	}
	if json.Valid(body) {
		var parsed any
		if err := json.Unmarshal(body, &parsed); err == nil {
			snap["json"] = parsed
			if truncated {
				snap["truncated"] = true
			}
			return snap
		}
	}
	text := string(body)
	if truncated {
		text += "...(truncated)"
		snap["truncated"] = true
	}
	snap["body"] = text
	return snap
}

// toGJSONPath 将 JSONPath 风格（$.code、$.data.id）转为 gjson 路径（code、data.id）。
func toGJSONPath(expr string) string {
	p := strings.TrimSpace(expr)
	p = strings.TrimPrefix(p, "$.")
	if p == "$" || p == "" {
		return ""
	}
	p = strings.TrimPrefix(p, "$")
	// 兼容误把平台快照字段 body 当作根：$.body.code -> code
	if strings.HasPrefix(p, "body.") {
		p = strings.TrimPrefix(p, "body.")
	}
	// gjson 使用点号下标：data.rows.0.id，不是 data.rows[0].id
	return jsonPathIndexBracket.ReplaceAllString(p, ".$1")
}

// jsonPathGet 在 HTTP 响应体上求值。
func jsonPathGet(body []byte, expr string) gjson.Result {
	primary := toGJSONPath(expr)
	if primary == "" {
		return gjson.Result{}
	}
	val := gjson.GetBytes(body, primary)
	if val.Exists() {
		return val
	}
	// 再尝试去掉 $.body. 后的路径
	if strings.HasPrefix(strings.TrimSpace(expr), "$.body.") {
		alt := toGJSONPath("$" + strings.TrimPrefix(strings.TrimSpace(expr), "$.body"))
		if alt != "" && alt != primary {
			if alt2 := gjson.GetBytes(body, alt); alt2.Exists() {
				return alt2
			}
		}
	}
	return val
}

// NormalizeJSONPathExpr 规范化入库表达式（统一为 $. 前缀便于展示，执行时再 toGJSONPath）。
func NormalizeJSONPathExpr(expr string) string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return expr
	}
	p := toGJSONPath(expr)
	if p == "" {
		return expr
	}
	return "$." + p
}
