package runner

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

type AssertionInput struct {
	Type       string `json:"type"`
	Expression string `json:"expression"`
	Operator   string `json:"operator"`
	Expected   string `json:"expected"`
}

type AssertionResult struct {
	Type       string `json:"type"`
	Expression string `json:"expression"`
	Passed     bool   `json:"passed"`
	Actual     string `json:"actual"`
	Expected   string `json:"expected"`
	Message    string `json:"message,omitempty"`
}

func EvalAssertions(list []AssertionInput, statusCode int, body []byte, durationMS int64) []AssertionResult {
	results := make([]AssertionResult, 0, len(list))
	for _, a := range list {
		results = append(results, evalOne(a, statusCode, body, durationMS))
	}
	return results
}

func evalOne(a AssertionInput, statusCode int, body []byte, durationMS int64) AssertionResult {
	r := AssertionResult{
		Type: a.Type, Expression: a.Expression, Expected: a.Expected,
	}
	op, expected := resolveAssertionOp(a)

	switch a.Type {
	case "status_code":
		r.Actual = strconv.Itoa(statusCode)
		r.Expected = expectedLabel(op, expected)
		r.Passed = compare(op, r.Actual, expected)
	case "duration_ms":
		r.Actual = strconv.FormatInt(durationMS, 10)
		threshold := a.Expression
		if threshold == "" {
			threshold = expected
		}
		if threshold == "" {
			threshold = "3000"
		}
		r.Expected = threshold
		r.Passed = compare(op, r.Actual, threshold)
	case "json_path":
		val := jsonPathGet(body, a.Expression)
		r.Actual = gjsonActualString(val)
		r.Expected = expectedLabel(op, expected)
		r.Passed = compareJSONPath(op, val, expected)
	default:
		r.Passed = false
		r.Message = "unknown assertion type: " + a.Type
	}
	if !r.Passed && r.Message == "" {
		r.Message = fmt.Sprintf("expected %s %s %s, got %s", a.Type, op, r.Expected, r.Actual)
	}
	return r
}

func compare(op, actual, expected string) bool {
	switch normalizeOp(op) {
	case "not_empty", "notempty", "is_not_empty":
		return isNotEmptyString(actual)
	case "empty", "is_empty":
		return !isNotEmptyString(actual)
	case "eq", "==":
		return actual == expected
	case "ne", "!=":
		return actual != expected
	case "gt":
		return toFloat(actual) > toFloat(expected)
	case "gte":
		return toFloat(actual) >= toFloat(expected)
	case "lt":
		return toFloat(actual) < toFloat(expected)
	case "lte":
		return toFloat(actual) <= toFloat(expected)
	case "contains":
		return strings.Contains(actual, expected)
	default:
		return actual == expected
	}
}

func compareJSONPath(op string, val gjson.Result, expected string) bool {
	switch normalizeOp(op) {
	case "not_empty", "notempty", "is_not_empty":
		return gjsonNotEmpty(val)
	case "empty", "is_empty":
		return !gjsonNotEmpty(val)
	case "ne", "!=":
		if strings.TrimSpace(expected) == "" {
			return gjsonNotEmpty(val)
		}
		return compare("ne", gjsonActualString(val), expected)
	default:
		return compare(op, gjsonActualString(val), expected)
	}
}

func normalizeOp(op string) string {
	return strings.ToLower(strings.TrimSpace(op))
}

// resolveAssertionOp 兼容 AI 把 not_empty 写在 expected 字段的情况。
func resolveAssertionOp(a AssertionInput) (op, expected string) {
	op = strings.TrimSpace(a.Operator)
	expected = strings.TrimSpace(a.Expected)
	expOp := normalizeOp(expected)
	if op == "" && (expOp == "not_empty" || expOp == "notempty" || expOp == "is_not_empty" || expOp == "empty" || expOp == "is_empty") {
		op = expected
		expected = ""
	}
	if op == "" {
		op = "eq"
	}
	if normalizeOp(op) == "not_empty" && (expOp == "not_empty" || expOp == "notempty") {
		expected = ""
	}
	return op, expected
}

func expectedLabel(op, expected string) string {
	switch normalizeOp(op) {
	case "not_empty", "notempty", "is_not_empty":
		return "(非空)"
	case "empty", "is_empty":
		return "(为空)"
	default:
		return expected
	}
}

func isNotEmptyString(s string) bool {
	return strings.TrimSpace(s) != ""
}

// gjsonNotEmpty 判断 JSONPath 取值存在且非空（null、""、[]、{} 视为空）。
func gjsonNotEmpty(val gjson.Result) bool {
	if !val.Exists() {
		return false
	}
	switch val.Type {
	case gjson.Null:
		return false
	case gjson.String:
		return strings.TrimSpace(val.String()) != ""
	case gjson.False:
		return true
	case gjson.True, gjson.Number:
		return true
	default:
		raw := strings.TrimSpace(val.Raw)
		if raw == "" || raw == "null" {
			return false
		}
		if raw == "[]" || raw == "{}" {
			return false
		}
		if val.IsArray() {
			return len(val.Array()) > 0
		}
		if val.IsObject() {
			return len(val.Map()) > 0
		}
		return true
	}
}

func gjsonActualString(val gjson.Result) string {
	if !val.Exists() {
		return ""
	}
	if val.Type == gjson.String {
		return val.String()
	}
	raw := strings.TrimSpace(val.Raw)
	if raw == "" {
		return ""
	}
	return raw
}

func toFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func ParseAssertionListJSON(raw string) []AssertionInput {
	if raw == "" {
		return nil
	}
	var list []AssertionInput
	_ = json.Unmarshal([]byte(raw), &list)
	return list
}

func AllPassed(results []AssertionResult) bool {
	if len(results) == 0 {
		return false
	}
	for _, r := range results {
		if !r.Passed {
			return false
		}
	}
	return true
}
