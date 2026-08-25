// Package inventory creates a compact controller and method inventory.
package inventory

import "github.com/AmadoMuerte/wailsdoc/schema"

type Inventory struct {
	SchemaVersion int                 `json:"schemaVersion"`
	Controllers   map[string][]string `json:"controllers"`
}

func FromAPI(api schema.API) Inventory {
	result := Inventory{SchemaVersion: api.SchemaVersion, Controllers: map[string][]string{}}
	for _, controller := range api.Controllers {
		methods := make([]string, len(controller.Methods))
		for index, method := range controller.Methods {
			methods[index] = method.Name
		}
		result.Controllers[controller.Name] = methods
	}
	return result
}
