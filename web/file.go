package web

import (
	lru "github.com/hashicorp/golang-lru/v2"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
			ctx.RespData = []byte("illegal path used")
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

type StaticResourceHandler struct {
	dir                     string
	pathPrefix              string
	cache                   *lru.Cache[string, []byte]
	extensionContentTypeMap map[string]string
	maxSize                 int
}

type StaticResourceHandlerOption func(sh *StaticResourceHandler)

func NewStaticResourceHandler(dir string, pathPrefix string, opts ...StaticResourceHandlerOption) (*StaticResourceHandler, error) {
	// key-value count <= 1000
	c, err := lru.New[string, []byte](1000)
	if err != nil {
		return nil, err
	}
	res := &StaticResourceHandler{
		dir:        dir,
		pathPrefix: pathPrefix,
		cache:      c,
		extensionContentTypeMap: map[string]string{
			"jpeg": "image/jpeg",
			"jpe":  "image/jpeg",
			"jpg":  "image/jpeg",
			"png":  "image/png",
			"pdf":  "image/pdf",
		},
		maxSize: 1024 * 1024 * 10,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func StaticResourceHandlerWithMaxSize(maxSize int) StaticResourceHandlerOption {
	return func(sh *StaticResourceHandler) {
		sh.maxSize = maxSize
	}
}

func StaticResourceHandlerWithCache(c *lru.Cache[string, []byte]) StaticResourceHandlerOption {
	return func(sh *StaticResourceHandler) {
		sh.cache = c
	}
}

func StaticResourceHandlerWithMoreExtension(extMap map[string]string) StaticResourceHandlerOption {
	return func(sh *StaticResourceHandler) {
		for ext, contentType := range extMap {
			sh.extensionContentTypeMap[ext] = contentType
		}
	}
}

func (sh *StaticResourceHandler) Handle(ctx *Context) {
	req, err := ctx.PathValue("file").String()
	if err != nil {
		ctx.RespCode = http.StatusBadRequest
		ctx.RespData = []byte(err.Error())
		return
	}
	path := filepath.Join(sh.dir, sh.pathPrefix, filepath.Clean(req))
	path, err = filepath.Abs(path)
	if !strings.Contains(path, sh.dir) {
		ctx.RespCode = http.StatusBadRequest
		ctx.RespData = []byte("illegal path used")
		return
	}
	var data []byte
	var ok bool
	if data, ok = sh.cache.Get(path); !ok {
		data, err = os.ReadFile(path)
		if err != nil {
			ctx.RespCode = http.StatusInternalServerError
			ctx.RespData = []byte(err.Error())
			return
		}
		if len(data) <= sh.maxSize {
			sh.cache.Add(path, data)
		}
	}
	header := ctx.Resp.Header()
	header.Set("Content-Type", sh.extensionContentTypeMap[filepath.Ext(path)[1:]])
	header.Set("Content-Length", strconv.Itoa(len(data)))
	ctx.RespCode = http.StatusOK
	ctx.RespData = data
	return
}
