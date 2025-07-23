package id

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type ID [idSize]byte

const (
	idSize   = 32
	Chars    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsLen = byte(62)
)

func New() ID {
	var id ID
	_, _ = rand.Read(id[:])
	for i := range id {
		id[i] = Chars[id[i]%charsLen]
	}
	return id
}

func (g ID) Bytes() []byte {
	return g[:]
}

// String returns a canonical string representation of the ID
func (g ID) String() string {
	// gg := g[:]
	// return *(*string)(unsafe.Pointer(&gg))
	return string(g[:])
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (g ID) MarshalBinary() (data []byte, err error) {
	return g.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (g *ID) UnmarshalBinary(data []byte) error {
	if len(data) != idSize {
		return fmt.Errorf("id: incorrect byte length of %d", len(data))
	}
	copy(g[:], data)
	return nil
}

// MarshalText hex encodes the id.
func (g ID) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(g.Bytes())), nil
}

// UnmarshalText decodes a hex-encoded string into an id.
func (g *ID) UnmarshalText(data []byte) error {
	str, err := hex.DecodeString(string(data))
	if err != nil {
		return err
	}
	copy(g[:], str)
	return nil
}
