package principal_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid/v5"
	//nolint:revive // dot-import lets calls read as if same-package; testpackage requires package principal_test
	. "github.com/pudottapommin/golib/pkg/principal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeResolver struct {
	identity Identity[string]
	err      error
}

func (f *fakeResolver) Resolve(_ *http.Request) (Identity[string], error) {
	return f.identity, f.err
}

func TestService_Authenticate_ResolverError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	svc := NewService(&fakeResolver{identity: nil, err: wantErr})

	identity, err := svc.Authenticate(httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Nil(t, identity)
	assert.ErrorIs(t, err, wantErr)
}

func TestService_Authenticate_NoValidator(t *testing.T) {
	t.Parallel()

	user := NewUser(uuid.Must(uuid.NewV4()).String(), "alice", []byte("stamp"))
	svc := NewService(&fakeResolver{identity: user, err: nil})

	identity, err := svc.Authenticate(httptest.NewRequest(http.MethodGet, "/", nil))

	require.NoError(t, err)
	assert.Equal(t, user, identity)
}

func TestService_Authenticate_ValidatorRejects(t *testing.T) {
	t.Parallel()

	user := NewUser(uuid.Must(uuid.NewV4()).String(), "alice", []byte("stamp"))
	svc := NewService(
		&fakeResolver{identity: user, err: nil},
		WithValidator[string](ValidatorFunc[string](func(_ context.Context, _ Identity[string]) error {
			return ErrIdentityRevoked
		})),
	)

	identity, err := svc.Authenticate(httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Nil(t, identity)
	assert.ErrorIs(t, err, ErrIdentityRevoked)
}

func TestService_Authenticate_ValidatorAccepts(t *testing.T) {
	t.Parallel()

	user := NewUser(uuid.Must(uuid.NewV4()).String(), "alice", []byte("stamp"))
	svc := NewService(
		&fakeResolver{identity: user, err: nil},
		WithValidator[string](ValidatorFunc[string](func(_ context.Context, _ Identity[string]) error {
			return nil
		})),
	)

	identity, err := svc.Authenticate(httptest.NewRequest(http.MethodGet, "/", nil))

	require.NoError(t, err)
	assert.Equal(t, user, identity)
}

func TestSecurityStampValidator_Match(t *testing.T) {
	t.Parallel()

	user := NewUser(uuid.Must(uuid.NewV4()).String(), "alice", []byte("current-stamp"))
	validator := SecurityStampValidator[string](func(_ context.Context, _ string) ([]byte, error) {
		return []byte("current-stamp"), nil
	})

	err := validator.Validate(t.Context(), user)

	assert.NoError(t, err)
}

func TestSecurityStampValidator_Mismatch(t *testing.T) {
	t.Parallel()

	user := NewUser(uuid.Must(uuid.NewV4()).String(), "alice", []byte("old-stamp"))
	validator := SecurityStampValidator[string](func(_ context.Context, _ string) ([]byte, error) {
		return []byte("current-stamp"), nil
	})

	err := validator.Validate(t.Context(), user)

	assert.ErrorIs(t, err, ErrIdentityRevoked)
}

func TestSecurityStampValidator_LookupError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("lookup failed")
	user := NewUser(uuid.Must(uuid.NewV4()).String(), "alice", []byte("stamp"))
	validator := SecurityStampValidator[string](func(_ context.Context, _ string) ([]byte, error) {
		return nil, wantErr
	})

	err := validator.Validate(t.Context(), user)

	assert.ErrorIs(t, err, wantErr)
}

func TestSecurityStampValidator_BothEmpty_FailsClosed(t *testing.T) {
	t.Parallel()

	user := NewUser(uuid.Must(uuid.NewV4()).String(), "alice", nil)
	validator := SecurityStampValidator(func(_ context.Context, _ string) ([]byte, error) {
		return nil, nil
	})

	err := validator.Validate(t.Context(), user)

	assert.ErrorIs(t, err, ErrIdentityRevoked)
}

func TestSecurityStampValidator_LookupEmpty_FailsClosed(t *testing.T) {
	t.Parallel()

	user := NewUser(uuid.Must(uuid.NewV4()).String(), "alice", []byte("stamp"))
	validator := SecurityStampValidator(func(_ context.Context, _ string) ([]byte, error) {
		return nil, nil
	})

	err := validator.Validate(t.Context(), user)

	assert.ErrorIs(t, err, ErrIdentityRevoked)
}

func TestSecurityStampValidator_IdentityEmpty_FailsClosed(t *testing.T) {
	t.Parallel()

	user := NewUser(uuid.Must(uuid.NewV4()).String(), "alice", nil)
	validator := SecurityStampValidator(func(_ context.Context, _ string) ([]byte, error) {
		return []byte("current-stamp"), nil
	})

	err := validator.Validate(t.Context(), user)

	assert.ErrorIs(t, err, ErrIdentityRevoked)
}
