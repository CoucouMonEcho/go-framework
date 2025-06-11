package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeDecodeResponse(t *testing.T) {
	testCases := []struct {
		name string

		resp *Response
	}{
		{
			name: "success",
			resp: &Response{
				MessageId:  123,
				Version:    12,
				Compress:   13,
				Serializer: 14,
				Error:      []byte("error"),
				Data:       []byte("hello, world"),
			},
		},
		{
			name: "no error",
			resp: &Response{
				MessageId:  123,
				Version:    12,
				Compress:   13,
				Serializer: 14,
				Data:       []byte("hello, world"),
			},
		},
		{
			name: "no data",
			resp: &Response{
				MessageId:  123,
				Version:    12,
				Compress:   13,
				Serializer: 14,
				Error:      []byte("error"),
			},
		},
		{
			name: "data with \n",
			resp: &Response{
				MessageId:  123,
				Version:    12,
				Compress:   13,
				Serializer: 14,
				Data:       []byte("hello, world\n\r"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.resp.CalculateHeaderLength()
			tc.resp.CalculateBodyLength()

			data := EncodeResp(tc.resp)
			resp := DecodeResp(data)
			assert.Equal(t, tc.resp, resp)

		})
	}
}
