package docsrepo

import "testing"

func TestLoadTestCasesBrief(t *testing.T) {
	root := "/Users/hanshiqian/cursor/qa-doc-generator"
	if !dirExists(root) {
		t.Skip("qa-doc-generator not at expected path")
	}
	tc, err := LoadTestCases(root, "v2.7.2", "brief")
	if err != nil {
		t.Fatal(err)
	}
	if tc.CaseCount == 0 {
		t.Fatal("expected cases")
	}
}
