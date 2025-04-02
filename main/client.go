package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	Ip   string
	Port int
	Name string
	conn net.Conn
	flag int
}

func NewClient(ip string, port int) *Client {
	client := &Client{
		Ip:   ip,
		Port: port,
		flag: 99,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		fmt.Println("net dial err", err)
		return nil
	}
	client.conn = conn
	return client
}

func (this *Client) dealResponse() {
	io.Copy(os.Stdout, this.conn)
}

func (this *Client) menu() bool {
	var flag int

	fmt.Println("1: public message")
	fmt.Println("2: private message")
	fmt.Println("3: rename")
	fmt.Println("0: exit")

	fmt.Scanln(&flag)

	if flag > -1 && flag < 4 {
		this.flag = flag
		return true
	} else {
		fmt.Println("err menu!")
		return false
	}
}

func (this *Client) selectUsers() bool {
	_, err := this.conn.Write([]byte("who\n"))
	if err != nil {
		fmt.Println("selectUsers write err", err)
		return false
	}
	return true
}

func (this *Client) privateMsg() bool {
	var userName string
	var chatMsg string

	this.selectUsers()

	fmt.Print("select one user: ")
	fmt.Scanln(&userName)
	for userName != "exit" {
		fmt.Print("input chatMsg: ")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to@" + userName + " " + chatMsg + "\n"
				_, err := this.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("chatMsg write err", err)
					break
				}
			}
			fmt.Print("input chatMsg: ")
			chatMsg = ""
			fmt.Scanln(&chatMsg)
		}
		fmt.Print("select one user: ")
		userName = ""
		fmt.Scanln(&userName)
	}
	return true
}

func (this *Client) publicMsg() bool {
	fmt.Print("input public msg: ")
	var chatMsg string
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := this.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("chatMsg write err", err)
				break
			}
		}
		chatMsg = ""
		fmt.Scanln(&chatMsg)
	}
	return true
}

func (this *Client) rename() bool {
	fmt.Print("input name: ")
	fmt.Scanln(&this.Name)
	_, err := this.conn.Write([]byte("rename:" + this.Name + "\n"))
	if err != nil {
		fmt.Println("rename write err", err)
		return false
	}
	return true
}

func (this *Client) Run() {
	for this.flag != 0 {
		for this.menu() != true {
		}
		switch this.flag {
		case 1:
			this.publicMsg()
			break
		case 2:
			this.privateMsg()
			break
		case 3:
			this.rename()
			break
		}
	}
}

var ip string
var port int

func init() {
	flag.StringVar(&ip, "i", "127.0.0.1", "ip default:127.0.0.1")
	flag.IntVar(&port, "p", 8888, "port default:8888")
}

func main() {
	flag.Parse()

	client := NewClient(ip, port)
	if client == nil {
		fmt.Println("connect server fail...")
		return
	}

	// listen
	go client.dealResponse()

	fmt.Println("connect server success!")
	client.Run()
}
