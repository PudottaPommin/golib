package testdata

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserID_Presets(t *testing.T) {
	orig := UserID("user_abc123")

	// Stringer & TextMarshaler / TextUnmarshaler
	assert.Equal(t, "user_abc123", orig.String())

	textBytes, err := orig.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, []byte("user_abc123"), textBytes)

	var textDec UserID
	err = textDec.UnmarshalText(textBytes)
	require.NoError(t, err)
	assert.Equal(t, orig, textDec)

	// JSON
	jsonBytes, err := json.Marshal(orig)
	require.NoError(t, err)
	assert.Equal(t, `"user_abc123"`, string(jsonBytes))

	var jsonDec UserID
	err = json.Unmarshal(jsonBytes, &jsonDec)
	require.NoError(t, err)
	assert.Equal(t, orig, jsonDec)

	// Binary
	binBytes, err := orig.MarshalBinary()
	require.NoError(t, err)
	var binDec UserID
	err = binDec.UnmarshalBinary(binBytes)
	require.NoError(t, err)
	assert.Equal(t, orig, binDec)

	// SQL Valuer
	val, err := orig.Value()
	require.NoError(t, err)
	assert.Equal(t, "user_abc123", val)

	// SQL Scanner
	var scanID UserID
	require.NoError(t, scanID.Scan("user_from_str"))
	assert.Equal(t, UserID("user_from_str"), scanID)

	require.NoError(t, scanID.Scan([]byte("user_from_bytes")))
	assert.Equal(t, UserID("user_from_bytes"), scanID)

	require.NoError(t, scanID.Scan(nil))
	assert.Equal(t, UserID(""), scanID)

	assert.Error(t, scanID.Scan(12345))
}

func TestAccountID_Presets(t *testing.T) {
	orig := AccountID(9876543210)

	// Stringer & Text
	assert.Equal(t, "9876543210", orig.String())

	textBytes, err := orig.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, []byte("9876543210"), textBytes)

	var textDec AccountID
	err = textDec.UnmarshalText(textBytes)
	require.NoError(t, err)
	assert.Equal(t, orig, textDec)

	// Invalid text
	var badText AccountID
	assert.Error(t, badText.UnmarshalText([]byte("invalid_int")))

	// JSON
	jsonBytes, err := json.Marshal(orig)
	require.NoError(t, err)
	assert.Equal(t, "9876543210", string(jsonBytes))

	var jsonDec AccountID
	err = json.Unmarshal(jsonBytes, &jsonDec)
	require.NoError(t, err)
	assert.Equal(t, orig, jsonDec)

	// Binary
	binBytes, err := orig.MarshalBinary()
	require.NoError(t, err)
	assert.Len(t, binBytes, 8)

	var binDec AccountID
	err = binDec.UnmarshalBinary(binBytes)
	require.NoError(t, err)
	assert.Equal(t, orig, binDec)

	// Invalid binary length
	var badBin AccountID
	assert.Error(t, badBin.UnmarshalBinary([]byte{1, 2, 3}))

	// SQL Valuer
	val, err := orig.Value()
	require.NoError(t, err)
	assert.Equal(t, int64(9876543210), val)

	// SQL Scanner
	var scanID AccountID
	require.NoError(t, scanID.Scan(int64(100)))
	assert.Equal(t, AccountID(100), scanID)

	require.NoError(t, scanID.Scan(int(200)))
	assert.Equal(t, AccountID(200), scanID)

	require.NoError(t, scanID.Scan(int32(300)))
	assert.Equal(t, AccountID(300), scanID)

	require.NoError(t, scanID.Scan("400"))
	assert.Equal(t, AccountID(400), scanID)

	require.NoError(t, scanID.Scan([]byte("500")))
	assert.Equal(t, AccountID(500), scanID)

	require.NoError(t, scanID.Scan(float64(600)))
	assert.Equal(t, AccountID(600), scanID)

	require.NoError(t, scanID.Scan(nil))
	assert.Equal(t, AccountID(0), scanID)

	assert.Error(t, scanID.Scan("not_a_number"))
	assert.Error(t, scanID.Scan(struct{}{}))
}

func TestRoleID_Presets(t *testing.T) {
	orig := RoleID(42)

	// Stringer & Text
	assert.Equal(t, "42", orig.String())

	textBytes, err := orig.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, []byte("42"), textBytes)

	var textDec RoleID
	err = textDec.UnmarshalText(textBytes)
	require.NoError(t, err)
	assert.Equal(t, orig, textDec)

	// JSON
	jsonBytes, err := json.Marshal(orig)
	require.NoError(t, err)
	assert.Equal(t, "42", string(jsonBytes))

	var jsonDec RoleID
	err = json.Unmarshal(jsonBytes, &jsonDec)
	require.NoError(t, err)
	assert.Equal(t, orig, jsonDec)

	// Binary
	binBytes, err := orig.MarshalBinary()
	require.NoError(t, err)
	assert.Len(t, binBytes, 8)

	var binDec RoleID
	err = binDec.UnmarshalBinary(binBytes)
	require.NoError(t, err)
	assert.Equal(t, orig, binDec)

	// SQL Valuer
	val, err := orig.Value()
	require.NoError(t, err)
	assert.Equal(t, int64(42), val)

	// SQL Scanner
	var scanID RoleID
	require.NoError(t, scanID.Scan(uint32(42)))
	assert.Equal(t, RoleID(42), scanID)

	require.NoError(t, scanID.Scan("42"))
	assert.Equal(t, RoleID(42), scanID)

	require.NoError(t, scanID.Scan([]byte("42")))
	assert.Equal(t, RoleID(42), scanID)

	require.NoError(t, scanID.Scan(nil))
	assert.Equal(t, RoleID(0), scanID)

	assert.Error(t, scanID.Scan("invalid"))
	assert.Error(t, scanID.Scan(struct{}{}))
}

func TestSmallID_Overflow(t *testing.T) {
	var s SmallID

	// int8 valid range: -128 to 127
	require.NoError(t, s.UnmarshalText([]byte("120")))
	assert.Equal(t, SmallID(120), s)

	// Overflow int8
	assert.Error(t, s.UnmarshalText([]byte("300")))

	// Scan overflow
	assert.Error(t, s.Scan("300"))
}
