package message

import (
	"bytes"
	"encoding/binary"
)

const (
	splitter     = '\n'
	pairSplitter = '\r'
)

type Request struct {
	// header
	HeadLength uint32
	BodyLength uint32
	MessageId  uint32
	Version    uint8
	Compress   uint8
	Serializer uint8

	ServiceName string
	MethodName  string

	Meta map[string]string

	// Data use any can not confirm type
	Data []byte
}

func EncodeReq(req *Request) []byte {
	data := make([]byte, req.HeadLength+req.BodyLength)
	cur := data

	// head length
	binary.BigEndian.PutUint32(cur[:4], req.HeadLength)
	cur = cur[4:]

	// body length
	binary.BigEndian.PutUint32(cur[:4], req.BodyLength)
	cur = cur[4:]

	// message id
	binary.BigEndian.PutUint32(cur[:4], req.MessageId)
	cur = cur[4:]

	// version
	cur[0] = req.Version
	cur = cur[1:]

	// compress
	cur[0] = req.Compress
	cur = cur[1:]

	// serializer
	cur[0] = req.Serializer
	cur = cur[1:]

	// service name
	lenServiceName := len(req.ServiceName)
	copy(cur[:lenServiceName], req.ServiceName)
	cur[lenServiceName] = splitter
	cur = cur[lenServiceName+1:]

	// method name
	lenMethodName := len(req.MethodName)
	copy(cur[:lenMethodName], req.MethodName)
	cur[lenMethodName] = splitter
	cur = cur[lenMethodName+1:]

	// meta
	for k, v := range req.Meta {
		copy(cur, k)
		cur[len(k)] = pairSplitter
		cur = cur[len(k)+1:]

		copy(cur, v)
		cur[len(v)] = splitter
		cur = cur[len(v)+1:]
	}

	// data
	copy(cur, req.Data)

	return data
}

func DecodeReq(data []byte) *Request {
	req := &Request{}

	// head length
	req.HeadLength = binary.BigEndian.Uint32(data[:4])
	head := data[4:req.HeadLength]

	// body length
	req.BodyLength = binary.BigEndian.Uint32(head[:4])
	head = head[4:]

	// message id
	req.MessageId = binary.BigEndian.Uint32(head[:4])
	head = head[4:]

	// version
	req.Version = head[0]
	head = head[1:]

	// compress
	req.Compress = head[0]
	head = head[1:]

	// serializer
	req.Serializer = head[0]
	head = head[1:]

	// service name
	index := bytes.IndexByte(head, splitter)
	req.ServiceName = string(head[:index])
	head = head[index+1:]

	// method name
	index = bytes.IndexByte(head, splitter)
	req.MethodName = string(head[:index])
	head = head[index+1:]

	// meta
	index = bytes.IndexByte(head, splitter)
	if index != -1 {
		req.Meta = make(map[string]string, 4)
		for index != -1 {
			pair := head[:index]
			pairIndex := bytes.IndexByte(pair, pairSplitter)
			k := string(pair[:pairIndex])
			v := string(pair[pairIndex+1:])
			req.Meta[k] = v

			head = head[index+1:]
			index = bytes.IndexByte(head, splitter)
		}
	}

	// data
	if req.BodyLength > 0 {
		req.Data = data[req.HeadLength:]
	}

	return req
}
