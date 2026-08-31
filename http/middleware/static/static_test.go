package static

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ghttp "github.com/pudottapommin/golib/http"
	"github.com/stretchr/testify/require"
)

type mockFile struct {
	name      string
	data      []byte
	zstdBytes []byte
	hash      string
	modTime   time.Time
	isDir     bool
	reader    *bytes.Reader
}

func newMockFile(name string, data, zstdBytes []byte, hash string, modTime time.Time) *mockFile {
	return &mockFile{
		name:      name,
		data:      data,
		zstdBytes: zstdBytes,
		hash:      hash,
		modTime:   modTime,
		reader:    bytes.NewReader(data),
	}
}

func (m *mockFile) Stat() (fs.FileInfo, error) {
	return &mockFileInfo{
		name:    m.name,
		size:    int64(len(m.data)),
		modTime: m.modTime,
		isDir:   m.isDir,
	}, nil
}

func (m *mockFile) Read(p []byte) (int, error) {
	return m.reader.Read(p)
}

func (m *mockFile) Seek(offset int64, whence int) (int64, error) {
	return m.reader.Seek(offset, whence)
}

func (m *mockFile) Close() error {
	return nil
}

func (m *mockFile) ZstdBytes() []byte {
	return m.zstdBytes
}

func (m *mockFile) Hash() string {
	return m.hash
}

type mockFileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (fi *mockFileInfo) Name() string       { return fi.name }
func (fi *mockFileInfo) Size() int64        { return fi.size }
func (fi *mockFileInfo) Mode() fs.FileMode  { return 0o444 }
func (fi *mockFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *mockFileInfo) IsDir() bool        { return fi.isDir }
func (fi *mockFileInfo) Sys() any           { return nil }

type mockFS struct {
	files map[string]*mockFile
}

func (mfs *mockFS) Open(name string) (fs.File, error) {
	f, ok := mfs.files[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return &mockFile{
		name:      f.name,
		data:      f.data,
		zstdBytes: f.zstdBytes,
		hash:      f.hash,
		modTime:   f.modTime,
		isDir:     f.isDir,
		reader:    bytes.NewReader(f.data),
	}, nil
}

type simpleReaderFile struct {
	name    string
	reader  io.Reader
	modTime time.Time
	size    int64
}

func (s *simpleReaderFile) Stat() (fs.FileInfo, error) {
	return &mockFileInfo{name: s.name, size: s.size, modTime: s.modTime, isDir: false}, nil
}

func (s *simpleReaderFile) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

func (s *simpleReaderFile) Close() error {
	return nil
}

type simpleReaderFS struct {
	file *simpleReaderFile
}

func (s *simpleReaderFS) Open(name string) (fs.File, error) {
	if name == s.file.name {
		return &simpleReaderFile{
			name:    s.file.name,
			reader:  bytes.NewReader([]byte("simple stream")),
			modTime: s.file.modTime,
			size:    s.file.size,
		}, nil
	}
	return nil, fs.ErrNotExist
}

func Test_Static_ZstdCompressed_ServedWhenAccepted(t *testing.T) {
	t.Parallel()

	uncompressedData := []byte("body { background: #fff; }")
	compressedData := []byte{0x28, 0xb5, 0x2f, 0xfd, 0x00, 0x01}
	modTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mfs := &mockFS{
		files: map[string]*mockFile{
			"style.css": newMockFile("style.css", uncompressedData, compressedData, "abc123hash", modTime),
		},
	}

	handler := New(mfs, WithEtag(), WithSetProd(true))

	req := httptest.NewRequest(http.MethodGet, "/style.css", nil)
	req.Header.Set(ghttp.HeaderAcceptEncoding, "gzip, deflate, zstd")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "zstd", rec.Header().Get(ghttp.HeaderContentEncoding))
	require.Equal(t, "text/css; charset=utf-8", rec.Header().Get(ghttp.HeaderContentType))
	require.Equal(t, ghttp.HeaderAcceptEncoding, rec.Header().Get(ghttp.HeaderVary))
	require.Equal(t, `"abc123hash"`, rec.Header().Get(ghttp.HeaderETag))
	require.Equal(t, compressedData, rec.Body.Bytes())
}

func Test_Static_Uncompressed_WithZstdBytesProviderReturningNil(t *testing.T) {
	t.Parallel()

	uncompressedData := []byte("hello uncompressed")
	modTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// ZstdBytes is nil (same as uncompressed file in asseter)
	mfs := &mockFS{
		files: map[string]*mockFile{
			"tiny.txt": newMockFile("tiny.txt", uncompressedData, nil, "hash456", modTime),
		},
	}

	handler := New(mfs)

	req := httptest.NewRequest(http.MethodGet, "/tiny.txt", nil)
	req.Header.Set(ghttp.HeaderAcceptEncoding, "gzip, zstd")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Empty(t, rec.Header().Get(ghttp.HeaderContentEncoding))
	require.Equal(t, "text/plain; charset=utf-8", rec.Header().Get(ghttp.HeaderContentType))
	require.Equal(t, uncompressedData, rec.Body.Bytes())
}

func Test_Static_AcceptEncoding_Formats(t *testing.T) {
	t.Parallel()

	uncompressedData := []byte("body { color: black; }")
	compressedData := []byte{0x28, 0xb5, 0x2f, 0xfd}
	modTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mfs := &mockFS{
		files: map[string]*mockFile{
			"app.css": newMockFile("app.css", uncompressedData, compressedData, "hash", modTime),
		},
	}

	tests := []struct {
		name           string
		acceptEncoding string
		expectZstd     bool
	}{
		{
			name:           "zstd in list",
			acceptEncoding: "gzip, deflate, br, zstd",
			expectZstd:     true,
		},
		{
			name:           "zstd with q-value",
			acceptEncoding: "gzip;q=1.0, zstd;q=0.8",
			expectZstd:     true,
		},
		{
			name:           "zstd with q=0 disabled",
			acceptEncoding: "gzip;q=1.0, zstd;q=0",
			expectZstd:     false,
		},
		{
			name:           "zstd with q=0.000 disabled",
			acceptEncoding: "zstd;q=0.000, gzip",
			expectZstd:     false,
		},
		{
			name:           "no zstd accepted",
			acceptEncoding: "gzip, deflate, br",
			expectZstd:     false,
		},
		{
			name:           "empty accept encoding",
			acceptEncoding: "",
			expectZstd:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			handler := New(mfs)
			req := httptest.NewRequest(http.MethodGet, "/app.css", nil)
			if tc.acceptEncoding != "" {
				req.Header.Set(ghttp.HeaderAcceptEncoding, tc.acceptEncoding)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			if tc.expectZstd {
				require.Equal(t, "zstd", rec.Header().Get(ghttp.HeaderContentEncoding))
				require.Equal(t, compressedData, rec.Body.Bytes())
			} else {
				require.Empty(t, rec.Header().Get(ghttp.HeaderContentEncoding))
				require.Equal(t, uncompressedData, rec.Body.Bytes())
			}
		})
	}
}

func Test_Static_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	mfs := &mockFS{
		files: map[string]*mockFile{
			"file.txt": newMockFile("file.txt", []byte("ok"), nil, "", time.Time{}),
		},
	}
	handler := New(mfs)

	disallowedMethods := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range disallowedMethods {
		req := httptest.NewRequest(method, "/file.txt", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	}

	for _, method := range []string{http.MethodGet, http.MethodHead} {
		req := httptest.NewRequest(method, "/file.txt", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	}
}

func Test_Static_NotFoundAndDirectories(t *testing.T) {
	t.Parallel()

	mfs := &mockFS{
		files: map[string]*mockFile{
			"dir": {name: "dir", isDir: true},
		},
	}
	handler := New(mfs)

	// Missing file
	req := httptest.NewRequest(http.MethodGet, "/missing.txt", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)

	// Directory request
	reqDir := httptest.NewRequest(http.MethodGet, "/dir", nil)
	recDir := httptest.NewRecorder()
	handler.ServeHTTP(recDir, reqDir)
	require.Equal(t, http.StatusNotFound, recDir.Code)
}

func Test_Static_CacheControlAndOptions(t *testing.T) {
	t.Parallel()

	mfs := &mockFS{
		files: map[string]*mockFile{
			"file.txt": newMockFile("file.txt", []byte("content"), nil, "hash123", time.Time{}),
		},
	}

	t.Run("development default", func(t *testing.T) {
		handler := New(mfs, WithSetProd(false))
		req := httptest.NewRequest(http.MethodGet, "/file.txt", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, "private, max-age=0, s-maxage=0, must-revalidate", rec.Header().Get(ghttp.HeaderCacheControl))
		require.Empty(t, rec.Header().Get(ghttp.HeaderETag))
	})

	t.Run("production with custom durations and etag", func(t *testing.T) {
		handler := New(mfs,
			WithSetProd(true, false),
			WithEtag(),
			WithMaxAge(time.Hour*48),
			WithSMaxAge(time.Hour*12),
		)
		req := httptest.NewRequest(http.MethodGet, "/file.txt", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, "private, max-age=172800, s-maxage=43200, immutable", rec.Header().Get(ghttp.HeaderCacheControl))
		require.Equal(t, `"hash123"`, rec.Header().Get(ghttp.HeaderETag))
	})

	t.Run("WithSetProd no args defaults true", func(t *testing.T) {
		handler := New(mfs, WithSetProd())
		req := httptest.NewRequest(http.MethodGet, "/file.txt", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Contains(t, rec.Header().Get(ghttp.HeaderCacheControl), "immutable")
	})
}

func Test_Static_NonReadSeekerFallback(t *testing.T) {
	t.Parallel()

	sfs := &simpleReaderFS{
		file: &simpleReaderFile{
			name:    "stream.txt",
			size:    13,
			modTime: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	handler := New(sfs)
	req := httptest.NewRequest(http.MethodGet, "/stream.txt", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, []byte("simple stream"), rec.Body.Bytes())
}

func Test_parseAcceptEncoding(t *testing.T) {
	t.Parallel()

	res := parseAcceptEncoding("gzip, deflate;q=0.5, br;q=1.0, zstd;q=0.8, invalid;q=0")
	require.True(t, res.Contains("gzip"))
	require.True(t, res.Contains("deflate"))
	require.True(t, res.Contains("br"))
	require.True(t, res.Contains("zstd"))
	require.False(t, res.Contains("invalid"))

	empty := parseAcceptEncoding("")
	require.Empty(t, empty)
}
