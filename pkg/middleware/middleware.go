package middleware

type (
	Key              string
	OptionsFn[T any] func(*T)
)

const (
	TemplateKey       Key = "template"
	AuthenticationKey Key = "auth"
)
