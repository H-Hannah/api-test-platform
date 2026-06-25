package docsrepo

import "testing"

func TestSortVersions(t *testing.T) {
	got := sortVersions([]string{"v2.7.0", "v2.7.2", "v2.6.6", "v2.8.0"})
	if len(got) != 4 || got[0] != "v2.8.0" || got[len(got)-1] != "v2.6.6" {
		t.Fatalf("unexpected order: %v", got)
	}
}

func TestCompareVersion(t *testing.T) {
	if compareVersion("v2.7.2", "v2.7.1") <= 0 {
		t.Fatal("expected v2.7.2 > v2.7.1")
	}
}
