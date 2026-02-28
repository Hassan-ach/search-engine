package template

import (
	"io"
	"text/template"

	"github.com/labstack/echo/v5"
)

type Template struct {
	template *template.Template
}

func NewTemplate() *Template {
	return &Template{template: template.Must(template.ParseGlob("*.html"))}
}

func (t *Template) Render(w io.Writer, name string, data any, c *echo.Context) error {
	return t.template.ExecuteTemplate(w, name, data)
}
