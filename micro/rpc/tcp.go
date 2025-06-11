package rpc

import (
	"encoding/binary"
	"net"
)

// 8 bytes = 64 bits
const numOfLengthBytes = 8

func ReadMsg(conn net.Conn) ([]byte, error) {
	// read the head identifies the length
	lenBs := make([]byte, numOfLengthBytes)
	_, err := conn.Read(lenBs)
	if err != nil {
		return nil, err
	}
	headerLength := binary.BigEndian.Uint32(lenBs[:4])
	bodyLength := binary.BigEndian.Uint32(lenBs[4:])
	length := headerLength + bodyLength
	// read message body
	data := make([]byte, length)
	_, err = conn.Read(data[8:])
	copy(data[:8], lenBs)
	return data, nil
}
