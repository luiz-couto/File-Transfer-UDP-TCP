package message

import (
	"fmt"
	"net"
	"strconv"
)

// Message Types
const (
	HelloType int = iota + 1
	ConnectionType
	InfoFileType
	OKType
	FimType
	FileType
	AckType
)

//Message define the message struct
type Message struct {
	Type    int
	Payload []byte
}

// NewMessage creates a new Message
func NewMessage() *Message {
	return &Message{}
}

//HELLO defines the HELLO message type
func (msg *Message) HELLO() {
	msg.Type = HelloType
}

//CONNECTION defines the HELLO message type
func (msg *Message) CONNECTION(port int) {
	msg.Type = ConnectionType
	msg.Payload = []byte(strconv.Itoa(port))
}

//INFOFILE define the INFO FILE message type
func (msg *Message) INFOFILE(fileName string, fileSize int) {
	msg.Type = InfoFileType

	fileNameBytes := []byte(fileName)
	fileSizeBytes := []byte(strconv.Itoa(fileSize))

	msg.Payload = append(fileNameBytes, fileSizeBytes...)
}

//OK defines the OK message type
func (msg *Message) OK() {
	msg.Type = OKType
}

//FIM defines the FIM message type
func (msg *Message) FIM() {
	msg.Type = FimType
}

// FILE defines the FILE message type
func (msg *Message) FILE(seqNumber int, payloadSize int, payload []byte) {
	msg.Type = FileType

	seqNumberBytes := []byte(strconv.Itoa(seqNumber))
	payloadSizeBytes := []byte(strconv.Itoa(payloadSize))

	msg.Payload = append(seqNumberBytes, payloadSizeBytes...)
	msg.Payload = append(msg.Payload, payload...)
}

// ACK defines the ACK message type
func (msg *Message) ACK(seqNumber int) {
	msg.Type = AckType
	msg.Payload = []byte(strconv.Itoa(seqNumber))
}

// Send sends the message for given connection
func (msg *Message) Send(conn net.Conn) {
	message := append([]byte(strconv.Itoa(msg.Type)), msg.Payload...)
	fmt.Fprintf(conn, string(message))
}
