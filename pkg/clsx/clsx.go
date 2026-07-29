// Package clsx is heavily inspired bz https://github.com/lukeed/clsx
package clsx

import (
	"github.com/valyala/bytebufferpool"
)

type (
	HandlerFn func() bool
	Map       map[string]Value
	valueKind uint8
	Value     struct {
		h    HandlerFn
		cls  string
		kind valueKind
		b    bool
	}
)

const (
	kindBoolean valueKind = iota
	kindFn
)

func Make(kv ...Value) Map {
	m := make(Map, len(kv))
	for _, v := range kv {
		switch v.kind {
		case kindBoolean, kindFn:
			m[v.cls] = v
		}
	}
	return m
}

func (m Map) String() string {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	wasWritten := false
	for k, v := range m {
		switch v.kind {
		case kindFn:
			if !v.h() {
				continue
			}
		case kindBoolean:
			if !v.b {
				continue
			}
		default:
			continue
		}
		if wasWritten {
			_ = buf.WriteByte(' ')
		} else {
			wasWritten = true
		}
		_, _ = buf.WriteString(k)
	}

	return buf.String()
}

func String(cls string) Value {
	return Boolean(cls, true)
}

func Boolean(cls string, cond bool) Value {
	return Value{kind: kindBoolean, cls: cls, b: cond}
}

func Handler(cls string, fn HandlerFn) Value {
	return Value{kind: kindFn, cls: cls, h: fn}
}
