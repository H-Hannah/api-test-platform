package runner

import "testing"

func TestJsonPathNotEmpty(t *testing.T) {
	body := []byte(`{"code":0,"data":{"token":"abc"},"items":[],"empty":""}`)
	tests := []struct {
		expr string
		want bool
	}{
		{"$.data.token", true},
		{"$.code", true},
		{"$.empty", false},
		{"$.items", false},
		{"$.missing", false},
	}
	for _, tc := range tests {
		a := AssertionInput{Type: "json_path", Expression: tc.expr, Operator: "not_empty"}
		r := evalOne(a, 200, body, 10)
		if r.Passed != tc.want {
			t.Fatalf("%s not_empty: passed=%v actual=%q", tc.expr, r.Passed, r.Actual)
		}
	}
}

func TestJsonPathBracketIndex(t *testing.T) {
	body := []byte(`{"data":{"rows":[{"id":"abc","theme":"t","summary":"s"}]}}`)
	for _, expr := range []string{"$.data.rows[0].id", "$.data.rows[0].theme"} {
		a := AssertionInput{Type: "json_path", Expression: expr, Operator: "not_empty"}
		r := evalOne(a, 200, body, 10)
		if !r.Passed {
			t.Fatalf("%s: passed=%v actual=%q msg=%s", expr, r.Passed, r.Actual, r.Message)
		}
	}
}

func TestNotEmptyInExpectedField(t *testing.T) {
	body := []byte(`{"data":{"token":"xyz"}}`)
	a := AssertionInput{Type: "json_path", Expression: "$.data.token", Operator: "", Expected: "not_empty"}
	r := evalOne(a, 200, body, 10)
	if !r.Passed || r.Expected != "(非空)" {
		t.Fatalf("got passed=%v expected=%q", r.Passed, r.Expected)
	}
}

func TestJsonPathNeEmptyString(t *testing.T) {
	body := []byte(`{"data":{"token":"x"}}`)
	a := AssertionInput{Type: "json_path", Expression: "$.data.token", Operator: "ne", Expected: ""}
	r := evalOne(a, 200, body, 10)
	if !r.Passed {
		t.Fatalf("ne empty expected should pass when token present: %v", r)
	}
}
