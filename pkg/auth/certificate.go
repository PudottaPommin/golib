package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

func NewSigningKey() (*ecdsa.PrivateKey, error) {
	return NewSigningKeyCurve(elliptic.P521())
}

func NewSigningKeyCurve(c elliptic.Curve) (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(c, rand.Reader)
}
