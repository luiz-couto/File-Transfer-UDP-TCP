package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func handleConnection(conn net.Conn) {
	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		msg = strings.TrimSpace(string(msg))
		fmt.Println(msg)

		fmt.Fprintf(conn, "Message Received!\n")
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
		go handleConnection(conn)
	}

}
