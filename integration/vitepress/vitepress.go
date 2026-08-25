// Package vitepress initializes and prepares the optional VitePress UI.
package vitepress

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Options struct {
	Directory string
	Schema    string
	Markdown  string
	Title     string
	Force     bool
}

func Init(root string, options Options) error {
	files := map[string]string{
		"package.json": packageJSON,
		"index.md":     strings.ReplaceAll(indexMarkdown, "{{TITLE}}", options.Title),
		filepath.Join(".vitepress", "config.mts"): strings.NewReplacer(
			"{{TITLE}}", options.Title,
			"{{SCHEMA}}", relative(filepath.Join(root, options.Directory, ".vitepress"), filepath.Join(root, options.Schema)),
		).Replace(configMTS),
	}
	for name, contents := range files {
		path := filepath.Join(root, options.Directory, name)
		if !options.Force {
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("%s already exists; use --force to replace it", path)
			}
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func Prepare(root string, options Options) error {
	target := filepath.Join(root, options.Directory, "api")
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	return copyDir(filepath.Join(root, options.Markdown), target)
}

func Run(root, directory, command string) error {
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm is required for VitePress %s", command)
	}
	cmd := exec.Command("npm", "run", command)
	cmd.Dir = filepath.Join(root, directory)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

func relative(from, to string) string {
	value, _ := filepath.Rel(from, to)
	return filepath.ToSlash(value)
}

func copyDir(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
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
}

const packageJSON = `{
  "name": "wailsdoc-site",
  "private": true,
  "type": "module",
  "scripts": {
    "build": "vitepress build .",
    "dev": "vitepress dev .",
    "preview": "vitepress preview ."
  },
  "devDependencies": {
    "vitepress": "1.6.4"
  }
}
`

const indexMarkdown = `---
layout: home
hero:
  name: {{TITLE}}
  text: Wails v2 API reference
  tagline: Generated directly from Go source.
  actions:
    - theme: brand
      text: Browse Controllers
      link: /api/README#controllers
    - theme: alt
      text: Browse Methods
      link: /api/METHODS
    - theme: alt
      text: Browse Types
      link: /api/README#types
---
`

const configMTS = `import { readFileSync } from "node:fs";
import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vitepress";

const api = JSON.parse(readFileSync(fileURLToPath(new URL("{{SCHEMA}}", import.meta.url)), "utf8"));
const controllers = api.controllers.map(({ name }) => ({ text: name, link: ` + "`/api/controllers/${name}`" + ` }));
const types = api.types.map(({ name }) => ({ text: name, link: ` + "`/api/types/${name}`" + ` }));

export default defineConfig({
  title: "{{TITLE}}",
  cleanUrls: true,
  themeConfig: {
    nav: [
      { text: "Home", link: "/" },
      { text: "API Reference", link: "/api/README" },
      { text: "Methods", link: "/api/METHODS" }
    ],
    sidebar: { "/api/": [
      { text: "Overview", link: "/api/README" },
      { text: "Methods", link: "/api/METHODS" },
      { text: "Controllers", items: controllers },
      { text: "Types", collapsed: true, items: types }
    ] },
    search: { provider: "local" }
  }
});
`
