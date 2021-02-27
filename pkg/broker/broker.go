package broker

import "sync"

/*
Worker defines the Worker structure
*/
type Worker struct {
	ID     int
	Source chan int
	Quit   chan struct{}
}

/*
ThreadSafeSlice defines the ThreadSafeSlice structure
*/
type ThreadSafeSlice struct {
	sync.Mutex
	Workers []*Worker
	nxtID   int
}

/*
NewBroker returns a new thread safe slice
*/
func NewBroker() *ThreadSafeSlice {
	return &ThreadSafeSlice{
		Workers: []*Worker{},
		nxtID:   0,
	}
}

/*
Push adds a new worker to the thread safe slice
*/
func (slice *ThreadSafeSlice) Push(w *Worker) {
	slice.Lock()
	defer slice.Unlock()

	w.ID = slice.nxtID
	slice.nxtID = slice.nxtID + 1
	slice.Workers = append(slice.Workers, w)
}

/*
Remove removes the given worker from the thread safe slice
*/
func (slice *ThreadSafeSlice) Remove(w *Worker) {
	slice.Lock()
	defer slice.Unlock()

	var newSlice []*Worker
	for _, v := range slice.Workers {
		if v.ID != w.ID {
			newSlice = append(newSlice, v)
		}
	}
	slice.Workers = newSlice
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
