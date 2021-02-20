package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/luiz-couto/File-Transfer-UDP-TCP/pkg/bytes"
	"github.com/luiz-couto/File-Transfer-UDP-TCP/pkg/message"
)

//Client DOC TODO
type Client struct {
	connTCP net.Conn
	connUDP *UDPConnection
}

//UDPConnection defines the udp connection object
type UDPConnection struct {
	UDP  *net.UDPConn
	port int
}

func (c *Client) startUDPConnection() {
	// Get a Free Port Number
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println(err)
		return
	}

	udpAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(listener.Addr().(*net.TCPAddr).Port))
	if err != nil {
		fmt.Println(err)
		return
	}

	udp, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	connUDP := &UDPConnection{
		UDP:  udp,
		port: udpAddr.Port,
	}

	c.connUDP = connUDP

}

func (c *Client) handleConnection() {
	for {
		msg, err := bufio.NewReader(c.connTCP).ReadBytes('\n')
		msg = msg[:len(msg)-1]
		if err != nil {
			fmt.Println(err)
			return
		}
		c.handleMsg(msg)
	}
}

func (c *Client) handleMsg(msg []byte) {
	msgID := bytes.ReadByteBlockAsInt(0, 2, msg)
	switch msgID {
	case message.HelloType:
		fmt.Println("Received HELLO")
		c.startUDPConnection()
		message.NewMessage().CONNECTION(c.connUDP.port).Send(c.connTCP)
	}
}

func main() {
	args := os.Args
	if len(args) == 1 {
		fmt.Println("usage ./server <port_number>")
		return
	}

	PORT := ":" + args[1]
	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		client := &Client{
			connTCP: conn,
		}
		go client.handleConnection()
	}

}
