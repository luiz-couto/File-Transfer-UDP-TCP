package message

import "strconv"

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
