package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2EGenerationAndExecution(t *testing.T) {
	tempDir := t.TempDir()

	// Create test package definition
	typesContent := `package fixture

type MemberID string
type ScoreID int64
type BadgeID uint32
`
	err := os.WriteFile(filepath.Join(tempDir, "types.go"), []byte(typesContent), 0o644)
	require.NoError(t, err)

	// Run generator CLI on fixture
	err = run([]string{
		"-dir=" + tempDir,
		"-type=MemberID,ScoreID,BadgeID",
		"-presets=all",
		"-output=types_gen.go",
	})
	require.NoError(t, err)

	// Create a test file in the temp package
	testContent := `package fixture

import (
	"encoding/json"
	"testing"
)

func TestGeneratedMethods(t *testing.T) {
	m := MemberID("mem_123")
	if m.String() != "mem_123" {
		t.Fatalf("unexpected string: %s", m.String())
	}
	jb, err := json.Marshal(m)
	if err != nil || string(jb) != "\"mem_123\"" {
		t.Fatalf("unexpected json: %s, %v", string(jb), err)
	}

	s := ScoreID(9999)
	if s.String() != "9999" {
		t.Fatalf("unexpected string: %s", s.String())
	}
	sb, err := s.MarshalBinary()
	if err != nil || len(sb) != 8 {
		t.Fatalf("unexpected binary: %v, %v", sb, err)
	}

	b := BadgeID(55)
	if b.String() != "55" {
		t.Fatalf("unexpected string: %s", b.String())
	}
	var bScan BadgeID
	if err := bScan.Scan("55"); err != nil || bScan != 55 {
		t.Fatalf("unexpected scan: %v, %v", bScan, err)
	}
}
`
	err = os.WriteFile(filepath.Join(tempDir, "types_test.go"), []byte(testContent), 0o644)
	require.NoError(t, err)

	// Copy go.mod or run go test in directory
	goModContent := `module fixturetest

go 1.27
`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0o644)
	require.NoError(t, err)

	cmd := exec.Command("go", "get", "github.com/pudottapommin/golib")
	cmd.Dir = tempDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "go mod tidy: %s", string(out))

	cmd = exec.Command("go", "test", "-v", ".")
	cmd.Dir = tempDir
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "go test failed: %s", string(out))
	assert.Contains(t, string(out), "PASS")
}
