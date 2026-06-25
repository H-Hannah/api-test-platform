package runner

import "testing"

func TestNormalizeJSONPathExpr(t *testing.T) {
	tests := []struct{ in, want string }{
		{"$.body.code", "$.code"},
		{"code", "$.code"},
		{"$.code", "$.code"},
	}
	for _, tc := range tests {
		if got := NormalizeJSONPathExpr(tc.in); got != tc.want {
			t.Fatalf("NormalizeJSONPathExpr(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestJsonPathGet(t *testing.T) {
	body := []byte(`{"code":0,"data":{"id":"x","rows":[{"id":"a"}]}}`)
	if v := jsonPathGet(body, "$.body.code"); !v.Exists() || v.String() != "0" {
		t.Fatalf("fallback $.body.code: %v", v)
	}
	if v := jsonPathGet(body, "$.code"); v.String() != "0" {
		t.Fatalf("$.code: %v", v)
	}
	if v := jsonPathGet(body, "$.data.rows[0].id"); !v.Exists() || v.String() != "a" {
		t.Fatalf("$.data.rows[0].id: %v", v)
	}
}

func TestBuildResponseSnapshotJSON(t *testing.T) {
	snap := buildResponseSnapshot(200, []byte(`{"code":0}`), 1024)
	if snap["json"] == nil {
		t.Fatalf("expected json field, got %v", snap)
	}
	if _, hasBody := snap["body"]; hasBody {
		t.Fatalf("json response should not use string body wrapper: %v", snap)
	}
}
