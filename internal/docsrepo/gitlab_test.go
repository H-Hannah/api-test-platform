package docsrepo

import "testing"

func TestParseGitLabRepoURL_Tree(t *testing.T) {
	raw := "https://gitlab.com/Keccak256-evg/qa/qa-doc-generator/-/tree/beta_20260618_v272/test-docs/v2.7.2/brief?ref_type=heads"
	ref, err := parseGitLabRepoURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	if ref.ProjectPath != "Keccak256-evg/qa/qa-doc-generator" {
		t.Fatalf("project: %s", ref.ProjectPath)
	}
	if ref.Ref != "beta_20260618_v272" {
		t.Fatalf("ref: %s", ref.Ref)
	}
	if ref.RepoPath != "test-docs/v2.7.2/brief" {
		t.Fatalf("path: %s", ref.RepoPath)
	}
	ver, rid := parseTestDocsMeta(ref.RepoPath)
	if ver != "v2.7.2" || rid != "brief" {
		t.Fatalf("meta: %s %s", ver, rid)
	}
}
