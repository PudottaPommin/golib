package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/pudottapommin/golib/pkg/template/ptempl"
)

type (
	TemplateOptFn[T ptempl.TemplateDataInterface]     OptionsFn[TemplateConfig[T]]
	TemplateFactoryFn[T ptempl.TemplateDataInterface] func(ctx echo.Context) *T
	TemplateConfig[T ptempl.TemplateDataInterface]    struct {
		Key         string
		FactoryFunc TemplateFactoryFn[T]
	}
)

func Template[T ptempl.TemplateDataInterface](opts ...TemplateOptFn[T]) echo.MiddlewareFunc {
	cfg := TemplateConfig[T]{
		Key: string(TemplateKey),
		FactoryFunc: func(c echo.Context) *T {
			if td, ok := ptempl.NewTemplateData(c.Echo()).(T); ok {
				return &td
			}
			return new(T)
		},
	}
	cfg.applyOptions(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			data := cfg.FactoryFunc(c)
			c.Set(cfg.Key, data)
			return next(c)
		}
	}
}

func (c *TemplateConfig[T]) applyOptions(opts ...TemplateOptFn[T]) *TemplateConfig[T] {
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithTemplateContextKey[T ptempl.TemplateDataInterface](key string) TemplateOptFn[T] {
	return func(tc *TemplateConfig[T]) {
		tc.Key = key
	}
}

func WithTemplateFactoryFunc[T ptempl.TemplateDataInterface](factory func(echo.Context) *T) TemplateOptFn[T] {
	return func(tc *TemplateConfig[T]) {
		tc.FactoryFunc = factory
	}
}
