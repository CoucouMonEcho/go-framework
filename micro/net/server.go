package net

import (
	"encoding/binary"
	"net"
)

// 8 bytes = 64 bits
const numOfLengthBytes = 8

// mockgen -destination=micro/net/mocks/net_conn.gen.go -package=mocks net Conn

func Serve(network, addr string) error {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if er := handleConn(conn); er != nil {
				_ = conn.Close()
			}
		}()
	}
}

func handleConn(conn net.Conn) error {
	for {
		bs := make([]byte, numOfLengthBytes)
		_, err := conn.Read(bs)
		if err != nil {
			return err
		}
		//if n != numOfLengthBytes {
		//	return errors.New("micro: read too many bytes")
		//}

		res := handleMsg(bs)
		_, err = conn.Write(res)
		if err != nil {
			return err
		}
		//if n != len(res) {
		//	return errors.New("micro: write too few bytes")
		//}
	}
}

//func handleConn(conn net.Conn) {
//	for {
//		bs := make([]byte, numOfLengthBytes)
//		_, err := conn.Read(bs)
//		if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || errors.Is(err, io.ErrUnexpectedEOF) {
//			_ = conn.Close()
//			return
//		}
//		if err != nil {
//			continue
//		}
//
//		res := handleMsg(bs)
//		_, err = conn.Write(res)
//		if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || errors.Is(err, io.ErrUnexpectedEOF) {
//			_ = conn.Close()
//			return
//		}
//	}
//}

func handleMsg(req []byte) []byte {
	res := make([]byte, 2*len(req))
	copy(res[:len(req)], req)
	copy(res[len(req):], req)
	return res
}

type Server struct {
	//network string
	//addr    string
}

//func NewServer(network, addr string) *Server {
//	return &Server{
//		network: network,
//		addr:    addr,
//	}
//}

func (s *Server) Start(network, addr string) error {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if er := s.handleConn(conn); er != nil {
				_ = conn.Close()
			}
		}()
	}
}

func (s *Server) handleConn(conn net.Conn) error {
	for {
		// read the head identifies the length
		lenBs := make([]byte, numOfLengthBytes)
		_, err := conn.Read(lenBs)
		if err != nil {
			return err
		}
		reqLen := binary.BigEndian.Uint64(lenBs)
		// read message body
		reqBs := make([]byte, reqLen)
		_, err = conn.Read(reqBs)
		if err != nil {
			return err
		}

		// handle
		respData := handleMsg(reqBs)

		// write the head identifies the length
		respLen := len(respData)
		res := make([]byte, respLen+numOfLengthBytes)
		binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(respLen))
		// write message body
		copy(res[numOfLengthBytes:], respData)
		// return
		_, err = conn.Write(res)
		if err != nil {
			return err
		}
	}
}
