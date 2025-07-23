package etag

import (
	"bytes"
	"hash/crc32"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pudottapommin/golib"
)

var buffers = golib.NewPool(func() *bytes.Buffer {
	return new(bytes.Buffer)
})

func New(opts ...OptsFn) gin.HandlerFunc {
	cfg := defaultConfig
	for i := range opts {
		opts[i](&cfg)
	}

	const crcPol = 0xD5828281
	crc32q := crc32.MakeTable(crcPol)

	return func(c *gin.Context) {
		if cfg.Next != nil && cfg.Next(c) {
			c.Next()
			return
		}

		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           buffers.Get(),
		}
		defer buffers.PutAndReset(w.body)
		c.Writer = w
		c.Next()

		if c.Err() != nil {
			return
		}
		if c.Writer.Status() != http.StatusOK {
			return
		}
		if c.Writer.Header().Get(headerETag) != "" {
			return
		}

		etagBuffer := buffers.Get()
		defer buffers.PutAndReset(etagBuffer)
		if cfg.Weak {
			etagBuffer.Write(weakPrefix)
		}
		etagBuffer.WriteByte('"')

		bodyLength := len(w.body.Bytes())
		if bodyLength > math.MaxUint32 {
			c.Status(http.StatusRequestEntityTooLarge)
			return
		}

		appendUint(etagBuffer, uint32(bodyLength))
		etagBuffer.WriteByte('-')
		appendUint(etagBuffer, crc32.Checksum(w.body.Bytes(), crc32q))
		etagBuffer.WriteByte('"')

		etag := etagBuffer.Bytes()
		clientEtag := []byte(c.Request.Header.Get(headerIfNoneMatch))

		if bytes.HasPrefix(clientEtag, weakPrefix) {
			if bytes.Equal(clientEtag[2:], etag) || bytes.Equal(clientEtag[2:], etag[2:]) {
				// todo: ??maybe reset body??
				c.Status(http.StatusNotModified)
				return
			}
			c.Writer.Header().Set(headerETag, string(etag))
			return
		}
		if bytes.Equal(clientEtag, etag) {
			// todo: ??maybe reset body??
			c.Status(http.StatusNotModified)
			return
		}

		c.Writer.Header().Set(headerETag, string(etag))
		if _, err := w.ResponseWriter.Write(w.body.Bytes()); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}
	}
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
	len  int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	l, err := w.body.Write(b)
	return l, err
}

func appendUint(buffer *bytes.Buffer, n uint32) {
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

	buffer.Write(buf[i:])
}
