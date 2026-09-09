package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// BaseKind represents the primitive kind of underlying type.
type BaseKind int

const (
	KindUnknown BaseKind = iota
	KindString
	KindInt
	KindUint
	KindUUID
)

// TypeInfo holds metadata about an inspected type.
type TypeInfo struct {
	Name        string   // e.g. "UserID"
	PackageName string   // e.g. "models"
	Kind        BaseKind // KindString, KindInt, KindUint
	Underlying  string   // e.g. "string", "int", "int64", "uint32"
}

// PackageInfo contains parsed information for a package in a directory.
type PackageInfo struct {
	PackageName string
	Dir         string
	Types       map[string]*TypeInfo
}

var stringTypes = map[string]bool{
	"string": true,
}

var signedIntTypes = map[string]bool{
	"int":   true,
	"int8":  true,
	"int16": true,
	"int32": true,
	"int64": true,
}

var unsignedIntTypes = map[string]bool{
	"uint":    true,
	"uint8":   true,
	"uint16":  true,
	"uint32":  true,
	"uint64":  true,
	"uintptr": true,
	"byte":    true,
}

// ClassifyBaseType classifies an underlying primitive type name into BaseKind.
func ClassifyBaseType(name string) (BaseKind, error) {
	if stringTypes[name] {
		return KindString, nil
	}
	if signedIntTypes[name] {
		return KindInt, nil
	}
	if unsignedIntTypes[name] {
		return KindUint, nil
	}
	if name == "uuid.UUID" {
		return KindUUID, nil
	}
	return KindUnknown, fmt.Errorf(
		"unsupported underlying type: %s (only string and integer types are supported)",
		name,
	)
}

// ParsePackage parses all Go source files in the given directory (excluding tests)
// and extracts all defined named types and their underlying types.
func ParsePackage(dir string) (*PackageInfo, error) {
	if dir == "" {
		dir = "."
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory %q: %w", dir, err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no Go package found in directory %q", dir)
	}

	// Pick the non-main package if multiple, or the first package
	var targetPkg *ast.Package
	for name, pkg := range pkgs {
		if targetPkg == nil || (targetPkg.Name == "main" && name != "main") {
			targetPkg = pkg
		}
	}

	if targetPkg == nil {
		return nil, fmt.Errorf("no Go package found in directory %q", dir)
	}

	pkgInfo := &PackageInfo{
		PackageName: targetPkg.Name,
		Dir:         dir,
		Types:       make(map[string]*TypeInfo),
	}

	// First pass: collect raw type declarations
	rawTypes := make(map[string]string)

	for _, file := range targetPkg.Files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				typeName := typeSpec.Name.Name

				if ident, ok := typeSpec.Type.(*ast.Ident); ok {
					rawTypes[typeName] = ident.Name
				} else if sel, ok := typeSpec.Type.(*ast.SelectorExpr); ok {
					if xIdent, ok := sel.X.(*ast.Ident); ok {
						rawTypes[typeName] = xIdent.Name + "." + sel.Sel.Name
					}
				}
			}
		}
	}

	// Second pass: resolve transitive aliases/definitions to find base types
	for typeName, rawUnderlying := range rawTypes {
		visited := make(map[string]bool)
		curr := rawUnderlying
		for {
			if visited[curr] {
				break
			}
			visited[curr] = true
			if next, exists := rawTypes[curr]; exists {
				curr = next
			} else {
				break
			}
		}

		kind, err := ClassifyBaseType(curr)
		if err == nil {
			pkgInfo.Types[typeName] = &TypeInfo{
				Name:        typeName,
				PackageName: pkgInfo.PackageName,
				Kind:        kind,
				Underlying:  curr,
			}
		}
	}

	return pkgInfo, nil
}

// FindType looks up a specific type name in the package info.
func (p *PackageInfo) FindType(typeName string) (*TypeInfo, error) {
	info, exists := p.Types[typeName]
	if !exists {
		return nil, fmt.Errorf(
			"type %q not found or has unsupported underlying type in package %q (directory %q)",
			typeName,
			p.PackageName,
			p.Dir,
		)
	}
	return info, nil
}

// InspectTypes parses a package directory and validates that all specified type names exist and are supported.
func InspectTypes(dir string, typeNames []string) (*PackageInfo, []*TypeInfo, error) {
	if len(typeNames) == 0 {
		return nil, nil, errors.New("no type names specified")
	}

	pkgInfo, err := ParsePackage(dir)
	if err != nil {
		return nil, nil, err
	}

	var result []*TypeInfo
	for _, name := range typeNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		info, err := pkgInfo.FindType(name)
		if err != nil {
			return nil, nil, err
		}
		result = append(result, info)
	}

	if len(result) == 0 {
		return nil, nil, errors.New("no valid type names provided")
	}

	return pkgInfo, result, nil
}
