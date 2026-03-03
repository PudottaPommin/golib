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
	algo := h.algo()
	if _, err := algo.Write(data); err != nil {
		return nil, err
	}
	return algo.Sum(nil), nil
}

func (h hasher) Verify(data, mac []byte) error {
	hashed, err := h.Hash(data)
	if err != nil {
		return err
	}
	if len(hashed) == len(mac) && subtle.ConstantTimeCompare(hashed, mac) == 1 {
		return nil
	}
	return fmt.Errorf("failed to verify due to: MAC mismatch")
}
