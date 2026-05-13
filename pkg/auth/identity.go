package auth

import (
	"github.com/gofrs/uuid/v5"
)

type (
	Identity interface {
		ID() uuid.UUID
		Username() string
	}
	identity struct {
		id       uuid.UUID
		username string
	}
)

var _ Identity = (*identity)(nil)

func NewIdentity(cv *CookieValue) (Identity, error) {
	return &identity{id: cv.ID, username: cv.Username}, nil
}

func (i *identity) ID() uuid.UUID {
	return i.id
}

func (i *identity) Username() string {
	return i.username
}
