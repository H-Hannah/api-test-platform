package ai

import (
	"encoding/json"
	"testing"
)

func TestAIAssertion_UnmarshalBoolExpected(t *testing.T) {
	raw := `{"type":"json_path","expression":"$.success","operator":"eq","expected":true}`
	var a AIAssertion
	if err := json.Unmarshal([]byte(raw), &a); err != nil {
		t.Fatal(err)
	}
	if a.Expected != "true" {
		t.Fatalf("got %q want true", a.Expected)
	}
}

func TestAIAssertion_UnmarshalNumberExpected(t *testing.T) {
	raw := `{"type":"status_code","expression":"200","operator":"eq","expected":200}`
	var a AIAssertion
	if err := json.Unmarshal([]byte(raw), &a); err != nil {
		t.Fatal(err)
	}
	if a.Expected != "200" {
		t.Fatalf("got %q want 200", a.Expected)
	}
}
