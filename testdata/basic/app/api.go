// Package app is a generic Wails-like API fixture.
package app

import "time"

// Role names a user role.
type Role string

// UserDTO is a recursive public user representation.
type UserDTO struct {
	// ID uniquely identifies the user.
	ID       int               `json:"id"`
	Name     string            `json:"name"`
	Active   bool              `json:"active"`
	Role     Role              `json:"role"`
	Parent   *UserDTO          `json:"parent,omitempty"`
	Children []*UserDTO        `json:"children,omitempty"`
	Labels   map[string]string `json:"labels"`
	Created  time.Time         `json:"created"`
	Profile  *ProfileDTO       `json:"profile,omitempty"`
}

// ProfileDTO contains nested user data.
type ProfileDTO struct {
	// Bio is the user-provided profile text.
	Bio string `json:"bio,omitempty"`
}

// UserController exposes user operations to a Wails frontend.
type UserController struct{}

// GetUser returns one user by identifier.
//
// Errors:
//   - user_not_found: no user has the requested identifier
func (c *UserController) GetUser(id int) (*UserDTO, error) { return nil, nil }

// ListUsers returns users grouped by role.
func (c *UserController) ListUsers(active bool) (map[Role][]UserDTO, error) { return nil, nil }

// Ping verifies that the backend is available.
func (c *UserController) Ping() error { return nil }
