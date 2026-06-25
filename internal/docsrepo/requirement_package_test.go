package docsrepo

import "testing"

func TestPickPRDFolderBrief(t *testing.T) {
	names := []string{"V2.7.2-Brief页面", "V2.7.2-Plan页面", "other"}
	if got := pickPRDFolder(names, "v2.7.2", "brief"); got != "V2.7.2-Brief页面" {
		t.Fatalf("got %q", got)
	}
}

func TestMatchOSWikiFeaturesBrief(t *testing.T) {
	features := []string{
		"20260506-proactive-tracker",
		"20260408-proactive-agent",
		"20260623-invest-plan",
	}
	matched := matchOSWikiFeatures("brief", features)
	if len(matched) == 0 {
		t.Fatal("expected matches for brief")
	}
}

func TestLoadRequirementPackageBriefIntegration(t *testing.T) {
	if gitLabToken() == "" {
		t.Skip("GITLAB_TOKEN not set")
	}
	pkg, err := LoadRequirementPackage("v2.7.2", "brief", "beta_20260618_v272")
	if err != nil {
		t.Fatal(err)
	}
	if pkg.CaseCount == 0 {
		t.Fatal("expected test cases from qa-doc-generator")
	}
	// PRD/BE may fallback to qa-doc-generator if GitHub token invalid
	if pkg.PRDText == "" && pkg.BeTechText == "" {
		t.Fatal("expected at least prd or be text")
	}
}

func TestClassifyDocFile(t *testing.T) {
	if classifyDocFile("tick-tracker-java-service-design.md", "requirements/v2.7.1/tracker/x.md", "") != "be" {
		t.Fatal("expected be")
	}
	if classifyDocFile("V2.7.2-Brief.md", "x", "technical_spec") != "be" {
		t.Fatal("expected be from source type")
	}
}
