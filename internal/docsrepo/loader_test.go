package docsrepo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPackage_sourcesJSON(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "qa-doc-generator")
	if _, err := os.Stat(root); err != nil {
		t.Skip("qa-doc-generator not found at", root)
	}
	pkg, err := LoadPackage(root, "requirements/v2.7.0/chat-voice-input")
	if err != nil {
		t.Fatal(err)
	}
	if pkg.PRDText == "" {
		t.Fatal("expected prd text")
	}
	if pkg.UIDesignText == "" {
		t.Fatal("expected ui design text")
	}
}
