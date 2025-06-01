package hasher

import (
	"encoding/base64"
	"errors"
)

var (
	ErrorWrongSaltLength      = errors.New("hasher: Wrong salt size in provided hash")
	ErrorWrongSubkeyLength    = errors.New("hasher: Wrong subkey length in provided hash")
	ErrorWeakHashingAlgorithm = errors.New("hasher: Password is hashed with weak algorightm")
	ErrorSaltNotFilled        = errors.New("hasher: Salt didn't receive all bytes required to fill")
	ErrorWrongAlgorithm       = errors.New("hasher: Password hashed with wrong algorithm")
)

type PasswordVerificationResult uint8

const (
	PasswordVerificationFailed PasswordVerificationResult = iota
	PasswordVerificationSuccess
	PasswordVerificationNeedsRehash
)

type hasherAlgorithm uint8

const (
	pbkdf2Algorithm   hasherAlgorithm = 0x01
	argon2idAlgorithm hasherAlgorithm = 0x02
)

type Hasher interface {
	Verify(hash, password string) (result PasswordVerificationResult, err error)
	Hash(password string) (hash string, err error)
}

type hasher struct {
	pbkfd2   *pbkdf2Hasher
	argon2id *argon2idHasher
}

func (h *hasher) Verify(hash string, password string) (result PasswordVerificationResult, err error) {
	buffer, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return
	}

	switch buffer[0] {
	case byte(pbkdf2Algorithm):
		result, err = h.pbkfd2.verify(buffer, []byte(password))
		if result == PasswordVerificationSuccess {
			result = PasswordVerificationNeedsRehash
		}
		return
	default:
		return h.argon2id.verify(buffer, []byte(password))
	}
}

func (h *hasher) Hash(password string) (hash string, err error) {
	return h.argon2id.Hash(password)
}

func NewHasher() Hasher {
	return &hasher{argon2id: newArgon2id(), pbkfd2: newPbkdf2()}
}

func readNetworkByteOrder(buffer []byte, offset int) uint {
	return uint(int(buffer[offset])<<24|int(buffer[offset+1])<<16|int(buffer[offset+2])<<8) | uint(buffer[offset+3])
}

func writeNetworkByteOrder(buffer []byte, offset int, value uint) {
	buffer[offset] = byte(value >> 24)
	buffer[offset+1] = byte(value >> 16)
	buffer[offset+2] = byte(value >> 8)
	buffer[offset+3] = byte(value)
}

func compareSubkeysInFixedTime(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}

	length := len(left)
	var num int
	for i := range length {
		num |= int(left[i]) - int(right[i])
	}

	return num == 0
}

func (r PasswordVerificationResult) String() string {
	switch r {
	case PasswordVerificationSuccess:
		return "Success"
	case PasswordVerificationNeedsRehash:
		return "NeedsRehash"
	case PasswordVerificationFailed:
		return "Failed"
	default:
		return ""
	}
}
