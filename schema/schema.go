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
	Errors      []Error     `json:"errors,omitempty"`
	Source      Source      `json:"source"`
}

type Error struct {
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

type Parameter struct {
	Name     string   `json:"name"`
	GoType   string   `json:"goType"`
	TSType   string   `json:"tsType,omitempty"`
	TypeRef  string   `json:"typeRef,omitempty"`
	TypeRefs []string `json:"typeRefs,omitempty"`
	Nullable bool     `json:"nullable,omitempty"`
}

type Return struct {
	Name     string   `json:"name,omitempty"`
	GoType   string   `json:"goType"`
	TSType   string   `json:"tsType,omitempty"`
	TypeRef  string   `json:"typeRef,omitempty"`
	TypeRefs []string `json:"typeRefs,omitempty"`
	Nullable bool     `json:"nullable,omitempty"`
}

type Type struct {
	Name          string   `json:"name"`
	QualifiedName string   `json:"qualifiedName"`
	Kind          string   `json:"kind"`
	GoType        string   `json:"goType,omitempty"`
	TSType        string   `json:"tsType,omitempty"`
	TypeRef       string   `json:"typeRef,omitempty"`
	TypeRefs      []string `json:"typeRefs,omitempty"`
	Description   string   `json:"description,omitempty"`
	Fields        []Field  `json:"fields,omitempty"`
	Source        Source   `json:"source"`
}

type Field struct {
	Name        string   `json:"name"`
	JSONName    string   `json:"jsonName,omitempty"`
	OmitEmpty   bool     `json:"omitempty,omitempty"`
	Nullable    bool     `json:"nullable,omitempty"`
	GoType      string   `json:"goType"`
	TSType      string   `json:"tsType,omitempty"`
	TypeRef     string   `json:"typeRef,omitempty"`
	TypeRefs    []string `json:"typeRefs,omitempty"`
	Description string   `json:"description,omitempty"`
}

type Source struct {
	File string `json:"file"`
	Line int    `json:"line"`
}
