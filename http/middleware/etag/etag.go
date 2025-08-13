package etag

import (
	"bytes"
	"hash/crc32"
	"math"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/valyala/bytebufferpool"
)

const (
	byteQuote = '"'
	byteDash  = '-'
)

func New(opts ...OptsFn) func(http.Handler) http.Handler {
	cfg := defaultConfig
	for i := range opts {
		opts[i](&cfg)
	}

	const crcPol = 0xD5828281
	crc32q := crc32.MakeTable(crcPol)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Next != nil && cfg.Next(w, r) {
				next.ServeHTTP(w, r)
				return
			}

			rw := &responseWriter{
				WrapResponseWriter: middleware.NewWrapResponseWriter(w, r.ProtoMajor),
				body:               bytebufferpool.Get(),
			}
			defer rw.TriggerSetStatus()
			defer bytebufferpool.Put(rw.body)
			next.ServeHTTP(rw, r)

			if rw.statusCode != nil && *rw.statusCode != http.StatusOK {
				return
			}
			if rw.Header().Get(headerETag) != "" {
				return
			}

			if rw.len == 0 {
				rw.len = len(rw.body.Bytes())
			}
			if rw.len > math.MaxUint32 {
				rw.WriteHeader(http.StatusRequestEntityTooLarge)
				return
			}
			etagBuffer := bytebufferpool.Get()
			defer bytebufferpool.Put(etagBuffer)

			if cfg.Weak {
				_, _ = etagBuffer.Write(weakPrefix)
			}
			_ = etagBuffer.WriteByte(byteQuote)

			bodyLength := len(rw.body.Bytes())
			if bodyLength > math.MaxUint32 {
				rw.WriteHeader(http.StatusRequestEntityTooLarge)
				return
			}

			appendUint(etagBuffer, uint32(bodyLength))
			_ = etagBuffer.WriteByte(byteDash)
			appendUint(etagBuffer, crc32.Checksum(rw.body.Bytes(), crc32q))
			_ = etagBuffer.WriteByte(byteQuote)

			etag := etagBuffer.Bytes()
			clientEtag := []byte(r.Header.Get(headerIfNoneMatch))

			if bytes.HasPrefix(clientEtag, weakPrefix) {
				if bytes.Equal(clientEtag[2:], etag) || bytes.Equal(clientEtag[2:], etag[2:]) {
					// todo: ??maybe reset body??
					rw.WriteHeader(http.StatusNotModified)
					return
				}
				rw.Header().Set(headerETag, string(etag))
				return
			}
			if bytes.Equal(clientEtag, etag) {
				// todo: ??maybe reset body??
				rw.WriteHeader(http.StatusNotModified)
				return
			}

			rw.Header().Set(headerETag, string(etag))
			if _, err := rw.WrapResponseWriter.Write(rw.body.Bytes()); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
		})
	}
}

type responseWriter struct {
	middleware.WrapResponseWriter
	body       *bytebufferpool.ByteBuffer
	statusCode *int
	len        int
}

func (w *responseWriter) TriggerSetStatus() {
	if w.statusCode == nil {
		return
	}
	w.WrapResponseWriter.WriteHeader(*w.statusCode)
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = &code
}

func (w *responseWriter) Write(b []byte) (int, error) {
	l, err := w.body.Write(b)
	return l, err
}

func appendUint(buffer *bytebufferpool.ByteBuffer, n uint32) {
	var b [20]byte
	buf := b[:]
	i := len(buf)
	var q uint32
	for n >= 10 {
		i--
		q = n / 10
		buf[i] = '0' + byte(n-q*10)
		n = q
	}
	i--
	buf[i] = '0' + byte(n)

	_, _ = buffer.Write(buf[i:])
}
