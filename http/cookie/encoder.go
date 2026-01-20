package cookie

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
)

type Encoder interface {
	Encode(src any) ([]byte, error)
	Decode(src []byte, dst any) error
}

type GobEncoder struct{}

func (e GobEncoder) Encode(src any) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(src); err != nil {
		return nil, fmt.Errorf("GobEncoder failed to encode: %w", err)
	}
	return buf.Bytes(), nil
}

func (e GobEncoder) Decode(src []byte, dst any) error {
	if err := gob.NewDecoder(bytes.NewReader(src)).Decode(dst); err != nil {
		return fmt.Errorf("GobEncoder failed to decode: %w", err)
	}
	return nil
}

type JSONEncoder struct{}

func (e JSONEncoder) Encode(src any) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(src); err != nil {
		return nil, fmt.Errorf("JSONEncoder failed to encode: %w", err)
	}
	return buf.Bytes(), nil
}

func (e JSONEncoder) Decode(src []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(src))
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("JSONEncoder failed to decode: %w", err)
	}
	return nil
}

type NoopEncoder struct{}

func (e NoopEncoder) Encode(src any) ([]byte, error) {
	if b, ok := src.([]byte); ok {
		return b, nil
	}
	return nil, errors.New("NoopEncoder can only encode []byte")
}

func (e NoopEncoder) Decode(src []byte, dst any) error {
	if b, ok := dst.(*[]byte); ok {
		*b = src
		return nil
	}
	return errors.New("NoopEncoder can only decode into *[]byte")
}

type Base64Encoder struct{}

func (b Base64Encoder) Encode(src any) ([]byte, error) {
	if src, ok := src.([]byte); ok {
		buf := make([]byte, base64.URLEncoding.EncodedLen(len(src)))
		base64.URLEncoding.Encode(buf, src)
		return buf, nil
	}
	return nil, errors.New("Base64Encoder can only encode []byte")
}

func (b Base64Encoder) Decode(src []byte, dst any) error {
	if dst, ok := dst.(*[]byte); ok {
		buf := make([]byte, base64.URLEncoding.DecodedLen(len(src)))
		n, err := base64.URLEncoding.Decode(buf, src)
		if err != nil {
			return fmt.Errorf("Base64Encoder failed to decode: %w", err)
		}
		*dst = buf[:n]
		return nil
	}
	return errors.New("Base64Encoder can only decode into *[]byte")
}
