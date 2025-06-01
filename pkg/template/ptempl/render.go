package ptempl

import (
	"context"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func Render[T TemplateDataInterface](c echo.Context, statusCode int, t templ.Component, data *T) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	ctx := context.WithValue(c.Request().Context(), TemplateKey, data)
	if err := t.Render(ctx, buf); err != nil {
		return err
	}
	return c.HTML(statusCode, buf.String())
}

func GetData[T TemplateDataInterface](ctx context.Context) *T {
	if data, ok := ctx.Value(TemplateKey).(*T); ok {
		return data
	}
	return nil
}
