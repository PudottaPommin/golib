package principal_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	gcookie "github.com/pudottapommin/golib/http/cookie"

	//nolint:revive // dot-import lets calls read as if same-package; testpackage requires package principal_test
	. "github.com/pudottapommin/golib/http/middleware/principal"
	principalpkg "github.com/pudottapommin/golib/pkg/principal"
	principalcookie "github.com/pudottapommin/golib/pkg/principal/cookie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAuthService[T comparable] struct {
	identity principalpkg.Identity[T]
	err      error
}

func (f *fakeAuthService[T]) Authenticate(_ *http.Request) (principalpkg.Identity[T], error) {
	return f.identity, f.err
}

type fakeRevoker struct {
	called bool
}

func (f *fakeRevoker) Revoke(http.ResponseWriter, *http.Request) error {
	f.called = true
	return nil
}

func terminalHandler(reached *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		*reached = true
		w.WriteHeader(http.StatusOK)
	})
}

func TestAuthentication_Success(t *testing.T) {
	t.Parallel()

	user := principalpkg.NewUser(uuid.NewV4().String(), "alice", nil)
	mw := NewAuthentication(&fakeAuthService[string]{identity: user, err: nil})

	var (
		reached bool
		fromCtx principalpkg.Identity[string]
	)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		fromCtx = FromContext[string](r.Context())
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	mw.Handler(next).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.True(t, reached)
	require.NotNil(t, fromCtx)
	assert.Equal(t, user.ID(), fromCtx.ID())
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthentication_DefaultFailureIs401(t *testing.T) {
	t.Parallel()

	mw := NewAuthentication(&fakeAuthService[string]{identity: nil, err: principalpkg.ErrInvalidIdentity})

	var reached bool
	rec := httptest.NewRecorder()
	mw.Handler(terminalHandler(&reached)).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.False(t, reached)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthentication_OnFailureOverride(t *testing.T) {
	t.Parallel()

	var gotErr error
	mw := NewAuthentication(
		&fakeAuthService[string]{identity: nil, err: principalpkg.ErrInvalidIdentity},
		WithAuthenticationOnFailure[string](func(w http.ResponseWriter, _ *http.Request, err error) {
			gotErr = err
			w.WriteHeader(http.StatusTeapot)
		}),
	)

	var reached bool
	rec := httptest.NewRecorder()
	mw.Handler(terminalHandler(&reached)).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.False(t, reached)
	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.ErrorIs(t, gotErr, principalpkg.ErrInvalidIdentity)
}

func TestAuthentication_AllowAnonymous(t *testing.T) {
	t.Parallel()

	mw := NewAuthentication(
		&fakeAuthService[string]{identity: nil, err: principalpkg.ErrNoIdentity},
		WithAllowAnonymous[string](true),
	)

	var (
		reached bool
		fromCtx principalpkg.Identity[string]
	)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		fromCtx = FromContext[string](r.Context())
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	mw.Handler(next).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.True(t, reached)
	assert.Nil(t, fromCtx)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthentication_Next_Skips(t *testing.T) {
	t.Parallel()

	mw := NewAuthentication(
		&fakeAuthService[string]{identity: nil, err: principalpkg.ErrInvalidIdentity},
		WithAuthenticationNext[string](func(http.ResponseWriter, *http.Request) bool {
			return true
		}),
	)

	var reached bool
	rec := httptest.NewRecorder()
	mw.Handler(terminalHandler(&reached)).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.True(t, reached)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthentication_Revoker_CalledOnInvalid_NotOnMissing(t *testing.T) {
	t.Parallel()

	revoker := &fakeRevoker{called: false}
	mw := NewAuthentication(
		&fakeAuthService[string]{identity: nil, err: principalpkg.ErrInvalidIdentity},
		WithRevoker[string](revoker),
	)

	rec := httptest.NewRecorder()
	mw.Handler(terminalHandler(new(bool))).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.True(t, revoker.called)

	revoker = &fakeRevoker{called: false}
	mw = NewAuthentication(
		&fakeAuthService[string]{identity: nil, err: principalpkg.ErrNoIdentity},
		WithRevoker[string](revoker),
	)

	rec = httptest.NewRecorder()
	mw.Handler(terminalHandler(new(bool))).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.False(t, revoker.called)
}

func TestAuthentication_EndToEnd_CookieStore(t *testing.T) {
	t.Parallel()

	hashKey := gcookie.GenerateRandomKey(32)
	blockKey := gcookie.GenerateRandomKey(32)
	store, err := principalcookie.NewCookieStore[string]("_identity", hashKey, blockKey)
	require.NoError(t, err)

	mw := NewAuthentication(principalpkg.NewService(store), WithRevoker[string](store))

	want := principalpkg.NewUser(uuid.NewV4().String(), "alice", []byte("stamp-1"))
	storeRec := httptest.NewRecorder()
	require.NoError(t, store.Store(storeRec, httptest.NewRequest(http.MethodGet, "/", nil), want))
	setCookie := storeRec.Result().Cookies()
	require.Len(t, setCookie, 1)

	var fromCtx principalpkg.Identity[string]
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fromCtx = FromContext[string](r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(setCookie[0])
	rec := httptest.NewRecorder()
	mw.Handler(next).ServeHTTP(rec, req)

	require.NotNil(t, fromCtx)
	assert.Equal(t, want.ID(), fromCtx.ID())
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Result().Cookies())

	tamperedValue := []byte(setCookie[0].Value)
	tamperedValue[len(tamperedValue)/2] ^= 0xFF
	tamperedCookie := &http.Cookie{Name: setCookie[0].Name, Value: string(tamperedValue)}

	fromCtx = nil
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(tamperedCookie)
	rec = httptest.NewRecorder()
	mw.Handler(next).ServeHTTP(rec, req)

	assert.Nil(t, fromCtx)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	revokeCookies := rec.Result().Cookies()
	require.Len(t, revokeCookies, 1)
	assert.Equal(t, -1, revokeCookies[0].MaxAge)
}
