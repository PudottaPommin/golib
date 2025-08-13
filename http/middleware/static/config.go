package static

import (
	"io/fs"
	"time"
)

type (
	OptsFn func(*mw)
	mw     struct {
		fs      fs.FS
		etag    bool
		isProd  bool
		maxAge  time.Duration
		sMaxAge time.Duration
	}
)

func New(fs fs.FS, opts ...OptsFn) *mw {
	m := &mw{fs: fs, maxAge: time.Hour * 24 * 14, sMaxAge: time.Hour * 24 * 1}
	for i := range opts {
		opts[i](m)
	}
	return m
}

func WithSetProd(isProd ...bool) OptsFn {
	return func(c *mw) {
		switch len(isProd) {
		case 0:
			c.isProd = true
		case 1:
			c.isProd = isProd[0]
		default:
			for _, b := range isProd {
				if b {
					c.isProd = true
					break
				}
			}
		}
	}
}

func WithEtag() OptsFn {
	return func(c *mw) {
		c.etag = true
	}
}

func WithMaxAge(d time.Duration) OptsFn {
	return func(c *mw) {
		c.maxAge = d
	}
}

func WithSMaxAge(d time.Duration) OptsFn {
	return func(c *mw) {
		c.sMaxAge = d
	}
}
