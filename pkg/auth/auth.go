package auth

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"slices"
	"time"
)

type (
	OptFn  func(*Config)
	Config struct {
		cookieName string
		signingKey *ecdsa.PrivateKey
		expiration time.Duration
		isSliding  bool
	}
)

func NewConfig(opts ...OptFn) *Config {
	cfg := &Config{
		cookieName: "auth",
		signingKey: nil,
		expiration: time.Hour * 8,
		isSliding:  false,
	}

	for _, o := range opts {
		o(cfg)
	}

	return cfg
}

var (
	ErrSecurityStampNotMatching = errors.New("auth: Security stamp doesn't match")
	ErrKeyNotVerified           = errors.New("auth: Unable to verify auth token")
	ErrorIdentityNotFound       = errors.New("auth: No identity found for requested ID")
	ErrorSecurityStampsDiffer   = errors.New("auth: Security stamps don't match")
)

func (c *Config) SigningKey() *ecdsa.PrivateKey {
	return c.signingKey
}

func (c *Config) SigningKeyPublic() *ecdsa.PublicKey {
	return c.signingKey.Public().(*ecdsa.PublicKey)
}

func WithCookieName(name string) func(*Config) {
	return func(c *Config) {
		c.cookieName = name
	}
}

func WithSigningKey(key *ecdsa.PrivateKey) func(*Config) {
	return func(c *Config) {
		c.signingKey = key
	}
}

func WithExpiration(d time.Duration) func(*Config) {
	return func(c *Config) {
		c.expiration = d
	}
}

func IsSliding() func(*Config) {
	return func(c *Config) {
		c.isSliding = true
	}
}

func ValidateSecurityStamp(a, b []byte) bool {
	return slices.Equal(a, b)
}

func encodeAuthToken(key *ecdsa.PrivateKey, cv CookieValue) (token string, err error) {
	signedBytes, err := key.Sign(rand.Reader, cv.Digest(), nil)
	if err != nil {
		return
	}

	cv.Signature = signedBytes

	jsonBytes, err := json.Marshal(cv)
	if err != nil {
		return
	}

	return base64.StdEncoding.EncodeToString(jsonBytes), nil
}

func decodeAuthToken(key *ecdsa.PrivateKey, s string) (cv *CookieValue, err error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return
	}

	if err = json.Unmarshal(decoded, &cv); err != nil {
		return
	}

	if !ecdsa.VerifyASN1(&key.PublicKey, cv.Digest(), cv.Signature) {
		return cv, ErrKeyNotVerified
	}

	return
}
