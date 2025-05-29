package rpc

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
					Invoke(gomock.Any(), &Request{
						ServiceName: "test-service",
						MethodName:  "GetById",
						Arg: &GetByIdReq{
							Id: 123,
						},
					}).
					Return(&Response{}, nil)
				return p
			},
			want: (*GetByIdResp)(nil),
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
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			err := setFuncField(tc.service, tc.mock(ctrl))
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			resp, err := tc.service.(*TestService).GetById(context.Background(), &GetByIdReq{Id: 123})
			assert.Equal(t, tc.want, resp)
		})
	}
}

type TestService struct {
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
}

func (t TestService) Name() string {
	return "test-service"
}

type GetByIdReq struct {
	Id int64
}

type GetByIdResp struct {
}
