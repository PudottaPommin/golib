package compressor

import (
	"net/http"

	"github.com/klauspost/compress/gzhttp"
)

type Wrapper func(http.Handler) http.HandlerFunc
type Handler func(http.Handler) http.Handler

func New() (Handler, error) {
	gzh, err := gzhttp.NewWrapper()
	if err != nil {
		return nil, err
	}
	return func(next http.Handler) http.Handler {
		return gzh(next)
	}, nil
}

func MustNew() Handler {
	gzh, err := New()
	if err != nil {
		panic(err)
	}
	return gzh
}

func NewWithWrapper(wrapper Wrapper) Handler {
	return func(next http.Handler) http.Handler {
		return wrapper(next)
	}
}
