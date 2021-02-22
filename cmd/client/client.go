package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/luiz-couto/File-Transfer-UDP-TCP/pkg/bytes"
	"github.com/luiz-couto/File-Transfer-UDP-TCP/pkg/message"
)

// File DOC TODO
type File struct {
	fileSize int
	fileName string
	content  []byte
}

// SlidingWindow DOC TODO
type SlidingWindow struct {
	windowSize int
	nextWindow []int
}

// Client DOC TODO
type Client struct {
	UDPconn   *net.UDPConn
	TCPconn   net.Conn
	SliWindow *SlidingWindow
	file      *File
}

// ReadFile DOC TODO
func ReadFile(fileName string) *File {
	fileContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
	}

	return &File{
		fileSize: len(fileContent),
		fileName: fileName,
		content:  fileContent,
	}
}

func (c *Client) startUDPConnection(port int) {
	connectTo := strings.Split(c.TCPconn.RemoteAddr().String(), ":")[0]

	addr, err := net.ResolveUDPAddr("udp", connectTo+":"+strconv.Itoa(port))
	udpConn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.UDPconn = udpConn
}

func (c *Client) handleMsg(msg []byte) {
	msgID := bytes.ReadByteBlockAsInt(0, 2, msg)
	switch msgID {
	case message.ConnectionType:
		fmt.Println("Received CONNECTION")
		port := bytes.ReadByteBlockAsInt(2, 6, msg)

		fmt.Println("Porto UDP is " + strconv.Itoa(port))
		c.startUDPConnection(port)

		message.NewMessage().INFOFILE(c.file.fileName, c.file.fileSize).Send(c.TCPconn)
	}
}

func main() {
	args := os.Args
	if len(args) != 4 {
		fmt.Println("usage ./client <server_address> <port_number> <file_name>")
		return
	}

	connectTo := args[1] + ":" + args[2]

	file := ReadFile(args[3])

	conn, err := net.Dial("tcp", connectTo)
	if err != nil {
		fmt.Println(err)
		return
	}

	message.NewMessage().HELLO().Send(conn)

	for {

		msg, err := bufio.NewReader(conn).ReadBytes('\n')
		msg = msg[:len(msg)-1]
		if err != nil {
			fmt.Println(err)
			return
		}

		client := &Client{
			TCPconn: conn,
			file:    file,
		}

		client.handleMsg(msg)

	}
}
