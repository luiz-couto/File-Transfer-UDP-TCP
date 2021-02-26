package broker

import "sync"

// Worker DOC TODO
type Worker struct {
	Source chan int
	Quit   chan struct{}
}

// ThreadSafeSlice DOC TODO
type ThreadSafeSlice struct {
	sync.Mutex
	Workers []*Worker
}

// NewBroker DOC TODO
func NewBroker() *ThreadSafeSlice {
	return &ThreadSafeSlice{
		Workers: []*Worker{},
	}
}

// Push DOC TODO
func (slice *ThreadSafeSlice) Push(w *Worker) {
	slice.Lock()
	defer slice.Unlock()

	slice.Workers = append(slice.Workers, w)
}

// Iter DOC TODO
func (slice *ThreadSafeSlice) Iter(routine func(*Worker)) {
	slice.Lock()
	defer slice.Unlock()

	for _, worker := range slice.Workers {
		routine(worker)
	}
}
