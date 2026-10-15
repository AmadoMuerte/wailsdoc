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
	if actual.UI.PackageManager != "auto" {
		t.Fatalf("package manager default = %q, want auto", actual.UI.PackageManager)
	}
}

func TestPackageManagerValidation(t *testing.T) {
	for _, value := range PackageManagers {
		cfg := Defaults()
		cfg.UI.PackageManager = value
		contents, err := Marshal(cfg)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), Filename)
		if err := os.WriteFile(path, contents, 0o644); err != nil {
			t.Fatal(err)
		}
		if loaded, err := Load(path); err != nil || loaded.UI.PackageManager != value {
			t.Fatalf("Load(%q) = %q, %v", value, loaded.UI.PackageManager, err)
		}
	}
	cfg := Defaults()
	cfg.UI.PackageManager = "cargo"
	contents, err := Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), Filename)
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("unknown ui.packageManager was accepted")
	}
}
