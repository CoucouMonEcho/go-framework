package message

type Response struct {
	// header
	HeadLength uint32
	BodyLength uint32
	MessageId  uint32
	Version    uint8
	Compress   uint8
	Serializer uint8

	Error []byte

	Data []byte
}
