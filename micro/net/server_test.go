package net

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go-framework/micro/net/mocks"
	"net"
	"testing"
)

func Test_handleConn(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) net.Conn

		wantErr error
	}{
		{
			name: "read error",
			mock: func(ctrl *gomock.Controller) net.Conn {
				conn := mocks.NewMockConn(ctrl)
				conn.EXPECT().
					Read(gomock.Any()).
					Return(0, errors.New("micro: read too many bytes"))
				return conn
			},
			wantErr: errors.New("micro: read too many bytes"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			err := handleConn(tc.mock(ctrl))
			assert.Equal(t, err, tc.wantErr)
		})
	}
}
