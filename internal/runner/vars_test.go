package runner

import (
	"testing"

	"api-test-platform/internal/store"
)

func TestBuildRunVarsFallbackBaseURL(t *testing.T) {
	env := &store.Environment{
		BaseURL:   "https://api.edgen.tech",
		Variables: `{}`,
	}
	vars := buildRunVars(env)
	if vars["base_url_edgen"] != "https://api.edgen.tech" {
		t.Fatalf("base_url_edgen=%q", vars["base_url_edgen"])
	}
}

func TestSubstituteBaseURLAlias(t *testing.T) {
	vars := map[string]string{"base_url": "https://api.edgen.tech"}
	got := substitute("{{base_url_edgen}}/v2/foo", vars)
	if got != "https://api.edgen.tech/v2/foo" {
		t.Fatalf("got %q", got)
	}
}

func TestValidateUnresolved(t *testing.T) {
	err := validateNoUnresolved("{{token}}", "请求头 Authorization")
	if err == nil {
		t.Fatal("expected error")
	}
}
