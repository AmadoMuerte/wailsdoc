// Package validation checks optional generated API constraints.
package validation

import (
	"fmt"
	"strings"

	"github.com/AmadoMuerte/wailsdoc/schema"
)

func ForbiddenFields(api schema.API, forbidden []string) error {
	for _, typ := range api.Types {
		for _, field := range typ.Fields {
			candidate := strings.ToLower(typ.Name + " " + field.Name + " " + field.JSONName)
			for _, name := range forbidden {
				if name != "" && strings.Contains(candidate, strings.ToLower(name)) {
					return fmt.Errorf("type %s exposes forbidden field %s", typ.Name, field.Name)
				}
			}
		}
	}
	return nil
}
