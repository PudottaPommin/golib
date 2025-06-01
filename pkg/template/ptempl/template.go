package ptempl

import (
	"github.com/labstack/echo/v4"
)

type (
	templateKeyType       string
	TemplateDataInterface interface {
		App() *echo.Echo
	}
	TemplateData struct {
		app *echo.Echo
	}
)

const TemplateKey templateKeyType = "ui/template"

func (td TemplateData) App() *echo.Echo {
	return td.app
}

func NewTemplateData(app *echo.Echo) TemplateDataInterface {
	return TemplateData{app: app}
}
