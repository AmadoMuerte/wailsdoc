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

func TestResolveManagerDetectsLockfiles(t *testing.T) {
	cases := []struct {
		name    string
		inRoot  string
		inDir   string
		managed string
		want    string
	}{
		{"bun lock", "bun.lock", "", "", "bun"},
		{"bun lockb", "bun.lockb", "", "", "bun"},
		{"bun fig", "bunfig.toml", "", "", "bun"},
		{"pnpm lock", "pnpm-lock.yaml", "", "", "pnpm"},
		{"pnpm workspace", "pnpm-workspace.yaml", "", "", "pnpm"},
		{"yarn", "yarn.lock", "", "", "yarn"},
		{"npm", "package-lock.json", "", "", "npm"},
		{"no marker", "", "", "", "npm"},
		{"explicit wins", "yarn.lock", "", "pnpm", "pnpm"},
		{"ui dir wins", "", "bun.lock", "", "bun"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			directory := filepath.Join("docs", "site")
			if err := os.MkdirAll(filepath.Join(root, directory), 0o755); err != nil {
				t.Fatal(err)
			}
			if testCase.inRoot != "" {
				if err := os.WriteFile(filepath.Join(root, testCase.inRoot), nil, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if testCase.inDir != "" {
				if err := os.WriteFile(filepath.Join(root, directory, testCase.inDir), nil, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			manager, err := ResolveManager(root, directory, testCase.managed)
			if err != nil {
				t.Fatal(err)
			}
			if manager != testCase.want {
				t.Fatalf("ResolveManager() = %q, want %q", manager, testCase.want)
			}
		})
	}
}

func TestResolveManagerRejectsUnknown(t *testing.T) {
	if _, err := ResolveManager(t.TempDir(), "docs/site", "cargo"); err == nil {
		t.Fatal("unknown package manager was accepted")
	}
}

func TestRunExecutesResolvedManager(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join("docs", "site")
	if err := os.MkdirAll(filepath.Join(root, directory), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, directory, "bun.lock"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	args := filepath.Join(t.TempDir(), "args.txt")
	script := "#!/bin/sh\nprintf '%s' \"$*\" > " + args + "\n"
	if err := os.WriteFile(filepath.Join(bin, "bun"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	if err := Run(root, directory, "build", "auto"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(args)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "run build" {
		t.Fatalf("bun invoked with %q, want %q", got, "run build")
	}
}
