// Package wailsdoc composes scanning, validation, and output generation.
package wailsdoc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AmadoMuerte/wailsdoc/bindings"
	"github.com/AmadoMuerte/wailsdoc/config"
	"github.com/AmadoMuerte/wailsdoc/integration/vitepress"
	"github.com/AmadoMuerte/wailsdoc/inventory"
	"github.com/AmadoMuerte/wailsdoc/renderer/markdown"
	"github.com/AmadoMuerte/wailsdoc/scanner"
	"github.com/AmadoMuerte/wailsdoc/schema"
	"github.com/AmadoMuerte/wailsdoc/validation"
)

func Generate(ctx context.Context, root string, cfg config.Config) (schema.API, error) {
	api, err := scanner.Scan(ctx, scanner.Options{Dir: root, Packages: cfg.Scan.Packages, Generator: "wailsdoc"})
	if err != nil {
		return schema.API{}, err
	}
	if err := validation.ForbiddenFields(api, cfg.Validation.ForbiddenFields); err != nil {
		return schema.API{}, err
	}
	if cfg.Wails.Bindings != "" {
		if err := bindings.Check(api, filepath.Join(root, cfg.Wails.Bindings)); err != nil {
			return schema.API{}, err
		}
	}
	if err := writeJSON(filepath.Join(root, cfg.Output.Schema), api); err != nil {
		return schema.API{}, err
	}
	if cfg.Output.Inventory != "" {
		if err := writeJSON(filepath.Join(root, cfg.Output.Inventory), inventory.FromAPI(api)); err != nil {
			return schema.API{}, err
		}
	}
	if _, err := markdown.Render(api, filepath.Join(root, cfg.Output.Markdown)); err != nil {
		return schema.API{}, err
	}
	if cfg.UI.Provider == "vitepress" {
		if err := vitepress.Prepare(root, vitepress.Options{Directory: cfg.UI.Directory, Markdown: cfg.Output.Markdown}); err != nil {
			return schema.API{}, err
		}
	}
	return api, nil
}

func Check(ctx context.Context, root string, cfg config.Config) error {
	temporary, err := os.MkdirTemp("", "wailsdoc-check-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)
	check := cfg
	expectedSchema := filepath.Join(temporary, "schema.json")
	expectedMarkdown := filepath.Join(temporary, "markdown")
	check.Output.Schema, _ = filepath.Rel(root, expectedSchema)
	check.Output.Markdown, _ = filepath.Rel(root, expectedMarkdown)
	expectedInventory := ""
	if cfg.Output.Inventory != "" {
		expectedInventory = filepath.Join(temporary, "inventory.json")
		check.Output.Inventory, _ = filepath.Rel(root, expectedInventory)
	}
	check.UI.Provider = "none"
	if _, err := Generate(ctx, root, check); err != nil {
		return err
	}
	for _, pair := range [][2]string{{expectedSchema, filepath.Join(root, cfg.Output.Schema)}, {expectedMarkdown, filepath.Join(root, cfg.Output.Markdown)}} {
		if equal, err := equalPath(pair[0], pair[1]); err != nil || !equal {
			return fmt.Errorf("generated files are outdated; run wailsdoc generate")
		}
	}
	if cfg.Output.Inventory != "" {
		if equal, err := equalPath(expectedInventory, filepath.Join(root, cfg.Output.Inventory)); err != nil || !equal {
			return fmt.Errorf("generated files are outdated; run wailsdoc generate")
		}
	}
	return nil
}

func writeJSON(path string, value any) error {
	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(contents, '\n'), 0o644)
}

func equalPath(left, right string) (bool, error) {
	leftInfo, err := os.Stat(left)
	if err != nil {
		return false, err
	}
	rightInfo, err := os.Stat(right)
	if err != nil || leftInfo.IsDir() != rightInfo.IsDir() {
		return false, err
	}
	if !leftInfo.IsDir() {
		a, err := os.ReadFile(left)
		if err != nil {
			return false, err
		}
		b, err := os.ReadFile(right)
		return string(a) == string(b), err
	}
	equal := true
	err = filepath.WalkDir(left, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, _ := filepath.Rel(left, path)
		match, err := equalPath(path, filepath.Join(right, relative))
		equal = equal && match
		return err
	})
	if err != nil || !equal {
		return equal, err
	}
	leftEntries, _ := files(left)
	rightEntries, _ := files(right)
	return leftEntries == rightEntries, nil
}

func files(root string) (int, error) {
	count := 0
	err := filepath.WalkDir(root, func(_ string, entry os.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			count++
		}
		return err
	})
	return count, err
}
