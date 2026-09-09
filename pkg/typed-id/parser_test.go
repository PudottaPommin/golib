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

type UnrelatedStruct struct {
	Name string
}

type UnrelatedFloat float64
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

	// Non-existent type
	_, err = pkgInfo.FindType("NonExistent")
	assert.Error(t, err)
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
