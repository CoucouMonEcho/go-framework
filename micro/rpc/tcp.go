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
	dataLen := binary.BigEndian.Uint64(lenBs)
	// read message body
	data := make([]byte, dataLen)
	_, err = conn.Read(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func EncodeMsg(data []byte) []byte {
	// write the head identifies the length
	reqLen := len(data)
	res := make([]byte, reqLen+numOfLengthBytes)
	binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(reqLen))
	copy(res[numOfLengthBytes:], data)
	return res
}
