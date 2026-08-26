// Package foo contains external API models.
package foo

import "example.com/domain/shared"

// Result is returned by an external service.
type Result struct {
	// Data contains the nested response payload.
	Data shared.Data `json:"data"`
}
