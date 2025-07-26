//go:build e2e

package web

import (
	"github.com/stretchr/testify/require"
	"html/template"
	"log"
	"mime/multipart"
	"path/filepath"
	"testing"
)

func TestFileUploader_Handle(t *testing.T) {
	tpl, err := template.ParseGlob("testdata/tpls/*.gohtml")
	require.NoError(t, err)

	engine := &GoTemplateEngine{
		T: tpl,
	}
	h := NewHTTPServer(ServerWithTemplateEngine(engine))
	h.Get("/upload", func(ctx *Context) {
		err := ctx.Render("upload.gohtml", nil)
		if err != nil {
			log.Println(err)
		}
	})
	fu := &FileUploader{
		// <input type="file" name="upload_file">
		FileField: "upload_file",
		DstPathFunc: func(header *multipart.FileHeader) string {
			return filepath.Join("testdata", "upload", header.Filename)
		},
	}
	h.Post("/upload", fu.Handle())
	h.Start(":8081")
}

func TestFileDownloader_Handle(t *testing.T) {
	h := NewHTTPServer()
	fd := &FileDownloader{
		// Dir: "testdata/download" it may lead to compatibility issues
		Dir: filepath.Join("testdata", "download"),
	}
	h.Get("/download", fd.Handle())
	h.Start(":8081")
}

func TestStaticResourceHandler_Handle(t *testing.T) {
	h := NewHTTPServer()
	s, err := NewStaticResourceHandler(filepath.Join("testdata", "static"), "js", StaticResourceHandlerWithMaxSize(), StaticResourceHandlerWithCache(), StaticResourceHandlerWithMoreExtension())
	require.NoError(t, err)
	// localhost:8081/static/xxx.jpg
	h.Get("/static/:file", s.Handle)
	h.Start(":8081")
}
