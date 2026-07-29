package binding

// CustomBinder is a high-performance generic binder for types implementing BindUnmarshaler.
type CustomBinder[T any, PT interface {
	*T
	BindUnmarshaler
}] struct{}

func (m CustomBinder[T, PT]) guard(a any) bool {
	switch a.(type) {
	case *T, *[]T:
		return true
	default:
		return false
	}
}

func (m CustomBinder[T, PT]) Mappable(a any) bool {
	return m.guard(a)
}

func (m CustomBinder[T, PT]) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*T))
}

func (m CustomBinder[T, PT]) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]T))
}

func (m CustomBinder[T, PT]) BindT(src string, dst *T) error {
	if dst == nil {
		return ErrDestinationNil
	}
	return PT(dst).UnmarshalBind(src)
}

func (m CustomBinder[T, PT]) BindManyT(src []string, dst *[]T) error {
	if dst == nil {
		return ErrDestinationNil
	}
	arr := make([]T, len(src))
	for i, v := range src {
		if err := PT(&arr[i]).UnmarshalBind(v); err != nil {
			return err
		}
	}
	*dst = arr
	return nil
}
