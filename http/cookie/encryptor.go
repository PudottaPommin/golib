package cookie

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"

	"golang.org/x/crypto/chacha20poly1305"
)

type Encryptor interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

type chacha20poly1305Encryptor struct {
	aead cipher.AEAD
}

func NewChacha20Poly1305Encryptor(key []byte) (Encryptor, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, errors.New("[NewChacha20Poly1305Encryptor] key length must be 32 bytes")
	}
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}
	return &chacha20poly1305Encryptor{aead: aead}, nil
}

func (c chacha20poly1305Encryptor) Encrypt(src []byte) ([]byte, error) {
	nonce := GenerateRandomKey(c.aead.NonceSize())
	return append(nonce, c.aead.Seal(nil, nonce[:], src, nil)...), nil
}

func (c chacha20poly1305Encryptor) Decrypt(src []byte) ([]byte, error) {
	data, err := c.aead.Open(nil, src[:c.aead.NonceSize()], src[c.aead.NonceSize():], nil)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type aesGcmEncryptor struct {
	aead cipher.AEAD
}

func NewDefaultEncryptor(key []byte) (Encryptor, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &aesGcmEncryptor{aead: aead}, nil
}

func (e aesGcmEncryptor) Encrypt(src []byte) ([]byte, error) {
	nonce := GenerateRandomKey(e.aead.NonceSize())
	return append(nonce, e.aead.Seal(nil, nonce, src, nil)...), nil
}

func (e aesGcmEncryptor) Decrypt(src []byte) ([]byte, error) {
	size := e.aead.NonceSize()
	if len(src) < size {
		return nil, errors.New("block size is greater than src length")
	}
	nonce := src[:size]
	src = src[size:]
	return e.aead.Open(nil, nonce, src, nil)
}
