package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

func RenderTestDataYAML(version, requirementID, requirementName string, envKeys []string, collections []TestDataCollectionAI, datasets []TestDataDatasetAI, notes string) string {
	var b strings.Builder
	b.WriteString("# 测试数据规格\n")
	b.WriteString("version: " + version + "\n")
	b.WriteString("requirement_id: " + requirementID + "\n")
	b.WriteString("requirement_name: " + requirementName + "\n\n")

	b.WriteString("env_keys:\n")
	for _, k := range envKeys {
		b.WriteString("  - " + k + "\n")
	}
	if notes != "" {
		b.WriteString("\ncoverage_notes: |\n")
		for _, line := range strings.Split(notes, "\n") {
			b.WriteString("  " + line + "\n")
		}
	}

	if len(collections) > 0 {
		b.WriteString("\ncollections:\n")
		for _, coll := range collections {
			b.WriteString("  - collection_key: " + coll.CollectionKey + "\n")
			b.WriteString("    name: " + yamlQuote(coll.Name) + "\n")
			if coll.Description != "" {
				b.WriteString("    description: " + yamlQuote(coll.Description) + "\n")
			}
			b.WriteString("    datasets:\n")
			for _, ds := range coll.Datasets {
				writeDatasetYAML(&b, ds, "      ")
			}
		}
		return b.String()
	}

	b.WriteString("\ndatasets:\n")
	for _, ds := range datasets {
		writeDatasetYAML(&b, ds, "  ")
	}
	return b.String()
}

func writeDatasetYAML(b *strings.Builder, ds TestDataDatasetAI, indent string) {
	b.WriteString(indent + "- dataset_key: " + ds.DatasetKey + "\n")
	b.WriteString(indent + "  name: " + yamlQuote(ds.Name) + "\n")
	if ds.CollectionKey != "" {
		b.WriteString(indent + "  collection_key: " + ds.CollectionKey + "\n")
	}
	if ds.Description != "" {
		b.WriteString(indent + "  description: " + yamlQuote(ds.Description) + "\n")
	}
	if len(ds.TcRefs) > 0 {
		b.WriteString(indent + "  tc_refs:\n")
		for _, t := range ds.TcRefs {
			b.WriteString(indent + "    - " + t + "\n")
		}
	}
	if len(ds.ApiBindings) > 0 {
		b.WriteString(indent + "  api_bindings:\n")
		for _, a := range ds.ApiBindings {
			b.WriteString(indent + "    - " + yamlQuote(a) + "\n")
		}
	}
	if len(ds.Variables) > 0 {
		b.WriteString(indent + "  variables:\n")
		for k, v := range ds.Variables {
			b.WriteString(indent + "    " + k + ": " + yamlQuote(v) + "\n")
		}
	}
	if ds.BodyOverride != "" {
		b.WriteString(indent + "  body_override: " + yamlQuote(ds.BodyOverride) + "\n")
	}
	b.WriteString(indent + "  obtain_type: " + ds.ObtainType + "\n")
	if ds.ObtainNote != "" {
		b.WriteString(indent + "  obtain_note: " + yamlQuote(ds.ObtainNote) + "\n")
	}
	b.WriteString(indent + "  owner: " + ds.Owner + "\n")
	if len(ds.Tags) > 0 {
		b.WriteString(indent + "  tags: [" + strings.Join(ds.Tags, ", ") + "]\n")
	}
	b.WriteString("\n")
}

func yamlQuote(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "\"\""
	}
	if strings.ContainsAny(s, ":\n\"'#") || strings.HasPrefix(s, "{") {
		escaped := strings.ReplaceAll(s, "\"", "\\\"")
		return "\"" + escaped + "\""
	}
	return s
}

func SuggestTestDataGitPath(version, requirementID string) string {
	return fmt.Sprintf("test-data/%s/%s/data-spec.yaml", version, requirementID)
}

func datasetAIToStore(productID int64, version, requirementID string, ds TestDataDatasetAI) *storeDataset {
	tcRefs, _ := json.Marshal(ds.TcRefs)
	apiBindings, _ := json.Marshal(ds.ApiBindings)
	vars, _ := json.Marshal(ds.Variables)
	if len(ds.Variables) == 0 {
		vars = []byte("{}")
	}
	headers, _ := json.Marshal(ds.HeadersOverride)
	if len(ds.HeadersOverride) == 0 {
		headers = []byte("[]")
	}
	tags, _ := json.Marshal(ds.Tags)
	owner := strings.TrimSpace(ds.Owner)
	if owner == "" {
		owner = "qa"
	}
	obtain := strings.TrimSpace(ds.ObtainType)
	if obtain == "" {
		obtain = "env"
	}
	return &storeDataset{
		ProductID:       productID,
		Version:         version,
		RequirementID:   requirementID,
		DatasetKey:      strings.TrimSpace(ds.DatasetKey),
		Name:            strings.TrimSpace(ds.Name),
		Description:     strings.TrimSpace(ds.Description),
		TcRefs:          string(tcRefs),
		ApiBindings:     string(apiBindings),
		Variables:       string(vars),
		HeadersOverride: string(headers),
		BodyOverride:    strings.TrimSpace(ds.BodyOverride),
		ObtainType:      obtain,
		ObtainNote:      strings.TrimSpace(ds.ObtainNote),
		Owner:           owner,
		Tags:            string(tags),
		Source:          "ai",
	}
}

// storeDataset mirrors store.TestDataset for import without circular import in yaml file.
type storeDataset struct {
	ProductID       int64
	Version         string
	RequirementID   string
	DatasetKey      string
	Name            string
	Description     string
	TcRefs          string
	ApiBindings     string
	Variables       string
	HeadersOverride string
	BodyOverride    string
	ObtainType      string
	ObtainNote      string
	Owner           string
	Tags            string
	Source          string
}
