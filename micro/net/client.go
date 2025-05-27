package net

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func Connect(network, addr string) error {
	conn, err := net.DialTimeout(network, addr, time.Second*3)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()
	for i := 0; i < 10; i++ {
		_, err := conn.Write([]byte("hello"))
		if err != nil {
			return err
		}
		res := make([]byte, 128)
		_, err = conn.Read(res)
		if err != nil {
			return err
		}
		fmt.Println(string(res))
	}
	return nil
}

type Client struct {
	network string
	addr    string
}

func NewClient(network, addr string) *Client {
	return &Client{
		network: network,
		addr:    addr,
	}
}

func (c *Client) Send(data string) (string, error) {
	conn, err := net.DialTimeout(c.network, c.addr, time.Second*3)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = conn.Close()
	}()

	// write the head identifies the length
	reqLen := len(data)
	req := make([]byte, reqLen+numOfLengthBytes)
	binary.BigEndian.PutUint64(req[:numOfLengthBytes], uint64(reqLen))
	// write message body
	copy(req[numOfLengthBytes:], data)
	_, err = conn.Write(req)
	if err != nil {
		return "", err
	}

	// read the head identifies the length
	lenBs := make([]byte, numOfLengthBytes)
	_, err = conn.Read(lenBs)
	if err != nil {
		return "", err
	}
	respLen := binary.BigEndian.Uint64(lenBs)
	// read message body
	respBs := make([]byte, respLen)
	_, err = conn.Read(respBs)
	if err != nil {
		return "", err
	}
	return string(respBs), nil

}
