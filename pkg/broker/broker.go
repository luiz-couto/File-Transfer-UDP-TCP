package broker

import "sync"

/*
Worker defines the Worker structure
*/
type Worker struct {
	Source chan int
	Quit   chan struct{}
}

/*
ThreadSafeSlice defines the ThreadSafeSlice structure
*/
type ThreadSafeSlice struct {
	sync.Mutex
	Workers []*Worker
}

/*
NewBroker returns a new thread safe slice
*/
func NewBroker() *ThreadSafeSlice {
	return &ThreadSafeSlice{
		Workers: []*Worker{},
	}
}

/*
Push adds a new worker to the thread safe slice
*/
func (slice *ThreadSafeSlice) Push(w *Worker) {
	slice.Lock()
	defer slice.Unlock()

	slice.Workers = append(slice.Workers, w)
}

/*
Iter iterates over the thread safe slice and apply the given
function to all the workers
*/
func (slice *ThreadSafeSlice) Iter(routine func(*Worker)) {
	slice.Lock()
	defer slice.Unlock()

	for _, worker := range slice.Workers {
		routine(worker)
	}
}
