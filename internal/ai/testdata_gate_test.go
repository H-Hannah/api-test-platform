package ai

import "testing"

func TestEvaluateTestDataGate(t *testing.T) {
	ok, reasons := EvaluateTestDataGate([]TestDataDatasetAI{
		{
			DatasetKey:  "DS001",
			Name:        "正常",
			TcRefs:      []string{"TC001"},
			ApiBindings: []string{"GET /v2/foo"},
			ObtainType:  "env",
			Variables:   map[string]string{"id": "{{user_id}}"},
		},
	}, []string{"token", "user_id"})
	if !ok {
		t.Fatalf("expected pass, got %v", reasons)
	}

	_, reasons = EvaluateTestDataGate(nil, nil)
	if len(reasons) == 0 {
		t.Fatal("expected failure for empty datasets")
	}
}
