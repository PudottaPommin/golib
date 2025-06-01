package auth

import (
	"github.com/google/uuid"
)

type (
	key      string
	Identity interface {
		ID() uuid.UUID
		Username() string
	}
	identity struct {
		id       uuid.UUID
		username string
	}
)

const identityContextKey key = "auth/identity"

func NewIdentity(cv *CookieValue) (Identity, error) {
	return &identity{id: cv.ID, username: cv.Username}, nil
}

func (i *identity) ID() uuid.UUID {
	return i.id
}

func (i *identity) Username() string {
	return i.username
}
