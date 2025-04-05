package simple

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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
		for _, user := range this.OnlineMap {
			user.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) broadCast(user *User, msg string) {
	this.Message <- "[" + user.Addr + "]" + user.Name + ":" + msg
}

func (this *Server) handler(conn net.Conn) {

	user := NewUser(conn, this)

	// online
	user.Online()

	// isLive
	isLive := make(chan bool)

	// receive msg
	go func() {
		for {
			buf := make([]byte, 4096)
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn read err:", err)
				return
			}
			user.DoMessage(string(buf[:n-1]))
			isLive <- true
		}
	}()

	// clog
	for {
		select {
		case <-isLive:
		case <-time.After(time.Second * 300):
			user.SendMessage("time out offline...")
			close(user.C)
			conn.Close()
			return
		}
	}
}

func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// close listen socket
	defer listener.Close()

	// listen socket
	go this.listenMessage()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// do handler
		go this.handler(conn)
	}

}
