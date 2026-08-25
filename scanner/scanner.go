// Package scanner extracts Wails-style controller APIs from Go source.
package scanner

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/AmadoMuerte/wailsdoc/schema"
	"golang.org/x/tools/go/packages"
)

type Options struct {
	Packages         []string
	Dir              string
	ControllerSuffix string
	Generator        string
}

func Scan(ctx context.Context, options Options) (schema.API, error) {
	if len(options.Packages) == 0 {
		return schema.API{}, fmt.Errorf("scan: no packages configured")
	}
	if options.ControllerSuffix == "" {
		options.ControllerSuffix = "Controller"
	}
	config := &packages.Config{
		Context: ctx,
		Dir:     options.Dir,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedDeps | packages.NeedModule,
		Tests: false,
	}
	loaded, err := packages.Load(config, options.Packages...)
	if err != nil {
		return schema.API{}, fmt.Errorf("scan: load packages: %w", err)
	}
	if packages.PrintErrors(loaded) > 0 {
		return schema.API{}, fmt.Errorf("scan: package loading failed")
	}
	result := schema.API{SchemaVersion: schema.Version, Generator: options.Generator, Controllers: []schema.Controller{}}
	s := state{api: &result, root: moduleRoot(loaded), packages: indexPackages(loaded), seen: map[string]bool{}}
	for _, pkg := range loaded {
		if pkg.Types == nil || pkg.TypesInfo == nil {
			continue
		}
		s.scanPackage(pkg, options.ControllerSuffix)
	}
	sort.Slice(result.Controllers, func(i, j int) bool { return result.Controllers[i].QualifiedName < result.Controllers[j].QualifiedName })
	sort.Slice(result.Types, func(i, j int) bool { return result.Types[i].QualifiedName < result.Types[j].QualifiedName })
	return result, nil
}

type state struct {
	api      *schema.API
	root     string
	packages map[*types.Package]*packages.Package
	seen     map[string]bool
}

func (s *state) scanPackage(pkg *packages.Package, suffix string) {
	for _, name := range pkg.Types.Scope().Names() {
		object, ok := pkg.Types.Scope().Lookup(name).(*types.TypeName)
		if !ok || !object.Exported() || !strings.HasSuffix(name, suffix) {
			continue
		}
		named, ok := types.Unalias(object.Type()).(*types.Named)
		if !ok {
			continue
		}
		if _, ok := named.Underlying().(*types.Struct); !ok {
			continue
		}
		controller := schema.Controller{
			Name: name, QualifiedName: qualified(object), Package: pkg.PkgPath,
			Description: objectComment(pkg, object), Methods: []schema.Method{},
		}
		methods := types.NewMethodSet(types.NewPointer(named))
		for index := 0; index < methods.Len(); index++ {
			selection := methods.At(index)
			function, ok := selection.Obj().(*types.Func)
			if !ok || !function.Exported() || selection.Obj().Pkg() != pkg.Types {
				continue
			}
			signature, ok := function.Type().(*types.Signature)
			if !ok {
				continue
			}
			method := schema.Method{
				Name: function.Name(), Description: objectComment(pkg, function),
				Parameters: []schema.Parameter{}, Returns: []schema.Return{}, Source: s.source(pkg, function.Pos()),
			}
			for parameter := 0; parameter < signature.Params().Len(); parameter++ {
				value := signature.Params().At(parameter)
				method.Parameters = append(method.Parameters, schema.Parameter{Name: value.Name(), GoType: typeString(value.Type(), pkg.Types), TypeRef: typeRef(value.Type())})
				s.resolve(value.Type())
			}
			for result := 0; result < signature.Results().Len(); result++ {
				value := signature.Results().At(result)
				method.Returns = append(method.Returns, schema.Return{Name: value.Name(), GoType: typeString(value.Type(), pkg.Types), TypeRef: typeRef(value.Type())})
				s.resolve(value.Type())
			}
			controller.Methods = append(controller.Methods, method)
		}
		sort.Slice(controller.Methods, func(i, j int) bool { return controller.Methods[i].Name < controller.Methods[j].Name })
		s.api.Controllers = append(s.api.Controllers, controller)
	}
}

func (s *state) resolve(typ types.Type) {
	switch current := typ.(type) {
	case *types.Pointer:
		s.resolve(current.Elem())
	case *types.Slice:
		s.resolve(current.Elem())
	case *types.Array:
		s.resolve(current.Elem())
	case *types.Map:
		s.resolve(current.Key())
		s.resolve(current.Elem())
	case *types.Alias:
		if object := current.Obj(); object != nil && s.projectType(object) {
			s.addType(object, current, types.Unalias(current))
		}
		s.resolve(types.Unalias(current))
	case *types.Named:
		object := current.Obj()
		if object == nil || !s.projectType(object) {
			return
		}
		s.addType(object, current, current.Underlying())
	}
}

func (s *state) addType(object *types.TypeName, declared, underlying types.Type) {
	key := qualified(object)
	if s.seen[key] {
		return
	}
	s.seen[key] = true
	pkg := s.packages[object.Pkg()]
	entry := schema.Type{Name: object.Name(), QualifiedName: key, Kind: kind(underlying), Description: objectComment(pkg, object), Source: s.source(pkg, object.Pos())}
	if structure, ok := underlying.(*types.Struct); ok {
		entry.Fields = []schema.Field{}
		for index := 0; index < structure.NumFields(); index++ {
			field := structure.Field(index)
			if !field.Exported() {
				continue
			}
			jsonName, omitEmpty, skip := jsonField(structure.Tag(index), field.Name())
			if skip {
				continue
			}
			entry.Fields = append(entry.Fields, schema.Field{Name: field.Name(), JSONName: jsonName, OmitEmpty: omitEmpty, GoType: typeString(field.Type(), object.Pkg()), TypeRef: typeRef(field.Type()), Description: objectComment(pkg, field)})
			s.resolve(field.Type())
		}
	} else {
		entry.GoType = typeString(declared, object.Pkg())
		entry.TypeRef = typeRef(underlying)
		s.resolve(underlying)
	}
	s.api.Types = append(s.api.Types, entry)
}

func (s *state) projectType(object *types.TypeName) bool {
	if object.Pkg() == nil || s.root == "" {
		return false
	}
	pkg := s.packages[object.Pkg()]
	return pkg != nil && pkg.Module != nil && filepath.Clean(pkg.Module.Dir) == s.root
}

func moduleRoot(packages_ []*packages.Package) string {
	var root string
	for _, pkg := range packages_ {
		if root == "" && pkg.Module != nil && pkg.Module.Dir != "" {
			root = filepath.Clean(pkg.Module.Dir)
		}
	}
	return root
}

func indexPackages(roots []*packages.Package) map[*types.Package]*packages.Package {
	result := map[*types.Package]*packages.Package{}
	seen := map[*packages.Package]bool{}
	var visit func(*packages.Package)
	visit = func(pkg *packages.Package) {
		if pkg == nil || seen[pkg] {
			return
		}
		seen[pkg] = true
		if pkg.Types != nil {
			result[pkg.Types] = pkg
		}
		for _, imported := range pkg.Imports {
			visit(imported)
		}
	}
	for _, root := range roots {
		visit(root)
	}
	return result
}

func qualified(object *types.TypeName) string {
	if object.Pkg() == nil {
		return object.Name()
	}
	return object.Pkg().Name() + "." + object.Name()
}

func typeRef(typ types.Type) string {
	switch current := typ.(type) {
	case *types.Named:
		if current.Obj().Pkg() == nil {
			return ""
		}
		return qualified(current.Obj())
	case *types.Pointer:
		return typeRef(current.Elem())
	case *types.Slice:
		return typeRef(current.Elem())
	case *types.Array:
		return typeRef(current.Elem())
	case *types.Map:
		return typeRef(current.Elem())
	case *types.Alias:
		if current.Obj().Pkg() == nil {
			return ""
		}
		return qualified(current.Obj())
	default:
		return ""
	}
}

func typeString(typ types.Type, current *types.Package) string {
	return types.TypeString(typ, func(pkg *types.Package) string {
		if pkg == current {
			return ""
		}
		return pkg.Name()
	})
}

func kind(typ types.Type) string {
	switch typ.(type) {
	case *types.Struct:
		return "struct"
	case *types.Interface:
		return "interface"
	default:
		return "alias"
	}
}

func jsonField(tag, fallback string) (string, bool, bool) {
	value := reflect.StructTag(tag).Get("json")
	parts := strings.Split(value, ",")
	if len(parts) > 0 && parts[0] == "-" {
		return "", false, true
	}
	name := fallback
	if len(parts) > 0 && parts[0] != "" {
		name = parts[0]
	}
	for _, option := range parts[1:] {
		if option == "omitempty" {
			return name, true, false
		}
	}
	return name, false, false
}

func objectComment(pkg *packages.Package, object types.Object) string {
	if pkg == nil || object == nil {
		return ""
	}
	for _, file := range pkg.Syntax {
		for _, declaration := range file.Decls {
			switch node := declaration.(type) {
			case *ast.FuncDecl:
				if pkg.TypesInfo.Defs[node.Name] == object {
					return commentText(node.Doc)
				}
			case *ast.GenDecl:
				for _, spec := range node.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok || pkg.TypesInfo.Defs[typeSpec.Name] != object {
						continue
					}
					if typeSpec.Doc != nil {
						return commentText(typeSpec.Doc)
					}
					return commentText(node.Doc)
				}
			}
		}
		var found string
		ast.Inspect(file, func(node ast.Node) bool {
			field, ok := node.(*ast.Field)
			if !ok || found != "" {
				return true
			}
			for _, name := range field.Names {
				if pkg.TypesInfo.Defs[name] == object {
					found = commentText(field.Doc)
					if found == "" {
						found = commentText(field.Comment)
					}
					return false
				}
			}
			return true
		})
		if found != "" {
			return found
		}
	}
	return ""
}

func commentText(group *ast.CommentGroup) string {
	if group == nil {
		return ""
	}
	return strings.TrimSpace(group.Text())
}

func (s *state) source(pkg *packages.Package, position token.Pos) schema.Source {
	if pkg == nil || !position.IsValid() || pkg.Fset == nil {
		return schema.Source{}
	}
	value := pkg.Fset.Position(position)
	file := filepath.Clean(value.Filename)
	if s.root != "" {
		if relative, err := filepath.Rel(s.root, file); err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			file = filepath.ToSlash(relative)
		}
	}
	return schema.Source{File: filepath.ToSlash(file), Line: value.Line}
}
