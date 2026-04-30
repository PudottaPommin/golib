package binding

var _ GenericBinder[string] = (*StringBinder)(nil)

type StringBinder struct{}

func (m StringBinder) guard(a any) bool {
	switch a.(type) {
	case *string, *[]string:
		return true
	default:
		return false
	}
}

func (m StringBinder) Mappable(a any) bool {
	switch a.(type) {
	case *string, *[]string:
		return true
	default:
		return false
	}
}

func (m StringBinder) Bind(src string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindT(src, dst.(*string))
}

func (m StringBinder) BindMany(src []string, dst any) error {
	if !m.guard(dst) {
		return ErrorDestinationTypeInvalid
	}
	return m.BindManyT(src, dst.(*[]string))
}

func (m StringBinder) BindT(src string, dst *string) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	*dst = src
	return nil
}

func (m StringBinder) BindManyT(src []string, dst *[]string) error {
	if dst == nil {
		return ErrorDestinationNil
	}
	*dst = src
	return nil
}
