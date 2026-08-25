package wailsdoc_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/AmadoMuerte/wailsdoc"
	"github.com/AmadoMuerte/wailsdoc/config"
)

func TestGenericProject(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	fixture := filepath.Join(filepath.Dir(file), "testdata", "basic")
	root := t.TempDir()
	copyFixture(t, fixture, root)
	cfg := config.Defaults()
	cfg.Project.Name = "Basic"
	cfg.Scan.Packages = []string{"./app"}
	first, err := wailsdoc.Generate(context.Background(), root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Controllers) != 1 || len(first.Controllers[0].Methods) != 2 || len(first.Types) != 3 {
		t.Fatalf("unexpected API: %#v", first)
	}
	user := first.Types[2]
	if user.Name != "UserDTO" || len(user.Fields) != 9 || !user.Fields[4].OmitEmpty {
		t.Fatalf("unexpected recursive DTO: %#v", user)
	}
	markdown, err := os.ReadFile(filepath.Join(root, cfg.Output.Markdown, "types", "UserDTO.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"UserDTO is a recursive", "ProfileDTO.md", "UserController.GetUser", "UserController.ListUsers"} {
		if !strings.Contains(string(markdown), expected) {
			t.Fatalf("Markdown lacks %q", expected)
		}
	}
	if err := wailsdoc.Check(context.Background(), root, cfg); err != nil {
		t.Fatal(err)
	}
	schemaPath := filepath.Join(root, cfg.Output.Schema)
	before, _ := os.ReadFile(schemaPath)
	if _, err := wailsdoc.Generate(context.Background(), root, cfg); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(schemaPath)
	if string(before) != string(after) {
		t.Fatal("generation is not deterministic")
	}
}

func copyFixture(t *testing.T, source, target string) {
	t.Helper()
	err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(source, path)
		destination := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destination, contents, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}
