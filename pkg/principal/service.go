package principal

import (
	"context"
	"crypto/subtle"
	"net/http"
)

type (
	ServiceOptFn[T comparable] func(*Service[T])
	Service[T comparable]      struct {
		resolver  IdentityResolver[T]
		validator Validator[T]
	}
)

var _ AuthenticationService[string] = (*Service[string])(nil)

func NewService[T comparable](resolver IdentityResolver[T], opts ...ServiceOptFn[T]) *Service[T] {
	s := &Service[T]{resolver: resolver, validator: nil}
	for i := range opts {
		opts[i](s)
	}
	return s
}

func WithValidator[T comparable](v Validator[T]) ServiceOptFn[T] {
	return func(s *Service[T]) {
		s.validator = v
	}
}

func (s *Service[T]) Authenticate(r *http.Request) (Identity[T], error) {
	identity, err := s.resolver.Resolve(r)
	if err != nil {
		return nil, err
	}
	if s.validator != nil {
		if err = s.validator.Validate(r.Context(), identity); err != nil {
			return nil, err
		}
	}
	return identity, nil
}

// SecurityStampValidator returns a Validator that looks up the current security stamp for an
// identity and rejects it with ErrIdentityRevoked if it no longer matches the stamp carried by
// the identity, using a constant-time comparison.
func SecurityStampValidator[T comparable](lookup func(ctx context.Context, id T) ([]byte, error)) ValidatorFunc[T] {
	return func(ctx context.Context, identity Identity[T]) error {
		stamp, err := lookup(ctx, identity.ID())
		if err != nil {
			return err
		}
		// An empty stamp on either side must not compare equal - subtle.ConstantTimeCompare
		// treats two empty slices as a match, which would let an identity with no stamp (or a
		// lookup that found none) sail through unrevoked.
		if len(stamp) == 0 || len(identity.SecurityStamp()) == 0 {
			return ErrIdentityRevoked
		}
		if subtle.ConstantTimeCompare(stamp, identity.SecurityStamp()) != 1 {
			return ErrIdentityRevoked
		}
		return nil
	}
}
