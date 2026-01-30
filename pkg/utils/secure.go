package utils

import "crypto/rand"

func GenerateRandomKey(length int) []byte {
	key := make([]byte, 0)
	_, _ = rand.Read(key)
	return key
}
