package message

import (
	"net"

	"github.com/luiz-couto/File-Transfer-UDP-TCP/pkg/bytes"
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
	body []byte
}

// NewMessage creates a new Message
func NewMessage() *Message {
	return &Message{}
}

//HELLO defines the HELLO message type
func (msg *Message) HELLO() *Message {
	id := bytes.WriteIntAsBytes(2, HelloType)

	msg.body = id
	return msg
}

//CONNECTION defines the HELLO message type
func (msg *Message) CONNECTION(port int) *Message {
	id := bytes.WriteIntAsBytes(2, ConnectionType)
	bPort := bytes.WriteIntAsBytes(4, port)

	msg.body = append(id, bPort...)
	return msg
}

//INFOFILE define the INFO FILE message type
func (msg *Message) INFOFILE(fileName string, fileSize int) *Message {
	id := bytes.WriteIntAsBytes(2, InfoFileType)
	bFileName := bytes.CreateByteBlock(15, []byte(fileName))
	bFileSize := bytes.WriteIntAsBytes(8, fileSize)

	msg.body = append(id, bFileName...)
	msg.body = append(msg.body, bFileSize...)
	return msg
}

//OK defines the OK message type
func (msg *Message) OK() *Message {
	id := bytes.WriteIntAsBytes(2, OKType)

	msg.body = id
	return msg
}

//FIM defines the FIM message type
func (msg *Message) FIM() *Message {
	id := bytes.WriteIntAsBytes(2, FimType)

	msg.body = id
	return msg
}

// FILE defines the FILE message type
func (msg *Message) FILE(seqNumber int, payloadSize int, payload []byte) *Message {
	id := bytes.WriteIntAsBytes(2, FileType)
	bSeqNumber := bytes.WriteIntAsBytes(4, seqNumber)
	bPayloadSize := bytes.WriteIntAsBytes(2, payloadSize)

	bPayload := bytes.CreateByteBlock(payloadSize, payload)

	msg.body = append(id, bSeqNumber...)
	msg.body = append(msg.body, bPayloadSize...)
	msg.body = append(msg.body, bPayload...)
	return msg
}

// ACK defines the ACK message type
func (msg *Message) ACK(seqNumber int) *Message {
	id := bytes.WriteIntAsBytes(2, AckType)
	bSeqNumber := bytes.WriteIntAsBytes(4, seqNumber)

	msg.body = append(id, bSeqNumber...)
	return msg
}

// Send sends the message for given connection
func (msg *Message) Send(conn net.Conn) {
	conn.Write(msg.body)
}

// SendFile sends the pkg file to the given connection
func (msg *Message) SendFile(conn *net.UDPConn) {
	conn.Write(msg.body)
}
