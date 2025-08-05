package gzip

import (
	"bytes"
	"errors"
	"github.com/CoucouMonEcho/go-framework/micro/rpc/compress"
	"github.com/klauspost/compress/gzip"
	"io"
)

var _ compress.Compressor = &Compressor{}

type Compressor struct{}

func (_ Compressor) Code() byte {
	return 1
}

func (_ Compressor) Compress(src []byte) ([]byte, error) {
	res := bytes.NewBuffer(nil)
	w := gzip.NewWriter(res)
	if _, err := w.Write(src); err != nil {
		return nil, err
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func (_ Compressor) Uncompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = r.Close()
	}()
	res, err := io.ReadAll(r)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}
	return res, nil
}
