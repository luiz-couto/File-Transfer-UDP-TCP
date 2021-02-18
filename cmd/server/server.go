package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

//UDPConnection defines the udp connection object
type UDPConnection struct {
	connUDP *net.UDPConn
	port    int
}

func startUDPConnection() (*UDPConnection, error) {
	// Get a Free Port Number
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	udpAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(listener.Addr().(*net.TCPAddr).Port))
	if err != nil {
		return nil, err
	}

	connUDP, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	return &UDPConnection{
		connUDP: connUDP,
		port:    udpAddr.Port,
	}, nil

}

func handleConnection(conn net.Conn) {
	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		msg = strings.TrimSpace(string(msg))
		fmt.Println(msg)

		UDP, err := startUDPConnection()
		if err != nil {
			fmt.Println(err)
			return
		}
		defer UDP.connUDP.Close()

		fmt.Fprintf(conn, "Message Received! Your UDP port is "+fmt.Sprint(UDP.port)+"\n")

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
