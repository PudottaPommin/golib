package hasher

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"hash"
	"unsafe"

	"golang.org/x/crypto/pbkdf2"
)

type (
	pbkdf2HashAlgorithm uint8
	pbkdf2Hasher        struct {
		algorithm pbkdf2HashAlgorithm
		saltSize  int
		keyLen    int
		iterCount int
	}
)

const (
	AlgorithmHMACSHA512 pbkdf2HashAlgorithm = iota
	AlgorithmHMACSHA256
	AlgorithmHMACSHA1
)

const pbkdf2Offset = 13

func NewPbkdf2() Hasher {
	return &pbkdf2Hasher{
		algorithm: AlgorithmHMACSHA512,
		saltSize:  128 / 8,
		keyLen:    256 / 8,
		iterCount: 210_000,
	}
}

func NewPbkdf2FromAlgo(algorithm pbkdf2HashAlgorithm) Hasher {
	return &pbkdf2Hasher{
		algorithm: algorithm,
		saltSize:  128 / 8,
		keyLen:    256 / 8,
		iterCount: 210_000,
	}
}

func (h *pbkdf2Hasher) Hash(password string) (hash string, err error) {
	salt := make([]byte, h.saltSize)
	_, _ = rand.Read(salt)
	// if n != h.saltSize {
	// 	return hash, ErrorSaltNotFilled
	// }

	prf := h.algorithm
	pwBytes := *(*[]byte)(unsafe.Pointer(&password))
	saltLen := len(salt)
	subkey := pbkdf2.Key(pwBytes, salt, h.iterCount, h.keyLen, prf.hashFunction())

	// [0] 0x01
	// [1:4] algorithm
	// [5:8] iteration count
	// [9:12] saltSize
	buffer := buffers.Get()
	defer func() {
		buffer.Reset()
		buffers.Put(buffer)
	}()
	buffer.Grow(pbkdf2Offset + saltLen + len(subkey))
	buffer.WriteByte(byte(pbkdf2Algorithm))
	writeNetworkByteOrder(buffer, 1, uint(h.algorithm))
	writeNetworkByteOrder(buffer, 5, uint(h.iterCount))
	writeNetworkByteOrder(buffer, 9, uint(saltLen))
	buffer.Write(salt)
	buffer.Write(subkey)

	hash = base64.StdEncoding.EncodeToString(buffer.Bytes())
	return
}

func (h *pbkdf2Hasher) Verify(hash, password string) (result PasswordVerificationResult, err error) {
	buffer, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return
	}

	return h.verify(buffer, []byte(password))
}

func (h *pbkdf2Hasher) verify(hash, password []byte) (result PasswordVerificationResult, err error) {
	if hash[0] != byte(pbkdf2Algorithm) {
		return PasswordVerificationFailed, ErrorWrongAlgorithm
	}

	prf := pbkdf2HashAlgorithm(readNetworkByteOrder(hash, 1))
	iterCount := int(readNetworkByteOrder(hash, 5))
	saltLen := int(readNetworkByteOrder(hash, 9))
	if saltLen < h.saltSize {
		return PasswordVerificationFailed, ErrorWrongSaltLength
	}

	salt := hash[pbkdf2Offset : pbkdf2Offset+saltLen]
	expectedKey := hash[pbkdf2Offset+saltLen:]
	keyLen := len(expectedKey)
	if keyLen < h.keyLen {
		return PasswordVerificationFailed, ErrorWrongSubkeyLength
	}

	actualKey := pbkdf2.Key(password, salt, iterCount, keyLen, prf.hashFunction())

	if compareSubkeysInFixedTime(expectedKey, actualKey) {
		result = PasswordVerificationSuccess
	}

	if result == PasswordVerificationSuccess && prf != AlgorithmHMACSHA512 {
		result = PasswordVerificationNeedsRehash
	}

	return
}

func (e pbkdf2HashAlgorithm) String() string {
	switch e {
	case AlgorithmHMACSHA512:
		return "HMAC-SHA-512"
	case AlgorithmHMACSHA256:
		return "HMAC-SHA-256"
	case AlgorithmHMACSHA1:
		return "HMAC-SHA-1"
	default:
		return "HMAC-SHA-512"
	}
}

func (e pbkdf2HashAlgorithm) hashFunction() func() hash.Hash {
	switch e {
	case AlgorithmHMACSHA512:
		return sha512.New
	case AlgorithmHMACSHA256:
		return sha256.New
	case AlgorithmHMACSHA1:
		return sha1.New
	default:
		return sha512.New
	}
}
