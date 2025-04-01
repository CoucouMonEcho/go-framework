package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	addr := conn.RemoteAddr().String()

	user := &User{
		Name:   addr,
		Addr:   addr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// listen
	go user.onMessage()

	return user
}

func (this *User) Online() {

	// online map
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// broad cast
	this.server.broadCast(this, "is online!")

}

func (this *User) Offline() {

	// online map
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// broad cast
	this.server.broadCast(this, "is offline...")

}

func (this *User) SendMessage(msg string) {
	this.conn.Write([]byte(msg))
}

func (this *User) DoMessage(msg string) {

	if msg == "who" {
		for _, user := range this.server.OnlineMap {
			this.SendMessage("[" + user.Addr + "]" + user.Name + ":is online...\n")
		}
	} else {
		// broad cast
		this.server.broadCast(this, msg)
	}

}

func (this *User) onMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
