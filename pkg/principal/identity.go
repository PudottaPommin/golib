package principal

type (
	Identity[T comparable] interface {
		ID() T
		Username() string
		SecurityStamp() []byte
	}
	User[T comparable] struct {
		id            T
		username      string
		securityStamp []byte
	}
)

var _ Identity[string] = (*User[string])(nil)

func NewUser[T comparable](id T, username string, securityStamp []byte) *User[T] {
	return &User[T]{id: id, username: username, securityStamp: securityStamp}
}

func (u *User[T]) ID() T {
	return u.id
}

func (u *User[T]) Username() string {
	return u.username
}

func (u *User[T]) SecurityStamp() []byte {
	return u.securityStamp
}
