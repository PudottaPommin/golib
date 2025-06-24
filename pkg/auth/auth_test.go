package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"slices"
	"testing"

	"github.com/gofrs/uuid/v5"
)

const certStr = `
-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDCi/XEV4A/X1Mcm86YeCkJeL2JATfd0CjA5dqMOBijtsG7Hj0tkc6tF
8sTcNYLiNUugBwYFK4EEACKhZANiAASPeF6h3MiurToQk0KGylqRQW/GhT+RGjZL
je/RX54dOG31It5TMaDHxI1xx/O7JSVFgU6KX8aB5uVH2VoIZgCjQzzFEqMu6OnP
aY/ZcvurufEBuk+xY6ts/5cq/0hV4Ns=
-----END EC PRIVATE KEY-----
`

func TestGenerateSignedCookie(t *testing.T) {
	key := getEcdsaKey(t)

	cv := NewCookieValue(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000000"), "username", []byte("test2"))
	token, err := EncodeAuthToken(key, cv)
	if err != nil {
		t.Error(err)
	}

	_ = token
	// t.Log(token)
}

func TestDecodeAuthToken(t *testing.T) {
	const s = "eyJpZCI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsInVzZXJuYW1lIjoidXNlcm5hbWUiLCJzZWN1cml0eVN0YW1wIjoiZEdWemREST0iLCJ0aW1lc3RhbXAiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsInNpZ25hdHVyZSI6Ik1HVUNNUUN0VDI3UVllU3JRQkhSaURrLzU4Ymg5aFRsaTNiejRlNmxBaVpFRVZOZENUYllvY2JtdFR4S1lsMTExckVBNUpjQ01CTGJhNllPQmlRWmVGaDgxL3doRG1EMnBJTFNHVGlkUjJjTlVTQ05GUXBVbEt6OStQUGtQOW15MVl5WURNdHFOUT09In0="

	key := getEcdsaKey(t)
	cv, err := DecodeAuthToken(key, s)
	if err != nil {
		t.Error(err)
	}

	if cv.ID.String() != "00000000-0000-0000-0000-000000000000" {
		t.Error("IDs don't match")
	}

	if cv.Username != "username" {
		t.Error("Usernames don't match")
	}

	if !slices.Equal(cv.SecurityStamp, []byte("test2")) {
		t.Error("SecurityStamp doesn't match")
	}
}

func getEcdsaKey(t *testing.T) *ecdsa.PrivateKey {
	block, _ := pem.Decode([]byte(certStr))
	if block == nil {
		t.Fatal("No block in ECDSA key\n")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		t.Fatal("Unable to parse X509\n")
	}

	return key
}
