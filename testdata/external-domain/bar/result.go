// Package bar contains another external API model.
package bar

// Result deliberately shares a type name with foo.Result.
type Result struct {
	Value string `json:"value"`
}
