package cookie_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"uuid"

	gcookie "github.com/pudottapommin/golib/http/cookie"
	"github.com/pudottapommin/golib/pkg/principal"

	//nolint:revive // dot-import lets calls read as if same-package; testpackage requires package cookie_test
	. "github.com/pudottapommin/golib/pkg/principal/cookie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCookieName = "_identity"

func newTestKeys() ([]byte, []byte) {
	return gcookie.GenerateRandomKey(32), gcookie.GenerateRandomKey(32)
}

func storeIdentity(t *testing.T, cs *Store[string], identity principal.Identity[string]) *http.Cookie {
	t.Helper()

	rec := httptest.NewRecorder()
	require.NoError(t, cs.Store(rec, httptest.NewRequest(http.MethodGet, "/", nil), identity))

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	return cookies[0]
}

func TestNewCookieStore_RejectsNilBlockKey(t *testing.T) {
	t.Parallel()

	hashKey, _ := newTestKeys()

	store, err := NewCookieStore[string](testCookieName, hashKey, nil)

	assert.Nil(t, store)
	assert.ErrorIs(t, err, ErrBlockKeyRequired)
}

func TestNewCookieStore_RejectsEmptyHashKey(t *testing.T) {
	t.Parallel()

	_, blockKey := newTestKeys()

	store, err := NewCookieStore[string](testCookieName, nil, blockKey)

	require.ErrorIs(t, err, ErrHashKeyRequired)
	assert.Nil(t, store)

	store, err = NewCookieStore[string](testCookieName, []byte{}, blockKey)

	require.ErrorIs(t, err, ErrHashKeyRequired)
	assert.Nil(t, store)
}

func TestCookieStore_RoundTrip(t *testing.T) {
	t.Parallel()

	hashKey, blockKey := newTestKeys()
	store, err := NewCookieStore[string](testCookieName, hashKey, blockKey)
	require.NoError(t, err)

	want := principal.NewUser(uuid.NewV4().String(), "alice", []byte("stamp-1"))
	setCookie := storeIdentity(t, store, want)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(setCookie)

	got, err := store.Resolve(req)

	require.NoError(t, err)
	assert.Equal(t, want.ID(), got.ID())
	assert.Equal(t, want.Username(), got.Username())
	assert.Equal(t, want.SecurityStamp(), got.SecurityStamp())
}

func TestCookieStore_Resolve_NoCookie(t *testing.T) {
	t.Parallel()

	hashKey, blockKey := newTestKeys()
	store, err := NewCookieStore[string](testCookieName, hashKey, blockKey)
	require.NoError(t, err)

	identity, err := store.Resolve(httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Nil(t, identity)
	assert.ErrorIs(t, err, principal.ErrNoIdentity)
}

func TestCookieStore_Resolve_Tampered(t *testing.T) {
	t.Parallel()

	hashKey, blockKey := newTestKeys()
	store, err := NewCookieStore[string](testCookieName, hashKey, blockKey)
	require.NoError(t, err)

	user := principal.NewUser(uuid.NewV4().String(), "alice", []byte("stamp-1"))
	setCookie := storeIdentity(t, store, user)

	tampered := []byte(setCookie.Value)
	tampered[len(tampered)/2] ^= 0xFF
	setCookie.Value = string(tampered)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(setCookie)

	identity, err := store.Resolve(req)

	assert.Nil(t, identity)
	assert.ErrorIs(t, err, principal.ErrInvalidIdentity)
}

func TestCookieStore_Resolve_Expired(t *testing.T) {
	t.Parallel()

	hashKey, blockKey := newTestKeys()
	store, err := NewCookieStore[string](testCookieName, hashKey, blockKey, gcookie.WithMaxAge(1))
	require.NoError(t, err)

	user := principal.NewUser(uuid.NewV4().String(), "alice", []byte("stamp-1"))
	setCookie := storeIdentity(t, store, user)

	time.Sleep(2 * time.Second)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(setCookie)

	identity, err := store.Resolve(req)

	require.ErrorIs(t, err, principal.ErrInvalidIdentity)
	assert.Nil(t, identity)
	assert.ErrorIs(t, err, gcookie.ErrCookieExpired)
}

func TestCookieStore_Revoke(t *testing.T) {
	t.Parallel()

	hashKey, blockKey := newTestKeys()
	store, err := NewCookieStore[string](testCookieName, hashKey, blockKey)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	require.NoError(t, store.Revoke(rec, httptest.NewRequest(http.MethodGet, "/", nil)))

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, testCookieName, cookies[0].Name)
	assert.Equal(t, -1, cookies[0].MaxAge)
	assert.Equal(t, "/", cookies[0].Path)
}

func TestCookieStore_CustomCookieAttributes(t *testing.T) {
	t.Parallel()

	hashKey, blockKey := newTestKeys()
	store, err := NewCookieStore[string](testCookieName, hashKey, blockKey)
	require.NoError(t, err)

	store.
		WithCookiePath("/account").
		WithCookieDomain("example.com").
		WithCookieSameSite(http.SameSiteLaxMode).
		WithCookieHTTPOnly(false).
		WithCookieSecure(false)

	user := principal.NewUser(uuid.NewV4().String(), "alice", []byte("stamp-1"))
	setCookie := storeIdentity(t, store, user)

	assert.Equal(t, "/account", setCookie.Path)
	assert.Equal(t, "example.com", setCookie.Domain)
	assert.Equal(t, http.SameSiteLaxMode, setCookie.SameSite)
	assert.False(t, setCookie.HttpOnly)
	assert.False(t, setCookie.Secure)
}
