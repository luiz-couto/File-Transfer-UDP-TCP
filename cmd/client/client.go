package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

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

	for {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print(">> ")

		msg, _ := reader.ReadString('\n')
		fmt.Fprintf(conn, msg+"\n")

		resp, err := bufio.NewReader(conn).ReadString('\n') // Change to ReadBytes!
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Print("[server] " + resp)
	}
}
