package ai

import "testing"

func TestFlattenTestDataCollections(t *testing.T) {
	result := TestDataGenerateAIResult{
		Collections: []TestDataCollectionAI{
			{
				CollectionKey: "tracker",
				Name:          "Tracker 测试集",
				Datasets: []TestDataDatasetAI{
					{Name: "正常", TcRefs: []string{"TC1"}, ApiBindings: []string{"GET /v2/t"}, ObtainType: "env"},
				},
			},
			{
				CollectionKey: "brief",
				Name:          "Brief 测试集",
				Datasets: []TestDataDatasetAI{
					{Name: "卡片流", TcRefs: []string{"TC2"}, ApiBindings: []string{"GET /v2/brief"}, ObtainType: "fixture", Variables: map[string]string{"x": "1"}},
				},
			},
		},
	}
	colls := flattenTestDataCollections(&result)
	if len(colls) != 2 {
		t.Fatalf("expected 2 collections, got %d", len(colls))
	}
	if len(result.Datasets) != 2 {
		t.Fatalf("expected 2 flat datasets, got %d", len(result.Datasets))
	}
	if result.Datasets[0].DatasetKey != "tracker-001" {
		t.Fatalf("unexpected key %s", result.Datasets[0].DatasetKey)
	}
}
