package ai

import (
	"testing"

	"api-test-platform/internal/store"
)

func TestPathOnly(t *testing.T) {
	if got := PathOnly("{{quest_trex_url}}/v1/a"); got != "/v1/a" {
		t.Fatalf("got %q", got)
	}
	if got := PathOnly("{{edgen_url}}/v2/foo"); got != "/v2/foo" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildURLTemplate(t *testing.T) {
	envs := []*store.Environment{{
		Variables: `{"edgen_url":"https://api.beta.ospprotocol.xyz"}`,
	}}
	records := []RawRecord{{
		URL:    "https://api.beta.ospprotocol.xyz/v2/bind",
		Method: "GET",
		Path:   "/v2/bind",
	}}
	item := AIAPIItem{Method: "GET", Path: "/v2/bind"}
	full, path := buildURLTemplate(records, item, envs)
	if full != "{{edgen_url}}/v2/bind" {
		t.Fatalf("full=%q", full)
	}
	if path != "/v2/bind" {
		t.Fatalf("path=%q", path)
	}
}

func TestBuildURLTemplateMultiEnv(t *testing.T) {
	envs := []*store.Environment{
		{Variables: `{"edgen_url":"https://api.beta.ospprotocol.xyz"}`},
		{Variables: `{"edgen_url":"https://api.edwealth.ai"}`},
	}
	for _, raw := range []string{
		"https://api.beta.ospprotocol.xyz/v2/bind",
		"https://api.edwealth.ai/v2/bind",
	} {
		records := []RawRecord{{URL: raw, Method: "GET", Path: "/v2/bind"}}
		item := AIAPIItem{Method: "GET", Path: "/v2/bind"}
		full, path := buildURLTemplate(records, item, envs)
		if full != "{{edgen_url}}/v2/bind" {
			t.Fatalf("raw=%q full=%q", raw, full)
		}
		if path != "/v2/bind" {
			t.Fatalf("raw=%q path=%q", raw, path)
		}
	}
}
