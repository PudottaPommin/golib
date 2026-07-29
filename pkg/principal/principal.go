package principal

import (
	"context"
	"net/http"
)

type (
	// IdentityResolver extracts an Identity from an incoming request. Resolve must return
	// ErrNoIdentity when no credential is present (the caller is anonymous - do not clear any
	// cookie), or ErrInvalidIdentity when a credential is present but fails verification
	// (tampered, malformed, or expired - the caller should clear it). That distinction is
	// load-bearing for callers such as http/middleware/principal.Authentication.
	IdentityResolver[T comparable] interface {
		Resolve(r *http.Request) (Identity[T], error)
	}
	// IdentityStorer persists an Identity, typically by writing a Set-Cookie header.
	IdentityStorer[T comparable] interface {
		Store(w http.ResponseWriter, r *http.Request, identity Identity[T]) error
	}
	// IdentityRevoker clears a previously stored Identity credential, e.g. on logout or when a
	// resolved credential turns out to be invalid or revoked.
	IdentityRevoker interface {
		Revoke(w http.ResponseWriter, r *http.Request) error
	}
	// IdentityStore composes the full lifecycle of an identity credential: resolve, store, and
	// revoke.
	IdentityStore[T comparable] interface {
		IdentityResolver[T]
		IdentityStorer[T]
		IdentityRevoker
	}
	// AuthenticationService authenticates a request into an Identity, typically composing an
	// IdentityResolver with a Validator (see Service).
	AuthenticationService[T comparable] interface {
		Authenticate(r *http.Request) (Identity[T], error)
	}
	// Validator inspects an already-resolved Identity and rejects it (e.g. with
	// ErrIdentityRevoked) if it should no longer be trusted.
	Validator[T comparable] interface {
		Validate(ctx context.Context, identity Identity[T]) error
	}
	// ValidatorFunc is a function adapter for Validator.
	ValidatorFunc[T comparable] func(ctx context.Context, identity Identity[T]) error
)

func (f ValidatorFunc[T]) Validate(ctx context.Context, identity Identity[T]) error {
	return f(ctx, identity)
}
