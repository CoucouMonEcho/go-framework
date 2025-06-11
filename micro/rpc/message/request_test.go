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
			tc.req.CalculateHeaderLength()
			tc.req.CalculateBodyLength()

			data := EncodeReq(tc.req)
			req := DecodeReq(data)
			assert.Equal(t, tc.req, req)

		})
	}
}
