package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/luiz-couto/File-Transfer-UDP-TCP/pkg/bytes"
	"github.com/luiz-couto/File-Transfer-UDP-TCP/pkg/message"
)

/*
maxBufferSize is the largest possible message size
*/
const maxBufferSize = 1008

/*
Pkg defines the Pkg structure
*/
type Pkg struct {
	seqNumber int
	payload   []byte
}

/*
FileBuffer defines the FileBuffer structure
*/
type FileBuffer struct {
	fileSize   int
	fileName   string
	pkgBuckets [][]Pkg
	rcvLog     map[int]bool
}

/*
Client defines the client structure
*/
type Client struct {
	connTCP    net.Conn
	connUDP    *UDPConnection
	fileBuffer *FileBuffer
	reader     *bufio.Reader
}

/*
UDPConnection defines the UDP structure
*/
type UDPConnection struct {
	UDP  *net.UDPConn
	port int
}

/*
startUDPConnection initiates the UDP connection and adds a pointer to that connection to the
client's structure
*/
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

/*
handleConnection is the thread of each client that waits for messages
*/
func (c *Client) handleConnection() {
	for {
		buf := make([]byte, 2)
		n, _ := io.ReadFull(c.reader, buf)

		if n == 0 {
			fmt.Println("Arquivo recebido com sucesso")
			fmt.Println("Conexão com o cliente finalizada")
			return
		}

		c.handleMsg(buf)
	}
}

/*
handleMsg decides the type of the message for the given bytes
and chooses how to handle it based on the type
*/
func (c *Client) handleMsg(msgType []byte) {
	msgID := bytes.ReadByteBlockAsInt(0, 2, msgType)
	switch msgID {
	case message.HelloType:
		fmt.Println("Received HELLO")
		c.startUDPConnection()
		message.NewMessage().CONNECTION(c.connUDP.port).Send(c.connTCP)

	case message.InfoFileType:
		fmt.Println("Received INFO_FILE")

		buf := make([]byte, 15)
		io.ReadFull(c.reader, buf)
		fileName := bytes.ReadByteBlockAsString(0, 15, buf)

		buf2 := make([]byte, 8)
		io.ReadFull(c.reader, buf2)
		fileSize := bytes.ReadByteBlockAsInt(0, 8, buf2)

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

/*
receiveFile waits for FILE messages and sends an ACK to the respective sequence number.
For each received packet, it is added to a client buffer. It will be running until the
expected file size is reached.
*/
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

		if _, ok := c.fileBuffer.rcvLog[seqNumber]; ok {
			continue
		}

		c.addToPkgBucket(seqNumber, payload)
		totalLen = totalLen + len(payload)
		c.fileBuffer.rcvLog[seqNumber] = true

		fmt.Printf("Total recebido (acumulado): %v bytes\n", totalLen)

		if totalLen == c.fileBuffer.fileSize {
			message.NewMessage().FIM().Send(c.connTCP)
			time.Sleep(100 * time.Millisecond)

			c.writeFile()

			c.connUDP.UDP.Close()
			c.connTCP.Close()
			break
		}
	}
}

/*
addToPkgBucket receives a sequence number and payload for a package and decides which
bucket that package will be added to or whether a new bucket will be created for that
package
*/
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

/*
writeFile will get the file payload and write the file in the current directory
*/
func (c *Client) writeFile() {
	file := getPayload(c.fileBuffer.pkgBuckets, len(c.fileBuffer.rcvLog)-1)

	dir, _ := os.Getwd()

	err := ioutil.WriteFile(dir+"/"+c.fileBuffer.fileName, file, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

/*
getPayload receives buckets of packages and joins them in the correct order,
returning the file in bytes
*/
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

		reader := bufio.NewReader(conn)

		client := &Client{
			connTCP: conn,
			reader:  reader,
		}
		go client.handleConnection()
	}

}
