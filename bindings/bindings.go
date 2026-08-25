// Package bindings validates Wails v2 JavaScript controller bindings.
package bindings

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AmadoMuerte/wailsdoc/schema"
)

func Check(api schema.API, directory string) error {
	known := map[string]bool{}
	for _, controller := range api.Controllers {
		known[controller.Name] = true
		path := filepath.Join(directory, controller.Name+".js")
		contents, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("controller %s has no binding: %w", controller.Name, err)
		}
		for _, method := range controller.Methods {
			if !regexp.MustCompile(`export function ` + regexp.QuoteMeta(method.Name) + `\b`).Match(contents) {
				return fmt.Errorf("binding %s is missing method %s", controller.Name, method.Name)
			}
		}
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".js") && !known[strings.TrimSuffix(entry.Name(), ".js")] {
			return fmt.Errorf("binding %s has no matching controller", entry.Name())
		}
	}
	return nil
}
