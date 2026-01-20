package cookie

import (
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
