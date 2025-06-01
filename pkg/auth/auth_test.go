package auth_test

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/pudottapommin/golib/pkg/auth"
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

	cv := auth.NewCookieValue(uuid.MustParse("00000000-0000-0000-0000-000000000000"), []byte("test2"))
	token, err := auth.EncodeAuthToken(key, cv)
	if err != nil {
		t.Error(err)
	}

	_ = token
	t.Log(token)
}

func TestDecodeAuthToken(t *testing.T) {
	const s = "eyJpZCI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsInNlY3VyaXR5U3RhbXAiOiJkR1Z6ZERJPSIsInRpbWVzdGFtcCI6IjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwic2lnbmF0dXJlIjoiTUdVQ01CK1o3MmxZNzRTelMzWHNuSWZsYURHTC92aVBwTHJRaEMwUHA1SUtHdFVTWmtMMUNhMVNLL0d2VFJxMWw0ckhJZ0l4QU0vV1lRK0FxTXJNdytRRWR0bUpRL3lIa1lzREYyNU1PUDJrWDQwSDB0bklFMmEyZzBLRUJiS1NZOEJ0cmRQbDFRPT0ifQ=="

	key := getEcdsaKey(t)
	cv, err := auth.DecodeAuthToken(key, s)
	if err != nil {
		t.Error(err)
	}

	if cv.ID.String() != "00000000-0000-0000-0000-000000000000" {
		t.Error("IDs don't match")
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
