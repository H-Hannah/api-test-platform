package store

import "testing"

func TestDatasetBelongsToAPI(t *testing.T) {
	api := &APIDefinition{ID: 5, Method: "GET", Path: "/v2/foo"}
	byReq := &TestDataset{RequirementID: "api-5"}
	byBind := &TestDataset{RequirementID: "REQ-1", ApiBindings: `["GET /v2/foo"]`}
	wrongBind := &TestDataset{RequirementID: "REQ-1", ApiBindings: `["GET /v2/bar"]`}
	substr := &TestDataset{RequirementID: "REQ-1", ApiBindings: `["GET /v2"]`}

	if !DatasetBelongsToAPI(byReq, api) {
		t.Fatal("expected requirement_id match")
	}
	if !DatasetBelongsToAPI(byBind, api) {
		t.Fatal("expected binding match")
	}
	if DatasetBelongsToAPI(wrongBind, api) {
		t.Fatal("expected binding mismatch")
	}
	if DatasetBelongsToAPI(substr, api) {
		t.Fatal("substring binding must not match")
	}
}

func TestAPIDefinitionFingerprintChanges(t *testing.T) {
	a := &APIDefinition{Method: "GET", Path: "/a", Headers: "[]", Body: ""}
	b := &APIDefinition{Method: "GET", Path: "/b", Headers: "[]", Body: ""}
	if APIDefinitionFingerprint(a) == APIDefinitionFingerprint(b) {
		t.Fatal("different paths should differ")
	}
}
