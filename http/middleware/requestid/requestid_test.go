package requestid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const testRequestID = "test-request-id"

func emptyHandler(c *gin.Context) {
	c.String(http.StatusOK, "")
}

func Test_RequestID_New(t *testing.T) {
	r := gin.New()
	r.Use(New())
	r.GET("/", emptyHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Header().Get(headerXRequestID))
}

func Test_RequestID_PassThrough(t *testing.T) {
	r := gin.New()
	r.Use(New())
	r.GET("/", emptyHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.Header.Set(headerXRequestID, testRequestID)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, testRequestID, w.Header().Get(headerXRequestID))
}

func Test_RequestID_WithCustomGenerator(t *testing.T) {
	r := gin.New()
	r.Use(New(WithGenerator(func() string {
		return testRequestID
	})))
	r.GET("/", emptyHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, testRequestID, w.Header().Get(headerXRequestID))
}

func Test_RequestID_WithCustomHeader(t *testing.T) {
	r := gin.New()
	r.Use(New(WithHeader(testRequestID)))
	r.GET("/", emptyHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Header().Get(headerXRequestID))
}

func Test_RequestID_HandlerGet(t *testing.T) {
	r := gin.New()
	r.Use(New())
	r.GET("/", func(c *gin.Context) {
		rid := Get(c)
		require.NotEmpty(t, rid)
		c.String(http.StatusOK, "")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Header().Get(headerXRequestID))
}
