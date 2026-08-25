package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	expected := Defaults()
	expected.Scan.Packages = []string{"./backend"}
	contents, err := Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), Filename)
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatal(err)
	}
	actual, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if actual.Scan.Packages[0] != "./backend" || actual.Output.Markdown == "" {
		t.Fatalf("unexpected config: %#v", actual)
	}
}
