package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePackage(t *testing.T) {
	tempDir := t.TempDir()

	src := `package testpkg

//typed-id: presets=json,sql
type UserID string

type (
	OrderID int64
	AccountID uint32
	Count int
)

type CustomString UserID
type CustomInt Count
type UuidID uuid.UUID
type UserUUID UuidID

type TokenID []byte

//typed-id:encoding=base64
type SecretID []byte

//typed-id:sql=hex
type HexSQLID []byte

//typed-id:sql=base64
type B64SQLID []byte

//typed-id: presets=json,sql; encoding=base64; sql=bytes
type MultiSettingID []byte

//typed-id: presets=text, sql encoding=hex
type SpaceSettingID []byte

type Meta map[string]any
type Flex map[any]any

type UnrelatedStruct struct {
	Name string
}

type UnrelatedFloat float64
type UnsupportedMap map[int]string
`
	err := os.WriteFile(filepath.Join(tempDir, "types.go"), []byte(src), 0644)
	require.NoError(t, err)

	pkgInfo, err := ParsePackage(tempDir)
	require.NoError(t, err)
	assert.Equal(t, "testpkg", pkgInfo.PackageName)

	// Valid types
	u, err := pkgInfo.FindType("UserID")
	require.NoError(t, err)
	assert.Equal(t, "UserID", u.Name)
	assert.Equal(t, KindString, u.Kind)
	assert.Equal(t, "string", u.Underlying)

	o, err := pkgInfo.FindType("OrderID")
	require.NoError(t, err)
	assert.Equal(t, "OrderID", o.Name)
	assert.Equal(t, KindInt, o.Kind)
	assert.Equal(t, "int64", o.Underlying)

	a, err := pkgInfo.FindType("AccountID")
	require.NoError(t, err)
	assert.Equal(t, "AccountID", a.Name)
	assert.Equal(t, KindUint, a.Kind)
	assert.Equal(t, "uint32", a.Underlying)

	// Transitive aliases
	cs, err := pkgInfo.FindType("CustomString")
	require.NoError(t, err)
	assert.Equal(t, KindString, cs.Kind)
	assert.Equal(t, "string", cs.Underlying)

	ci, err := pkgInfo.FindType("CustomInt")
	require.NoError(t, err)
	assert.Equal(t, KindInt, ci.Kind)
	assert.Equal(t, "int", ci.Underlying)

	uuidType, err := pkgInfo.FindType("UuidID")
	require.NoError(t, err)
	assert.Equal(t, KindUUID, uuidType.Kind)
	assert.Equal(t, "uuid.UUID", uuidType.Underlying)

	userUUID, err := pkgInfo.FindType("UserUUID")
	require.NoError(t, err)
	assert.Equal(t, KindUUID, userUUID.Kind)
	assert.Equal(t, "uuid.UUID", userUUID.Underlying)

	// Unsupported types
	_, err = pkgInfo.FindType("UnrelatedStruct")
	assert.Error(t, err)

	_, err = pkgInfo.FindType("UnrelatedFloat")
	assert.Error(t, err)

	_, err = pkgInfo.FindType("UnsupportedMap")
	assert.Error(t, err)

	// Non-existent type
	_, err = pkgInfo.FindType("NonExistent")
	assert.Error(t, err)

	// []byte types
	tok, err := pkgInfo.FindType("TokenID")
	require.NoError(t, err)
	assert.Equal(t, "TokenID", tok.Name)
	assert.Equal(t, KindBytes, tok.Kind)
	assert.Equal(t, "[]byte", tok.Underlying)
	assert.Equal(t, ByteEncoding(""), tok.ByteEncoding) // no directive

	sec, err := pkgInfo.FindType("SecretID")
	require.NoError(t, err)
	assert.Equal(t, "SecretID", sec.Name)
	assert.Equal(t, KindBytes, sec.Kind)
	assert.Equal(t, "[]byte", sec.Underlying)
	assert.Equal(t, ByteEncodingBase64, sec.ByteEncoding) // directive
	assert.Equal(t, SQLEncoding(""), sec.SQLEncoding)

	hexSql, err := pkgInfo.FindType("HexSQLID")
	require.NoError(t, err)
	assert.Equal(t, SQLEncodingHex, hexSql.SQLEncoding)

	b64Sql, err := pkgInfo.FindType("B64SQLID")
	require.NoError(t, err)
	assert.Equal(t, SQLEncodingBase64, b64Sql.SQLEncoding)

	ms, err := pkgInfo.FindType("MultiSettingID")
	require.NoError(t, err)
	assert.Equal(t, ByteEncodingBase64, ms.ByteEncoding)
	assert.Equal(t, SQLEncodingBytes, ms.SQLEncoding)
	assert.Equal(t, []Preset{PresetJSON, PresetSQL}, ms.Presets)

	ss, err := pkgInfo.FindType("SpaceSettingID")
	require.NoError(t, err)
	assert.Equal(t, ByteEncodingHex, ss.ByteEncoding)
	assert.Equal(t, []Preset{PresetText, PresetSQL}, ss.Presets)

	// UserID presets
	assert.Equal(t, []Preset{PresetJSON, PresetSQL}, u.Presets)

	// Map types
	meta, err := pkgInfo.FindType("Meta")
	require.NoError(t, err)
	assert.Equal(t, KindMap, meta.Kind)
	assert.Equal(t, "map[string]any", meta.Underlying)

	flex, err := pkgInfo.FindType("Flex")
	require.NoError(t, err)
	assert.Equal(t, KindMap, flex.Kind)
	assert.Equal(t, "map[any]any", flex.Underlying)
}

func TestParsePackage_DirectiveErrors(t *testing.T) {
	tempDir := t.TempDir()

	// Invalid preset
	badPreset := `package testpkg
//typed-id: presets=invalid
type BadPresetID string
`
	err := os.WriteFile(filepath.Join(tempDir, "types.go"), []byte(badPreset), 0644)
	require.NoError(t, err)

	_, err = ParsePackage(tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown preset")

	// Invalid encoding
	badEnc := `package testpkg
//typed-id: encoding=invalid
type BadEncID []byte
`
	err = os.WriteFile(filepath.Join(tempDir, "types.go"), []byte(badEnc), 0644)
	require.NoError(t, err)

	_, err = ParsePackage(tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid encoding")

	// Unknown directive
	badKey := `package testpkg
//typed-id: foo=bar
type BadKeyID string
`
	err = os.WriteFile(filepath.Join(tempDir, "types.go"), []byte(badKey), 0644)
	require.NoError(t, err)

	_, err = ParsePackage(tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown directive")
}

func TestInspectTypes(t *testing.T) {
	tempDir := t.TempDir()

	src := `package sample

type UserID string
type OrderID int64
`
	err := os.WriteFile(filepath.Join(tempDir, "sample.go"), []byte(src), 0644)
	require.NoError(t, err)

	pkgInfo, types, err := InspectTypes(tempDir, []string{"UserID", "OrderID"})
	require.NoError(t, err)
	assert.Equal(t, "sample", pkgInfo.PackageName)
	assert.Len(t, types, 2)
	assert.Equal(t, "UserID", types[0].Name)
	assert.Equal(t, "OrderID", types[1].Name)

	// Error when one type is missing
	_, _, err = InspectTypes(tempDir, []string{"UserID", "UnknownID"})
	assert.Error(t, err)

	// Error when no types specified
	_, _, err = InspectTypes(tempDir, []string{})
	assert.Error(t, err)
}
