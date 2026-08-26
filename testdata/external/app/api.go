// Package app exposes a Wails-like API.
package app

import (
	"example.com/domain/bar"
	"example.com/domain/foo"
)

// ExternalController returns models from dependency packages.
type ExternalController struct{}

// Foo returns a nested external result.
func (*ExternalController) Foo() (foo.Result, error) { return foo.Result{}, nil }

// Bar returns a distinct type with the same simple name.
func (*ExternalController) Bar() (*bar.Result, error) { return nil, nil }
