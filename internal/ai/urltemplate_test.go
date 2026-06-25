package ai

import "testing"

func TestBuildFullURLTemplate(t *testing.T) {
	if got := BuildFullURLTemplate("anchor", "/v1/link"); got != "{{base_url_anchor}}/v1/link" {
		t.Fatalf("anchor: %q", got)
	}
	if got := BuildFullURLTemplate("quest", "/tasks"); got != "{{base_url_quest}}/tasks" {
		t.Fatalf("quest: %q", got)
	}
	if got := BuildFullURLTemplate("", "/health"); got != "{{base_url}}/health" {
		t.Fatalf("default: %q", got)
	}
}

func TestPathOnly(t *testing.T) {
	if got := PathOnly("{{base_url_quest}}/v1/a"); got != "/v1/a" {
		t.Fatalf("got %q", got)
	}
}
