package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/AmadoMuerte/wailsdoc"
	"github.com/AmadoMuerte/wailsdoc/config"
	"github.com/AmadoMuerte/wailsdoc/integration/vitepress"
)

var version string

func main() {
	if len(os.Args) < 2 {
		help()
		return
	}
	var err error
	switch os.Args[1] {
	case "init":
		err = initCommand(os.Args[2:])
	case "generate":
		err = generateCommand(os.Args[2:], false)
	case "check":
		err = generateCommand(os.Args[2:], true)
	case "serve", "build":
		err = uiCommand(os.Args[1], os.Args[2:])
	case "version", "--version":
		fmt.Println(buildVersion())
	case "help", "--help", "-h":
		help()
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "wailsdoc:", err)
		os.Exit(1)
	}
}

func buildVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func initCommand(args []string) error {
	flags := flag.NewFlagSet("wailsdoc init", flag.ContinueOnError)
	packagePath := flags.String("package", ".", "Wails API Go package")
	provider := flags.String("ui", "none", "documentation UI: none or vitepress")
	name := flags.String("name", "Wails App", "project name")
	force := flags.Bool("force", false, "replace existing files")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *provider != "none" && *provider != "vitepress" {
		return fmt.Errorf("unsupported UI provider %q", *provider)
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(root, config.Filename)
	if !*force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists; use --force to replace it", config.Filename)
		}
	}
	cfg := config.Defaults()
	cfg.Project.Name, cfg.Project.Title = *name, *name+" Wails API"
	cfg.Scan.Packages = []string{*packagePath}
	cfg.UI.Provider = *provider
	contents, err := config.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		return err
	}
	if *provider == "vitepress" {
		return vitepress.Init(root, vitepress.Options{Directory: cfg.UI.Directory, Schema: cfg.Output.Schema, Markdown: cfg.Output.Markdown, Title: cfg.Project.Title, Force: *force})
	}
	return nil
}

func generateCommand(args []string, check bool) error {
	flags := flag.NewFlagSet("wailsdoc generate", flag.ContinueOnError)
	configPath := flags.String("config", config.Filename, "configuration file")
	packages := flags.String("package", "", "override Wails API Go package")
	if err := flags.Parse(args); err != nil {
		return err
	}
	root, path, cfg, err := load(*configPath)
	if err != nil {
		return err
	}
	if *packages != "" {
		cfg.Scan.Packages = []string{*packages}
	}
	if check {
		if err := wailsdoc.Check(context.Background(), root, cfg); err != nil {
			return err
		}
		fmt.Println("WailsDoc is up to date.")
		return nil
	}
	api, err := wailsdoc.Generate(context.Background(), root, cfg)
	if err != nil {
		return err
	}
	fmt.Printf("Generated %d controllers and %d types using %s.\n", len(api.Controllers), len(api.Types), path)
	return nil
}

func uiCommand(command string, args []string) error {
	flags := flag.NewFlagSet("wailsdoc "+command, flag.ContinueOnError)
	configPath := flags.String("config", config.Filename, "configuration file")
	if err := flags.Parse(args); err != nil {
		return err
	}
	root, _, cfg, err := load(*configPath)
	if err != nil {
		return err
	}
	if cfg.UI.Provider != "vitepress" {
		return fmt.Errorf("ui.provider must be vitepress")
	}
	if _, err := wailsdoc.Generate(context.Background(), root, cfg); err != nil {
		return err
	}
	return vitepress.Run(root, cfg.UI.Directory, map[string]string{"serve": "dev", "build": "build"}[command])
}

func load(path string) (string, string, config.Config, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", "", config.Config{}, err
	}
	cfg, err := config.Load(absolute)
	return filepath.Dir(absolute), path, cfg, err
}

func help() {
	fmt.Println(`WailsDoc generates API documentation from Wails v2 Go controllers.

Usage:
  wailsdoc <command> [flags]

Commands:
  init       Initialize wailsdoc.yaml and an optional VitePress site
  generate   Generate schema, inventory, Markdown, and UI content
  check      Fail when generated files differ from source
  serve      Start the configured VitePress development server
  build      Build the configured VitePress site
  version    Print the WailsDoc version`)
}
