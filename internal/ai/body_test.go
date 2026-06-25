package ai

import "testing"

func TestResolveAPIBodyFromRecord(t *testing.T) {
	records := []RawRecord{{
		Method:      "POST",
		Path:        "/v1/foo",
		URL:         "https://api.example.com/v1/foo",
		RequestBody: `{"page":1}`,
	}}
	item := AIAPIItem{Method: "POST", Path: "/v1/foo", Body: ""}
	got := resolveAPIBody(item, records)
	if got != `{"page":1}` {
		t.Fatalf("got %q", got)
	}
}
