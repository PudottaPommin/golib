package ptempl

type (
	templateKeyType       string
	TemplateDataInterface interface {
		// App() *echo.Echo
	}
	TemplateData struct {
		// app *echo.Echo
	}
)

const TemplateKey templateKeyType = "ui/template"

func NewTemplateData() TemplateDataInterface {
	return TemplateData{}
}
