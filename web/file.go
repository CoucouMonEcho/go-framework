package web

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// normally use OSS
type FileUploader struct {
	FileField   string
	DstPathFunc func(file *multipart.FileHeader) string
}

func (fu *FileUploader) Handle() HandlerFunc {
	// return handler
	// do some check before handle
	if fu.FileField == "" {
		fu.FileField = "file"
	}
	if fu.DstPathFunc == nil {
		fu.DstPathFunc = func(header *multipart.FileHeader) string {
			return filepath.Join("file", "upload", header.Filename)
		}
	}
	return func(ctx *Context) {
		file, header, err := ctx.Req.FormFile(fu.FileField)
		defer file.Close()
		if err != nil {
			ctx.RespCode = http.StatusInternalServerError
			ctx.RespData = []byte(err.Error())
			return
		}
		// path decision left to user
		dst := fu.DstPathFunc(header)
		err = os.MkdirAll(dst, 0o666)
		if err != nil {
			ctx.RespCode = http.StatusInternalServerError
			ctx.RespData = []byte(err.Error())
			return
		}
		// O_WRONLY write
		// O_TRUNC delete if exists
		// O_CREATE create
		dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o666)
		defer dstFile.Close()
		if err != nil {
			ctx.RespCode = http.StatusInternalServerError
			ctx.RespData = []byte(err.Error())
			return
		}
		// buffer default 32 * 1024
		// extension: reusable buffer
		_, err = io.CopyBuffer(dstFile, file, nil)
		if err != nil {
			ctx.RespCode = http.StatusInternalServerError
			ctx.RespData = []byte(err.Error())
			return
		}
		ctx.RespCode = http.StatusOK
	}
}

// option design for code practice

//type FileUploaderOption func(*FileUploader) *FileUploader
//
//func NewFileUploader(opts ...FileUploaderOption) *FileUploader {
//	fu := &FileUploader{
//		FileField: "file",
//		DstPathFunc: func(header *multipart.FileHeader) string {
//			return filepath.Join("file", "upload", header.Filename)
//		},
//	}
//	for _, opt := range opts {
//		fu = opt(fu)
//	}
//	return fu
//}
//
//func (fu FileUploader) HandleFunc(ctx *Context) {
//	// file upload
//}

// normally use OSS
type FileDownloader struct {
	Dir string
}

func (fd *FileDownloader) Handle() HandlerFunc {
	return func(ctx *Context) {
		req, err := ctx.QueryValue("file").String()
		if err != nil {
			ctx.RespCode = http.StatusBadRequest
			ctx.RespData = []byte(err.Error())
			return
		}
		path := filepath.Join(fd.Dir, filepath.Clean(req))
		path, err = filepath.Abs(path)
		if !strings.Contains(path, fd.Dir) {
			ctx.RespCode = http.StatusBadRequest
			ctx.RespData = []byte(err.Error())
			return
		}
		fn := filepath.Base(path)

		header := ctx.Resp.Header()
		// this header means download the file as an attachment
		header.Set("Content-Disposition", "attachment; filename="+fn)
		// this header means nothing, just for desc
		header.Set("Content-Description", "File Transfer")
		// this header means file type, octet-stream means arbitrary binary stream,
		// the content type cannot be identified,
		// so it is recommended to download rather than open directly
		header.Set("Content-Type", "application/octet-stream")
		// this header means nothing, it is used by MIME mail protocols,
		// it is not a standard field in HTTP, just for compatibility enhancement
		header.Set("Content-Transfer-Encoding", "binary")
		// this header means do not use browser cache
		header.Set("Expires", "0")
		// this header means even browser has cached, also need to confirm again before file used
		header.Set("Cache-Control", "must-revalidate")
		// this header means cache, it is for HTTP/1.0,
		// which "public" is use cache, "no-cache" is not use
		header.Set("Pragma", "no-cache")

		// no cache
		http.ServeFile(ctx.Resp, ctx.Req, path)
	}
}
