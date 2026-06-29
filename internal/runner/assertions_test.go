package runner

import "testing"

func TestAllPassedEmptyIsFalse(t *testing.T) {
	if AllPassed(nil) {
		t.Fatal("empty assertions must not pass")
	}
	if AllPassed([]AssertionResult{}) {
		t.Fatal("empty assertion results must not pass")
	}
}
