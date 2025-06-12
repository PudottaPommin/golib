package ptempl

import (
	"context"

	"github.com/a-h/templ"
)

func Render[T TemplateDataInterface](c context.Context, t templ.Component, data *T) (string, error) {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	ctx := context.WithValue(c, TemplateKey, data)
	if err := t.Render(ctx, buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GetData[T TemplateDataInterface](ctx context.Context) *T {
	if data, ok := ctx.Value(TemplateKey).(*T); ok {
		return data
	}
	return nil
}
