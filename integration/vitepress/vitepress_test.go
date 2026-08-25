package vitepress

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitIsConservative(t *testing.T) {
	root := t.TempDir()
	options := Options{Directory: "docs/site", Schema: "docs/api.json", Markdown: "docs/api", Title: "Example API"}
	if err := Init(root, options); err != nil {
		t.Fatal(err)
	}
	config, err := os.ReadFile(filepath.Join(root, options.Directory, ".vitepress", "config.mts"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(config), "Waxlight") || !strings.Contains(string(config), "Example API") {
		t.Fatalf("non-generic config: %s", config)
	}
	if err := Init(root, options); err == nil {
		t.Fatal("second init overwrote existing files")
	}
}
