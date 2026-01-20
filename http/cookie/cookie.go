// Package cookie is heavily inspired by https://github.com/gorilla/securecookie
package cookie

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"strconv"
	"time"
)

const macSeparator byte = '|'

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
		encoder:    &GobEncoder{},
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
	// HMAC
	buf = []byte(fmt.Sprintf("%s%c%d%c%s%c", name, macSeparator, sc.timestamp(), macSeparator, buf, macSeparator))
	h, err := sc.mac.Hash(buf[:len(buf)-1])
	if err != nil {
		return nil, fmt.Errorf("[Cookie] failed to Secure: %w", err)
	}
	buf = append(buf[len(name)+1:], h...)

	// Encode to base64
	if buf, err = sc.urlEncoder.Encode(buf); err != nil {
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
		return fmt.Errorf("[Cookie] failed to Decrypt: value is too long %d", len(value))
	}
	var (
		buf []byte
		err error
	)
	// Decode from base64
	if err = sc.urlEncoder.Decode([]byte(value), &buf); err != nil {
		return fmt.Errorf("[Cookie] failed to Decrypt: %w", err)
	}
	// Verify MAC
	parts := bytes.SplitN(buf, []byte{macSeparator}, 3)
	if len(parts) != 3 {
		return fmt.Errorf("[Cookie] failed to Decrypt: wrong MAC separators count")
	}
	// We remake the buffer from name + separator and part 1+2
	buf = append([]byte(name+string(macSeparator)), buf[:len(buf)-len(parts[2])-1]...)
	if err = sc.mac.Verify(buf, parts[2]); err != nil {
		return fmt.Errorf("[Cookie] failed to Decrypt: %w", err)
	}
	// Verify dates
	var t1 int64
	if t1, err = strconv.ParseInt(string(parts[0]), 10, 64); err != nil {
		return fmt.Errorf("[Cookie] failed to Decrypt: invalid timestamp format")
	}
	t2 := sc.timestamp()
	if sc.minAge != 0 && t2-t1 < sc.minAge {
		return fmt.Errorf("[Cookie] failed to Decrypt: cookie is too new")
	}
	if sc.maxAge != 0 && t2-t1 > sc.maxAge {
		return fmt.Errorf("[Cookie] failed to Decrypt: cookie has expired")
	}
	// Decrypt if encrypted
	buf = parts[1]
	if sc.encryptor != nil {
		if buf, err = sc.encryptor.Decrypt(buf); err != nil {
			return fmt.Errorf("[Cookie] failed to Decrypt: %w", err)
		}
	}
	// Deserialize
	if err = sc.encoder.Decode(buf, dst); err != nil {
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
