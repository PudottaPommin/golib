package binding

type dummyUnmarshaler struct{}

func (d *dummyUnmarshaler) UnmarshalBind(string) error { return nil }

var _ Binder[dummyUnmarshaler] = (*CustomBinder[dummyUnmarshaler, *dummyUnmarshaler])(nil)

// CustomBinder is a high-performance generic binder for types implementing BindUnmarshaler.
type CustomBinder[T any, PT interface {
	*T
	BindUnmarshaler
}] struct{}

func (m CustomBinder[T, PT]) Bind(src string, dst *T) error {
	if dst == nil {
		return ErrDestinationNil
	}
	return PT(dst).UnmarshalBind(src)
}

func (m CustomBinder[T, PT]) BindMany(src []string, dst *[]T) error {
	if dst == nil {
		return ErrDestinationNil
	}
	if cap(*dst) < len(src) {
		*dst = make([]T, len(src))
	} else {
		*dst = (*dst)[:len(src)]
	}
	for i, v := range src {
		if err := PT(&(*dst)[i]).UnmarshalBind(v); err != nil {
			return err
		}
	}
	return nil
}

// BindT is provided for backward compatibility.
func (m CustomBinder[T, PT]) BindT(src string, dst *T) error {
	return m.Bind(src, dst)
}

// BindManyT is provided for backward compatibility.
func (m CustomBinder[T, PT]) BindManyT(src []string, dst *[]T) error {
	return m.BindMany(src, dst)
}
