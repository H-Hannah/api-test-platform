package ai

import "testing"

func TestResolveAPIHeadersFromRecord(t *testing.T) {
	records := []RawRecord{{
		Method: "POST",
		Path:   "/v1/foo",
		URL:    "https://api.example.com/v1/foo",
		RequestHeaders: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer secret-token-long-value",
			"Cookie":        "a=b",
			"X-Chain-Id":    "1",
			"Origin":        "https://www.trex.xyz",
		},
	}}
	item := AIAPIItem{Method: "POST", Path: "/v1/foo", Headers: nil}
	got := resolveAPIHeaders(item, records)
	if len(got) < 3 {
		t.Fatalf("expected headers, got %v", got)
	}
	var auth, ct, chain, origin string
	for _, h := range got {
		switch h.Name {
		case "Authorization":
			auth = h.Value
		case "Content-Type":
			ct = h.Value
		case "X-Chain-Id":
			chain = h.Value
		case "Origin":
			origin = h.Value
		}
	}
	if ct != "application/json" || auth != "Bearer {{token}}" || chain != "1" || origin == "" {
		t.Fatalf("ct=%s auth=%s chain=%s origin=%s", ct, auth, chain, origin)
	}
}

func TestShouldDropHeaderKeepsOrigin(t *testing.T) {
	if shouldDropHeader("Origin") {
		t.Fatal("Origin should be kept")
	}
	if !shouldDropHeader("Cookie") {
		t.Fatal("Cookie should drop")
	}
}

func TestDefaultHeadersForGET(t *testing.T) {
	got := defaultHeadersForMethod("GET", "")
	if len(got) != 1 || got[0].Name != "Accept" {
		t.Fatalf("got %v", got)
	}
}
