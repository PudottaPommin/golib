package principal

import (
	"context"
	"net/http"

	principalpkg "github.com/pudottapommin/golib/pkg/principal"
)

type (
	key string

	AuthenticationOptsFn[T comparable] func(*Authentication[T])
	Authentication[T comparable]       struct {
		// Next defines function to skip middleware when returned true
		//
		// Optional, Default: nil
		Next func(http.ResponseWriter, *http.Request) bool
		// Service resolves and validates the identity for an incoming request.
		Service principalpkg.AuthenticationService[T]
		// Revoker clears the identity credential when authentication fails with an error other
		// than principal.ErrNoIdentity - it breaks the corrupt-cookie 401 loop.
		//
		// Optional, Default: nil
		Revoker principalpkg.IdentityRevoker
		// AllowAnonymous lets a request with no credential continue to next with no identity in
		// context, instead of failing.
		//
		// Optional, Default: false
		AllowAnonymous bool
		// OnSuccess is called after a successful authentication, before next runs.
		//
		// Optional, Default: nil
		OnSuccess func(http.ResponseWriter, *http.Request, principalpkg.Identity[T])
		// OnFailure handles a failed authentication.
		//
		// Optional, Default: writes http.StatusUnauthorized
		OnFailure func(http.ResponseWriter, *http.Request, error)
	}

	AuthorizationOptsFn[T comparable] func(*Authorization[T])
	Authorization[T comparable]       struct {
		// Next defines function to skip middleware when returned true
		//
		// Optional, Default: nil
		Next func(http.ResponseWriter, *http.Request) bool
		// Authorize decides if the Identity found in context is allowed to proceed.
		//
		// Optional, Default: returns true whenever an identity is present
		Authorize func(principalpkg.Identity[T]) bool
		// OnSuccess is called after a successful authorization, before next runs.
		//
		// Optional, Default: nil
		OnSuccess func(http.ResponseWriter, *http.Request, principalpkg.Identity[T])
		// OnFailure handles a failed authorization.
		//
		// Optional, Default: writes http.StatusForbidden
		OnFailure func(http.ResponseWriter, *http.Request, principalpkg.Identity[T])
	}
)

const ContextKey key = "golib.principal"

func NewAuthentication[T comparable](
	service principalpkg.AuthenticationService[T],
	opts ...AuthenticationOptsFn[T],
) *Authentication[T] {
	m := &Authentication[T]{
		Next:           nil,
		Service:        service,
		Revoker:        nil,
		AllowAnonymous: false,
		OnSuccess:      nil,
		OnFailure: func(w http.ResponseWriter, _ *http.Request, _ error) {
			w.WriteHeader(http.StatusUnauthorized)
		},
	}
	for i := range opts {
		opts[i](m)
	}
	return m
}

// WithAuthenticationNext sets the Next hook.
func WithAuthenticationNext[T comparable](next func(http.ResponseWriter, *http.Request) bool) AuthenticationOptsFn[T] {
	return func(m *Authentication[T]) {
		m.Next = next
	}
}

// WithRevoker sets the Revoker used to clear a rejected identity credential.
func WithRevoker[T comparable](revoker principalpkg.IdentityRevoker) AuthenticationOptsFn[T] {
	return func(m *Authentication[T]) {
		m.Revoker = revoker
	}
}

// WithAllowAnonymous sets AllowAnonymous.
func WithAllowAnonymous[T comparable](allow bool) AuthenticationOptsFn[T] {
	return func(m *Authentication[T]) {
		m.AllowAnonymous = allow
	}
}

// WithAuthenticationOnSuccess sets OnSuccess.
func WithAuthenticationOnSuccess[T comparable](
	fn func(http.ResponseWriter, *http.Request, principalpkg.Identity[T]),
) AuthenticationOptsFn[T] {
	return func(m *Authentication[T]) {
		m.OnSuccess = fn
	}
}

// WithAuthenticationOnFailure sets OnFailure.
func WithAuthenticationOnFailure[T comparable](
	fn func(http.ResponseWriter, *http.Request, error),
) AuthenticationOptsFn[T] {
	return func(m *Authentication[T]) {
		m.OnFailure = fn
	}
}

func NewAuthorization[T comparable](
	opts ...AuthorizationOptsFn[T],
) *Authorization[T] {
	m := &Authorization[T]{
		Next: nil,
		Authorize: func(identity principalpkg.Identity[T]) bool {
			return identity != nil
		},
		OnSuccess: nil,
		OnFailure: func(w http.ResponseWriter, _ *http.Request, _ principalpkg.Identity[T]) {
			w.WriteHeader(http.StatusForbidden)
		},
	}
	for i := range opts {
		opts[i](m)
	}
	return m
}

// WithAuthorizationNext sets the Next hook.
func WithAuthorizationNext[T comparable](next func(http.ResponseWriter, *http.Request) bool) AuthorizationOptsFn[T] {
	return func(m *Authorization[T]) {
		m.Next = next
	}
}

// WithAuthorize sets the Authorize predicate.
func WithAuthorize[T comparable](fn func(principalpkg.Identity[T]) bool) AuthorizationOptsFn[T] {
	return func(m *Authorization[T]) {
		m.Authorize = fn
	}
}

// WithAuthorizationOnSuccess sets OnSuccess.
func WithAuthorizationOnSuccess[T comparable](
	fn func(http.ResponseWriter, *http.Request, principalpkg.Identity[T]),
) AuthorizationOptsFn[T] {
	return func(m *Authorization[T]) {
		m.OnSuccess = fn
	}
}

// WithAuthorizationOnFailure sets OnFailure.
func WithAuthorizationOnFailure[T comparable](
	fn func(http.ResponseWriter, *http.Request, principalpkg.Identity[T]),
) AuthorizationOptsFn[T] {
	return func(m *Authorization[T]) {
		m.OnFailure = fn
	}
}

func FromContext[T comparable](ctx context.Context) principalpkg.Identity[T] {
	v, ok := ctx.Value(ContextKey).(principalpkg.Identity[T])
	if !ok {
		return nil
	}
	return v
}
