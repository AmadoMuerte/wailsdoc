package scanner

import (
	"go/token"
	"go/types"
	"testing"
)

func TestTSType(t *testing.T) {
	pkg := types.NewPackage("example.com/api", "api")
	dto := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "Request", nil), types.NewStruct(nil, nil), nil)
	cases := []struct {
		name string
		typ  types.Type
		want string
	}{
		{"primitive", types.Typ[types.String], "string"},
		{"pointer", types.NewPointer(dto), "Request | null"},
		{"slice", types.NewSlice(types.NewPointer(dto)), "(Request | null)[]"},
		{"array", types.NewArray(types.Typ[types.Int], 3), "number[]"},
		{"map", types.NewMap(types.Typ[types.String], dto), "Record<string, Request>"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := tsType(test.typ, pkg); got != test.want {
				t.Fatalf("tsType() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestTypeRefsIncludesMapKeyAndValue(t *testing.T) {
	pkg := types.NewPackage("example.com/api", "api")
	key := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "Key", nil), types.Typ[types.String], nil)
	value := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "Value", nil), types.NewStruct(nil, nil), nil)
	refs := typeRefs(types.NewMap(key, types.NewSlice(value)))
	if len(refs) != 2 || refs[0] != "example.com/api.Key" || refs[1] != "example.com/api.Value" {
		t.Fatalf("unexpected references: %#v", refs)
	}
}
