// Package shared contains nested external models.
package shared

// Data is a recursive external payload.
type Data struct {
	// Name identifies the payload for display.
	Name string `json:"name"`
	Next *Data  `json:"next,omitempty"`
}
