package id

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
)

type ID [idSize]byte

const (
	idSize   = 12
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

// MarshalJSON hex encodes the id.
func (g ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.String())
}

// UnmarshalJSON decodes a hex-encoded string into an id.
func (g *ID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	copy(g[:], s)
	return nil
}
