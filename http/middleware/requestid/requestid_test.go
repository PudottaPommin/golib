package requestid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

const testRequestID = "test-request-id"

func emptyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func Test_RequestID_New(t *testing.T) {
	r := http.NewServeMux()
	r.Handle("/", New().Handler(http.HandlerFunc(emptyHandler)))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Header().Get(headerXRequestID))
}

func Test_RequestID_PassThrough(t *testing.T) {
	r := http.NewServeMux()
	r.Handle("/", New().Handler(http.HandlerFunc(emptyHandler)))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.Header.Set(headerXRequestID, testRequestID)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, testRequestID, w.Header().Get(headerXRequestID))
}

func Test_RequestID_WithCustomGenerator(t *testing.T) {
	r := http.NewServeMux()
	r.Handle("/", New(WithGenerator(func() string {
		return testRequestID
	})).Handler(http.HandlerFunc(emptyHandler)))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, testRequestID, w.Header().Get(headerXRequestID))
}

func Test_RequestID_WithCustomHeader(t *testing.T) {
	r := http.NewServeMux()
	r.Handle("/", New(WithHeader(testRequestID)).Handler(http.HandlerFunc(emptyHandler)))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Header().Get(headerXRequestID))
}

func Test_RequestID_HandlerGet(t *testing.T) {
	r := http.NewServeMux()
	r.Handle("/", New().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := Get(r)
		require.NotEmpty(t, rid)
		w.WriteHeader(http.StatusOK)
	})))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Header().Get(headerXRequestID))
}
