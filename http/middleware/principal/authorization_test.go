package principal_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid/v5"
	//nolint:revive // dot-import lets calls read as if same-package; testpackage requires package principal_test
	. "github.com/pudottapommin/golib/http/middleware/principal"
	principalpkg "github.com/pudottapommin/golib/pkg/principal"
	"github.com/stretchr/testify/assert"
)

func requestWithIdentity[T comparable](identity principalpkg.Identity[T]) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if identity != nil {
		r = r.WithContext(context.WithValue(r.Context(), ContextKey, identity))
	}
	return r
}

func TestAuthorization_NoIdentity_DefaultsTo403(t *testing.T) {
	t.Parallel()

	mw := NewAuthorization[string]()

	var reached bool
	rec := httptest.NewRecorder()
	mw.Handler(terminalHandler(&reached)).ServeHTTP(rec, requestWithIdentity[string](nil))

	assert.False(t, reached)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAuthorization_WithIdentity_DefaultAllows(t *testing.T) {
	t.Parallel()

	mw := NewAuthorization[string]()
	user := principalpkg.NewUser(uuid.Must(uuid.NewV4()).String(), "alice", nil)

	var reached bool
	rec := httptest.NewRecorder()
	mw.Handler(terminalHandler(&reached)).ServeHTTP(rec, requestWithIdentity[string](user))

	assert.True(t, reached)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthorization_PredicateFalse(t *testing.T) {
	t.Parallel()

	mw := NewAuthorization[string](WithAuthorize(func(principalpkg.Identity[string]) bool {
		return false
	}))
	user := principalpkg.NewUser(uuid.Must(uuid.NewV4()).String(), "alice", nil)

	var reached bool
	rec := httptest.NewRecorder()
	mw.Handler(terminalHandler(&reached)).ServeHTTP(rec, requestWithIdentity[string](user))

	assert.False(t, reached)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAuthorization_PredicateTrue(t *testing.T) {
	t.Parallel()

	mw := NewAuthorization[string](WithAuthorize(func(principalpkg.Identity[string]) bool {
		return true
	}))
	user := principalpkg.NewUser(uuid.Must(uuid.NewV4()).String(), "alice", nil)

	var reached bool
	rec := httptest.NewRecorder()
	mw.Handler(terminalHandler(&reached)).ServeHTTP(rec, requestWithIdentity[string](user))

	assert.True(t, reached)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthorization_Next_Skips(t *testing.T) {
	t.Parallel()

	mw := NewAuthorization[string](WithAuthorizationNext[string](func(http.ResponseWriter, *http.Request) bool {
		return true
	}))

	var reached bool
	rec := httptest.NewRecorder()
	mw.Handler(terminalHandler(&reached)).ServeHTTP(rec, requestWithIdentity[string](nil))

	assert.True(t, reached)
	assert.Equal(t, http.StatusOK, rec.Code)
}
