// Package schema defines the machine-readable Wails API documentation model.
package schema

const Version = 2

type API struct {
	SchemaVersion int          `json:"schemaVersion"`
	Generator     string       `json:"generator,omitempty"`
	Controllers   []Controller `json:"controllers"`
	Types         []Type       `json:"types,omitempty"`
}

type Controller struct {
	Name          string   `json:"name"`
	QualifiedName string   `json:"qualifiedName"`
	Description   string   `json:"description,omitempty"`
	Package       string   `json:"package"`
	Methods       []Method `json:"methods"`
}

type Method struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  []Parameter `json:"parameters"`
	Returns     []Return    `json:"returns"`
	Source      Source      `json:"source"`
}

type Parameter struct {
	Name    string `json:"name"`
	GoType  string `json:"goType"`
	TypeRef string `json:"typeRef,omitempty"`
}

type Return struct {
	Name    string `json:"name,omitempty"`
	GoType  string `json:"goType"`
	TypeRef string `json:"typeRef,omitempty"`
}

type Type struct {
	Name          string  `json:"name"`
	QualifiedName string  `json:"qualifiedName"`
	Kind          string  `json:"kind"`
	GoType        string  `json:"goType,omitempty"`
	TypeRef       string  `json:"typeRef,omitempty"`
	Description   string  `json:"description,omitempty"`
	Fields        []Field `json:"fields,omitempty"`
	Source        Source  `json:"source"`
}

type Field struct {
	Name        string `json:"name"`
	JSONName    string `json:"jsonName,omitempty"`
	OmitEmpty   bool   `json:"omitempty,omitempty"`
	GoType      string `json:"goType"`
	TypeRef     string `json:"typeRef,omitempty"`
	Description string `json:"description,omitempty"`
}

type Source struct {
	File string `json:"file"`
	Line int    `json:"line"`
}
