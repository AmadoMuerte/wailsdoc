// Package config loads the small wailsdoc.yaml project configuration.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const Filename = "wailsdoc.yaml"

type Config struct {
	Version    int        `yaml:"version"`
	Project    Project    `yaml:"project"`
	Scan       Scan       `yaml:"scan"`
	Output     Output     `yaml:"output"`
	UI         UI         `yaml:"ui"`
	Wails      Wails      `yaml:"wails,omitempty"`
	Validation Validation `yaml:"validation,omitempty"`
}

type Project struct {
	Name  string `yaml:"name"`
	Title string `yaml:"title,omitempty"`
}

type Scan struct {
	Packages []string `yaml:"packages"`
}

type Output struct {
	Schema    string `yaml:"schema"`
	Markdown  string `yaml:"markdown"`
	Inventory string `yaml:"inventory,omitempty"`
}

type UI struct {
	Provider  string `yaml:"provider"`
	Directory string `yaml:"directory,omitempty"`
}

type Wails struct {
	Bindings string `yaml:"bindings,omitempty"`
}

type Validation struct {
	ForbiddenFields []string `yaml:"forbiddenFields,omitempty"`
}

func Defaults() Config {
	return Config{
		Version: 1,
		Project: Project{Name: "Wails App", Title: "Wails API"},
		Scan:    Scan{Packages: []string{"."}},
		Output:  Output{Schema: "docs/generated/wails-api.json", Markdown: "docs/generated/api"},
		UI:      UI{Provider: "none", Directory: "docs/site"},
	}
}

func Load(path string) (Config, error) {
	result := Defaults()
	contents, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	if err := yaml.Unmarshal(contents, &result); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if result.Version != 1 {
		return Config{}, fmt.Errorf("unsupported config version %d", result.Version)
	}
	if len(result.Scan.Packages) == 0 {
		return Config{}, fmt.Errorf("scan.packages must not be empty")
	}
	if result.Output.Schema == "" || result.Output.Markdown == "" {
		return Config{}, fmt.Errorf("output.schema and output.markdown are required")
	}
	if result.UI.Provider != "none" && result.UI.Provider != "vitepress" {
		return Config{}, fmt.Errorf("unsupported UI provider %q", result.UI.Provider)
	}
	return result, nil
}

func Marshal(value Config) ([]byte, error) {
	return yaml.Marshal(value)
}
