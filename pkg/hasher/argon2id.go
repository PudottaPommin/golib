package hasher

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/argon2"
)

const argon2idOffset = 17

type argon2idHasher struct {
	time     uint32
	saltSize int
	memory   uint32
	threads  uint8
	keyLen   uint32
}

func (h *argon2idHasher) Hash(password string) (hash string, err error) {
	salt := make([]byte, h.saltSize)
	n, err := rand.Read(salt)
	if err != nil {
		return
	}

	if n != h.saltSize {
		return hash, ErrorSaltNotFilled
	}

	saltLen := len(salt)
	subkey := argon2.IDKey([]byte(password), salt, h.time, h.memory, h.threads, h.keyLen)

	// [0] 0x02 version
	// [1:4] time
	// [5:8] memory
	// [9:12] threads
	// [13:16] saltLen
	buffer := make([]byte, argon2idOffset+saltLen+len(subkey))
	buffer[0] = byte(argon2idAlgorithm)
	writeNetworkByteOrder(buffer, 1, uint(h.time))
	writeNetworkByteOrder(buffer, 5, uint(h.memory))
	writeNetworkByteOrder(buffer, 9, uint(h.threads))
	writeNetworkByteOrder(buffer, 13, uint(saltLen))

	copy(buffer[argon2idOffset:], salt)
	copy(buffer[argon2idOffset+saltLen:], subkey)

	hash = base64.StdEncoding.EncodeToString(buffer)
	return
}

func (h *argon2idHasher) Verify(hash, password string) (result PasswordVerificationResult, err error) {
	buffer, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return
	}

	return h.verify(buffer, []byte(password))
}

func (h *argon2idHasher) verify(hash, password []byte) (result PasswordVerificationResult, err error) {
	if hash[0] != byte(argon2idAlgorithm) {
		return PasswordVerificationFailed, ErrorWrongAlgorithm
	}

	time := uint32(readNetworkByteOrder(hash, 1))
	memory := uint32(readNetworkByteOrder(hash, 5))
	threads := uint8(readNetworkByteOrder(hash, 9))
	saltLen := int(readNetworkByteOrder(hash, 13))
	if saltLen < h.saltSize {
		return PasswordVerificationFailed, ErrorWrongSaltLength
	}

	if argon2idOffset+saltLen > len(hash) {
		return PasswordVerificationFailed, errors.New("Offset + saltlen are out of bounds of buffer")
	}

	salt := hash[argon2idOffset:(argon2idOffset + saltLen)]
	keyLen := uint32(len(hash) - argon2idOffset - saltLen)
	if keyLen < h.keyLen {
		return PasswordVerificationFailed, ErrorWrongSubkeyLength
	}

	expectedKey := hash[argon2idOffset+saltLen : argon2idOffset+saltLen+int(keyLen)]
	actualKey := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	if compareSubkeysInFixedTime(expectedKey, actualKey) {
		result = PasswordVerificationSuccess
	}

	return
}

func NewArgon2id() Hasher {
	return newArgon2id()
}

func newArgon2id() *argon2idHasher {
	return &argon2idHasher{
		time:     1,
		saltSize: 128 / 8,
		memory:   46 * 1024,
		threads:  1,
		keyLen:   256 / 8,
	}
}
