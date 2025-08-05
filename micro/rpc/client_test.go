package rpc

import (
	"context"
	"errors"
	"github.com/CoucouMonEcho/go-framework/micro/rpc/compress/donothing"
	"github.com/CoucouMonEcho/go-framework/micro/rpc/message"
	"github.com/CoucouMonEcho/go-framework/micro/rpc/serialize/json"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_setFuncField(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) Proxy

		service Service

		wantErr error
		want    any
	}{
		{
			name:    "success",
			service: &TestService{},
			mock: func(ctrl *gomock.Controller) Proxy {
				p := NewMockProxy(ctrl)
				p.EXPECT().
					Invoke(gomock.Any(), &message.Request{
						ServiceName: "test-service",
						MethodName:  "GetById",
						Data:        []byte(`{"Id":123}`),
					}).
					Return(&message.Response{
						Data: []byte(`{"Msg":"hello, world"}`),
					}, nil)
				return p
			},
			want: &GetByIdResp{Msg: "hello, world"},
		},
		{
			name:    "nil",
			service: nil,
			wantErr: errors.New("rpc: service is nil"),
		},
		{
			name:    "not ptr",
			service: TestService{},
			wantErr: errors.New("rpc: service must be a pointer to struct"),
		},
	}
	s := &json.Serializer{}
	c := &donothing.Compressor{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := tc.mock(ctrl)
			err := setFuncField(tc.service, mock, s, c)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			resp, err := tc.service.(*TestService).GetById(context.Background(), &GetByIdReq{Id: 123})
			require.NoError(t, err)
			assert.Equal(t, tc.want, resp)
		})
	}
}
