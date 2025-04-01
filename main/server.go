package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	// online map
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// msg chan
	Message chan string
}

func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

func (this *Server) listenMessage() {
	for {
		msg := <-this.Message

		this.mapLock.Lock()

		// send all
		for _, client := range this.OnlineMap {
			client.C <- msg
		}

		this.mapLock.Unlock()
	}
}

func (this *Server) broadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) handler(conn net.Conn) {
	//fmt.Println("connect success!")

	// lock
	this.mapLock.Lock()

	// online map
	user := NewUser(conn)
	this.OnlineMap[user.Name] = user

	// unlock
	this.mapLock.Unlock()

	// broad cast
	this.broadCast(user, "is online!")

	// clog
	select {}
}

func (this *Server) Start() {

	// listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// listen
	go this.listenMessage()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}

		// do handler
		go this.handler(conn)
	}

	// close listen socket
	listener.Close()
}
