package docsrepo

import (
	"os"
	"testing"
)

func TestListTestDocsCatalogLocal(t *testing.T) {
	root := "/Users/hanshiqian/cursor/qa-doc-generator"
	if _, err := os.Stat(root); err != nil {
		t.Skip("qa-doc-generator not found")
	}
	os.Setenv("DOCS_REPO_ROOT", root)
	os.Unsetenv("GITLAB_DOCS_PROJECT")
	cat, err := ListTestDocsCatalog("", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Versions) == 0 {
		t.Fatal("expected versions")
	}
	reqs, err := ListTestDocsCatalog(cat.Versions[0], "")
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs.Requirements) == 0 {
		t.Fatalf("expected requirements under %s", cat.Versions[0])
	}
}
