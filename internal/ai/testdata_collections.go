package ai

import (
	"fmt"
	"strings"
)

// flattenTestDataCollections 将 collections 展开为平铺 datasets，并补全 collection 元数据。
func flattenTestDataCollections(result *TestDataGenerateAIResult) []TestDataCollectionAI {
	if len(result.Collections) == 0 {
		if len(result.Datasets) == 0 {
			return nil
		}
		return []TestDataCollectionAI{{
			CollectionKey: "default",
			Name:          "默认测试集",
			Datasets:      result.Datasets,
		}}
	}
	var flat []TestDataDatasetAI
	collections := result.Collections
	for ci := range collections {
		coll := &collections[ci]
		if strings.TrimSpace(coll.CollectionKey) == "" {
			coll.CollectionKey = fmt.Sprintf("col-%d", ci+1)
		}
		if strings.TrimSpace(coll.Name) == "" {
			coll.Name = coll.CollectionKey + " 测试集"
		}
		for di := range coll.Datasets {
			ds := &coll.Datasets[di]
			if strings.TrimSpace(ds.CollectionKey) == "" {
				ds.CollectionKey = coll.CollectionKey
			}
			if strings.TrimSpace(ds.CollectionName) == "" {
				ds.CollectionName = coll.Name
			}
			if strings.TrimSpace(ds.DatasetKey) == "" {
				ds.DatasetKey = fmt.Sprintf("%s-%03d", coll.CollectionKey, di+1)
			}
			ds.Tags = ensureCollectionTag(ds.Tags, coll.CollectionKey)
			flat = append(flat, *ds)
		}
	}
	result.Datasets = flat
	return collections
}

func ensureCollectionTag(tags []string, collectionKey string) []string {
	prefix := "collection:" + collectionKey
	for _, t := range tags {
		if t == prefix {
			return tags
		}
	}
	return append(append([]string{}, tags...), prefix)
}

func BuildTestDataStats(datasets []TestDataDatasetAI, envKeys []string, collections []TestDataCollectionAI) TestDataStats {
	st := TestDataStats{
		TotalDatasets:    len(datasets),
		TotalCollections: len(collections),
		ByObtain:         map[string]int{},
		EnvKeyCount:      len(envKeys),
	}
	if st.TotalCollections == 0 && len(datasets) > 0 {
		st.TotalCollections = 1
	}
	for _, ds := range datasets {
		ot := strings.TrimSpace(ds.ObtainType)
		if ot == "" {
			ot = "unknown"
		}
		st.ByObtain[ot]++
	}
	return st
}
