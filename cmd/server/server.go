package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/luiz-couto/File-Transfer-UDP-TCP/pkg/bytes"
	"github.com/luiz-couto/File-Transfer-UDP-TCP/pkg/message"
)

//maxBufferSize DOC TODO
const maxBufferSize = 1008

// Pkg DOC TODO
type Pkg struct {
	seqNumber int
	payload   []byte
}

// FileBuffer DOC TODO
type FileBuffer struct {
	fileSize   int
	fileName   string
	pkgBuckets [][]Pkg
	rcvLog     map[int]bool
}

//Client DOC TODO
type Client struct {
	connTCP    net.Conn
	connUDP    *UDPConnection
	fileBuffer *FileBuffer
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
		msg, err := bufio.NewReader(c.connTCP).ReadBytes(255)

		if len(msg) == 0 {
			fmt.Println("FINISH CLIENT CONNECTION")
			return
		}

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

	case message.InfoFileType:
		fmt.Println("Received INFO_FILE")
		fileName := bytes.ReadByteBlockAsString(2, 17, msg)
		fileSize := bytes.ReadByteBlockAsInt(17, 25, msg)

		fileBuffer := &FileBuffer{
			fileName: fileName,
			fileSize: fileSize,
			rcvLog:   make(map[int]bool),
		}

		c.fileBuffer = fileBuffer

		message.NewMessage().OK().Send(c.connTCP)

		c.receiveFile()

	}
}

func (c *Client) receiveFile() {
	totalLen := 0
	for {
		buffer := make([]byte, maxBufferSize)
		n, _, err := c.connUDP.UDP.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			return
		}

		msg := buffer[:n-1]
		msgID := bytes.ReadByteBlockAsInt(0, 2, msg)
		if msgID != message.FileType {
			continue
		}

		fmt.Println("Received FILE")
		seqNumber := bytes.ReadByteBlockAsInt(2, 6, msg)
		payloadSize := bytes.ReadByteBlockAsInt(6, 8, msg)
		payload := msg[8 : 8+payloadSize]

		message.NewMessage().ACK(seqNumber).Send(c.connTCP)
		time.Sleep(100 * time.Millisecond)

		if _, ok := c.fileBuffer.rcvLog[seqNumber]; ok {
			continue
		}

		c.addToPkgBucket(seqNumber, payload)
		totalLen = totalLen + len(payload)
		c.fileBuffer.rcvLog[seqNumber] = true

		fmt.Println(totalLen)

		if totalLen == c.fileBuffer.fileSize {
			message.NewMessage().FIM().Send(c.connTCP)
			time.Sleep(100 * time.Millisecond)

			c.writeFile()

			// for _, bckt := range c.fileBuffer.pkgBuckets {
			// 	fmt.Printf("[")
			// 	for _, pkg := range bckt {
			// 		fmt.Printf("%v, ", pkg.seqNumber)
			// 	}
			// 	fmt.Printf("]\n")
			// }

			c.connUDP.UDP.Close()
			c.connTCP.Close()
			break
		}
	}
}

func (c *Client) addToPkgBucket(seqNum int, payload []byte) {
	newPkg := Pkg{
		seqNumber: seqNum,
		payload:   payload,
	}

	for i, bucket := range c.fileBuffer.pkgBuckets {
		lastPkg := bucket[len(bucket)-1]
		if seqNum == lastPkg.seqNumber+1 {
			c.fileBuffer.pkgBuckets[i] = append(bucket, newPkg)
			return
		}
	}

	var newBucket []Pkg
	newBucket = append(newBucket, newPkg)

	c.fileBuffer.pkgBuckets = append(c.fileBuffer.pkgBuckets, newBucket)
}

func (c *Client) writeFile() {
	file := getPayload(c.fileBuffer.pkgBuckets, len(c.fileBuffer.rcvLog)-1)

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)

	err := ioutil.WriteFile(basepath+"/"+c.fileBuffer.fileName, file, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func getPayload(buckets [][]Pkg, lastPkgIdx int) []byte {
	var final []byte
	var zeroBckt []Pkg
	for _, v := range buckets {
		if v[0].seqNumber == 0 {
			zeroBckt = v
		}
	}

	currBckt := zeroBckt
	for {
		for _, v := range currBckt {
			final = append(final, v.payload...)
		}
		if currBckt[len(currBckt)-1].seqNumber == lastPkgIdx {
			break
		}

		nxtFirstPkg := currBckt[len(currBckt)-1].seqNumber + 1

		for _, v := range buckets {
			if v[0].seqNumber == nxtFirstPkg {
				currBckt = v
			}
		}
	}

	return final
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
