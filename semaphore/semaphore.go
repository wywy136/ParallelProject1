package semaphore

import (
	"sync"
)

type Semaphore struct {
	value int
	mutex *sync.Mutex
	cond *sync.Cond
}

func NewSemaphore(capacity int) *Semaphore {
	mutex := &sync.Mutex{}
	cond := sync.NewCond(mutex)
	return &Semaphore{capacity, mutex, cond}
}

func (s *Semaphore) Up() {
	s.mutex.Lock()
	s.value ++
	s.cond.Signal()
	s.mutex.Unlock()
}

func (s *Semaphore) Down() {
	s.mutex.Lock()
	for s.value <= 0 {
		s.cond.Wait()
	}
	s.value --	
	s.mutex.Unlock()
}
