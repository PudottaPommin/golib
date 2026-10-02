package testdata

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"testing"
	"uuid"

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

func TestTokenID_Presets(t *testing.T) {
	orig := TokenID{0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe}

	// String should be hex
	assert.Equal(t, "deadbeefcafe", orig.String())

	// Text round-trip
	textBytes, err := orig.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, []byte("deadbeefcafe"), textBytes)

	var textDec TokenID
	err = textDec.UnmarshalText(textBytes)
	require.NoError(t, err)
	assert.Equal(t, orig, textDec)

	// Invalid hex text
	var badText TokenID
	assert.Error(t, badText.UnmarshalText([]byte("zzzz")))

	// JSON round-trip (hex string)
	jsonBytes, err := json.Marshal(orig)
	require.NoError(t, err)
	assert.Equal(t, `"deadbeefcafe"`, string(jsonBytes))

	var jsonDec TokenID
	err = json.Unmarshal(jsonBytes, &jsonDec)
	require.NoError(t, err)
	assert.Equal(t, orig, jsonDec)

	// Binary round-trip
	binBytes, err := orig.MarshalBinary()
	require.NoError(t, err)
	assert.Equal(t, []byte(orig), binBytes)

	var binDec TokenID
	err = binDec.UnmarshalBinary(binBytes)
	require.NoError(t, err)
	assert.Equal(t, orig, binDec)

	// SQL Value (raw []byte by default)
	val, err := orig.Value()
	require.NoError(t, err)
	assert.Equal(t, []byte(orig), val)

	// SQL Scan from string (hex)
	var scanID TokenID
	require.NoError(t, scanID.Scan("deadbeefcafe"))
	assert.Equal(t, orig, scanID)

	// SQL Scan from []byte (raw)
	require.NoError(t, scanID.Scan([]byte{0xde, 0xad}))
	assert.Equal(t, TokenID{0xde, 0xad}, scanID)

	// SQL Scan nil
	require.NoError(t, scanID.Scan(nil))
	assert.Nil(t, []byte(scanID))

	// SQL Scan unsupported type
	assert.Error(t, scanID.Scan(12345))

	// SQL Scan invalid hex string
	assert.Error(t, scanID.Scan("zzzz"))

	// Binder
	var bindID TokenID
	require.NoError(t, bindID.UnmarshalBind("deadbeefcafe"))
	assert.Equal(t, orig, bindID)

	// Binder invalid
	assert.Error(t, bindID.UnmarshalBind("zzzz"))

	// Nil value
	var nilID TokenID
	val, err = nilID.Value()
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestSecretID_Presets(t *testing.T) {
	orig := SecretID{0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe}
	expectedB64 := base64.RawURLEncoding.EncodeToString([]byte(orig))

	// String should be base64
	assert.Equal(t, expectedB64, orig.String())

	// Text round-trip
	textBytes, err := orig.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, []byte(expectedB64), textBytes)

	var textDec SecretID
	err = textDec.UnmarshalText(textBytes)
	require.NoError(t, err)
	assert.Equal(t, orig, textDec)

	// JSON round-trip (base64 string)
	jsonBytes, err := json.Marshal(orig)
	require.NoError(t, err)

	var jsonDec SecretID
	err = json.Unmarshal(jsonBytes, &jsonDec)
	require.NoError(t, err)
	assert.Equal(t, orig, jsonDec)

	// Binary round-trip (same as raw bytes)
	binBytes, err := orig.MarshalBinary()
	require.NoError(t, err)

	var binDec SecretID
	err = binDec.UnmarshalBinary(binBytes)
	require.NoError(t, err)
	assert.Equal(t, orig, binDec)

	// SQL Value (raw []byte by default)
	val, err := orig.Value()
	require.NoError(t, err)
	assert.Equal(t, []byte(orig), val)

	// SQL Scan from string (base64)
	var scanID SecretID
	require.NoError(t, scanID.Scan(expectedB64))
	assert.Equal(t, orig, scanID)

	// SQL Scan from []byte (raw)
	require.NoError(t, scanID.Scan([]byte{0xca, 0xfe}))
	assert.Equal(t, SecretID{0xca, 0xfe}, scanID)

	// SQL Scan nil
	require.NoError(t, scanID.Scan(nil))
	assert.Nil(t, []byte(scanID))

	// Binder
	var bindID SecretID
	require.NoError(t, bindID.UnmarshalBind(expectedB64))
	assert.Equal(t, orig, bindID)

	// Verify hex and base64 produce different results for same data
	hexStr := hex.EncodeToString([]byte(orig))
	assert.NotEqual(t, hexStr, expectedB64)
}

func TestMetadata_Presets(t *testing.T) {
	orig := Metadata{"key": "value", "count": float64(42)}

	// String should be JSON
	s := orig.String()
	assert.Contains(t, s, `"key"`)
	assert.Contains(t, s, `"value"`)

	// Text round-trip
	textBytes, err := orig.MarshalText()
	require.NoError(t, err)

	var textDec Metadata
	err = textDec.UnmarshalText(textBytes)
	require.NoError(t, err)
	assert.Equal(t, orig, textDec)

	// JSON round-trip
	jsonBytes, err := json.Marshal(orig)
	require.NoError(t, err)

	var jsonDec Metadata
	err = json.Unmarshal(jsonBytes, &jsonDec)
	require.NoError(t, err)
	assert.Equal(t, orig, jsonDec)

	// SQL Value (JSON string)
	val, err := orig.Value()
	require.NoError(t, err)
	valStr, ok := val.(string)
	require.True(t, ok)
	assert.Contains(t, valStr, `"key"`)

	// SQL Scan from string
	var scanID Metadata
	require.NoError(t, scanID.Scan(valStr))
	assert.Equal(t, orig, scanID)

	// SQL Scan from []byte
	var scanBytes Metadata
	require.NoError(t, scanBytes.Scan([]byte(valStr)))
	assert.Equal(t, orig, scanBytes)

	// SQL Scan nil
	require.NoError(t, scanID.Scan(nil))
	assert.Nil(t, (map[string]any)(scanID))

	// SQL Scan unsupported type
	assert.Error(t, scanID.Scan(12345))

	// Nil value
	var nilMeta Metadata
	val, err = nilMeta.Value()
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestFlexData_Presets(t *testing.T) {
	orig := FlexData{"key": "value", "nested": map[string]any{"a": float64(1)}}

	// String should be JSON
	s := orig.String()
	assert.Contains(t, s, `"key"`)

	// Text round-trip
	textBytes, err := orig.MarshalText()
	require.NoError(t, err)

	var textDec FlexData
	err = textDec.UnmarshalText(textBytes)
	require.NoError(t, err)
	assert.Equal(t, "value", textDec["key"])

	// JSON round-trip
	jsonBytes, err := json.Marshal(orig)
	require.NoError(t, err)

	var jsonDec FlexData
	err = json.Unmarshal(jsonBytes, &jsonDec)
	require.NoError(t, err)
	assert.Equal(t, "value", jsonDec["key"])

	// SQL Value
	val, err := orig.Value()
	require.NoError(t, err)
	assert.IsType(t, "", val)

	// SQL Scan from string
	var scanID FlexData
	require.NoError(t, scanID.Scan(val))
	assert.Equal(t, "value", scanID["key"])

	// SQL Scan nil
	require.NoError(t, scanID.Scan(nil))
	assert.Nil(t, (map[any]any)(scanID))

	// Nil value
	var nilFlex FlexData
	val, err = nilFlex.Value()
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestHexSQLID_Presets(t *testing.T) {
	orig := HexSQLID{0xde, 0xad, 0xbe, 0xef}

	// SQL Value (explicit hex string)
	val, err := orig.Value()
	require.NoError(t, err)
	assert.Equal(t, "deadbeef", val)

	// SQL Scan from hex string
	var scanID HexSQLID
	require.NoError(t, scanID.Scan("deadbeef"))
	assert.Equal(t, orig, scanID)

	// SQL Scan from raw []byte
	require.NoError(t, scanID.Scan([]byte{0xde, 0xad, 0xbe, 0xef}))
	assert.Equal(t, orig, scanID)

	// SQL Scan nil
	require.NoError(t, scanID.Scan(nil))
	assert.Nil(t, []byte(scanID))
}

func TestBase64SQLID_Presets(t *testing.T) {
	orig := Base64SQLID{0xde, 0xad, 0xbe, 0xef}
	expectedB64 := base64.RawURLEncoding.EncodeToString([]byte(orig))

	// SQL Value (explicit base64 string)
	val, err := orig.Value()
	require.NoError(t, err)
	assert.Equal(t, expectedB64, val)

	// SQL Scan from base64 string
	var scanID Base64SQLID
	require.NoError(t, scanID.Scan(expectedB64))
	assert.Equal(t, orig, scanID)

	// SQL Scan from raw []byte
	require.NoError(t, scanID.Scan([]byte{0xde, 0xad, 0xbe, 0xef}))
	assert.Equal(t, orig, scanID)

	// SQL Scan nil
	require.NoError(t, scanID.Scan(nil))
	assert.Nil(t, []byte(scanID))
}

func TestCustomSecretID_Presets(t *testing.T) {
	orig := CustomSecretID{0xca, 0xfe, 0xba, 0xbe}
	expectedB64 := base64.RawURLEncoding.EncodeToString([]byte(orig))

	// JSON round-trip (base64)
	jsonBytes, err := json.Marshal(orig)
	require.NoError(t, err)

	var jsonDec CustomSecretID
	err = json.Unmarshal(jsonBytes, &jsonDec)
	require.NoError(t, err)
	assert.Equal(t, orig, jsonDec)

	// SQL Value (raw []byte)
	val, err := orig.Value()
	require.NoError(t, err)
	assert.Equal(t, []byte(orig), val)

	// SQL Scan from raw []byte
	var scanID CustomSecretID
	require.NoError(t, scanID.Scan([]byte{0xca, 0xfe, 0xba, 0xbe}))
	assert.Equal(t, orig, scanID)

	// SQL Scan from base64 string
	require.NoError(t, scanID.Scan(expectedB64))
	assert.Equal(t, orig, scanID)
}

func TestSQLOnlyUUID_ScanValue(t *testing.T) {
	parsedUUID, err := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	require.NoError(t, err)

	orig := SQLOnlyUUID(parsedUUID)

	// SQL Valuer
	val, err := orig.Value()
	require.NoError(t, err)
	assert.Equal(t, "6ba7b810-9dad-11d1-80b4-00c04fd430c8", val)

	// SQL Scan from raw 16 bytes
	var scanID SQLOnlyUUID
	rawBytes := []byte(parsedUUID[:])
	require.NoError(t, scanID.Scan(rawBytes))
	assert.Equal(t, orig, scanID)

	// SQL Scan from 36-byte formatted text slice
	var textScanID SQLOnlyUUID
	require.NoError(t, textScanID.Scan([]byte("6ba7b810-9dad-11d1-80b4-00c04fd430c8")))
	assert.Equal(t, orig, textScanID)

	// SQL Scan from string
	var strScanID SQLOnlyUUID
	require.NoError(t, strScanID.Scan("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
	assert.Equal(t, orig, strScanID)

	// SQL Scan nil resets to Nil UUID
	require.NoError(t, strScanID.Scan(nil))
	assert.Equal(t, SQLOnlyUUID(uuid.Nil()), strScanID)

	// SQL Scan errors
	assert.Error(t, scanID.Scan("not-a-uuid"))
	assert.Error(t, scanID.Scan([]byte("short-bytes")))
	assert.Error(t, scanID.Scan(12345))
}

func BenchmarkTokenID_MarshalText(b *testing.B) {
	tok := TokenID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = tok.MarshalText()
	}
}

func BenchmarkTokenID_UnmarshalText(b *testing.B) {
	tok := TokenID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	text, _ := tok.MarshalText()
	var dec TokenID
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = dec.UnmarshalText(text)
	}
}

func BenchmarkTokenID_MarshalBinary(b *testing.B) {
	tok := TokenID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = tok.MarshalBinary()
	}
}

func BenchmarkTokenID_UnmarshalBinary(b *testing.B) {
	raw := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	var dec TokenID
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = dec.UnmarshalBinary(raw)
	}
}

func BenchmarkSecretID_MarshalText(b *testing.B) {
	sec := SecretID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = sec.MarshalText()
	}
}

func BenchmarkSecretID_UnmarshalText(b *testing.B) {
	sec := SecretID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	text, _ := sec.MarshalText()
	var dec SecretID
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = dec.UnmarshalText(text)
	}
}
