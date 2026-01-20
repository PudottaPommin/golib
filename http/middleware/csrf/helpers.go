package csrf

import (
	"crypto/subtle"

	"github.com/pudottapommin/golib/http/cookie"
)

func validateToken(realToken, requestToken []byte) bool {
	if len(realToken) != len(requestToken) {
		return false
	}
	return subtle.ConstantTimeCompare(realToken, requestToken) == 1
}

func xorToken(token, mask []byte) []byte {
	n := len(token)
	if len(mask) < n {
		n = len(mask)
	}
	masked := make([]byte, n)
	for i := range n {
		masked[i] = token[i] ^ mask[i]
	}
	return masked
}

func maskToken(token []byte) []byte {
	mask := cookie.GenerateRandomKey(defaultTokenLength)
	return append(mask, xorToken(mask, token)...)
}

func unmaskToken(masked []byte) []byte {
	if len(masked) != 2*defaultTokenLength {
		return nil
	}
	mask := masked[:defaultTokenLength]
	token := masked[defaultTokenLength:]
	return xorToken(token, mask)
}
