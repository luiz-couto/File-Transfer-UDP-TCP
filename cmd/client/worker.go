package main

import "sync"

// Worker DOC TODO
type Worker struct {
	source chan int
	quit   chan struct{}
}

// ThreadSafeSlice DOC TODO
type ThreadSafeSlice struct {
	sync.Mutex
	workers []*Worker
}

// Push DOC TODO
func (slice *ThreadSafeSlice) Push(w *Worker) {
	slice.Lock()
	defer slice.Unlock()

	slice.workers = append(slice.workers, w)
}

// Iter DOC TODO
func (slice *ThreadSafeSlice) Iter(routine func(*Worker)) {
	slice.Lock()
	defer slice.Unlock()

	for _, worker := range slice.workers {
		routine(worker)
	}
}
