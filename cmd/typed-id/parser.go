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
	KindBytes
	KindMap
)

// ByteEncoding controls how []byte is rendered in text/JSON contexts.
type ByteEncoding string

const (
	ByteEncodingHex    ByteEncoding = "hex"
	ByteEncodingBase64 ByteEncoding = "base64"
)

// SQLEncoding controls how []byte is stored and retrieved in SQL database drivers.
type SQLEncoding string

const (
	SQLEncodingBytes  SQLEncoding = "bytes"
	SQLEncodingHex    SQLEncoding = "hex"
	SQLEncodingBase64 SQLEncoding = "base64"
)

// TypeDirectives stores per-type configuration parsed from doc comments.
type TypeDirectives struct {
	ByteEncoding ByteEncoding
	SQLEncoding  SQLEncoding
	Presets      []Preset
}

// TypeInfo holds metadata about an inspected type.
type TypeInfo struct {
	Name         string       // e.g. "UserID"
	PackageName  string       // e.g. "models"
	Kind         BaseKind     // KindString, KindInt, KindUint, KindBytes, KindMap
	Underlying   string       // e.g. "string", "int64", "[]byte", "map[string]any"
	ByteEncoding ByteEncoding // Only set for KindBytes
	SQLEncoding  SQLEncoding  // Only set for KindBytes (bytes, hex, base64)
	Presets      []Preset     // Configured presets from doc-comment (empty if not specified)
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
	if name == "[]byte" {
		return KindBytes, nil
	}
	if name == "map[string]any" || name == "map[any]any" {
		return KindMap, nil
	}
	return KindUnknown, fmt.Errorf(
		"unsupported underlying type: %s (only string, integer, []byte, map[string]any, and map[any]any types are supported)",
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

	// First pass: collect raw type declarations and directives
	rawTypes := make(map[string]string)
	typeDirectives := make(map[string]TypeDirectives)

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

				dirs, err := parseTypeDirectives(genDecl.Doc, typeSpec.Doc)
				if err != nil {
					return nil, fmt.Errorf("failed to parse directives for type %q: %w", typeName, err)
				}
				if dirs.ByteEncoding != "" || dirs.SQLEncoding != "" || len(dirs.Presets) > 0 {
					typeDirectives[typeName] = dirs
				}

				if ident, ok := typeSpec.Type.(*ast.Ident); ok {
					rawTypes[typeName] = ident.Name
				} else if sel, ok := typeSpec.Type.(*ast.SelectorExpr); ok {
					if xIdent, ok := sel.X.(*ast.Ident); ok {
						rawTypes[typeName] = xIdent.Name + "." + sel.Sel.Name
					}
				} else if arr, ok := typeSpec.Type.(*ast.ArrayType); ok {
					// []byte: ArrayType with nil Len (slice) and Elt = "byte" or "uint8"
					if arr.Len == nil {
						if eltIdent, ok := arr.Elt.(*ast.Ident); ok {
							if eltIdent.Name == "byte" || eltIdent.Name == "uint8" {
								rawTypes[typeName] = "[]byte"
							}
						}
					}
				} else if mp, ok := typeSpec.Type.(*ast.MapType); ok {
					keyStr := identName(mp.Key)
					valStr := identName(mp.Value)
					composite := "map[" + keyStr + "]" + valStr
					if composite == "map[string]any" || composite == "map[any]any" {
						rawTypes[typeName] = composite
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
			dirs := typeDirectives[typeName]
			ti := &TypeInfo{
				Name:        typeName,
				PackageName: pkgInfo.PackageName,
				Kind:        kind,
				Underlying:  curr,
				Presets:     dirs.Presets,
			}
			if kind == KindBytes {
				ti.ByteEncoding = dirs.ByteEncoding
				ti.SQLEncoding = dirs.SQLEncoding
			}
			pkgInfo.Types[typeName] = ti
		}
	}

	return pkgInfo, nil
}

// identName extracts the name from an *ast.Ident expression.
func identName(expr ast.Expr) string {
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

// parseDirectiveLine parses a single comment line for typed-id directives.
func parseDirectiveLine(line string) (TypeDirectives, error) {
	var dir TypeDirectives
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "//")
	trimmed = strings.TrimPrefix(trimmed, "/*")
	trimmed = strings.TrimSuffix(trimmed, "*/")
	trimmed = strings.TrimSpace(trimmed)

	var rest string
	if after, ok := strings.CutPrefix(trimmed, "typed-id:"); ok {
		rest = after
	} else if after, ok := strings.CutPrefix(trimmed, "typed-id "); ok {
		rest = after
	} else {
		return dir, nil
	}

	// Replace semicolons with spaces
	rest = strings.ReplaceAll(rest, ";", " ")
	tokens := strings.Fields(rest)
	if len(tokens) == 0 {
		return dir, nil
	}

	type kv struct {
		key string
		val string
	}
	var pairs []kv
	for _, tok := range tokens {
		if k, v, ok := strings.Cut(tok, "="); ok {
			pairs = append(pairs, kv{key: strings.ToLower(strings.TrimSpace(k)), val: strings.TrimSpace(v)})
		} else if len(pairs) > 0 {
			// Continuation of previous value (e.g. presets=json, sql)
			pairs[len(pairs)-1].val += tok
		} else {
			return dir, fmt.Errorf("invalid directive token %q (expected key=value)", tok)
		}
	}

	for _, pair := range pairs {
		val := strings.Trim(pair.val, `,"'`)
		switch pair.key {
		case "encoding", "byte-encoding":
			switch ByteEncoding(val) {
			case ByteEncodingHex, ByteEncodingBase64:
				dir.ByteEncoding = ByteEncoding(val)
			default:
				return dir, fmt.Errorf("invalid encoding %q (supported: hex, base64)", val)
			}
		case "sql", "sql-encoding":
			switch SQLEncoding(val) {
			case SQLEncodingBytes, SQLEncodingHex, SQLEncodingBase64:
				dir.SQLEncoding = SQLEncoding(val)
			default:
				return dir, fmt.Errorf("invalid sql encoding %q (supported: bytes, hex, base64)", val)
			}
		case "preset", "presets":
			rawParts := strings.Split(val, ",")
			var pList []Preset
			for _, p := range rawParts {
				p = strings.TrimSpace(p)
				if p != "" {
					pList = append(pList, Preset(p))
				}
			}
			norm, err := NormalizePresets(pList)
			if err != nil {
				return dir, err
			}
			dir.Presets = norm
		default:
			return dir, fmt.Errorf("unknown directive %q (supported: presets, encoding, sql)", pair.key)
		}
	}
	return dir, nil
}

// parseTypeDirectives scans comment groups for typed-id directives and merges them.
func parseTypeDirectives(groups ...*ast.CommentGroup) (TypeDirectives, error) {
	var merged TypeDirectives
	for _, cg := range groups {
		if cg == nil {
			continue
		}
		for _, c := range cg.List {
			dir, err := parseDirectiveLine(c.Text)
			if err != nil {
				return merged, err
			}
			if dir.ByteEncoding != "" {
				merged.ByteEncoding = dir.ByteEncoding
			}
			if dir.SQLEncoding != "" {
				merged.SQLEncoding = dir.SQLEncoding
			}
			if len(dir.Presets) > 0 {
				merged.Presets = dir.Presets
			}
		}
	}
	return merged, nil
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
