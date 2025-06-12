package id

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type ID [idSize]byte

const (
	idSize     = 21
	Chars32    = "abcdefghijklmnopqrstuvwxyz012345"
	Chars64    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	chars32Len = byte(32)
	chars64Len = byte(64)
)

// New return cryptographically secure random bytes.
func New() ID {
	var id ID
	_, _ = rand.Read(id[:])
	for i := range id {
		id[i] = Chars64[id[i]%chars64Len]
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
// // It will return an error if the slice isn't 21 bytes long.
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
	var x string
	err := json.Unmarshal(data, &x)
	if err == nil {
		str, e := hex.DecodeString(x)
		copy(g[:], str)
		err = e
	}
	return err
}
