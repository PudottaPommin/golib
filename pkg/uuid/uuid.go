package uuid

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"uuid"
)

var (
	Nil                 = UUID(uuid.Nil())
	ErrConvertTypeError = errors.New("uuid: failed to convert type")
)

type UUID uuid.UUID //nolint:recvcheck // noop

func (u *UUID) Scan(src any) error {
	if src == nil {
		*u = Nil
		return nil
	}

	switch src := src.(type) {
	case []byte:
		if len(src) == len(Nil) {
			copy(u[:], src)
			return nil
		}
		return u.UnmarshalText(src)
	case string:
		return u.UnmarshalText([]byte(src))
	}
	return fmt.Errorf("%w %T to UUID", ErrConvertTypeError, src)
}

func (u UUID) Value() (driver.Value, error) {
	return uuid.UUID(u).String(), nil
}

func (u *UUID) UnmarshalText(b []byte) error {
	zero := uuid.Nil()
	if err := zero.UnmarshalText(b); err != nil {
		return fmt.Errorf("uuid: failed to unmarshal: %w", err)
	}
	*u = UUID(zero)
	return nil
}

func (u UUID) MarshalText() ([]byte, error) {
	v, err := uuid.UUID(u).MarshalText()
	if err != nil {
		return nil, fmt.Errorf("uuid: failed to marshal: %w", err)
	}
	return v, nil
}

func (u UUID) String() string {
	return uuid.UUID(u).String()
}

func Parse(s string) (UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return Nil, err
	}
	return UUID(id), nil
}

func NewV7() UUID {
	return UUID(uuid.NewV7())
}

func NewV4() UUID {
	return UUID(uuid.NewV4())
}

func New() UUID {
	return UUID(uuid.New())
}

func NewV7T[T ~[16]byte]() T {
	return T(NewV7())
}

func NewV4T[T ~[16]byte]() T {
	return T(NewV4())
}

func NewT[T ~[16]byte]() T {
	return T(NewV4())
}
