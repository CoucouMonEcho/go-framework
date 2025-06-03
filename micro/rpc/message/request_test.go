package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeDecodeRequest(t *testing.T) {
	testCases := []struct {
		name string

		req *Request
	}{
		{
			name: "success",
			req: &Request{
				MessageId:   123,
				Version:     12,
				Compress:    13,
				Serializer:  14,
				ServiceName: "test-service",
				MethodName:  "GetById",
				Meta: map[string]string{
					"trace-id": "123456",
					"a/b":      "a",
				},
				Data: []byte("hello, world"),
			},
		},
		{
			name: "no meta",
			req: &Request{
				MessageId:   123,
				Version:     12,
				Compress:    13,
				Serializer:  14,
				ServiceName: "test-service",
				MethodName:  "GetById",
				//Meta: map[string]string{
				//	"trace-id": "123456",
				//	"a/b":      "a",
				//},
				Data: []byte("hello, world"),
			},
		},
		{
			//FIXME
			name: "no data",
			req: &Request{
				MessageId:   123,
				Version:     12,
				Compress:    13,
				Serializer:  14,
				ServiceName: "test-service",
				MethodName:  "GetById",
				Meta: map[string]string{
					"trace-id": "123456",
					"a/b":      "a",
				},
				//Data: []byte("hello, world"),
			},
		},
		{
			name: "data with \n",
			req: &Request{
				MessageId:   123,
				Version:     12,
				Compress:    13,
				Serializer:  14,
				ServiceName: "test-service",
				MethodName:  "GetById",
				Meta: map[string]string{
					"trace-id": "123456",
					"a/b":      "a",
				},
				Data: []byte("hello, world\n\r"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.calculateHeaderLength()
			tc.req.calculateBodyLength()

			data := EncodeReq(tc.req)
			req := DecodeReq(data)
			assert.Equal(t, tc.req, req)

		})
	}
}

func (req *Request) calculateHeaderLength() {
	req.HeadLength = 15 + uint32(len(req.ServiceName)) + 1 + uint32(len(req.MethodName)) + 1
	for k, v := range req.Meta {
		req.HeadLength += uint32(len(k) + 1 + len(v) + 1)
	}
}

func (req *Request) calculateBodyLength() {
	req.BodyLength = uint32(len(req.Data))
}
