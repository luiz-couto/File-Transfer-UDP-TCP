package broker

// Broker DOC TODO
type Broker struct {
	stopCh    chan struct{}
	publishCh chan interface{}
	subCh     chan chan interface{}
	unsubCh   chan chan interface{}
}

// NewBroker DOC TODO
func NewBroker() *Broker {
	return &Broker{
		stopCh:    make(chan struct{}),
		publishCh: make(chan interface{}, 10),
		subCh:     make(chan chan interface{}, 10),
		unsubCh:   make(chan chan interface{}, 10),
	}
}

// Start DOC TODO
func (b *Broker) Start() {
	subs := map[chan interface{}]struct{}{}
	for {
		select {
		case <-b.stopCh:
			return
		case msgCh := <-b.subCh:
			subs[msgCh] = struct{}{}
		case msgCh := <-b.unsubCh:
			delete(subs, msgCh)
		case msg := <-b.publishCh:
			for msgCh := range subs {
				// msgCh is buffered, use non-blocking send to protect the broker:
				select {
				case msgCh <- msg:
				default:
				}
			}
		}
	}
}

// Stop DOC TODO
func (b *Broker) Stop() {
	close(b.stopCh)
}

// Subscribe DOC TODO
func (b *Broker) Subscribe() chan interface{} {
	msgCh := make(chan interface{}, 5)
	b.subCh <- msgCh
	return msgCh
}

// Unsubscribe DOC TODO
func (b *Broker) Unsubscribe(msgCh chan interface{}) {
	b.unsubCh <- msgCh
}

// Publish DOC TODO
func (b *Broker) Publish(msg interface{}) {
	b.publishCh <- msg
}
