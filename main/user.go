package main

import (
	"net"
	"strings"
)

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
	} else if len(msg) > 7 && msg[:7] == "rename:" {
		newName := strings.Split(msg, ":")[1]
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMessage("this name is already used!")
		}
		this.server.mapLock.Lock()

		delete(this.server.OnlineMap, this.Name)
		this.server.OnlineMap[newName] = this

		this.server.mapLock.Unlock()

		this.Name = newName
		this.SendMessage("rename -> " + this.Name + " success!\n")
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
