package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/luiz-couto/File-Transfer-UDP-TCP/pkg/bytes"
	"github.com/luiz-couto/File-Transfer-UDP-TCP/pkg/message"
)

var globalQuit chan struct{} = make(chan struct{})

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
	UDPconn         *net.UDPConn
	TCPconn         net.Conn
	SliWindow       *SlidingWindow
	file            *File
	tss             *threadSafeSlice
	endTransmission bool
}

type worker struct {
	source chan int
	quit   chan struct{}
}

type threadSafeSlice struct {
	sync.Mutex
	workers []*worker
}

func (slice *threadSafeSlice) Push(w *worker) {
	slice.Lock()
	defer slice.Unlock()

	slice.workers = append(slice.workers, w)
}

func (slice *threadSafeSlice) Iter(routine func(*worker)) {
	slice.Lock()
	defer slice.Unlock()

	for _, worker := range slice.workers {
		routine(worker)
	}
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

func (c *Client) getNextWindow() []int {
	var nxtWin []int

	for i := 0; i < c.SliWindow.windowSize; i++ {
		if len(c.SliWindow.lostPkgs) > 0 {
			nxtWin = append(nxtWin, c.SliWindow.lostPkgs[0])

			c.SliWindow.mutex.Lock()
			c.SliWindow.lostPkgs = RemoveFromSlice(c.SliWindow.lostPkgs, c.SliWindow.lostPkgs[0])
			c.SliWindow.mutex.Unlock()

			continue
		}

		if c.SliWindow.nxtPkg != len(c.SliWindow.pkgs) {
			nxtWin = append(nxtWin, c.SliWindow.nxtPkg)
			c.SliWindow.nxtPkg = c.SliWindow.nxtPkg + 1
		}
	}
	return nxtWin
}

func (c *Client) checkNxtWindowEmpty(nxtWindow []int) bool {
	for _, v := range nxtWindow {
		if v != -1 {
			return false
		}
	}
	return true
}

func (c *Client) waitForAck(ctx context.Context, seqNum int, cancel context.CancelFunc, w *worker) {
	w.source = make(chan int, 10)
	w.quit = globalQuit

	go func() {
		fmt.Println("Started thread " + strconv.Itoa(seqNum))
		defer cancel()
		for {
			select {
			case rcvAck := <-w.source:
				fmt.Println("ACK -> " + strconv.Itoa(rcvAck) + " / Thread " + strconv.Itoa(seqNum))
				if rcvAck == seqNum {
					fmt.Println(time.Now().Format(time.RFC850) + "PASSSSOUU AQQQ -> " + "ACK -> " + strconv.Itoa(seqNum))
					return
				}

			case <-ctx.Done():
				fmt.Println("TIMEOUT: " + strconv.Itoa(seqNum))
				c.SliWindow.mutex.Lock()
				c.SliWindow.lostPkgs = append(c.SliWindow.lostPkgs, seqNum)
				c.SliWindow.mutex.Unlock()
				return

			case <-w.quit:
				return
			}
		}
	}()

}

func (c *Client) sendNxtWindow(nxtWindow []int) {
	for _, seqNum := range nxtWindow {
		pkg := c.SliWindow.pkgs[seqNum]

		ctx, cancel := context.WithTimeout(context.Background(), 3200*time.Millisecond)

		w := &worker{}
		c.waitForAck(ctx, seqNum, cancel, w)
		c.tss.Push(w)

		message.NewMessage().FILE(seqNum, len(pkg), pkg).SendFile(c.UDPconn)
	}
}

func (c *Client) startFileTransmission() {
	pkgs := bytes.DivideInPackages(c.file.content, 1000)
	wsize := len(pkgs) / 2

	sliWin := &SlidingWindow{
		windowSize: wsize,
		lostPkgs:   []int{},
		nxtPkg:     0,
		pkgs:       pkgs,
	}

	tss := &threadSafeSlice{
		workers: []*worker{},
	}

	c.SliWindow = sliWin
	c.tss = tss

	for {

		if c.endTransmission {
			fmt.Println("ENDING CONNECTION")
			c.UDPconn.Close()
			c.TCPconn.Close()
			break
		}

		nxtWin := c.getNextWindow()
		time.Sleep(100 * time.Millisecond)
		if len(nxtWin) > 0 {
			fmt.Println(nxtWin)
		}

		c.sendNxtWindow(nxtWin)
	}
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

	case message.OKType:
		fmt.Println("Received OK")
		go c.startFileTransmission()

	case message.AckType:
		fmt.Println("Received ACK")
		seqNum := bytes.ReadByteBlockAsInt(2, 6, msg)
		//c.ack.Publish(seqNum)

		c.tss.Iter(func(w *worker) { w.source <- seqNum })

		return

	case message.FimType:
		fmt.Println("Received FIM")
		close(globalQuit)

		c.tss.Mutex.Lock()
		c.SliWindow.lostPkgs = []int{}
		c.tss.Mutex.Unlock()

		c.endTransmission = true
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

	client := &Client{
		TCPconn:         conn,
		file:            file,
		endTransmission: false,
	}

	for {

		fmt.Println("Estou escutando...")
		msg, err := bufio.NewReader(conn).ReadBytes('\n')

		if err != nil {
			fmt.Println(err)
			return
		}

		client.handleMsg(msg)

	}
}
