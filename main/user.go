package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

func NewUser(conn net.Conn) *User {
	addr := conn.RemoteAddr().String()

	user := &User{
		Name: addr,
		Addr: addr,
		C:    make(chan string),
		conn: conn,
	}

	// listen
	go user.onMessage()

	return user
}

func (this *User) onMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
