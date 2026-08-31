package static

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	ghttp "github.com/pudottapommin/golib/http"
	"github.com/pudottapommin/golib/pkg/set"
	"github.com/pudottapommin/golib/pkg/utils"
)

type (
	ZstdBytesProvider interface {
		ZstdBytes() []byte
	}
	HashedFileProvider interface {
		Hash() string
	}
)

func (m *mw) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	m.handleRequest(w, r, r.URL.Path)
}

func (m *mw) handleRequest(w http.ResponseWriter, r *http.Request, name string) {
	f, err := m.fs.Open(utils.PathJoinRelX(name))
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if fi.IsDir() {
		http.NotFound(w, r)
		return
	}

	var rs io.ReadSeeker
	if rseeker, ok := f.(io.ReadSeeker); ok {
		rs = rseeker
	} else {
		data, err := io.ReadAll(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rs = bytes.NewReader(data)
	}

	m.serveFile(w, r, fi.Name(), fi.ModTime(), fi.Size(), rs)
}

func (m *mw) serveFile(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	modTime time.Time,
	size int64,
	rs io.ReadSeeker,
) {
	setWellKnownContentType(w, name)
	w.Header().Add(ghttp.HeaderVary, ghttp.HeaderAcceptEncoding)
	encodings := parseAcceptEncoding(r.Header.Get(ghttp.HeaderAcceptEncoding))
	if zstdBytes, ok := checkZSTD(encodings, rs); ok {
		rd := bytes.NewReader(zstdBytes)
		if w.Header().Get(ghttp.HeaderContentType) == "" {
			w.Header().Set(ghttp.HeaderContentType, "application/octet-stream")
		}
		w.Header().Set(ghttp.HeaderContentEncoding, "zstd")
		setFileHeaders(w, m, rs, len(zstdBytes))
		http.ServeContent(w, r, name, modTime, rd)
		return
	}
	setFileHeaders(w, m, rs, int(size))
	http.ServeContent(w, r, name, modTime, rs)
}

func checkZSTD(encodings set.Set[string], rs io.ReadSeeker) ([]byte, bool) {
	if encodings.Contains("zstd") {
		if compressed, ok := rs.(ZstdBytesProvider); ok {
			if zstdBytes := compressed.ZstdBytes(); len(zstdBytes) > 0 {
				return zstdBytes, true
			}
		}
	}
	return nil, false
}

func parseAcceptEncoding(s string) set.Set[string] {
	types := make(set.Set[string])
	for p := range strings.SplitSeq(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		name, params, found := strings.Cut(p, ";")
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if found {
			params = strings.TrimSpace(params)
			if strings.HasPrefix(params, "q=0") {
				qVal := strings.TrimPrefix(params, "q=")
				if qVal == "0" || qVal == "0.0" || qVal == "0.00" || qVal == "0.000" {
					continue
				}
			}
		}
		types.Add(name)
	}
	return types
}

func setWellKnownContentType(w http.ResponseWriter, file string) {
	mimeType := ghttp.ResolveWellKnownMimeType(path.Ext(file))
	if mimeType != "" {
		w.Header().Set(ghttp.HeaderContentType, mimeType)
	}
}

func setFileHeaders(w http.ResponseWriter, m *mw, rs io.ReadSeeker, size int) {
	w.Header().Set(ghttp.HeaderContentLength, fmt.Sprintf("%d", size))
	if hf, ok := rs.(HashedFileProvider); m.etag && ok {
		w.Header().Set(ghttp.HeaderETag, fmt.Sprintf(`"%s"`, hf.Hash()))
	}
	if m.isProd {
		w.Header().
			Set(ghttp.HeaderCacheControl, fmt.Sprintf(`private, max-age=%d, s-maxage=%d, immutable`, int(m.maxAge.Seconds()), int(m.sMaxAge.Seconds())))
	} else {
		w.Header().Set(ghttp.HeaderCacheControl, fmt.Sprintf(`private, max-age=%d, s-maxage=%d, must-revalidate`, 0, 0))
	}
}
