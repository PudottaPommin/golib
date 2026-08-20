package binding

var _ Binder[string] = (*StringBinder)(nil)

// StringGenericBinder binds string inputs to string types (including custom ~string types).
type StringGenericBinder[T ~string] struct{}

func (m StringGenericBinder[T]) Bind(s string, dst *T) error {
	if dst == nil {
		return ErrDestinationNil
	}
	*dst = T(s)
	return nil
}

func (m StringGenericBinder[T]) BindMany(s []string, dst *[]T) error {
	if dst == nil {
		return ErrDestinationNil
	}
	if cap(*dst) < len(s) {
		*dst = make([]T, len(s))
	} else {
		*dst = (*dst)[:len(s)]
	}
	for i, v := range s {
		(*dst)[i] = T(v)
	}
	return nil
}

// BindT is provided for backward compatibility.
func (m StringGenericBinder[T]) BindT(src string, dst *T) error {
	return m.Bind(src, dst)
}

// BindManyT is provided for backward compatibility.
func (m StringGenericBinder[T]) BindManyT(src []string, dst *[]T) error {
	return m.BindMany(src, dst)
}

// StringBinder is an alias for StringGenericBinder[string].
type StringBinder = StringGenericBinder[string]
