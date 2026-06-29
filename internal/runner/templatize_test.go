package runner

import (
	"testing"

	"api-test-platform/internal/store"
)

func TestTemplatizeURL(t *testing.T) {
	vars := map[string]string{
		"edgen_url": "https://api.beta.ospprotocol.xyz",
		"trex_url":  "https://api.trex.beta.dipbit.xyz",
	}
	raw := "https://api.beta.ospprotocol.xyz/v2/platform/bind"
	got := Templatize(raw, vars)
	want := "{{edgen_url}}/v2/platform/bind"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestTemplatizeLongestFirst(t *testing.T) {
	vars := map[string]string{
		"base":     "https://api.example.com",
		"base_api": "https://api.example.com/v1",
	}
	raw := "https://api.example.com/v1/users"
	got := Templatize(raw, vars)
	if got != "{{base_api}}/users" {
		t.Fatalf("got %q", got)
	}
}

func TestSubstituteRoundTrip(t *testing.T) {
	vars := map[string]string{"edgen_url": "https://api.edgen.tech"}
	tpl := "{{edgen_url}}/v2/foo"
	got := substitute(tpl, vars)
	if got != "https://api.edgen.tech/v2/foo" {
		t.Fatalf("got %q", got)
	}
}

func TestTemplatizeFromEnvironments(t *testing.T) {
	envs := []*store.Environment{
		{Variables: `{"edgen_url":"https://api.beta.ospprotocol.xyz"}`},
		{Variables: `{"edgen_url":"https://api.edwealth.ai"}`},
	}
	for _, raw := range []string{
		"https://api.beta.ospprotocol.xyz/v2/platform/bind",
		"https://api.edwealth.ai/v2/platform/bind",
	} {
		got := TemplatizeFromEnvironments(raw, envs)
		want := "{{edgen_url}}/v2/platform/bind"
		if got != want {
			t.Fatalf("raw=%q got %q want %q", raw, got, want)
		}
	}
}
