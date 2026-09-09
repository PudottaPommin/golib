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
`
	err := os.WriteFile(filepath.Join(tempDir, "types.go"), []byte(typesSrc), 0o644)
	require.NoError(t, err)

	// Single type default output name
	err = run([]string{"-dir=" + tempDir, "-type=MemberID", "-presets=all"})
	require.NoError(t, err)

	genPath := filepath.Join(tempDir, "memberid_typed_id.go")
	assert.FileExists(t, genPath)
	content, err := os.ReadFile(genPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "func (id MemberID) MarshalJSON()")

	// Multiple types with custom output
	customOut := "custom_gen.go"
	err = run([]string{"-dir=" + tempDir, "-type=MemberID,OrgID,UuidID", "-presets=json,sql", "-output=" + customOut})
	require.NoError(t, err)

	customPath := filepath.Join(tempDir, customOut)
	assert.FileExists(t, customPath)
	content, err = os.ReadFile(customPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "func (id MemberID) MarshalJSON()")
	assert.Contains(t, string(content), "func (id OrgID) Value()")
	assert.Contains(t, string(content), "func (id UuidID) MarshalJSON()")
	assert.Contains(t, string(content), "func (id UuidID) Value()")
	assert.NotContains(t, string(content), "MarshalBinary") // binary omitted

	// Missing type flag error
	err = run([]string{"-dir=" + tempDir})
	assert.Error(t, err)

	// Unknown type error
	err = run([]string{"-dir=" + tempDir, "-type=NonExistentID"})
	assert.Error(t, err)
}
