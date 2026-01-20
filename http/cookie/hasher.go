package cookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"hash"
)

type Hasher interface {
	Hash([]byte) ([]byte, error)
	Verify(data, mac []byte) error
}

type hasher struct {
	key []byte
}

func (h hasher) algo() hash.Hash {
	return hmac.New(sha256.New, h.key)
}

func (h hasher) Hash(data []byte) ([]byte, error) {
	hasher := h.algo()
	if _, err := hasher.Write(data); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

func (h hasher) Verify(data, mac []byte) error {
	data, err := h.Hash(data)
	if err != nil {
		return fmt.Errorf("[Hasher] failed to verify: %w", err)
	}
	if len(data) == len(mac) && subtle.ConstantTimeCompare(data, mac) == 1 {
		return nil
	}
	return fmt.Errorf("[Hasher] failed to verify: MAC mismatch")
}
