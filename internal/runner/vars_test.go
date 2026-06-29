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
	if vars["base_url"] != "https://api.edgen.tech" {
		t.Fatalf("base_url=%q", vars["base_url"])
	}
}

func TestSubstituteURLVar(t *testing.T) {
	vars := map[string]string{"edgen_url": "https://api.edgen.tech", "base_url": "https://api.edgen.tech"}
	got := substitute("{{edgen_url}}/v2/foo", vars)
	if got != "https://api.edgen.tech/v2/foo" {
		t.Fatalf("got %q", got)
	}
}

func TestSubstituteURLFallback(t *testing.T) {
	vars := map[string]string{"base_url": "https://api.edgen.tech"}
	got := substitute("{{trex_url}}/v2/foo", vars)
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
