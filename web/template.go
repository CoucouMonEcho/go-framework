package web

import (
	"bytes"
	"context"
	"html/template"
)

type TemplateEngine interface {
	Render(ctx context.Context, name string, data any) ([]byte, error)

	// middleware could not work
	// Render(ctx context.Context, name string, data any, writer io.Writer) error

	// use web.Context maybe coupling
	// Render(ctx Context)

	// AddTemplate() should not provide by web engine
}

var _ TemplateEngine = &GoTemplateEngine{}

type GoTemplateEngine struct {
	T *template.Template
}

func (gt *GoTemplateEngine) Render(_ context.Context, name string, data any) ([]byte, error) {
	buffer := &bytes.Buffer{}
	err := gt.T.ExecuteTemplate(buffer, name, data)
	return buffer.Bytes(), err
}
