package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/luiz-couto/File-Transfer-UDP-TCP/pkg/broker"
	"github.com/luiz-couto/File-Transfer-UDP-TCP/pkg/bytes"
	"github.com/luiz-couto/File-Transfer-UDP-TCP/pkg/message"
)

var globalQuit chan struct{} = make(chan struct{})

const maxPkgLen = 1000

// File DOC TODO
type File struct {
	fileSize int
	fileName string
	content  []byte
}

// SlidingWindow DOC TODO
type SlidingWindow struct {
	windowSize int
	lostPkgs   []int
	pkgs       map[int][]byte
	nxtPkg     int
	mutex      sync.Mutex
}

// Client DOC TODO
type Client struct {
	UDPconn   *net.UDPConn
	TCPconn   net.Conn
	address   string
	SliWindow *SlidingWindow
	file      *File
	tss       *broker.ThreadSafeSlice
	sndNxt    chan struct{}
	reader    *bufio.Reader
}

/*
RemoveFromSlice return a new slice without any occurence of the
target int t
*/
func RemoveFromSlice(vs []int, t int) []int {
	var newSlice []int
	for _, v := range vs {
		if v != t {
			newSlice = append(newSlice, v)
		}
	}
	return newSlice
}

// ReadFile DOC TODO
func ReadFile(fileName string) (*File, error) {
	fileContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return &File{
		fileSize: len(fileContent),
		fileName: filepath.Base(fileName),
		content:  fileContent,
	}, nil
}

func (c *Client) startUDPConnection(port int) {
	addr, err := net.ResolveUDPAddr("udp", c.address+":"+strconv.Itoa(port))
	udpConn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.UDPconn = udpConn
}

func (c *Client) waitForAck(ctx context.Context, seqNum int, cancel context.CancelFunc, w *broker.Worker) {
	w.Source = make(chan int, 1000)
	w.Quit = globalQuit

	go func() {
		fmt.Println("Started thread " + strconv.Itoa(seqNum))
		defer cancel()
		for {
			select {
			case rcvAck := <-w.Source:
				//fmt.Println("ACK -> " + strconv.Itoa(rcvAck) + " / Thread " + strconv.Itoa(seqNum))
				if rcvAck == seqNum {
					c.SliWindow.mutex.Lock()
					fmt.Println(time.Now().Format(time.RFC850) + "PASSSSOUU AQQQ -> " + "ACK -> " + strconv.Itoa(seqNum))
					c.sndNxt <- struct{}{}
					return
				}

			case <-ctx.Done():
				fmt.Println("TIMEOUT: " + strconv.Itoa(seqNum))

				c.SliWindow.mutex.Lock()
				c.SliWindow.lostPkgs = append(c.SliWindow.lostPkgs, seqNum)

				c.sndNxt <- struct{}{}

				return

			case <-w.Quit:
				return
			}
		}
	}()
}

func (c *Client) sendNxtPkg() {
	defer c.SliWindow.mutex.Unlock()

	var nxtSeqNum int

	if len(c.SliWindow.lostPkgs) > 0 {
		nxtSeqNum = c.SliWindow.lostPkgs[0]
		c.SliWindow.lostPkgs = RemoveFromSlice(c.SliWindow.lostPkgs, c.SliWindow.lostPkgs[0])
	} else {
		if c.SliWindow.nxtPkg != len(c.SliWindow.pkgs) {
			nxtSeqNum = c.SliWindow.nxtPkg
			c.SliWindow.nxtPkg = c.SliWindow.nxtPkg + 1
		} else {
			return
		}
	}

	pkg := c.SliWindow.pkgs[nxtSeqNum]
	ctx, cancel := context.WithTimeout(context.Background(), 3200*time.Millisecond)

	w := &broker.Worker{}
	c.waitForAck(ctx, nxtSeqNum, cancel, w)
	c.tss.Push(w)

	message.NewMessage().FILE(nxtSeqNum, len(pkg), pkg).SendFile(c.UDPconn)
}

func (c *Client) startFileTransmission() {
	pkgs := bytes.DivideInPackages(c.file.content, 1000)
	wsize := len(pkgs) / 2
	if wsize > 7 {
		wsize = 7
	}

	sliWin := &SlidingWindow{
		windowSize: wsize,
		lostPkgs:   []int{},
		nxtPkg:     0,
		pkgs:       pkgs,
	}

	tss := broker.NewBroker()

	c.SliWindow = sliWin
	c.tss = tss

	// Send first window
	for i := 0; i < c.SliWindow.windowSize; i++ {
		c.SliWindow.mutex.Lock()
		c.sendNxtPkg()
	}

	for {
		select {
		case <-c.sndNxt:
			c.sendNxtPkg()

		case <-globalQuit:
			fmt.Println("ENDING CONNECTION")
			c.UDPconn.Close()
			c.TCPconn.Close()
			return
		}
	}
}

func (c *Client) handleMsg(msgType []byte) {
	msgID := bytes.ReadByteBlockAsInt(0, 2, msgType)
	switch msgID {
	case message.ConnectionType:
		fmt.Println("Received CONNECTION")

		buf := make([]byte, 4)
		io.ReadFull(c.reader, buf)

		port := bytes.ReadByteBlockAsInt(0, 4, buf)

		fmt.Println("Porto UDP is " + strconv.Itoa(port))
		c.startUDPConnection(port)

		message.NewMessage().INFOFILE(c.file.fileName, c.file.fileSize).Send(c.TCPconn)

	case message.OKType:
		fmt.Println("Received OK")
		go c.startFileTransmission()

	case message.AckType:
		fmt.Println("Received ACK")

		buf := make([]byte, 4)
		io.ReadFull(c.reader, buf)

		seqNum := bytes.ReadByteBlockAsInt(0, 4, buf)

		c.tss.Iter(func(w *broker.Worker) { w.Source <- seqNum })

		return

	case message.FimType:
		fmt.Println("Received FIM")
		close(globalQuit)

		c.tss.Mutex.Lock()
		c.SliWindow.lostPkgs = []int{}
		c.tss.Mutex.Unlock()
	}
}

func validateFileName(name string) bool {
	if len([]byte(name)) > 15 || strings.Count(name, ".") != 1 || len(strings.Split(name, ".")[1]) != 3 || !isASCII(name) {
		return false
	}
	return true
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func main() {
	args := os.Args
	if len(args) != 4 {
		fmt.Println("usage ./client <server_address> <port_number> <file_name>")
		return
	}

	if !validateFileName(filepath.Base(args[3])) {
		fmt.Println("Nome não permitido.")
		return
	}

	// Check if is IPv6
	address := args[1]
	if !(strings.Count(address, ":") < 2) {
		address = "[" + args[1] + "]"
	}

	file, err := ReadFile(args[3])
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := net.Dial("tcp", address+":"+args[2])
	if err != nil {
		fmt.Println(err)
		return
	}

	reader := bufio.NewReader(conn)

	message.NewMessage().HELLO().Send(conn)

	client := &Client{
		TCPconn: conn,
		file:    file,
		sndNxt:  make(chan struct{}),
		address: address,
		reader:  reader,
	}

	for {
		fmt.Println("Estou escutando...")
		buf := make([]byte, 2)

		n, _ := io.ReadFull(reader, buf)
		if n == 0 {
			return
		}

		client.handleMsg(buf)
	}
}
