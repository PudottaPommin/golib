package cookie

import (
	"encoding/base64"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecureCookie_Secure(t *testing.T) {
	const cookieName = "secure_cookie"
	message := "Čakanie na Godota"
	key := []byte{255, 147, 74, 252, 113, 10, 175, 93, 193, 253, 244, 6, 171, 222, 124, 226, 43, 248, 96, 116, 70, 175, 231, 56, 214, 173, 170, 186, 58, 86, 44, 156}

	sc, err := New(key, key)
	assert.NoError(t, err)
	assert.NotNil(t, sc)

	secured, err := sc.Secure(cookieName, message)
	assert.NoError(t, err)
	assert.NotNil(t, secured)

	var dst string
	err = sc.Decrypt(cookieName, string(secured), &dst)
	assert.NoError(t, err)
	assert.Equal(t, message, dst)
}

func TestEncryptor_Encryption(t *testing.T) {
	key := []byte("1234567890123456")
	enc, err := NewDefaultEncryptor(key)
	assert.NoError(t, err)
	assert.NotNil(t, enc)

	message := "test message"
	encrypted, err := enc.Encrypt([]byte(message))
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := enc.Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, message, string(decrypted))
}

func TestGobEncoder(t *testing.T) {
	message := "test message"

	enc := GobEncoder{}
	encoded, err := enc.Encode(message)
	assert.NoError(t, err)
	assert.NotEmpty(t, encoded)

	var decodedMessage string
	err = enc.Decode(encoded, &decodedMessage)
	assert.NoError(t, err)
	assert.Equal(t, message, decodedMessage)
}

func TestJSONEncoder(t *testing.T) {
	message := "test message"

	enc := JSONEncoder{}
	encoded, err := enc.Encode(message)
	assert.NoError(t, err)
	assert.NotEmpty(t, encoded)

	var decodedMessage string
	err = enc.Decode(encoded, &decodedMessage)
	assert.NoError(t, err)
	assert.Equal(t, message, decodedMessage)
}

func TestNoopEncoder(t *testing.T) {
	message := []byte("test message")

	enc := NoopEncoder{}
	encoded, err := enc.Encode(message)
	assert.NoError(t, err)
	assert.NotEmpty(t, encoded)

	var decodedMessage []byte
	err = enc.Decode(encoded, &decodedMessage)
	assert.NoError(t, err)
	assert.Equal(t, message, decodedMessage)
}

type mockEncryptor struct {
	data []byte
}

func (m *mockEncryptor) Encrypt(src []byte) ([]byte, error) {
	return m.data, nil
}

func (m *mockEncryptor) Decrypt(src []byte) ([]byte, error) {
	return src, nil
}

func TestMACMismatchWithSeparator(t *testing.T) {
	hashKey := []byte("very-secret-hash-key-32-bytes-!!")
	// We use a mock encryptor that returns data containing the separator '|'
	mockEnc := &mockEncryptor{data: []byte("part1|part2")}

	sc, err := New(hashKey, nil, WithEncryptor(mockEnc), WithEncoder(&NoopEncoder{}))
	assert.NoError(t, err)

	name := "test-cookie"
	value := []byte("some-value")

	secured, err := sc.Secure(name, value)
	assert.NoError(t, err)

	var decrypted []byte
	err = sc.Decrypt(name, string(secured), &decrypted)
	assert.NoError(t, err, "Should be able to decrypt even if data contains separator")
	assert.Equal(t, mockEnc.data, decrypted)
}

func TestSecureCookie_DecryptErrors(t *testing.T) {
	key := make([]byte, 32)
	sc, _ := New(key, key)

	t.Run("Value too long", func(t *testing.T) {
		err := sc.Decrypt("test", string(make([]byte, 5000)), nil)
		assert.ErrorIs(t, err, ErrValueTooLong)
	})

	t.Run("Malformed base64", func(t *testing.T) {
		err := sc.Decrypt("test", "invalid-base64-!!!", nil)
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrCookieTooShort)
	})

	t.Run("Cookie too short", func(t *testing.T) {
		val := base64.URLEncoding.EncodeToString([]byte("short"))
		err := sc.Decrypt("test", val, nil)
		assert.ErrorIs(t, err, ErrCookieTooShort)
	})

	t.Run("Value length mismatch", func(t *testing.T) {
		// [ts:8][vlen:4]...
		buf := make([]byte, 12)
		binary.BigEndian.PutUint32(buf[8:], 100) // vlen = 100, but buf is only 12
		val := base64.URLEncoding.EncodeToString(buf)
		err := sc.Decrypt("test", val, nil)
		assert.ErrorIs(t, err, ErrValueLengthMismatch)
	})

	t.Run("MAC verification failure", func(t *testing.T) {
		secured, _ := sc.Secure("test", "message")
		// Decode first to tamper with payload
		var buf []byte
		_ = sc.urlEncoder.Decode([]byte(secured), &buf)
		buf[len(buf)-1] ^= 0xFF

		// Re-encode
		tampered, _ := sc.urlEncoder.Encode(buf)
		err := sc.Decrypt("test", string(tampered), nil)
		assert.ErrorIs(t, err, ErrVerificationFailed)
	})

	t.Run("Expired cookie", func(t *testing.T) {
		scExpired, _ := New(key, key, WithMaxAge(-1)) // -1 means always expired
		secured, _ := sc.Secure("test", "message")
		err := scExpired.Decrypt("test", string(secured), nil)
		assert.ErrorIs(t, err, ErrCookieExpired)
	})
}
