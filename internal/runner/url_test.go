package runner

import "testing"

func TestResolveRequestURL(t *testing.T) {
	vars := map[string]string{
		"base_url":        "https://api.trex.xyz",
		"base_url_trex":   "https://api.trex.xyz",
		"base_url_quest":  "https://quest.trex.xyz",
		"base_url_anchor": "https://anchor.trex.xyz",
	}

	cases := []struct {
		path, tpl, want string
	}{
		{"/v1/a", "", "https://api.trex.xyz/v1/a"},
		{"/v1/a", "{{base_url_quest}}/v1/a", "https://quest.trex.xyz/v1/a"},
		{"{{base_url_anchor}}/v1/link", "", "https://anchor.trex.xyz/v1/link"},
		{"", "{{base_url_edgen}}/v2/bind", "https://api.trex.xyz/v2/bind"},
		{"", "{{base_url_trex}}/health", "https://api.trex.xyz/health"},
	}
	for _, c := range cases {
		got := ResolveRequestURL(c.path, c.tpl, vars)
		if got != c.want {
			t.Errorf("ResolveRequestURL(%q,%q)=%q want %q", c.path, c.tpl, got, c.want)
		}
	}
}
