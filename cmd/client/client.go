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

func handleMsg(msg []byte) {
	msgID, _ := strconv.Atoi(bytes.ReadByteBlockAsString(0, 2, msg))
	switch msgID {
	case message.ConnectionType:
		fmt.Println("Received CONNECTION")
		port := bytes.ReadByteBlockAsString(2, 6, msg)
		fmt.Println("Porto UDP is " + port)
	}
}

func main() {
	args := os.Args
	if len(args) != 3 {
		fmt.Println("usage ./client <server_address> <port_number>")
		return
	}

	connectTo := args[1] + ":" + args[2]

	conn, err := net.Dial("tcp", connectTo)
	if err != nil {
		fmt.Println(err)
		return
	}

	message.NewMessage().HELLO().Send(conn)

	for {

		msg, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		handleMsg(msg)

	}
}
