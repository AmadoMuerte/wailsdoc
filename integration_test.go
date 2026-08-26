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
	cfg.Project.Title = "Basic API"
	cfg.Scan.Packages = []string{"./app"}
	first, err := wailsdoc.Generate(context.Background(), root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Controllers) != 1 || len(first.Controllers[0].Methods) != 3 || len(first.Types) != 3 {
		t.Fatalf("unexpected API: %#v", first)
	}
	user := first.Types[2]
	if user.Name != "UserDTO" || len(user.Fields) != 9 || !user.Fields[4].OmitEmpty {
		t.Fatalf("unexpected recursive DTO: %#v", user)
	}
	if user.Fields[0].Description != "ID uniquely identifies the user." || user.Fields[4].TSType != "UserDTO | null" || !user.Fields[4].Nullable {
		t.Fatalf("field documentation or TypeScript metadata missing: %#v", user.Fields)
	}
	controller := first.Controllers[0]
	if controller.Methods[0].Name != "GetUser" || len(controller.Methods[0].Errors) != 1 || controller.Methods[0].Errors[0].Code != "user_not_found" {
		t.Fatalf("method errors missing: %#v", controller.Methods)
	}
	markdown, err := os.ReadFile(filepath.Join(root, cfg.Output.Markdown, "types", "UserDTO.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"UserDTO is a recursive", "ProfileDTO.md", "UserController.GetUser", "UserController.ListUsers", "ID uniquely identifies the user.", "Optional", "Nullable"} {
		if !strings.Contains(string(markdown), expected) {
			t.Fatalf("Markdown lacks %q", expected)
		}
	}
	controllerMarkdown, err := os.ReadFile(filepath.Join(root, cfg.Output.Markdown, "controllers", "UserController.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"GetUser(id: number): Promise<UserDTO | null>", "`user_not_found`", "const getUser = await GetUser(0)", "Ping(): Promise<void>"} {
		if !strings.Contains(string(controllerMarkdown), expected) {
			t.Fatalf("controller Markdown lacks %q", expected)
		}
	}
	index, _ := os.ReadFile(filepath.Join(root, cfg.Output.Markdown, "README.md"))
	if !strings.Contains(string(index), "# Basic API") {
		t.Fatal("project title was not rendered")
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

func TestExternalTypes(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	fixture := filepath.Join(filepath.Dir(file), "testdata", "external")
	root := t.TempDir()
	copyFixture(t, fixture, root)
	copyFixture(t, filepath.Join(filepath.Dir(file), "testdata", "external-domain"), filepath.Join(filepath.Dir(root), "external-domain"))
	cfg := config.Defaults()
	cfg.Scan.Packages = []string{"./app"}
	api, err := wailsdoc.Generate(context.Background(), root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(api.Types) != 3 {
		t.Fatalf("expected three reachable external types, got %#v", api.Types)
	}
	controller := api.Controllers[0]
	if controller.Methods[1].Returns[0].TSType != "foo.Result" {
		t.Fatalf("external TS name missing: %#v", controller.Methods)
	}
	for _, name := range []string{"foo.Result.md", "bar.Result.md", "Data.md"} {
		if _, err := os.Stat(filepath.Join(root, cfg.Output.Markdown, "types", name)); err != nil {
			t.Fatal(err)
		}
	}
	fooPage, err := os.ReadFile(filepath.Join(root, cfg.Output.Markdown, "types", "foo.Result.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"external service", "Data.md", "Data contains the nested response payload", "ExternalController.Foo"} {
		if !strings.Contains(string(fooPage), expected) {
			t.Fatalf("external page lacks %q:\n%s", expected, fooPage)
		}
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
