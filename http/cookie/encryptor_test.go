package cookie

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/chacha20poly1305"
)

func TestNewChacha20Poly1305Encryptor(t *testing.T) {
	key, err := base64.StdEncoding.DecodeString("8RVCgKJOQgFOtAWkat0EDp8Z9EgX6z1kMar3GIcTvCo=")
	assert.NoError(t, err)
	if len(key) != chacha20poly1305.KeySize {
		t.Fatalf("invalid key length: %d", len(key))
	}

	enc, err := NewChacha20Poly1305Encryptor(key)
	assert.NoError(t, err)

	const testMessage = "test message"
	b, err := enc.Encrypt([]byte(testMessage))
	assert.NoError(t, err)
	assert.NotEmpty(t, b)
	assert.NotEqual(t, testMessage, string(b))

	dec, err := enc.Decrypt(b)
	assert.NoError(t, err)
	assert.Equal(t, testMessage, string(dec))
}
