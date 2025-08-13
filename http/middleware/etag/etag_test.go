package etag

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	ghttp "github.com/pudottapommin/golib/http"
	"github.com/stretchr/testify/require"
)

func Test_ETag_Next(t *testing.T) {
	t.Parallel()
	r := http.NewServeMux()
	r.Handle("/a", New(WithNext(func(_ http.ResponseWriter, _ *http.Request) bool {
		return true
	}))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})))

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func Test_ETag_NotStatusOk(t *testing.T) {
	t.Parallel()
	r := http.NewServeMux()
	r.Handle("/", New()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})))

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
}

func Test_ETag_NoBody(t *testing.T) {
	t.Parallel()
	r := http.NewServeMux()
	r.Handle("/", New()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})))

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func Test_ETag_NewEtag(t *testing.T) {
	t.Parallel()
	t.Run("without HeaderIfNoneMatch", func(t *testing.T) {
		t.Parallel()
		testETagNewEtag(t, false, false)
	})
	t.Run("with HeaderIfNoneMatch and not matched", func(t *testing.T) {
		t.Parallel()
		testETagNewEtag(t, true, false)
	})
	t.Run("with HeaderIfNoneMatch and matched", func(t *testing.T) {
		t.Parallel()
		testETagNewEtag(t, true, true)
	})
}

func testETagNewEtag(t *testing.T, headerIfNoneMatch, matched bool) {
	t.Helper()
	r := http.NewServeMux()
	r.Handle("/", New()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, World!"))
		w.WriteHeader(http.StatusOK)
	})))

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	require.NoError(t, err)
	if headerIfNoneMatch {
		etag := `"non-match"`
		if matched {
			etag = `"13-1831710635"`
		}
		req.Header.Set(ghttp.HeaderIfNoneMatch, etag)
	}
	r.ServeHTTP(w, req)

	if !headerIfNoneMatch || !matched {
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, `"13-1831710635"`, w.Header().Get(headerETag))
		return
	}

	require.Equal(t, http.StatusNotModified, w.Code)
	b, err := io.ReadAll(w.Body)
	require.NoError(t, err)
	require.Empty(t, b)
}

func Test_ETag_WeakEtag(t *testing.T) {
	t.Parallel()
	t.Run("without HeaderIfNoneMatch", func(t *testing.T) {
		t.Parallel()
		testETagWeakEtag(t, false, false)
	})
	t.Run("with HeaderIfNoneMatch and not matched", func(t *testing.T) {
		t.Parallel()
		testETagWeakEtag(t, true, false)
	})
	t.Run("with HeaderIfNoneMatch and matched", func(t *testing.T) {
		t.Parallel()
		testETagWeakEtag(t, true, true)
	})
}

func testETagWeakEtag(t *testing.T, headerIfNoneMatch, matched bool) {
	t.Helper()
	r := http.NewServeMux()
	r.Handle("/", New(WithWeak())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, World!"))
		w.WriteHeader(http.StatusOK)
	})))

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	require.NoError(t, err)
	if headerIfNoneMatch {
		etag := `W/"non-match"`
		if matched {
			etag = `W/"13-1831710635"`
		}
		req.Header.Set(ghttp.HeaderIfNoneMatch, etag)
	}
	r.ServeHTTP(w, req)

	if !headerIfNoneMatch || !matched {
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, `W/"13-1831710635"`, w.Header().Get(headerETag))
		return
	}

	require.Equal(t, http.StatusNotModified, w.Code)
	b, err := io.ReadAll(w.Body)
	require.NoError(t, err)
	require.Empty(t, b)
}

func Test_ETag_CustomEtag(t *testing.T) {
	t.Parallel()
	t.Run("without HeaderIfNoneMatch", func(t *testing.T) {
		t.Parallel()
		testETagCustomEtag(t, false, false)
	})
	t.Run("with HeaderIfNoneMatch and not matched", func(t *testing.T) {
		t.Parallel()
		testETagCustomEtag(t, true, false)
	})
	t.Run("with HeaderIfNoneMatch and matched", func(t *testing.T) {
		t.Parallel()
		testETagCustomEtag(t, true, true)
	})
}

func testETagCustomEtag(t *testing.T, headerIfNoneMatch, matched bool) {
	t.Helper()
	r := http.NewServeMux()
	r.Handle("/", New()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerETag, `"custom"`)
		if bytes.Equal([]byte(r.Header.Get(ghttp.HeaderIfNoneMatch)), []byte(`"custom"`)) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		_, _ = w.Write([]byte("Hello, World!"))
		w.WriteHeader(http.StatusOK)
	})))

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	require.NoError(t, err)
	if headerIfNoneMatch {
		etag := `"non-match"`
		if matched {
			etag = `"custom"`
		}
		req.Header.Set(ghttp.HeaderIfNoneMatch, etag)
	}
	r.ServeHTTP(w, req)

	if !headerIfNoneMatch || !matched {
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, `"custom"`, w.Header().Get(headerETag))
		return
	}

	require.Equal(t, http.StatusNotModified, w.Code)
	b, err := io.ReadAll(w.Body)
	require.NoError(t, err)
	require.Empty(t, b)
}

func Test_ETag_CustomEtagPut(t *testing.T) {
	t.Parallel()
	r := http.NewServeMux()
	r.Handle("/", New()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerETag, `"custom"`)
		if !bytes.Equal([]byte(r.Header.Get(ghttp.HeaderIfMatch)), []byte(`"custom"`)) {
			w.WriteHeader(http.StatusPreconditionFailed)
			return
		}
		_, _ = w.Write([]byte("Hello, World!"))
		w.WriteHeader(http.StatusOK)
	})))

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, "/", nil)
	require.NoError(t, err)
	req.Header.Set(ghttp.HeaderIfMatch, `"non-match"`)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusPreconditionFailed, w.Code)
}

// func Benchmark_Etag(b *testing.B) {
// 	r := http.NewServeMux()
// 	r.Handle("/", New()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		_, _ = w.Write([]byte("Hello, World!"))
// 		w.WriteHeader(http.StatusOK)
// 	})))
//
// 	h := r.Handler()
//
// 	w := httptest.NewRecorder()
// 	gctx := gin.CreateTestContextOnly(w, r)
// 	gctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
//
// 	b.ReportAllocs()
// 	b.ResetTimer()
//
// 	for n := 0; n < b.N; n++ {
// 		h.ServeHTTP(w, gctx.Request)
// 	}
//
// 	require.Equal(b, http.StatusOK, gctx.Writer.Status())
// 	require.Equal(b, `"13-1831710635"`, gctx.Writer.Header().Get(headerETag))
// }
