package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIRun(t *testing.T) {
	tempDir := t.TempDir()

	typesSrc := `package testdomain
import "uuid"
	
type MemberID string
type OrgID int64
type UuidID uuid.UUID
type TokenID []byte

//typed-id:encoding=base64
type SecretID []byte

//typed-id: presets=json encoding=base64
type JsonOnlyID []byte

type Meta map[string]any
`
	err := os.WriteFile(filepath.Join(tempDir, "types.go"), []byte(typesSrc), 0o644)
	require.NoError(t, err)

	// Single type default output name
	err = run([]string{"-dir=" + tempDir, "-type=MemberID", "-presets=all"})
	require.NoError(t, err)

	genPath := filepath.Join(tempDir, "memberid.typed_id.go")
	assert.FileExists(t, genPath)
	content, err := os.ReadFile(genPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "func (id MemberID) MarshalJSONTo(in *jsontext.Encoder) error")

	// Multiple types with custom output including new types
	customOut := "custom_gen.go"
	err = run([]string{"-dir=" + tempDir, "-type=MemberID,OrgID,UuidID,TokenID,SecretID,Meta", "-presets=json,sql", "-output=" + customOut})
	require.NoError(t, err)

	customPath := filepath.Join(tempDir, customOut)
	assert.FileExists(t, customPath)
	content, err = os.ReadFile(customPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "func (id MemberID) MarshalJSONTo(in *jsontext.Encoder) error")
	assert.Contains(t, string(content), "func (id OrgID) Value()")
	assert.Contains(t, string(content), "func (id UuidID) MarshalJSONTo(in *jsontext.Encoder) error")
	assert.Contains(t, string(content), "func (id UuidID) Value()")
	assert.NotContains(t, string(content), "MarshalBinary") // binary omitted

	// TokenID should use hex (default)
	assert.Contains(t, string(content), "func (id TokenID) MarshalJSONTo(enc *jsontext.Encoder) error")
	assert.Contains(t, string(content), "hex.EncodeToString")

	// SecretID should use base64 (from directive, overrides default)
	assert.Contains(t, string(content), "func (id SecretID) MarshalJSONTo(enc *jsontext.Encoder) error")
	assert.Contains(t, string(content), "base64.RawURLEncoding")

	// Meta should have JSON and SQL
	assert.Contains(t, string(content), "func (id Meta) MarshalJSONTo(enc *jsontext.Encoder) error")
	assert.Contains(t, string(content), "func (id *Meta) Scan(src any) error")

	// TokenID should use raw []byte for SQL by default
	assert.Contains(t, string(content), "func (id TokenID) Value() (driver.Value, error)")
	assert.Contains(t, string(content), "return []byte(id), nil")

	// Test -byte-encoding=base64 flag (TokenID should now use base64)
	flagOut := "flag_gen.go"
	err = run([]string{"-dir=" + tempDir, "-type=TokenID", "-presets=text", "-byte-encoding=base64", "-output=" + flagOut})
	require.NoError(t, err)
	flagContent, err := os.ReadFile(filepath.Join(tempDir, flagOut))
	require.NoError(t, err)
	assert.Contains(t, string(flagContent), "base64.RawURLEncoding")
	assert.NotContains(t, string(flagContent), "hex.EncodeToString")

	// Test -sql-byte-encoding=hex flag (TokenID should use hex for SQL Value)
	sqlFlagOut := "sql_flag_gen.go"
	err = run([]string{"-dir=" + tempDir, "-type=TokenID", "-presets=sql", "-sql-byte-encoding=hex", "-output=" + sqlFlagOut})
	require.NoError(t, err)
	sqlFlagContent, err := os.ReadFile(filepath.Join(tempDir, sqlFlagOut))
	require.NoError(t, err)
	assert.Contains(t, string(sqlFlagContent), "return hex.EncodeToString([]byte(id)), nil")

	// Test per-type presets (JsonOnlyID specifies presets=json, should override -presets=all)
	jsonOnlyOut := "json_only_gen.go"
	err = run([]string{"-dir=" + tempDir, "-type=JsonOnlyID", "-presets=all", "-output=" + jsonOnlyOut})
	require.NoError(t, err)
	jsonOnlyContent, err := os.ReadFile(filepath.Join(tempDir, jsonOnlyOut))
	require.NoError(t, err)
	assert.Contains(t, string(jsonOnlyContent), "func (id JsonOnlyID) MarshalJSONTo(enc *jsontext.Encoder) error")
	assert.NotContains(t, string(jsonOnlyContent), "func (id JsonOnlyID) String()")
	assert.NotContains(t, string(jsonOnlyContent), "func (id JsonOnlyID) Value()")
	assert.NotContains(t, string(jsonOnlyContent), "MarshalBinary")

	// Missing type flag error
	err = run([]string{"-dir=" + tempDir})
	assert.Error(t, err)

	// Unknown type error
	err = run([]string{"-dir=" + tempDir, "-type=NonExistentID"})
	assert.Error(t, err)

	// Invalid byte-encoding flag
	err = run([]string{"-dir=" + tempDir, "-type=TokenID", "-byte-encoding=invalid"})
	assert.Error(t, err)

	// Invalid sql-byte-encoding flag
	err = run([]string{"-dir=" + tempDir, "-type=TokenID", "-sql-byte-encoding=invalid"})
	assert.Error(t, err)
}
