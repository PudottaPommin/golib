// Package cookie is heavily inspired by https://github.com/gorilla/securecookie
package cookie

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

var (
	ErrValueTooLong        = errors.New("value is too long")
	ErrCookieTooShort      = errors.New("cookie is too short")
	ErrValueLengthMismatch = errors.New("value length mismatch")
	ErrCookieTooNew        = errors.New("cookie is too new")
	ErrCookieExpired       = errors.New("cookie has expired")
	ErrDecryptionFailed    = errors.New("decryption failed")
	ErrVerificationFailed  = errors.New("mac verification failed")
)

type Cookie struct {
	encoder    Encoder
	encryptor  Encryptor
	mac        Hasher
	urlEncoder Encoder

	maxLength int
	maxAge    int64
	minAge    int64
}

func New(hashKey, blockKey []byte, options ...OptFn) (sc *Cookie, err error) {
	sc = &Cookie{
		mac:        &hasher{key: hashKey},
		encoder:    &JSONEncoder{},
		urlEncoder: &Base64Encoder{},
		maxLength:  4096,
	}
	for _, opt := range options {
		opt(sc)
	}
	if sc.encryptor == nil && blockKey != nil {
		if sc.encryptor, err = NewDefaultEncryptor(blockKey); err != nil {
			return
		}
	}
	return
}

// Binary format:
// [timestamp: 8][val_len: 4][value: N][mac: M]
// Total size = 12 + N + M

func (sc *Cookie) Secure(name string, value any) ([]byte, error) {
	var (
		buf []byte
		err error
	)
	// Serialize
	if buf, err = sc.encoder.Encode(value); err != nil {
		return nil, fmt.Errorf("[Cookie] failed to Secure: %w", err)
	}
	// Encrypt
	if sc.encryptor != nil {
		if buf, err = sc.encryptor.Encrypt(buf); err != nil {
			return nil, fmt.Errorf("[Cookie] failed to Secure: %w", err)
		}
	}

	ts := sc.timestamp()
	
	// Create payload for signing
	// We sign: name + ts(8) + value
	sigBuf := make([]byte, len(name)+8+len(buf))
	copy(sigBuf, name)
	binary.BigEndian.PutUint64(sigBuf[len(name):], uint64(ts))
	copy(sigBuf[len(name)+8:], buf)

	h, err := sc.mac.Hash(sigBuf)
	if err != nil {
		return nil, fmt.Errorf("[Cookie] failed to Secure: %w", err)
	}

	// Final binary blob: [ts:8][vlen:4][val:N][mac:M]
	final := make([]byte, 8+4+len(buf)+len(h))
	binary.BigEndian.PutUint64(final[0:], uint64(ts))
	binary.BigEndian.PutUint32(final[8:], uint32(len(buf)))
	copy(final[12:], buf)
	copy(final[12+len(buf):], h)

	// Encode to base64
	if buf, err = sc.urlEncoder.Encode(final); err != nil {
		return nil, fmt.Errorf("[Cookie] failed to Secure: %w", err)
	}
	// Check length, if provided
	if sc.maxLength != 0 && len(buf) > sc.maxLength {
		return nil, fmt.Errorf("[Cookie] failed to Secure: value is too long %d", len(buf))
	}

	return buf, nil
}

func (sc *Cookie) Decrypt(name, value string, dst any) error {
	if sc.maxLength != 0 && len(value) > sc.maxLength {
		return fmt.Errorf("[Cookie] failed to Decrypt %d: %w", len(value), ErrValueTooLong)
	}
	var (
		buf []byte
		err error
	)
	// Decode from base64
	if err = sc.urlEncoder.Decode([]byte(value), &buf); err != nil {
		return fmt.Errorf("[Cookie] failed to Decrypt: %w", err)
	}

	if len(buf) < 12 {
		return fmt.Errorf("[Cookie] failed to Decrypt: %w", ErrCookieTooShort)
	}

	ts := int64(binary.BigEndian.Uint64(buf[0:8]))
	vlen := int(binary.BigEndian.Uint32(buf[8:12]))

	if len(buf) < 12+vlen {
		return fmt.Errorf("[Cookie] failed to Decrypt: %w", ErrValueLengthMismatch)
	}

	val := buf[12 : 12+vlen]
	mac := buf[12+vlen:]

	// Verify MAC
	// We sign: name + ts(8) + value
	sigBuf := make([]byte, len(name)+8+len(val))
	copy(sigBuf, name)
	binary.BigEndian.PutUint64(sigBuf[len(name):], uint64(ts))
	copy(sigBuf[len(name)+8:], val)

	if err = sc.mac.Verify(sigBuf, mac); err != nil {
		return fmt.Errorf("[Cookie] failed to Decrypt: %w: %w", ErrVerificationFailed, err)
	}

	// Verify dates
	t2 := sc.timestamp()
	if sc.minAge != 0 && t2-ts < sc.minAge {
		return fmt.Errorf("[Cookie] failed to Decrypt: %w", ErrCookieTooNew)
	}
	if sc.maxAge != 0 && t2-ts > sc.maxAge {
		return fmt.Errorf("[Cookie] failed to Decrypt: %w", ErrCookieExpired)
	}

	// Decrypt if encrypted
	if sc.encryptor != nil {
		if val, err = sc.encryptor.Decrypt(val); err != nil {
			return fmt.Errorf("[Cookie] failed to Decrypt: %w: %w", ErrDecryptionFailed, err)
		}
	}
	// Deserialize
	if err = sc.encoder.Decode(val, dst); err != nil {
		return fmt.Errorf("[Cookie] failed to Decrypt: %w", err)
	}
	return nil
}

func (sc *Cookie) timestamp() int64 {
	return time.Now().UTC().Unix()
}

func GenerateRandomKey(length int) []byte {
	key := make([]byte, length)
	_, _ = rand.Read(key)
	return key
}
