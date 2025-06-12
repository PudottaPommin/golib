package auth

import "crypto/ecdsa"

func EncodeAuthToken(key *ecdsa.PrivateKey, cv *CookieValue) (token string, err error) {
	return encodeAuthToken(key, cv)
}

func DecodeAuthToken(key *ecdsa.PrivateKey, s string) (cv *CookieValue, err error) {
	return decodeAuthToken(key, s)
}
