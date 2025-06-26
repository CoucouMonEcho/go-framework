package compress

type Compressor interface {
	Code() uint8
	Compress(data []byte) ([]byte, error)
	Uncompress(data []byte) ([]byte, error)
}
