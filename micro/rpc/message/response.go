package message

import (
	"encoding/binary"
)

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

func EncodeResp(resp *Response) []byte {
	data := make([]byte, resp.HeadLength+resp.BodyLength)
	cur := data

	// head length
	binary.BigEndian.PutUint32(cur[:4], resp.HeadLength)
	cur = cur[4:]

	// body length
	binary.BigEndian.PutUint32(cur[:4], resp.BodyLength)
	cur = cur[4:]

	// message id
	binary.BigEndian.PutUint32(cur[:4], resp.MessageId)
	cur = cur[4:]

	// version
	cur[0] = resp.Version
	cur = cur[1:]

	// compress
	cur[0] = resp.Compress
	cur = cur[1:]

	// serializer
	cur[0] = resp.Serializer
	cur = cur[1:]

	// error
	copy(cur, resp.Error)
	cur = cur[len(resp.Error):]

	// data
	copy(cur, resp.Data)

	return data
}

func DecodeResp(data []byte) *Response {
	resp := &Response{}

	// head length
	resp.HeadLength = binary.BigEndian.Uint32(data[:4])
	head := data[4:resp.HeadLength]

	// body length
	resp.BodyLength = binary.BigEndian.Uint32(head[:4])
	head = head[4:]

	// message id
	resp.MessageId = binary.BigEndian.Uint32(head[:4])
	head = head[4:]

	// version
	resp.Version = head[0]
	head = head[1:]

	// compress
	resp.Compress = head[0]
	head = head[1:]

	// serializer
	resp.Serializer = head[0]
	head = head[1:]

	// error
	if resp.HeadLength > 15 {
		resp.Error = head
	}

	// data
	if resp.BodyLength > 0 {
		resp.Data = data[resp.HeadLength:]
	}

	return resp
}

func (resp *Response) CalculateHeaderLength() {
	resp.HeadLength = 15 + uint32(len(resp.Error))
}

func (resp *Response) CalculateBodyLength() {
	resp.BodyLength = uint32(len(resp.Data))
}
