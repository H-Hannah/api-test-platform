package docsrepo

import (
	"os"
	"testing"
)

func TestLoadFromGitLabLive(t *testing.T) {
	if os.Getenv("GITLAB_TOKEN") == "" {
		t.Skip("GITLAB_TOKEN not set")
	}
	u := "https://gitlab.com/Keccak256-evg/qa/qa-doc-generator/-/tree/beta_20260618_v272/test-docs/v2.7.2/brief?ref_type=heads"
	tc, err := LoadTestCasesFromGitLabURL(u)
	if err != nil {
		t.Fatal(err)
	}
	if tc.CaseCount == 0 {
		t.Fatal("no cases")
	}
	if tc.Version != "v2.7.2" || tc.RequirementID != "brief" {
		t.Fatalf("meta: %s %s", tc.Version, tc.RequirementID)
	}
}
