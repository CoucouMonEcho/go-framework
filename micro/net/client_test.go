package net

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	go func() {
		err := Serve("tcp", ":8082")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	err := Connect("tcp", ":8082")
	t.Log(err)
}

func TestClient_Send(t *testing.T) {
	s := &Server{}
	go func() {
		err := s.Start("tcp", ":8082")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	c := NewClient("tcp", ":8082")
	resp, err := c.Send("hello world")
	require.NoError(t, err)
	require.Equal(t, "hello worldhello world", resp)
}
