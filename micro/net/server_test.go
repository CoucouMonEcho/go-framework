package net

import (
	"code-practise/micro/net/mocks"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func Test_handleConn(t *testing.T) {
	tests := []struct {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			err := handleConn(tt.mock(ctrl))
			assert.Equal(t, err, tt.wantErr)
		})
	}
}
