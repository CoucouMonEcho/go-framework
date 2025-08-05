package donothing

import "github.com/CoucouMonEcho/go-framework/micro/rpc/compress"

var _ compress.Compressor = &Compressor{}

type Compressor struct{}

func (_ Compressor) Code() byte {
	return 0
}

func (_ Compressor) Compress(src []byte) ([]byte, error) {
	return src, nil
}

func (_ Compressor) Uncompress(src []byte) ([]byte, error) {
	return src, nil
}
