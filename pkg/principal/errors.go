package principal

import "errors"

var (
	// ErrNoIdentity indicates no credential was present on the request. The caller is anonymous
	// - callers must not clear any cookie in response to this error.
	ErrNoIdentity = errors.New("principal: no identity present")
	// ErrInvalidIdentity indicates a credential was present but failed verification (tampered,
	// malformed, or otherwise unreadable). Callers should clear the cookie in response.
	ErrInvalidIdentity = errors.New("principal: identity is invalid")
	// ErrIdentityRevoked indicates a credential verified successfully but its security stamp no
	// longer matches, meaning it was revoked after issuance. Callers should clear the cookie.
	ErrIdentityRevoked = errors.New("principal: identity has been revoked")
)
